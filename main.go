package main

import (
	"ws-vpn-go/client"
)

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

// Upgrade HTTP connection to WebSocket
/* var upgrader = websocket.Upgrader{
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

} */

func main() {

	// tunelAddress := "192.168.0.16:51820"

	/* serverInterfaceName := "tunServer"
	serverAaddress := "10.0.0.20/24"; */

	//go createClient(clientInterfaceName, clientAaddress, tunelAddress)

	// go createWebSocketServer("/ws")

	url := "ws://localhost:8080/ws"
	clientInterfaceName := "tunClient"
	clientAaddress := "10.0.0.10/24"

	client := client.NewClient(url)
	client.SetInterfaceAddress(clientAaddress)
	client.SetInterfaceName(clientInterfaceName)
	client.Init()
	client.Run()

	select {}
}
