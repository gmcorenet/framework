package cache

import (
	"sync"
	"time"
)

type CacheItemInterface interface {
	GetKey() string
	Get() interface{}
	IsHit() bool
	Set(value interface{}) *CacheItem
	ExpiresAt(t time.Time) *CacheItem
	ExpiresAfter(d time.Duration) *CacheItem
	IsExpired() bool
}

type CacheItem struct {
	key        string
	value      interface{}
	expiration time.Time
	hit        bool
}

func NewCacheItem(key string, value interface{}) *CacheItem {
	return &CacheItem{
		key:        key,
		value:      value,
		expiration: time.Time{},
	}
}

func (c *CacheItem) GetKey() string {
	return c.key
}

func (c *CacheItem) Get() interface{} {
	return c.value
}

func (c *CacheItem) IsHit() bool {
	return c.hit
}

func (c *CacheItem) Set(value interface{}) *CacheItem {
	c.value = value
	return c
}

func (c *CacheItem) ExpiresAt(t time.Time) *CacheItem {
	c.expiration = t
	return c
}

func (c *CacheItem) ExpiresAfter(d time.Duration) *CacheItem {
	c.expiration = time.Now().Add(d)
	return c
}

func (c *CacheItem) IsExpired() bool {
	if c.expiration.IsZero() {
		return false
	}
	return time.Now().After(c.expiration)
}

type CachePoolInterface interface {
	GetItem(key string) CacheItemInterface
	GetItems(keys []string) map[string]CacheItemInterface
	HasItem(key string) bool
	Clear() bool
	DeleteItem(key string) bool
	DeleteItems(keys []string) bool
	Save(item CacheItemInterface) bool
	SaveDeferred(item CacheItemInterface) bool
	Commit() bool
}

type ArrayCache struct {
	items map[string]*CacheItem
	mu    sync.RWMutex
}

func NewArrayCache() *ArrayCache {
	return &ArrayCache{
		items: make(map[string]*CacheItem),
	}
}

func (c *ArrayCache) GetItem(key string) CacheItemInterface {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if !ok || item.IsExpired() {
		newItem := NewCacheItem(key, nil)
		newItem.hit = false
		return newItem
	}

	item.hit = true
	return item
}

func (c *ArrayCache) GetItems(keys []string) map[string]CacheItemInterface {
	result := make(map[string]CacheItemInterface)
	for _, key := range keys {
		result[key] = c.GetItem(key)
	}
	return result
}

func (c *ArrayCache) HasItem(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return false
	}
	return !item.IsExpired()
}

func (c *ArrayCache) Clear() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*CacheItem)
	return true
}

func (c *ArrayCache) DeleteItem(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	return true
}

func (c *ArrayCache) DeleteItems(keys []string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, key := range keys {
		delete(c.items, key)
	}
	return true
}

func (c *ArrayCache) Save(item CacheItemInterface) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[item.GetKey()] = item.(*CacheItem)
	return true
}

func (c *ArrayCache) SaveDeferred(item CacheItemInterface) bool {
	return c.Save(item)
}

func (c *ArrayCache) Commit() bool {
	return true
}

type ChainCache struct {
	pools []CachePoolInterface
}

func NewChainCache(pools ...CachePoolInterface) *ChainCache {
	return &ChainCache{pools: pools}
}

func (c *ChainCache) GetItem(key string) CacheItemInterface {
	for _, pool := range c.pools {
		item := pool.GetItem(key)
		if item.IsHit() {
			return item
		}
	}
	return NewCacheItem(key, nil)
}

func (c *ChainCache) GetItems(keys []string) map[string]CacheItemInterface {
	result := make(map[string]CacheItemInterface)
	for _, key := range keys {
		result[key] = c.GetItem(key)
	}
	return result
}

func (c *ChainCache) HasItem(key string) bool {
	for _, pool := range c.pools {
		if pool.HasItem(key) {
			return true
		}
	}
	return false
}

func (c *ChainCache) Clear() bool {
	for _, pool := range c.pools {
		pool.Clear()
	}
	return true
}

func (c *ChainCache) DeleteItem(key string) bool {
	for _, pool := range c.pools {
		pool.DeleteItem(key)
	}
	return true
}

func (c *ChainCache) DeleteItems(keys []string) bool {
	for _, pool := range c.pools {
		pool.DeleteItems(keys)
	}
	return true
}

func (c *ChainCache) Save(item CacheItemInterface) bool {
	for _, pool := range c.pools {
		pool.Save(item)
	}
	return true
}

func (c *ChainCache) SaveDeferred(item CacheItemInterface) bool {
	return c.Save(item)
}

func (c *ChainCache) Commit() bool {
	for _, pool := range c.pools {
		pool.Commit()
	}
	return true
}

type PoolManager struct {
	pools map[string]CachePoolInterface
	mu    sync.RWMutex
}

func NewPoolManager() *PoolManager {
	return &PoolManager{
		pools: make(map[string]CachePoolInterface),
	}
}

func (m *PoolManager) Register(name string, pool CachePoolInterface) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pools[name] = pool
}

func (m *PoolManager) Get(name string) CachePoolInterface {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pools[name]
}

func (m *PoolManager) All() map[string]CachePoolInterface {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]CachePoolInterface, len(m.pools))
	for k, v := range m.pools {
		result[k] = v
	}
	return result
}
