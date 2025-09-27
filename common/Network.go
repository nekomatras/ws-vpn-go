package common

import (
	"fmt"
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

