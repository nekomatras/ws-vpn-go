package wstunnel

import (
	"sync"
	"ws-vpn-go/common"
	"github.com/gorilla/websocket"
)

type ConnectionRegister struct {
	mutex    sync.RWMutex
	register map[common.IpAddress]*websocket.Conn
}

func NewConnectionRegister() ConnectionRegister {
	return ConnectionRegister{
		register: make(map[common.IpAddress]*websocket.Conn),
	}
}

func (register *ConnectionRegister) Add(key common.IpAddress, value *websocket.Conn) {
	register.mutex.Lock()
	defer register.mutex.Unlock()
	register.register[key] = value
}

func (register *ConnectionRegister) Remove(key common.IpAddress) {
	register.mutex.Lock()
	defer register.mutex.Unlock()
	delete(register.register, key)
}

func (register *ConnectionRegister) Contains(key common.IpAddress) bool {
	register.mutex.RLock()
	defer register.mutex.RUnlock()
	_, exist := register.register[key]
	return exist
}

func (register *ConnectionRegister) Get(key common.IpAddress) (*websocket.Conn, bool) {
	register.mutex.RLock()
	defer register.mutex.RUnlock()
	tun, res := register.register[key]
	return tun, res
}

func (register *ConnectionRegister) Update(key common.IpAddress, value *websocket.Conn) bool {
	register.mutex.Lock()
	defer register.mutex.Unlock()
	_, res := register.register[key]

	if !res {
		return false
	}

	register.register[key] = value

	return true
}