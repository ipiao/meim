package kcp

import (
	"errors"
	"net"

	"github.com/ipiao/meim"
	kcp "github.com/xtaci/kcp-go"
)

func init() {
	meim.RegisterMakeListener("kcp", kcpMakeListener)
}

func kcpMakeListener(cfg *meim.ListenerConfig) (ln net.Listener, err error) {
	if cfg.Options == nil || cfg.Options["BlockCrypt"] == nil {
		return nil, errors.New("KCP BlockCrypt must be configured in cfg.Options")
	}

	return kcp.ListenWithOptions(cfg.Address, cfg.Options["BlockCrypt"].(kcp.BlockCrypt), 10, 3)
}

func WrapBlockCryptOption(s *meim.Server, blockCrypt kcp.BlockCrypt) {
	meim.WithOptions(map[string]interface{}{
		"BlockCrypt": blockCrypt,
	})(s)
}
