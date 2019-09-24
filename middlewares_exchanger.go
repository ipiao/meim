package meim

var (
	exc MessageExchanger
)

func SendMessage(sender, receiver int64, message *Message) {
	//interMessage := &InternalMessage{
	//	Message:   message,
	//	Sender:    sender,
	//	Receiver:  receiver,
	//	Timestamp: util.NowMillSec(),
	//}

}
