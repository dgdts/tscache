package consistenthash

import (
	"hash/crc32"
	"reflect"
	"testing"
)

// TestNewMap tests the creation of a new consistent hash map.
func TestNewMap(t *testing.T) {
	// Create a new consistent hash map with 3 replicas and default hash function
	m := NewMap(3, nil)

	// Check if the hash function is set to default crc32.ChecksumIEEE
	if m.hash == nil || reflect.ValueOf(m.hash).Pointer() != reflect.ValueOf(crc32.ChecksumIEEE).Pointer() {
		t.Errorf("Expected default hash function to be set")
	}

	// Check if the number of replicas is set correctly
	if m.replicas != 3 {
		t.Errorf("Expected replicas to be set to 3, got %d", m.replicas)
	}

	// Check if the hashMap is initialized
	if m.hashMap == nil {
		t.Errorf("Expected hashMap to be initialized")
	}
}

// TestSelectNode tests selecting a node for a given key from the consistent hash map.
func TestSelectNode(t *testing.T) {
	// Create a new consistent hash map with 2 replicas and default hash function
	m := NewMap(2, nil)

	// Create nodes
	node1 := &Node{Name: "node1"}
	node2 := &Node{Name: "node2"}

	// Add nodes to the hash map
	m.Add(node1, node2)

	// Select nodes for keys
	selectedNode1, _ := m.SelectNode("key1")
	selectedNode2, _ := m.SelectNode("key2")

	// Check if the correct nodes are selected for keys
	if selectedNode1 != node1 {
		t.Errorf("Expected node1 to be selected for key1, got %v", selectedNode1)
	}
	if selectedNode2 != node2 {
		t.Errorf("Expected node2 to be selected for key2, got %v", selectedNode2)
	}
}

// TestSelectNodeEmptyMap tests selecting a node when the consistent hash map is empty.
func TestSelectNodeEmptyMap(t *testing.T) {
	// Create a new consistent hash map with 2 replicas and default hash function
	m := NewMap(2, nil)

	// Select node for a key
	_, err := m.SelectNode("key")

	// Check if an error is returned when the map is empty
	if err == nil || err.Error() != "no nodes available" {
		t.Errorf("Expected error 'no nodes available', got %v", err)
	}
}

// TestGetNode tests the SelectNode method of the Map struct.
func TestGetNode(t *testing.T) {
	// Create a new Map instance with size 1 and a hash function converting bytes to numbers
	m := NewMap(1, func(data []byte) uint32 {
		return uint32(data[0] - '0')
	})

	// Add nodes to the map
	m.Add(&Node{"2"}, &Node{"4"}, &Node{"8"})

	// Test node selection
	node1, _ := m.SelectNode("2")
	node2, _ := m.SelectNode("3")
	node3, _ := m.SelectNode("5")

	// Check if nodes are returned as expected
	if node1.Name != "2" || node2.Name != "4" || node3.Name != "8" {
		t.Fatalf("Test failed, %s, %s, %s", node1.Name, node2.Name, node3.Name)
	}

	// Add a new node to the map
	m.Add(&Node{"6"})

	// Re-select node "5" and check if it has been updated to node "6"
	node3, _ = m.SelectNode("5")

	// Check if the node is returned as expected
	if node3.Name != "6" {
		t.Fatalf("Test failed, %s", node3.Name)
	}
}
