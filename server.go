package meim

import (
	"math/rand"
	"sync"
	"time"

	"github.com/ipiao/meim/conf"

	"github.com/ipiao/meim/libs/netutil"
	xtime "github.com/ipiao/meim/libs/time"
	"github.com/ipiao/meim/libs/utils"
	"github.com/ipiao/meim/log"
	"github.com/zhenjl/cityhash"
)

type Server struct {
	c          *conf.Config
	wg         sync.WaitGroup // wait for all channel close when server closed
	backoff    netutil.BackOff
	round      *conf.Round    // accept round store
	buckets    []*conf.Bucket // subkey bucket
	bucketSize uint32
	serverID   string
	unid       *utils.UniqueId
	closed     chan struct{}
}

// NewServer returns a new Server.
func NewServer(c *conf.Config) *Server {
	s := &Server{
		c:      c,
		round:  NewRound(*c.Round),
		closed: make(chan struct{}),
		backoff: &netutil.BackOffConfig{
			MaxDelay:  time.Second * 10,
			BaseDelay: time.Millisecond * 100,
			Factor:    1.2,
			Jitter:    0.6,
		},
	}
	// init bucket
	s.buckets = make([]*conf.Bucket, c.Bucket.Size)
	s.bucketSize = uint32(c.Bucket.Size)
	for i := 0; i < c.Bucket.Size; i++ {
		s.buckets[i] = NewBucket(c.Bucket)
	}
	s.serverID = c.ServerID
	s.unid = utils.NewUniqueId(1000, 1<<24-1)
	return s
}

// 开始运行服务
func (s *Server) Run() {
	go s.onlineproc()
	InitNetListeners(s)
}

// Buckets return all buckets.
func (s *Server) Buckets() []*conf.Bucket {
	return s.buckets
}

// Bucket get the bucket by subkey.
func (s *Server) Bucket(subKey string) *conf.Bucket {
	idx := cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % s.bucketSize
	if s.c.Debug {
		log.Infof("%s hit channel bucket index: %d use cityhash", subKey, idx)
	}
	return s.buckets[idx]
}

// RandServerHeartbeat 获取随机心跳超时时间
func (s *Server) RandServerHeartbeat() time.Duration {
	return time.Duration(s.c.Protocol.MinHeartbeatInterval +
		xtime.Duration(rand.Int63n(int64(s.c.Protocol.MaxHeartbeatInterval-s.c.Protocol.MinHeartbeatInterval))))
}

// Close close the server.
func (s *Server) Close() (err error) {
	close(s.closed)
	for _, bucket := range s.buckets {
		for _, ch := range bucket.chs {
			ch.Close() // 主动关闭掉所有的连接
		}
	}
	s.wg.Wait()
	return
}

// 定期更新房间的实际总数量
func (s *Server) onlineproc() {
	for {
		var (
			allRoomsCount map[string]int32
			err           error
		)
		roomCount := make(map[string]int32)
		for _, bucket := range s.buckets {
			for roomID, count := range bucket.EachRoomsCount() {
				roomCount[roomID] += count
			}
		}
		if allRoomsCount, err = Handler.RenewOnlineCount(); err != nil {
			time.Sleep(time.Second * 3)
			continue
		}
		for _, bucket := range s.buckets {
			bucket.UpdateRoomsCount(allRoomsCount)
		}
		time.Sleep(time.Second * 30)
	}
}
