package tscache

import (
	"fmt"
	"log"
	"sync"
	"syscall"
	"tscache/lru"
	pb "tscache/tscachepb"
)

// Getter is an interface for getting data based on a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc is an adapter function that allows using ordinary functions as Getter interfaces.
type GetterFunc func(key string) ([]byte, error)

// Get calls the GetterFunc function itself.
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group represents a cache group that encapsulates a cache and its associated peers.
type Group struct {
	name      string     // name is the name of the cache group.
	getter    Getter     // getter is the callback function to fetch data if it's not in the cache.
	mainCache cache      // mainCache is the main LRU cache.
	peers     PeerPicker // peers is the peer picker for selecting remote peers.

	mu sync.Mutex       // mu is used for synchronizing access to the Group.
	m  map[string]*call // m maps each key to its corresponding call.
}

// call represents an in-flight or completed call to Get.
type call struct {
	wg  sync.WaitGroup // wg is used to wait for the completion of the call.
	val interface{}    // val is the value returned by the call.
	err error          // err is the error returned by the call.
}

var (
	mu     sync.RWMutex              // mu guards the groups map.
	groups = make(map[string]*Group) // groups maps cache group names to their corresponding Group instances.
)

// NewGroup creates and returns a new cache Group with the specified name, cache size, and getter function.
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}

	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			lru:        lru.NewCache(cacheBytes, nil),
			cacheBytes: cacheBytes,
		},
	}
	groups[name] = g
	return g
}

// GetGroup returns the cache Group associated with the given name.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Do executes the function fn if the key is not in the cache.
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := &call{}
	g.m[key] = c
	c.wg.Add(1)
	g.mu.Unlock()

	c.val, c.err = fn()

	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}

// Get retrieves the value for a given key from the cache.
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is empty")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Printf("[pid:%d][%s Group] Cache hit", syscall.Getpid(), g.name)
		return v, nil
	}

	return g.load(key)
}

// RegisterNodes registers the peer picker for selecting remote peers.
func (g *Group) RegisterNodes(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// load loads the value for a key either from a peer or locally.
func (g *Group) load(key string) (value ByteView, err error) {
	data, err := g.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[TSCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	return data.(ByteView), err
}

// getFromPeer fetches the value for a key from a remote peer.
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	request := &pb.Request{Group: g.name, Key: key}
	response := &pb.Response{}
	err := peer.Get(request, response)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{B: response.GetValue()}, nil
}

// getLocally fetches the value for a key from the local cache or getter.
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{B: cloneBytes(bytes)}
	g.mainCache.add(key, value)
	return value, nil
}
