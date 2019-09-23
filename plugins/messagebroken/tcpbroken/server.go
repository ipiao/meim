package tcpbroken

import (
	"math/rand"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/ipiao/meim"
	"github.com/ipiao/meim/log"
)

var (
	clients      ClientSet
	clientsMutex sync.RWMutex // clients的全局锁
	Subcmd       int
	Unsubcmd     int
	DC           meim.DataCreator
)

//新增
func AddClient(client *Client) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	clients.Add(client)
}

// 移除
func RemoveClient(client *Client) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	clients.Remove(client)
}

// clone clients
func GetClientSet() ClientSet {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()
	s := NewClientSet()
	for c := range clients {
		s.Add(c)
	}
	return s
}

// 查找用户客户端
func FindClientSet(userId int64) ClientSet {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()
	s := NewClientSet()
	for c := range clients {
		if c.route.ContainUserID(userId) {
			s.Add(c)
		}
	}
	return s
}

// 判断用户是否在线
func IsUserOnline(uid int64) bool {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	for c := range clients {
		if c.route.IsUserOnline(uid) {
			return true
		}
	}
	return false
}

// 客户端处理
func handleClient(conn *net.TCPConn) {
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Duration(10 * 60 * time.Second))
	client := NewClient(conn)
	log.Infof("new client: %s", conn.RemoteAddr())
	client.Run()
}

// 监听开始
func Listen(f func(*net.TCPConn), listenAddr string) {
	listen, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Listen error: %v", err.Error())
		return
	}
	tcpListener, ok := listen.(*net.TCPListener)
	if !ok {
		log.Fatalf("listen error")
		return
	}
	log.Infof("start listen at %s", listenAddr)
	for {
		client, err := tcpListener.AcceptTCP()
		if err != nil {
			return
		}
		f(client)
	}
}

// 这个是tcpRouter的服务器
// 运行
func Run(addr string) {
	rand.Seed(time.Now().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	clients = NewClientSet()

	Listen(handleClient, addr)
}
