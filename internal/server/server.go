// Package server implements the TCP server that accepts connections and dispatches to handlers
// Entry point for network traffic, listening on TCP port and creates goroutine per connection
package server

import (
	"net"
	"shard_cache/internal/metrics"
	"shard_cache/internal/store"
)

// Server manages the TCP listener and connection lifecycle
type Server struct {
	addr     string
	listener net.Listener
	store    *store.Store
	metrics  *metrics.Metrics
	workers  *WorkerPool 
}

// New creates a new Server that listens to the given address
// Constructor encapsulates initialisation; callers do not need to know which fields exist
func New(addr string, st *store.Store, m *metrics.Metrics, numWorkers, queueSize int) *Server {
	return &Server{
		addr:    addr,
		store:   st,
		metrics: m,
		workers: NewWorkerPool(numWorkers, queueSize, st, m),
	}
}

// Start begins listening for TCP connections
// Start is blocking; it runs the accept loop forever until error 
// Caller will call this in the main goroutine
//
// Note: handleConn processes a single client connection
// Separate method for the per-connection logic and lifecycle, 
// since each connection is its own goroutine
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		go s.handleConn(conn)
	}
}



// Shutdown gracefully stops the server
//
// Abrupt termination (Ctrl+C) leaves connections hanging and work incomplete
// Graceful shutdown: stop accepting new connections, let existing work finish,
// then exit cleanly
func (s *Server) Shutdown() {
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			return
		}
	}
	s.workers.Shutdown()
}
