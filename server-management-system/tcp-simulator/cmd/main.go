package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/vcs-sms/tcp-simulator/simulator"
)

func main() {
	// Load configuration
	cfg := simulator.LoadConfig()

	// Create simulator manager
	manager := simulator.NewManager(cfg)

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start control loop in a goroutine
	go manager.RunControlLoop()

	// Wait for shutdown signal
	<-quit

	// Shutdown
	manager.Shutdown()
}
