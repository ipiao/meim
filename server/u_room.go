package server

import (
	"sync"

	"github.com/ipiao/meim"
	"github.com/ipiao/meim/protocol"
)

// Room 存放房间里面的所有channel
type Room struct {
	ID        string
	rLock     sync.RWMutex
	next      *Channel
	dropped   bool
	Online    int32 // dirty read is ok
	AllOnline int32
}

// NewRoom 新建一个room，制定id
func NewRoom(id string) (r *Room) {
	r = new(Room)
	r.ID = id
	r.dropped = false
	r.next = nil
	r.Online = 0
	return
}

// Put 将channel放进room
func (r *Room) Put(ch *Channel) (err error) {
	r.rLock.Lock()
	if !r.dropped {
		if r.next != nil {
			r.next.Prev = ch
		}
		ch.Next = r.next
		ch.Prev = nil
		r.next = ch // insert to header
		r.Online++
	} else {
		err = meim.ErrRoomDropped
	}
	r.rLock.Unlock()
	return
}

// Del 从room里面删除一个channel，返回房间是否没人了
func (r *Room) Del(ch *Channel) bool {
	r.rLock.Lock()
	if ch.Next != nil {
		// if not footer
		ch.Next.Prev = ch.Prev
	}
	if ch.Prev != nil {
		// if not header
		ch.Prev.Next = ch.Next
	} else {
		r.next = ch.Next
	}
	r.Online--
	r.dropped = r.Online == 0
	r.rLock.Unlock()
	return r.dropped
}

// Push 推送消息到房间，如果channel满了，就忽略
func (r *Room) Push(p *protocol.Proto) {
	r.rLock.RLock()
	for ch := r.next; ch != nil; ch = ch.Next {
		_ = ch.Push(p)
	}
	r.rLock.RUnlock()
}

// Close 关闭房间
func (r *Room) Close() {
	r.rLock.RLock()
	for ch := r.next; ch != nil; ch = ch.Next {
		ch.Close()
	}
	r.rLock.RUnlock()
}

// OnlineNum 所有在线人数
func (r *Room) OnlineNum() int32 {
	if r.AllOnline > 0 {
		return r.AllOnline
	}
	return r.Online
}
