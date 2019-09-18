package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Value struct {
	v int
}

func fn1(a *Value) {
	a.v += 1
}

func fn2(a *Value) {
	a.v += 2
}

func TestMergeFunc(t *testing.T) {
	val := &Value{v: 0}

	fn := MergeFunc(fn1, fn2)
	vfn, ok := fn.(func(*Value))
	assert.True(t, ok)
	vfn(val) // +3
	assert.Equal(t, val.v, 3)

	fnn := MergeFunc(fn, fn2)
	vfn2, ok := fnn.(func(*Value))
	assert.True(t, ok)
	vfn2(val)                 // +5
	assert.Equal(t, val.v, 8) // +8
}
