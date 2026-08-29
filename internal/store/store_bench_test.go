package store

import (
	"fmt"
	"sync"
	"testing"
)

// NOTES FROM READING:
// Go's testing package has built-in benchmarking. Functions named Benchmark* receive 
// a *testing.B and run the code b.N times. The framework adjusts b.N to get stable measurements.
//
// Run benchmarks:
//   go test -bench=. ./internal/store/
//
// Run with memory stats:
//   go test -bench=. -benchmem ./internal/store/
//
// Run specific benchmark:
//   go test -bench=BenchmarkSharded -benchmem ./internal/store/

// BenchmarkBaselineStore_Concurrent measures the single-mutex store under load
//
// b.RunParallel simulates real-world concurrent access. Spawns GOMAXPROCS goroutines,
// each running the benchmark function in a loop
func BenchmarkBaselineStore_Concurrent(b *testing.B) {
	s := NewBaseline()

	// Pre-populate with data so Gets can hit
	for _, k := range benchKeys {
		s.Set(k, []byte("value"))
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {   // pb.Next() returns false when this goroutine's share of b.N is done
			key := benchKeys[i%1000]

			if i%10 == 0 {
				s.Set(key, []byte("value"))
			} else {
				s.Get(key)
			}
			i++
		}
	})
}

// BenchmarkShardedStore_Concurrent measures the sharded store under load
func BenchmarkShardedStore_Concurrent(b *testing.B) {
	s := New()

	// Pre-populate — same as baseline for fair comparison
	for _, k := range benchKeys {
		s.Set(k, []byte("value"))
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := benchKeys[i%1000]

			if i%10 == 0 {
				s.Set(key, []byte("value"))
			} else {
				s.Get(key)
			}
			i++
		}
	})
}

// BenchmarkComparison runs both stores at various concurrency levels
//
// Shows how performance scales with goroutine count
func BenchmarkComparison(b *testing.B) {
	concurrencyLevels := []int{1, 10, 100, 1000, 10000}

	for _, n := range concurrencyLevels {
		b.Run(fmt.Sprintf("Baseline/goroutines-%d", n), func(b *testing.B) {
			s := NewBaseline()

			// Pre-populate
			for _, k := range benchKeys {
				s.Set(k, []byte("value"))
			}

			var wg sync.WaitGroup // coordinate the different groups

			opsPerGoroutine := b.N / n
			if opsPerGoroutine < 1 {
				opsPerGoroutine = 1
			}
			b.ResetTimer()

			for i := 0; i < n; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < opsPerGoroutine; j++ {
						key := benchKeys[j%1000]
						if j%10 == 0 {
							s.Set(key, []byte("value"))
						} else {
							s.Get(key)
						}
					}
				}(i)
			}
			wg.Wait()
		})

		// Sharded store benchmark at this concurrency level
		b.Run(fmt.Sprintf("Sharded/goroutines-%d", n), func(b *testing.B) {
			s := New()

			// Pre-populate
			for _, k := range benchKeys {
				s.Set(k, []byte("value"))
			}

			var wg sync.WaitGroup
			opsPerGoroutine := b.N / n
			if opsPerGoroutine < 1 {
				opsPerGoroutine = 1
			}

			b.ResetTimer()

			for i := 0; i < n; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < opsPerGoroutine; j++ {
						key := benchKeys[j%1000]
						if j%10 == 0 {
							s.Set(key, []byte("value"))
						} else {
							s.Get(key)
						}
					}
				}(i)
			}
			wg.Wait()
		})
	}
}

// Helper, suppress unused import warning
var _ = sync.WaitGroup{}

// Helper, pre-computed keys to avoid fmt.Sprintf overhead in hot path
var benchKeys = func() []string {
	keys := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
	}
	return keys
}()