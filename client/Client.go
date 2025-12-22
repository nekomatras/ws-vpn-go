package client

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"ws-vpn-go/client/tunnel/wstunnel"
	"ws-vpn-go/common"
	netinterface "ws-vpn-go/common/interface"
	"ws-vpn-go/common/tunnel"
)

const Timeout = 5 * time.Second
const Repeat = 5

type Client struct {
	netInterface *netinterface.NetworkInterface
	tunnel       tunnel.Tunnel

	key           string
	interfaceName string

	remoteAddress string
	tunnelPath    string
	registerPath  string

	ipAddress common.IpAddress
	mtu       uint

	logger *slog.Logger
}

func New(remoteAddress string, tunnelPath string, registerPath string, key string, interfaceName string, logger *slog.Logger) *Client {
	logger.Info(fmt.Sprintf("Create client: Target URL: \"%s\"", remoteAddress+tunnelPath))
	return &Client{
		remoteAddress: remoteAddress,
		tunnelPath:    tunnelPath,
		registerPath:  registerPath,
		tunnel:        wstunnel.New(remoteAddress, tunnelPath, key, logger),
		key:           key,
		interfaceName: interfaceName,
		logger:        logger}
}

func (client *Client) Start() error {
	var err error

	info, err := client.tryRegister()
	if err != nil {
		client.logger.Error(fmt.Sprintf("Unable to get informantion from server: %v", err))
		return err
	}

	client.logger.Info(fmt.Sprintf("Got server info: %+v", info))

	client.ipAddress = common.GetIpFromString(info.ClientIp)
	client.mtu = info.MTU

	client.netInterface = netinterface.New(info.ClientIp, client.interfaceName, client.mtu, client.logger)
	err = client.netInterface.Init()
	if err != nil {
		client.logger.Error(fmt.Sprintf("Unable to setup interface: %v", err))
		return err
	}

	err = client.tunnel.ReserveConnection(client.ipAddress)
	if err != nil {
		return err
	}

	err = client.tryRunTunnel()
	if err != nil {
		return err
	}

	go client.tunnel.WriteTo(*client.netInterface.Interface())
	go client.netInterface.WriteTo(client.tunnel.WriteToTunnel)

	return nil
}

func (client *Client) tryRunTunnel() error {
	var err error
	repeat := Repeat

	for repeat != 0 {
		err = client.tunnel.Run()
		if err == nil {
			return nil
		}

		repeat--
		client.logger.Warn(fmt.Sprintf("Unable to connect tunnel: %v; Retry %d times", err, repeat))
	}

	return err
}

func (client *Client) tryRegister() (common.ServerInfo, error) {
	var err error
	repeat := Repeat

	info := common.ServerInfo{}

	for repeat != 0 {
		info, err = client.register()
		if err == nil {
			return info, err
		}

		repeat--
		client.logger.Warn(fmt.Sprintf("Unable to get server data: %v; Retry %d times", err, repeat))
	}

	return info, err
}

func (client *Client) register() (common.ServerInfo, error) {
	{
		httpClient := http.Client{
			Timeout: Timeout,
		}

		url := client.remoteAddress + client.registerPath
		client.logger.Info(fmt.Sprintf("Try to register: %s", url))

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return common.ServerInfo{}, fmt.Errorf("cannot create request: %w", err)
		}

		req.Header.Set("Key", client.key)
		req.Header.Set("MAC", "a0:ad:9f:81:3f:cc")

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
