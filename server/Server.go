package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"ws-vpn-go/common"
	netinterface "ws-vpn-go/common/interface"
	"ws-vpn-go/common/tunnel"
	"ws-vpn-go/server/contentmanager"
	"ws-vpn-go/server/networkmanager"
	"ws-vpn-go/server/tunnel/wstunnel"
)

type Server struct {
	key        string
	serverInfo common.ServerInfo

	listenAddress string
	registerPath  string

	netInterface   *netinterface.NetworkInterface
	tunnel         tunnel.Tunnel
	netManager     *networkmanager.Networkmanager
	contentManager *contentmanager.ContentManager

	logger *slog.Logger
}

func New(
	network string,
	interfaceName string,
	mtu uint,
	listenAddress string,
	registerPath string,
	tunnelPath string,
	key string,
	pagePath string,
	staticPath string,
	logger *slog.Logger) (*Server, error) {
	var err error

	logger.Info(fmt.Sprintf("Create server: WebSocket URL: \"%s%s\"; MTU: %d", listenAddress, tunnelPath, mtu))

	networkManager, err := networkmanager.New(network)
	if err != nil {
		return nil, err
	}

	contentManager, err := contentmanager.New(pagePath, staticPath)
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
		netInterface:   netinterface.New(address.String(), interfaceName, mtu, logger),
		tunnel:         wstunnel.New(tunnelPath, key, info, logger),
		netManager:     networkManager,
		contentManager: contentManager,
		logger:         logger,
		key:            key,
		serverInfo:     info,
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

	serverMux.HandleFunc(server.registerPath, server.registerHandler)

	server.contentManager.RegisterHandlers(serverMux)

	server.tunnel.RegisterHandlers(serverMux)
	server.tunnel.SetConnectionCloseHandler(server.netManager.FreeAddress)

	server.tunnel.Run()
	go http.ListenAndServe(server.listenAddress, serverMux)

	go server.tunnel.WriteTo(*server.netInterface.Interface())
	go server.netInterface.WriteTo(server.tunnel.WriteToTunnel)

	return nil
}

func (server *Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	if common.CheckKey(r, server.key) {

		macString := r.Header.Get("MAC")
		mac := common.GetMacFromString(macString)
		if mac == common.GetAllZeroMac() {
			server.contentManager.WriteContentToResponse(w)
		}

		clientIp, ok := server.netManager.GetAddress(mac)

		if !ok {
			clientIp, err := server.netManager.AssignAddress(mac)

			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			server.tunnel.ReserveConnection(clientIp)
		}

		clientInfo := server.serverInfo
		clientInfo.ClientIp = clientIp.String()

		clientInfo.WriteToResponse(w)
		w.WriteHeader(http.StatusAccepted)
	} else {
		server.contentManager.WriteContentToResponse(w)
	}
}
