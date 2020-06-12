package quic

import (
	"errors"
	"net"

	meim "github.com/ipiao/meim"
	quicconn "github.com/marten-seemann/quic-conn"
)

func init() {
	meim.RegisterListenerMaker("quic", quicListenerMaker)
}

func quicListenerMaker(cfg *meim.Config) (ln net.Listener, err error) {
	if cfg.TlsConfig == nil {
		return nil, errors.New("TlsConfig must be configured in cfg")
	}
	return quicconn.Listen("udp", cfg.Address, cfg.TlsConfig)
}
