package meim

import (
	"net"
	"time"

	"github.com/ipiao/meim/log"
)

// 为什么不直接用net.Conn?
// 考虑兼容 websocket.Conn等其他情况
type Conn interface {
	Read([]byte) (int, error)
	Write(b []byte) (int, error) // 写
	RemoteAddr() net.Addr        // 远端地址
	LocalAddr() net.Addr         // 本地地址
	Close() error                // 关闭
}

// tcp连接
type NetConn struct {
	net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewNetConn(conn net.Conn, rto, wto time.Duration) *NetConn {
	if tc, ok := conn.(*net.TCPConn); ok {
		tc.SetLinger(10)
	}
	return &NetConn{
		Conn:         conn,
		readTimeout:  rto,
		writeTimeout: wto,
	}
}

func (conn *NetConn) Read(buff []byte) (int, error) {
	if conn.readTimeout > 0 {
		conn.Conn.SetReadDeadline(time.Now().Add(conn.readTimeout))
	}
	n, err := conn.Conn.Read(buff)
	// n, err := io.ReadFull(conn.Conn, buff)
	if err != nil {
		log.Debugf("read error: %s, addr: %s", err, conn.RemoteAddr())
	}
	return n, err
}

func (conn *NetConn) Write(b []byte) (int, error) {
	if conn.writeTimeout > 0 {
		conn.Conn.SetWriteDeadline(time.Now().Add(conn.writeTimeout))
	}
	n, err := conn.Conn.Write(b)
	if err != nil {
		log.Debugf("write error: %s, addr: %s", err, conn.RemoteAddr())
	}
	return n, err
}

func IsTimeout(err error) bool {
	if e, ok := err.(net.Error); ok {
		return e.Timeout()
	}
	return false
}
