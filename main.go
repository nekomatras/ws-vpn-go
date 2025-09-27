package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"
	"ws-vpn-go/client"
	"ws-vpn-go/common"
	"ws-vpn-go/server"
)

func main() {

	baseLogger := common.NewLogger(os.Stdout, slog.LevelDebug)

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

		logger := common.GetLoggerWithName(baseLogger, "Client")

		client := client.New(*remoteUrl, logger)
		client.SetInterfaceAddress(clientAaddress)
		client.SetInterfaceName(clientInterfaceName)
		error := client.Init()

		if !client.IsInited() {
			logger.Error(error.Error())
			os.Exit(-1)
		}

		for {

			client.ConnectToRemote()

			if client.IsConnectedToRemote() {
				client.Run()
				break
			}

			logger.Warn(fmt.Sprintf("Unable to connect remote WS, wait %s and retry...", TimeToWait))
			time.Sleep(TimeToWait)
		}

	} else if *mode == "server" {

		logger := common.GetLoggerWithName(baseLogger, "Server")

		server := server.New(*remoteUrl, logger)
		server.SetInterfaceAddress(remoteInterfaceAaddress)
		server.SetInterfaceName(remoteInterfaceName)
		error := server.Init()
		if error != nil {
			logger.Error(error.Error())
			os.Exit(-1)
		}

	} else {
		baseLogger.Error(fmt.Sprintf("Unknown mode: -mode=%s", *mode))
		os.Exit(-1)
	}

	select {}
}
