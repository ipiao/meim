package gate

import (
	"container/list"
	"sync"

	"github.com/ipiao/meim/protocol"
	"github.com/ipiao/meim/server"
)

type Client struct {
	conn server.Conn

	closed    bool                   // 是否关闭
	mch       chan *protocol.Message // 一般消息下发通道, message channel
	lmsch     chan int               // 长消息下发(信号),long message signal channel
	lmessages *list.List             // 长消息存储队列
	extch     chan func(*Client)     // 外部时间队列, external event channel
	mu        sync.Mutex             // 锁
}

func NewClient(conn server.Conn, mchSize, extchSize, readMax int) *Client {
	return &Client{
		conn:      conn,
		mch:       make(chan *protocol.Message),
		lmsch:     make(chan int, 1),
		lmessages: list.New(),
		extch:     make(chan func(*Client), extchSize),
	}
}

func (c *Client) read() {
	for {
		// protocol.ReadLimitMessage(c.conn, 32*1024)
	}
}

func (c *Client) write() {

}
