package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"ws-vpn-go/common"
	netinterface "ws-vpn-go/common/interface"
	"ws-vpn-go/common/tunnel"
	"ws-vpn-go/server/contentmanager"
	"ws-vpn-go/server/networkmanager"
	"ws-vpn-go/server/tunnel/wstunnel"
)

type Server struct {
	key            string
	serverInfo     common.ServerInfo

	listenAddress  string
	registerPath   string

	secure         bool
	privateKeyPath string
	chainPath      string

	netInterface   *netinterface.NetworkInterface
	tunnel         tunnel.Tunnel
	netManager     *networkmanager.Networkmanager
	contentManager *contentmanager.ContentManager

	logger         *slog.Logger
}

func New(
	config *common.Config,
	logger *slog.Logger) (*Server, error) {
	var err error

	logger.Info(fmt.Sprintf(
		"Create server: WebSocket URL: \"%s%s\"; MTU: %d",
		config.ListenAddress,
		config.TunnelPath,
		config.MTU,
	))

	networkManager, err := networkmanager.New(config.Network)
	if err != nil {
		return nil, err
	}

	contentManager, err := contentmanager.New(config.DefaultPagePath, config.StaticFolderPath, logger)
	if err != nil {
		return nil, err
	}

	address, err := networkManager.AssignRouterAddress()
	if err != nil {
		return nil, err
	}

	info := common.ServerInfo{
		MTU:                   config.MTU,
		GatewayIp:             address.String(),
		TunnelPath:            config.TunnelPath,
	}

	server := &Server{
		netInterface:   netinterface.New(
			fmt.Sprintf(
				"%s/%d",
				address.String(),
				networkManager.GetSubNet(),
			),
			config.InterfaceName,
			config.MTU,
			logger,
		),
		tunnel:         wstunnel.New(config.TunnelPath, config.Key, info, logger),
		netManager:     networkManager,
		contentManager: contentManager,
		listenAddress:  config.ListenAddress,
		registerPath:   config.RegisterPath,
		logger:         logger,
		key:            config.Key,
		serverInfo:     info,
		secure:         config.Secure,
		privateKeyPath: config.PrivateKey,
		chainPath:      config.Chain,
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

	go func() {
		if server.secure {
			err := http.ListenAndServeTLS(
				server.listenAddress,
				server.chainPath,
				server.privateKeyPath,
				serverMux)
			if err != nil {
				server.logger.Error(err.Error())
				os.Exit(-1)
			}
		} else {
			err := http.ListenAndServe(server.listenAddress, serverMux)
			if err != nil {
				server.logger.Error(err.Error())
				os.Exit(-1)
			}
		}
	}()

	go server.tunnel.WriteTo(*server.netInterface.Interface())
	go server.netInterface.WriteTo(server.tunnel.WriteToTunnel)

	return nil
}

func (server *Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	server.logger.Info(fmt.Sprintf("Process register from: %s", common.GetRealIP(r)))
	if common.CheckKey(r, server.key) {

		macString := r.Header.Get("MAC")
		mac := common.GetMacFromString(macString)
		if mac == common.GetAllZeroMac() {
			server.contentManager.WriteContentToResponse(w, r)
			return
		}

		clientIp, ok := server.netManager.GetAddress(mac)

		if !ok {
			var err error;
			clientIp, err = server.netManager.AssignAddress(mac)

			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			server.logger.Info(fmt.Sprintf("[%s] Reserve connection. Clinet IP: %s", common.GetRealIP(r), clientIp.String()))
			server.tunnel.ReserveConnection(clientIp)
		}

		clientInfo := server.serverInfo
		clientInfo.ClientIp = fmt.Sprintf("%s/%d", clientIp.String(), server.netManager.GetSubNet())

		server.logger.Info(fmt.Sprintf("[%s] Send info: %+v", common.GetRealIP(r), clientInfo))
		clientInfo.WriteToResponse(w)
		w.WriteHeader(http.StatusAccepted)
	} else {
		server.logger.Error(fmt.Sprintf("[%s] Try to register with wrong key", common.GetRealIP(r)))
		server.contentManager.WriteContentToResponse(w, r)
	}
}
