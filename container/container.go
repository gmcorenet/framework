package container

import (
	"fmt"
	"sync"
)

type Container struct {
	services map[string]interface{}
	mu       sync.RWMutex
}

func NewContainer() *Container {
	return &Container{
		services: make(map[string]interface{}),
	}
}

func (c *Container) Set(key string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[key] = service
}

func (c *Container) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	service, ok := c.services[key]
	if !ok {
		return nil, fmt.Errorf("service not found: %s", key)
	}
	return service, nil
}

func (c *Container) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.services[key]
	return ok
}

func (c *Container) Make(key string) (interface{}, error) {
	return c.Get(key)
}

func (c *Container) Bind(key string, factory func() interface{}) {
	c.Set(key, factory())
}

func (c *Container) Singleton(key string, instance interface{}) {
	c.Set(key, instance)
}

