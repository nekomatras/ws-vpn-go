package wstunnel

import (
	"io"
	"fmt"
	"net/http"
	"log/slog"
	"sync"
	"time"

	"ws-vpn-go/common"

	"github.com/gorilla/websocket"
)

const InitialBackoff = 1 * time.Second
const MaxBackoff = 30 * time.Second

type WsTunnel struct {
	remoteURL string
	key string

	clientIp     common.IpAddress
	sessionToken string

	mutex    sync.Mutex
	wsTunnel *websocket.Conn

	logger *slog.Logger
}

func New(remoteAddress string, tunnelPath string, key string, logger *slog.Logger) *WsTunnel {
	return &WsTunnel{
		remoteURL: "wss://" + remoteAddress + tunnelPath,
		key: key,
		logger: logger,
	}
}

func (tunnel *WsTunnel) Run() error {
	tunnel.reconnect()
	return nil
}

func (tunnel *WsTunnel) WriteTo(target io.Writer) error {
	for {
		conn := tunnel.reconnect()

		_, message, err := conn.ReadMessage()
		if err != nil {
			tunnel.logger.Warn(fmt.Sprintf("WebSocket read error: %v; reconnecting", err))
			tunnel.invalidate(conn)
			continue
		}

		_, err = target.Write(message)
		if err != nil {
			tunnel.logger.Error(fmt.Sprintf("Interface write error: %v", err))
			return fmt.Errorf("interface write error: %w", err)
		}
	}
}

func (tunnel *WsTunnel) WriteToTunnel(_ common.IpAddress, packet []byte) error {
	conn := tunnel.reconnect()

	if err := conn.WriteMessage(websocket.BinaryMessage, packet); err != nil {
		tunnel.invalidate(conn)
		return fmt.Errorf("web-socket write error: %w", err)
	}

	return nil
}

func (tunnel *WsTunnel) RegisterHandlers(mux *http.ServeMux) error {
	errMsg := "method RegisterHandlers not implemented"
	tunnel.logger.Error("Not implementod method called: " + errMsg)
	return fmt.Errorf("%s", errMsg)
}

func (tunnel *WsTunnel) ReserveConnection(ip common.IpAddress, token string) error {

	if tunnel.clientIp != common.GetAllZeroIp() {
		return fmt.Errorf(
			"unable to set client ip address [%s], its already seted to %s", ip.String(), tunnel.clientIp.String())
	}

	if ip == common.GetAllZeroIp() {
		return fmt.Errorf(
			"unable to set client ip address [%s]", ip.String())
	}

	tunnel.clientIp = ip
	tunnel.sessionToken = token
	return nil
}

func (tunnel *WsTunnel) SetConnectionCloseHandler(handler func (common.IpAddress)) {
	errMsg := "method SetConnectionCloseHandler not implemented"
	tunnel.logger.Error("Not implementod method called: " + errMsg)
}

// reconnect returns the current live connection, dialing a new one if
// necessary. Concurrent callers (the read loop in WriteTo and the write
// path in WriteToTunnel both hit this on a dropped connection) share the
// same in-flight dial instead of racing on tunnel.wsTunnel: only the first
// caller actually dials, the rest block on the mutex and reuse its result.
// It retries indefinitely with exponential backoff so a transient network
// blip (wifi switch, sleep/wake, brief outage) recovers on its own.
func (tunnel *WsTunnel) reconnect() *websocket.Conn {
	tunnel.mutex.Lock()
	defer tunnel.mutex.Unlock()

	if tunnel.wsTunnel != nil {
		return tunnel.wsTunnel
	}

	backoff := InitialBackoff
	attempt := 0
	for {
		attempt++
		conn, err := tunnel.dial()
		if err == nil {
			tunnel.wsTunnel = conn
			return conn
		}

		tunnel.logger.Warn(fmt.Sprintf(
			"[%s] Unable to connect tunnel (attempt %d): %v; retry in %s",
			tunnel.remoteURL, attempt, err, backoff))

		time.Sleep(backoff)
		backoff *= 2
		if backoff > MaxBackoff {
			backoff = MaxBackoff
		}
	}
}

// invalidate drops the cached connection if it is still the one that just
// failed, so the next reconnect() call dials a fresh one. Guarded so a
// stale failure from an already-replaced connection can't clobber a newer
// live one.
func (tunnel *WsTunnel) invalidate(conn *websocket.Conn) {
	tunnel.mutex.Lock()
	defer tunnel.mutex.Unlock()

	if tunnel.wsTunnel == conn {
		conn.Close()
		tunnel.wsTunnel = nil
	}
}

func (tunnel *WsTunnel) dial() (*websocket.Conn, error) {
	header := make(http.Header)
	header.Add("Key", tunnel.key)
	header.Add("ClientIP", tunnel.clientIp.String())
	header.Add("SessionToken", tunnel.sessionToken)

	conn, _, err := websocket.DefaultDialer.Dial(tunnel.remoteURL, header)
	if err != nil {
		return nil, fmt.Errorf("web-socket connection error: %w", err)
	}

	return conn, nil
}