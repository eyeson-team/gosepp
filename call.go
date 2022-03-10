package gosepp

import (
	"context"
	"fmt"
)

// CallID custom callID type
type CallID string

// Call is an abstraction of the gosepp messaging based interface.
type Call struct {
	sepp               *GoSepp
	confID             string
	clientID           string
	callID             CallID
	terminationHandler func()
	sdpUpdateHandler   func(Sdp)
	cancel             context.CancelFunc
	termCh             chan bool
}

// NewCall initializes an instant of a call.
func NewCall(callInfo CallInfoInterface) (*Call, error) {

	sepp, err := NewGoSepp(callInfo.GetSigEndpoint(), callInfo.GetAuthToken())
	if err != nil {
		return nil, err
	}
	return &Call{sepp: sepp,
		confID:   callInfo.GetConfID(),
		clientID: callInfo.GetClientID(),
		termCh:   make(chan bool),
	}, nil
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

func startDispatch(ctx context.Context, sepp *GoSepp,
	termHandler func(), sdpUpdateHandler func(Sdp), termCh chan<- bool) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-sepp.RcvCh():
			if !ok {
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
			DisplayName: displayname},
	}); err != nil {
		return nil, nil, fmt.Errorf("failed to send message: %s", err)
	}

	// wait for call accepted or rejected
	select {
	case msg, ok := <-c.sepp.RcvCh():
		if !ok {
			return nil, nil, fmt.Errorf("Failed to receive")
		}
		// dispatch messages
		switch m := msg.(type) {
		case *MsgCallAccepted:
			callID := CallID(m.Data.CallID)
			c.callID = callID
			// start dispatcher as goroutine
			go startDispatch(callCtx, c.sepp, c.terminationHandler,
				c.sdpUpdateHandler, c.termCh)

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
