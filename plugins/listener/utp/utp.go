package utp

import (
	"net"

	"github.com/anacrolix/utp"
	"github.com/ipiao/meim/comect"
)

func init() {
	comect.RegisterMakeListener("utp", utpMakeListener)
}

func utpMakeListener(cfg *comect.ListenerConfig) (ln net.Listener, err error) {
	return utp.Listen(cfg.Address)
}
