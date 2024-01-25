package gosepp

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// CallID custom callID type
type CallID string

// Call is an abstraction of the gosepp messaging based interface.
type Call struct {
	sepp                *GoSepp
	confID              string
	clientID            string
	callID              CallID
	terminationHandler  func()
	sdpUpdateHandler    func(Sdp)
	memberlistHandler   func(MsgMemberlistData)
	sourceUpdateHandler func(MsgSourceUpdateData)
	cancel              context.CancelFunc
	termCh              chan bool
	logger              Logger
	customCAFile        string
	platform            string
}

// CallOption defines the options interface
type CallOption func(*Call)

// WithCustomCAFile This option configures this library
// to use a custom-CA instead of the systemCA.
func WithCustomCAFile(customCAFile string) CallOption {
	return func(c *Call) {
		c.customCAFile = customCAFile
	}
}

// WithPlatformVersion allows to specify the platform-version
// string which will be used during call-setup.
func WithPlatformVersion(platform string) CallOption {
	return func(c *Call) {
		c.platform = platform
	}
}

// NewCall initializes an instance of a call.
func NewCall(callInfo CallInfoInterface, logger Logger, options ...CallOption) (*Call, error) {

	if logger == nil {
		logger = &silentLogger{}
	}

	call := &Call{
		confID:   callInfo.GetConfID(),
		clientID: callInfo.GetClientID(),
		termCh:   make(chan bool),
		logger:   logger,
	}

	for _, opt := range options {
		opt(call)
	}

	var tlsConfig *tls.Config
	if len(call.customCAFile) > 0 {
		// Load CA cert
		caCert, err := ioutil.ReadFile(call.customCAFile)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("Failed to append CAcert")
		}
		tlsConfig = &tls.Config{
			RootCAs: caCertPool,
		}
	}

	sepp, err := NewGoSepp(callInfo.GetSigEndpoint(), callInfo.GetAuthToken(),
		tlsConfig, logger)
	if err != nil {
		return nil, err
	}

	call.sepp = sepp
	return call, nil
}

// SetTerminatedHandler sets the termination handler which is
// called when the call is terminated.
// Must be set-up before start.
func (c *Call) SetTerminatedHandler(handler func()) {
	c.terminationHandler = handler
}

// SetSDPUpdateHandler sets the sdp-update handler which is
// called if the remote end is sending an updated
// sdp.
// Must be set-up before start.
func (c *Call) SetSDPUpdateHandler(handler func(Sdp)) {
	c.sdpUpdateHandler = handler
}

// SetMemberlistHandler set handler to be called on change of
// the memberlist.
func (c *Call) SetMemberlistHandler(handler func(MsgMemberlistData)) {
	c.memberlistHandler = handler
}

// SetSourceUpdateHandler set handler to be called if the podium
// layout changes.
func (c *Call) SetSourceUpdateHandler(handler func(MsgSourceUpdateData)) {
	c.sourceUpdateHandler = handler
}

func startDispatch(ctx context.Context, logger Logger, sepp *GoSepp,
	termHandler func(), sdpUpdateHandler func(Sdp),
	memberlistHandler func(MsgMemberlistData),
	sourceUpdateHandler func(MsgSourceUpdateData), termCh chan<- bool) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-sepp.RcvCh():
			if !ok {
				logger.Info("Channel closed. Stopping dispatch")
				return
			}
			// dispatch messages
			switch m := msg.(type) {
			case *MsgCallTerminated:
				// try to signal on the term channel
				select {
				case termCh <- true:
				default:
					//log.Println("Timout when calling term channel")
				}
				if termHandler != nil {
					termHandler()
				}
			case *MsgSdpUpdate:
				if sdpUpdateHandler != nil {
					sdpUpdateHandler(m.Data.Sdp)
				}
			case *MsgMemberlist:
				if memberlistHandler != nil {
					memberlistHandler(m.Data)
				}
			case *MsgSourceUpdate:
				if sourceUpdateHandler != nil {
					sourceUpdateHandler(m.Data)
				}
			default:
			}
		}
	}
}

