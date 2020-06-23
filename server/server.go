package server

import (
	"time"

	"github.com/ipiao/meim/conf"
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

// Close close the server.
func (s *Server) Close() (err error) {
	return
}

// 定期更新房间的实际数量
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
			time.Sleep(time.Second)
			continue
		}
		for _, bucket := range s.buckets {
			bucket.UpdateRoomsCount(allRoomsCount)
		}
		time.Sleep(time.Second * 10)
	}
}
