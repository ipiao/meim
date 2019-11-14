package meim_v2

import (
	"container/list"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/ipiao/meim.v2/protocol"
	"github.com/ipiao/meim/log"
	"go.uber.org/atomic"
)

type Client struct {
	Conn
	PluginI
	cfg ClientConfig

	mch       chan *protocol.Message // 一般消息下发通道, message channel
	lMessages *list.List             // 长消息存储队列(非阻塞消息下发,能忍受延时)
	lSignal   chan int               // 长消息下发(信号),long message signal channel
	eventCh   chan func(*Client)     // 外部事件队列, external event channel
	mu        sync.Mutex             // 锁
	closed    atomic.Bool            // 是否关闭

	UID      int64       // 用户id
	Version  int16       // 版本
	UserData interface{} // 用户其他私有数据
}

// 新建客户端
func NewClient(conn Conn, cfg ClientConfig, plugin PluginI) *Client {
	return &Client{
		Conn:      conn,
		PluginI:   plugin,
		mch:       make(chan *protocol.Message, cfg.PendingWriteChanLen),
		lMessages: list.New(),
		lSignal:   make(chan int, 1),
		eventCh:   make(chan func(*Client), cfg.PendingEventChanLen),
		cfg:       cfg,
	}
}

func (client *Client) Log() string {
	return fmt.Sprintf("uid %d, addr %s", client.UID, client.RemoteAddr())
}

func (client *Client) String() string {
	return fmt.Sprintf(" uid: %d, addr: %s, data: %v", client.UID, client.RemoteAddr(), client.UserData)
}

// 发送一般消息
func (client *Client) EnqueueMessage(msg *protocol.Message) bool {
	if msg == nil { // 不能手动传 nil
		log.Warnf("can't send nil message to client: %s", client.Log())
		return false
	}
	if client.closed.Load() { // 已关闭
		log.Infof("can't send message to closed client %s", client.Log())
		return false
	}

	select {
	case client.mch <- msg:
		return true
	case <-time.After(client.cfg.EnqueueTimeout):
		log.Infof("client %s enqueue message timeout", client.Log())
		return false
	}
}

// 发送非阻塞消息
func (client *Client) EnqueueNonBlockMessage(msg *protocol.Message) bool {
	if client.closed.Load() { // 已关闭
		log.Infof("can't send message to closed client %s", client.Log())
		return false
	}

	dropped := false
	client.mu.Lock()
	if client.lMessages.Len() >= client.cfg.MaxPendingWriteNBSize {
		//队列阻塞，丢弃之前的消息
		client.lMessages.Remove(client.lMessages.Front())
		dropped = true
	}

	client.lMessages.PushBack(msg)
	client.mu.Unlock()

	if dropped {
		log.Info("client %s lMessage queue is full, drop a message", client.Log())
	}

	//nonblock
	select {
	case client.lSignal <- 1:
	default:
	}
	return true
}

//发送等待队列中的消息
func (client *Client) SendLMessages() {
	var messages *list.List
	client.mu.Lock()
	if client.lMessages.Len() == 0 {
		client.mu.Unlock()
		return
	}
	messages = client.lMessages
	client.lMessages = list.New()
	client.mu.Unlock()

	e := messages.Front()
	for e != nil {
		msg := e.Value.(*protocol.Message)
		protocol.WriteLimitMessage(client.Conn, msg, client.cfg.MaxWriteBodySize)
		e = e.Next()
	}
}

// TODO 机制确保事件执行
// NotifyEventDropped
func (client *Client) EnqueueEvent(fn func(*Client)) bool {
	if client.closed.Load() { // 已关闭
		log.Infof("can't add event to closed client %s", client.Log())
		return false
	}

	select {
	case client.eventCh <- fn:
		return true
	case <-time.After(client.cfg.EnqueueTimeout):
		log.Infof("client %s enqueue event timeout", client.Log())
		return false
	default:
		return false
	}
}

// 如果不能入队列，就直接处理
func (client *Client) EnsureEvent(fn func(*Client)) {
	if !client.EnqueueEvent(fn) {
		fn(client)
	}
}

// Run 阻塞进行，连接读写
func (client *Client) Run() {
	go client.read()
	client.write()
}

// 读
func (client *Client) read() {
	for {
		msg := &protocol.Message{
			Header: client.CreateHeader(),
		}
		err := protocol.ReadLimitMessage(client.Conn, msg, client.cfg.MaxReadBodySize)
		if err != nil {
			log.Infof("client %s read error: %s", client.Log(), err)
			client.Close()
			break
		}
		client.HandleMessage(client, msg)
	}
}

// 写
func (client *Client) write() {
	//发送在线消息
	for {
		select {
		case msg := <-client.mch:
			if msg == nil { // 一定是通过close传进来的
				if client.UID != 0 {
					log.Infof("client %s socket closed", client.Log())
				}
				client.flush()
				return
			}
			err := protocol.WriteLimitMessage(client.Conn, msg, client.cfg.MaxWriteBodySize)
			if err != nil {
				if _, ok := err.(net.Error); ok || err == io.EOF { // maybe write on closed connection
					log.Infof("[write-nil] client %s, msg : %s, err: %s", client.Log(), msg, err)
				} else {
					log.Warnf("[write-err] client %s, msg : %s, err: %s", client.Log(), msg, err)
				}
				client.HandleOnWriteError(client, msg, err)
				//// 防止线程问题
				//if client.closed.CAS(false, true) {
				//	client.flush()
				//}
				//return
			}

		case <-client.lSignal:
			client.SendLMessages()

		case fn := <-client.eventCh:
			if fn != nil {
				fn(client)
			}
		}
	}
}

// 发送一般消息
// DEBUG: flush 之后出现了 client.mch <-
func (client *Client) flush() {
	if !client.closed.Load() { // 防止发送端继续发送数据
		return
	}
	//close(client.mch)
	//for msg := range client.mch {
	//	protocol.WriteLimitMessage(client.Conn, msg, client.cfg.MaxWriteBodySize)
	//}
	//close(client.eventCh)
	//for fn := range client.eventCh {
	//	fn(client)
	//}
	//client.SendLMessages()
}

// Close 不会立即关闭连接，而是给连接发送nil信号
// 重写Close 避免直接调用底层的Conn close
func (client *Client) Close() {
	if client.closed.CAS(false, true) {
		log.Infof("try close client %s", client.Log())
		client.mch <- nil
	}
}
