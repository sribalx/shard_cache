package server

import (
	"net"

	"shard_cache/internal/metrics"
	"shard_cache/internal/protocol"
	"shard_cache/internal/store"
)

// Job represents a unit of work for the worker pool
// 
// This decouples connection reading from request processing
// Connection handler reads frames and submits jobs, worker process them
type Job struct {
	Conn   net.Conn        // connection to write response to
	Frame  *protocol.Frame // parsed request frame
}

// WorkerPool manages a fixed number of worker goroutines
//
// A fixed pool provides:
// - Bounded concurrency (predictable resource usage)
// - Backpressure (when queue is full, callers block or get rejected)
type WorkerPool struct {
	jobs       chan Job          // Buffered channel for incoming work
	store      *store.Store       // Reference to KV store
	metrics    *metrics.Metrics   // Reference to metrics
	numWorkers int                // how many worker goroutines to spawn
}

// NewWorkerPool creates and starts a worker pool
func NewWorkerPool(numWorkers, queueSize int, st *store.Store, m *metrics.Metrics) *WorkerPool {
	wp := WorkerPool{
		jobs:       make(chan Job, queueSize),
		store:      st, 
		metrics:    m, 
		numWorkers: numWorkers}
	for i := 0; i < numWorkers; i++ {
		go wp.worker()
	}
	return &wp
}

// Submit adds a job to the queue
// Backpressure is non-blocking: use select with default to detect full queue
//
// RETURNS: true if job was queued, false if queue is full
func (wp *WorkerPool) Submit(job Job) bool {
	select {
	case wp.jobs <- job:
		return true
	default:
		return false // queue full
	}
}

// worker is the main loop for a single worker goroutine
//
// Note: `for job := range wp.jobs` loops until the channel is closed
//        When we close(wp.jobs) during shutdown, all workers exit cleanly
func (wp *WorkerPool) worker() {
	for job := range wp.jobs {
		wp.process(job)
	}
}

// process handles a single job, the actual GET/SET/DELETE logic

func (wp *WorkerPool) process(job Job) {

	var response *protocol.Frame
	switch job.Frame.Op {
	case protocol.OpGet:
		val, found := wp.store.Get(string(job.Frame.Key))
		if found {
				wp.metrics.RecordHit()
				response = &protocol.Frame{Op: protocol.OpValue, Value: val}
			} else {
				wp.metrics.RecordMiss()
				response = &protocol.Frame{Op:protocol.OpNotFound}
			}
	case protocol.OpSet:
		wp.store.Set(string(job.Frame.Key), job.Frame.Value)
		wp.metrics.RecordSet()
		response = &protocol.Frame{Op: protocol.OpOK}
	case protocol.OpDelete:
		wp.store.Delete(string(job.Frame.Key))
		wp.metrics.RecordDelete()
		response = &protocol.Frame{Op: protocol.OpOK}
	default: 
		response = &protocol.Frame{Op: protocol.OpError}
	}
	err := sendResponse(job.Conn, response)
	if err != nil {
		return
	}
	wp.metrics.RecordBytesOut(int64(len(response.Value)))
}

// Shutdown gracefully stops the worker pool
//
// During server shutdown, we want workers to finish current jobs before exiting, and
// closing the jobs channel signals workers to stop after draining the queue
func (wp *WorkerPool) Shutdown() {
	close(wp.jobs)
}

