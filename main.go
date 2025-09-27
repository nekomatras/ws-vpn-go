package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"time"
	"ws-vpn-go/client"
	"ws-vpn-go/common"
	"ws-vpn-go/server"
)

func main() {

	baseLogger := common.NewLogger(os.Stdout, slog.LevelDebug)
	logger := common.GetLoggerWithName(baseLogger, "123")

	logger.Debug("Test")
	logger.Warn("Aboba")
	logger.Error("ABABO")

	mode := flag.String("mode", "none", "Select mode")
	remoteUrl := flag.String(
		"remote",
		"ws://8.8.8.8:8080/ws",
		"Set remote WebSocket URL. Example: \"ws://8.8.8.8:8080/ws\"")
	flag.Parse()

	TimeToWait := 5 * time.Second

	clientInterfaceName := "tunClient"
	clientAaddress := "10.0.0.5/24"

	remoteInterfaceAaddress := "10.0.0.1/24"
	remoteInterfaceName := "tunServer"

	if *mode == "client" {

		client := client.New(*remoteUrl)
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

		server := server.New(*remoteUrl)
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
