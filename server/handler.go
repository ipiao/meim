package server

// 处理器
type HandlerPlugin interface {
	RenewOnlineCount() (map[string]int32, error)
}

type DefaultHandler struct {
}

func (h *DefaultHandler) RenewOnlineCount() (map[string]int32, error) {
	return nil, nil
}

var Handler HandlerPlugin = &DefaultHandler{}
