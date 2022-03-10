package gosepp

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const SeppEndpoint string = "wss://sig.eyeson.com/call"

// GoSepp Confserver signaling.
type GoSepp struct {
	wsURL             *url.URL
	wsClient          *websocket.Conn
	run               bool
	rcvCh             chan MsgInterface
	wsDialer          *websocket.Dialer
	senderWaitGroup   sync.WaitGroup
	receiverWaitGroup sync.WaitGroup
	sendCh            chan []byte
	connectStatusCh   chan bool
	receiverCtxCancel context.CancelFunc
	authToken         string
}

// NewGoSepp returns a new GoSepp client.
func NewGoSepp(baseURL, authToken string) (*GoSepp, error) {
	return NewGoSeppCustomCerts(baseURL, "", "", "", false, true, authToken)
}

// NewGoSeppCustomCerts returns a new GoSepp client using custom certs.
func NewGoSeppCustomCerts(baseURL string, certFile, keyFile, caFile string,
	insecure bool, useSystemCAPool bool, authToken string) (*GoSepp, error) {
	var tlsConfig *tls.Config
	var err error
	if len(certFile) > 0 && len(keyFile) > 0 {
		tlsConfig, err = createTLSConfig(certFile, keyFile, caFile, useSystemCAPool)
		if err != nil {
			return nil, err
		}
	} else {
		tlsConfig = &tls.Config{}
	}

	if insecure {
		tlsConfig.InsecureSkipVerify = insecure
	}
	d := websocket.Dialer{TLSClientConfig: tlsConfig}
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	receiverCtx, receiverCancel := context.WithCancel(context.Background())
	rtm := &GoSepp{
		wsURL:             parsedURL,
		rcvCh:             make(chan MsgInterface, 1),
		wsDialer:          &d,
		sendCh:            make(chan []byte, 1),
		connectStatusCh:   make(chan bool, 1),
		receiverCtxCancel: receiverCancel,
		run:               true,
		authToken:         authToken}

	rtm.start(receiverCtx)
	rtm.sender()
	return rtm, nil
}

func createTLSConfig(certFile, keyFile, caFile string, useSystemCAPool bool) (*tls.Config, error) {
	// load cert, key, and CA-file
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	var caCertPool *x509.CertPool
	if useSystemCAPool {
		caCertPool, err = x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
	} else {
		if len(caFile) > 0 {
			// Load CA cert
			caCert, err := ioutil.ReadFile(caFile)
			if err != nil {
				return nil, err
			}
			caCertPool = x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				fmt.Println("Failed to append CAcert")
			}
		}
	}

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

// RcvCh get the channel where message adhering to the ConfMsgInterface
// can be retrieved.
func (rtm *GoSepp) RcvCh() chan MsgInterface {
	return rtm.rcvCh
}

// ConnectStatusCh allow to monitor the websockets connection status.
func (rtm *GoSepp) ConnectStatusCh() chan bool {
	return rtm.connectStatusCh
}

func (rtm *GoSepp) connect(parentCtx context.Context) error {
	ctx, cancel := context.WithTimeout(parentCtx, 8*time.Second)
	defer cancel()

	requestHeader := make(http.Header)
	if len(rtm.authToken) > 0 {
		requestHeader.Add("Authorization", fmt.Sprintf("Bearer %s", rtm.authToken))
	}
	c, _, err := rtm.wsDialer.DialContext(ctx, rtm.wsURL.String(), requestHeader)
	if err == nil {
		rtm.wsClient = c
	}
	return err
}

// Stop the internal messaging loop.
func (rtm *GoSepp) Stop() {

	// 1. stop receive-path
	rtm.run = false
	if wsClient := rtm.wsClient; wsClient != nil {
		wsClient.Close()
	}

	// cancel receiver-ctx. So any possible running connect
	// will return.
	rtm.receiverCtxCancel()
	rtm.receiverWaitGroup.Wait()
	// receiver is done now. So it's save to close the rcvCh
	close(rtm.rcvCh)
	close(rtm.connectStatusCh)

	close(rtm.sendCh)
	rtm.senderWaitGroup.Wait()
}

// SendMsg sends a message over the underlying websocket.
// In order to support concurrent writes, messages
// are send through an internal channel.
// Therefore messages are not sent immediately down
// the wire.
func (rtm *GoSepp) SendMsg(msg interface{}) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if rtm.run {
		rtm.sendCh <- b
	} else {
		return fmt.Errorf("Not running")
	}
	return nil

}

func (rtm *GoSepp) sender() {
	rtm.senderWaitGroup.Add(1)
	go func() {
		defer rtm.senderWaitGroup.Done()
		for {
			pingInterval := time.After(3 * time.Second)
			select {
			case <-pingInterval:
				if wsClient := rtm.wsClient; wsClient != nil {
					err := wsClient.WriteMessage(websocket.PingMessage, []byte("keepalive"))
					if err != nil {
						log.Println("failed to send ping.")
					}
				}
			case msg, ok := <-rtm.sendCh:
				if !ok {
					// exit sender
					return
				}
				if wsClient := rtm.wsClient; wsClient != nil {
					err := wsClient.WriteMessage(websocket.TextMessage, msg)
					if err != nil {
						log.Println("failed to send.")
					} else {
						//log.Printf("Sent message: %s\n", msg)
					}
				}
			}
		}
	}()
}

func (rtm *GoSepp) start(ctx context.Context) {
	rtm.receiverWaitGroup.Add(1)

	go func() {
		defer rtm.receiverWaitGroup.Done()
		for rtm.run == true {
			// try to connect
			err := rtm.connect(ctx)
			if err != nil {
				log.Printf("Failed to connect to %s [%s]. Retrying.", rtm.wsURL, err)
				rtm.connectStatusCh <- false
				if rtm.run {
					time.Sleep(2 * time.Second)
				}
				continue
			}
			rtm.connectStatusCh <- true

			// start recv and send loop
			for {
				messageType, message, err := rtm.wsClient.ReadMessage()
				if err != nil {
					log.Println("read failed with:", err)
					break
				}

				//log.Println("Received:", string(message))

				if messageType == websocket.TextMessage {
					// parse
					var msgBase MsgBase
					err := json.Unmarshal(message, &msgBase)
					if err != nil {
						log.Printf("Failed to unmarshal [%s].\n", err)
						continue
					}
					msgInitFunc, ok := SeppMsgTypes[msgBase.Type]
					if !ok {
						log.Printf("Message-type %s not supported.", msgBase.Type)
						continue
					}
					interf := msgInitFunc()
					err = json.Unmarshal(message, interf)
					if err != nil {
						log.Println("Failed to unmarshal.")
						continue
					}
					rtm.rcvCh <- interf
				}

			}

		}
	}()
}
