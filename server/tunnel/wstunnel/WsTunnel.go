package wstunnel

import (
	"io"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"encoding/hex"

	"ws-vpn-go/common"

	"github.com/gorilla/websocket"
)

type Tunel = *websocket.Conn

type WsTunnel struct {

	serverInfo common.ServerInfo

	localWebSocketURL string
	key string

	wsTunnel       Tunel
	upgrader websocket.Upgrader

	tunnelError error

	receivedPackageCh       chan []byte
	clinetConnectionRegister ConnectionRegister

	logger *slog.Logger
}

func New(wsUrl string, key string, serverInfo common.ServerInfo, logger *slog.Logger) *WsTunnel {
	tunnel := &WsTunnel{
		localWebSocketURL:        wsUrl,
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

	return tunnel
}

func (tunnel *WsTunnel) checkKey(request *http.Request) bool {
	key := request.Header.Get("Key")
	if key == tunnel.key {
		return true
	} else {
		return false
	}
}

func (tunnel *WsTunnel) checkClientAddress(request *http.Request) bool {
	clientAddress := request.Header.Get("ClientIP")
	ip := common.GetIpFromString(clientAddress)
	return !tunnel.clinetConnectionRegister.Contains(ip)
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

	tunnel.clinetConnectionRegister.Add(clientIP, ws)
	defer tunnel.clinetConnectionRegister.Remove(clientIP)

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

func (tunnel *WsTunnel) infoHandler(w http.ResponseWriter, r *http.Request) {
	if tunnel.checkKey(r) {
		tunnel.serverInfo.WriteToResponse(w)
		w.WriteHeader(http.StatusAccepted)
	} else {
		tunnel.defaultHandler(w, r)
	}
}

func (tunnel *WsTunnel) defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "САДОВОД-ЛЮБИТЕЛЬ")
	w.WriteHeader(http.StatusOK)
}

func (tunnel *WsTunnel) checkBeforeUpgrade(request *http.Request) bool {
	if tunnel.checkKey(request) && tunnel.checkClientAddress(request) {
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

func (tunnel *WsTunnel) listenWebSocket() error {
	var err error

	u, err := url.Parse(tunnel.localWebSocketURL)
	if err != nil {
		return fmt.Errorf("Unable to parse WS URL: %w", err)
	}

	http.HandleFunc(u.Path, tunnel.connectionHandler)
	http.HandleFunc("/info", tunnel.infoHandler)
	http.HandleFunc("/", tunnel.defaultHandler) //добавить проброс из вне

	err = http.ListenAndServe(u.Host, nil)
	if err != nil {
		return fmt.Errorf("Unable to listen WS connections: %w", err)
	}

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
	tunnel.tunnelError = tunnel.listenWebSocket()
	if tunnel.tunnelError != nil {
		tunnel.logger.Error(fmt.Sprintf("Error while WS listen: %v", tunnel.tunnelError))
	}

	return tunnel.tunnelError
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

	return fmt.Errorf("Write channel closed, exiting writer loop")
}

func (tunnel *WsTunnel) WriteToTunnel(target common.IpAddress, packet []byte) error {

	client, ok := tunnel.clinetConnectionRegister.Get(target)

	if ok {
		err := client.WriteMessage(websocket.BinaryMessage, packet)
		if err != nil {
			return err
		}
		tunnel.logger.Debug(fmt.Sprintf("Server interface got package:\n%s", hex.Dump(packet)))
	} else {
		return fmt.Errorf("Unknown target %s", target.String())
	}

	return nil
}