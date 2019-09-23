package meim

import (
	"net"
	"testing"
	"time"

	"github.com/ipiao/meim/log"
)

func TestServer(t *testing.T) {
	lncfg := &ListenerConfig{
		Network: "tcp",
		Address: "0.0.0.0:1234",
	}

	s := NewServerWithConfig(lncfg)
	SetExternalPlugin(NewExternalImp())
	s.Run()
}

type plugins struct {
	c map[Conn]chan struct{}
}

func (p *plugins) HandleConnAccepted(conn Conn) {
	log.Info("conn accepted")
	p.c[conn] = make(chan struct{})

	go func() {
		time.Sleep(time.Second * 3)
		log.Info("3 seconds later...")
		p.HandleCloseConn(conn)
	}()

	<-p.c[conn]
}

func (p *plugins) HandleConnClosed(conn Conn) {
	log.Info("conn closed")
}

func (p *plugins) HandleCloseConn(conn Conn) {
	log.Info("close conn")
	p.c[conn] <- struct{}{}
}

func TestClient(t *testing.T) {
	conn, _ := net.Dial("tcp", "127.0.0.1:1234")
	time.Sleep(time.Second * 5)
	conn.Close()
}
