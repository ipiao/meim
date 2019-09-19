package tcprouter

import (
	"net"
	"time"

	"github.com/ipiao/meim/gate"
	"github.com/ipiao/meim/log"
)

type TCPBrokenClient struct {
	addr string
	conn net.Conn
}

func NewTCPRouterClient(addr string) *TCPBrokenClient {
	tr := &TCPBrokenClient{
		addr: addr,
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

func (tr *TCPBrokenClient) SendMessage(msg *gate.InternalMessage) error {
	return nil
}
