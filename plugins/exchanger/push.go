package exchanger

import (
	"github.com/ipiao/meim"
	"github.com/ipiao/meim/log"
)

type NopPusher struct{}

func (p *NopPusher) PushMessage(msg *meim.InternalMessage) bool {
	log.Infof("msg %v not pushed", msg)
	return false
}
