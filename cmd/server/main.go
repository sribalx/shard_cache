package main

import (
	"log"

	"shard_cache/internal/metrics"
	"shard_cache/internal/server"
	"shard_cache/internal/store"
)

// Default port for cache server
const defaultAddr = ":9000"

// Main entry point
func main() {
	// TODO: signal handling for graceful shutdown
	log.Printf("starting server on %s", defaultAddr)
	st := store.New()
	m := &metrics.Metrics{}

	srv := server.New(defaultAddr, st, m)
	err := srv.Start()
	if err != nil {
		log.Fatal(err)
	}
}
