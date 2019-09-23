package unix

import (
	"net"

	"github.com/ipiao/meim/server"
)

func init() {
	server.RegisterMakeListener("unix", unixMakeListener)
}

func unixMakeListener(cfg *server.ListenerConfig) (ln net.Listener, err error) {
	laddr, err := net.ResolveUnixAddr("unix", cfg.Address)
	if err != nil {
		return nil, err
	}
	return net.ListenUnix("unix", laddr)
}
