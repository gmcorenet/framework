package container

import (
	"reflect"
)

type Container struct {
	services  map[string]interface{}
	instances map[string]interface{}
}

func New() *Container {
	return &Container{
		services:  make(map[string]interface{}),
		instances: make(map[string]interface{}),
	}
}

func (c *Container) Set(id string, concrete interface{}) *Container {
	c.services[id] = concrete
	return c
}

func (c *Container) Get(id string) interface{} {
	if instance, ok := c.instances[id]; ok {
		return instance
	}

	if service, ok := c.services[id]; ok {
		c.instances[id] = service
		return service
	}

	return nil
}

func (c *Container) Has(id string) bool {
	_, hasService := c.services[id]
	_, hasInstance := c.instances[id]
	return hasService || hasInstance
}

func (c *Container) Call(function interface{}, params map[string]interface{}) interface{} {
	fn := reflect.ValueOf(function)
	fnType := fn.Type()

	args := make([]reflect.Value, fnType.NumIn())

	for i := 0; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)
		if service, ok := c.services[paramType.Name()]; ok {
			args[i] = reflect.ValueOf(service)
		} else if val, ok := params[fnType.Param(i).Name]; ok {
			args[i] = reflect.ValueOf(val)
		} else {
			args[i] = reflect.Zero(paramType)
		}
	}

	return fn.Call(args)
}