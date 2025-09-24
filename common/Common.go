package common

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/songgao/water"

	"github.com/gorilla/websocket"
)

type Interface = *water.Interface
type Tunel = *websocket.Conn

func CreateInterface(name string) (Interface, error) {
	parameters := water.PlatformSpecificParams{}

	if name == "" {
		parameters.Name = name
	}

	ifce, err := water.New(water.Config{
		DeviceType:             water.TUN,
		PlatformSpecificParams: parameters,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create address to interface: %w", err)
	}

	return ifce, nil
}

func SetupInterface(iface Interface, address string) error {

	err := exec.Command("ip", "addr", "add", address, "dev", iface.Name()).Run()
	if err != nil {
		return fmt.Errorf("Failed to add address to interface %s: %w", iface.Name(), err)
	}

	err = exec.Command("ip", "link", "set", "dev", iface.Name(), "up").Run()
	if err != nil {
		return fmt.Errorf("Failed to add address to interface %s: %w", iface.Name(), err)
	}

	return nil
}

type IpAddress struct {
	A uint8
	B uint8
	C uint8
	D uint8
}

func MakeIpAddress(a uint8, b uint8, c uint8, d uint8) IpAddress {
	return IpAddress{a, b, c, d}
}

func GetIpFromString(ipString string) IpAddress {

	ip, _, err := net.ParseCIDR(ipString)
	if err != nil {
		return GetAllZeroIp()
	}

	ipBytes := ip.To4()
	if ipBytes == nil {
		return GetAllZeroIp()
	}

	return IpAddress{ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3]}
}

func GetAllZeroIp() IpAddress {
	return IpAddress{}
}

func GetIpFromPacket(packet []byte) (IpAddress, IpAddress, error) {

	if len(packet) < 20 { // минимальный размер IPv4 заголовка
		return IpAddress{}, IpAddress{}, fmt.Errorf("Package length < 20")
	}

	// IPv4
	ipHeaderLen := int(packet[0]&0x0F) * 4 // размер заголовка в байтах
	if len(packet) < ipHeaderLen {
		return IpAddress{}, IpAddress{}, fmt.Errorf("Package length < ipHeaderLen")
	}

	// Адрес назначения: байты 16-19
	destIP := MakeIpAddress(packet[16], packet[17], packet[18], packet[19])
	// Адрес источника: байты 12-15 (если нужно)
	srcIP := MakeIpAddress(packet[12], packet[13], packet[14], packet[15])

	return srcIP, destIP, nil
}

func (ip IpAddress) String() string {
    return fmt.Sprintf("%d.%d.%d.%d", ip.A, ip.B, ip.C, ip.D)
}

type MacAddress struct {
	A uint8
	B uint8
	C uint8
	D uint8
	E uint8
	F uint8
}