package loadbalancer

import (
	"errors"
	"fmt"
	"math"
	"net/http/httputil"
	"net/url"

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

type proxy struct {
	server *httputil.ReverseProxy
	count  int64
}

type LoadBalancer struct {
	candidates map[string]*proxy
	l          *lock.Lock
}

func New() *LoadBalancer {
	return &LoadBalancer{
		candidates: make(map[string]*proxy),
		l:          new(lock.Lock),
	}
}

func (l *LoadBalancer) Add(target string) error {
	l.l.Lock()
	defer l.l.Unlock()

	if _, ok := l.candidates[target]; ok {
		return ErrAlreadyExist()
	}

	l.candidates[target] = &proxy{
		count: 0,
	}
	url, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("LoadBalancer.Add: url.Parse: %s", err)
	}
	l.candidates[target].server = httputil.NewSingleHostReverseProxy(url)
	return nil
}

func (l *LoadBalancer) Append(target string, server *httputil.ReverseProxy) error {
	l.l.Lock()
	defer l.l.Unlock()

	if _, ok := l.candidates[target]; ok {
		return ErrAlreadyExist()
	}

	l.candidates[target].count = 0
	l.candidates[target].server = server
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

func (l *LoadBalancer) Get() (string, *httputil.ReverseProxy, error) {
	l.l.Lock()
	defer l.l.Unlock()

	min := int64(math.MaxInt64)
	target := ""
	for k, v := range l.candidates {
		if v.count < min {
			target = k
			min = v.count
		}
	}

	if target == "" {
		return "", nil, ErrEmptyValue()
	}

	l.candidates[target].count++
	return target, l.candidates[target].server, nil
}

func (l *LoadBalancer) Restore(target string) error {
	l.l.Lock()
	defer l.l.Unlock()

	if _, ok := l.candidates[target]; !ok {
		return ErrNotExist()
	}

	l.candidates[target].count--
	return nil
}
