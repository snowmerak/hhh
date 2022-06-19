package lock

import (
	"runtime"
	"sync/atomic"
)

type Lock struct {
	i int64
}

func (l *Lock) Lock() {
	for !atomic.CompareAndSwapInt64(&l.i, 0, 1) {
		runtime.Gosched()
	}
}
func (l *Lock) Unlock() {
	atomic.StoreInt64(&l.i, 0)
}
