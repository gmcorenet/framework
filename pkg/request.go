package pkg

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type Request struct {
	method  string
	uri     string
	params  map[string]string
	headers map[string]string
	body    map[string]interface{}
	query   map[string]string
}

var (
	requestInstance *Request
	requestOnce     sync.Once
)

func NewRequest() *Request {
	requestOnce.Do(func() {
		requestInstance = &Request{
			method:  http.MethodGet,
			uri:     "/",
			params:  make(map[string]string),
			headers: make(map[string]string),
			body:    make(map[string]interface{}),
			query:   make(map[string]string),
		}
	})
	return requestInstance
}

func NewRequestFromHTTP(r *http.Request) *Request {
	return &Request{
		method:  r.Method,
		uri:     r.RequestURI,
		params:  make(map[string]string),
		headers: make(map[string]string),
		body:    make(map[string]interface{}),
		query:   make(map[string]string),
	}
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

func (r *Request) SetMethod(method string) *Request {
	r.method = method
	return r
}

func (r *Request) SetURI(uri string) *Request {
	r.uri = uri
	return r
}

func (r *Request) SetHeader(name, value string) *Request {
	r.headers[name] = value
	return r
}

func (r *Request) SetBody(body map[string]interface{}) *Request {
	r.body = body
	return r
}

func (r *Request) SetQuery(query map[string]string) *Request {
	r.query = query
	return r
}

func (r *Request) SetParam(key, value string) *Request {
	r.params[key] = value
	return r
}

func (r *Request) IsAJAX() bool {
	return r.headers["X-Requested-With"] == "XMLHttpRequest"
}

func (r *Request) IsJson() bool {
	return strings.Contains(r.headers["Accept"], "application/json")
}

func (r *Request) GetClientIP() string {
	return r.headers["X-Forwarded-For"]
}

func (r *Request) GetBearerToken() string {
	auth := r.headers["Authorization"]
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

func (r *Request) HasHeader(name string) bool {
	_, ok := r.headers[name]
	return ok
}

func (r *Request) MethodIs(methods ...string) bool {
	for _, m := range methods {
		if r.method == m {
			return true
		}
	}
	return false
}

func (r *Request) MarshalJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Request) Validate() error {
	if r.method == "" {
		return &RequestError{Message: "method is required"}
	}
	return nil
}

type RequestError struct {
	Message string
}

func (e *RequestError) Error() string {
	return e.Message
}

var numberRegex = regexp.MustCompile(`^-?\d+(\.\d+)?`)

func parseInt64(s string) (int64, bool) {
	m := numberRegex.FindString(s)
	if m == "" {
		return 0, false
	}
	n := int64(0)
	negative := false
	for _, c := range m {
		if c == '-' {
			negative = true
			continue
		}
		n = n*10 + int64(c-'0')
	}
	if negative {
		n = -n
	}
	return n, true
}
