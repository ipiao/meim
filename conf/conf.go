package conf

import (
	"crypto/tls"

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

var Conf = &Config{}
