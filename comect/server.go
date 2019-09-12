package comect

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

	mu        sync.RWMutex // 锁
	clients   ClientSet    // 客户端集
	clientsMu sync.RWMutex // 客户端锁

	closeSig  chan bool      // 服务结束信号
	lnwg      sync.WaitGroup // listener的等待组
	clientswg sync.WaitGroup // clients的等待组

	Plugins PluginContainer // 插件槽
}

// 新建服务
func NewServer(options ...OptionFn) *Server {
	s := new(Server)
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

	if s.clients == nil {
		s.clients = NewClientSet()
	}

	if s.Plugins == nil {
		s.Plugins = &pluginContainer{}
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
	s.startShutdownListener()

	ln, err := s.makeListener()
	if err != nil {
		log.Fatal(err)
	}
	s.ln = ln
	go s.serveListener(ln)
}

func (s *Server) serveListener(ln net.Listener) error {
	s.lnwg.Add(1)
	defer s.lnwg.Done()

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
				log.Infof("server accept error: %v\nretrying in %s", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		tempDelay = 0

		s.handleConn(conn)
	}
}

// 处理连接
func (s *Server) handleConn(conn net.Conn) {
	s.clientsMu.Lock()
	// 如果连接数量超过限制,直接返回
	// 理论上在auth时候就要避免
	if len(s.clients) >= s.maxConnNum {
		s.clientsMu.Unlock()
		conn.Close()
		log.Warnf("too many connections: %d", len(s.clients))
		return
	}

	netConn := NewNetConn(conn, s.readTimeout, s.writeTimeout)
	client := NewClient(s.Plugins.HandleConnAccept(netConn))

	s.clients.Add(client)
	s.clientsMu.Unlock()
	s.clientswg.Add(1)
	log.Debug("new client: %s", client.RemoteAddr())

	go func() {
		// if gate.OnNewAgent != nil {
		// 	gate.OnNewAgent(agent)
		// }
		// agent.Run() // 阻塞运行
		// // cleanup
		// conn.Close()
		// gate.agentSetMu.Lock()
		// gate.agentSet.Remove(agent)
		// gate.agentSetMu.Unlock()
		// agent.OnClose()

		// gate.wgConns.Done()
	}()
}
