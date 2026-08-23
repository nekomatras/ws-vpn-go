package wstunnel

import (
	"sync"
	"ws-vpn-go/common"
	"github.com/gorilla/websocket"
)

type registeredConnection struct {
	conn  *websocket.Conn
	token string
	// epoch bumps on every Add/Update call. A scheduled expiry check
	// captures the epoch at disconnect time and only acts if it's still
	// current, so a reconnect (or a later disconnect) that happened in
	// the meantime cancels it implicitly.
	epoch uint64
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
	previous := register.register[key]
	register.register[key] = registeredConnection{conn: nil, token: token, epoch: previous.epoch + 1}
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

// Update attaches (or detaches, if value is nil) a connection for an
// existing reservation. It returns the entry's new epoch, which the caller
// can hand to RemoveIfStale to schedule an expiry check for this specific
// disconnect event.
func (register *ConnectionRegister) Update(key common.IpAddress, value *websocket.Conn) (uint64, bool) {
	register.mutex.Lock()
	defer register.mutex.Unlock()
	entry, res := register.register[key]

	if !res {
		return 0, false
	}

	entry.conn = value
	entry.epoch++
	register.register[key] = entry

	return entry.epoch, true
}

// RemoveIfStale removes the reservation for key if it is still on the given
// epoch and still has no live connection, i.e. nothing (a reconnect or a
// newer disconnect) has touched it since the epoch was captured. Returns
// whether the reservation was actually removed.
func (register *ConnectionRegister) RemoveIfStale(key common.IpAddress, epoch uint64) bool {
	register.mutex.Lock()
	defer register.mutex.Unlock()
	entry, exist := register.register[key]

	if !exist || entry.epoch != epoch || entry.conn != nil {
		return false
	}

	delete(register.register, key)
	return true
}