package store

import (
	"sync"
)

// BaselineStore is a simple single-mutex store for benchmarking comparison
//
// This store uses ONE mutex for the entire map, and all operations serialise
// EXPECTED RESULT:
// At low concurrency: similar to sharded store
// At high concurrency (1000+ goroutines): sharded store wins by 10-30x
type BaselineStore struct {
	mu   sync.Mutex
	data map[string][]byte
}

func NewBaseline() *BaselineStore {
	bl := BaselineStore {
		data: make(map[string][]byte),
	}
	return &bl
}

// Get retrieves a value by key
// NOTE: Using Mutex, not RWMutex is deliberate, for worst case locking
func (s *BaselineStore) Get(key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[key]
	return v, ok
}

func (s *BaselineStore) Set(key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *BaselineStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}
