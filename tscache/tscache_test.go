package tscache

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"testing"
)

// TestGroup tests the functionality of the Group struct.
func TestGroup(t *testing.T) {
	// Define a mock database.
	var db = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}

	// Initialize a map to track the number of times each key is loaded.
	loadCounts := make(map[string]int, len(db))

	// Create a new cache group named "MainCache" with a maximum size of 1000 bytes.
	ts := NewGroup("MainCache", int64(1000), GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[Code data] search key", key)
			// Simulate fetching data from the mock database.
			if data, ok := db[key]; ok {
				// Increment the load count for the key.
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(data), nil
			}
			// Return an error if the data is not found in the database.
			return nil, fmt.Errorf("can't found data")
		}))

	// Test retrieving values from the cache.
	for k, v := range db {
		// Check if the retrieved value matches the expected value.
		if view, err := ts.Get(k); err != nil || view.String() != v {
			t.Fatal("Failed to get value")
		}
		// Check if the key was loaded only once.
		if _, err := ts.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatal("cache miss")
		}
	}

	// Repeat the tests to ensure cached values are retrieved without loading again.
	for k, v := range db {
		if view, err := ts.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value")
		}
		if _, err := ts.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatal("cache miss")
		}
	}

	// Test retrieving a value for a key that does not exist in the database.
	if view, err := ts.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}

// TestGroup_Get tests the Get method of the Group struct.
func TestGroup_Get(t *testing.T) {
	// Create a new cache group with a mock getter function.
	group := NewGroup("test-group", 100, GetterFunc(func(key string) ([]byte, error) {
		if key == "key1" {
			return []byte("value1"), nil
		}
		return nil, errors.New("not found")
	}))

	// Test case: key exists in the cache.
	byteView, err := group.Get("key1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectedValue := []byte("value1")
	if !bytes.Equal(byteView.ByteSlice(), expectedValue) {
		t.Errorf("expected value %v, got %v", expectedValue, byteView.ByteSlice())
	}

	// Test case: key does not exist in the cache.
	byteView, err = group.Get("key2")
	if err == nil || len(byteView.ByteSlice()) != 0 {
		t.Errorf("expected error, got nil, and non-empty byte view")
	}
}

// TestGroup_GetLocally tests the getLocally method of the Group struct.
func TestGroup_GetLocally(t *testing.T) {
	// Create a new cache group with a mock getter function.
	group := NewGroup("test-group", 100, GetterFunc(func(key string) ([]byte, error) {
		if key == "key1" {
			return []byte("value1"), nil
		}
		return nil, errors.New("not found")
	}))

	// Test case: key exists in the getter.
	byteView, err := group.getLocally("key1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectedValue := []byte("value1")
	if !bytes.Equal(byteView.ByteSlice(), expectedValue) {
		t.Errorf("expected value %v, got %v", expectedValue, byteView.ByteSlice())
	}

	// Test case: key does not exist in the getter.
	byteView, err = group.getLocally("key2")
	if err == nil || len(byteView.ByteSlice()) != 0 {
		t.Errorf("expected error, got nil, and non-empty byte view")
	}
}

// TestGroup_Load tests the load method of the Group struct.
func TestGroup_Load(t *testing.T) {
	// Create a new cache group with a mock getter function.
	group := NewGroup("test-group", 100, GetterFunc(func(key string) ([]byte, error) {
		if key == "key1" {
			return []byte("value1"), nil
		}
		return nil, errors.New("not found")
	}))

	// Test case: key exists in the cache.
	byteView, err := group.load("key1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectedValue := []byte("value1")
	if !bytes.Equal(byteView.ByteSlice(), expectedValue) {
		t.Errorf("expected value %v, got %v", expectedValue, byteView.ByteSlice())
	}

	// Test case: key does not exist in the cache or peers.
	byteView, err = group.load("key2")
	if err == nil {
		t.Errorf("expected error")
	}
	if len(byteView.ByteSlice()) != 0 {
		t.Errorf("expected empty byte view, got %v", byteView.ByteSlice())
	}
}
