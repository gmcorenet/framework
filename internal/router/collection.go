package router

type RouteCollection struct {
	routes []*Route
}

func NewRouteCollection() *RouteCollection {
	return &RouteCollection{routes: []*Route{}}
}

func (c *RouteCollection) Add(route *Route) {
	c.routes = append(c.routes, route)
}

func (c *RouteCollection) All() []*Route {
	return c.routes
}

func (c *RouteCollection) Match(method, uri string) *Route {
	for _, route := range c.routes {
		if route.Matches(method, uri) {
			return route
		}
	}
	return nil
}