package gate

import (
	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/server"
)

// server 的外部函数实现
//
// 在自己的项目中可以通过继承ExternalImp进行函数重写

type ExternalImp struct {
	*Router
	conns map[server.Conn]*Client
}

func NewExternalImp() *ExternalImp {
	return &ExternalImp{
		Router: NewRouter(),
		conns:  make(map[server.Conn]*Client),
	}
}

func (eimp *ExternalImp) HandleConnAccepted(conn server.Conn) {
	client := NewClient(conn, nil)

	eimp.mu.Lock()
	eimp.conns[conn] = client
	eimp.mu.Unlock()

	exts.HandleAuthClient(client)
	if client.dc == nil {
		log.Errorf("conn %s DC not set", client.Log())
		return
	}
	client.Run()
}

func (eimp *ExternalImp) HandleCloseConn(conn server.Conn) {
	eimp.mu.RLock()
	client, ok := eimp.conns[conn]
	eimp.mu.RUnlock()
	if ok {
		client.Close()
	}
}

func (eimp *ExternalImp) HandleConnClosed(conn server.Conn) {
	eimp.mu.RLock()
	client, ok := eimp.conns[conn]
	eimp.mu.RUnlock()
	if ok {
		exts.HandleClientClosed(client)
		eimp.mu.Lock()
		delete(eimp.conns, conn)
		eimp.mu.Unlock()
	}
}

//var eimp *ExternalImp
//
//func init() {
//	eimp = NewExternalImp()
//	server.SetExternalPlugin(eimp)
//}
