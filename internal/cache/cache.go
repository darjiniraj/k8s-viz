package cache

import (
	"sync"
	"time"
)

type Item struct {
	Value      interface{}
	Expiration int64
}

type MemoryCache struct {
	sync.RWMutex
	items map[string]Item
}

func New() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]Item),
	}
}

func (c *MemoryCache) Set(key string, value interface{}, duration time.Duration) {
	c.Lock()
	defer c.Unlock()
	c.items[key] = Item{
		Value:      value,
		Expiration: time.Now().Add(duration).UnixNano(),
	}
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()
	item, found := c.items[key]
	if !found || time.Now().UnixNano() > item.Expiration {
		return nil, false
	}
	return item.Value, true
}

func (c *MemoryCache) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.items, key)
}