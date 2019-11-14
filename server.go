package meim_v2

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ipiao/meim.v2/log"
)

// 服务
type Server struct {
	cfg              *Config        // 配置
	ln               net.Listener   //
	clients          ClientSet      // 连接到当前服务的客户端列表
	clientsLock      sync.RWMutex   // 客户端锁
	lnWaitGroup      sync.WaitGroup //
	clientsWaitGroup sync.WaitGroup //
	closeSignal      chan bool      // 接收进程结束信号

	plugin     PluginI
	pluginOnce sync.Once
}

// 新键服务，必须要有地址
func NewServer(addr string, opts ...ConfigOption) *Server {
	cfg := &Config{
		Address: addr,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return NewServerWithConfig(cfg)
}

// 带配置
func NewServerWithConfig(cfg *Config) *Server {
	cfg.Init()
	s := &Server{
		cfg:         cfg,
		clients:     NewClientSet(),
		closeSignal: make(chan bool),
	}
	return s
}

// 注册设置服务插件
func (s *Server) RegisterPlugin(p PluginI) {
	s.pluginOnce.Do(func() {
		s.plugin = p
	})
}

// Run 运行服务
func (s *Server) Run() {
	if s.plugin == nil {
		s.plugin = DefaultPlugin()
	}
	s.startShutdownListener()

	ln, err := s.makeListener()
	if err != nil {
		log.Fatal(err)
	}
	s.ln = ln
	go s.serveListener()

	// 结束服务
	<-s.closeSignal
	s.Shutdown()
}

// 监听中断信号
func (s *Server) startShutdownListener() {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		sig := <-c
		log.Infof("receive signal %s...", sig)
		s.closeSignal <- true
	}()
}

func (s *Server) makeListener() (ln net.Listener, err error) {
	ml := listenerMakers[s.cfg.Network]
	if ml == nil {
		return nil, fmt.Errorf("can not make listener of %s", s.cfg.Network)
	}
	return ml(s.cfg)
}

func (s *Server) serveListener() {
	log.Infof("start listener: %s@%s...", s.cfg.Network, s.cfg.Address)
	s.lnWaitGroup.Add(1)
	defer s.lnWaitGroup.Done()
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
			return
		}
		tempDelay = 0
		s.handleNetConn(conn)
	}
}

// 处理连接
func (s *Server) handleNetConn(conn net.Conn) {
	s.clientsLock.Lock()
	// 如果连接数量超过限制,直接返回
	if count := len(s.clients); count >= s.cfg.MaxConnections {
		s.clientsLock.Unlock()
		conn.Close()
		log.Warnf("too many connections: %d, connection %d dropped", count, conn.RemoteAddr())
		return
	}

	netConn := NewNetConn(conn, s.cfg.ReadTimeout, s.cfg.WriteTimeout)
	client := NewClient(netConn, s.cfg.ClientConfig, s.plugin)
	s.clients.Add(client)
	s.clientsLock.Unlock()
	s.clientsWaitGroup.Add(1)

	go func() {
		log.Debugf("run new client: %s", client.RemoteAddr())
		client.Run()    // 这里面进行Conn消息收发处理等,阻塞
		netConn.Close() //	netConn.Close()
		s.clientsLock.Lock()
		s.clients.Remove(client)
		s.clientsLock.Unlock()
		s.clientsWaitGroup.Done()
	}()
}

// 结束服务
func (s *Server) Shutdown() {

	//关闭listener
	s.closeListener()

	// 处理ConnClose
	s.clientsLock.Lock()
	for client := range s.clients {
		client.Close()
	}
	s.clientsLock.Unlock()
	s.clientsWaitGroup.Wait()
}

// 结束服务listener
func (s *Server) closeListener() {
	s.ln.Close()
	s.lnWaitGroup.Wait()
}

// 返回当前客户端列表
func (s *Server) ClientSet() ClientSet {
	s.clientsLock.RLock()
	defer s.clientsLock.RUnlock()
	return s.clients.Clone()
}
