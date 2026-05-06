package routing

import (
	"log"
	"net/http"
	"reflect"

	"github.com/gmcorenet/framework/container"
	"github.com/gmcorenet/framework/router"
	"github.com/gmcorenet/framework/routing/annotations"
)

type RouteAttr struct {
	Path    string
	Action  string
	Methods []string
	Name    string
	Public  bool
}

type RouteProvider interface {
	ControllerRoutes() []RouteAttr
}

type controllerFactory func() interface{}

var registry = map[string]controllerFactory{}

func Register(factory func() interface{}) {
	ctrl := factory()
	ctrlType := reflect.TypeOf(ctrl)
	if ctrlType.Kind() == reflect.Ptr {
		ctrlType = ctrlType.Elem()
	}
	id := annotations.TypeToServiceID(ctrlType.Name())
	registry[id] = factory
}

type serviceFactory func() interface{}

var services = map[string]serviceFactory{}

func RegisterService(name string, factory func() interface{}) {
	services[name] = factory
}

type MiddlewareProvider func(ctr *container.Container, r *router.Router) (func(http.Handler) http.Handler, bool)

var middlewareProviders []MiddlewareProvider

func RegisterMiddlewareProvider(provider MiddlewareProvider) {
	middlewareProviders = append(middlewareProviders, provider)
}

func ApplyMiddlewares(ctr *container.Container, r *router.Router, handler http.Handler) http.Handler {
	for _, provider := range middlewareProviders {
		if mw, ok := provider(ctr, r); ok && mw != nil {
			handler = mw(handler)
		}
	}
	return handler
}

func PopulateControllers(ctr *container.Container) {
	for name, factory := range services {
		ctr.Set(name, factory())
	}
	for id, factory := range registry {
		ctr.Set(id, factory())
	}
}

func InjectAll(ctr *container.Container) {
	keys := ctr.Keys()
	for _, id := range keys {
		svc, err := ctr.Get(id)
		if err != nil || svc == nil {
			continue
		}
		ctr.Inject(svc)
	}
}

func MakeHandler(svc interface{}, action string) router.Handler {
	method := reflect.ValueOf(svc).MethodByName(action)
	if !method.IsValid() {
		return nil
	}
	return makeHandler(method)
}

func makeHandler(method reflect.Value) router.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		method.Call([]reflect.Value{
			reflect.ValueOf(w),
			reflect.ValueOf(r),
			reflect.ValueOf(params),
		})
	}
}

var _ = log.Default
