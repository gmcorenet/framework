package router

import (
	"fmt"
	"regexp"
	"strings"
)

type RouteCollection struct {
	routes []*Route
	named  map[string]*Route
}

func NewRouteCollection() *RouteCollection {
	return &RouteCollection{
		routes: []*Route{},
		named:  make(map[string]*Route),
	}
}

func (c *RouteCollection) Add(route *Route) *RouteCollection {
	if route.name != "" {
		c.named[route.name] = route
	}
	c.routes = append(c.routes, route)
	return c
}

func (c *RouteCollection) All() []*Route {
	return c.routes
}

func (c *RouteCollection) Get(name string) *Route {
	if route, ok := c.named[name]; ok {
		return route
	}
	return nil
}

func (c *RouteCollection) Match(method, uri string) *Route {
	for _, route := range c.routes {
		if route.Matches(method, uri) {
			return route
		}
	}
	return nil
}

func (c *RouteCollection) MatchByName(name string) *Route {
	return c.Get(name)
}

func (c *RouteCollection) NamedRoutes() map[string]*Route {
	return c.named
}

func (c *RouteCollection) URL(name string, params map[string]string) (string, error) {
	route := c.Get(name)
	if route == nil {
		return "", fmt.Errorf("route not found: %s", name)
	}

	return route.GenerateURL(params)
}

func (r *Route) GenerateURL(params map[string]string) (string, error) {
	path := r.path

	for key, value := range r.defaults {
		if _, ok := params[key]; !ok {
			params[key] = value
		}
	}

	paramPattern := regexp.MustCompile(`\{(\w+)\}`)
	matches := paramPattern.FindAllStringSubmatch(path, -1)

	for _, match := range matches {
		placeholder := match[0]
		key := match[1]

		if val, ok := params[key]; ok && val != "" {
			path = strings.Replace(path, placeholder, val, 1)
		} else if requirement, ok := r.requirements[key]; ok {
			if ok && requirement != "" {
				continue
			}
			return "", fmt.Errorf("missing parameter: %s", key)
		}
	}

	unresolved := paramPattern.FindAllString(path, -1)
	if len(unresolved) > 0 {
		return "", fmt.Errorf("missing parameters: %v", unresolved)
	}

	return path, nil
}

func (c *RouteCollection) GetRoutesByMethod(method string) []*Route {
	var result []*Route
	for _, route := range c.routes {
		if route.method == method {
			result = append(result, route)
		}
	}
	return result
}

func (c *RouteCollection) GetRoutesByPath(pathPrefix string) []*Route {
	var result []*Route
	for _, route := range c.routes {
		if strings.HasPrefix(route.path, pathPrefix) {
			result = append(result, route)
		}
	}
	return result
}

func (c *RouteCollection) Count() int {
	return len(c.routes)
}

func (c *RouteCollection) CountNamed() int {
	return len(c.named)
}
