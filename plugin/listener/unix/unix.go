package unix

import (
	"net"

	meim "github.com/ipiao/meim.v2"
)

func init() {
	meim.RegisterListenerMaker("unix", unixListenerMaker)
}

func unixListenerMaker(cfg *meim.Config) (ln net.Listener, err error) {
	laddr, err := net.ResolveUnixAddr("unix", cfg.Address)
	if err != nil {
		return nil, err
	}
	return net.ListenUnix("unix", laddr)
}
