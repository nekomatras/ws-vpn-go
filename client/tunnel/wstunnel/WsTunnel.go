package wstunnel

import (
	"io"
	"fmt"
	"net/http"
	"log/slog"

	"ws-vpn-go/common"

	"github.com/gorilla/websocket"
)

type WsTunnel struct {
	remoteURL string
	key string

	clientIp common.IpAddress

	wsTunnel *websocket.Conn

	logger *slog.Logger
}

func New(remoteAddress string, tunnelPath string, key string, logger *slog.Logger) *WsTunnel {
	return &WsTunnel{
		remoteURL: "ws://" + remoteAddress + tunnelPath,
		key: key,
		logger: logger,
	}
}

func (tunnel *WsTunnel) Run() error {
	var err error

	header := make(http.Header)
	header.Add("Key", tunnel.key)
	header.Add("ClientIP", tunnel.clientIp.String())

	tunnel.wsTunnel, _, err = websocket.DefaultDialer.Dial(tunnel.remoteURL, header)
	if err != nil {
		return fmt.Errorf("web-socket connection error: %w", err)
	}

	return nil
}

func (tunnel *WsTunnel) WriteTo(target io.Writer) error {
	for {
		_, message, err := tunnel.wsTunnel.ReadMessage()
		if err != nil {
			tunnel.logger.Error(fmt.Sprintf("WebSocket read error: %v", err))
			continue
		}

		_, err = target.Write(message)
		if err != nil {
			tunnel.logger.Error(fmt.Sprintf("Interface write error: %v", err))
			break
		}
	}

	return fmt.Errorf("write channel closed, exiting writer loop")
}

func (tunnel *WsTunnel) WriteToTunnel(_ common.IpAddress, packet []byte) error {
	err := tunnel.wsTunnel.WriteMessage(websocket.BinaryMessage, packet)
		if err != nil {
			return fmt.Errorf("web-socket write error: %v", err)
		}

	return nil
}

func (tunnel *WsTunnel) RegisterHandlers(mux *http.ServeMux) error {
	errMsg := "method RegisterHandlers not implemented"
	tunnel.logger.Error("Not implementod method called: " + errMsg)
	return fmt.Errorf("%s", errMsg)
}

func (tunnel *WsTunnel) ReserveConnection(ip common.IpAddress) error {

	if tunnel.clientIp != common.GetAllZeroIp() {
		return fmt.Errorf(
			"unable to set client ip address [%s], its already seted to %s", ip.String(), tunnel.clientIp.String())
	}

	tunnel.clientIp = ip
	return nil
}

func (tunnel *WsTunnel) SetConnectionCloseHandler(handler func (common.IpAddress)) {
	errMsg := "method SetConnectionCloseHandler not implemented"
	tunnel.logger.Error("Not implementod method called: " + errMsg)
}