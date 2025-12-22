package wstunnel

import (
	"io"
	"fmt"
	"net/http"
	"log/slog"

	"strings"

	"ws-vpn-go/common"

	"github.com/gorilla/websocket"
)

const Repeat = 5

type WsTunnel struct {
	remoteURL string
	key string

	clientIp common.IpAddress

	wsTunnel *websocket.Conn

	logger *slog.Logger
}

func New(remoteAddress string, tunnelPath string, key string, logger *slog.Logger) *WsTunnel {
	return &WsTunnel{
		remoteURL: "wss://" + remoteAddress + tunnelPath,
		key: key,
		logger: logger,
	}
}

func (tunnel *WsTunnel) Run() error {
	return tunnel.tryConnectToRemote()
}

func (tunnel *WsTunnel) WriteTo(target io.Writer) error {
	for {
		_, message, err := tunnel.wsTunnel.ReadMessage()
		if err != nil {
			tunnel.logger.Error(fmt.Sprintf("WebSocket read error: %v", err))
			reconnectErr := tunnel.tryConnectToRemote()
			if reconnectErr != nil {
				return reconnectErr
			}
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
	var err error

	err = tunnel.wsTunnel.WriteMessage(websocket.BinaryMessage, packet)
		if err != nil {
			reconnectErr := tunnel.tryConnectToRemote()
			if reconnectErr == nil {
				err = tunnel.wsTunnel.WriteMessage(websocket.BinaryMessage, packet)
				if err == nil {
					return nil
				}
			}
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

	if ip == common.GetAllZeroIp() {
		return fmt.Errorf(
			"unable to set client ip address [%s]", ip.String())
	}

	tunnel.clientIp = ip
	return nil
}

func (tunnel *WsTunnel) SetConnectionCloseHandler(handler func (common.IpAddress)) {
	errMsg := "method SetConnectionCloseHandler not implemented"
	tunnel.logger.Error("Not implementod method called: " + errMsg)
}

func (tunnel *WsTunnel) tryConnectToRemote() error {
	var err error
	repeat := Repeat

	header := make(http.Header)
	header.Add("Key", tunnel.key)
	header.Add("ClientIP", tunnel.clientIp.String())

	wsPath := strings.Replace(tunnel.remoteURL, "https://", "", 1)

	for repeat != 0 {
		tunnel.wsTunnel, _, err = websocket.DefaultDialer.Dial(wsPath, header)
		if err == nil {
			return nil
		}

		repeat--
		err = fmt.Errorf("web-socket connection error: %w", err)
		tunnel.logger.Warn(fmt.Sprintf("[%s] Unadle to connect tunnel: %v; Repeat: %d", wsPath, err, repeat))
	}

	return err
}