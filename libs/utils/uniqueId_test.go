package utils

import (
	"fmt"
	"testing"
)

func TestGenerate(t *testing.T) {
	u := NewUniqueId(0, 3)

	var idc = make(chan uint32, 999999999)
	go func() {
		for id := range idc {
			fmt.Println(id)
		}
	}()
	for i := 0; i < 20; i++ {
		go func() {
			for {
				id := u.Generate()
				if id > u.max {
					t.FailNow()
				}
				idc <- id
			}
		}()
	}
	select {}
}
