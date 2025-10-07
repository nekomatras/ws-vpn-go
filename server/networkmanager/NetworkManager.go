package networkmanager

import (
	"fmt"
	"net"
	"sync"

	"ws-vpn-go/common"
)

type Networkmanager struct {
	mutex    sync.RWMutex
	adresses []common.IpAddress
	assignedAdresses map[common.IpAddress]struct{}
}

func New(subnet string) (*Networkmanager, error) {

	if len(subnet) >= 3 && subnet[len(subnet)-3:] != "/24" {
		return nil, fmt.Errorf("support only 255.255.255.0 or X.X.X.0/24 mask")
	}

	availableAddresses, err := getIpListBySubNet(subnet)
	if err != nil {
		return nil, err
	}

	return &Networkmanager{
		adresses: availableAddresses,
		assignedAdresses: make(map[common.IpAddress]struct{}),
	}, nil
}

func (manager *Networkmanager) FreeAddress(address common.IpAddress) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	delete(manager.assignedAdresses, address)
}

func (manager *Networkmanager) AssignRouterAddress() (common.IpAddress, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	for _, address := range manager.adresses {
		if address.D == 1 {
			manager.assignedAdresses[address] = struct{}{}
			return address, nil
		}
	}

	return common.GetAllZeroIp(), fmt.Errorf("unable to assign X.X.X.1 address")
}

func (manager *Networkmanager) AssignAddress() (common.IpAddress, error) {

	availableAddress, err := manager.findAvailableAddress()
	if err != nil {
		return common.GetAllZeroIp(), fmt.Errorf("unable to get ip address: %v", err)
	}

	manager.mutex.Lock()

	_, contains := manager.assignedAdresses[availableAddress]
	if !contains {
		manager.assignedAdresses[availableAddress] = struct{}{}

		manager.mutex.Unlock()
		return availableAddress, nil
	}

	manager.mutex.Unlock()
	return manager.AssignAddress()
}

func (manager *Networkmanager) findAvailableAddress() (common.IpAddress, error) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	for _, ip := range manager.adresses {
		_, contains := manager.assignedAdresses[ip]
		if contains {
			return ip, nil
		}
	}

	return common.GetAllZeroIp(), fmt.Errorf("no one address available")
}

func getIpListBySubNet(subnet string) ([]common.IpAddress, error) {

	ip, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}

	adresses := make([]common.IpAddress, 0, 254)

	for i := uint8(1); i < 255; i++ {
		addr := common.NewIpAddress(ip[0], ip[1], ip[2], i)
		adresses = append(adresses, addr)
	}

	return adresses, nil
}