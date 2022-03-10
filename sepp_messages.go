package gosepp

// Messages types
const (
	MsgTypeCallStart      string = "call_start"
	MsgTypeCallRejected   string = "call_rejected"
	MsgTypeCallAccepted   string = "call_accepted"
	MsgTypeSdpUpdate      string = "sdp_update"
	MsgTypeCallTerminate  string = "call_terminate"
	MsgTypeCallTerminated string = "call_terminated"
)

// SeppMsgTypes defines a mapping of message types
// and an interface function which create a messages
// adhering to the MsgInterface.
var SeppMsgTypes = map[string]func() MsgInterface{
	MsgTypeCallStart:      func() MsgInterface { return &MsgCallStart{} },
	MsgTypeCallRejected:   func() MsgInterface { return &MsgCallRejected{} },
	MsgTypeCallAccepted:   func() MsgInterface { return &MsgCallAccepted{} },
	MsgTypeSdpUpdate:      func() MsgInterface { return &MsgSdpUpdate{} },
	MsgTypeCallTerminate:  func() MsgInterface { return &MsgCallTerminate{} },
	MsgTypeCallTerminated: func() MsgInterface { return &MsgCallTerminated{} },
}

// MsgInterface define a messages which allows to get and modify
// the base-message. This helps to dispatch matches without
// having to deserialize the whole message.
type MsgInterface interface {
	GetMsgID() string
	GetType() string
	GetFrom() string
	GetTo() string
	SetFrom(string)
	SetTo(string)
}

// MsgBase base struct for all conf messages.
type MsgBase struct {
	Type  string `json:"type"`
	MsgID string `json:"msg_id"`
	From  string `json:"from"`
	To    string `json:"to"`
}

// GetMsgID get the message-id of a conf message.
func (msg *MsgBase) GetMsgID() string {
	return msg.MsgID
}

// GetType get the message-type of a conf message.
func (msg *MsgBase) GetType() string {
	return msg.Type
}

// GetTo retrieves the message to header.
func (msg *MsgBase) GetTo() string {
	return msg.To
}

// SetTo allows to set the message base to header.
func (msg *MsgBase) SetTo(to string) {
	msg.To = to
}

// GetFrom retrieves the from header.
func (msg *MsgBase) GetFrom() string {
	return msg.From
}

// SetFrom allows to set the from header of that message.
func (msg *MsgBase) SetFrom(from string) {
	msg.From = from
}

// Sdp combines the actual sdp with an type.
// The type can be either "offer" or "answer".
type Sdp struct {
	SdpType string `json:"type"`
	Sdp     string `json:"sdp"`
}

// MsgCallStartData carries data of for the call_start message.
type MsgCallStartData struct {
	Sdp         Sdp    `json:"sdp"`
	DisplayName string `json:"display_name"`
}

// MsgCallStart message
type MsgCallStart struct {
	MsgBase
	Data MsgCallStartData `json:"data"`
}

// MsgCallRejectedData data
type MsgCallRejectedData struct {
	RejectCode int `json:"reject_code"`
}

// MsgCallRejected message
type MsgCallRejected struct {
	MsgBase
	Data MsgCallRejectedData `json:"data"`
}

// MsgCallAcceptedData data
type MsgCallAcceptedData struct {
	CallID string `json:"call_id"`
	Sdp    Sdp    `json:"sdp"`
}

// MsgCallAccepted message
type MsgCallAccepted struct {
	MsgBase
	Data MsgCallAcceptedData `json:"data"`
}

// MsgSdpUpdateData data
type MsgSdpUpdateData struct {
	CallID string `json:"call_id"`
	Sdp    Sdp    `json:"sdp"`
}

// MsgSdpUpdate message
type MsgSdpUpdate struct {
	MsgBase
	Data MsgSdpUpdateData `json:"data"`
}

// MsgCallTerminateData data
type MsgCallTerminateData struct {
	CallID   string `json:"call_id"`
	TermCode int    `json:"term_code"`
}

// MsgCallTerminate message
type MsgCallTerminate struct {
	MsgBase
	Data MsgCallTerminateData `json:"data"`
}

// MsgCallTerminatedData data
type MsgCallTerminatedData struct {
	CallID   string `json:"call_id"`
	TermCode int    `json:"term_code"`
}

// MsgCallTerminated message
type MsgCallTerminated struct {
	MsgBase
	Data MsgCallTerminatedData `json:"data"`
}
