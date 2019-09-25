package meim

import (
	"github.com/ipiao/meim/log"
)

// server和gate 的外部函数实现
// example for ExternalPlugin
//
type (
	MessageHandler func(client *Client, msg *Message)
)

type ExternalImp struct {
	defaultHandler MessageHandler         // 当cmd处理函数未被注册时候,统一处理
	handlers       map[int]MessageHandler // 处理函数,按cmd
	onAuthClient   func(*Client) bool     // 处理客户端认证
	onClientClosed func(*Client)          //
}

func NewExternalImp() *ExternalImp {
	return &ExternalImp{
		//Router:   NewRouter(),
		handlers: make(map[int]MessageHandler),
	}
}

func (e *ExternalImp) HandleAuthClient(client *Client) bool {
	if e.onAuthClient == nil {
		return false
	}
	// 现在至少要求DC不为nil
	if !e.onAuthClient(client) || client.DC == nil {
		return false
	}
	return true
}

func (e *ExternalImp) HandleMessage(client *Client, msg *Message) {
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

func (e *ExternalImp) SetOnAuthClient(h func(*Client) bool) {
	if e.onAuthClient != nil {
		log.Warnf("onAuthClient already set, will be replaced")
	}
	e.onAuthClient = h
}

func (e *ExternalImp) SetMsgHandler(cmd int, h MessageHandler) {
	if _, ok := e.handlers[cmd]; ok {
		log.Warnf("cmd %d handler already exists, will be replaced", cmd)
	}
	e.handlers[cmd] = h
}

func (e *ExternalImp) SetDefauleHandler(h MessageHandler) {
	if e.defaultHandler != nil {
		log.Warnf("defaultHandler already set, will be replaced")
	}
	e.defaultHandler = h
}

func (e *ExternalImp) SetOnClientClosed(h func(*Client)) {
	if e.onClientClosed != nil {
		log.Warnf("onClientClosed already set, will be replaced")
	}
	e.onClientClosed = h
}

func (e *ExternalImp) Copy() *ExternalImp {
	return &ExternalImp{
		//Router:         e.Router,
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
//	SetExternalPlugin(eimp)
//	gate.SetExternalPlugin(eimp)
//}
