package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"shard_cache/internal/metrics"
	"shard_cache/internal/server"
	"shard_cache/internal/store"
)

// Default port for cache server
const defaultAddr = ":9000"

// Main entry point
func main() {
	log.Printf("starting server on %s", defaultAddr)
	
	st := store.New()
	m := &metrics.Metrics{}

	numWorkers := runtime.NumCPU()
	queueSize := 1000

	srv := server.New(defaultAddr, st, m, numWorkers, queueSize)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		err := srv.Start()
		if err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()
	<-sigCh
	log.Printf("shutting down server on %s...", defaultAddr)
	srv.Shutdown()
	log.Print("shutdown complete")
}
