package meim

import (
	"time"

	"github.com/ipiao/meim/log"

	"github.com/ipiao/meim/protocol"
)

// 处理器
type HandlerPlugin interface {
	RenewOnlineCount() (map[string]int32, error)

	// authChannel完成必要到信息认证之后，还要对Channel对应的用户信息进行设置
	// Mid,rid,accepts,key
	HandleAuth(*Channel, *protocol.Proto) (mid int64, key, rid string, accepts []int32, hb time.Duration, err error) // 认证channel
	HandleProto(*Channel, *protocol.Proto) error                                                                     // 处理channel 收到的消息
	HandleClosed(*Channel)                                                                                           // 处理channel 收到的消息
}

type DefaultHandler struct {
}

func (h *DefaultHandler) RenewOnlineCount() (map[string]int32, error) {
	log.Info("=====RenewOnlineCount=====")
	return nil, nil
}
func (h *DefaultHandler) HandleAuth(ch *Channel, p *protocol.Proto) (mid int64, key, rid string, accepts []int32, hb time.Duration, err error) {
	log.Infof("=====HandleAuth=====, channel: %s, proto:", ch, p)

	return
}
func (h *DefaultHandler) HandleProto(ch *Channel, p *protocol.Proto) error {
	log.Infof("=====HandleProto=====, channel: %s, proto:", ch, p)
	return nil
}
func (h *DefaultHandler) HandleClosed(ch *Channel) {
	log.Infof("=====HandleClosed=====, channel: %s", ch)
	return
}

var Handler HandlerPlugin = &DefaultHandler{}
