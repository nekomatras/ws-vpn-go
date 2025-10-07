package server

import (
	"fmt"
	"log/slog"

	"ws-vpn-go/common"
	netinterface "ws-vpn-go/server/interface"
	"ws-vpn-go/server/tunnel/wstunnel"
)

const DefaultMTU = 1500

type Server struct {
	serverInfo common.ServerInfo

	netInterface *netinterface.NetworkInterface
	tunnel       *wstunnel.WsTunnel
	logger       *slog.Logger
}

func New(
	interfaceAddress string,
	interfaceName string,
	mtu uint,
	wsUrl string,
	key string,
	logger *slog.Logger) *Server {

	logger.Info(fmt.Sprintf("Create server: WebSocket URL: \"%s\"; MTU: %d", wsUrl, mtu))

	info := common.ServerInfo{
		MTU:                   mtu,
		InternalServerAddress: interfaceAddress,
	}

	server := &Server{
		netInterface: netinterface.New(interfaceAddress, interfaceName, mtu, logger),
		tunnel:       wstunnel.New(wsUrl, key, info, logger),
		logger:       logger,
		serverInfo:   info,
	}

	return server
}

func (server *Server) Start() error {
	var err error

	err = server.netInterface.Init()
	if err != nil {
		server.logger.Error(fmt.Sprintf("Unable to setup interface: %v", err))
		return err
	}

	go server.tunnel.Listen()

	go server.tunnel.WriteTo(*server.netInterface.Interface())
	go server.netInterface.WriteTo(server.tunnel.WriteToTunnel)

	return nil
}
