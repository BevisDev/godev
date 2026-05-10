package framework

import "sync"

type Container struct {
	mu   sync.RWMutex
	deps map[string]interface{}
}

func NewContainer() *Container {
	return &Container{
		deps: make(map[string]interface{}),
	}
}

func (c *Container) Set(key string, val interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deps[key] = val
}

func (c *Container) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.deps[key]
}

func Get[T any](c *Container, key string) T {
	return c.Get(key).(T)
}
