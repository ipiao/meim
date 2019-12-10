package meim

import (
	"github.com/ipiao/meim.v2/protocol"
)

// Ring ring proto buffer.
type Ring struct {
	// 读指针 read-pointer
	rp uint64
	// 写指针 write-pointer
	wp uint64
	// 环大小
	num uint64
	// 掩码，因为指针递增的
	mask uint64
	// TODO split cacheline, many cpu cache line size is 64
	// pad [40]byte
	data []protocol.Proto
}

// NewRing 创建一个环缓冲
func NewRing(num int) *Ring {
	r := new(Ring)
	r.init(uint64(num))
	return r
}

// Init 初始化环还
func (r *Ring) Init(num int) {
	r.init(uint64(num))
}

// 获取到大于num的最小2^N值
func (r *Ring) init(num uint64) {
	// 2^N
	if num&(num-1) != 0 {
		for num&(num-1) != 0 {
			num &= num - 1
		}
		num = num << 1
	}
	r.data = make([]protocol.Proto, num)
	r.num = num
	r.mask = r.num - 1
}

// Get 从缓冲中读取一个数据
func (r *Ring) Get() (proto *protocol.Proto, err error) {
	if r.rp == r.wp {
		return nil, ErrRingEmpty
	}
	proto = &r.data[r.rp&r.mask]
	return
}

// GetAdv 递增读索引 Get-Advance
func (r *Ring) GetAdv() {
	r.rp++
}

// Set 获取到缓冲池里面的可用缓冲
func (r *Ring) Set() (proto *protocol.Proto, err error) {
	if r.wp-r.rp >= r.num {
		return nil, ErrRingFull
	}
	proto = &r.data[r.wp&r.mask]
	return
}

// SetAdv 递增写读索引 Set-Advance
func (r *Ring) SetAdv() {
	r.wp++
}

// Reset 重置 ring.
func (r *Ring) Reset() {
	r.rp = 0
	r.wp = 0
	// prevent pad compiler optimization
	// r.pad = [40]byte{}
}
