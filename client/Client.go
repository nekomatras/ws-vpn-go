package client

import (
	"log"
	"fmt"
	"errors"
	"encoding/hex"
	"ws-vpn-go/common"

	"github.com/gorilla/websocket"
)



const DefaultMTU = 1500

type Client struct {

    remoteWebSocketURL string

	interfaceName string
	interfaceAddress string

	mtu uint
	isInited bool

	localInterface common.Interface
	wsTunnel common.Tunel
}

func NewClient(wsUrl string) *Client {
    return &Client{remoteWebSocketURL: wsUrl, mtu: DefaultMTU, isInited: false}
}

func (client *Client) RemoteWebSocketURL() string {
	return client.remoteWebSocketURL
}

func (client *Client) InterfaceName() string {
return client.interfaceName
}

func (client *Client) MTU() uint {
	return client.mtu
}

func (client *Client) SetInterfaceName(interfaceName string) {
	client.interfaceName = interfaceName
}

func (client *Client) SetInterfaceAddress(interfaceAddress string) {
	client.interfaceAddress = interfaceAddress
}

func (client *Client) Init() error {
	var err error

	client.localInterface, err = common.CreateInterface(client.interfaceName)
	if err != nil {
		return fmt.Errorf("Interface creation error: %w", err)
	}

	err = common.SetupInterface(client.localInterface, client.interfaceAddress)
	if err != nil {
		return fmt.Errorf("Interface setup error: %w", err)
	}

	client.wsTunnel, _, err = websocket.DefaultDialer.Dial(client.remoteWebSocketURL, nil)
	if err != nil {
		return fmt.Errorf("WebSocket connection error: %w", err)
	}

	log.Printf("WebSocket connected to remore: %s", client.remoteWebSocketURL)
	client.isInited = true

	return nil
}

func (client *Client) Run() error {

	if (!client.isInited) {
		return errors.New("Client must be initialized")
	}

	go client.inretfaceLoop()
	go client.tunelLoop()

	return nil
}

func (client *Client) inretfaceLoop() {

	buf := make([]byte, client.mtu)

	for {

		n, err := client.localInterface.Read(buf)
		if err != nil {
			log.Fatalf("Interface read error: %v", err)
		}

		log.Printf("Client interface got package:\n%s",hex.Dump(buf[:n]))

		err = client.wsTunnel.WriteMessage(websocket.TextMessage, buf[:n])
		if err != nil {
			log.Fatalf("WebSocket write error: %v", err)
		}
	}
}

func (client *Client) tunelLoop() {

	for {
		_, message, err := client.wsTunnel.ReadMessage()
		if err != nil {
			log.Fatalf("WebSocket read error: %v", err)
		}

		_, err = client.localInterface.Write(message)
		if err != nil {
			log.Fatalf("Interface write error: %v", err)
		}
	}

}

func (client *Client) closeWebSocketConnection() error {

	err := client.wsTunnel.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return fmt.Errorf("WebSocket error sending close messager: %w", err)
	}

	return nil
}