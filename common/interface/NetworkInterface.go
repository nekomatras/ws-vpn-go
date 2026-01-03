package netinterface

import (
	"fmt"
	"log/slog"

	"ws-vpn-go/common"
)

type NetworkInterface struct {

	address              string
	name                 string
	mtu                  uint

	localInterface       common.Interface

	logger               *slog.Logger
}

func New(
	interfaceAddress string,
	interfaceName string,
	mtu uint,
	logger *slog.Logger) *NetworkInterface {

	return &NetworkInterface{
		address: interfaceAddress,
		name: interfaceName,
		mtu: mtu,
		logger: logger,
	}
}

func (netInterface *NetworkInterface) Interface() *common.Interface {
	return &netInterface.localInterface
}

func (netInterface *NetworkInterface) Init() error {
	var err error

	netInterface.logger.Info("Setup interface...")
	netInterface.localInterface, err = common.CreateInterface(netInterface.name)
	if err != nil {
		return fmt.Errorf("interface creation error: %w", err)
	}

	err = common.SetupInterface(netInterface.localInterface, netInterface.address, netInterface.mtu)
	if err != nil {
		return fmt.Errorf("interface setup error: %w", err)
	}

	netInterface.logger.Info(fmt.Sprintf("Create VPN interface %s", netInterface.name))

	return err
}

func (netInterface *NetworkInterface) SetupRoutingSettings(
	remoteAddress string,
	gatewayAddress common.IpAddress,
	routeTable int,
	redirectByMark int) error {
	var err error;

	err = common.SetupRouting(
		netInterface.localInterface,
		remoteAddress,
		gatewayAddress,
		routeTable)
	if err != nil {
		return fmt.Errorf("routing setup error: %w", err)
	}

	err = common.SetIpRule(routeTable, redirectByMark)
	if err != nil {
		return fmt.Errorf("ip rule setup error: %w", err)
	}

	return err
}

func (netInterface *NetworkInterface) WriteTo(tunnelWriter func(common.IpAddress, []byte) error) error {

	buf := make([]byte, netInterface.mtu)
	for {
		n, err := netInterface.localInterface.Read(buf)
		if err != nil {
			return err
		}

		packet := buf[:n]
		sourceIp, destIp, err := common.GetIpFromPacket(packet)
		if err != nil {
			netInterface.logger.Debug(fmt.Sprintf("Unable to get IP from paket: %s", err))
			continue
		}

		writeErr := tunnelWriter(destIp, packet)
		if writeErr != nil {
			netInterface.logger.Debug(fmt.Sprintf("[%s -> %s] Unable to write to tunnel: %v",
				sourceIp.String(),
				destIp.String(),
				writeErr))
		} else {
			netInterface.logger.Debug(fmt.Sprintf("[%s -> %s] Write to tunnel, package len: %d",
				sourceIp.String(),
				destIp.String(),
				n))
		}
	}
}