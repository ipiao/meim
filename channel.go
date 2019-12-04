package meim

import (
	"bufio"
	"sync"

	"github.com/ipiao/meim.v2/protocol"
)

var (
	ProtoReady  = &protocol.Message{}
	ProtoFinish = &protocol.Message{}
)

// Channel 消息推送器推送的时候使用，将消息传递给连接的写线程
type Channel struct {
	Room     *Room //如果room不是nil，代表是房间的channel
	CliProto Ring
	signal   chan *protocol.Message
	Writer   bufio.Writer
	Reader   bufio.Reader
	Next     *Channel
	Prev     *Channel

	Mid      int64
	Key      string
	IP       string
	watchOps map[int32]struct{}
	mutex    sync.RWMutex
}

// NewChannel 创建一个通道
// cli 消息缓冲环大小
// svr 信号接收缓冲
func NewChannel(cli, svr int) *Channel {
	c := new(Channel)
	c.CliProto.Init(cli)
	c.signal = make(chan *protocol.Message, svr)
	c.watchOps = make(map[int32]struct{})
	return c
}

// Watch 监听一些操作
func (c *Channel) Watch(accepts ...int32) {
	c.mutex.Lock()
	for _, op := range accepts {
		c.watchOps[op] = struct{}{}
	}
	c.mutex.Unlock()
}

// UnWatch 取消监听操作
func (c *Channel) UnWatch(accepts ...int32) {
	c.mutex.Lock()
	for _, op := range accepts {
		delete(c.watchOps, op)
	}
	c.mutex.Unlock()
}

// NeedPush 验证操作是否被监听
func (c *Channel) NeedPush(op int32) bool {
	c.mutex.RLock()
	if _, ok := c.watchOps[op]; ok {
		c.mutex.RUnlock()
		return true
	}
	c.mutex.RUnlock()
	return false
}

// Push 推送一条消息
func (c *Channel) Push(p *protocol.Message) (err error) {
	select {
	case c.signal <- p:
	default:
	}
	return
}

// Ready 检查通道是否就绪或关闭
func (c *Channel) Ready() *protocol.Message {
	return <-c.signal
}

// Signal 发送就绪信号到通道
func (c *Channel) Signal() {
	//c.signal <- rpc.ProtoReady
}

// Close 发送关闭信号到通道
func (c *Channel) Close() {
	//c.signal <- rpc.ProtoFinish
}
