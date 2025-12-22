package common

import (
	"net"
	"fmt"
)

type IpAddress struct {
	A uint8
	B uint8
	C uint8
	D uint8
}

func NewIpAddress(a uint8, b uint8, c uint8, d uint8) IpAddress {
	return IpAddress{a, b, c, d}
}

func GetIpFromString(ipString string) IpAddress {
	var ip net.IP;
	var err error;

	ip, _, err = net.ParseCIDR(ipString)
	if err != nil {
		ip = net.ParseIP(ipString)
		if ip == nil {
			return GetAllZeroIp()
		}
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
		return IpAddress{}, IpAddress{}, fmt.Errorf("package length < 20")
	}

	// IPv4
	ipHeaderLen := int(packet[0]&0x0F) * 4 // размер заголовка в байтах
	if len(packet) < ipHeaderLen {
		return IpAddress{}, IpAddress{}, fmt.Errorf("package length < ipHeaderLen")
	}

	// Адрес назначения: байты 16-19
	destIP := NewIpAddress(packet[16], packet[17], packet[18], packet[19])
	// Адрес источника: байты 12-15 (если нужно)
	srcIP := NewIpAddress(packet[12], packet[13], packet[14], packet[15])

	return srcIP, destIP, nil
}

func (ip IpAddress) String() string {
    return fmt.Sprintf("%d.%d.%d.%d", ip.A, ip.B, ip.C, ip.D)
}