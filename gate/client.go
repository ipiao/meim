package gate

import (
	"container/list"

	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/protocol"
	"github.com/ipiao/meim/server"
)

type Client struct {
	Connection
}

func NewClient(conn server.Conn, dc protocol.DataCreator) *Client {
	client := new(Client)
	client.conn = conn
	client.mch = make(chan *protocol.Message, 16)
	client.lmsch = make(chan int, 1)
	client.lmessages = list.New()
	client.extch = make(chan func(*Client), 8)
	client.dc = dc
	return nil
}

func (c *Client) Read() {
	for {
		msg, err := protocol.ReadLimitMessage(c.conn, c.dc, 128*1024)
		if err != nil {
			log.Info("client read error:", err)
			c.HandleCloseClient()
			break
		}
		log.Debug(msg)
	}
}

func (c *Client) Write() {
	//发送在线消息
	for {
		select {
		case msg := <-c.mch:
			if msg == nil {
				log.Infof("client:%d socket closed", c.uid)
				c.conn.Close()
				break
			}
			protocol.WriteMessage(c.conn, msg)

		case <-c.lmsch:
			c.SendLMessages()
			break

		}
	}
}

//发送等待队列中的消息
func (c *Client) SendLMessages() {
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

func (c *Client) HandleCloseClient() {
	if c.closed.CAS(false, true) {
		c.mch <- nil
	}
}

func (client *Client) Run() {
	go client.Read()
	client.Write()
}
