package main

import (
	"net"

	"github.com/ipiao/meim/protocol"

	"github.com/ipiao/meim/log"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:5678")
	if err != nil {
		log.Fatal(err)
	}
	protocol.WriteTo(conn, &protocol.Proto{
		Ver:         1,
		Op:          protocol.OpAuth,
		Seq:         1,
		Compress:    0,
		ContentType: 0,
		Body:        nil,
	})

	select {}
}
