package circuitbreaker

import (
	"errors"
	"net/http/httptest"
	"net/http/httputil"
	"time"

	"github.com/snowmerak/hhh/loadbalancer"
	"github.com/snowmerak/hhh/system/lock"
)

var errAlreadyExist = errors.New("already exist")

func ErrAlreadyExist() error {
	return errAlreadyExist
}

type CurcuitBreaker struct {
	lock     *lock.Lock
	balancer *loadbalancer.LoadBalancer
	set      map[string]*httputil.ReverseProxy
}

func New(balancer *loadbalancer.LoadBalancer) *CurcuitBreaker {
	cb := &CurcuitBreaker{
		lock:     new(lock.Lock),
		balancer: balancer,
		set:      make(map[string]*httputil.ReverseProxy),
	}
	go func() {
		for {
			names := []string{}
			oks := []*httputil.ReverseProxy{}
			for k, v := range cb.set {
				req := httptest.NewRequest("GET", "/", nil)
				v.Director(req)
				statusCode := req.Response.StatusCode
				if statusCode == 200 {
					names = append(names, k)
					oks = append(oks, v)
					delete(cb.set, k)
				}
			}
			for i := 0; i < len(names) && i < len(oks); i++ {
				cb.balancer.Append(names[i], oks[i])
			}
			time.Sleep(time.Second)
		}
	}()
	return cb
}

func (b *CurcuitBreaker) Add(target string, server *httputil.ReverseProxy) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if _, ok := b.set[target]; ok {
		return ErrAlreadyExist()
	}

	b.set[target] = server
	return nil
}
