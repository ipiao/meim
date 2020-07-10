package time

import (
	"sync"
	xtime "time"
)

const (
	timerFormat      = "2006-01-02 15:04:05"
	infiniteDuration = xtime.Duration(1<<63 - 1)
)

// TimerData 定时任务数据，双向链表
type TimerData struct {
	Key    string
	expire xtime.Time
	fn     func()
	next   *TimerData
	pre    *TimerData
}

// Delay 延迟时长
func (td *TimerData) Delay() xtime.Duration {
	return xtime.Until(td.expire)
}

// ExpireString 到期时间的字符串表示
func (td *TimerData) ExpireString() string {
	return td.expire.Format(timerFormat)
}

// Timer 定时器.
type Timer struct {
	lock       sync.Mutex
	free       *TimerData // 里面的pre无效
	timers     *TimerData
	timersTail *TimerData // 最后一个
	signal     *xtime.Timer
	num        int
	timerNum   int // 等待执行的链长
	freeNum    int // 空闲链长
}

// 创建一个定时器，带任务书局缓存的，可以提好内存利用率
// 链表操作，添加任务基本是O(n)时间复杂度
func NewTimer(num int) (t *Timer) {
	t = new(Timer)
	t.init(num)
	return t
}

func (t *Timer) Init(num int) {
	t.init(num)
}

func (t *Timer) init(num int) {
	t.signal = xtime.NewTimer(infiniteDuration)
	t.num = num
	t.grow()
	go t.start()
}

// 增长定时任务的任务数据缓冲
func (t *Timer) grow() {
	var (
		i   int
		td  *TimerData
		tds = make([]TimerData, t.num)
	)
	t.free = &(tds[0])
	td = t.free
	for i = 1; i < t.num; i++ {
		td.next = &(tds[i])
		td = td.next
	}
	td.next = nil
	t.freeNum += t.num
	//log.Debugf("timer grow, freeNum is %d now", t.freeNum)
}

// 获取一个空闲的数据内存
func (t *Timer) get() (td *TimerData) {
	if td = t.free; td == nil {
		t.grow()
		td = t.free
	}
	t.free = td.next
	td.next = nil
	t.freeNum--
	return
}

// put 将数据内存放回
// TODO 不减，在内存数据足够多的情况下应释放掉
func (t *Timer) put(td *TimerData) {
	if t.freeNum > 2*t.num && t.freeNum >= t.timerNum { // 内存池足够的情况下，不放回了，等待gc
		//log.Debugf("freeNum is %d, timerNum is %d, td released", t.freeNum, t.timerNum)
		td.pre = nil
		td.next = nil
		return
	}
	if t.free != nil {
		t.free.pre = td
	}
	td.pre = nil
	td.next = t.free
	t.free = td
	t.freeNum++
}

// 添加一个任务，如果时间没有固定性，将是O(n)复杂度
func (t *Timer) Add(expire xtime.Duration, fn func()) (td *TimerData) {
	if fn == nil {
		panic("fn can not be nil")
	}
	t.lock.Lock()
	td = t.get()
	td.expire = xtime.Now().Add(expire)
	td.fn = fn
	t.add(td)
	t.lock.Unlock()
	return
}

// 删除一个任务， O(1) 复杂度
func (t *Timer) Del(td *TimerData) {
	t.lock.Lock()
	t.del(td)
	t.put(td)
	t.lock.Unlock()
}

// add 将任务添加到链表相应位置
func (t *Timer) add(td *TimerData) {
	tail := t.timersTail
	if tail == nil {
		t.timers = td
		t.timersTail = td
	} else if !td.expire.Before(tail.expire) {
		td.pre = t.timersTail
		t.timersTail.next = td
		t.timersTail = td
	} else {
		for tail != nil && tail.expire.After(td.expire) {
			tail = tail.pre
		}
		if tail == nil {
			td.next = t.timers
			t.timers = td
		} else {
			td.pre = tail
			td.next = tail.next
		}
	}

	// add to the minheap last node
	if td.pre == nil {
		// if first node, signal start goroutine
		d := td.Delay()
		t.signal.Reset(d)
		//log.Debugf("add td to head, reset duration to %s", d)
	}
	t.timerNum++
}

func (t *Timer) del(td *TimerData) {
	if td.fn == nil {
		return
	}
	if t.timers == td {
		t.timers = td.next
		if t.timers != nil {
			t.timers.pre = nil
		}
	}

	if t.timersTail == td {
		t.timersTail = td.pre
		if t.timersTail != nil {
			t.timersTail.next = nil
		}
	}

	td.fn = nil
	// remove item is the last node
	t.timerNum--
}

// Set 更新任务的时间数据
func (t *Timer) Set(td *TimerData, expire xtime.Duration) {
	t.lock.Lock()
	t.del(td)
	td.expire = xtime.Now().Add(expire)
	t.add(td)
	t.lock.Unlock()
}

// start start the timer.
func (t *Timer) start() {
	for {
		t.expire()
		<-t.signal.C
	}
}

// expire 超时执行
func (t *Timer) expire() {
	var (
		fn func()
		td *TimerData
		d  xtime.Duration
	)
	t.lock.Lock()
	for {
		if t.timers == nil {
			//log.Debugf("timers is nil, duration reset to NAN")
			d = infiniteDuration
			break
		}
		td = t.timers
		if d = td.Delay(); d > 0 {
			break
		}
		fn = td.fn
		// let caller put back
		t.del(td)
		t.lock.Unlock() // 解锁，防止执行任务的时候，进行添加或者删除等其他任务操作，引起死锁

		fn()
		t.lock.Lock()
	}
	t.signal.Reset(d)
	t.lock.Unlock()
}