// Start the call. On success the call-id and sdp is returned,
// else an error.
func (c *Call) Start(ctx context.Context, sdp Sdp, displayname string) (*CallID, *Sdp, error) {
	if len(c.callID) > 0 {
		return nil, nil, fmt.Errorf("call already in progress")
	}

	callCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	// wait for connected
	select {
	case connected, ok := <-c.sepp.ConnectStatusCh():
		if !ok || !connected {
			return nil, nil, fmt.Errorf("Failed to connect")
		}
	case <-callCtx.Done():
		return nil, nil, fmt.Errorf("Timeout. Failed to connect")
	}

	// send start call message
	if err := c.sepp.SendMsg(MsgCallStart{
		MsgBase: MsgBase{
			Type: MsgTypeCallStart,
			From: c.clientID,
			To:   c.confID,
		},
		Data: MsgCallStartData{
			Sdp:         sdp,
			DisplayName: displayname,
			Platform:    c.platform,
		},
	}); err != nil {
		return nil, nil, fmt.Errorf("failed to send message: %s", err)
	}

	for {
		// wait for call accepted or rejected
		select {
		case msg, ok := <-c.sepp.RcvCh():
			if !ok {
				return nil, nil, fmt.Errorf("Failed to receive")
			}
			// dispatch messages
			switch m := msg.(type) {
			case *MsgMemberlist:
				// Continue if a memberlist was received.
				continue
			case *MsgCallAccepted:
				callID := CallID(m.Data.CallID)
				c.callID = callID
				// start dispatcher as goroutine
				go startDispatch(callCtx, c.logger, c.sepp, c.terminationHandler,
					c.sdpUpdateHandler, c.memberlistHandler, c.sourceUpdateHandler,
					c.termCh)

				return &callID, &m.Data.Sdp, nil
			case *MsgCallRejected:
				return nil, nil, fmt.Errorf("Call rejected: %d", m.Data.RejectCode)
			default:
				return nil, nil, fmt.Errorf("Protocol error. Msg-type: %s", m.GetType())
			}
		case <-callCtx.Done():
			return nil, nil, fmt.Errorf("Timeout")
		}
	}

}

// Terminate the active call.
func (c *Call) Terminate(ctx context.Context) error {
	if len(c.callID) == 0 {
		return fmt.Errorf("no active call")
	}
	// send start call message
	if err := c.sepp.SendMsg(MsgCallTerminate{
		MsgBase: MsgBase{
			Type: MsgTypeCallTerminate,
			From: c.clientID,
			To:   c.confID,
		},
		Data: MsgCallTerminateData{
			CallID: string(c.callID)},
	}); err != nil {
		return fmt.Errorf("failed to send message: %s", err)
	}

	// wait for terminated
	select {
	case <-ctx.Done():
		return fmt.Errorf("timeout")
	case <-c.termCh:
	}

	return nil
}

// UpdateSDP sends and sdp update to the remote end.
func (c *Call) UpdateSDP(ctx context.Context, sdp Sdp) error {
	if len(c.callID) == 0 {
		return fmt.Errorf("no active call")
	}
	// send start call message
	if err := c.sepp.SendMsg(MsgSdpUpdate{
		MsgBase: MsgBase{
			Type: MsgTypeSdpUpdate,
			From: c.clientID,
			To:   c.confID,
		},
		Data: MsgSdpUpdateData{
			CallID: string(c.callID),
			Sdp:    sdp},
	}); err != nil {
		return fmt.Errorf("failed to send message: %s", err)
	}
	return nil
}

// TurnOffVideo mutes or unmute video
func (c *Call) TurnOffVideo(ctx context.Context, off bool) error {
	if len(c.callID) == 0 {
		return fmt.Errorf("no active call")
	}
	if err := c.sepp.SendMsg(MsgMuteVideo{
		MsgBase: MsgBase{
			Type: MsgTypeMuteVideo,
			From: c.clientID,
			To:   c.confID,
		},
		Data: MsgMuteVideoData{
			CallID: string(c.callID),
			On:     off},
	}); err != nil {
		return fmt.Errorf("failed to send message: %s", err)
	}
	return nil
}

// Close this call.
// Shuts down connection to the signaling service,
// but does _not_ terminate the call.
func (c *Call) Close() {
	if c.cancel != nil {
		c.cancel()
	}
	if c.sepp != nil {
		c.sepp.Stop()
	}
}
