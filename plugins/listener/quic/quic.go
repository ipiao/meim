package quic

import (
	"errors"
	"net"

	"github.com/ipiao/meim/server"
	quicconn "github.com/marten-seemann/quic-conn"
)

func init() {
	server.RegisterMakeListener("quic", quicMakeListener)
}

func quicMakeListener(cfg *server.ListenerConfig) (ln net.Listener, err error) {
	if cfg.TLSConfig == nil {
		return nil, errors.New("TLSConfig must be configured in cfg.Options")
	}
	return quicconn.Listen("udp", cfg.Address, cfg.TLSConfig)
}
