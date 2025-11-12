//go:build integration

package command

import (
	"context"
	"testing"

	"github.com/radio-control/rcc/internal/audit"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
	"github.com/radio-control/rcc/test/fixtures"
	"github.com/radio-control/rcc/test/integration/fakes"
)

// TestCoverage_RealComponents tests real orchestrator with mock adapter to demonstrate coverage.
func TestCoverage_RealComponents(t *testing.T) {
	// Create real components
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)
	radioManager := radio.NewManager()
	auditLogger, err := audit.NewLogger("/tmp/audit_test")
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	
	// Create orchestrator with real components
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)
	orchestrator.SetAuditLogger(auditLogger)

	// Create fake adapter (only external dependency mocked)
	fakeAdapter := fakes.NewFakeAdapter("test-radio-coverage")
	
	// Load capabilities and set active
	err = radioManager.LoadCapabilities("test-radio-coverage", fakeAdapter, 5)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}
	
	err = radioManager.SetActive("test-radio-coverage")
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}
	
	orchestrator.SetActiveAdapter(fakeAdapter)

	// Test real orchestrator methods
	ctx := context.Background()
	
	// This should succeed and cover real orchestrator code
	err = orchestrator.SetPower(ctx, "test-radio-coverage", 25.0)
	if err != nil {
		t.Errorf("SetPower failed: %v", err)
	}
	
	// This should succeed and cover real orchestrator code
	err = orchestrator.SetChannel(ctx, "test-radio-coverage", 2412.0)
	if err != nil {
		t.Errorf("SetChannel failed: %v", err)
	}
	
	// This should succeed and cover real orchestrator code
	err = orchestrator.SetChannelByIndex(ctx, "test-radio-coverage", 1, radioManager)
	if err != nil {
		t.Errorf("SetChannelByIndex failed: %v", err)
	}
	
	// This should succeed and cover real orchestrator code
	state, err := orchestrator.GetState(ctx, "test-radio-coverage")
	if err != nil {
		t.Errorf("GetState failed: %v", err)
	}
	if state == nil {
		t.Error("Expected non-nil state")
	}

	// Clean up
	telemetryHub.Stop()
	
	t.Logf("✅ Coverage test: Real orchestrator methods executed successfully")
}
