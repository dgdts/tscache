package lru

import "container/list"

type Cache struct {
	maxBytes  int64
	nbytes    int64
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

func NewCache(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		nbytes:    0,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if data, ok := c.cache[key]; ok {
		c.ll.MoveToFront(data)
		kv := data.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

func (c *Cache) RemoveOldst() {
	data := c.ll.Back()
	userData := data.Value.(*entry)
	if userData != nil {
		delete(c.cache, userData.key)
		c.ll.Remove(data)
		c.nbytes = c.nbytes - int64(len(userData.key)) - int64(userData.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(userData.key, userData.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if data, ok := c.cache[key]; ok {
		userData := data.Value.(*entry)
		oldSize := userData.value.Len()
		c.ll.MoveToFront(data)
		c.nbytes = c.nbytes - int64(oldSize) + int64(value.Len())
	} else {
		newData := c.ll.PushFront(&entry{
			key:   key,
			value: value,
		})
		c.cache[key] = newData
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldst()
	}
}

func (c *Cache) Len() int {
	return len(c.cache)
}
