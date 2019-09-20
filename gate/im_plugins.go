package gate

import (
	"github.com/ipiao/meim/protocol"
)

// 分布式用户消息交换路由
type MessageBroken interface {
	Connect()                                                                     // 连接
	Subscribe(uid int64)                                                          // 注册用户
	UnSubscribe(uid int64)                                                        // 注销用户
	SendMessage(msg *protocol.InternalMessage) error                              // 发布消息
	ReceiveMessage() (*protocol.InternalMessage, error)                           // 接收消息
	SyncMessage(msg *protocol.InternalMessage) (*protocol.InternalMessage, error) // 同步消息请求
	Close()
}

// 本地消息分发,外部发送到本服务的消息
type Dispatcher interface {
	DispatchMessage(*protocol.InternalMessage)
}

// 进行消息发布
type Publisher interface {
	PublishMessage(*protocol.InternalMessage)
}

// 消息交换机
type MessageExchanger interface {
	Dispatcher
	Publisher
	MessageBroken
}

// 消息推送,作为附属
type Pusher interface {
	PushMessage(msg *protocol.InternalMessage)
}

// example
type ExchangerExample struct {
	MessageBroken
	*Router
	pusher Pusher
}

//var (
//	exc MessageExchanger = new(ExchangerExample)
//)

// 举个例子
func (exc *ExchangerExample) DispatchMessage(msg *protocol.InternalMessage) {
	if msg.Receiver > 0 {
		clients := exc.FindClientSet(msg.Receiver)
		if len(clients) == 0 { // 不在线,推送消息
			exc.pusher.PushMessage(msg)
		}
		for client := range clients {
			client.EnqueueMessage(&msg.Message)
		}
	} else {
		// TODO,消息群发布
	}
}

func (exc *ExchangerExample) PublishMessage(msg *protocol.InternalMessage) {
	if exc.IsOnline(msg.Receiver) {
		exc.DispatchMessage(msg)
	} else {
		if msg.Receiver > 0 {
			if exc.SendMessage(msg) != nil {
				exc.pusher.PushMessage(msg)
			}
		} else {
			// TODO,消息群发布
		}
	}
}
