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

var help = flag.Bool(
	"help",
	false,
	"Get manual")

var configFile = flag.String(
	"config-path",
	"/etc/ws-wpn.conf",
	"Set this flag to use configuration file. Example: -config-path=/etc/ws-wpn.conf")

var configMode = flag.Bool(
	"config",
	false,
	"Set this flag to use parameters from config file.")

var envieronmentMode = flag.Bool(
	"env",
	false,
	"Set this flag to use envieronment variables.")

func main() {
	flag.Parse()

	if (*help) {
		fmt.Println("Not implemented yet...")
		os.Exit(1)
	}

	baseLogger := common.NewLogger(os.Stdout, slog.LevelDebug)

	if *configMode == *envieronmentMode {
        baseLogger.Error("Unable to start. You should specify configuration mode: --env or --config")
        os.Exit(1)
    }

	var configPtr *common.Config;
	var configErr error;
	var configType string;

	if *configMode {
		configPtr, configErr = common.LoadConfigFromFile(*configFile)
		configType = *configFile
	} else if *envieronmentMode {
		configPtr, configErr = common.LoadConfigFromEnvieronment()
		configType = "Envieronment Variables"
	} else {
		baseLogger.Error("Unknown config mode...")
		os.Exit(1)
	}

	if (configErr != nil) {
		baseLogger.Error(configErr.Error())
		os.Exit(-1)
	}

	config := *configPtr

	baseLogger.Info(fmt.Sprintf("Start with config [%s]:\n%+v", configType, config))

	switch config.Mode {
	case "client":

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

	case "server":

		logger := common.GetLoggerWithName(baseLogger, "Server")

		server, err := server.New(
			config.Network,
			config.InterfaceName,
			config.MTU,
			config.ListenAddress,
			config.RegisterPath,
			config.TunnelPath,
			config.Key,
			config.DefaultPagePath,
			config.StaticFolderPath,
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

	default:
		baseLogger.Error(fmt.Sprintf("Unknown mode: %s", config.Mode))
		os.Exit(-1)
	}

	select {}
}
