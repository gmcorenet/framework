package kernel

import (
	"github.com/gmcorenet/framework/internal/container"
	"github.com/gmcorenet/framework/internal/router"
	"github.com/gmcorenet/framework/pkg"
)

type Kernel struct {
	router    *router.Router
	container *container.Container
}

func New(r *router.Router, c *container.Container) *Kernel {
	return &Kernel{
		router:    r,
		container: c,
	}
}

func (k *Kernel) Handle(request *pkg.Request) *pkg.Response {
	route := k.router.Resolve(request.Method(), request.URI())

	if route == nil {
		return pkg.NewResponse("Not Found", 404)
	}

	params := route.ExtractParams(request.URI())
	request.WithParams(params)

	handler := route.Handler()

	if fn, ok := handler.(func(*pkg.Request) *pkg.Response); ok {
		return fn(request)
	}

	return k.container.Call(handler, params).(*pkg.Response)
}

func (k *Kernel) Terminate() {
}