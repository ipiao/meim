package meim

import (
	"net"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	lncfg := &ListenerConfig{
		Network: "tcp",
		Address: "0.0.0.0:1234",
	}

	s := NewServerWithConfig(lncfg)
	s.plugin = NewExternalImp()
	//SetExternalPlugin(NewExternalImp())
	s.Run()
}

func TestClient(t *testing.T) {
	conn, _ := net.Dial("tcp", "127.0.0.1:1234")
	time.Sleep(time.Second * 5)
	conn.Close()
}
