package networkmanager

import (
	"fmt"
	"net"
	"sync"

	"ws-vpn-go/common"
)

type Networkmanager struct {
	mutex    sync.RWMutex
	addressPool []common.IpAddress

	assignedMacByIp map[common.IpAddress]common.MacAddress
	assignedIpByMac map[common.MacAddress]common.IpAddress
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
		addressPool: availableAddresses,
		assignedMacByIp: make(map[common.IpAddress]common.MacAddress),
		assignedIpByMac: make(map[common.MacAddress]common.IpAddress),
	}, nil
}

func (manager *Networkmanager) FreeAddress(ip common.IpAddress) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	mac, ok := manager.assignedMacByIp[ip]
	if !ok {
		return
	}

	delete(manager.assignedMacByIp, ip)
	delete(manager.assignedIpByMac, mac)
}

func (manager *Networkmanager) AssignRouterAddress() (common.IpAddress, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	for _, address := range manager.addressPool {
		if address.D == 1 {
			manager.assignedMacByIp[address] = common.GetAllZeroMac()
			manager.assignedIpByMac[common.GetAllZeroMac()] = address
			return address, nil
		}
	}

	return common.GetAllZeroIp(), fmt.Errorf("unable to assign X.X.X.1 address")
}

func (manager *Networkmanager) GetAddress(mac common.MacAddress) (common.IpAddress, bool) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	ip, ok := manager.assignedIpByMac[mac]
	if ok {
		return ip, true
	} else {
		return common.GetAllZeroIp(), false
	}
}

func (manager *Networkmanager) AssignAddress(mac common.MacAddress) (common.IpAddress, error) {

	manager.mutex.RLock()

	ip, ok := manager.assignedIpByMac[mac]
	if ok {
		return ip, nil
	}

	availableAddress, err := manager.findAvailableAddress()
	if err != nil {
		return common.GetAllZeroIp(), fmt.Errorf("unable to get ip address: %v", err)
	}

	manager.mutex.RUnlock()
	manager.mutex.Lock()

	_, contains := manager.assignedMacByIp[availableAddress]
	if !contains {
		manager.assignedMacByIp[availableAddress] = mac
		manager.assignedIpByMac[mac] = availableAddress

		manager.mutex.Unlock()
		return availableAddress, nil
	}

	manager.mutex.Unlock()
	return manager.AssignAddress(mac)
}

func (manager *Networkmanager) findAvailableAddress() (common.IpAddress, error) {
	for _, ip := range manager.addressPool {
		_, contains := manager.assignedMacByIp[ip]
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