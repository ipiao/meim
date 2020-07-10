package meim

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/ipiao/meim/libs/bufio"
	"github.com/ipiao/meim/libs/bytes"
	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/protocol"
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

	CliProto Ring                 // 客户单数据环，Ring控制了读写差异，不至于在某条消息的处理结果一直得不到返回的情况下，读处理还在继续
	signal   chan *protocol.Proto // 信号流，主动推

	conn net.Conn // 原始网络连接

	// rb、wb仅作为记录以待回收，不要使用他们
	rb     *bytes.Buffer
	wb     *bytes.Buffer
	Writer bufio.Writer
	Reader bufio.Reader

	Key      string
	IP       string
	Mid      int64
	watchOps map[int32]struct{}
	mutex    sync.RWMutex

	CID int             // channel的连接索引，用于轮训索引
	Ctx context.Context // 用户上下文
}

func (c *Channel) String() string {
	return fmt.Sprintf("mid: %d,cid: %s,key %s, addr: %s", c.Mid, c.CID, c.Key, c.conn.RemoteAddr().String())
}

// NewChannel 创建一个通道
// cli 消息缓冲环大小
// svr 信号接收缓冲
func NewChannel(ctx context.Context, conn net.Conn, rb, wb *bytes.Buffer, cid, srvProto, cliProto int) *Channel { // 不带指针，利于GC
	c := new(Channel)
	c.CliProto.Init(cliProto)
	c.signal = make(chan *protocol.Proto, srvProto)
	c.watchOps = make(map[int32]struct{})
	c.CID = cid
	c.Ctx = ctx
	c.conn = conn
	c.rb = rb
	c.wb = wb
	c.Reader.ResetBuffer(conn, rb.Bytes())
	c.Writer.ResetBuffer(conn, wb.Bytes())
	c.IP, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
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
		log.Warnf("signal channel is full: %s", c.String())
	}
	return
}

// Ready 检查通道是否就绪或关闭
func (c *Channel) Ready() *protocol.Proto {
	return <-c.signal
}

// Signal 发送就绪信号到通道
func (c *Channel) Signal() {
	c.signal <- SignalReady
}

// Close 发送关闭信号到通道
func (c *Channel) Close() {
	c.signal <- SignalFinish
}
