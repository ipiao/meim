package gate

import (
	"sync"

	"github.com/ipiao/meim/server"
)

var (
	HandleClientClosed func(*Client) // 关闭客户端之后的处理
)

type serverPlugins struct {
	mu      sync.RWMutex
	clients map[server.Conn]*Client
}

func (sp *serverPlugins) HandleConnAccepted(conn server.Conn) {
	cilent := NewClient(conn, nil)
	sp.mu.Lock()
	sp.clients[conn] = cilent
	sp.mu.Unlock()
	cilent.Run()
}

func (sp *serverPlugins) HandleCloseConn(conn server.Conn) {
	sp.mu.RLock()
	client, ok := sp.clients[conn]
	sp.mu.RUnlock()
	if ok {
		client.Close()
	}
}

func (sp *serverPlugins) HandleConnClosed(conn server.Conn) {
	sp.mu.RLock()
	client, ok := sp.clients[conn]
	sp.mu.RUnlock()
	if ok {
		HandleClientClosed(client)
		sp.mu.Lock()
		delete(sp.clients, conn)
		sp.mu.Unlock()
	}
}

var sps *serverPlugins

func init() {
	sps = &serverPlugins{
		clients: make(map[server.Conn]*Client),
	}
	// server.Plugins.SetPlugin(sps)
	server.HandleCloseConn = sps.HandleConnClosed
	server.HandleConnAccepted = sps.HandleConnAccepted
	server.HandleConnClosed = sps.HandleConnClosed
}

// func Clients() ClientSet {
// 	set := NewClientSet()

// 	sps.mu.RLock()
// 	defer sps.mu.RUnlock()

// 	for _, v := range sps.clients {
// 		set[v] = struct{}{}
// 	}
// 	return set
// }

// func HasClient(c *Client) bool {
// 	sps.mu.RLock()
// 	defer sps.mu.RUnlock()

// 	_, ok := sps.clients[c.Conn]
// 	return ok
// }
