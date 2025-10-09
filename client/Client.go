package client

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"ws-vpn-go/common"

	"github.com/gorilla/websocket"
)

const DefaultMTU = 1500

type Client struct {
	remoteWebSocketURL string

	interfaceName    string
	interfaceAddress string

	key string

	mtu                 uint
	isInited            bool
	isConnectedToRemote bool

	localInterface common.Interface
	wsTunnel       common.Tunnel

	logger         *slog.Logger
}

func New(wsUrl string, logger *slog.Logger) *Client {
	logger.Info(fmt.Sprintf("Create client: Target URL: \"%s\"; MTU: %d", wsUrl, DefaultMTU))
	return &Client{
		remoteWebSocketURL:  wsUrl,
		mtu:                 DefaultMTU,
		isInited:            false,
		isConnectedToRemote: false,
		logger: logger}
}

















func (client *Client) SetKey(key string) {
	client.key = key
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
	client.logger.Info(fmt.Sprintf("Set interface name: %s", interfaceName))
	client.interfaceName = interfaceName
}

func (client *Client) SetInterfaceAddress(interfaceAddress string) {
	client.logger.Info(fmt.Sprintf("Set interface address: %s", interfaceAddress))
	client.interfaceAddress = interfaceAddress
}

func (client *Client) ConnectToRemote() error {
	var err error

	if client.isConnectedToRemote {
		return fmt.Errorf("already connected to remote web-socket")
	}

	header := make(http.Header)
	header.Add("ClientIP", client.interfaceAddress)
	header.Add("Key", client.key)

	client.wsTunnel, _, err = websocket.DefaultDialer.Dial(client.remoteWebSocketURL, header)
	if err != nil {
		return fmt.Errorf("web-socket connection error: %w", err)
	}

	client.logger.Info(fmt.Sprintf("WebSocket connected to remore: %s", client.remoteWebSocketURL))
	client.isConnectedToRemote = true

	return nil
}

func (client *Client) Init() error {
	var err error
	client.logger.Info("Setup interface...")
	client.localInterface, err = common.CreateInterface(client.interfaceName)
	if err != nil {
		return fmt.Errorf("interface creation error: %w", err)
	}

	err = common.SetupInterface(client.localInterface, client.interfaceAddress, client.mtu)
	if err != nil {
		return fmt.Errorf("interface setup error: %w", err)
	}

	client.isInited = true
	client.logger.Info(fmt.Sprintf("Create VPN interface %s", client.interfaceName))

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
			client.logger.Error(fmt.Sprintf("Interface read error: %v", err))
			os.Exit(-1)
		}

		client.logger.Debug(fmt.Sprintf("Client interface got package:\n%s", hex.Dump(buf[:n])))

		err = client.wsTunnel.WriteMessage(websocket.BinaryMessage, buf[:n])
		if err != nil {
			client.logger.Error(fmt.Sprintf("WebSocket write error: %v", err))
			os.Exit(-1)
		}
	}
}

func (client *Client) tunelLoop() {

	for {
		_, message, err := client.wsTunnel.ReadMessage()
		if err != nil {
			client.logger.Error(fmt.Sprintf("WebSocket read error: %v", err))
			os.Exit(-1)
		}

		_, err = client.localInterface.Write(message)
		if err != nil {
			client.logger.Error(fmt.Sprintf("Interface write error: %v", err))
			os.Exit(-1)
		}
	}

}

func (client *Client) closeWebSocketConnection() error {

	err := client.wsTunnel.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return fmt.Errorf("web-socket error sending close messager: %w", err)
	}

	return nil
}
