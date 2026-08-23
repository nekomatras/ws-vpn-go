package wstunnel

import (
	"sync"
	"ws-vpn-go/common"
	"github.com/gorilla/websocket"
)

type registeredConnection struct {
	conn  *websocket.Conn
	token string
}

type ConnectionRegister struct {
	mutex    sync.RWMutex
	register map[common.IpAddress]registeredConnection
}

func NewConnectionRegister() ConnectionRegister {
	return ConnectionRegister{
		register: make(map[common.IpAddress]registeredConnection),
	}
}

// Add reserves an ip for the given session token, replacing any previous
// reservation (and its connection) for that ip. This is called once per
// successful /register request, so re-registering rotates the token and
// invalidates whatever session previously held the ip.
func (register *ConnectionRegister) Add(key common.IpAddress, token string) {
	register.mutex.Lock()
	defer register.mutex.Unlock()
	register.register[key] = registeredConnection{conn: nil, token: token}
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

// CheckToken verifies that token is the session token issued for key by the
// most recent /register call, using a constant-time comparison.
func (register *ConnectionRegister) CheckToken(key common.IpAddress, token string) bool {
	register.mutex.RLock()
	defer register.mutex.RUnlock()
	entry, exist := register.register[key]
	if !exist {
		return false
	}
	return common.CheckToken(entry.token, token)
}

func (register *ConnectionRegister) Get(key common.IpAddress) (*websocket.Conn, bool) {
	register.mutex.RLock()
	defer register.mutex.RUnlock()
	entry, res := register.register[key]
	return entry.conn, res
}

func (register *ConnectionRegister) Update(key common.IpAddress, value *websocket.Conn) bool {
	register.mutex.Lock()
	defer register.mutex.Unlock()
	entry, res := register.register[key]

	if !res {
		return false
	}

	entry.conn = value
	register.register[key] = entry

	return true
}