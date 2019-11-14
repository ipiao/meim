package utp

import (
	"net"

	"github.com/anacrolix/utp"
	meim "github.com/ipiao/meim.v2"
)

func init() {
	meim.RegisterListenerMaker("utp", utpListenerMaker)
}

func utpListenerMaker(cfg *meim.Config) (ln net.Listener, err error) {
	return utp.Listen(cfg.Address)
}
