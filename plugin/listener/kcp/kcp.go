package kcp

import (
	"errors"
	"net"

	meim "github.com/ipiao/meim"
	kcp "github.com/xtaci/kcp-go"
)

func init() {
	meim.RegisterListenerMaker("kcp", kcpListenerMaker)
}

func kcpListenerMaker(cfg *meim.NetworkConfig) (ln net.Listener, err error) {
	if cfg.Options == nil || cfg.Options["BlockCrypt"] == nil {
		return nil, errors.New("KCP BlockCrypt must be configured in cfg.Options")
	}
	return kcp.ListenWithOptions(cfg.Address, cfg.Options["BlockCrypt"].(kcp.BlockCrypt), 10, 3)
}

// WithBlockCrypt option
func WithBlockCrypt(blockCrypt kcp.BlockCrypt) meim.ConfigOption {
	return meim.WithOptions(map[string]interface{}{
		"BlockCrypt": blockCrypt,
	})
}
