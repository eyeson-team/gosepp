package gosepp

// CallInfoInterface defines a configuration interface,
// to which the init struct of NewCall must comply.
type CallInfoInterface interface {
	GetSigEndpoint() string
	GetAuthToken() string
	GetClientID() string
	GetConfID() string
}

// CallInfo is the default implementation of the
// CallInfoInterface.
type CallInfo struct {
	SigEndpoint string
	AuthToken   string
	ClientID    string
	ConfID      string
}

// GetSigEndpoint returns the sip-sepp endpoint.
func (i *CallInfo) GetSigEndpoint() string {
	return i.SigEndpoint
}

// GetAuthToken returns the jwt-auth token
// used as bearer authorization token.
func (i *CallInfo) GetAuthToken() string {
	return i.AuthToken
}

// GetClientID returns the clientID which
// is the initiator of this call.
func (i *CallInfo) GetClientID() string {
	return i.ClientID
}

// GetConfID returns the confID which
// is the destination of this call.
func (i *CallInfo) GetConfID() string {
	return i.ConfID
}
