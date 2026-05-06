package routing

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/gmcorenet/framework/container"
	"github.com/gmcorenet/framework/router"
	"gopkg.in/yaml.v3"
)

type RouteDefinition struct {
	Path       string   `yaml:"path"`
	Controller string   `yaml:"controller"`
	Action     string   `yaml:"action"`
	Methods    []string `yaml:"methods"`
	Name       string   `yaml:"name"`
	Public     bool     `yaml:"public"`
}

type RouteConfig struct {
	Prefix string                    `yaml:"prefix"`
	Routes map[string]RouteDefinition `yaml:"routes"`
}

type ExposedRoute struct {
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
}

func LoadRoutes(path string) (*RouteConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read routes file %s: %w", path, err)
	}

	var cfg RouteConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse routes file %s: %w", path, err)
	}

	return &cfg, nil
}

var pathParamRegex = regexp.MustCompile(`\{(\w+)\}`)

func convertPathParams(path string) string {
	return pathParamRegex.ReplaceAllString(path, `(?P<$1>[^/]+)`)
}

func RegisterRoutes(r *router.Router, c *container.Container, cfg *RouteConfig) ([]ExposedRoute, error) {
	var exposed []ExposedRoute

	for routeName, def := range cfg.Routes {
		controller, err := c.Get(def.Controller)
		if err != nil {
			return nil, fmt.Errorf("route %q: controller %q not found in container: %w", routeName, def.Controller, err)
		}

		handler, err := resolveAction(controller, def.Action)
		if err != nil {
			return nil, fmt.Errorf("route %q: %w", routeName, err)
		}

		methods := def.Methods
		if len(methods) == 0 {
			methods = []string{"GET"}
		}

		fullPath := cfg.Prefix + def.Path
		regexPath := convertPathParams(fullPath)

		route, err := r.Handle(regexPath, methods, handler)
		if err != nil {
			return nil, fmt.Errorf("route %q: invalid pattern %q: %w", routeName, fullPath, err)
		}

		name := def.Name
		if name == "" {
			name = routeName
		}
		route.Name = name

		if def.Public {
			exposed = append(exposed, ExposedRoute{
				Name:    name,
				Path:    fullPath,
				Methods: methods,
			})
		}
	}

	return exposed, nil
}

func resolveAction(controller interface{}, action string) (router.Handler, error) {
	ctrlVal := reflect.ValueOf(controller)

	method := ctrlVal.MethodByName(action)
	if !method.IsValid() {
		return nil, fmt.Errorf("action %q not found on controller %T", action, controller)
	}

	methodType := method.Type()

	if methodType.NumIn() != 3 {
		return nil, fmt.Errorf("action %q has wrong signature: expected func(http.ResponseWriter, *http.Request, map[string]string), got %d params", action, methodType.NumIn())
	}

	if methodType.NumOut() != 0 {
		return nil, fmt.Errorf("action %q must return nothing", action)
	}

	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		method.Call([]reflect.Value{
			reflect.ValueOf(w),
			reflect.ValueOf(r),
			reflect.ValueOf(params),
		})
	}, nil
}

func normalizeControllerID(name string) string {
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	return strings.ToLower(name)
}
