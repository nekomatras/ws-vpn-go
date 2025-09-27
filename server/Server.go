package server

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"ws-vpn-go/common"

	"github.com/gorilla/websocket"
)

const DefaultMTU = 1500

type Server struct {
	localWebSocketURL string

	interfaceName    string
	interfaceAddress string

	mtu      uint
	isInited bool

	localInterface common.Interface
	wsTunnel       common.Tunel

	writeToInterfaceCh       chan []byte
	clinetConnectionRegister map[common.IpAddress]*websocket.Conn

	logger         *slog.Logger
	upgrader       websocket.Upgrader
}

func New(wsUrl string, logger *slog.Logger) *Server {

	logger.Info(fmt.Sprintf("Create server: WebSocket URL: \"%s\"; MTU: %d", wsUrl, DefaultMTU))

	server:= &Server{
		localWebSocketURL:        wsUrl,
		mtu:                      DefaultMTU,
		isInited:                 false,
		writeToInterfaceCh:       make(chan []byte, 256),
		clinetConnectionRegister: make(map[common.IpAddress]*websocket.Conn),
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: server.checkBeforeUpgrade,
	}

	server.upgrader = upgrader

	return server
}

func (server *Server) LocalWebSocketURL() string {
	return server.localWebSocketURL
}

func (server *Server) InterfaceName() string {
	return server.interfaceName
}

func (server *Server) MTU() uint {
	return server.mtu
}

func (server *Server) IsInited() bool {
	return server.isInited
}

func (server *Server) SetInterfaceName(interfaceName string) {
	server.logger.Info(fmt.Sprintf("Set interface name: %s", interfaceName))
	server.interfaceName = interfaceName
}

func (server *Server) SetInterfaceAddress(interfaceAddress string) {
	server.logger.Info(fmt.Sprintf("Set interface address: %s", interfaceAddress))
	server.interfaceAddress = interfaceAddress
}

func (server *Server) checkBeforeUpgrade(request *http.Request) bool {
	server.logger.Info(fmt.Sprintf("Try to open WS connection to: %s", request.RemoteAddr))
	return true
}

func (server *Server) connectionHandler(w http.ResponseWriter, r *http.Request) {

	sourceIp := r.RemoteAddr

	ws, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		server.logger.Error(fmt.Sprintf("Connection upgrade failed for %s; Error: %s", sourceIp, err))
		return
	}

	defer ws.Close()

	clientIP := common.GetIpFromString(r.Header.Get("ClientIp"))
	if clientIP == common.GetAllZeroIp() {
		server.logger.Error(fmt.Sprintf("Unable to read target IP in request from: %s. Connection closed", sourceIp))
		return
	}

	server.clinetConnectionRegister[clientIP] = ws
	defer delete(server.clinetConnectionRegister, clientIP)

	server.logger.Info(fmt.Sprintf("Client %s connected from %s", clientIP.String(), sourceIp))

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				server.logger.Warn(fmt.Sprintf("[%s] Unable to read from WS: %s", clientIP.String(), err))
			}
			break
		}

		server.logger.Debug(fmt.Sprintf("[%s] Got message from WS connection:\n%s", clientIP.String(), hex.Dump(message)))

		server.writeToInterface(message)
	}

	server.logger.Info(fmt.Sprintf("[%s] Client from %s disconnected", clientIP.String(), sourceIp))
}

func (server *Server) listenWebSocket() error {
	var err error

	u, err := url.Parse(server.localWebSocketURL)
	if err != nil {
		server.logger.Error(err.Error())
		os.Exit(-1)
	}

	http.HandleFunc(u.Path, server.connectionHandler)

	err = http.ListenAndServe(u.Host, nil)
	if err != nil {
		return fmt.Errorf("ListenAndServe error: %w", err)
	}

	server.logger.Info(fmt.Sprintf("Server listening on:%s", server.localWebSocketURL))

	return nil
}

func (server *Server) Init() error {
	var err error

	server.logger.Info("Setup interface...")
	server.localInterface, err = common.CreateInterface(server.interfaceName)
	if err != nil {
		return fmt.Errorf("Interface creation error: %w", err)
	}

	err = common.SetupInterface(server.localInterface, server.interfaceAddress)
	if err != nil {
		return fmt.Errorf("Interface setup error: %w", err)
	}

	server.isInited = true
	server.logger.Info(fmt.Sprintf("Create VPN interface %s", server.interfaceName))

	go func() {
		err = server.listenWebSocket()
		if err != nil {
			server.logger.Error(fmt.Sprintf("Error while WS listen: %w", err))
			os.Exit(-1)
		}
	}()

	go server.writerToInterfaceLoop()
	go server.readerFromInterfaceLoop()

	return nil
}

func (server *Server) writerToInterfaceLoop() {
	for {
		select {
		case packet, ok := <-server.writeToInterfaceCh:
			if !ok { // канал закрыт
				server.logger.Error("[IF] Write channel closed, exiting writer loop")
				os.Exit(-1)
				return
			}
			_, err := server.localInterface.Write(packet)
			if err != nil {
				server.logger.Error(fmt.Sprintf("[IF] Write error: %s", err))
			}
			server.logger.Debug(fmt.Sprintf("Send to interface, packets remain in channale: %d", len(server.writeToInterfaceCh)))
		}
	}
}

func (server *Server) readerFromInterfaceLoop() {
	buf := make([]byte, 1500)
	for {
		n, err := server.localInterface.Read(buf)
		if err != nil {
			server.logger.Error(fmt.Sprintf("[IF] Read error: %s", err))
			os.Exit(-1)
			return
		}

		packet := buf[:n]
		sourceIp, destIp, err := common.GetIpFromPacket(packet)
		if err != nil {
			server.logger.Debug(fmt.Sprintf("Unable to get IP from paket: %s", err))
			continue
		}

		client, ok := server.clinetConnectionRegister[destIp]
		if ok {
			client.WriteMessage(websocket.BinaryMessage, packet)
			server.logger.Debug(fmt.Sprintf("Server interface got package:\n%s", hex.Dump(packet)))
		} else {
			server.logger.Warn(fmt.Sprintf("Got package from %s to unknown dest ip address %s", sourceIp.String(), destIp.String()))
		}

	}
}

func (server *Server) writeToInterface(data []byte) {
	select {
	case server.writeToInterfaceCh <- data:
	default:
		server.logger.Error("[IF] Write channel full, dropping packet")
	}
}
