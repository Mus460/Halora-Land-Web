package cache

import (
	"sync"
	"time"
)

type entry struct {
	value  any
	expiry time.Time
}

// Cache is a minimal thread-safe in-memory TTL cache used for
// small, mostly-static reference data (e.g. AHSP master data).
type Cache struct {
	mu    sync.Mutex
	ttl   time.Duration
	items map[string]entry
}

func New(ttl time.Duration) *Cache {
	return &Cache{ttl: ttl, items: make(map[string]entry)}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.items[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expiry) {
		delete(c.items, key)
		return nil, false
	}
	return e.value, true
}

func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry{value: value, expiry: time.Now().Add(c.ttl)}
}

// Clear drops every entry; used to invalidate cached lists on mutations.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	clear(c.items)
}
