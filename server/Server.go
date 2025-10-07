package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"ws-vpn-go/common"
	"ws-vpn-go/server/interface"
	"ws-vpn-go/server/networkmanager"
	"ws-vpn-go/server/tunnel"
	"ws-vpn-go/server/tunnel/wstunnel"
)

const DefaultMTU = 1500

type Server struct {
	key          string
	serverInfo   common.ServerInfo

	netInterface *netinterface.NetworkInterface
	tunnel       tunnel.Tunnel
	netManager   *networkmanager.Networkmanager
	logger       *slog.Logger
}

func New(
	network string,
	interfaceName string,
	mtu uint,
	wsUrl string,
	key string,
	logger *slog.Logger) (*Server, error) {
	var err error

	logger.Info(fmt.Sprintf("Create server: WebSocket URL: \"%s\"; MTU: %d", wsUrl, mtu))

	networkManager, err := networkmanager.New(network)
	if err != nil {
		return nil, err
	}

	address, err := networkManager.AssignRouterAddress()
	if err != nil {
		return nil, err
	}

	info := common.ServerInfo{
		MTU:                   mtu,
		InternalServerAddress: address.String(),
	}

	server := &Server{
		netInterface: netinterface.New(address.String(), interfaceName, mtu, logger),
		tunnel:       wstunnel.New(wsUrl, key, info, logger),
		netManager:   networkManager,
		logger:       logger,
		key:          key,
		serverInfo:   info,
	}

	return server, nil
}

func (server *Server) Start() error {

	err := server.netInterface.Init()
	if err != nil {
		server.logger.Error(fmt.Sprintf("Unable to setup interface: %v", err))
		return err
	}

	serverMux := http.NewServeMux()

	serverMux.HandleFunc("/info", server.infoHandler) //раскидать по отдельным классам
	serverMux.HandleFunc("/", server.defaultHandler)
	server.tunnel.RegisterHandlers(serverMux)

	go server.tunnel.Listen()
	go http.ListenAndServe("0.0.0.0:443", serverMux) //TODO

	go server.tunnel.WriteTo(*server.netInterface.Interface())
	go server.netInterface.WriteTo(server.tunnel.WriteToTunnel)

	return nil
}

func (server *Server) infoHandler(w http.ResponseWriter, r *http.Request) {
	if common.CheckKey(r, server.key) {
		server.serverInfo.WriteToResponse(w)
		w.WriteHeader(http.StatusAccepted)
	} else {
		server.defaultHandler(w, r)
	}
}

func (tunnel *Server) defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "САДОВОД-ЛЮБИТЕЛЬ")
	w.WriteHeader(http.StatusOK)
}
