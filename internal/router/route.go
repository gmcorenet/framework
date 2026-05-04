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
	name         string
	defaults     map[string]string
	requirements map[string]string
	host         string
	schemes      []string
	options      map[string]interface{}
}

func NewRoute(method, path string, handler interface{}) *Route {
	return &Route{
		method:       method,
		path:         path,
		handler:      handler,
		middlewares:  []interface{}{},
		defaults:     make(map[string]string),
		requirements: make(map[string]string),
		options:      make(map[string]interface{}),
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

func (r *Route) SetName(name string) *Route {
	r.name = name
	return r
}

func (r *Route) Name() string {
	return r.name
}

func (r *Route) SetDefaults(defaults map[string]string) *Route {
	r.defaults = defaults
	return r
}

func (r *Route) SetDefault(key, value string) *Route {
	r.defaults[key] = value
	return r
}

func (r *Route) Defaults() map[string]string {
	return r.defaults
}

func (r *Route) GetDefault(key string) string {
	if val, ok := r.defaults[key]; ok {
		return val
	}
	return ""
}

func (r *Route) SetRequirements(requirements map[string]string) *Route {
	r.requirements = requirements
	return r
}

func (r *Route) SetRequirement(key, pattern string) *Route {
	r.requirements[key] = pattern
	return r
}

func (r *Route) Requirements() map[string]string {
	return r.requirements
}

func (r *Route) GetRequirement(key string) string {
	if val, ok := r.requirements[key]; ok {
		return val
	}
	return ""
}

func (r *Route) HasRequirement(key string) bool {
	_, ok := r.requirements[key]
	return ok
}

func (r *Route) SetHost(host string) *Route {
	r.host = host
	return r
}

func (r *Route) Host() string {
	return r.host
}

func (r *Route) SetSchemes(schemes []string) *Route {
	r.schemes = schemes
	return r
}

func (r *Route) Schemes() []string {
	return r.schemes
}

func (r *Route) SetOption(key string, value interface{}) *Route {
	r.options[key] = value
	return r
}

func (r *Route) GetOption(key string) interface{} {
	if val, ok := r.options[key]; ok {
		return val
	}
	return nil
}

func (r *Route) Options() map[string]interface{} {
	return r.options
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

			if val, ok := r.defaults[key]; ok && params[key] == "" {
				params[key] = val
			}
		}
	}

	for key, val := range r.defaults {
		if _, ok := params[key]; !ok {
			params[key] = val
		}
	}

	return params
}

func (r *Route) BuildPattern() *regexp.Regexp {
	pathPattern := r.path

	for key, requirement := range r.requirements {
		placeholder := "{" + key + "}"
		if strings.Contains(pathPattern, placeholder) {
			pathPattern = strings.ReplaceAll(pathPattern, placeholder, "("+requirement+")")
		}
	}

	pathPattern = regexp.MustCompile(`\{(\w+)\}`).ReplaceAllString(pathPattern, "([^/]+)")
	pathPattern = "^" + pathPattern + "$"

	return regexp.MustCompile(pathPattern)
}

func (r *Route) Matches(method, uri string) bool {
	if r.method != method {
		return false
	}

	pattern := r.BuildPattern()
	return pattern.MatchString(uri)
}

func (r *Route) MatchesWithScheme(method, scheme, uri string) bool {
	if !r.Matches(method, uri) {
		return false
	}

	if len(r.schemes) > 0 {
		found := false
		for _, s := range r.schemes {
			if s == scheme {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (r *Route) ValidateRequirements() error {
	pattern := regexp.MustCompile(`\{(\w+)\}`)
	paramNames := pattern.FindAllStringSubmatch(r.path, -1)

	for _, match := range paramNames {
		paramName := match[1]
		if _, ok := r.requirements[paramName]; !ok {
			r.requirements[paramName] = "[^/]+"
		}
	}

	return nil
}
