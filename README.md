# ShardCache

A concurrent in-memory key-value cache server in Go. Built as a learning project.

---

## Why I Built This

I was exposed to Go heavily in the past year, but I found myself writing Go without really understanding what makes it Go. I didn't have a strong command of the Go primitives, and lacked the mental model to appreciate the concurrency strengths and abstractions Go presents. 

More broadly, concurrency was a gap in my CS knowledge. I'd learnt about mutexes and channels in OS class, but I'd never built something where the choice between a mutex and a channel actually mattered. I wanted to get my hands dirty with a few things:

- Fine-grained locking (and why 64)
- Memory pooling (why `sync.Pool` exists)
- Binary protocols for speed
- The tradeoffs that real systems like Redis and Memcached make

This definitely isn't a Redis replacement, because it currently lacks the distributed systems strengths. But is does give me a good understanding of why Redis made the choices it did. 

---

## Architecture

```mermaid
flowchart TB
    clients@{ shape: processes, label: "TCP Clients" }
    tcp_server@{ shape: lean-r, label: "TCP Server\n(:9000)" }
    conn_handler@{ shape: rounded, label: "Connection Handler\n(1 goroutine per conn)" }
    frame_parser@{ shape: rounded, label: "Frame Parser\n(binary protocol)" }
    job_channel@{ shape: das, label: "Job Channel\n(buffered)" }
    worker_pool@{ shape: rounded, label: "Worker Pool\n(NumCPU workers)" }
    
    subgraph store["Sharded Store (64 shards)"]
        direction LR
        shard0@{ shape: cylinder, label: "Shard 0\nRWMutex + map" }
        shard1@{ shape: cylinder, label: "Shard 1\nRWMutex + map" }
        shardN@{ shape: cylinder, label: "...\nShard 63" }
    end
    
    hash_fn@{ shape: diamond, label: "FNV-1a\nhash" }
    
    subgraph aux["Support Components"]
        direction TB
        buffer_pool@{ shape: rounded, label: "Buffer Pool\n(sync.Pool)" }
        metrics@{ shape: rounded, label: "Metrics\n(atomics)" }
    end

    %% Request flow
    clients <-- "Binary Frames" --> tcp_server
    tcp_server -- "Accept" --> conn_handler
    conn_handler -- "Read Header" --> frame_parser
    frame_parser -- "Submit Job" --> job_channel
    job_channel -- "Dequeue" --> worker_pool
    worker_pool -- "Key" --> hash_fn
    hash_fn -- "shard[hash & 63]" --> store
    
    %% Response flow
    worker_pool -- "Response" --> conn_handler
    
    %% Auxiliary
    conn_handler -. "Get/Put" .-> buffer_pool
    worker_pool -. "Record" .-> metrics

    classDef core fill:#00574b,stroke-width:2px,color:#fff;
    worker_pool:::core
    
    classDef storage fill:#1565c0,stroke-width:1px,color:#fff;
    shard0:::storage
    shard1:::storage
    shardN:::storage
    
    classDef auxiliary fill:#00897b,stroke-width:1px,color:#fff;
    buffer_pool:::auxiliary
    metrics:::auxiliary
```

---

## Build and Run

```bash
# Start the server (listens on :9000, pprof on :6060)
go run ./cmd/server

# In another terminal, use the client
go run ./cmd/client set foo bar    # OK
go run ./cmd/client get foo        # Value: bar
go run ./cmd/client del foo        # OK
go run ./cmd/client get foo        # NOT FOUND
```

---

## Run Tests

```bash
go test ./...                      # All tests
go test -race ./...                # Race detector
go test -bench=. ./internal/store  # Benchmarks
```

Fuzz testing (let it run for 30+ seconds):

```bash
go test -fuzz=FuzzDecodeHeader -fuzztime=30s ./internal/protocol/
go test -fuzz=FuzzDecodePayload -fuzztime=30s ./internal/protocol/
go test -fuzz=FuzzRoundTrip -fuzztime=30s ./internal/protocol/
```

---

## Binary Protocol

```
┌───────────┬───────────┬───────────┬───────────┬───────────────────┐
│ Start (1B)│ OpCode(1B)│ KeyLen(2B)│ ValLen(4B)│ Payload (K+V)     │
└───────────┴───────────┴───────────┴───────────┴───────────────────┘
     0xCA      0x01-03      big-endian   big-endian   key || value
```

Request opcodes: GET (0x01), SET (0x02), DELETE (0x03)

Response opcodes: OK (0x80), ERROR (0x81), VALUE (0x82), NOT_FOUND (0x83)

---

## Benchmark Results

Compared a single-mutex baseline against the 64-shard implementation:

```
BenchmarkComparison/Baseline/goroutines-1       27.48 ns/op
BenchmarkComparison/Sharded/goroutines-1        39.03 ns/op   (sharding overhead)

BenchmarkComparison/Baseline/goroutines-100     129.7 ns/op
BenchmarkComparison/Sharded/goroutines-100      31.12 ns/op   ~4x faster

BenchmarkComparison/Baseline/goroutines-1000    160.8 ns/op
BenchmarkComparison/Sharded/goroutines-1000     32.18 ns/op   ~5x faster

BenchmarkComparison/Baseline/goroutines-10000   149.5 ns/op
BenchmarkComparison/Sharded/goroutines-10000    31.11 ns/op   ~5x faster
```

**Result: ~5x throughput improvement under contention.**

The improvement plateaus at 5x because I'm running on 8 cores. Even with 64 shards, only 8 goroutines are truly running at once — so at most 8 shards are being contended simultaneously. The other 56 shards are just chilling.

**TODO:** Deploy to a larger EC2 instance (32+ cores) to test whether the improvement scales further. Hypothesis: with more cores, more shards see real contention, and the gap should widen.

---

## What I Learned

**TO DO**

---

## File Structure

```
├── cmd/
│   ├── server/main.go          # Entry point, signal handling
│   └── client/main.go          # CLI client for testing
│
├── internal/
│   ├── protocol/
│   │   ├── opcodes.go          # Constants
│   │   ├── frame.go            # Encode/Decode
│   │   ├── frame_test.go       # Unit tests
│   │   └── fuzz_test.go        # Fuzz tests
│   │
│   ├── server/
│   │   ├── server.go           # TCP listener
│   │   ├── conn.go             # Connection handler
│   │   └── worker.go           # Worker pool
│   │
│   ├── store/
│   │   ├── shard.go            # Single shard (RWMutex + map)
│   │   ├── store.go            # 64-shard store with FNV-1a routing
│   │   ├── store_baseline.go   # Single-mutex baseline for benchmarks
│   │   ├── store_test.go       # Unit tests
│   │   └── store_bench_test.go # Benchmarks
│   │
│   ├── pool/
│   │   └── buffer.go           # sync.Pool for byte buffers
│   │
│   └── metrics/
│       └── metrics.go          # Atomic counters
│
├── go.mod
└── README.md
```

---

## What I'd Do Differently

If I were building a production cache, I'd probably start single-threaded like Redis and only add threading for network I/O if profiling showed it was needed. Sharding the KV layer was a great learning exercise, but the insight is that it solves a problem that doesn't dominate in this workload.

Also, I'd add TTL support. A cache without expiration is just a memory leak with extra steps.
