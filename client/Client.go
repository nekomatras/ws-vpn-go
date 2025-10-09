package client

import (
	"fmt"
	"time"
	"log/slog"
	"net/http"
	"encoding/json"

	"ws-vpn-go/common"
	"ws-vpn-go/common/tunnel"
	"ws-vpn-go/common/interface"
	"ws-vpn-go/client/tunnel/wstunnel"
)

const DefaultMTU = 1500

type Client struct {

	netInterface *netinterface.NetworkInterface
	tunnel       tunnel.Tunnel

	key string
	interfaceName string

	remoteAddress string
	tunnelPath string
	registerPath string

	ipAddress common.IpAddress
	mtu uint

	logger         *slog.Logger
}

func New(remoteAddress string, tunnelPath string, registerPath string, key string, interfaceName string, logger *slog.Logger) *Client {
	logger.Info(fmt.Sprintf("Create client: Target URL: \"%s\"", remoteAddress + tunnelPath))
	return &Client{
		remoteAddress: remoteAddress,
		tunnelPath: tunnelPath,
		registerPath: registerPath,
		tunnel: wstunnel.New(remoteAddress, tunnelPath, key, logger),
		key: key,
		logger: logger}
}


func (client *Client) Start() error {
	var err error

	info, err := client.register()
	if err != nil {
		client.logger.Error(fmt.Sprintf("Unable to get informantion from server: %v", err))
		return err
	}

	client.ipAddress = common.GetIpFromString(info.ClientIp)
	client.mtu = info.MTU



	client.netInterface = netinterface.New(client.ipAddress.String(), client.interfaceName, client.mtu, client.logger)
	err = client.netInterface.Init()
	if err != nil {
		client.logger.Error(fmt.Sprintf("Unable to setup interface: %v", err))
		return err
	}

	err = client.tunnel.ReserveConnection(client.ipAddress)
	if err != nil {
		return err
	}

	err = client.tunnel.Run()
	if err != nil {
		return err
	}

	go client.tunnel.WriteTo(*client.netInterface.Interface())
	go client.netInterface.WriteTo(client.tunnel.WriteToTunnel)

	return nil
}

func (client *Client) register() (common.ServerInfo, error) {
	{
		httpClient := http.Client{
			Timeout: 5 * time.Second, // таймаут на всякий случай
		}

		req, err := http.NewRequest("GET", client.remoteAddress + client.registerPath, nil)
		if err != nil {
			return common.ServerInfo{}, fmt.Errorf("cannot create request: %w", err)
		}

		req.Header.Set("Key", client.key)

		resp, err := httpClient.Do(req)
		if err != nil {
			return common.ServerInfo{}, fmt.Errorf("cannot make request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return common.ServerInfo{}, fmt.Errorf("server returned status: %s", resp.Status)
		}

		var info common.ServerInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			return common.ServerInfo{}, fmt.Errorf("cannot decode response: %w", err)
		}

		return info, nil
	}
}