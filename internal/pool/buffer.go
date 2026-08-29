// Package pool provides reusable byte buffers to reduce allocations
// 
// Every make([]byte, n) allocated heap memory and at high throughput, this constant allocation
// can cause frequent GC pauses, so sync.Pool allows us to reuse buffers
package pool

import (
	"sync"
)


// DefaultBufferSize is the size of pooled buffers
// Common page size, good for payloads. If a payload exceeds, can allocate a larger buffer but don't return to pool
const DefaultBufferSize = 4096

// bufferPool is the global pool for byte buffers
//
// sync.Pool:
// - Thread safe without explicit locking
// - Automatically clears unused objects during GC
// - per processor caching reduces contention
var bufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, DefaultBufferSize)
		return &buf
	},
}

// GetBuffer retrieves a buffer from the pool
// Returns a byte slice of at least DefaultBufferSize capacity, with full capacity
// Note: Caller should zero or overwrite completely before using, might have stale data.
func GetBuffer() []byte {
	return *bufferPool.Get().(*[]byte)
}

// PutBuffer returns a buffer to the pool for reuse
//
// Usage:
// - Only return buffers of Default BufferSize capacity
// - Don't return oversized buffers
// - After put, make sure caller is not using the buffer at all anymore. 
func PutBuffer(buf []byte) {
	if cap(buf) != DefaultBufferSize{
		return
	}
	buf = buf[:cap(buf)]
	bufferPool.Put(&buf)
}