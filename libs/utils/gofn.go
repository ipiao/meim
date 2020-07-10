package utils

import (
	"log"
	"runtime"
)

func RecoverDo(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1<<16)
			n := runtime.Stack(buf, false)
			buf = buf[:n]
			log.Println(string(buf))
		}
	}()
	fn()
}

func GoFunc(fn func()) {
	go RecoverDo(fn)
}
