package gate

import (
	"github.com/ipiao/meim/server"
)

// server 和 gate的外部函数实现

type externalImp struct {
	*Router
	conns map[server.Conn]*Client
}

func newExternalImp() *externalImp {
	return &externalImp{
		Router: NewRouter(),
		conns:  make(map[server.Conn]*Client),
	}
}

func (sp *externalImp) HandleConnAccepted(conn server.Conn) {
	cilent := NewClient(conn, nil)

	sp.mu.Lock()
	sp.conns[conn] = cilent
	sp.mu.Unlock()

	cilent.Run()
}

func (sp *externalImp) HandleCloseConn(conn server.Conn) {
	sp.mu.RLock()
	client, ok := sp.conns[conn]
	sp.mu.RUnlock()
	if ok {
		client.Close()
	}
}

func (sp *externalImp) HandleConnClosed(conn server.Conn) {
	sp.mu.RLock()
	client, ok := sp.conns[conn]
	sp.mu.RUnlock()
	if ok {
		HandleClientClosed(client)
		sp.mu.Lock()
		delete(sp.conns, conn)
		sp.mu.Unlock()
	}
}

var eimp *externalImp

func init() {
	eimp = newExternalImp()

	server.HandleCloseConn = eimp.HandleConnClosed
	server.HandleConnAccepted = eimp.HandleConnAccepted
	server.HandleConnClosed = eimp.HandleConnClosed
}
