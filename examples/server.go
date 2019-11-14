package main

import (
	meim "github.com/ipiao/meim.v2"
	"github.com/ipiao/meim.v2/protocol"
	"github.com/ipiao/meim/log"
)

func main() {
	s := meim.NewServer("0.0.0.0:1234")

	plugin := &meim.Plugin{
		HeaderCreator: protocol.DefaultHeaderCreator,
		MessageHandler: func(client *meim.Client, message *protocol.Message) {
			log.Infof("client %s, handle message %s", client.Log(), message.Log())
			if message.Header.Seq() == 10 {
				client.EnqueueMessage(&protocol.Message{})
			}
		},
	}
	s.RegisterPlugin(plugin)
	s.Run()
}
