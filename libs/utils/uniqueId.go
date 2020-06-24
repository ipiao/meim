package utils

import "sync/atomic"

type UniqueId struct {
	min uint32
	max uint32
	cur uint32
}

func (id *UniqueId) Generate() uint32 {
	if id.cur >= id.max {
		if !atomic.CompareAndSwapUint32(&id.cur, id.max, id.min) {
			if id.cur >= id.max {
				atomic.StoreUint32(&id.cur, id.min)
			}
		}
	}
	return atomic.AddUint32(&id.cur, 1)
}

func NewUniqueId(min, max uint32) *UniqueId {
	return &UniqueId{
		min: min,
		max: max,
		cur: min,
	}
}
