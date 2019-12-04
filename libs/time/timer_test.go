package time

//
//import (
//	"testing"
//	"time"
//
//	"github.com/ipiao/meim.v2/log"
//)
//
//func getJob(i int) func() {
//	return func() {
//		log.Debugf("task : %d", i)
//	}
//}
//
//func TestTimer(t *testing.T) {
//	defer log.Flush()
//	timer := NewTimer(100)
//	tds := make([]*TimerData, 100)
//	for i := 0; i < 10; i++ {
//		tds[i] = timer.Add(time.Duration(10-i)*time.Second, getJob(i))
//	}
//	//for i := 0; i < 10; i++ {
//	//	tds[i] = timer.Add(time.Duration(i)*time.Second+5*time.Minute, nil)
//	//}
//	//printTimer(timer)
//	//for i := 0; i < 10; i++ {
//	//	log.Debugf("td: %s, %s, %d", tds[i].Key, tds[i].ExpireString(), tds[i].index)
//	//	timer.Del(tds[i])
//	//}
//	//printTimer(timer)
//	//for i := 0; i < 10; i++ {
//	//	tds[i] = timer.Add(time.Duration(i)*time.Second+5*time.Minute, nil)
//	//}
//	//printTimer(timer)
//	//for i := 0; i < 10; i++ {
//	//	timer.Del(tds[i])
//	//}
//	printTimer(timer)
//	timer.Add(time.Second, nil)
//	time.Sleep(time.Second * 20)
//	if len(timer.timers) != 0 {
//		t.FailNow()
//	}
//}
//
//func printTimer(timer *Timer) {
//	log.Debugf("----------timers: %d ----------", len(timer.timers))
//	for i := 0; i < len(timer.timers); i++ {
//		log.Debugf("timer: %s, %s, index: %d", timer.timers[i].Key, timer.timers[i].ExpireString(), timer.timers[i].index)
//	}
//	log.Debugf("--------------------")
//}
//
//func TestDelayTask(t *testing.T) {
//	defer log.Flush()
//	timer := NewTimer(100)
//	fn := func() {
//		log.Debugf("timer task exec at %s", time.Now())
//		time.Sleep(time.Second)
//	}
//	timer.Add(time.Second, fn)
//	timer.Add(time.Second, fn)
//	timer.Add(time.Second, fn)
//
//	time.Sleep(time.Second * 10)
//	if len(timer.timers) != 0 {
//		t.FailNow()
//	}
//}
