package gosepp

// Messages types
const (
	MsgTypeCallStart        string = "call_start"
	MsgTypeCallRejected     string = "call_rejected"
	MsgTypeCallAccepted     string = "call_accepted"
	MsgTypeSdpUpdate        string = "sdp_update"
	MsgTypeCallTerminate    string = "call_terminate"
	MsgTypeCallTerminated   string = "call_terminated"
	MsgTypeCallResume       string = "call_resume"
	MsgTypeCallResumed      string = "call_resumed"
	MsgTypeChat             string = "chat"
	MsgTypeSetPresenter     string = "set_presenter"
	MsgTypeDesktopstreaming string = "desktopstreaming"
	MsgTypeMuteVideo        string = "mute_video"
	MsgTypeSourceUpdate     string = "source_update"
	MsgTypeMemberlist       string = "memberlist"
	MsgTypeRecording        string = "recording"
)

// SeppMsgTypes defines a mapping of message types
// and an interface function which create a messages
// adhering to the MsgInterface.
var SeppMsgTypes = map[string]func() MsgInterface{
	MsgTypeCallStart:        func() MsgInterface { return &MsgCallStart{} },
	MsgTypeCallRejected:     func() MsgInterface { return &MsgCallRejected{} },
	MsgTypeCallAccepted:     func() MsgInterface { return &MsgCallAccepted{} },
	MsgTypeSdpUpdate:        func() MsgInterface { return &MsgSdpUpdate{} },
	MsgTypeCallTerminate:    func() MsgInterface { return &MsgCallTerminate{} },
	MsgTypeCallTerminated:   func() MsgInterface { return &MsgCallTerminated{} },
	MsgTypeCallResume:       func() MsgInterface { return &MsgCallResume{} },
	MsgTypeCallResumed:      func() MsgInterface { return &MsgCallResumed{} },
	MsgTypeChat:             func() MsgInterface { return &MsgChat{} },
	MsgTypeSetPresenter:     func() MsgInterface { return &MsgSetPresenter{} },
	MsgTypeDesktopstreaming: func() MsgInterface { return &MsgDesktopstreaming{} },
	MsgTypeMuteVideo:        func() MsgInterface { return &MsgMuteVideo{} },
	MsgTypeSourceUpdate:     func() MsgInterface { return &MsgSourceUpdate{} },
	MsgTypeMemberlist:       func() MsgInterface { return &MsgMemberlist{} },
	MsgTypeRecording:        func() MsgInterface { return &MsgRecording{} },
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

// MsgCallResumeData carries data for the call_resume message.
type MsgCallResumeData struct {
	Sdp    Sdp    `json:"sdp"`
	CallID string `json:"call_id"`
}

// MsgCallResume message
type MsgCallResume struct {
	MsgBase
	Data MsgCallResumeData `json:"data"`
}

// MsgCallResumedData data
type MsgCallResumedData struct {
	CallID string `json:"call_id"`
	Sdp    Sdp    `json:"sdp"`
}

// MsgCallResumed message
type MsgCallResumed struct {
	MsgBase
	Data MsgCallResumedData `json:"data"`
}

// MsgChatData data
type MsgChatData struct {
	CallID    string `json:"call_id"`
	ClientID  string `json:"cid"`
	Content   string `json:"content"`
	ID        string `json:"id"`
	Timestamp string `json:"ts"`
}

// MsgChat chat message
type MsgChat struct {
	MsgBase
	Data MsgChatData `json:"data"`
}

// MsgSetPresenterData data
type MsgSetPresenterData struct {
	CallID   string `json:"call_id"`
	On       bool   `json:"on"`
	ClientID string `json:"cid"`
}

// MsgSetPresenter message
type MsgSetPresenter struct {
	MsgBase
	Data MsgSetPresenterData `json:"data"`
}

// MsgDesktopstreamingData data
type MsgDesktopstreamingData struct {
	CallID   string `json:"call_id"`
	On       bool   `json:"on"`
	ClientID string `json:"cid"`
}

// MsgDesktopstreaming message
type MsgDesktopstreaming struct {
	MsgBase
	Data MsgDesktopstreamingData `json:"data"`
}

// MsgMuteVideoData data
type MsgMuteVideoData struct {
	CallID   string `json:"call_id"`
	On       bool   `json:"on"`
	ClientID string `json:"cid"`
}

// MsgMuteVideo message
type MsgMuteVideo struct {
	MsgBase
	Data MsgMuteVideoData `json:"data"`
}

// Dimension specifying position on podium
type Dimension struct {
	Width  int `json:"w"`
	Height int `json:"h"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

// MsgSourceUpdateData holds data for the podium configuration
type MsgSourceUpdateData struct {
	CallID             string      `json:"call_id"`
	AudioSources       []int       `json:"asrc"`
	VideoSources       []int       `json:"vsrc"`
	Broadcast          *bool       `json:"bcast,omitempty"`
	Dimensions         []Dimension `json:"dims"`
	Layout             int         `json:"l"`
	Sources            []string    `json:"src"`
	TextOverlay        *bool       `json:"tovl,omitempty"`
	PresenterSrc       *int        `json:"psrc,omitempty"`
	DesktopstreamerSrc *int        `json:"dsrc,omitempty"`
}

// MsgSourceUpdate message
type MsgSourceUpdate struct {
	MsgBase
	Data MsgSourceUpdateData `json:"data"`
}

// MsgRecordingData recording status stuff
type MsgRecordingData struct {
	CallID  string `json:"call_id"`
	Active  bool   `json:"active"`
	Enabled bool   `json:"enabled"`
}

// MsgRecording message
type MsgRecording struct {
	MsgBase
	Data MsgRecordingData `json:"data"`
}

// Member participant on memberlist
type Member struct {
	ClientID string  `json:"cid"`
	Platform *string `json:"p,omitempty"`
}

// Media media on memberlist
type Media struct {
	MediaID string `json:"mid"`
	PlayID  string `json:"playid"`
}

// MsgMemberlistData memberlist data
type MsgMemberlistData struct {
	CallID string   `json:"call_id"`
	Count  int      `json:"count"`
	Add    []Member `json:"add"`
	Del    []string `json:"del"`
	Media  []Media  `json:"media"`
}

// MsgMemberlist message
type MsgMemberlist struct {
	MsgBase
	Data MsgMemberlistData `json:"data"`
}
