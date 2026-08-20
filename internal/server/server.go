// Package server implements the TCP server that accepts connections and dispatches to handlers
// Entry point for network traffic, listening on TCP port and creates goroutine per connection
package server

import (
	"net"
	"shard_cache/internal/metrics"
	"shard_cache/internal/store"
)

// Server manages the TCP listener and connection lifecycle
// TODO: references to worker pool
type Server struct {
	addr     string
	listener net.Listener
	store    *store.Store
	metrics  *metrics.Metrics
}

// New creates a new Server that listens to the given address
// Constructor encapsulates initialisation; callers do not need to know which fields exist
func New(addr string, st *store.Store, m *metrics.Metrics) *Server {
	return &Server{
		addr:    addr,
		store:   st,
		metrics: m,
	}
}

// Start begins listening for TCP connections
// Start is blocking; it runs the accept loop forever until error. Caller will call this in the main goroutine.
// TODO: graceful shutdown, worker pool for backpressure
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

// Note: handleConn processes a single client connection
// Separate method for the per-connection logic and lifecycle, since each connection is its own goroutine
