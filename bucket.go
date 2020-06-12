package meim

import (
	"sync"
	"sync/atomic"

	"github.com/ipiao/meim/protocol"
)

// bucket 初始化配置可选项
type BucketOptions struct {
	// Bucket is bucket config.
	Size    int // bucket数量
	Channel int // channel初始化数量

	Room          int    // room初始化数量
	RoutineAmount uint64 // room的后台推送线程数量
	RoutineSize   int    // room每个推送线程的缓冲大小
}

// Bucket 统一管理通道
type Bucket struct {
	options *BucketOptions
	cLock   sync.RWMutex        // protect the channels for chs
	chs     map[string]*Channel // map sub key to a channel

	// room
	rooms       map[string]*Room           // bucket room channels
	routines    []chan *protocol.RoomProto // 不同的通道执行房间消息发送
	routinesNum uint64                     // 用作计数，每次房间消息轮询换一个通道

	ipChannels map[string]int32 // 每个ip对应的通道个数
}

// NewBucket new 新建bucket，存放连接的channel
func NewBucket(opts *BucketOptions) (b *Bucket) {
	b = new(Bucket)
	b.chs = make(map[string]*Channel, opts.Channel)
	b.ipChannels = make(map[string]int32)
	b.options = opts

	b.rooms = make(map[string]*Room, opts.Room)
	b.routines = make([]chan *protocol.RoomProto, opts.RoutineAmount)
	for i := uint64(0); i < opts.RoutineAmount; i++ {
		c := make(chan *protocol.RoomProto, opts.RoutineSize)
		b.routines[i] = c
		go b.roomProc(c)
	}
	return
}

// ChannelCount 通道数量，相当个人连接数
func (b *Bucket) ChannelCount() int {
	return len(b.chs)
}

// RoomCount 返回当前房间数量
func (b *Bucket) RoomCount() int {
	return len(b.rooms)
}

// RoomsCount 获取所有房间的当前节点在线数量
func (b *Bucket) EachRoomsCount() (res map[string]int32) {
	var (
		roomID string
		room   *Room
	)
	b.cLock.RLock()
	res = make(map[string]int32)
	for roomID, room = range b.rooms {
		if room.Online > 0 {
			res[roomID] = room.Online
		}
	}
	b.cLock.RUnlock()
	return
}

// ChangeRoom 进入或者离开房间
func (b *Bucket) ChangeRoom(newRid string, ch *Channel) (err error) {
	var (
		newRoom *Room
		ok      bool
		oldRoom = ch.Room
	)

	// 如果房间没人了就删除房间
	if oldRoom != nil && oldRoom.Del(ch) {
		b.DelRoom(oldRoom)
	}
	// 离开房间
	if newRid == "" {
		ch.Room = nil
		return
	}
	// 进入新房间
	b.cLock.Lock()
	if newRoom, ok = b.rooms[newRid]; !ok {
		newRoom = NewRoom(newRid)
		b.rooms[newRid] = newRoom
	}
	b.cLock.Unlock()
	if err = newRoom.Put(ch); err != nil {
		return
	}
	ch.Room = newRoom
	return
}

// Put 存放一个新channel
func (b *Bucket) Put(rid string, ch *Channel) (err error) {
	var (
		room *Room
		ok   bool
	)
	b.cLock.Lock()
	// 如果channel key 存在，实际是同一个端，关闭旧通道
	if dch := b.chs[ch.Key]; dch != nil {
		dch.Close()
	}
	b.chs[ch.Key] = ch
	if rid != "" {
		if room, ok = b.rooms[rid]; !ok {
			room = NewRoom(rid)
			b.rooms[rid] = room
		}
		ch.Room = room
	}
	b.ipChannels[ch.IP]++
	b.cLock.Unlock()
	if room != nil {
		err = room.Put(ch)
	}
	return
}

// Del 删除channel
func (b *Bucket) Del(dch *Channel) {
	var (
		ok   bool
		ch   *Channel
		room *Room
	)
	b.cLock.Lock()
	if ch, ok = b.chs[dch.Key]; ok {
		room = ch.Room
		if ch == dch {
			delete(b.chs, ch.Key)
		}
		// 删减ip数量
		if b.ipChannels[ch.IP] > 1 {
			b.ipChannels[ch.IP]--
		} else {
			delete(b.ipChannels, ch.IP)
		}
	}
	b.cLock.Unlock()
	if room != nil && room.Del(ch) {
		// 如果房间已空，从bucket中删除房间
		b.DelRoom(room)
	}
}

// Channel 通过key获取channel
func (b *Bucket) Channel(key string) (ch *Channel) {
	b.cLock.RLock()
	ch = b.chs[key]
	b.cLock.RUnlock()
	return
}

// Broadcast 广播消息到所有
func (b *Bucket) Broadcast(p *protocol.Proto, op int32) {
	var ch *Channel
	b.cLock.RLock()
	for _, ch = range b.chs {
		if !ch.NeedPush(op) {
			continue
		}
		_ = ch.Push(p)
	}
	b.cLock.RUnlock()
}

// Room 根据房间号获取房间
func (b *Bucket) Room(rid string) (room *Room) {
	b.cLock.RLock()
	room = b.rooms[rid]
	b.cLock.RUnlock()
	return
}

// DelRoom 删除房间
func (b *Bucket) DelRoom(room *Room) {
	b.cLock.Lock()
	delete(b.rooms, room.ID)
	b.cLock.Unlock()
	room.Close()
}

// BroadcastRoom 广播房间消息
func (b *Bucket) BroadcastRoom(msg *protocol.RoomProto) {
	num := atomic.AddUint64(&b.routinesNum, 1) % b.options.RoutineAmount
	b.routines[num] <- msg
}

// Rooms 获取所有在线人数大于0的房间号
func (b *Bucket) Rooms() (res map[string]struct{}) {
	var (
		roomID string
		room   *Room
	)
	res = make(map[string]struct{})
	b.cLock.RLock()
	for roomID, room = range b.rooms {
		if room.Online > 0 {
			res[roomID] = struct{}{}
		}
	}
	b.cLock.RUnlock()
	return
}

// IPCount 获取所有的ip和连接数量
func (b *Bucket) IPCount() (res map[string]struct{}) {
	var (
		ip string
	)
	b.cLock.RLock()
	res = make(map[string]struct{}, len(b.ipChannels))
	for ip = range b.ipChannels {
		res[ip] = struct{}{}
	}
	b.cLock.RUnlock()
	return
}

// UpRoomsCount 更新房间总人数，非单个节点
func (b *Bucket) UpRoomsCount(roomCountMap map[string]int32) {
	var (
		roomID string
		room   *Room
	)
	b.cLock.RLock()
	for roomID, room = range b.rooms {
		room.AllOnline = roomCountMap[roomID]
	}
	b.cLock.RUnlock()
}

// 房间进程，执行房间消息推送
func (b *Bucket) roomProc(c chan *protocol.RoomProto) {
	for {
		arg := <-c
		if room := b.Room(arg.RoomId); room != nil {
			room.Push(arg.Proto)
		}
	}
}
