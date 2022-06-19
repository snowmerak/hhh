package loadbalancer

import (
	"errors"
	"math"

	"github.com/snowmerak/hhh/system/lock"
)

var errNotExist = errors.New("not exist")

func ErrNotExist() error {
	return errNotExist
}

var errEmptyValue = errors.New("empty value")

func ErrEmptyValue() error {
	return errEmptyValue
}

var errAlreadyExist = errors.New("already exist")

func ErrAlreadyExist() error {
	return errAlreadyExist
}

type LoadBalancer struct {
	candidates map[string]int64
	l          *lock.Lock
}

func New() *LoadBalancer {
	return &LoadBalancer{
		candidates: make(map[string]int64),
		l:          new(lock.Lock),
	}
}

func (l *LoadBalancer) Add(target string) error {
	l.l.Lock()
	defer l.l.Unlock()

	if _, ok := l.candidates[target]; ok {
		return ErrAlreadyExist()
	}

	l.candidates[target] = 0
	return nil
}

func (l *LoadBalancer) Sub(target string) error {
	l.l.Lock()
	defer l.l.Unlock()

	if _, ok := l.candidates[target]; !ok {
		return ErrNotExist()
	}

	delete(l.candidates, target)
	return nil
}

func (l *LoadBalancer) Get() (string, error) {
	l.l.Lock()
	defer l.l.Unlock()

	min := int64(math.MaxInt64)
	target := ""
	for k, v := range l.candidates {
		if v < min {
			target = k
			min = v
		}
	}

	if target == "" {
		return "", ErrEmptyValue()
	}

	l.candidates[target]++
	return target, nil
}

func (l *LoadBalancer) Restore(target string) error {
	l.l.Lock()
	defer l.l.Unlock()

	if _, ok := l.candidates[target]; !ok {
		return ErrNotExist()
	}

	l.candidates[target]--
	return nil
}
