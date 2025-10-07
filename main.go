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

var configFile = flag.String(
	"config",
	"/etc/ws-wpn.conf",
	"Set path to configuration file. Example: -config=/etc/ws-wpn.conf")

func main() {
	flag.Parse()

	baseLogger := common.NewLogger(os.Stdout, slog.LevelDebug)

	config, err := common.LoadConfig(*configFile)
	if (err != nil) {
		baseLogger.Error(err.Error())
		os.Exit(-1)
	}

	TimeToWait := 5 * time.Second

	fmt.Println(*config)

	if config.Mode == "client" {

		logger := common.GetLoggerWithName(baseLogger, "Client")

		client := client.New(config.RemoteUrl, logger)
		//client.SetInterfaceAddress(config.InterfaceAddress)
		client.SetInterfaceName(config.InterfaceName)
		client.SetKey(config.Key)
		err := client.Init()

		if !client.IsInited() {
			logger.Error(err.Error())
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

	} else if config.Mode == "server" {

		logger := common.GetLoggerWithName(baseLogger, "Server")

		server, err := server.New(
			config.Network,
			config.InterfaceName,
			config.MTU,
			config.RemoteUrl,
			config.Key,
			logger)

		if err != nil {
			logger.Error(err.Error())
			os.Exit(-1)
		}

		err = server.Start()
		if err != nil {
			logger.Error(err.Error())
			os.Exit(-1)
		}

	} else {
		baseLogger.Error(fmt.Sprintf("Unknown mode: %s", config.Mode))
		os.Exit(-1)
	}

	select {}
}
