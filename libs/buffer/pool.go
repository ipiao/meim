package buffer

import (
	"bytes"
	"sync"
)

type Pool struct {
	sync.Pool
}

// 新建一个buf pool
func New(initSize int) *Pool {
	return &Pool{
		Pool: sync.Pool{New: func() interface{} {
			b := bytes.NewBuffer(make([]byte, initSize))
			b.Reset()
			return b
		}},
	}
}

// 获取
func (bp *Pool) Get() *bytes.Buffer {
	return bp.Pool.Get().(*bytes.Buffer)
}

// 放回
func (bp *Pool) Put(b *bytes.Buffer) {
	b.Reset()
	bp.Pool.Put(b)
}
