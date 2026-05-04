package router

type Router struct {
	routes      *RouteCollection
	middlewares []interface{}
}

func New() *Router {
	return &Router{
		routes: NewRouteCollection(),
	}
}

func (r *Router) Get(path string, handler interface{}) *Route {
	route := NewRoute("GET", path, handler)
	r.routes.Add(route)
	return route
}

func (r *Router) Post(path string, handler interface{}) *Route {
	route := NewRoute("POST", path, handler)
	r.routes.Add(route)
	return route
}

func (r *Router) Put(path string, handler interface{}) *Route {
	route := NewRoute("PUT", path, handler)
	r.routes.Add(route)
	return route
}

func (r *Router) Delete(path string, handler interface{}) *Route {
	route := NewRoute("DELETE", path, handler)
	r.routes.Add(route)
	return route
}

func (r *Router) Patch(path string, handler interface{}) *Route {
	route := NewRoute("PATCH", path, handler)
	r.routes.Add(route)
	return route
}

func (r *Router) Options(path string, handler interface{}) *Route {
	route := NewRoute("OPTIONS", path, handler)
	r.routes.Add(route)
	return route
}

func (r *Router) Any(path string, handler interface{}) *Router {
	r.Get(path, handler)
	r.Post(path, handler)
	r.Put(path, handler)
	r.Delete(path, handler)
	r.Patch(path, handler)
	return r
}

func (r *Router) Match(methods []string, path string, handler interface{}) *Route {
	for _, method := range methods {
		route := NewRoute(method, path, handler)
		r.routes.Add(route)
	}
	return nil
}

func (r *Router) Middleware(middleware interface{}) *Router {
	r.middlewares = append(r.middlewares, middleware)
	return r
}

func (r *Router) Group(callback func(*Router), middlewares []interface{}) *Router {
	group := New()
	callback(group)
	for _, route := range group.routes.All() {
		for _, m := range r.middlewares {
			route.AddMiddleware(m)
		}
		for _, m := range middlewares {
			route.AddMiddleware(m)
		}
		r.routes.Add(route)
	}
	return r
}

func (r *Router) Resolve(method, uri string) *Route {
	return r.routes.Match(method, uri)
}

func (r *Router) Routes() *RouteCollection {
	return r.routes
}

func (r *Router) URL(name string, params map[string]string) (string, error) {
	return r.routes.URL(name, params)
}

func (r *Router) NamedRoute(name string) *Route {
	return r.routes.Get(name)
}

func (r *Router) AddRoute(route *Route) *Router {
	r.routes.Add(route)
	return r
}
