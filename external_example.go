package meim

import (
	"github.com/ipiao/meim/log"
)

// server和gate 的外部函数实现
// example for ExternalPlugin
//
type (
	MessageHandler func(client *Client, msg *Message)
	Filter         func(MessageHandler) MessageHandler
)

type ExternalImp struct {
	defaultHandler MessageHandler // 当cmd处理函数未被注册时候,统一处理
	defaultFilters []Filter
	handlers       map[int]MessageHandler // 处理函数,按cmd
	onAuthClient   func(*Client) bool     // 处理客户端认证
	onClientClosed func(*Client)          //
	beforeWrite    MessageHandler
}

func NewExternalImp() *ExternalImp {
	return &ExternalImp{
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

func (e *ExternalImp) HandleBeforeWriteMessage(client *Client, message *Message) {
	if e.beforeWrite != nil {
		e.beforeWrite(client, message)
	}
}

func (e *ExternalImp) SetOnAuthClient(h func(*Client) bool) {
	if e.onAuthClient != nil {
		log.Warnf("onAuthClient already set, will be replaced")
	}
	e.onAuthClient = h
}

func (e *ExternalImp) SetMsgHandler(cmd int, h MessageHandler, filters ...Filter) {
	if _, ok := e.handlers[cmd]; ok {
		log.Warnf("cmd %d handler already exists, will be replaced", cmd)
	}
	e.handlers[cmd] = filterHandler(h, append(filters, e.defaultFilters...))
}

func (e *ExternalImp) SetDefaultHandler(h MessageHandler, filters ...Filter) {
	e.defaultFilters = append(filters, e.defaultFilters...)
	if e.defaultHandler != nil {
		log.Warnf("defaultHandler already set, will be replaced")
	}
	e.defaultHandler = filterHandler(h, e.defaultFilters)
}

func (e *ExternalImp) SetBeforeWrite(h MessageHandler) {
	if e.beforeWrite != nil {
		log.Warnf("beforeWrite already set, will be replaced")
	}
	e.beforeWrite = h
}

func (e *ExternalImp) SetOnClientClosed(h func(*Client)) {
	if e.onClientClosed != nil {
		log.Warnf("onClientClosed already set, will be replaced")
	}
	e.onClientClosed = h
}

func (e *ExternalImp) Clone() *ExternalImp {
	imp := &ExternalImp{
		//Router:         e.Router,
		onAuthClient:   e.onAuthClient,
		onClientClosed: e.onClientClosed,
		defaultHandler: e.defaultHandler,
		beforeWrite:    e.beforeWrite,
	}
	handlers := make(map[int]MessageHandler)
	for cmd, h := range e.handlers {
		handlers[cmd] = h
	}
	var filters []Filter
	for _, filter := range e.defaultFilters {
		filters = append(filters, filter)
	}
	imp.defaultFilters = filters
	imp.handlers = handlers
	return imp
}

func filterHandler(h MessageHandler, filters []Filter) MessageHandler {
	if h == nil {
		h = func(client *Client, msg *Message) {
			// do nothing
		}
	}
	ret := h
	for _, filter := range filters {
		ret = filter(ret)
	}
	return ret
}

//// 通过反射注册
//func (e *ExternalImp) RegisterHandler(cmd int, i interface{}, filters ...Filter) {
//	if i == nil {
//		e.SetMsgHandler(cmd, nil, filters...)
//		return
//	}
//	fn := reflect.ValueOf(i)
//	if fn.Kind() != reflect.Func {
//		panic("invalid handler")
//	}
//	if fn.Type().NumIn() != 2 {
//		panic("invalid args number of handler")
//	}
//	// 限定参数类型
//	if fn.Type().In(0) != reflect.TypeOf(&Client{}) ||
//		!fn.Type().In(1).Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {
//		panic("invalid args type of handler")
//	}
//	if fn.Type().NumOut() > 0 {
//		for i := 0; i < fn.Type().NumOut(); i++ {
//			if !fn.Type().Out(i).Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {
//				panic("invalid return type of handler")
//			}
//		}
//	}
//	f := func(client *Client, msg *Message) {
//		body := msg.Body.(*dc.ProtoData)
//		in := []reflect.Value{
//			reflect.ValueOf(client),
//			reflect.ValueOf(body.Message),
//		}
//		outs := fn.Call(in)
//		for _, out := range outs {
//			var ahdr ProtocolHeader
//			ahdr = msg.Header.Clone()
//			acmd, ok := client.DC.GetCmd2(out.Type())
//			if !ok {
//				log.Warnf("write unregistered msg: %v", msg)
//			}
//			ahdr.SetCmd(acmd)
//			amsg := &Message{
//				Header: ahdr,
//				Body:   dc.NewProtoData(out.Interface().(proto.Message)),
//			}
//			client.EnqueueMessage(amsg)
//		}
//	}
//	e.SetMsgHandler(cmd, f, filters...)
//}

//var eimp *ExternalImp
//
//func init() {
//	eimp = NewExternalImp()
//	SetExternalPlugin(eimp)
//	gate.SetExternalPlugin(eimp)
//}

// filter 防止panic
// just for example
func FilterRecover(fn MessageHandler) MessageHandler {
	return func(client *Client, msg *Message) {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("[handler][recover][%s] client: %s, request: %+v, info:\n %v",
					client.DC.GetDescription(msg.Header.Cmd()), client, msg, r)
			}
		}()
		if msg == nil {
			log.Warnf("[handler] client %s receive nil msg", client)
			return
		}
		fn(client, msg)
	}
}
