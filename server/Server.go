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

	listenAddress string
	registerPath string

	netInterface *netinterface.NetworkInterface
	tunnel       tunnel.Tunnel
	netManager   *networkmanager.Networkmanager
	logger       *slog.Logger
}

func New(
	network string,
	interfaceName string,
	mtu uint,
	listenAddress string,
	registerPath string,
	tunnelPath string,
	key string,
	logger *slog.Logger) (*Server, error) {
	var err error

	logger.Info(fmt.Sprintf("Create server: WebSocket URL: \"%s%s\"; MTU: %d", listenAddress, tunnelPath, mtu))

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
		tunnel:       wstunnel.New(tunnelPath, key, info, logger),
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

	serverMux.HandleFunc(server.registerPath, server.registerHandler) //раскидать по отдельным классам
	serverMux.HandleFunc("/", server.defaultHandler)

	server.tunnel.RegisterHandlers(serverMux)
	server.tunnel.SetConnectionCloseHandler(server.netManager.FreeAddress)

	go server.tunnel.Listen()
	go http.ListenAndServe(server.listenAddress, serverMux) //TODO

	go server.tunnel.WriteTo(*server.netInterface.Interface())
	go server.netInterface.WriteTo(server.tunnel.WriteToTunnel)

	return nil
}

func (server *Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	if common.CheckKey(r, server.key) {

		macString := r.Header.Get("MAC")
		mac := common.GetMacFromString(macString)
		if mac == common.GetAllZeroMac() {
			server.defaultHandler(w, r)
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
		server.defaultHandler(w, r)
	}
}

func (tunnel *Server) defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "САДОВОД-ЛЮБИТЕЛЬ")
	w.WriteHeader(http.StatusOK)
}
