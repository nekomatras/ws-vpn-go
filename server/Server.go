package server

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
}

func NewServer(wsUrl string) *Server {
	log.Printf("Create server: WebSocket URL: \"%s\"; MTU: %d", wsUrl, DefaultMTU)
	return &Server{
		localWebSocketURL:        wsUrl,
		mtu:                      DefaultMTU,
		isInited:                 false,
		writeToInterfaceCh:       make(chan []byte, 256),
		clinetConnectionRegister: make(map[common.IpAddress]*websocket.Conn),
	}
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
	log.Printf("Set interface name: %s", interfaceName)
	server.interfaceName = interfaceName
}

func (server *Server) SetInterfaceAddress(interfaceAddress string) {
	log.Printf("Set interface address: %s", interfaceAddress)
	server.interfaceAddress = interfaceAddress
}

// Upgrade HTTP connection to WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		log.Printf("Try to open WS connection to: %s", r.RemoteAddr)
		return true // проверить ключ из заголовка
	},
}

/*
0. В хендлере в цикле читаем пакеты от клиента и сразу кидаем в интерфейс
1. На этапе авторизации получайм от клиента ключ и его локальный ip (нужно проверять, что ip реально его...)
2. Создаем канал из клиентского WS и кладем его в мапу по его ip
3. В хендлере читаем из канала и кидаем все в канал интерфейса
4. Создаем поток, который будет читать интерфейс и каждый входящий пакет будет кидать в соответствующий ip назначения WS канал
*/

func (server *Server) connectionHandler(w http.ResponseWriter, r *http.Request) {

	sourceIp := r.RemoteAddr

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Connection upgrade failed for %s; Error: %s", sourceIp, err)
		return
	}

	defer ws.Close()

	clientIP := common.GetIpFromString(r.Header.Get("ClientIp"))
	if clientIP == common.GetAllZeroIp() {
		log.Printf("Unable to read target IP in request from: %s. Connection closed", sourceIp)
		return
	}

	server.clinetConnectionRegister[clientIP] = ws
	defer delete(server.clinetConnectionRegister, clientIP)

	log.Printf("Client %s connected from %s", clientIP.String(), sourceIp)

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("[%s] Unable to read from WS: %s", clientIP.String(), err)
			}
			break
		}

		log.Printf("[%s] Got message from WS connection:\n%s", clientIP.String(), hex.Dump(message))

		server.writeToInterface(message)
	}

	log.Printf("[%s] Client from %s disconnected", clientIP.String(), sourceIp)
}

func (server *Server) listenWebSocket() error {
	var err error

	u, err := url.Parse(server.localWebSocketURL)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc(u.Path, server.connectionHandler)

	err = http.ListenAndServe(u.Host, nil)
	if err != nil {
		return fmt.Errorf("ListenAndServe error: %w", err)
	}

	log.Printf("Server listening on:%s", server.localWebSocketURL)

	return nil
}

func (server *Server) Init() error {
	var err error

	log.Println("Setup interface...")
	server.localInterface, err = common.CreateInterface(server.interfaceName)
	if err != nil {
		return fmt.Errorf("Interface creation error: %w", err)
	}

	err = common.SetupInterface(server.localInterface, server.interfaceAddress)
	if err != nil {
		return fmt.Errorf("Interface setup error: %w", err)
	}

	server.isInited = true
	log.Printf("Create VPN interface %s", server.interfaceName)

	go func() {
		err = server.listenWebSocket()
		if err != nil {
			log.Fatalf("Error while WS listen: %w", err)
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
				log.Fatal("[IF] Write channel closed, exiting writer loop")
				return
			}
			_, err := server.localInterface.Write(packet)
			if err != nil {
				log.Printf("[IF] Write error: %s", err)
			}
			fmt.Printf("Send to interface, packets remain in channale: %d", len(server.writeToInterfaceCh))
		}
	}
}

func (server *Server) readerFromInterfaceLoop() {
	buf := make([]byte, 1500)
	for {
		n, err := server.localInterface.Read(buf)
		if err != nil {
			log.Printf("[IF] Read error: %s", err)
			return
		}

		packet := buf[:n]
		sourceIp, destIp, err := common.GetIpFromPacket(packet)
		if err != nil {
			log.Printf("Unable to get IP from paket: %s", err)
			continue
		}

		client, ok := server.clinetConnectionRegister[destIp]
		if ok {
			client.WriteMessage(websocket.BinaryMessage, packet)
			log.Printf("Server interface got package:\n%s", hex.Dump(packet))
		} else {
			log.Printf("Got package from %s to unknown dest ip address %s", sourceIp.String(), destIp.String())
		}

	}
}

func (server *Server) writeToInterface(data []byte) {
	select {
	case server.writeToInterfaceCh <- data:
	default:
		log.Println("[IF] Write channel full, dropping packet")
	}
}
