package router

import (
	"regexp"
	"strings"
)

type Route struct {
	method       string
	path         string
	handler      interface{}
	middlewares  []interface{}
}

func NewRoute(method, path string, handler interface{}) *Route {
	return &Route{
		method:      method,
		path:        path,
		handler:     handler,
		middlewares: []interface{}{},
	}
}

func (r *Route) Method() string {
	return r.method
}

func (r *Route) Path() string {
	return r.path
}

func (r *Route) Handler() interface{} {
	return r.handler
}

func (r *Route) AddMiddleware(m interface{}) *Route {
	r.middlewares = append(r.middlewares, m)
	return r
}

func (r *Route) Middlewares() []interface{} {
	return r.middlewares
}

func (r *Route) ExtractParams(uri string) map[string]string {
	params := make(map[string]string)
	pattern := regexp.MustCompile(`\{(\w+)\}`)
	paramNames := pattern.FindAllStringSubmatch(r.path, -1)

	pathParts := strings.Split(strings.Trim(r.path, "/"), "/")
	uriParts := strings.Split(strings.Trim(uri, "/"), "/")

	if len(pathParts) != len(uriParts) {
		return params
	}

	for i, part := range pathParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			key := strings.Trim(part, "{}")
			params[key] = uriParts[i]
		}
	}

	return params
}

func (r *Route) Matches(method, uri string) bool {
	if r.method != method {
		return false
	}

	pathPattern := regexp.MustCompile(`\{(\w+)\}`)
	patternStr := pathPattern.ReplaceAllString(r.path, "[^/]+")
	patternStr = "^" + patternStr + "$"

	pattern := regexp.MustCompile(patternStr)
	return pattern.MatchString(uri)
}