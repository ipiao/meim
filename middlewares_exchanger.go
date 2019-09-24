package meim

import (
	"time"

	"github.com/ipiao/meim/log"
)

// 消息交换机,处理消息的对内分发和对外分发
// TODO add feedback for not dispatched or not published
type Exchanger struct {
	MessageBroker                                //
	Pusher                                       //
	InternalMessageHandler                       //
	router                 *Router               // for example,取 ExternalImp的Router
	pubCh                  chan *InternalMessage //
}

func NewMessageExchanger(router *Router, broker MessageBroker, pusher Pusher, handler InternalMessageHandler) *Exchanger {
	if pusher == nil {
		pusher = new(NopPusher)
	}
	return &Exchanger{
		MessageBroker:          broker,
		Pusher:                 pusher,
		router:                 router,
		InternalMessageHandler: handler,
		pubCh:                  make(chan *InternalMessage, 256),
	}
}

// 直接下发
// 单纯的进行消息下发,未考虑业务消息
func (exc *Exchanger) DispatchMessage(msg *InternalMessage) bool {
	return exc.DispatchMessage(msg)
}

func (exc *Exchanger) dispatchMessage(msg *InternalMessage) bool {
	// TODO 使用goroutin池
	// 用go避免阻塞
	client := exc.router.FindClient(msg.Receiver)
	if client == nil {
		return exc.PushMessage(msg)
	} else {
		go exc.DispatchClientMessage(client, msg)
		return true
	}
}

// 对客户端直接下发
func (exc *Exchanger) DispatchClientMessage(client *Client, msg *InternalMessage) bool {
	go client.EnqueueMessage(msg.Message)
	return true
}

// 发布消息到Broker
func (exc *Exchanger) PublishMessage(msg *InternalMessage) bool {
	select {
	case exc.pubCh <- msg:
		return true
	case <-time.After(time.Second * 3):
		log.Warnf("publish message timeout: %v", msg)
	}
	return false
}

func (exc *Exchanger) publishMessage(msg *InternalMessage) bool {
	if exc.SendMessage(msg) != nil {
		return exc.PushMessage(msg)
	}
	return true
}

func (exc *Exchanger) handleRead(closedCh chan bool) {
	for {
		msg, err := exc.ReceiveMessage()
		if err == nil || msg == nil {
			log.Infof("exchanger receive err or nil message: err: %s,msg: %v", err, msg)
			close(closedCh)
			return
		}
		exc.HandleInternalMessage(msg)
	}
}

func (exc *Exchanger) handleWrite(closedCh chan bool) {
	for {
		select {
		case _ = <-closedCh:
			log.Info("exchanger closed")
			return
		case msg := <-exc.pubCh:
			exc.publishMessage(msg)
		}
	}
}

// 运行一个Exchanger
func (exc *Exchanger) Run() {
	for {
		exc.Connect()
		exc.runOnce()
	}
}

func (exc *Exchanger) runOnce() {
	defer exc.Close()
	closedCh := make(chan bool)

	go exc.handleRead(closedCh)
	exc.handleWrite(closedCh)
}
