package meim

import (
	"crypto/tls"
	"time"
)

// OptionFn 对服务的参数进行插入式配置
type OptionFn func(*Server)

// WithOptions sets multiple options.
func WithOptions(ops map[string]interface{}) OptionFn {
	return func(s *Server) {
		if s.lncfg.Options == nil {
			s.lncfg.Options = make(map[string]interface{})
		}
		for k, op := range ops {
			s.lncfg.Options[k] = op
		}
	}
}

// WithTLSConfig sets tls.Config.
func WithTLSConfig(cfg *tls.Config) OptionFn {
	return func(s *Server) {
		s.lncfg.TLSConfig = cfg
	}
}

// WithReadTimeout sets readTimeout.
func WithReadTimeout(readTimeout time.Duration) OptionFn {
	return func(s *Server) {
		s.readTimeout = readTimeout
	}
}

// WithWriteTimeout sets writeTimeout.
func WithWriteTimeout(writeTimeout time.Duration) OptionFn {
	return func(s *Server) {
		s.writeTimeout = writeTimeout
	}
}

// WithMaxConn sets maxConn.
func WithMaxConn(n int) OptionFn {
	return func(s *Server) {
		s.maxConn = n
	}
}

// WithWriteTimeout sets writeTimeout.
func WithNetwork(network string) OptionFn {
	return func(s *Server) {
		s.lncfg.Network = network
	}
}

// WithListenAddr sets listener addr.
func WithListenAddr(addr string) OptionFn {
	return func(s *Server) {
		s.lncfg.Address = addr
	}
}

// WithExternalPlugin sets server plugin
func WithExternalPlugin(plugin ExternalPlugin) OptionFn {
	return func(s *Server) {
		s.plugin = plugin
	}
}
