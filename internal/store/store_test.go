package store

import (
	"fmt"
	"sync"
	"testing"
)

// TestBasicOperations verifies Set/Get/Delete work correctly in single-threaded use.
func TestBasicOperations(t *testing.T) {
	// TODO: implement
	s := New()
	s.Set("key1", []byte("value1"))

	val, ok := s.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}
	if string(val) != "value1" {
		t.Errorf("got %q, expected %q", string(val), "value1")
	}

	_, ok = s.Get("key2")
	if ok {
		t.Fatal("expected key2 to not exist")
	}

	deleted := s.Delete("key1")
	if !deleted {
		t.Error("expected Delete to return true")
	}

	_, ok = s.Get("key1")
	if ok {
		t.Error("expected key1 to be gone after delete")
	}

	deleted = s.Delete("key1")
	if deleted {
		t.Error("expected Delete to return false for already-deleted key")
	}
}

// TestConcurrentAccess verifies the store is safe under concurrent access.
// RUN WITH go test -race ./internal/store/...
func TestConcurrentAccess(t *testing.T) {
	// TODO: implement
	// Hint: Use fmt.Sprintf("key-%d", id) for unique keys per goroutine
	s := New()
	var wg sync.WaitGroup

	numGoroutines := 100
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			key := fmt.Sprintf("key-%d", id)
			val := fmt.Appendf(nil, "val-%d", id)

			s.Set(key, val)

			got, ok := s.Get(key)
			if !ok {
				t.Errorf("goroutine %d: key missing", id)
			}
			if string(got) != string(val) {
				t.Errorf("goroutine %d: got %q, expected %q", id, got, val)
			}

			s.Delete(key)
		}(i)
	}
	wg.Wait()
}

// TestShardDistribution verifies keys are distributed across shards.
func TestShardDistribution(t *testing.T) {
	// TODO: implement
	hits := make(map[uint64]int)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		hash := fnv1a(key)
		shardIdx := hash & (NumShards - 1)
		hits[shardIdx]++
	}

	for idx, count := range hits {
		if count > 50 {
			t.Errorf("shard %d has %d hits, want <= 50", idx, count)
		}
	}

	if len(hits) < 50 {
		t.Errorf("only %d shards were hit, want >= 50", len(hits))
	}
}

// TestOverwriteValue verifies Set overwrites existing values.
func TestOverwriteValue(t *testing.T) {
	// TODO: implement
	s := New()
	s.Set("key", []byte("first"))
	s.Set("key", []byte("second"))
	val, ok := s.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if string(val) != "second" {
		t.Errorf("got %q, expected %q", string(val), "second")
	}
}
