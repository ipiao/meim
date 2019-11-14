package meim_v2

import (
	"crypto/tls"
	"time"

	"github.com/ipiao/meim.v2/log"
)

var (
	DefaultReadTimeout           = time.Minute * 15
	DefaultWriteTimeout          = time.Second * 15
	DefaultMaxReadBodySize       = 1 << 12
	DefaultMaxWriteBodySize      = 1 << 20
	DefaultMaxConnections        = 65536
	DefaultPendingWriteChanLen   = 16
	DefaultPendingEventChanLen   = 1
	DefaultMaxPendingWriteNBSize = 512
	DefaultEnqueueTimeout        = time.Second * 5
)

type Config struct {
	Network        string                 // 网络类型
	Address        string                 // 监听地址
	MaxConnections int                    // 最大连接数
	TlsConfig      *tls.Config            // tls配置
	Options        map[string]interface{} // 其他选项，支持自定义
	ClientConfig
}

func (c *Config) Init() {
	if c.Network == "" {
		c.Network = "tcp"
	}
	if c.Address == "" {
		log.Fatal("can not start server with empty listener address")
	}
	if c.MaxConnections == 0 {
		c.MaxConnections = DefaultMaxConnections
	}
	c.ClientConfig.Init()
}

type ClientConfig struct {
	ReadTimeout           time.Duration // 读超时
	WriteTimeout          time.Duration // 写超时
	MaxReadBodySize       int           // 最大读的body字长
	MaxWriteBodySize      int           // 最大写的body字长
	PendingWriteChanLen   int           // 等待写长度
	PendingEventChanLen   int           // 等待事件队列长
	MaxPendingWriteNBSize int           // 最大non-block写size
	EnqueueTimeout        time.Duration // 超时
}

func (c *ClientConfig) Init() {
	if c.ReadTimeout == 0 {
		c.ReadTimeout = DefaultReadTimeout
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = DefaultWriteTimeout
	}
	if c.MaxReadBodySize == 0 {
		c.MaxReadBodySize = DefaultMaxReadBodySize
	}
	if c.MaxWriteBodySize == 0 {
		c.MaxWriteBodySize = DefaultMaxWriteBodySize
	}
	if c.PendingWriteChanLen == 0 {
		c.PendingWriteChanLen = DefaultPendingWriteChanLen
	}
	if c.PendingEventChanLen == 0 {
		c.PendingEventChanLen = DefaultPendingEventChanLen
	}
	if c.MaxPendingWriteNBSize == 0 {
		c.MaxPendingWriteNBSize = DefaultMaxPendingWriteNBSize
	}
	if c.EnqueueTimeout == 0 {
		c.EnqueueTimeout = DefaultEnqueueTimeout
	}
}

// ConfigOption 对服务的参数进行插入式配置
type ConfigOption func(*Config)

// WithOptions 注入自定义的参数
func WithOptions(ops map[string]interface{}) ConfigOption {
	return func(c *Config) {
		if c.Options == nil {
			c.Options = make(map[string]interface{})
		}
		for k, op := range ops {
			c.Options[k] = op
		}
	}
}

// WithTLSConfig sets tls.Config.
func WithTLSConfig(cfg *tls.Config) ConfigOption {
	return func(c *Config) {
		c.TlsConfig = cfg
	}
}

// WithReadTimeout sets readTimeout.
func WithReadTimeout(readTimeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.ReadTimeout = readTimeout
	}
}

// WithWriteTimeout sets writeTimeout.
func WithWriteTimeout(writeTimeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.WriteTimeout = writeTimeout
	}
}

// WithMaxConn sets maxConn.
func WithMaxConnections(n int) ConfigOption {
	return func(c *Config) {
		c.MaxConnections = n
	}
}

// WithWriteTimeout sets writeTimeout.
func WithNetwork(network string) ConfigOption {
	return func(c *Config) {
		c.Network = network
	}
}

// WithListenAddress sets listener addr.
func WithListenAddress(addr string) ConfigOption {
	return func(c *Config) {
		c.Address = addr
	}
}
