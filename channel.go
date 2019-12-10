package meim

import (
	"io"
	"sync"

	"github.com/ipiao/meim.v2/log"
	"github.com/ipiao/meim.v2/protocol"
)

var (
	SignalReady  = &protocol.Proto{}
	SignalFinish = &protocol.Proto{}
)

// Channel 消息推送器推送的时候使用，将消息传递给连接的写线程
// 相当于client
type Channel struct {
	Room *Room //如果room不是nil，代表是房间的channel

	Next *Channel // 暂时只有Room使用
	Prev *Channel // 暂时只有Room使用

	CliProto Ring
	signal   chan *protocol.Proto
	Writer   io.Writer
	Reader   io.Reader

	Key      string
	IP       string
	Mid      int64
	watchOps map[int32]struct{}
	mutex    sync.RWMutex
}

// NewChannel 创建一个通道
// cli 消息缓冲环大小
// svr 信号接收缓冲
func NewChannel(cli, svr int) *Channel {
	c := new(Channel)
	c.CliProto.Init(cli)
	c.signal = make(chan *protocol.Proto, svr)
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
func (c *Channel) Push(p *protocol.Proto) (err error) {
	select {
	case c.signal <- p:
	default:
		log.Warnf("signal channel is full")
	}
	return
}

// Ready 检查通道是否就绪或关闭
func (c *Channel) Signal() *protocol.Proto {
	return <-c.signal
}

// Signal 发送就绪信号到通道
func (c *Channel) Ready() {
	c.signal <- SignalReady
}

// Close 发送关闭信号到通道
func (c *Channel) Close() {
	c.signal <- SignalFinish
}
