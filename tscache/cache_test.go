package tscache

import (
	"sync"
	"testing"
)

func TestCache_AddAndGet(t *testing.T) {
	// Initialize cache
	c := &cache{
		cacheBytes: 1000, // Set maximum cache size
	}

	// Add items to cache
	c.add("key1", ByteView{B: []byte("value1")})
	c.add("key2", ByteView{B: []byte("value2")})

	// Retrieve items from cache
	value1, found1 := c.get("key1")
	value2, found2 := c.get("key2")

	// Verify retrieval
	if !found1 || value1.String() != "value1" {
		t.Errorf("Expected value1: value1, got: %s", value1)
	}
	if !found2 || value2.String() != "value2" {
		t.Errorf("Expected value2: value2, got: %s", value2)
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	// Initialize cache
	c := &cache{
		cacheBytes: 1000, // Set maximum cache size
	}

	// Add items to cache
	c.add("key1", ByteView{B: []byte("value1")})

	// Create wait group for concurrent access
	var wg sync.WaitGroup
	wg.Add(2)

	// Concurrent access to cache
	go func() {
		defer wg.Done()
		// Retrieve item from cache
		value, found := c.get("key1")
		// Verify retrieval
		if !found || value.String() != "value1" {
			t.Errorf("Expected value: value1, got: %s", value)
		}
	}()

	go func() {
		defer wg.Done()
		// Add item to cache
		c.add("key2", ByteView{B: []byte("value2")})
	}()

	// Wait for goroutines to finish
	wg.Wait()
}
