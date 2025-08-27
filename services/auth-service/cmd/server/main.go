package main

import (
	"context"
	"log"
	"time"

	"auth-service/internal/transport/server"

	_ "github.com/lib/pq"
)

func main() {
	// Create root context with timeout for initialization
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize unified server (gRPC + REST)
	srv, err := server.NewServer(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	// Run the server
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
