package gate

import "github.com/ipiao/meim/protocol"

// 分布式用户路由
type DistributedRouter interface {
	Register(uid, node int64)   // 注册用户节点
	DeRegister(uid, node int64) // 注销用户节点

	SendMessage(msg *ExchangeMessage) // 发布消息
	ReceiveMessage() *ExchangeMessage // 接收消息
}

// 交换消息
type ExchangeMessage struct {
	protocol.Message       // 发送的消息体
	Receiver         int64 // 接收人
	Sender           int64 // 发送人
	Timestamp        int64 // 时间戳,ms
}
