package kcp

import (
	"errors"
	"net"

	"github.com/ipiao/meim/comect"
	"github.com/xtaci/kcp-go"
)

func init() {
	comect.RegisterMakeListener("kcp", kcpMakeListener)
}

func kcpMakeListener(cfg *comect.ListenerConfig) (ln net.Listener, err error) {
	if cfg.Options == nil || cfg.Options["BlockCrypt"] == nil {
		return nil, errors.New("KCP BlockCrypt must be configured in cfg.Options")
	}

	return kcp.ListenWithOptions(cfg.Address, cfg.Options["BlockCrypt"].(kcp.BlockCrypt), 10, 3)
}

func WrapBlockCryptOption(s *comect.Server, blockCrypt kcp.BlockCrypt) {
	comect.WithOptions(map[string]interface{}{
		"BlockCrypt": blockCrypt,
	})(s)
}
