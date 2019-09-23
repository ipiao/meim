package gate

import (
	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/protocol"
	"github.com/ipiao/meim/server"
)

// server和gate 的外部函数实现
// example
//
type ExternalImp struct {
	*Router
	conns          map[server.Conn]*Client
	handlers       map[int]func(*Client, *protocol.Message) // 处理函数,按cmd
	onAuthClient   func(*Client)
	onClientClosed func(*Client)
	defaultHandler func(*Client, *protocol.Message) // 当cmd处理函数未被注册时候,统一处理
}

func NewExternalImp() *ExternalImp {
	return &ExternalImp{
		Router:   NewRouter(),
		conns:    make(map[server.Conn]*Client),
		handlers: make(map[int]func(*Client, *protocol.Message)),
	}
}

func (e *ExternalImp) HandleConnAccepted(conn server.Conn) {
	client := NewClient(conn, nil)

	e.mu.Lock()
	e.conns[conn] = client
	e.mu.Unlock()

	exts.HandleAuthClient(client)
	if client.dc == nil {
		log.Errorf("conn %s DC not set", client.Log())
		return
	}
	client.Run()
}

func (e *ExternalImp) HandleCloseConn(conn server.Conn) {
	e.mu.RLock()
	client, ok := e.conns[conn]
	e.mu.RUnlock()
	if ok {
		client.Close()
	}
}

func (e *ExternalImp) HandleConnClosed(conn server.Conn) {
	e.mu.RLock()
	client, ok := e.conns[conn]
	e.mu.RUnlock()
	if ok {
		exts.HandleClientClosed(client)
		e.mu.Lock()
		delete(e.conns, conn)
		e.mu.Unlock()
	}
}

func (e *ExternalImp) GetConnClient(conn server.Conn) *Client {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.conns[conn]
}

func (e *ExternalImp) HandleAuthClient(client *Client) {
	if e.onAuthClient != nil {
		e.onAuthClient(client)
	}
}

func (e *ExternalImp) HandleMessage(client *Client, msg *protocol.Message) {
	if h, ok := e.handlers[msg.Header.Cmd()]; ok {
		h(client, msg)
	} else {
		if e.defaultHandler != nil {
			e.defaultHandler(client, msg)
		} else {
			log.Warnf("unsupported msg, cmd : %d", msg.Header.Cmd())
		}
	}
}

func (e *ExternalImp) HandleClientClosed(client *Client) {
	if e.onClientClosed != nil {
		e.onClientClosed(client)
	}
}

func (e *ExternalImp) SetOnAuthClient(h func(*Client)) {
	if e.onAuthClient != nil {
		log.Warnf("onAuthClient already set, will be replaced")
	}
	e.onAuthClient = h
}

func (e *ExternalImp) SetMsgHandler(cmd int, h func(*Client, *protocol.Message)) {
	if _, ok := e.handlers[cmd]; ok {
		log.Warnf("cmd %d handler already exists, will be replaced", cmd)
	}
	e.handlers[cmd] = h
}

func (e *ExternalImp) SetOnClientClosed(h func(*Client)) {
	if e.onClientClosed != nil {
		log.Warnf("onClientClosed already set, will be replaced")
	}
	e.onClientClosed = h
}

func (e *ExternalImp) Copy() *ExternalImp {
	return &ExternalImp{
		Router:         e.Router,
		conns:          e.conns,
		handlers:       e.handlers,
		onAuthClient:   e.onAuthClient,
		onClientClosed: e.onClientClosed,
		defaultHandler: e.defaultHandler,
	}
}

//var eimp *ExternalImp
//
//func init() {
//	eimp = NewExternalImp()
//	server.SetExternalPlugin(eimp)
//	gate.SetExternalPlugin(eimp)
//}
