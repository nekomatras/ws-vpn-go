package common

import (
	"fmt"
	"strings"
	"strconv"
)

type MacAddress struct {
	A uint8
	B uint8
	C uint8
	D uint8
	E uint8
	F uint8
}

func GetMacFromString(mac string) MacAddress {
	parts := strings.Split(mac, ":")
	if len(parts) != 6 {
		return GetAllZeroMac()
	}

	bytes := make([]uint8, 6)
	for i, part := range parts {
		val, err := strconv.ParseUint(part, 16, 8)
		if err != nil {
			return GetAllZeroMac()
		}
		bytes[i] = uint8(val)
	}

	return MacAddress{
		A: bytes[0],
		B: bytes[1],
		C: bytes[2],
		D: bytes[3],
		E: bytes[4],
		F: bytes[5],
	}
}

func GetAllZeroMac() MacAddress {
	return MacAddress{}
}

func (m MacAddress) String() string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
		m.A, m.B, m.C, m.D, m.E, m.F)
}