package gate

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/protocol"
	"github.com/ipiao/meim/server"
	"go.uber.org/atomic"
)

const (
	MESSAGE_QUEUE_LIMIT = 1000
)

type Connection struct {
	conn server.Conn

	closed    atomic.Bool            // 是否关闭
	mch       chan *protocol.Message // 一般消息下发通道, message channel
	lmsch     chan int               // 长消息下发(信号),long message signal channel
	lmessages *list.List             // 长消息存储队列(非阻塞消息下发)
	extch     chan func(*Client)     // 外部时间队列, external event channel
	dc        protocol.DataCreator   //
	mu        sync.Mutex             // 锁

	uid      int64       // 用户id
	userdata interface{} // 用户数据
}

func (c *Connection) Log() string {
	return fmt.Sprintf("uid %d, addr %s", c.uid, c.conn.RemoteAddr())
}

func (c *Connection) String() string {
	return fmt.Sprintf(" uid: %d, addr: %s, data: %+v", c.uid, c.conn.RemoteAddr(), c.userdata)
}

// 发送一般消息
func (c *Connection) EnqueueMessage(msg *protocol.Message) bool {
	if c.closed.Load() { // 已关闭
		log.Infof("can't send message to closed connection %s", c.Log())
		return false
	}

	select {
	case c.mch <- msg:
		return true
	case <-time.After(time.Second * 60):
		log.Infof("send message to mch timed out %s", c.Log())
		return false
	}
}

// 发送非阻塞消息
func (c *Connection) EnqueueNonBlockMessage(msg *protocol.Message) bool {
	if c.closed.Load() { // 已关闭
		log.Infof("can't send message to closed connection %s", c.Log())
		return false
	}

	dropped := false
	c.mu.Lock()
	if c.lmessages.Len() >= MESSAGE_QUEUE_LIMIT {
		//队列阻塞，丢弃之前的消息
		c.lmessages.Remove(c.lmessages.Front())
		dropped = true
	}

	c.lmessages.PushBack(msg)
	c.mu.Unlock()

	if dropped {
		log.Info("connection %s message queue full, drop a message", c.Log())
	}

	//nonblock
	select {
	case c.lmsch <- 1:
	default:
	}
	return true
}

// 发送一般消息
func (c *Connection) FlushMessage() {
	if c.closed.Load() {
		return
	}
	for msg := range c.mch {
		protocol.WriteMessage(c.conn, msg)
	}
	c.SendLMessages()
}

//发送等待队列中的消息
func (c *Connection) SendLMessages() {
	var messages *list.List
	c.mu.Lock()
	if c.lmessages.Len() == 0 {
		c.mu.Unlock()
		return
	}
	messages = c.lmessages
	c.lmessages = list.New()
	c.mu.Unlock()

	e := messages.Front()
	for e != nil {
		msg := e.Value.(*protocol.Message)
		protocol.WriteMessage(c.conn, msg)
		e = e.Next()
	}
}
