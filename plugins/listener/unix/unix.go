package unix

import (
	"net"

	"github.com/ipiao/meim/comect"
)

func init() {
	comect.RegisterMakeListener("unix", unixMakeListener)
}

func unixMakeListener(cfg *comect.ListenerConfig) (ln net.Listener, err error) {
	laddr, err := net.ResolveUnixAddr("unix", cfg.Address)
	if err != nil {
		return nil, err
	}
	return net.ListenUnix("unix", laddr)
}
