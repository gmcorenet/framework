package pkg

import (
	"net/http"
	"regexp"
	"strings"
)

type Request struct {
	method  string
	uri     string
	params  map[string]string
	headers map[string]string
	body    map[string]interface{}
	query   map[string]string
}

var requestInstance *Request

func NewRequest() *Request {
	if requestInstance == nil {
		requestInstance = &Request{
			method:  http.MethodGet,
			uri:     "/",
			params:  make(map[string]string),
			headers: make(map[string]string),
			body:    make(map[string]interface{}),
			query:   make(map[string]string),
		}
	}
	return requestInstance
}

func (r *Request) Method() string {
	return r.method
}

func (r *Request) URI() string {
	return r.uri
}

func (r *Request) Params() map[string]string {
	return r.params
}

func (r *Request) WithParams(params map[string]string) *Request {
	r.params = params
	return r
}

func (r *Request) Header(name string) string {
	return r.headers[name]
}

func (r *Request) Headers() map[string]string {
	return r.headers
}

func (r *Request) Body() map[string]interface{} {
	return r.body
}

func (r *Request) Query(key string) string {
	return r.query[key]
}

func (r *Request) Get(key string) string {
	return r.query[key]
}

func (r *Request) Post(key string) interface{} {
	return r.body[key]
}