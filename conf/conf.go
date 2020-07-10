package conf

import (
	"crypto/tls"
	"time"

	xtime "github.com/ipiao/meim/libs/time"
)

// 提炼conf好单独管理
// Config is comet config.
type Config struct {
	Debug    bool // 用作开启调试
	ServerID string

	Protocol *Protocol
	Networks []*Network
	Bucket   *Bucket
	Round    *Round
	Channel  *Channel
}

type Protocol struct {
	HandshakeTimeout     xtime.Duration
	MinHeartbeatInterval xtime.Duration
	MaxHeartbeatInterval xtime.Duration
}

type Network struct {
	Protocol  string // 协议类型, tcp,ws
	Bind      []string
	TlsConfig *tls.Config // tls配置
	Sndbuf    int
	Rcvbuf    int
	KeepAlive bool
	Accept    int
}

// bucket 初始化配置可选项
type Bucket struct {
	// Bucket is bucket config.
	Size    int // bucket数量
	Channel int // channel初始化数量

	Room          int    // room初始化数量
	RoutineAmount uint64 // room的后台推送线程数量
	RoutineSize   int    // room每个推送线程的缓冲大小
}

type Round struct {
	Timer        int
	TimerSize    int
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

type Channel struct {
	SvrProto int
	CliProto int
}

var Conf = Default()

// Default new a config with specified defualt value.
func Default() *Config {
	return &Config{
		Debug:    true,
		ServerID: "1",
		Protocol: &Protocol{
			HandshakeTimeout:     xtime.Duration(time.Second * 15),
			MinHeartbeatInterval: xtime.Duration(time.Second * 10),
			MaxHeartbeatInterval: xtime.Duration(time.Second * 30),
		},
		Networks: []*Network{{
			Protocol:  "tcp",
			Bind:      []string{"0.0.0.0:5678"},
			Sndbuf:    4096,
			Rcvbuf:    4096,
			KeepAlive: false,
			Accept:    1,
		}},
		Bucket: &Bucket{
			Size:          2,
			Channel:       128,
			Room:          16,
			RoutineAmount: 4,
			RoutineSize:   64,
		},
		Round: &Round{
			Timer:        8,
			TimerSize:    1024,
			Reader:       32,
			ReadBuf:      1024,
			ReadBufSize:  8192,
			Writer:       32,
			WriteBuf:     1024,
			WriteBufSize: 8192,
		},
		Channel: &Channel{
			SvrProto: 16,
			CliProto: 16,
		},
	}
}
