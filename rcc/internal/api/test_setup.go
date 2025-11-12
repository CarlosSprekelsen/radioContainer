package api

import (
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/adapter/silvusmock"
	"github.com/radio-control/rcc/internal/audit"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
)

// setupAPITest creates a fully wired API test environment with SilvusMock
func setupAPITest(t *testing.T) (*Server, *radio.Manager, *command.Orchestrator, adapter.IRadioAdapter) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	t.Cleanup(func() { hub.Stop() })

	// Create radio manager and register SilvusMock
	rm := radio.NewManager()
	adapter := silvusmock.NewSilvusMock("silvus-001", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412},
		{Index: 6, FrequencyMhz: 2437},
		{Index: 11, FrequencyMhz: 2462},
	})

	// Load capabilities to register adapter
	rm.LoadCapabilities("silvus-001", adapter, 5*time.Second)
	rm.SetActive("silvus-001")

	// Create orchestrator and set radio manager
	orch := command.NewOrchestrator(hub, cfg)
	orch.SetRadioManager(rm)

	// Set up audit logger
	auditLogger, err := audit.NewLogger(t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	orch.SetAuditLogger(auditLogger)

	// Set the active adapter on the orchestrator
	orch.SetActiveAdapter(adapter)

	// Create API server
	server := NewServer(hub, orch, rm, 30*time.Second, 30*time.Second, 120*time.Second)

	return server, rm, orch, adapter
}

// setupAPITestWithFault creates API test environment with specific fault mode
func setupAPITestWithFault(t *testing.T, faultMode string) (*Server, *radio.Manager, *command.Orchestrator, *silvusmock.SilvusMock) {
	server, rm, orch, adapterIface := setupAPITest(t)

	// Type-assert to SilvusMock to set fault mode
	adapter, ok := adapterIface.(*silvusmock.SilvusMock)
	if !ok {
		t.Fatalf("Failed to type-assert adapter to *silvusmock.SilvusMock")
	}

	// Set fault mode if specified
	if faultMode != "" {
		adapter.SetFaultMode(faultMode)
	}

	return server, rm, orch, adapter
}
