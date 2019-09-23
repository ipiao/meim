package kcp

import (
	"errors"
	"net"

	"github.com/ipiao/meim/server"
	kcp "github.com/xtaci/kcp-go"
)

func init() {
	server.RegisterMakeListener("kcp", kcpMakeListener)
}

func kcpMakeListener(cfg *server.ListenerConfig) (ln net.Listener, err error) {
	if cfg.Options == nil || cfg.Options["BlockCrypt"] == nil {
		return nil, errors.New("KCP BlockCrypt must be configured in cfg.Options")
	}

	return kcp.ListenWithOptions(cfg.Address, cfg.Options["BlockCrypt"].(kcp.BlockCrypt), 10, 3)
}

func WrapBlockCryptOption(s *server.Server, blockCrypt kcp.BlockCrypt) {
	server.WithOptions(map[string]interface{}{
		"BlockCrypt": blockCrypt,
	})(s)
}
