package wstunnel

import (
	"io"
	"fmt"
	"log/slog"
	"net/http"
	"encoding/hex"

	"ws-vpn-go/common"

	"github.com/gorilla/websocket"
)

type Tunel = *websocket.Conn

type WsTunnel struct {

	serverInfo common.ServerInfo

	tunnelPath               string
	key                      string

	upgrader                 websocket.Upgrader

	receivedPackageCh        chan []byte
	clinetConnectionRegister ConnectionRegister

	closeConnectionHandler   func(common.IpAddress)

	logger                   *slog.Logger
}

func New(tunnelPath string, key string, serverInfo common.ServerInfo, logger *slog.Logger) *WsTunnel {
	tunnel := &WsTunnel{
		tunnelPath:               tunnelPath,
		key:                      key,
		serverInfo:               serverInfo,
		receivedPackageCh:        make(chan []byte, 256),
		clinetConnectionRegister: NewConnectionRegister(),
		logger:                   logger,
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: tunnel.checkBeforeUpgrade,
	}
	tunnel.upgrader = upgrader

	tunnel.closeConnectionHandler = func (ip common.IpAddress)  {
		tunnel.logger.Error(
			fmt.Sprintf("Use default connection close handler for ip %s", ip.String()))
	}

	return tunnel
}

func (tunnel *WsTunnel) checkClientAddress(request *http.Request) bool {
	clientAddress := request.Header.Get("ClientIP")
	ip := common.GetIpFromString(clientAddress)
	return !tunnel.clinetConnectionRegister.Contains(ip)
}

//TODO: добавить освобождение адреса на сервере
func (tunnel *WsTunnel) handleConnectionClosed(clientIp common.IpAddress) {
	tunnel.clinetConnectionRegister.Remove(clientIp)
	tunnel.closeConnectionHandler(clientIp)
}

func (tunnel *WsTunnel) connectionHandler(w http.ResponseWriter, r *http.Request) {

	sourceIp := r.RemoteAddr

	ws, err := tunnel.upgrader.Upgrade(w, r, nil)
	if err != nil {
		tunnel.logger.Error(fmt.Sprintf("Connection upgrade failed for %s; Error: %s", sourceIp, err))
		return
	}

	defer ws.Close()

	clientIP := common.GetIpFromString(r.Header.Get("ClientIp"))
	if clientIP == common.GetAllZeroIp() {
		tunnel.logger.Error(fmt.Sprintf("Unable to read target IP in request from: %s. Connection closed", sourceIp))
		return
	}

	if !tunnel.clinetConnectionRegister.Update(clientIP, ws) {
		tunnel.logger.Warn(fmt.Sprintf("Unable to setup connection from [%s]: client ip [%s] is not registered", sourceIp, clientIP))
		return
	}

	defer tunnel.handleConnectionClosed(clientIP)

	tunnel.logger.Info(fmt.Sprintf("Client %s connected from %s", clientIP.String(), sourceIp))

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				tunnel.logger.Warn(fmt.Sprintf("[%s] Unable to read from WS: %s", clientIP.String(), err))
			}
			break
		}

		tunnel.logger.Debug(fmt.Sprintf("[%s] Got message from WS connection:\n%s", clientIP.String(), hex.Dump(message)))

		tunnel.writeToChannel(message)
	}

	tunnel.logger.Info(fmt.Sprintf("[%s] Client from %s disconnected", clientIP.String(), sourceIp))
}

func (tunnel *WsTunnel) checkBeforeUpgrade(request *http.Request) bool {
	if common.CheckKey(request, tunnel.key) && tunnel.checkClientAddress(request) {
		tunnel.logger.Info(fmt.Sprintf("Try to open WS connection from: %s", request.RemoteAddr))
		return true
	} else {
		tunnel.logger.Warn(fmt.Sprintf("Try to open WS connection with wrong key [%s] or address [%s]; Connection form: %s",
			request.Header.Get("Key"),
			request.Header.Get("ClientIP"),
			request.RemoteAddr))
		return false
	}
}

func (tunnel *WsTunnel) RegisterHandlers(mux *http.ServeMux) error {
	mux.HandleFunc(tunnel.tunnelPath, tunnel.connectionHandler)
	return nil
}

func (tunnel *WsTunnel) SetConnectionCloseHandler(handler func (common.IpAddress)) {
	tunnel.closeConnectionHandler = handler
}

func (tunnel *WsTunnel) ReserveConnection(ip common.IpAddress) error {
	if (tunnel.clinetConnectionRegister.Contains(ip)) {
		return fmt.Errorf("connection for [%s] already exist", ip.String())
	}

	tunnel.clinetConnectionRegister.Add(ip, nil)
	return nil
}

func (tunnel *WsTunnel) writeToChannel(data []byte) {
	select {
	case tunnel.receivedPackageCh <- data:
	default:
		tunnel.logger.Error("[IF] Write channel full, dropping packet")
	}
}

func (tunnel *WsTunnel) Listen() error {
	return nil
}

func (tunnel *WsTunnel) WriteTo(target io.Writer) error {

	for packet := range tunnel.receivedPackageCh {
		_, err := target.Write(packet)
		if err != nil {
			tunnel.logger.Error(fmt.Sprintf("Write error: %s", err))
		}
		tunnel.logger.Debug(
			fmt.Sprintf(
				"Send to interface, packets remain in channale: %d",
				len(tunnel.receivedPackageCh)))
	}

	return fmt.Errorf("write channel closed, exiting writer loop")
}

func (tunnel *WsTunnel) WriteToTunnel(target common.IpAddress, packet []byte) error {

	client, ok := tunnel.clinetConnectionRegister.Get(target)

	if ok {

		if client == nil {
			return fmt.Errorf("addres [%s] registered, but connection is not exist", target.String())
		}

		err := client.WriteMessage(websocket.BinaryMessage, packet)
		if err != nil {
			return err
		}
		tunnel.logger.Debug(fmt.Sprintf("Server interface got package:\n%s", hex.Dump(packet)))
	} else {
		return fmt.Errorf("unknown target %s", target.String())
	}

	return nil
}