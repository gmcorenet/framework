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

func (r *Router) Get(path string, handler interface{}) *Router {
	r.routes.Add(NewRoute("GET", path, handler))
	return r
}

func (r *Router) Post(path string, handler interface{}) *Router {
	r.routes.Add(NewRoute("POST", path, handler))
	return r
}

func (r *Router) Put(path string, handler interface{}) *Router {
	r.routes.Add(NewRoute("PUT", path, handler))
	return r
}

func (r *Router) Delete(path string, handler interface{}) *Router {
	r.routes.Add(NewRoute("DELETE", path, handler))
	return r
}

func (r *Router) Patch(path string, handler interface{}) *Router {
	r.routes.Add(NewRoute("PATCH", path, handler))
	return r
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