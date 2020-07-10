package meim

import (
	"crypto/tls"
	"net"

	"github.com/ipiao/meim/conf"

	"github.com/ipiao/meim/log"
)

// listener
type (
	ListenerCreator func(s *conf.Network, addr string) (ln net.Listener, err error)
)

// 监听器创建方法
var protocolListenerCreators = make(map[string]ListenerCreator)

// RegisterListenerMaker 注册监听器制造函数
func RegisterListenerCreator(protocol string, ml ListenerCreator) {
	if _, ok := protocolListenerCreators[protocol]; ok {
		log.Warnf("protocol listener of %s already exists, it will be replaced", protocol)
	}
	protocolListenerCreators[protocol] = ml
}

func LoadListenerCreator(protocol string) ListenerCreator {
	return protocolListenerCreators[protocol]
}

func init() {
	RegisterListenerCreator("tcp", makeTcpListenerCreator("tcp"))
	RegisterListenerCreator("tcp4", makeTcpListenerCreator("tcp4"))
	RegisterListenerCreator("tcp6", makeTcpListenerCreator("tcp6"))
	RegisterListenerCreator("http", makeTcpListenerCreator("tcp"))
	RegisterListenerCreator("ws", makeTcpListenerCreator("tcp"))
	RegisterListenerCreator("wss", makeTcpListenerCreator("tcp"))
}

func makeTcpListenerCreator(protocol string) ListenerCreator {
	return func(cfg *conf.Network, bind string) (ln net.Listener, err error) {
		if cfg.TlsConfig == nil {
			ln, err = net.Listen(protocol, bind)
		} else {
			ln, err = tls.Listen(protocol, bind, cfg.TlsConfig)
		}
		return ln, err
	}
}
