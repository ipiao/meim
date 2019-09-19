package gate

import "github.com/ipiao/meim/protocol"

// 分布式用户消息路由
type MessageBroken interface {
	Connect()                                        // 连接
	Subscribe(uid int64)                             // 注册用户
	UnSubscribe(uid int64)                           // 注销用户
	SendMessage(msg *protocol.InternalMessage) error // 发布消息
	ReceiveMessage() *protocol.InternalMessage       // 接收消息
}
