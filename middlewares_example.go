package meim

//var (
//	exc MessageExchanger = new(ExchangerExample)
//)

// just a example
type ExchangerExample struct {
	MessageBroker
	*Router
	Pusher
}

// 举个例子
func (exc *ExchangerExample) DispatchMessage(msg *InternalMessage) {
	if msg.Receiver > 0 {
		clients := exc.FindClientSet(msg.Receiver)
		if len(clients) == 0 { // 不在线,推送消息
			exc.PushMessage(msg)
		}
		for client := range clients {
			client.EnqueueMessage(&msg.Message)
		}
	} else {
		// TODO,消息群发布
	}
}

func (exc *ExchangerExample) PublishMessage(msg *InternalMessage) {
	if exc.IsOnline(msg.Receiver) {
		exc.DispatchMessage(msg)
	} else {
		if msg.Receiver > 0 {
			if exc.SendMessage(msg) != nil {
				exc.PushMessage(msg)
			}
		} else {
			// TODO,消息群发布
		}
	}
}
