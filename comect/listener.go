package comect

import (
	"crypto/tls"
	"net"
)

// 监听服务,适配不同的网络服务

type ListenerConfig struct {
	Network   string
	Address   string
	TLSConfig *tls.Config
	Options   map[string]interface{} // 为了在不囊括某些配置的时候,能够支持自定义listener的配置信息
}

// listener
type MakeListener func(cfg *ListenerConfig) (ln net.Listener, err error)

// 监听器注册
var makeListeners = make(map[string]MakeListener)

// RegisterMakeListener registers a MakeListener for network.
func RegisterMakeListener(network string, ml MakeListener) {
	if _, ok := makeListeners[network]; ok {

	}
	makeListeners[network] = ml
}

func init() {
	makeListeners["tcp"] = tcpMakeListener("tcp")
	makeListeners["tcp4"] = tcpMakeListener("tcp4")
	makeListeners["tcp6"] = tcpMakeListener("tcp6")
	makeListeners["http"] = tcpMakeListener("tcp")
}

func tcpMakeListener(network string) func(cfg *ListenerConfig) (ln net.Listener, err error) {
	return func(cfg *ListenerConfig) (ln net.Listener, err error) {
		if cfg.TLSConfig == nil {
			ln, err = net.Listen(network, cfg.Address)
		} else {
			ln, err = tls.Listen(network, cfg.Address, cfg.TLSConfig)
		}
		return ln, err
	}
}

func udpMakeListener(network string) func(cfg *ListenerConfig) (ln net.Listener, err error) {
	return func(cfg *ListenerConfig) (ln net.Listener, err error) {
		if cfg.TLSConfig == nil {
			ln, err = net.Listen(network, cfg.Address)
		} else {
			ln, err = tls.Listen(network, cfg.Address, cfg.TLSConfig)
		}
		return ln, err
	}
}
