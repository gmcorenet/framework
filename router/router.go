package router

import (
	"net/http"
	"regexp"
	"strings"
)

type Handler func(w http.ResponseWriter, r *http.Request, params map[string]string)

type Route struct {
	Pattern    *regexp.Regexp
	Methods    []string
	Handler    Handler
	Middleware []Middleware
}

type Middleware func(http.Handler) http.Handler

type Router struct {
	routes     []Route
	prefix     string
	middleware []Middleware
}

func New() *Router {
	return &Router{
		routes:     make([]Route, 0),
		middleware: make([]Middleware, 0),
	}
}

func (r *Router) Use(m Middleware) {
	r.middleware = append(r.middleware, m)
}

func (r *Router) Group(prefix string) *Router {
	return &Router{
		routes:     r.routes,
		prefix:     r.prefix + prefix,
		middleware: r.middleware,
	}
}

func (r *Router) Handle(pattern string, methods []string, handler Handler) {
	pattern = r.prefix + pattern
	pattern = "^" + strings.Trim(pattern, "/") + "/?$"
	r.routes = append(r.routes, Route{
		Pattern:    regexp.MustCompile(pattern),
		Methods:    methods,
		Handler:    handler,
		Middleware: r.middleware,
	})
}

func (r *Router) GET(pattern string, handler Handler) {
	r.Handle(pattern, []string{"GET", "HEAD"}, handler)
}

func (r *Router) POST(pattern string, handler Handler) {
	r.Handle(pattern, []string{"POST"}, handler)
}

func (r *Router) PUT(pattern string, handler Handler) {
	r.Handle(pattern, []string{"PUT"}, handler)
}

func (r *Router) DELETE(pattern string, handler Handler) {
	r.Handle(pattern, []string{"DELETE"}, handler)
}

func (r *Router) PATCH(pattern string, handler Handler) {
	r.Handle(pattern, []string{"PATCH"}, handler)
}

func (r *Router) OPTIONS(pattern string, handler Handler) {
	r.Handle(pattern, []string{"OPTIONS"}, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := strings.Trim(req.URL.Path, "/")

	for _, route := range r.routes {
		if !r.matchesMethod(route, req.Method) {
			continue
		}

		matches := route.Pattern.FindStringSubmatch(path)
		if matches != nil {
			params := r.extractParams(route.Pattern, path)
			route.Handler(w, req, params)
			return
		}
	}

	http.NotFound(w, req)
}

func (r *Router) matchesMethod(route Route, method string) bool {
	for _, m := range route.Methods {
		if m == method {
			return true
		}
	}
	return false
}

func (r *Router) extractParams(pattern *regexp.Regexp, path string) map[string]string {
	names := pattern.SubexpNames()
	matches := pattern.FindStringSubmatch(path)

	params := make(map[string]string)
	for i, match := range matches {
		if i == 0 {
			continue
		}
		if names[i] != "" {
			params[names[i]] = match
		}
	}
	return params
}

