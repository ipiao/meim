package exchanger

import (
	"github.com/ipiao/meim"
	"github.com/ipiao/meim/log"
)

// 消息交换机
// TODO add feedback for not dispatched or not published
type Exchanger struct {
	meim.MessageBroker                            //
	meim.Pusher                                   //
	router             *meim.Router               // for example,取 ExternalImp的Router
	pubCh              chan *meim.InternalMessage // publish chan
}

func New(router *meim.Router, broker meim.MessageBroker, pusher meim.Pusher) *Exchanger {
	if pusher == nil {
		pusher = new(NopPusher)
	}
	return &Exchanger{
		MessageBroker: broker,
		Pusher:        pusher,
		router:        router,
		pubCh:         make(chan *meim.InternalMessage, 128),
	}
}

func (exc *Exchanger) handleRead(closedCh chan bool) {
	for {
		msg, err := exc.ReceiveMessage()
		if err == nil || msg == nil {
			log.Infof("exchanger receive err or nil message: err: %s,msg: %v", err, msg)
			close(closedCh)
			return
		}
		if msg.Receiver != 0 { // 给给发送的用户
			client := exc.router.FindClient(msg.Receiver)
			client.EnqueueEvent(func(c *meim.Client) {
				meim.HandleMessage(c, msg.Message)
			})
		} else { // 没有指定用户的,就认为是系统消息,或者通知消息

		}
	}
}

func (exc *Exchanger) handleWrite(closedCh chan bool) {

	for {
		select {
		case _ = <-closedCh:
			log.Info("exchanger closed")
			return
		case msg := <-exc.pubCh:
			exc.PublishMessage(msg)
		}
	}
}

// 举个例子
// within handle message
// 单纯的进行消息下发,未考虑业务消息
func (exc *Exchanger) DispatchMessage(msg *meim.InternalMessage) bool {
	// TODO 使用goroutin池
	// 用go避免阻塞
	client := exc.router.FindClient(msg.Receiver)
	if client == nil {
		return exc.PushMessage(msg)
	} else {
		go client.EnqueueMessage(msg.Message)
	}
	return true
}

func (exc *Exchanger) DispatchClientMessage(client *meim.Client, msg *meim.InternalMessage) {
	// TODO 使用goroutin池
	// 用go避免阻塞
	go client.EnqueueMessage(msg.Message)
}

func (exc *Exchanger) PublishMessage(msg *meim.InternalMessage) bool {
	if exc.SendMessage(msg) != nil {
		return exc.PushMessage(msg)
	}
	return true
}
