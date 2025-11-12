// Package main implements Radio Control Container entry point.
// Architecture: IEEE 42010 + arc42 structure per docs/radio_control_container_architecture_v1.md
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/radio-control/rcc/internal/api"
	"github.com/radio-control/rcc/internal/audit"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
)

const (
	// Service configuration
	DefaultPort = "8000"
	DefaultAddr = ":" + DefaultPort
	Version     = "1.0.0"
)

func main() {
	log.Printf("Starting Radio Control Container v%s", Version)

	// Step 1: Load configuration
	// Source: Architecture §6.1 Initialization
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Println("Configuration loaded successfully")

	// Step 2: Initialize telemetry hub
	// Source: Architecture §6.1 Initialization
	telemetryHub := telemetry.NewHub(cfg)
	if telemetryHub == nil {
		log.Fatal("Failed to create telemetry hub")
	}
	log.Println("Telemetry hub initialized")

	// Step 3: Initialize audit logger
	// Source: Architecture §6.1 Initialization
	auditLogger, err := audit.NewLogger("logs")
	if err != nil {
		log.Fatalf("Failed to initialize audit logger: %v", err)
	}
	log.Println("Audit logger initialized")

	// Step 4: Initialize radio manager
	// Source: Architecture §6.1 Initialization
	radioManager := radio.NewManager()
	if radioManager == nil {
		log.Fatal("Failed to create radio manager")
	}
	log.Println("Radio manager initialized")

	// Step 5: Create command orchestrator
	// Source: Architecture §6.1 Initialization
	orchestrator := command.NewOrchestrator(telemetryHub, cfg)
	orchestrator.SetAuditLogger(auditLogger)

	// Step 6: Create API server with all components
	// Source: Architecture §6.1 Initialization
	server := api.NewServer(telemetryHub, orchestrator, radioManager, 30*time.Second, 30*time.Second, 120*time.Second)
	if server == nil {
		log.Fatal("Failed to create API server")
	}
	log.Println("API server created")

	// Step 7: Start HTTP server
	// Source: Architecture §6.1 Initialization
	addr := getServerAddress()
	log.Printf("Starting HTTP server on %s", addr)

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := server.Start(addr); err != nil {
			serverErr <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	// Wait a moment for server to start
	time.Sleep(100 * time.Millisecond)

	// Log successful startup
	log.Printf("Radio Control Container started successfully")
	log.Printf("Health endpoint: http://localhost%s/api/v1/health", addr)
	log.Printf("API base URL: http://localhost%s/api/v1", addr)

	// Set up graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal or server error
	select {
	case sig := <-shutdown:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
	case err := <-serverErr:
		log.Printf("Server error: %v", err)
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop telemetry hub
	telemetryHub.Stop()
	log.Println("Telemetry hub stopped")

	// Stop audit logger
	if err := auditLogger.Close(); err != nil {
		log.Printf("Error closing audit logger: %v", err)
	}
	log.Println("Audit logger closed")

	// Stop HTTP server
	if err := server.Stop(ctx); err != nil {
		log.Printf("Error stopping HTTP server: %v", err)
	} else {
		log.Println("HTTP server stopped gracefully")
	}

	log.Println("Radio Control Container shutdown complete")
}

// getServerAddress returns the server address from environment or default.
func getServerAddress() string {
	if addr := os.Getenv("RCC_ADDR"); addr != "" {
		return addr
	}
	return DefaultAddr
}
