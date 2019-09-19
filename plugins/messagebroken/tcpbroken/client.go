package tcpbroken

import (
	"net"

	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/protocol"
)

type Client struct {
	name  string
	wt    chan *protocol.InternalMessage
	conn  *net.TCPConn
	route *Route
}

// 新建client
func NewClient(conn *net.TCPConn) *Client {
	client := new(Client)
	client.name = conn.RemoteAddr().String()
	client.conn = conn
	client.wt = make(chan *protocol.InternalMessage, 16)
	client.route = NewRoute()
	return client
}

// 读开始
func (client *Client) Read() {
	AddClient(client)
	for {
		msg, err := client.read()
		if err != nil || msg == nil {
			RemoveClient(client)
			client.wt <- nil
			break
		}
		client.HandleMessage(msg)
	}
}

// 实际消息处理
func (client *Client) HandleMessage(msg *protocol.InternalMessage) {
	cmd := msg.Header.Cmd()
	log.Infof("client: %s handle msg cmd: %d", client.name, cmd)
	switch cmd {
	case Subcmd:
		client.HandleSubscribe(msg.Sender)
	case Unsubcmd:
		client.HandleUnsubscribe(msg.Sender)
	default:
		client.HandlePublish(msg)
	}
}

// 处理订阅消息
func (client *Client) HandleSubscribe(uid int64) {
	log.Infof("client: %s subscribe uid:%d", client.name, uid)
	client.route.AddUserID(uid)
}

// 处理取消订阅
func (client *Client) HandleUnsubscribe(uid int64) {
	log.Infof("client: %s unsubscribe uid:%d", client.name, uid)
	client.route.RemoveUserID(uid)
}

// 处理发布单聊消息
func (client *Client) HandlePublish(msg *protocol.InternalMessage) {
	cmd := msg.Header.Cmd()
	log.Infof("client: %s publish message uid:%d cmd:%s", client.name, msg.Receiver, cmd)

	s := FindClientSet(msg.Receiver)
	offline := true
	for c := range s {
		if c.route.IsUserOnline(msg.Receiver) {
			offline = false
		}
	}

	if offline {
		//用户不在线,推送消息到终端
		// dopush
	}

	for c := range s {
		if client == c { //不发送给自身
			continue
		}
		c.wt <- msg
	}
}

// 真正写
func (client *Client) Write() {
	for {
		msg := <-client.wt
		if msg == nil {
			client.close()
			log.Infof("client %s socket closed", client.name)
			break
		}
		client.send(msg)
	}
}

func (client *Client) Run() {
	go client.Write()
	go client.Read()
	go client.Push()
}

func (client *Client) read() (*protocol.InternalMessage, error) {
	return protocol.ReadInternalMessage(client.conn, DC)
}

func (client *Client) send(msg *protocol.InternalMessage) error {
	return protocol.WriteInternalMessage(client.conn, msg)
}

func (client *Client) close() {
	client.conn.Close()
}

func (client *Client) Push() {

}

//

type ClientSet map[*Client]struct{}

func NewClientSet() ClientSet {
	return make(map[*Client]struct{})
}

func (set ClientSet) Add(c *Client) {
	set[c] = struct{}{}
}

func (set ClientSet) IsMember(c *Client) bool {
	if _, ok := set[c]; ok {
		return true
	}
	return false
}

func (set ClientSet) Remove(c *Client) {
	if _, ok := set[c]; !ok {
		return
	}
	delete(set, c)
}

func (set ClientSet) Count() int {
	return len(set)
}

// 只是浅复制
func (set ClientSet) Clone() ClientSet {
	n := make(map[*Client]struct{})
	for k, v := range set {
		n[k] = v
	}
	return n
}
