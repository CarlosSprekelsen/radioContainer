//go:build integration

package orchestrator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/audit"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
	"github.com/radio-control/rcc/test/fixtures"
	"github.com/radio-control/rcc/test/integration/fakes"
)

func TestOrchestratorIntegration_CommandValidation(t *testing.T) {
	// Test orchestrator command validation with real components
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)

	// Create real components
	radioManager := radio.NewManager()
	auditLogger, err := audit.NewLogger("/tmp/audit_test")
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}

	// Create orchestrator with real components
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)
	orchestrator.SetAuditLogger(auditLogger)

	// Test with invalid radio ID (should get NOT_FOUND)
	invalidRadioID := "nonexistent-radio"
	err = orchestrator.SetChannel(context.Background(), invalidRadioID, 2412.0)
	if err == nil {
		t.Error("Expected error for invalid radio ID")
	}

	// Should get NOT_FOUND, not UNAVAILABLE
	if err != nil && !errors.Is(err, command.ErrNotFound) {
		t.Errorf("Expected command.ErrNotFound, got: %v", err)
	}

	// Test with valid radio ID but no adapter (should get UNAVAILABLE)
	validRadioID := "test-radio-001"

	// Create a fake adapter but don't set it as active
	fakeAdapter := fakes.NewFakeAdapter("test-radio-001")

	// Load capabilities for the radio
	err = radioManager.LoadCapabilities(validRadioID, fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}

	// Set radio as active
	err = radioManager.SetActive(validRadioID)
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}

	// Now test - should get UNAVAILABLE because no active adapter set
	err = orchestrator.SetChannel(context.Background(), validRadioID, 2412.0)
	if err == nil {
		t.Error("Expected error for radio without active adapter")
	}

	// Should get UNAVAILABLE for missing active adapter
	if err != nil && !errors.Is(err, adapter.ErrUnavailable) {
		t.Errorf("Expected adapter.ErrUnavailable, got: %v", err)
	}

	// Clean up
	telemetryHub.Stop()
}

func TestOrchestratorIntegration_PowerCommand(t *testing.T) {
	// Test power command validation with real components
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)

	// Create real components
	radioManager := radio.NewManager()
	auditLogger, err := audit.NewLogger("/tmp/audit_test")
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}

	// Create orchestrator with real components
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)
	orchestrator.SetAuditLogger(auditLogger)

	// Test power command with valid radio but no adapter
	radioID := "test-radio-002"

	// Create a fake adapter but don't set it as active
	fakeAdapter := fakes.NewFakeAdapter("test-radio-002")

	// Load capabilities for the radio
	err = radioManager.LoadCapabilities(radioID, fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}

	// Set radio as active
	err = radioManager.SetActive(radioID)
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}

	// Test power command - should get UNAVAILABLE because no active adapter set
	err = orchestrator.SetPower(context.Background(), radioID, 5.0)
	if err == nil {
		t.Error("Expected error for radio without active adapter")
	}

	// Verify error type
	if err != nil && !errors.Is(err, adapter.ErrUnavailable) {
		t.Errorf("Expected adapter.ErrUnavailable, got: %v", err)
	}

	// Clean up
	telemetryHub.Stop()
}

func TestOrchestratorIntegration_ChannelByIndex(t *testing.T) {
	// Test channel by index command with real components
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)

	// Create real components
	radioManager := radio.NewManager()
	auditLogger, err := audit.NewLogger("/tmp/audit_test")
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}

	// Create orchestrator with real components
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)
	orchestrator.SetAuditLogger(auditLogger)

	// Test with valid radio but no adapter
	radioID := "test-radio-003"

	// Create a fake adapter but don't set it as active
	fakeAdapter := fakes.NewFakeAdapter("test-radio-003")

	// Load capabilities for the radio
	err = radioManager.LoadCapabilities(radioID, fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}

	// Set radio as active
	err = radioManager.SetActive(radioID)
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}

	// Test channel by index - should get UNAVAILABLE because no active adapter set
	err = orchestrator.SetChannelByIndex(context.Background(), radioID, 6, radioManager)
	if err == nil {
		t.Error("Expected error for radio without active adapter")
	}

	// Verify error type
	if err != nil && !errors.Is(err, adapter.ErrUnavailable) {
		t.Errorf("Expected adapter.ErrUnavailable, got: %v", err)
	}

	// Clean up
	telemetryHub.Stop()
}

func TestOrchestratorIntegration_GetState(t *testing.T) {
	// Test get state command with real components
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)

	// Create real components
	radioManager := radio.NewManager()
	auditLogger, err := audit.NewLogger("/tmp/audit_test")
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}

	// Create orchestrator with real components
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)
	orchestrator.SetAuditLogger(auditLogger)

	// Test with valid radio but no adapter
	radioID := "test-radio-004"

	// Create a fake adapter but don't set it as active
	fakeAdapter := fakes.NewFakeAdapter("test-radio-004")

	// Load capabilities for the radio
	err = radioManager.LoadCapabilities(radioID, fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}

	// Set radio as active
	err = radioManager.SetActive(radioID)
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}

	// Test get state - should get UNAVAILABLE because no active adapter set
	_, err = orchestrator.GetState(context.Background(), radioID)
	if err == nil {
		t.Error("Expected error for radio without active adapter")
	}

	// Verify error type
	if err != nil && !errors.Is(err, adapter.ErrUnavailable) {
		t.Errorf("Expected adapter.ErrUnavailable, got: %v", err)
	}

	// Clean up
	telemetryHub.Stop()
}

func TestOrchestratorIntegration_TimingConstraints(t *testing.T) {
	// Test timing constraints with real components
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)

	// Create real components
	radioManager := radio.NewManager()
	auditLogger, err := audit.NewLogger("/tmp/audit_test")
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}

	// Create orchestrator with real components
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)
	orchestrator.SetAuditLogger(auditLogger)

	// Test timing constraints are loaded from config
	if cfg.CommandTimeoutSetPower == 0 {
		t.Error("Command timeout should be loaded from config")
	}
	if cfg.CommandTimeoutSetChannel == 0 {
		t.Error("Command timeout should be loaded from config")
	}

	// Clean up
	telemetryHub.Stop()
}
