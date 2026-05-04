package router

import (
	"fmt"
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
	Name       string
}

type Middleware func(http.Handler) http.Handler

type Router struct {
	routes            []Route
	prefix           string
	middleware       []Middleware
	notFound         http.Handler
	methodNotAllowed http.Handler
}

func New() *Router {
	return &Router{
		routes:     make([]Route, 0),
		middleware: make([]Middleware, 0),
		notFound:   http.HandlerFunc(http.NotFound),
	}
}

func (r *Router) Use(m Middleware) {
	r.middleware = append(r.middleware, m)
}

func (r *Router) Group(prefix string, callback func(*Router), middlewares ...Middleware) *Router {
	group := &Router{
		routes:      make([]Route, 0),
		prefix:      r.prefix + prefix,
		middleware:  append(r.middleware, middlewares...),
		notFound:    r.notFound,
	}

	callback(group)

	for _, route := range group.routes {
		r.routes = append(r.routes, route)
	}

	return r
}

func (r *Router) Handle(pattern string, methods []string, handler Handler) (*Route, error) {
	fullPattern := r.prefix + pattern
	fullPattern = "^" + strings.Trim(fullPattern, "/") + "/?$"

	compiled, err := regexp.Compile(fullPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid route pattern: %w", err)
	}

	route := Route{
		Pattern:    compiled,
		Methods:    methods,
		Handler:    handler,
		Middleware: make([]Middleware, len(r.middleware)),
	}
	copy(route.Middleware, r.middleware)

	r.routes = append(r.routes, route)
	return &r.routes[len(r.routes)-1], nil
}

func (r *Router) GET(pattern string, handler Handler) (*Route, error) {
	return r.Handle(pattern, []string{"GET", "HEAD"}, handler)
}

func (r *Router) POST(pattern string, handler Handler) (*Route, error) {
	return r.Handle(pattern, []string{"POST"}, handler)
}

func (r *Router) PUT(pattern string, handler Handler) (*Route, error) {
	return r.Handle(pattern, []string{"PUT"}, handler)
}

func (r *Router) DELETE(pattern string, handler Handler) (*Route, error) {
	return r.Handle(pattern, []string{"DELETE"}, handler)
}

func (r *Router) PATCH(pattern string, handler Handler) (*Route, error) {
	return r.Handle(pattern, []string{"PATCH"}, handler)
}

func (r *Router) OPTIONS(pattern string, handler Handler) (*Route, error) {
	return r.Handle(pattern, []string{"OPTIONS"}, handler)
}

func (r *Router) Match(methods []string, pattern string, handler Handler) (*Route, error) {
	return r.Handle(pattern, methods, handler)
}

func (r *Router) Any(pattern string, handler Handler) (*Route, error) {
	return r.Handle(pattern, []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}, handler)
}

func (r *Router) SetNotFound(handler http.Handler) {
	r.notFound = handler
}

func (r *Router) SetNotFoundFunc(fn func(http.ResponseWriter, *http.Request)) {
	r.notFound = http.HandlerFunc(fn)
}

func (r *Router) SetMethodNotAllowed(handler http.Handler) {
	r.methodNotAllowed = handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := strings.Trim(req.URL.Path, "/")
	methodMatched := false

	for _, route := range r.routes {
		if !r.matchesMethod(route, req.Method) {
			methodMatched = true
			continue
		}

		matches := route.Pattern.FindStringSubmatch(path)
		if matches != nil {
			params := r.extractParams(route.Pattern, path)

			finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				route.Handler(w, r, params)
			})

			for i := len(route.Middleware) - 1; i >= 0; i-- {
				finalHandler = route.Middleware[i](finalHandler)
			}

			finalHandler.ServeHTTP(w, req)
			return
		}
	}

	if methodMatched && r.methodNotAllowed != nil {
		r.methodNotAllowed.ServeHTTP(w, req)
		return
	}

	r.notFound.ServeHTTP(w, req)
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

func (r *Router) NamedRoutes() map[string]*Route {
	result := make(map[string]*Route)
	for i := range r.routes {
		if r.routes[i].Name != "" {
			result[r.routes[i].Name] = &r.routes[i]
		}
	}
	return result
}

func (r *Router) Route(name string) *Route {
	for i := range r.routes {
		if r.routes[i].Name == name {
			return &r.routes[i]
		}
	}
	return nil
}

func (r *Router) URL(name string, params map[string]string) (string, error) {
	route := r.Route(name)
	if route == nil {
		return "", &RouteNotFoundError{Name: name}
	}

	path := route.Pattern.String()
	for key, value := range params {
		path = strings.ReplaceAll(path, "(?P<"+key+">[^/]+)", value)
	}
	return path, nil
}

type RouteNotFoundError struct {
	Name string
}

func (e *RouteNotFoundError) Error() string {
	return "route not found: " + e.Name
}

func (r *Router) Routes() []Route {
	return r.routes
}

func (r *Router) Count() int {
	return len(r.routes)
}
