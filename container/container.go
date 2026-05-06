package container

import (
	"fmt"
	"reflect"
	"sync"
)

type Container struct {
	services  map[string]interface{}
	instances map[string]interface{}
	aliases   map[string]string
	tags      map[string][]string
	factories map[string]func() interface{}
	mu        sync.RWMutex
}

func NewContainer() *Container {
	return &Container{
		services:  make(map[string]interface{}),
		instances: make(map[string]interface{}),
		aliases:   make(map[string]string),
		tags:      make(map[string][]string),
		factories: make(map[string]func() interface{}),
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

	if instance, ok := c.instances[key]; ok {
		return instance, nil
	}

	if service, ok := c.services[key]; ok {
		return service, nil
	}

	if alias, ok := c.aliases[key]; ok {
		if instance, ok := c.instances[alias]; ok {
			return instance, nil
		}
		if service, ok := c.services[alias]; ok {
			return service, nil
		}
	}

	return nil, fmt.Errorf("service not found: %s", key)
}

func (c *Container) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, hasService := c.services[key]
	_, hasInstance := c.instances[key]
	_, hasAlias := c.aliases[key]
	return hasService || hasInstance || hasAlias
}

func (c *Container) Make(key string) (interface{}, error) {
	return c.Get(key)
}

func (c *Container) Bind(key string, factory func() interface{}) {
	c.Set(key, factory())
}

func (c *Container) Singleton(key string, instance interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[key] = instance
	c.instances[key] = instance
}

func (c *Container) Call(function interface{}, params map[string]interface{}) ([]interface{}, error) {
	fn := reflect.ValueOf(function)
	fnType := fn.Type()

	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("function must be a func, got %s", fnType.Kind())
	}

	args := make([]reflect.Value, fnType.NumIn())
	for i := 0; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)
		service, _ := c.Get(paramType.Name())
		if service != nil {
			args[i] = reflect.ValueOf(service)
		} else if val, ok := params[paramType.Name()]; ok {
			args[i] = reflect.ValueOf(val)
		} else if paramType.Kind() == reflect.Ptr {
			args[i] = reflect.Zero(paramType)
		} else {
			args[i] = reflect.Zero(paramType)
		}
	}

	results := fn.Call(args)
	output := make([]interface{}, len(results))
	for i, v := range results {
		output[i] = v.Interface()
	}
	return output, nil
}

func (c *Container) Inject(target interface{}) error {
	t := reflect.ValueOf(target)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer to struct")
	}

	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Type().Field(i)
		fieldValue := t.Field(i)

		injectTag := field.Tag.Get("inject")
		if injectTag == "" {
			continue
		}

		serviceID := injectTag
		if serviceID == "" {
			serviceID = field.Name
		}

		if !fieldValue.CanSet() {
			continue
		}

		service, err := c.Get(serviceID)
		if err != nil {
			continue
		}

		fieldValue.Set(reflect.ValueOf(service))
	}

	return nil
}

func (c *Container) Factory(key string, factory func() interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[key] = factory
}

func (c *Container) GetOrMake(key string) interface{} {
	c.mu.RLock()
	if instance, ok := c.instances[key]; ok {
		c.mu.RUnlock()
		return instance
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if instance, ok := c.instances[key]; ok {
		return instance
	}

	if factory, ok := c.factories[key]; ok {
		instance := factory()
		c.instances[key] = instance
		return instance
	}

	if service, ok := c.services[key]; ok {
		c.instances[key] = service
		return service
	}

	return nil
}

func (c *Container) Alias(alias, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aliases[alias] = key
}

func (c *Container) Tag(key string, tags ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, tag := range tags {
		found := false
		for _, existingID := range c.tags[tag] {
			if existingID == key {
				found = true
				break
			}
		}
		if !found {
			c.tags[tag] = append(c.tags[tag], key)
		}
	}
}

func (c *Container) Tagged(tagName string) []interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]interface{}, 0)
	for _, serviceID := range c.tags[tagName] {
		if service, ok := c.services[serviceID]; ok {
			result = append(result, service)
		} else if instance, ok := c.instances[serviceID]; ok {
			result = append(result, instance)
		}
	}
	return result
}

func (c *Container) Remove(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.services, id)
	delete(c.instances, id)
	delete(c.factories, id)

	for tag, serviceIDs := range c.tags {
		newIDs := make([]string, 0)
		for _, sid := range serviceIDs {
			if sid != id {
				newIDs = append(newIDs, sid)
			}
		}
		c.tags[tag] = newIDs
	}

	for alias, targetID := range c.aliases {
		if targetID == id {
			delete(c.aliases, alias)
		}
	}
}

func (c *Container) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	seen := make(map[string]struct{})
	for k := range c.services {
		seen[k] = struct{}{}
	}
	for k := range c.instances {
		seen[k] = struct{}{}
	}

	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	return keys
}
