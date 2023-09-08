package main

import (
	"log"
	"time"

	"github.com/eyeson-team/gosepp/v3"
)

func main() {
	clientID := "client-id"
	confID := "conf-id"
	sdp := "dummy-sdp"
	jwtToken := "signed-token"

	sepp, err := gosepp.NewGoSepp("wss://sig.eyeson.com/call", jwtToken, nil, nil)
	if err != nil {
		log.Fatalf("failed: %s", err)
	}

	// wait for connected
	select {
	case connected, ok := <-sepp.ConnectStatusCh():
		if !ok || !connected {
			log.Fatalf("Failed to connect")
		}
	case <-time.After(2 * time.Second):
		log.Fatalf("Failed to connect")
	}

	if err := sepp.SendMsg(gosepp.MsgCallStart{
		MsgBase: gosepp.MsgBase{
			Type: gosepp.MsgTypeCallStart,
			From: clientID,
			To:   confID,
		},
		Data: gosepp.MsgCallStartData{
			Sdp:         gosepp.Sdp{SdpType: "offer", Sdp: sdp},
			DisplayName: clientID},
	}); err != nil {
		log.Fatalf("failed to send message:", err)
	}

	// wait for call accepted
	select {
	case msg, ok := <-sepp.RcvCh():
		if !ok {
			log.Fatalf("Failed to receive")
		}
		// dispatch messages
		switch m := msg.(type) {
		case *gosepp.MsgCallAccepted:
			log.Printf("Call Accepted: call-id: ", m.Data.CallID)
		default:
			log.Fatalf("Accept call failed")
		}
	case <-time.After(2 * time.Second):
		log.Fatalf("Waited too long for accept")
	}

}
