package common

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/songgao/water"
)

type Interface = *water.Interface
type Tunnel = *websocket.Conn

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

	err := exec.Command("ip", "addr", "add", address, "dev", iface.Name()).Run()
	if err != nil {
		return fmt.Errorf("failed to add address to interface %s: %w", iface.Name(), err)
	}

	err = exec.Command("ip", "link", "set", "dev", iface.Name(), "mtu", fmt.Sprintf("%d", mtu)).Run()
	if err != nil {
		return fmt.Errorf("failed to set mtu on interface %s: %w", iface.Name(), err)
	}

	err = exec.Command("ip", "link", "set", "dev", iface.Name(), "up").Run()
	if err != nil {
		return fmt.Errorf("failed to add address to interface %s: %w", iface.Name(), err)
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

