package loadbalancer

import "net/http"

type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func NewResponse() *Response {
	return &Response{
		StatusCode: 200,
		Headers:    make(http.Header),
		Body:       []byte{},
	}
}

func (r *Response) WriteHeader(code int) {
	r.StatusCode = code
}

func (r *Response) Header() http.Header {
	return r.Headers
}

func (r *Response) Write(p []byte) (int, error) {
	r.Body = append(r.Body, p...)
	return len(p), nil
}
