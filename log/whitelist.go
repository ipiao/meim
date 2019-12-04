package log

// 白名单用户，日志单独输出，用户debug
// 当然可以日志插件/钩子的形式出现
// TODO 实际应用利用context

import (
	"log"
	"os"
)

var whitelist *Whitelist

type Whitelist struct {
	log  *log.Logger
	list map[int64]struct{} // whitelist for debug
}

// InitWhitelist 初始化whitelist
func InitWhitelist(wl []int64, wlFile string) (err error) {
	var (
		mid int64
		f   *os.File
	)
	if f, err = os.OpenFile(wlFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644); err == nil {
		whitelist = new(Whitelist)
		whitelist.log = log.New(f, "", log.LstdFlags)
		whitelist.list = make(map[int64]struct{})
		for _, mid = range wl {
			whitelist.list[mid] = struct{}{}
		}
	}
	return
}

// Contains 判断返回用户是否在白名单里面
func (w *Whitelist) Contains(uid int64) (ok bool) {
	if uid > 0 {
		_, ok = w.list[uid]
	}
	return
}

// Printf 打印白名单日志
func (w *Whitelist) Printf(format string, v ...interface{}) {
	w.log.Printf(format, v...)
}
