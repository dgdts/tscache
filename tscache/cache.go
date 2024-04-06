package tscache

import (
	"sync"
	"tscache/lru"
)

// cache is a synchronized cache structure.
type cache struct {
	mu         sync.Mutex // Mutex for synchronization
	lru        *lru.Cache // LRU cache instance
	cacheBytes int64      // Maximum cache size in bytes
}

// add adds a key-value pair to the cache.
// It initializes the LRU cache if it's nil.
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.NewCache(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

// get retrieves the value associated with the given key from the cache.
// It returns the value and a boolean indicating whether the key was found.
func (c *cache) get(key string) (ByteView, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return ByteView{}, false
	}

	if ret, ok := c.lru.Get(key); ok {
		return ret.(ByteView), true
	}
	return ByteView{}, false
}
