package utp

import (
	"net"

	"github.com/anacrolix/utp"
	"github.com/ipiao/meim"
)

func init() {
	meim.RegisterMakeListener("utp", utpMakeListener)
}

func utpMakeListener(cfg *meim.ListenerConfig) (ln net.Listener, err error) {
	return utp.Listen(cfg.Address)
}
