package client



import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"ws-vpn-go/common"

	"github.com/gorilla/websocket"
)

const DefaultMTU = 1500

type Client struct {
	remoteWebSocketURL string

	interfaceName    string
	interfaceAddress string

	mtu                 uint
	isInited            bool
	isConnectedToRemote bool

	localInterface common.Interface
	wsTunnel       common.Tunel
}

func New(wsUrl string) *Client {
	log.Printf("Create client: Target URL: \"%s\"; MTU: %d", wsUrl, DefaultMTU)
	return &Client{
		remoteWebSocketURL:  wsUrl,
		mtu:                 DefaultMTU,
		isInited:            false,
		isConnectedToRemote: false}
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

func (client *Client) IsInited() bool {
	return client.isInited
}

func (client *Client) IsConnectedToRemote() bool {
	return client.isConnectedToRemote
}

func (client *Client) SetInterfaceName(interfaceName string) {
	log.Printf("Set interface name: %s", interfaceName)
	client.interfaceName = interfaceName
}

func (client *Client) SetInterfaceAddress(interfaceAddress string) {
	log.Printf("Set interface address: %s", interfaceAddress)
	client.interfaceAddress = interfaceAddress
}

func (client *Client) ConnectToRemote() error {
	var err error

	if client.isConnectedToRemote {
		return fmt.Errorf("Already connected to remote WebSocket")
	}

	header := make(http.Header)
	header.Add("ClientIP", client.interfaceAddress)

	client.wsTunnel, _, err = websocket.DefaultDialer.Dial(client.remoteWebSocketURL, header)
	if err != nil {
		return fmt.Errorf("WebSocket connection error: %w", err)
	}

	log.Printf("WebSocket connected to remore: %s", client.remoteWebSocketURL)
	client.isConnectedToRemote = true

	return nil
}

func (client *Client) Init() error {
	var err error
	log.Println("Setup interface...")
	client.localInterface, err = common.CreateInterface(client.interfaceName)
	if err != nil {
		return fmt.Errorf("Interface creation error: %w", err)
	}

	err = common.SetupInterface(client.localInterface, client.interfaceAddress)
	if err != nil {
		return fmt.Errorf("Interface setup error: %w", err)
	}

	client.isInited = true
	log.Printf("Create VPN interface %s", client.interfaceName)

	return nil
}

func (client *Client) Run() error {

	if !client.isInited {
		return errors.New("Client must be initialized")
	}

	if !client.isConnectedToRemote {
		return errors.New("Client must be connected to remote")
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

		log.Printf("Client interface got package:\n%s", hex.Dump(buf[:n]))

		err = client.wsTunnel.WriteMessage(websocket.BinaryMessage, buf[:n])
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
