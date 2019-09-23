package regmq

import (
	"fmt"
	"testing"

	"github.com/ipiao/meim/plugins/dc"
)

func TestRediQ(t *testing.T) {
	keyFunc := func(uid int64) string {
		return fmt.Sprintf("USER_NODE:%d", uid)
	}
	redi := NewRedisRegistry2("127.0.0.1:6379", "", 1, keyFunc)
	mqcfg := &RabbitBrokerConfig{
		Url:      "amqp://scote:Be1sElJjlvDW@127.0.0.1:5672",
		Node:     1,
		Channels: ChannelPub | ChannelSub,
	}
	dc := dc.NewDataCreator()
	mq := NewRabbitBroker(mqcfg, dc, nil)

	rmb := NewRegisterMQ(redi, mq)
	rmb.Connect()

	rmb.Close()
}
