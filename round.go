package meim

import (
	"github.com/ipiao/meim.v2/libs/bytes"
	"github.com/ipiao/meim.v2/libs/time"
)

// RoundOptions round 操作
type RoundOptions struct {
	Timer        int
	TimerSize    int
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

// Round 连接从round里轮循换取读写缓冲等
type Round struct {
	readers []bytes.Pool
	writers []bytes.Pool
	timers  []time.Timer
	options *RoundOptions
}

// NewRound 新建一个Round
func NewRound(opts *RoundOptions) (r *Round) {
	var i int
	r = &Round{options: opts}
	// reader
	r.readers = make([]bytes.Pool, r.options.Reader)
	for i = 0; i < r.options.Reader; i++ {
		r.readers[i].Init(r.options.ReadBuf, r.options.ReadBufSize)
	}
	// writer
	r.writers = make([]bytes.Pool, r.options.Writer)
	for i = 0; i < r.options.Writer; i++ {
		r.writers[i].Init(r.options.WriteBuf, r.options.WriteBufSize)
	}
	// timer
	r.timers = make([]time.Timer, r.options.Timer)
	for i = 0; i < r.options.Timer; i++ {
		r.timers[i].Init(r.options.TimerSize)
	}
	return
}

// Timer获取一个定时器
func (r *Round) Timer(rn int) *time.Timer {
	return &(r.timers[rn%r.options.Timer])
}

// Reader 获取一个读内存池
func (r *Round) Reader(rn int) *bytes.Pool {
	return &(r.readers[rn%r.options.Reader])
}

// Writer 获取一个写内存池
func (r *Round) Writer(rn int) *bytes.Pool {
	return &(r.writers[rn%r.options.Writer])
}
