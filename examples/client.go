package main

import (
	"time"

	"github.com/ipiao/meim.v2/protocol"

	meim "github.com/ipiao/meim.v2"
	"github.com/ipiao/meim.v2/client"
	"github.com/ipiao/meim.v2/log"
)

func main() {
	cli, err := client.DialTCP("127.0.0.1:1234", meim.ClientConfig{}, nil)
	if err != nil {
		log.Fatal(err)
	}
	go writeTask(cli)
	cli.Run()
}

func writeTask(cli *meim.Client) {
	var seq int32
	tick := time.NewTicker(time.Second)
	for {
		select {
		case <-tick.C:
			msg := &protocol.Message{
				Header: cli.CreateHeader(),
			}
			seq++
			msg.Header.SetSeq(seq)
			cli.EnqueueMessage(msg)
		}
	}
}
