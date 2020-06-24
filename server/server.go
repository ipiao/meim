package server

import (
	"math/rand"
	"time"

	"github.com/ipiao/meim/conf"
	xtime "github.com/ipiao/meim/libs/time"
	"github.com/ipiao/meim/log"
	"github.com/zhenjl/cityhash"
)

type Server struct {
	c         *conf.Config
	round     *Round    // accept round store
	buckets   []*Bucket // subkey bucket
	bucketIdx uint32
	serverID  string
}

// NewServer returns a new Server.
func NewServer(c *conf.Config) *Server {
	s := &Server{
		c:     c,
		round: NewRound(*c.Round),
	}
	// init bucket
	s.buckets = make([]*Bucket, c.Bucket.Size)
	s.bucketIdx = uint32(c.Bucket.Size)
	for i := 0; i < c.Bucket.Size; i++ {
		s.buckets[i] = NewBucket(c.Bucket)
	}
	s.serverID = c.ServerID
	go s.onlineproc()
	return s
}

// Buckets return all buckets.
func (s *Server) Buckets() []*Bucket {
	return s.buckets
}

// Bucket get the bucket by subkey.
func (s *Server) Bucket(subKey string) *Bucket {
	idx := cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % s.bucketIdx
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
