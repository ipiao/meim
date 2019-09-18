package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ipiao/meim/log"
)

const (
	DEFAULT_READ_TIMEOUT  = time.Minute * 15
	DEFAULT_WRITE_TIMEOUT = time.Second * 10
	DEFAULT_MAX_CONNNUM   = 65536
)

// 服务
type Server struct {
	lncfg        *ListenerConfig // 配置
	ln           net.Listener    // 单一服务监听
	readTimeout  time.Duration   // 读超时
	writeTimeout time.Duration   // 写超时
	maxConnNum   int             // 限制最大连接数

	mu      sync.RWMutex // 锁
	conns   ConnSet      // 客户端集
	connsMu sync.RWMutex // 客户端锁

	closeSig chan bool      // 服务结束信号
	wgLn     sync.WaitGroup // listener的等待组
	wgConns  sync.WaitGroup // clients的等待组

}

// 新建服务
func NewServer(options ...OptionFn) *Server {
	return NewServerWithConfig(nil, options...)
}

// 新建服务
func NewServerWithConfig(cfg *ListenerConfig, options ...OptionFn) *Server {
	s := new(Server)
	s.lncfg = cfg

	s.init()
	for _, op := range options {
		op(s)
	}
	return s
}

// 初始化默认配置
func (s *Server) init() {
	if s.lncfg == nil {
		s.lncfg = &ListenerConfig{
			Network: "tcp",
		}
	}
	if s.readTimeout == 0 {
		s.readTimeout = DEFAULT_READ_TIMEOUT
	}
	if s.writeTimeout == 0 {
		s.writeTimeout = DEFAULT_WRITE_TIMEOUT
	}
	if s.maxConnNum == 0 {
		s.maxConnNum = DEFAULT_MAX_CONNNUM
	}

	if s.conns == nil {
		s.conns = NewConnSet()
	}

	if s.closeSig == nil {
		s.closeSig = make(chan bool)
	}
}

// 监听中断信号
func (s *Server) startShutdownListener() {
	go func() {
		log.Infof("server pid: %d", os.Getpid())
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		sig := <-c
		log.Infof("recv signal: %s", sig)
		s.closeSig <- true
	}()
}

func (s *Server) makeListener() (ln net.Listener, err error) {
	ml := makeListeners[s.lncfg.Network]
	if ml == nil {
		return nil, fmt.Errorf("can not make listener for %s", s.lncfg.Network)
	}
	return ml(s.lncfg)
}

func (s *Server) Run() {

	// check plugin
	if err := CheckExternalHandlers(); err != nil {
		log.Fatal(err)
	}

	s.startShutdownListener()

	ln, err := s.makeListener()
	if err != nil {
		log.Fatal(err)
	}
	s.ln = ln
	go s.serveListener()

	// 结束服务
	<-s.closeSig

	//关闭listener
	s.closeListener()
	// 处理ConnsClose
	s.connsMu.Lock()
	for conn := range s.conns {
		HandleCloseConn(conn)
	}
	s.connsMu.Unlock()
	s.wgConns.Wait()
	log.Infof("wait all agent onclose done")

}

func (s *Server) serveListener() {
	s.wgLn.Add(1)
	defer s.wgLn.Done()

	var tempDelay time.Duration
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Warnf("server accept error: %v\nretrying in %s", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			// 这里一般是主动断开程序造成的错误
			log.Warnf("server listener error: %s", err)
			return
		}
		tempDelay = 0

		s.handleConn(conn)
	}
}

// 处理连接
func (s *Server) handleConn(conn net.Conn) {
	s.connsMu.Lock()
	// 如果连接数量超过限制,直接返回
	// 理论上在auth时候就要避免
	if count := len(s.conns); count >= s.maxConnNum {
		s.connsMu.Unlock()
		conn.Close()
		log.Warnf("too many connections: %d", count)
		return
	}

	netConn := NewNetConn(conn, s.readTimeout, s.writeTimeout)

	s.conns.Add(netConn)
	s.connsMu.Unlock()
	s.wgConns.Add(1)
	log.Debug("new conn: %s", conn.RemoteAddr())

	go func() {
		HandleConnAccepted(netConn) // 这里面进行Conn消息收发处理等,阻塞
		// 阻塞条件结束
		netConn.Close()
		s.connsMu.Lock()
		s.conns.Remove(netConn)
		s.connsMu.Unlock()

		HandleConnClosed(netConn)
		s.wgConns.Done()
	}()
}

func (s *Server) closeListener() {
	s.ln.Close()
	s.wgLn.Wait()
}

func (s *Server) ConnSet() ConnSet {
	s.connsMu.RLock()
	defer s.connsMu.RUnlock()
	return s.conns.Clone()
}
