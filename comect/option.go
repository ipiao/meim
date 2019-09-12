package comect

import (
	"crypto/tls"
	"time"
)

// OptionFn configures options of server.
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
func WithTLSConfig(tlscfg *tls.Config) OptionFn {
	return func(s *Server) {
		s.lncfg.TLSConfig = tlscfg
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

// WithWriteTimeout sets writeTimeout.
func WithNetwork(network string) OptionFn {
	return func(s *Server) {
		s.lncfg.Network = network
	}
}
