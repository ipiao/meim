package meim

import (
	"github.com/ipiao/meim/log"
)

// 服务应该会用到的组件,插件

// 分布式用户消息交换
type MessageBroker interface {
	Connect()                                                   // 连接
	Subscribe(uid int64)                                        // 注册用户
	UnSubscribe(uid int64)                                      // 注销用户
	SendMessage(msg *InternalMessage) error                     // 发布消息
	ReceiveMessage() (*InternalMessage, error)                  // 接收消息
	SyncMessage(msg *InternalMessage) (*InternalMessage, error) // 同步消息请求
	Close()
}

// 消息推送,作为附属
type Pusher interface {
	PushMessage(msg *InternalMessage) bool
}

type NopPusher struct{}

func (p *NopPusher) PushMessage(msg *InternalMessage) bool {
	log.Infof("msg %v not pushed", msg)
	return false
}

// 本地消息分发,外部发送到本服务的消息
type Dispatcher interface {
	DispatchMessage(*InternalMessage) bool
}

// 进行消息发布
type Publisher interface {
	PublishMessage(*InternalMessage) bool
}

type InternalMessageHandler interface {
	HandleInternalMessage(msg *InternalMessage) // 和Client不一样的地方
}

// 消息交换机
type MessageExchanger interface {
	MessageBroker
	Dispatcher
	Publisher
	InternalMessageHandler
}
