package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/eyeson-team/gosepp/v3"
)

func main() {
	authTokenFlag := flag.String("auth-token", "", "JWT token")
	clientIDFlag := flag.String("client-id", "", "Client-ID to use")
	confIDFlag := flag.String("conf-id", "", "Confserver-ID to connect to")
	flag.Parse()

	ci := &gosepp.CallInfo{
		SigEndpoint: "wss://sig.eyeson.com/call",
		AuthToken:   *authTokenFlag,
		ClientID:    *clientIDFlag,
		ConfID:      *confIDFlag,
	}

	call, err := gosepp.NewCall(ci, nil)
	if err != nil {
		log.Fatalf("failed: %s", err)
	}

	defer call.Close()

	call.SetSDPUpdateHandler(func(sdp gosepp.Sdp) {
		log.Printf("Sdp update with type %s sdp: %s\n", sdp.SdpType, sdp.Sdp)
	})

	call.SetTerminatedHandler(func() {
		log.Println("Call terminated")
	})

	callID, sdp, err := call.Start(context.Background(),
		gosepp.Sdp{SdpType: "offer", Sdp: "dummy-sdp"}, "[Guest] Bla")
	if err != nil {
		log.Fatalf("Call failed with: %s", err)
	}

	log.Printf("Call with id %s and sdp %s", *callID, sdp.Sdp)

	time.Sleep(10 * time.Second)

	log.Println("Terminating call")
	if err = call.Terminate(context.Background()); err != nil {
		log.Printf("Termination failed: %s\n", err)
	}
}
