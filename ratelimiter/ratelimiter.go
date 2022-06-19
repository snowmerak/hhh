package ratelimiter

import (
	"time"

	"github.com/snowmerak/hhh/system/lock"
)

type RateLimiter struct {
	lock       *lock.Lock
	unit       int64
	maxConnPer float64
	prevTime   int64
	prevCount  int64
	curCount   int64
	nextTime   int64
}

func New(maxConnPer float64, unit time.Duration) *RateLimiter {
	now := int64(time.Now().UnixNano())
	return &RateLimiter{
		lock:       new(lock.Lock),
		unit:       int64(unit),
		maxConnPer: maxConnPer,
		prevTime:   now - int64(unit),
		prevCount:  0,
		curCount:   0,
		nextTime:   now + int64(unit),
	}
}

func (r *RateLimiter) TryTake() bool {
	r.lock.Lock()
	defer r.lock.Unlock()

	now := int64(time.Now().UnixNano())

	if now < r.prevTime {
		return false
	}

	for now > r.nextTime {
		r.prevTime = r.nextTime - r.unit
		r.prevCount = r.curCount
		r.curCount = 0
		r.nextTime = r.nextTime + r.unit
	}

	req := float64(r.prevCount)*float64(-now+r.prevTime+2*r.unit)/float64(r.unit) + float64(r.curCount+1)
	if req > r.maxConnPer {
		return false
	}

	r.curCount++

	return true
}

func (r *RateLimiter) Restore(count int64) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.prevCount -= count
	if r.prevCount < 0 {
		r.prevCount = 0
	}
}
