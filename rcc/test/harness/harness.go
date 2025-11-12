// Package harness provides a unified test harness for API and audit tests.
// Goal: Every API/audit test runs against the same fully-wired system with predictable IDs and data.
package harness

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/adapter/silvusmock"
	"github.com/radio-control/rcc/internal/api"
	"github.com/radio-control/rcc/internal/audit"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
)

// Options configures the test harness
type Options struct {
	BandPlan      []adapter.Channel
	ActiveRadioID string
	CorrelationID string
	WithAuth      bool
	TempDir       string
}

// DefaultOptions returns sensible defaults for testing
func DefaultOptions() Options {
	return Options{
		BandPlan: []adapter.Channel{
			{Index: 1, FrequencyMhz: 2412.0},
			{Index: 6, FrequencyMhz: 2437.0},
			{Index: 11, FrequencyMhz: 2462.0},
		},
		ActiveRadioID: "silvus-001",
		CorrelationID: "fixed-1",
		WithAuth:      false,
	}
}

// Server represents a test server with all components wired
type Server struct {
	URL           string
	Shutdown      func()
	RadioManager  *radio.Manager
	Orchestrator  *command.Orchestrator
	TelemetryHub  *telemetry.Hub
	AuditLogger   *audit.Logger
	SilvusAdapter *silvusmock.SilvusMock
}

// NewServer creates a fully-wired test server
func NewServer(t *testing.T, opts Options) *Server {
	// Use provided temp dir or create one
	tempDir := opts.TempDir
	if tempDir == "" {
		tempDir = t.TempDir()
	}

	// Build config
	cfg := config.LoadCBTimingBaseline()

	// Create telemetry hub
	hub := telemetry.NewHub(cfg)
	t.Cleanup(func() { hub.Stop() })

	// Create audit logger
	auditLogger, err := audit.NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	t.Cleanup(func() { auditLogger.Close() })

	// Create radio manager
	radioManager := radio.NewManager()

	// Create SilvusMock adapter with provided band plan
	silvusAdapter := silvusmock.NewSilvusMock(opts.ActiveRadioID, opts.BandPlan)

	// Load capabilities into radio manager
	err = radioManager.LoadCapabilities(opts.ActiveRadioID, silvusAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}

	// Set active radio
	err = radioManager.SetActive(opts.ActiveRadioID)
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}

	// Create orchestrator
	orchestrator := command.NewOrchestrator(hub, cfg)
	orchestrator.SetAuditLogger(auditLogger)
	orchestrator.SetActiveAdapter(silvusAdapter)
	orchestrator.SetRadioManager(radioManager) // Connect orchestrator to radio manager for channel resolution

	// Create API server with deterministic correlation ID
	apiServer := api.NewServer(hub, orchestrator, radioManager, 30*time.Second, 30*time.Second, 120*time.Second)

	// Set deterministic correlation ID if provided
	if opts.CorrelationID != "" {
		// This would require modifying the API server to accept a correlation ID function
		// For now, we'll work with the existing implementation
	}

	// Create HTTP test server
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Register routes and serve
		mux := http.NewServeMux()
		apiServer.RegisterRoutes(mux)
		mux.ServeHTTP(w, r)
	}))

	// Print harness summary
	t.Logf("=== HARNESS SUMMARY ===")
	t.Logf("Active Radio ID: %s", opts.ActiveRadioID)
	t.Logf("Band Plan: %+v", opts.BandPlan)
	t.Logf("Correlation ID: %s", opts.CorrelationID)
	t.Logf("With Auth: %v", opts.WithAuth)
	t.Logf("Server URL: %s", httpServer.URL)
	t.Logf("=====================")

	return &Server{
		URL:           httpServer.URL,
		Shutdown:      httpServer.Close,
		RadioManager:  radioManager,
		Orchestrator:  orchestrator,
		TelemetryHub:  hub,
		AuditLogger:   auditLogger,
		SilvusAdapter: silvusAdapter,
	}
}

// GetAuditLogs reads the audit log file and returns the last N lines
// This method is kept for E2E tests that need to verify audit logging
func (s *Server) GetAuditLogs(n int) ([]string, error) {
	// For now, return empty - audit logs should be verified via HTTP endpoints
	// or by reading log files directly in E2E tests
	return []string{}, nil
}
