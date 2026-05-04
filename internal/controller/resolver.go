package controller

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
)

type Resolver struct {
	container  interface{}
	arguments  *ArgumentResolver
	factories  map[string]ControllerFactory
}

type ControllerFactory func() interface{}

type ControllerInfo struct {
	Type       reflect.Type
	Factory    ControllerFactory
	Methods    []string
}

func NewResolver(container interface{}) *Resolver {
	return &Resolver{
		container:  container,
		arguments:  NewArgumentResolver(container),
		factories:  make(map[string]ControllerFactory),
	}
}

func (r *Resolver) Register(controllerName string, factory ControllerFactory) {
	r.factories[controllerName] = factory
}

func (r *Resolver) Resolve(controllerName, method string, w http.ResponseWriter, req *http.Request, params map[string]string) error {
	factory, ok := r.factories[controllerName]
	if !ok {
		return fmt.Errorf("controller not registered: %s", controllerName)
	}

	ctrl := factory()

	ctrlType := reflect.TypeOf(ctrl)
	if ctrlType.Kind() != reflect.Ptr {
		ctrlType = reflect.PtrTo(ctrlType)
	}

	methodVal := ctrlType.MethodByName(method)
	if !methodVal.IsValid() {
		return fmt.Errorf("method not found: %s on controller %s", method, controllerName)
	}

	ctx := req.Context()
	ctx = context.WithValue(ctx, "request", req)
	ctx = context.WithValue(ctx, "response", w)
	ctx = context.WithValue(ctx, "params", params)
	ctx = context.WithValue(ctx, "container", r.container)

	arguments, err := r.arguments.Resolve(methodVal, ctx, req, w, params)
	if err != nil {
		return fmt.Errorf("failed to resolve arguments for %s.%s: %w", controllerName, method, err)
	}

	results := methodVal.Func.Call(arguments)

	if len(results) > 0 {
		if lastErr, ok := results[len(results)-1].Interface().(error); ok && lastErr != nil {
			return lastErr
		}
	}

	return nil
}

func (r *Resolver) GetControllers() map[string]ControllerFactory {
	return r.factories
}