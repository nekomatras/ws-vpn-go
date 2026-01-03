package networkmanager

import (
	"fmt"
	"net"
	"sync"

	"ws-vpn-go/common"
)

type SyncMap[K comparable, V any] struct {
	mutex sync.RWMutex
	data  map[K]V
}

func MakeSyncMap[K comparable, V any]() SyncMap[K, V] {
	return SyncMap[K, V]{
		data: make(map[K]V),
	}
}

func (s *SyncMap[K, V]) Get(key K) (V, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	val, ok := s.data[key]
	return val, ok
}

func (s *SyncMap[K, V]) Set(key K, value V) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[key] = value
}

func (s *SyncMap[K, V]) Delete(key K) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.data, key)
}

// --------------------------------------------------------------------

type Networkmanager struct {
	// mutex       sync.RWMutex
	addressPool []common.IpAddress
	subNet uint16

	assignedMacByIp SyncMap[common.IpAddress, common.MacAddress]
	assignedIpByMac SyncMap[common.MacAddress, common.IpAddress]
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
		addressPool:     availableAddresses,
		subNet: 24,
		assignedMacByIp: MakeSyncMap[common.IpAddress, common.MacAddress](),
		assignedIpByMac: MakeSyncMap[common.MacAddress, common.IpAddress](),
	}, nil
}

func (manager *Networkmanager) FreeAddress(ip common.IpAddress) {
	mac, ok := manager.assignedMacByIp.Get(ip)
	if !ok {
		return
	}

	manager.assignedMacByIp.Delete(ip)
	manager.assignedIpByMac.Delete(mac)
}

func (manager *Networkmanager) AssignRouterAddress() (common.IpAddress, error) {
	for _, address := range manager.addressPool {
		if address.D == 1 {
			manager.assignedMacByIp.Set(address, common.GetAllZeroMac())
			manager.assignedIpByMac.Set(common.GetAllZeroMac(), address)
			return address, nil
		}
	}

	return common.GetAllZeroIp(), fmt.Errorf("unable to assign X.X.X.1 address")
}

func (manager *Networkmanager) GetSubNet() uint16 {
	return manager.subNet
}

func (manager *Networkmanager) GetAddress(mac common.MacAddress) (common.IpAddress, bool) {
	ip, ok := manager.assignedIpByMac.Get(mac)
	if ok {
		return ip, true
	} else {
		return common.GetAllZeroIp(), false
	}
}

func (manager *Networkmanager) AssignAddress(mac common.MacAddress) (common.IpAddress, error) {
	ip, ok := manager.assignedIpByMac.Get(mac)
	if ok {
		return ip, nil
	}

	availableAddress, err := manager.findAvailableAddress()
	if err != nil {
		return common.GetAllZeroIp(), fmt.Errorf("unable to get ip address: %v", err)
	}

	_, contains := manager.assignedMacByIp.Get(availableAddress)
	if !contains {
		manager.assignedMacByIp.Set(availableAddress, mac)
		manager.assignedIpByMac.Set(mac, availableAddress)

		return availableAddress, nil
	}

	return manager.AssignAddress(mac)
}

func (manager *Networkmanager) findAvailableAddress() (common.IpAddress, error) {
	for _, ip := range manager.addressPool {
		_, contains := manager.assignedMacByIp.Get(ip)
		if !contains {
			return ip, nil
		}
	}

	return common.GetAllZeroIp(), fmt.Errorf("no one address available")
}

func getIpListBySubNet(subnet string) ([]common.IpAddress, error) {

	rawIp, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}
	ip := common.ConvertIpAddress(rawIp)

	adresses := make([]common.IpAddress, 0, 254)

	for i := uint8(1); i < 255; i++ {
		ip.D = i
		addr := ip
		adresses = append(adresses, addr)
	}

	return adresses, nil
}
