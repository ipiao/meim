package comect

import (
	"io"
	"net"
	"time"
)

// 为什么不直接用net.Conn?
// 考虑兼容 websocket.Conn等其他情况
type Conn interface {
	Read(n int) ([]byte, error)  // 读取指定字节数
	Write(b []byte) (int, error) // 写
	RemoteAddr() net.Addr        // 远端地址
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

func (conn *NetConn) Read(n int) ([]byte, error) {
	if conn.readTimeout > 0 {
		conn.Conn.SetReadDeadline(time.Now().Add(conn.readTimeout))
	}
	buff := make([]byte, n)
	_, err := io.ReadFull(conn.Conn, buff)
	return buff, err
}

func (conn *NetConn) Write(b []byte) (int, error) {
	if conn.writeTimeout > 0 {
		conn.Conn.SetWriteDeadline(time.Now().Add(conn.writeTimeout))
	}
	return conn.Conn.Write(b)
}

func IsTimeout(err error) bool {
	if e, ok := err.(net.Error); ok {
		return e.Timeout()
	}
	return false
}

// 连接集合
type ConnSet map[Conn]struct{}

func NewConnSet() ConnSet {
	return make(map[Conn]struct{})
}

func (set ConnSet) Add(c Conn) {
	set[c] = struct{}{}
}

func (set ConnSet) IsMember(c Conn) bool {
	if _, ok := set[c]; ok {
		return true
	}
	return false
}

func (set ConnSet) Remove(c Conn) {
	if _, ok := set[c]; !ok {
		return
	}
	delete(set, c)
}

func (set ConnSet) Count() int {
	return len(set)
}

// 只是浅复制
func (set ConnSet) Clone() ConnSet {
	n := make(map[Conn]struct{})
	for k, v := range set {
		n[k] = v
	}
	return n
}
