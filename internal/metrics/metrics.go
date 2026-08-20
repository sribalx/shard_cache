// Package metrics provides lock-free telemetry using atomic operations
// Why ATOMICS: metrics are updated on every request, and using mutex will add lock contention
// Atomic operations are lock-free, no blocking, and only nanosecond overhead
package metrics

import (
	"sync/atomic"
)

// Metircs holds operational counters for the KV cache across shards
type Metrics struct {
	Hits     int64 // successful Get (key existed)
	Misses   int64 // failed Get (key not found)
	Sets     int64 // Set operations
	Deletes  int64 // Delete operations
	BytesIn  int64 // total bytes received into cache (keys + values)
	BytesOut int64 // total bytes sent from cache (only value)
}

func (m *Metrics) RecordHit() {
	atomic.AddInt64(&m.Hits, 1)
}

func (m *Metrics) RecordMiss() {
	atomic.AddInt64(&m.Misses, 1)
}

func (m *Metrics) RecordSet() {
	atomic.AddInt64(&m.Sets, 1)
}

func (m *Metrics) RecordDelete() {
	atomic.AddInt64(&m.Deletes, 1)
}

func (m *Metrics) RecordBytesIn(n int64) {
	atomic.AddInt64(&m.BytesIn, n)
}

func (m *Metrics) RecordBytesOut(n int64) {
	atomic.AddInt64(&m.BytesOut, n)
}

// Snapshot returns a point-in-time copy of all metrics
// Reading multiple fields is not atomic as a group, but Snapsjot uses atomic loads for each field
// Ensures each read is consistent at a moment in time
func (m *Metrics) Snapshot() Metrics {
	return Metrics{
		Hits:     atomic.LoadInt64(&m.Hits),
		Misses:   atomic.LoadInt64(&m.Misses),
		Sets:     atomic.LoadInt64(&m.Sets),
		Deletes:  atomic.LoadInt64(&m.Deletes),
		BytesIn:  atomic.LoadInt64(&m.BytesIn),
		BytesOut: atomic.LoadInt64(&m.BytesOut),
	}
}
