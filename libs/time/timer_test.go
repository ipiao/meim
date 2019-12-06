package time

import (
	"testing"
	"time"

	"github.com/ipiao/meim.v2/log"
)

func getJob(i int) func() {
	return func() {
		log.Debugf("exec at: %s, task : %d", time.Now(), i)
	}
}

func TestDelayTask(t *testing.T) {
	defer log.Flush()
	timer := NewTimer(100)
	fn := func() {
		log.Debugf("timer task exec at %s", time.Now())
		time.Sleep(time.Second)
	}
	timer.Add(time.Second, fn)
	timer.Add(time.Second, fn)
	timer.Add(time.Second, fn)

	time.Sleep(time.Second * 10)
	if timer.timerNum != 0 {
		t.FailNow()
	}
}

func TestTimer(t *testing.T) {
	defer log.Flush()
	timer := NewTimer(10)

	log.Debugf("round 1, print 0 - 99\n")
	for i := 0; i < 50; i++ {
		timer.Add(time.Second, getJob(i))
	}

	log.Debugf("round 2, print 50 - 1")
	//for i := 0; i < 100; i++ {
	//	timer.Add(time.Millisecond*time.Duration(100-i)*100, getJob(i))
	//}
	time.Sleep(time.Second * 5)
	if timer.timerNum != 0 {
		t.FailNow()
	}
}
