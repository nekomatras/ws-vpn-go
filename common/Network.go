package common

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

type Interface = *water.Interface
type Tunnel = *websocket.Conn

const PRIORITY = 100;

func CreateInterface(name string) (Interface, error) {
	parameters := water.PlatformSpecificParams{}

	if name != "" {
		parameters.Name = name
	}

	ifce, err := water.New(water.Config{
		DeviceType:             water.TUN,
		PlatformSpecificParams: parameters,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create address to interface: %w", err)
	}

	return ifce, nil
}

func SetupInterface(iface Interface, address string, mtu uint) error {
	ifName := iface.Name()

	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return fmt.Errorf("failed to get link %s: %w", ifName, err)
	}

	ipNet, err := netlink.ParseAddr(address)
	if err != nil {
		return fmt.Errorf("failed to parse address %s: %w", address, err)
	}

	if err := netlink.AddrAdd(link, ipNet); err != nil {
		// EEXIST допустим, если адрес уже есть
		if !os.IsExist(err) {
			return fmt.Errorf("failed to add address to %s: %w", ifName, err)
		}
	}

	if err := netlink.LinkSetMTU(link, int(mtu)); err != nil {
		return fmt.Errorf("failed to set mtu on %s: %w", ifName, err)
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to set %s up: %w", ifName, err)
	}

	return nil
}

func getDefaultGatewayForVpnServerAddress(vpnAddress string) (IpAddress, IpAddress, error) {

	ips, err := net.LookupIP(vpnAddress)
	if err != nil {
		return GetAllZeroIp(), GetAllZeroIp(), fmt.Errorf("failed to resolve vpn address: %v", err)
	}

	var dstIP net.IP
	for _, ip := range ips {
		if ip.To4() != nil {
			dstIP = ip
			break
		}
	}
	if dstIP == nil {
		return GetAllZeroIp(), GetAllZeroIp(), fmt.Errorf("failed resolve ipv4: %v", err)
	}

	route, err := netlink.RouteGet(dstIP)
	if err != nil {
		return GetAllZeroIp(), GetAllZeroIp(), fmt.Errorf("failed get route: %v", err)
	}

	if len(route) == 0 {
		return GetAllZeroIp(), GetAllZeroIp(), fmt.Errorf("no route found")
	}

	return ConvertIpAddress(dstIP), ConvertIpAddress(route[0].Gw), nil
}

func SetIpRule(routeTable int, redirectByMark int) error {

	rule := netlink.NewRule()
	rule.Table = routeTable
	rule.Priority = PRIORITY

	if redirectByMark > 0 {
		mask := uint32(0xffffffff)
		rule.Mark = uint32(redirectByMark)
		rule.Mask = &mask
	}

	return netlink.RuleAdd(rule)
}

func SetupRouting(
	iface Interface,
	remoteAddress string,
	gatewayAddress IpAddress,
	routeTable int) error {
	var err error;
	var defaultGateway IpAddress;
	var remoteIp IpAddress;

	// ip route add default via <VPN_GATEWAY> dev <VPN_INTERFACE_NAME> table <VPN_ROUTING_TABLE>
	link, err := netlink.LinkByName(iface.Name())
	if err != nil {
		return fmt.Errorf("failed to get interface index %s: %w", iface.Name(), err)
	}

	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Gw:        gatewayAddress.GetNetIp4(),
		Table:     routeTable,
	}

	err = netlink.RouteReplace(route)
	if err != nil {
		return fmt.Errorf("failed to add default route: %w", err)
	}

	// ip route add <VPN_SERVER_ADDRESS> via <CURRENT_DEFAULT_GATEWAY> table <VPN_ROUTING_TABLE>
	defaultGateway, remoteIp, err = getDefaultGatewayForVpnServerAddress(remoteAddress)
	if err != nil {
		return fmt.Errorf("failed get default gateway: %w", err)
	}

	vpnServer := &net.IPNet{
		IP:   remoteIp.GetNetIp4(),
		Mask: net.CIDRMask(32, 32),
	}

	routeToVpnServer := &netlink.Route{
		Dst:   vpnServer,
		Gw:    defaultGateway.GetNetIp4(),
		Table: routeTable,
	}

	err = netlink.RouteReplace(routeToVpnServer)
	if err != nil {
		return fmt.Errorf("failed to add route to vpn server: %w", err)
	}

	return nil
}


func flushTable(table int) error {
	routes, err := netlink.RouteListFiltered(
		netlink.FAMILY_ALL,
		&netlink.Route{
			Table: table,
		},
		netlink.RT_FILTER_TABLE,
	)
	if err != nil {
		return err
	}

	for _, r := range routes {
		route := r
		_ = netlink.RouteDel(&route)
	}

	return nil
}

func ResetRouting(routeTable int) error {
	// ip route flush table <VPN_ROUTING_TABLE>
	err := flushTable(routeTable)
	if err != nil {
		return fmt.Errorf("failed to flush table %d: %w", routeTable, err)
	}

	return nil
}

func GetRealIP(r *http.Request) string {

    xff := r.Header.Get("X-Forwarded-For")
    if xff != "" {
        parts := strings.Split(xff, ",")
        return strings.TrimSpace(parts[0])
    }

    if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
        return xrip
    }

    return r.RemoteAddr
}

