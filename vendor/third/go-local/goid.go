package local

import (
	"third/goroutine"
)

func Goid() uint64 {
	return uint64(goroutine.GoroutineId())
}
