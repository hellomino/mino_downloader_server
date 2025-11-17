package utils

import (
	"minodl/log"
	"runtime/debug"
)

// SafeCall 并行安全调用
func SafeCall(call func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("SafeCall panic: %v, stack:%s", r, string(debug.Stack()))
		}
	}()
	call()
}
