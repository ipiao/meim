package tcpbroken

import (
	"net"
	"time"

	"github.com/ipiao/meim/log"
	"github.com/ipiao/meim/protocol"
)

type TCPBrokenClient struct {
	addr     string
	dc       protocol.DataCreator
	conn     net.Conn
	subCmd   int
	unsubCmd int
}

func NewTCPRouterClient(addr string, dc protocol.DataCreator, subCmd, unsubCmd int) *TCPBrokenClient {
	tr := &TCPBrokenClient{
		addr: addr,
		dc:   dc,
	}
	return tr
}

func (tr *TCPBrokenClient) Connect() {
	nsleep := 100
	for {
		conn, err := net.Dial("tcp", tr.addr)
		if err != nil {
			log.Infof("tcpRouter server error: %v", err)
			nsleep *= 2
			if nsleep > 60*1000 {
				nsleep = 60 * 1000
			}
			log.Infof("tcpRouter connect sleep: %d ms", nsleep)
			time.Sleep(time.Duration(nsleep) * time.Millisecond)
			continue
		}
		tconn := conn.(*net.TCPConn)
		tconn.SetKeepAlive(true)
		tconn.SetKeepAlivePeriod(time.Duration(10 * 60 * time.Second))
		log.Infof("tcpRouter connected")
		tr.conn = tconn
		return
	}
}

func (tr *TCPBrokenClient) SendMessage(msg *protocol.InternalMessage) error {
	data, err := protocol.MarshalInternalMessage(msg)
	if err == nil {
		tr.conn.Write(data)
	}
	return err
}

func (tr *TCPBrokenClient) ReceiveMessage() (*protocol.InternalMessage, error) {
	return protocol.ReadInternalMessage(tr.conn, tr.dc)
}

func (tr *TCPBrokenClient) Subscribe(uid int64) {
	msg := new(protocol.InternalMessage)
	msg.Header = tr.dc.CreateHeader()
	msg.Header.SetCmd(tr.subCmd)
	msg.Sender = uid
	tr.SendMessage(msg)
}

func (tr *TCPBrokenClient) UnSubscribe(uid int64) {
	msg := new(protocol.InternalMessage)
	msg.Header = tr.dc.CreateHeader()
	msg.Header.SetCmd(tr.unsubCmd)
	msg.Sender = uid
	tr.SendMessage(msg)
}
