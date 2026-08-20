// Package store implements the sharded in-memory key-value storage engine
package store

import (
	"sync"
)

// shard is a single partition in the key-value store
// Each shard has its own RWMutex, allowing concurrent reads but exclusive writes, since caches are read-heavy
type shard struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// newShard returns an initialised shard with empty map (cannot be nil map)
func newShard() *shard {
	return &shard{data: make(map[string][]byte)}
}

// Get retrieves a value by key
// Uses RLock, multiple reads allowed, and only blocks if a writer holds a lock
func (s *shard) Get(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

// Set stores a KV pair
// Uses Lock because writing to the map requires exclusive access
func (s *shard) Set(key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Delete removes a key from the shard
// Returns true if key existed and was deleted, false if it doesn't exist
func (s *shard) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.data[key]
	delete(s.data, key)
	return exists
}
