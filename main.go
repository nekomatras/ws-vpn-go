package main

import (
	"log"
	"net"
	"os"
	"os/exec"
	"encoding/hex"
	"github.com/songgao/water"
	"github.com/gorilla/websocket"
)

func createInterface(name ...string) *water.Interface {
	parameters := water.PlatformSpecificParams{}

	if len(name) > 0 {
		parameters.Name = name[0]
	}

	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: parameters,
	})
	if err != nil {
		log.Fatalf("Failed to create address to interface: %v", err)
		os.Exit(1)
	}
	log.Println("TUN interface created:", ifce.Name())
	return ifce
}

func setupInterface(iface *water.Interface, address string) {
	err := exec.Command("ip", "addr", "add", address, "dev", iface.Name()).Run()
	if err != nil {
		log.Fatalf("Failed to add address to interface %s: %v", iface.Name(), err)
		os.Exit(1)
	}
	err = exec.Command("ip", "link", "set", "dev", iface.Name(), "up").Run()
	if err != nil {
		log.Fatalf("Failed to add address to interface %s: %v", iface.Name(), err)
		os.Exit(1)
	}
}

func createClient(interfaceName string, interfaceAddress string, serverArrdess string) {

	iface := createInterface(interfaceName)
	setupInterface(iface, interfaceAddress)

	/* conn, err := net.Dial("udp", serverArrdess)
	if err != nil {
		log.Fatal(err)
	} */

	MTU := 1500

	buf := make([]byte, MTU)
	for {
		n, err := iface.Read(buf)
		if err != nil {
			log.Fatal("tun read:", err)
		}
		log.Printf("Client interface got package:\n%s",hex.Dump(buf[:n]))

		/* _, err = conn.Write(buf[:n])
		if err != nil {
			log.Fatal("udp write:", err)
		} */
	}
}

func createServer(interfaceName string, address string) {

	/* iface := createInterface(interfaceName)
	setupInterface(iface, address)

	MTU := 1500 */


	/* addr := net.UDPAddr{Port: 51820}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close() */


	/* buf := make([]byte, MTU)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatal("udp read:", err)
		}
		log.Printf("Server interface got package:\n%s",hex.Dump(buf[:n]))
		_, err = iface.Write(buf[:n]) // пишем в tun
		if err != nil {
			log.Fatal("tun write:", err)
		}
	} */
}

func createWebSocketConnection(wsUrl string) {
	url := "ws://localhost:8080/ws"

	// Подключаемся к серверу
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Ошибка подключения:", err)
	}
	defer conn.Close()

	log.Println("Подключено к серверу!")

	// Отправляем сообщение серверу
	err = conn.WriteMessage(websocket.TextMessage, []byte("Привет, сервер!"))
	if err != nil {
		log.Println("Ошибка отправки:", err)
		return
	}

	// Читаем ответ от сервера
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Println("Ошибка чтения:", err)
		return
	}

	log.Printf("Ответ от сервера: %s\n", message)
}

func main() {

	tunelAddress := "192.168.0.16:51820"

	clientInterfaceName := "tunClient"
	clientAaddress := "10.0.0.10/24";

	/* serverInterfaceName := "tunServer"
	serverAaddress := "10.0.0.20/24"; */

	go createClient(clientInterfaceName, clientAaddress, tunelAddress)

	select {}
}
