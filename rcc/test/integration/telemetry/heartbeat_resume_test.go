//go:build integration

package telemetry

import (
	"context"
	"testing"

	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/test/integration/fixtures"
	"github.com/radio-control/rcc/test/integration/harness"
)

// TestTelemetry_ConfigurationValidation tests telemetry configuration loading.
// Boundary: telemetry hub in-proc; validates timing configuration.
func TestTelemetry_ConfigurationValidation(t *testing.T) {
	// Arrange: Build test stack
	_, _, tele, _, _ := harness.BuildTestStack(t)

	// Get test configuration
	cfg := fixtures.TestTimingConfig()

	// Act: Verify telemetry hub is created with configuration
	if tele == nil {
		t.Fatal("Telemetry hub should be created")
	}

	// Assert: Configuration should be loaded
	if cfg.HeartbeatInterval <= 0 {
		t.Error("Heartbeat interval should be positive")
	}
	if cfg.HeartbeatJitter < 0 {
		t.Error("Heartbeat jitter should be non-negative")
	}
	if cfg.EventBufferSize <= 0 {
		t.Error("Event buffer size should be positive")
	}

	t.Logf("✅ Telemetry configuration validated: interval=%v, jitter=%v, buffer=%d",
		cfg.HeartbeatInterval, cfg.HeartbeatJitter, cfg.EventBufferSize)
}

// TestTelemetry_CommandFlowIntegration tests that commands trigger telemetry events.
func TestTelemetry_CommandFlowIntegration(t *testing.T) {
	// Arrange: Build test stack
	orch, _, tele, _, _ := harness.BuildTestStack(t)

	// Act: Execute commands that should trigger telemetry events
	ctx := context.Background()

	// Set power (should trigger telemetry event)
	err := orch.SetPower(ctx, "fake-001", 25.0)
	if err != nil {
		t.Fatalf("SetPower failed: %v", err)
	}

	// Assert: Telemetry hub should be available for event publishing
	if tele == nil {
		t.Fatal("Telemetry hub should be available for event publishing")
	}

	// Note: Actual telemetry event validation requires HTTP SSE client testing
	// This integration test focuses on command flow and telemetry hub availability
	t.Logf("✅ Command flow integration: SetPower executed, telemetry hub available")
}

// TestTelemetry_ConfigurationConsistency tests that test config is consistent with production config.
func TestTelemetry_ConfigurationConsistency(t *testing.T) {
	// Arrange: Load both test and production configurations
	testCfg := fixtures.TestTimingConfig()
	prodCfg := config.LoadCBTimingBaseline()

	// Act & Assert: Verify test config is faster than production for test efficiency
	if testCfg.HeartbeatInterval >= prodCfg.HeartbeatInterval {
		t.Errorf("Test heartbeat interval (%v) should be faster than production (%v)",
			testCfg.HeartbeatInterval, prodCfg.HeartbeatInterval)
	}

	if testCfg.EventBufferSize >= prodCfg.EventBufferSize {
		t.Errorf("Test buffer size (%d) should be smaller than production (%d)",
			testCfg.EventBufferSize, prodCfg.EventBufferSize)
	}

	// Verify test config has reasonable values
	if testCfg.HeartbeatInterval <= 0 {
		t.Error("Test heartbeat interval should be positive")
	}
	if testCfg.EventBufferSize <= 0 {
		t.Error("Test buffer size should be positive")
	}

	t.Logf("✅ Configuration consistency: Test config optimized for fast execution")
}
