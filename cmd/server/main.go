package main

import (
		"log"

		"shard_cache/internal/server"
)

// Default port for cache server
const defaultAddr = ":9000"

// Main entry point 
func main() {
	// TODO: signal handling for graceful shutdown
	log.Printf("starting server on %s", defaultAddr)
	srv := server.New(defaultAddr)
	err := srv.Start()
	if err != nil {
		log.Fatal(err)
	}
}
