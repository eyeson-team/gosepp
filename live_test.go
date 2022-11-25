package gosepp

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func TestSeppLive(t *testing.T) {
	clientID := "61f90c61d4319200113fbd61"
	confID := "61f90c42d4319200113fbd5f"

	liveToken := os.Getenv("LIVE_TOKEN")
	if len(liveToken) == 0 {
		t.Skip("Skip cause env LIVE_TOKEN not set")
		return
	}

	seppServerAddress := "sig.eyeson.com"
	sepp, err := NewGoSepp(fmt.Sprintf("wss://%s/call", seppServerAddress),
		liveToken, nil, nil)
	if err != nil {
		t.Fatalf("failed: %s", err)
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

	if err := sepp.SendMsg(MsgCallStart{
		MsgBase: MsgBase{
			Type: MsgTypeCallStart,
			From: clientID,
			To:   confID,
		},
		Data: MsgCallStartData{
			Sdp:         Sdp{SdpType: "offer", Sdp: ""},
			DisplayName: clientID},
	}); err != nil {
		fmt.Println("failed to send message:", err)
		return
	}

	fmt.Println("Message sent")

	select {
	case msg, ok := <-sepp.RcvCh():
		fmt.Println("msg:", msg)
		if !ok {
			fmt.Println("Failed to recv")
		}
	}

	time.Sleep(10 * time.Second)

}
