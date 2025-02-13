package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	cache map[string]cacheEntry
	mux   *sync.RWMutex
}

type cacheEntry struct {
	createAt time.Time
	val      []byte
}

func NewCache(interval time.Duration) *Cache {
	cache := &Cache{
		cache: make(map[string]cacheEntry),
		mux:   &sync.RWMutex{},
	}
	go cache.reapLoop(interval)
	return cache
}

func (c *Cache) Add(key string, val []byte) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.cache[key] = cacheEntry{
		createAt: time.Now(),
		val:      val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	entry, ok := c.cache[key]
	if !ok {
		return nil, false
	}
	return entry.val, true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		c.reap(interval)
	}
}

func (c *Cache) reap(interval time.Duration) {
	timeAgo := time.Now().Add(-interval)
	var keysToDelete []string

	c.mux.RLock()
	for k, v := range c.cache {
		if v.createAt.Before(timeAgo) {
			keysToDelete = append(keysToDelete, k)
		}
	}
	c.mux.RUnlock()

	if len(keysToDelete) > 0 {
		c.mux.Lock()
		defer c.mux.Unlock()
		for _, k := range keysToDelete {
			delete(c.cache, k)
		}
	}
}