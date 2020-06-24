package server

import "github.com/ipiao/meim/protocol"

// 处理器
type HandlerPlugin interface {
	RenewOnlineCount() (map[string]int32, error)

	// authChannel完成必要到信息认证之后，还要对Channel对应的用户信息进行设置
	// Mid,rid,accepts,key
	HandleAuth(*Channel, *protocol.Proto) bool   // 认证channel
	HandleProto(*Channel, *protocol.Proto) error // 处理channel 收到的消息
	HandleClosed(*Channel)                       // 处理channel 收到的消息
}

type DefaultHandler struct {
}

func (h *DefaultHandler) RenewOnlineCount() (map[string]int32, error) {
	return nil, nil
}
func (h *DefaultHandler) HandleAuth(*Channel, *protocol.Proto) bool {
	return true
}
func (h *DefaultHandler) HandleProto(*Channel, *protocol.Proto) error {
	return nil
}
func (h *DefaultHandler) HandleClosed(*Channel) {
	return
}

var Handler HandlerPlugin = &DefaultHandler{}
