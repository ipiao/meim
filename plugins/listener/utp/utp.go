package utp

import (
	"net"

	"github.com/anacrolix/utp"
	"github.com/ipiao/meim/server"
)

func init() {
	server.RegisterMakeListener("utp", utpMakeListener)
}

func utpMakeListener(cfg *server.ListenerConfig) (ln net.Listener, err error) {
	return utp.Listen(cfg.Address)
}
