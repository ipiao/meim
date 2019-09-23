package exchanger

import "github.com/ipiao/meim"

// 消息交换机
type Exchanger struct {
	*meim.Router                          // for example,取 ExternalImp的Router
	meim.MessageBroker                    //
	meim.Pusher                           //
	mch                chan *meim.Message // publish chan
}

func New(router *meim.Router, broker meim.MessageBroker, pusher meim.Pusher) *Exchanger {
	return &Exchanger{
		Router:        router,
		MessageBroker: broker,
		Pusher:        pusher,
	}
}

// 举个例子
// within handle message
func (exc *Exchanger) DispatchMessage(msg *meim.InternalMessage) {
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

func (exc *Exchanger) PublishMessage(msg *meim.InternalMessage) {
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
