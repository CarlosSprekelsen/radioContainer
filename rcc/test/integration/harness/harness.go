//go:build integration

// Package harness provides a minimal in-process integration test harness.
// Boundary: command+radio+adapter+telemetry+audit; no HTTP; deterministic.
package harness

import (
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/audit"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
	"github.com/radio-control/rcc/test/integration/fakes"
	"github.com/radio-control/rcc/test/integration/fixtures"
)

// BuildCommandStack wires real implementations via public constructors only.
// Returns components via their public interfaces/ports with automatic cleanup.
// Only mocks external radio adapter - all internal components are real.
func BuildCommandStack(t *testing.T, seed Radios) (orch command.OrchestratorPort, rm *radio.Manager, tele *telemetry.Hub, auditLogger *audit.Logger, adapter adapter.IRadioAdapter) {
	// Create test config
	cfg := fixtures.TestTimingConfig()

	// Create real telemetry hub
	tele = telemetry.NewHub(cfg)

	// Register cleanup for telemetry hub
	t.Cleanup(func() {
		tele.Stop()
	})

	// Create real audit logger (writes to temp dir for tests)
	auditLogger, err := audit.NewLogger("/tmp/audit_test")
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}

	// Create real radio manager
	rm = radio.NewManager()

	// Create orchestrator with real components
	orchestrator := command.NewOrchestratorWithRadioManager(tele, cfg, rm)
	orchestrator.SetAuditLogger(auditLogger)

	// Seed radios if provided and return the first adapter
	if seed != nil {
		for id, fakeAdapter := range seed {
			if err := rm.LoadCapabilities(id, fakeAdapter, 5*time.Second); err != nil {
				t.Fatalf("Failed to seed radio %s: %v", id, err)
			}
			orchestrator.SetActiveAdapter(fakeAdapter)
			adapter = fakeAdapter // Return the adapter for test verification
		}
	}

	return orchestrator, rm, tele, auditLogger, adapter
}

// Radios represents a collection of radios to seed for testing.
type Radios map[string]adapter.IRadioAdapter

// BuildTestStack creates a complete test stack with fake adapters and fixtures.
func BuildTestStack(t *testing.T) (orch command.OrchestratorPort, rm *radio.Manager, tele *telemetry.Hub, auditLogger *audit.Logger, adapter adapter.IRadioAdapter) {
	// Create fake adapters (only external dependency mocked)
	fakeAdapter := fakes.NewFakeAdapter("fake-001").
		WithInitial(20.0, 2412.0, nil) // No channels needed for basic tests

	seedRadios := Radios{
		"fake-001": fakeAdapter,
	}

	// Build the command stack
	return BuildCommandStack(t, seedRadios)
}
