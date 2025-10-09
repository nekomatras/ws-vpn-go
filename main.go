package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
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



	if config.Mode == "client" {

		logger := common.GetLoggerWithName(baseLogger, "Client")

		client := client.New(
			config.RemoteAddress,
			config.TunnelPath,
			config.RegisterPath,
			config.Key,
			config.InterfaceName,
			logger)

		err := client.Start()
		if err != nil {
			logger.Error(err.Error())
			os.Exit(-1)
		}

	} else if config.Mode == "server" {

		logger := common.GetLoggerWithName(baseLogger, "Server")

		server, err := server.New(
			config.Network,
			config.InterfaceName,
			config.MTU,
			config.ListenAddress,
			config.RegisterPath,
			config.TunnelPath,
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
