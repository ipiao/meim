package util

import "reflect"

// 使用的时候需要各种形式一致
// 方法内未检验
func MergeFunc(fns ...interface{}) interface{} {
	realfns := []reflect.Value{}
	for _, fn := range fns {
		if fn != nil {
			realfns = append(realfns, reflect.ValueOf(fn))
		}
	}

	if len(realfns) == 0 {
		return nil
	}

	res := reflect.MakeFunc(realfns[0].Type(), func(args []reflect.Value) (results []reflect.Value) {
		for _, rfn := range realfns {
			rfn.Call(args)
		}
		return
	})

	return res.Interface()
}
