package regmq

import (
	"errors"

	"github.com/ipiao/meim"
)

var (
	ErrorUserNodeNotFound = errors.New("user node not found")
)

// Register-Broker
type RegisterMQ struct {
	reg Registry //
	mq  MQBroker //
}

// 注册登录
type Registry interface {
	Register(uid int64, node int) // 暂时只是允许单端单点登录,在注册登录前业务要自行查证唯一性
	DeRegister(uid int64)         // 取消注册
	GetUserNode(uid int64) int    // 获取用户node
	Close()
}

type MQBroker interface {
	Node() int
	ReceiveMessage() *meim.InternalMessage
	SendMessage(node int, msg *meim.InternalMessage) error
	SyncMessage(node int, msg *meim.InternalMessage) (*meim.InternalMessage, error)
	Close()
}

func (tr *RegisterMQ) Connect() {}

func (tr *RegisterMQ) Subscribe(uid int64) {
	tr.reg.Register(uid, tr.mq.Node())
}

func (tr *RegisterMQ) UnSubscribe(uid int64) {
	tr.reg.DeRegister(uid)
}

func (tr *RegisterMQ) SendMessage(msg *meim.InternalMessage) error {
	node := tr.reg.GetUserNode(msg.Receiver)
	if node == 0 {
		return ErrorUserNodeNotFound
	}
	return tr.mq.SendMessage(node, msg)
}

func (tr *RegisterMQ) ReceiveMessage() (*meim.InternalMessage, error) {
	return tr.mq.ReceiveMessage(), nil
}

func (tr *RegisterMQ) SyncMessage(msg *meim.InternalMessage) (*meim.InternalMessage, error) {
	node := tr.reg.GetUserNode(msg.Receiver)
	if node == 0 {
		return nil, ErrorUserNodeNotFound
	}
	return tr.mq.SyncMessage(node, msg)
}

func (tr *RegisterMQ) Close() {
	tr.reg.Close()
	tr.mq.Close()
}

func NewRegisterMQ(reg Registry, mq MQBroker) *RegisterMQ {
	return &RegisterMQ{
		reg: reg,
		mq:  mq,
	}
}
