package utils

import (
	"fmt"
	"runtime/debug"
)

func GoSafe(fn func()) {
	go RunSafe(fn)
}

func RunSafe(fn func()) {
	//todo panic 记录、上报
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("RunSafe capture crash,msg: %s \n %s ", fmt.Sprint(p), string(debug.Stack()))
		}
	}()
	fn()
}
