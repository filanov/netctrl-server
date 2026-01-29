package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/filanov/netctrl-server/internal/config"
	"github.com/filanov/netctrl-server/internal/server"
	"github.com/filanov/netctrl-server/internal/storage/memory"
)

func main() {
	// Load configuration
	cfg, err := config.LoadOrDefault("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting netctrl-server in %s mode", cfg.Server.Environment)

	// Initialize storage
	var storage = memory.New()

	// Create server
	srv := server.New(cfg, storage)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		srv.Stop()
	case err := <-errChan:
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server shutdown complete")
}
