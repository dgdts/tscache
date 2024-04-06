package consistenthash

import (
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash represents a function that generates a hash value for given data.
type Hash func(data []byte) uint32

// Node represents a node in the consistent hash ring.
type Node struct {
	Name string // Name of the node
}

// Map represents the consistent hash map.
type Map struct {
	hash     Hash          // Hash function to use
	replicas int           // Number of replicas of each node in the hash ring
	keys     []int         // Sorted list of hash keys
	hashMap  map[int]*Node // Mapping of hash keys to nodes
}

// NewMap creates and initializes a new consistent hash map.
// If hash function is not provided, it defaults to crc32.ChecksumIEEE.
func NewMap(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]*Node),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds nodes to the consistent hash map.
func (m *Map) Add(nodes ...*Node) {
	for _, node := range nodes {
		for i := 0; i < m.replicas; i++ {
			// Compute hash for the node with replica index
			hash := int(m.hash([]byte(node.Name + strconv.Itoa(i))))
			// Append hash to the list of keys
			m.keys = append(m.keys, hash)
			// Map the hash to the node
			m.hashMap[hash] = node
		}
	}
	// Sort the list of keys for binary search
	sort.Ints(m.keys)
}

// SelectNode selects the node responsible for a given key.
func (m *Map) SelectNode(key string) (*Node, error) {
	if len(m.keys) == 0 {
		return nil, errors.New("no nodes available")
	}

	// Compute hash for the key
	hash := int(m.hash([]byte(key)))

	// Find the index of the first key greater than or equal to the hash
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// Get the node mapped to the selected key's hash
	return m.hashMap[m.keys[idx%len(m.keys)]], nil
}
