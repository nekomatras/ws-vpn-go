package main

import (
	"flag"
	"log"
	"time"
	"ws-vpn-go/client"
	"ws-vpn-go/server"
)

func main() {

	mode := flag.String("mode", "none", "Select mode")
	remoteUrl := flag.String(
		"remote",
		"ws://7.7.7.7:8080/ws",
		"Set remote WebSocket URL. Example: \"ws://7.7.7.7:8080/ws\"")
	flag.Parse()

	TimeToWait := 5 * time.Second

	// remoteUrl := "ws://7.7.7.7:8080/ws"

	clientInterfaceName := "tunClient"
	clientAaddress := "10.0.0.5/24"

	remoteInterfaceAaddress := "10.0.0.1/24"
	remoteInterfaceName := "tunServer"

	if *mode == "client" {

		client := client.NewClient(*remoteUrl)
		client.SetInterfaceAddress(clientAaddress)
		client.SetInterfaceName(clientInterfaceName)
		error := client.Init()

		if !client.IsInited() {
			log.Fatalln(error)
		}

		for {

			client.ConnectToRemote()

			if client.IsConnectedToRemote() {
				client.Run()
				break
			}

			log.Printf("Unable to connect remote WS, wait %s and retry...", TimeToWait)
			time.Sleep(TimeToWait)
		}

	} else if *mode == "server" {

		server := server.NewServer(*remoteUrl)
		server.SetInterfaceAddress(remoteInterfaceAaddress)
		server.SetInterfaceName(remoteInterfaceName)
		error := server.Init()
		if error != nil {
			log.Fatalln(error)
		}
		log.Println("------------------")

	} else {
		log.Fatalf("Unknown mode: -mode=%s", *mode)
	}

	select {}
}
