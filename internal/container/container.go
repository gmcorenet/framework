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

func New() *Container {
	return &Container{
		services:  make(map[string]interface{}),
		instances: make(map[string]interface{}),
		aliases:   make(map[string]string),
		tags:      make(map[string][]string),
		factories: make(map[string]func() interface{}),
	}
}

func (c *Container) Set(id string, concrete interface{}) *Container {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[id] = concrete
	return c
}

func (c *Container) Get(id string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if instance, ok := c.instances[id]; ok {
		return instance
	}

	if service, ok := c.services[id]; ok {
		return service
	}

	if alias, ok := c.aliases[id]; ok {
		if instance, ok := c.instances[alias]; ok {
			return instance
		}
		if service, ok := c.services[alias]; ok {
			return service
		}
	}

	return nil
}

func (c *Container) Has(id string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, hasService := c.services[id]
	_, hasInstance := c.instances[id]
	_, hasAlias := c.aliases[id]

	if hasService || hasInstance || hasAlias {
		return true
	}

	return false
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

		if service := c.Get(paramType.Name()); service != nil {
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

func (c *Container) Alias(alias, id string) *Container {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aliases[alias] = id
	return c
}

func (c *Container) HasAlias(alias string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.aliases[alias]
	return ok
}

func (c *Container) GetAlias(id string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if alias, ok := c.aliases[id]; ok {
		return alias
	}
	return ""
}

func (c *Container) GetAliases() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]string)
	for k, v := range c.aliases {
		result[k] = v
	}
	return result
}

func (c *Container) RemoveAlias(alias string) *Container {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.aliases, alias)
	return c
}

func (c *Container) Tag(serviceId string, tags ...string) *Container {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, tag := range tags {
		for _, existingId := range c.tags[tag] {
			if existingId == serviceId {
				continue
			}
		}
		c.tags[tag] = append(c.tags[tag], serviceId)
	}

	return c
}

func (c *Container) Tagged(tagName string) []interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]interface{}, 0)
	for _, serviceId := range c.tags[tagName] {
		if service, ok := c.services[serviceId]; ok {
			result = append(result, service)
		} else if instance, ok := c.instances[serviceId]; ok {
			result = append(result, instance)
		}
	}

	return result
}

func (c *Container) HasTag(tagName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.tags[tagName]
	return ok && len(c.tags[tagName]) > 0
}

func (c *Container) GetTags() map[string][]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string][]string)
	for k, v := range c.tags {
		result[k] = v
	}
	return result
}

func (c *Container) GetTaggedServices(tagName string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	services := make([]string, len(c.tags[tagName]))
	copy(services, c.tags[tagName])
	return services
}

func (c *Container) RegisterTagged(tagName string, factory func(interface{})) *Container {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, serviceId := range c.tags[tagName] {
		if service, ok := c.services[serviceId]; ok {
			factory(service)
		}
	}

	return c
}

func (c *Container) Singleton(id string, concrete interface{}) *Container {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[id] = concrete
	c.instances[id] = concrete
	return c
}

func (c *Container) Factory(id string, factory func() interface{}) *Container {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[id] = factory
	return c
}

func (c *Container) GetOrMake(id string) interface{} {
	c.mu.RLock()
	if instance, ok := c.instances[id]; ok {
		c.mu.RUnlock()
		return instance
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if instance, ok := c.instances[id]; ok {
		return instance
	}

	if factory, ok := c.factories[id]; ok {
		instance := factory()
		c.instances[id] = instance
		return instance
	}

	if service, ok := c.services[id]; ok {
		c.instances[id] = service
		return service
	}

	return nil
}

func (c *Container) Inject(target interface{}) *Container {
	t := reflect.ValueOf(target)
	if t.Kind() != reflect.Ptr {
		return c
	}

	t = t.Elem()
	for i := 0; i < t.NumField(); i++ {
		field := t.Type().Field(i)
		fieldValue := t.Field(i)

		injectTag := field.Tag.Get("inject")
		if injectTag == "" {
			continue
		}

		serviceId := injectTag
		if serviceId == "" {
			serviceId = field.Name
		}

		if !fieldValue.CanSet() {
			continue
		}

		service := c.Get(serviceId)
		if service == nil {
			continue
		}

		fieldValue.Set(reflect.ValueOf(service))
	}

	return c
}

func (c *Container) Services() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	services := make([]string, 0, len(c.services))
	for id := range c.services {
		services = append(services, id)
	}
	return services
}

func (c *Container) Instances() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	instances := make([]string, 0, len(c.instances))
	for id := range c.instances {
		instances = append(instances, id)
	}
	return instances
}

func (c *Container) Remove(id string) *Container {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.services, id)
	delete(c.instances, id)

	for tag, serviceIds := range c.tags {
		newIds := make([]string, 0)
		for _, sid := range serviceIds {
			if sid != id {
				newIds = append(newIds, sid)
			}
		}
		c.tags[tag] = newIds
	}

	delete(c.factories, id)

	for alias, targetId := range c.aliases {
		if targetId == id {
			delete(c.aliases, alias)
		}
	}

	return c
}

func (c *Container) Clear() *Container {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = make(map[string]interface{})
	c.instances = make(map[string]interface{})
	c.aliases = make(map[string]string)
	c.tags = make(map[string][]string)
	c.factories = make(map[string]func() interface{})

	return c
}

func (c *Container) Wipe(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.services[id]; ok {
		delete(c.services, id)
	}

	if _, ok := c.instances[id]; ok {
		delete(c.instances, id)
	}

	return nil
}

func (c *Container) Do(tagName string, callback func(interface{}) error) error {
	c.mu.RLock()
	serviceIds := make([]string, len(c.tags[tagName]))
	copy(serviceIds, c.tags[tagName])
	c.mu.RUnlock()

	for _, id := range serviceIds {
		service := c.Get(id)
		if service == nil {
			return fmt.Errorf("service not found: %s", id)
		}
		if err := callback(service); err != nil {
			return err
		}
	}

	return nil
}
