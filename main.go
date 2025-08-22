package main

import (
	"log"
	// "net"
	// "os"
	// "os/exec"
	// "encoding/hex"
	// "github.com/songgao/water"

	"github.com/gorilla/websocket"
	"net/http"
)

// func createInterface(name ...string) *water.Interface {
// 	parameters := water.PlatformSpecificParams{}

// 	if len(name) > 0 {
// 		parameters.Name = name[0]
// 	}

// 	ifce, err := water.New(water.Config{
// 		DeviceType: water.TUN,
// 		PlatformSpecificParams: parameters,
// 	})
// 	if err != nil {
// 		log.Fatalf("Failed to create address to interface: %v", err)
// 		os.Exit(1)
// 	}
// 	log.Println("TUN interface created:", ifce.Name())
// 	return ifce
// }

// func setupInterface(iface *water.Interface, address string) {
// 	err := exec.Command("ip", "addr", "add", address, "dev", iface.Name()).Run()
// 	if err != nil {
// 		log.Fatalf("Failed to add address to interface %s: %v", iface.Name(), err)
// 		os.Exit(1)
// 	}
// 	err = exec.Command("ip", "link", "set", "dev", iface.Name(), "up").Run()
// 	if err != nil {
// 		log.Fatalf("Failed to add address to interface %s: %v", iface.Name(), err)
// 		os.Exit(1)
// 	}
// }

// func createClient(interfaceName string, interfaceAddress string, serverArrdess string) {

// 	iface := createInterface(interfaceName)
// 	setupInterface(iface, interfaceAddress)

// 	/* conn, err := net.Dial("udp", serverArrdess)
// 	if err != nil {
// 		log.Fatal(err)
// 	} */

// 	MTU := 1500

// 	buf := make([]byte, MTU)
// 	for {
// 		n, err := iface.Read(buf)
// 		if err != nil {
// 			log.Fatal("tun read:", err)
// 		}
// 		log.Printf("Client interface got package:\n%s",hex.Dump(buf[:n]))

// 		/* _, err = conn.Write(buf[:n])
// 		if err != nil {
// 			log.Fatal("udp write:", err)
// 		} */
// 	}
// }

// func createServer(interfaceName string, address string) {

// 	/* iface := createInterface(interfaceName)
// 	setupInterface(iface, address)

// 	MTU := 1500 */


// 	/* addr := net.UDPAddr{Port: 51820}
// 	conn, err := net.ListenUDP("udp", &addr)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer conn.Close() */


// 	/* buf := make([]byte, MTU)
// 	for {
// 		n, err := conn.Read(buf)
// 		if err != nil {
// 			log.Fatal("udp read:", err)
// 		}
// 		log.Printf("Server interface got package:\n%s",hex.Dump(buf[:n]))
// 		_, err = iface.Write(buf[:n]) // пишем в tun
// 		if err != nil {
// 			log.Fatal("tun write:", err)
// 		}
// 	} */
// }

func createWebSocketClinet(wsUrl string) {

	// Подключаемся к серверу
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		log.Fatal("[C] Connection error:", err)
	}
	defer conn.Close()

	log.Println("[C] Connected to server")

	// Отправляем сообщение серверу
	err = conn.WriteMessage(websocket.TextMessage, []byte("MESSAGE"))
	if err != nil {
		log.Println("[C] Send error:", err)
		return
	}

	// Читаем ответ от сервера
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Println("[C] Read error", err)
		return
	}

	log.Printf("[C] Server response: %s\n", message)

    // Close connection gracefully
	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("[S] Error sending close message:", err)
		return
	}
}

// Upgrade HTTP connection to WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all connections; adjust in production
	},
}

func connectionHandler(w http.ResponseWriter, r *http.Request) {

	// Upgrade initial GET request to a WebSocket
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("[S] Upgrade error:", err)
        return
    }

    defer ws.Close()

	log.Println("[S] Client connected")

    // Infinite loop to read messages
    for {
        messageType, message, err := ws.ReadMessage()
        if err != nil {
            if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
                log.Println("[S] Read error:", err)
            }
            break
        }
        log.Printf("[S] Received: %s\n", message)

        // Echo the message back to client
        err = ws.WriteMessage(messageType, message)
        if err != nil {
            log.Println("[S] Write error:", err)
            break
        }
    }

    log.Println("[S] Client disconnected")
}

func createWebSocketServer(wsUrl string) {

	http.HandleFunc(wsUrl, connectionHandler)

    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe error:", err)
    }

	log.Println("Server listening on :8080/ws")

}

func main() {

	// tunelAddress := "192.168.0.16:51820"

	// clientInterfaceName := "tunClient"
	// clientAaddress := "10.0.0.10/24";

	/* serverInterfaceName := "tunServer"
	serverAaddress := "10.0.0.20/24"; */

	//go createClient(clientInterfaceName, clientAaddress, tunelAddress)

    url := "ws://localhost:8080/ws"

	go createWebSocketServer("/ws")
    go createWebSocketClinet(url)

	select {}
}
