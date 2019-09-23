package unix

import (
	"net"

	"github.com/ipiao/meim"
)

func init() {
	meim.RegisterMakeListener("unix", unixMakeListener)
}

func unixMakeListener(cfg *meim.ListenerConfig) (ln net.Listener, err error) {
	laddr, err := net.ResolveUnixAddr("unix", cfg.Address)
	if err != nil {
		return nil, err
	}
	return net.ListenUnix("unix", laddr)
}
