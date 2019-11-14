package meim_v2

import (
	"crypto/tls"
	"net"

	"github.com/ipiao/meim/log"
)

// listener
type ListenerMaker func(s *Config) (ln net.Listener, err error)

// 监听器创建方法
var listenerMakers = make(map[string]ListenerMaker)

// RegisterListenerMaker 注册监听器制造函数
func RegisterListenerMaker(network string, ml ListenerMaker) {
	if _, ok := listenerMakers[network]; ok {
		log.Warnf("listener maker of %s already exists, it will be replaced", network)
	}
	listenerMakers[network] = ml
}

func init() {
	RegisterListenerMaker("tcp", tcpListenerMaker("tcp"))
	RegisterListenerMaker("tcp4", tcpListenerMaker("tcp4"))
	RegisterListenerMaker("tcp6", tcpListenerMaker("tcp6"))
	RegisterListenerMaker("http", tcpListenerMaker("tcp"))
}

func tcpListenerMaker(network string) ListenerMaker {
	return func(cfg *Config) (ln net.Listener, err error) {
		if cfg.TlsConfig == nil {
			ln, err = net.Listen(network, cfg.Address)
		} else {
			ln, err = tls.Listen(network, cfg.Address, cfg.TlsConfig)
		}
		return ln, err
	}
}
