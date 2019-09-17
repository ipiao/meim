package gate

import (
	"sync"

	"github.com/ipiao/meim/server"
)

type serverPlugins struct {
	sync.RWMutex
	clients map[server.Conn]*Client
}

func (*serverPlugins) HandleConnAccepted(server.Conn) {

}

func (*serverPlugins) HandleConnClosed(server.Conn) {

}

func (*serverPlugins) HandleCloseConn(server.Conn) {

}

func (*serverPlugins) HandleWrite(server.Conn) {

}

func init() {
	server.Plugins.SetPlugin(&serverPlugins{})
}
