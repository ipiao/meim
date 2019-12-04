package netutil

import (
	"math/rand"
	"time"
)

// DefaultBackOffConfig 是一般的机制
var DefaultBackOffConfig = BackOffConfig{
	MaxDelay:  60 * time.Second,
	BaseDelay: 100 * time.Millisecond,
	Factor:    1.6,
	Jitter:    0.2,
}

// BackOff 定义了方法执行失败后的重试时间机制
type BackOff interface {
	// BackOff 根据重试次数获取下一次的重试时间
	BackOff(retries int) time.Duration
}

// BackOffConfig defines the parameters for the default backoff strategy.
type BackOffConfig struct {
	// 最大延迟时间
	MaxDelay time.Duration
	// 初始延迟时间
	BaseDelay time.Duration
	// factor 是每次重试的固定增加时间,乘因子
	Factor float64
	// jitter 提供随机增加时间范围，还是乘
	Jitter float64
	// 最大重试此时，如果小于等于0,则无限重试
	MaxRetries int
}

// BackOff 返回下次重试的等待时间
func (bc *BackOffConfig) BackOff(retries int) time.Duration {
	if retries == 0 {
		return bc.BaseDelay
	}
	if bc.MaxRetries > 0 && retries > bc.MaxRetries {
		panic("backoff retry times out")
	}

	backOff, max := float64(bc.BaseDelay), float64(bc.MaxDelay)
	for backOff < max && retries > 0 {
		backOff *= bc.Factor
		retries--
	}
	if backOff > max {
		backOff = max
	}
	// Randomize backoff delays so that if a cluster of requests start at
	// the same time, they won't operate in lockstep.
	backOff *= 1 + bc.Jitter*(rand.Float64()*2-1)
	if backOff < 0 {
		return 0
	}
	return time.Duration(backOff)
}
