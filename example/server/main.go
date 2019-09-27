package main

import (
	"fmt"
	"time"

	"github.com/ipiao/meim"
	"github.com/ipiao/meim/plugins/broker/regmq"
	"github.com/ipiao/meim/plugins/dc"
)

// 需要实现的东西
var (
	DC     meim.DataCreator  // 数据协议,消息创建器
	eimp   *meim.ExternalImp // 插件实现
	exc    *meim.Exchanger   // 消息交换机
	router *meim.Router
)

type InternalMessageHandler struct {
	*meim.ExternalImp
}

func (h *InternalMessageHandler) HandleInternalMessage(msg *meim.InternalMessage) {

}

func keyFunc(uid int64) string {
	return fmt.Sprintf("USER_NODE:%d", uid)
}

func main() {
	router = meim.NewRouter()
	rabbit := regmq.NewRabbitMQ(&regmq.RabbitMQConfig{
		ExchangeName: "meim",
		ExchangeKind: "topic",
		Url:          "amqp://scote:Be1sElJjlvDW@127.0.0.1:5672",
		RPCTimeout:   time.Second * 5,
		SendTimeout:  time.Second * 3,
		ChanSize:     512,
		Channels:     regmq.ChannelPub | regmq.ChannelSub,
		QueuePrefix:  "message",
		Node:         1,
	}, DC, nil)
	reg := regmq.NewRedisRegistry2("127.0.0.1", "6379", 1, "USER_NODE")
	broker := regmq.NewRegisterMQ(reg, rabbit)
	exc = meim.NewMessageExchanger(broker, nil, &InternalMessageHandler{eimp}, router)

	DC = dc.NewProtoDataCreator()
	eimp = meim.NewExternalImp()
	eimp.SetDefaultHandler(func(client *meim.Client, msg *meim.Message) { // 写回
		client.EnqueueMessage(msg)
	})
	eimp.SetOnAuthClient(func(client *meim.Client) bool {
		client.DC = DC
		return true
	})
}
