package server

import "github.com/ipiao/meim/protocol"

// 处理器
type HandlerPlugin interface {
	RenewOnlineCount() (map[string]int32, error)

	// authChannel完成必要到信息认证之后，还要对Channel对应的用户信息进行设置
	// Mid,rid,accepts,key
	AuthChannel(*Channel, *protocol.Proto) bool // 认证channel
}

type DefaultHandler struct {
}

func (h *DefaultHandler) RenewOnlineCount() (map[string]int32, error) {
	return nil, nil
}
func (h *DefaultHandler) AuthChannel(*Channel, *protocol.Proto) bool {
	return true
}

var Handler HandlerPlugin = &DefaultHandler{}
