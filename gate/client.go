package gate

import (
	"container/list"
	"time"

	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/protocol"
	"github.com/ipiao/meim/server"
)

type Client struct {
	Connection
}

func NewClient(conn server.Conn, dc protocol.DataCreator) *Client {
	client := new(Client)
	client.Conn = conn
	client.mch = make(chan *protocol.Message, 16)
	client.lmsch = make(chan int, 1)
	client.lmessages = list.New()
	client.extch = make(chan func(*Client), 8)
	client.enqueueTimeout = time.Second * 10
	client.dc = dc
	return nil
}

func (c *Client) read() {
	for {
		msg, err := protocol.ReadLimitMessage(c.Conn, c.dc, 128*1024)
		if err != nil {
			log.Info("client read error:", err)
			c.Close()
			break
		}
		exts.HandleMessage(c, msg)
	}
}

func (c *Client) write() {
	//发送在线消息
	for {
		select {
		case msg := <-c.mch:
			if msg == nil {
				log.Infof("client:%d socket closed", c.uid)
				c.FlushMessage()
				c.Conn.Close()
				break
			}
			protocol.WriteMessage(c.Conn, msg)

		case <-c.lmsch:
			c.SendLMessages()
			break

		}
	}
}

func (c *Client) Close() {
	if c.closed.CAS(false, true) {
		log.Infof("try close client %s", c.Log())
		c.mch <- nil
	}
}

func (client *Client) Run() {
	client.write()
	go client.read()
}
