package command

import (
	"context"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/radio"
)

// TestPublishPowerChangedEventFailure tests publishPowerChangedEvent failure scenarios
func TestPublishPowerChangedEventFailure(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	telemetryHub := NewFailingTelemetryHub(cfg, true)
	defer telemetryHub.Stop()
	auditLogger := &MockAuditLogger{}

	orchestrator := NewOrchestrator(telemetryHub.Hub, cfg)
	orchestrator.SetAuditLogger(auditLogger)
	orchestrator.SetActiveAdapter(&MockAdapter{})

	ctx := context.Background()

	// Test SetPower with telemetry publish failure
	err := orchestrator.SetPower(ctx, "radio-01", 30.0)
	if err != nil {
		t.Errorf("SetPower should not fail due to telemetry publish error: %v", err)
	}

	// Verify that a fault event was published for the telemetry failure
	if len(telemetryHub.events) == 0 {
		t.Error("Expected fault event for telemetry publish failure")
	}

	// Check that the fault event has the correct structure
	faultEvent := telemetryHub.events[len(telemetryHub.events)-1]
	if faultEvent.Type != "fault" {
		t.Errorf("Expected fault event type, got %s", faultEvent.Type)
	}

	if faultEvent.Data["message"] != "Failed to publish power changed event" {
		t.Errorf("Expected fault message about power event, got %s", faultEvent.Data["message"])
	}
}

// TestPublishChannelChangedEventFailure tests publishChannelChangedEvent failure scenarios
func TestPublishChannelChangedEventFailure(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	telemetryHub := NewFailingTelemetryHub(cfg, true)
	defer telemetryHub.Stop()
	auditLogger := &MockAuditLogger{}

	orchestrator := NewOrchestrator(telemetryHub.Hub, cfg)
	orchestrator.SetAuditLogger(auditLogger)
	orchestrator.SetActiveAdapter(&MockAdapter{})

	ctx := context.Background()

	// Test SetChannel with telemetry publish failure
	err := orchestrator.SetChannel(ctx, "radio-01", 2412.0)
	if err != nil {
		t.Errorf("SetChannel should not fail due to telemetry publish error: %v", err)
	}

	// Verify that a fault event was published for the telemetry failure
	if len(telemetryHub.events) == 0 {
		t.Error("Expected fault event for telemetry publish failure")
	}

	// Check that the fault event has the correct structure
	faultEvent := telemetryHub.events[len(telemetryHub.events)-1]
	if faultEvent.Type != "fault" {
		t.Errorf("Expected fault event type, got %s", faultEvent.Type)
	}

	if faultEvent.Data["message"] != "Failed to publish channel changed event" {
		t.Errorf("Expected fault message about channel event, got %s", faultEvent.Data["message"])
	}
}

// TestResolveChannelIndexWithSilvusBandPlan tests resolveChannelIndex with Silvus band plan
func TestResolveChannelIndexWithSilvusBandPlan(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()

	// Create a test band plan
	cfg.SilvusBandPlan = &config.SilvusBandPlan{
		Models: map[string]map[string][]config.SilvusChannel{
			"Silvus-Scout": {
				"2.4GHz": {
					{ChannelIndex: 1, FrequencyMhz: 2412.0},
					{ChannelIndex: 2, FrequencyMhz: 2417.0},
					{ChannelIndex: 3, FrequencyMhz: 2422.0},
				},
			},
		},
	}

	orchestrator := NewOrchestrator(nil, cfg)

	// Create a mock radio manager that returns model and band
	mockRadioManager := &MockRadioManager{
		Radios: map[string]*radio.Radio{
			"radio-01": {
				ID:    "radio-01",
				Model: "Silvus-Scout",
			},
		},
	}
	orchestrator.SetRadioManager(mockRadioManager)

	ctx := context.Background()

	// Test resolveChannelIndex with valid channel
	frequency, err := orchestrator.resolveChannelIndex(ctx, "radio-01", 1, mockRadioManager)
	if err != nil {
		t.Errorf("resolveChannelIndex failed: %v", err)
	}

	if frequency != 2412.0 {
		t.Errorf("Expected frequency 2412.0, got %f", frequency)
	}

	// Test resolveChannelIndex with invalid channel
	_, err = orchestrator.resolveChannelIndex(ctx, "radio-01", 99, mockRadioManager)
	if err == nil {
		t.Error("Expected error for invalid channel index")
	}
}

// TestResolveChannelIndexWithMissingBandPlan tests resolveChannelIndex when band plan is missing
func TestResolveChannelIndexWithMissingBandPlan(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	cfg.SilvusBandPlan = nil // No band plan

	orchestrator := NewOrchestrator(nil, cfg)

	// Create a mock radio manager
	mockRadioManager := &MockRadioManager{
		Radios: map[string]*radio.Radio{
			"radio-01": {
				ID: "radio-01",
				Capabilities: &adapter.RadioCapabilities{
					Channels: []adapter.Channel{
						{Index: 1, FrequencyMhz: 2412.0},
						{Index: 2, FrequencyMhz: 2417.0},
					},
				},
			},
		},
	}
	orchestrator.SetRadioManager(mockRadioManager)

	ctx := context.Background()

	// Test resolveChannelIndex should fall back to radio manager
	frequency, err := orchestrator.resolveChannelIndex(ctx, "radio-01", 1, mockRadioManager)
	if err != nil {
		t.Errorf("resolveChannelIndex failed: %v", err)
	}

	if frequency != 2412.0 {
		t.Errorf("Expected frequency 2412.0, got %f", frequency)
	}
}

// TestResolveChannelIndexWithInvalidRadio tests resolveChannelIndex with invalid radio
func TestResolveChannelIndexWithInvalidRadio(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	orchestrator := NewOrchestrator(nil, cfg)

	// Create a mock radio manager that doesn't have the radio
	mockRadioManager := &MockRadioManager{
		Radios: map[string]*radio.Radio{},
	}
	orchestrator.SetRadioManager(mockRadioManager)

	ctx := context.Background()

	// Test resolveChannelIndex with invalid radio
	_, err := orchestrator.resolveChannelIndex(ctx, "invalid-radio", 1, mockRadioManager)
	if err == nil {
		t.Error("Expected error for invalid radio")
	}
}

// TestResolveChannelIndexWithZeroIndex tests resolveChannelIndex with zero index
func TestResolveChannelIndexWithZeroIndex(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	orchestrator := NewOrchestrator(nil, cfg)

	// Create a mock radio manager
	mockRadioManager := &MockRadioManager{
		Radios: map[string]*radio.Radio{
			"radio-01": {
				ID: "radio-01",
				Capabilities: &adapter.RadioCapabilities{
					Channels: []adapter.Channel{
						{Index: 1, FrequencyMhz: 2412.0},
					},
				},
			},
		},
	}
	orchestrator.SetRadioManager(mockRadioManager)

	ctx := context.Background()

	// Test resolveChannelIndex with zero index (should fail validation)
	_, err := orchestrator.resolveChannelIndex(ctx, "radio-01", 0, mockRadioManager)
	if err == nil {
		t.Error("Expected error for zero channel index")
	}
}

// TestResolveChannelIndexWithNegativeIndex tests resolveChannelIndex with negative index
func TestResolveChannelIndexWithNegativeIndex(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	orchestrator := NewOrchestrator(nil, cfg)

	// Create a mock radio manager
	mockRadioManager := &MockRadioManager{
		Radios: map[string]*radio.Radio{
			"radio-01": {
				ID: "radio-01",
				Capabilities: &adapter.RadioCapabilities{
					Channels: []adapter.Channel{
						{Index: 1, FrequencyMhz: 2412.0},
					},
				},
			},
		},
	}
	orchestrator.SetRadioManager(mockRadioManager)

	ctx := context.Background()

	// Test resolveChannelIndex with negative index (should fail validation)
	_, err := orchestrator.resolveChannelIndex(ctx, "radio-01", -1, mockRadioManager)
	if err == nil {
		t.Error("Expected error for negative channel index")
	}
}

// TestResolveChannelIndexWithMissingChannels tests resolveChannelIndex when radio has no channels
func TestResolveChannelIndexWithMissingChannels(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	orchestrator := NewOrchestrator(nil, cfg)

	// Create a mock radio manager with radio that has no channels
	mockRadioManager := &MockRadioManager{
		Radios: map[string]*radio.Radio{
			"radio-01": {
				ID: "radio-01",
				Capabilities: &adapter.RadioCapabilities{
					Channels: []adapter.Channel{}, // No channels
				},
			},
		},
	}
	orchestrator.SetRadioManager(mockRadioManager)

	ctx := context.Background()

	// Test resolveChannelIndex with no channels
	_, err := orchestrator.resolveChannelIndex(ctx, "radio-01", 1, mockRadioManager)
	if err == nil {
		t.Error("Expected error for radio with no channels")
	}
}

// TestResolveChannelIndexWithNilCapabilities tests resolveChannelIndex when radio has nil capabilities
func TestResolveChannelIndexWithNilCapabilities(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	orchestrator := NewOrchestrator(nil, cfg)

	// Create a mock radio manager with radio that has nil capabilities
	mockRadioManager := &MockRadioManager{
		Radios: map[string]*radio.Radio{
			"radio-01": {
				ID:           "radio-01",
				Capabilities: nil, // No capabilities
			},
		},
	}
	orchestrator.SetRadioManager(mockRadioManager)

	ctx := context.Background()

	// Test resolveChannelIndex with nil capabilities
	_, err := orchestrator.resolveChannelIndex(ctx, "radio-01", 1, mockRadioManager)
	if err == nil {
		t.Error("Expected error for radio with nil capabilities")
	}
}

// TestResolveChannelIndexTimeout tests resolveChannelIndex with timeout
func TestResolveChannelIndexTimeout(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	// Set a very short timeout for testing
	cfg.CommandTimeoutSetChannel = 1 * time.Millisecond
	orchestrator := NewOrchestrator(nil, cfg)

	// Create a mock radio manager
	mockRadioManager := &MockRadioManager{
		Radios: map[string]*radio.Radio{
			"radio-01": {
				ID: "radio-01",
				Capabilities: &adapter.RadioCapabilities{
					Channels: []adapter.Channel{
						{Index: 1, FrequencyMhz: 2412.0},
					},
				},
			},
		},
	}
	orchestrator.SetRadioManager(mockRadioManager)

	ctx := context.Background()

	// Test resolveChannelIndex with timeout
	_, err := orchestrator.resolveChannelIndex(ctx, "radio-01", 1, mockRadioManager)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

// TestGetRadioModelAndBandWithInvalidRadio tests getRadioModelAndBand with invalid radio
func TestGetRadioModelAndBandWithInvalidRadio(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	orchestrator := NewOrchestrator(nil, cfg)

	// Create a mock radio manager that doesn't have the radio
	mockRadioManager := &MockRadioManager{
		Radios: map[string]*radio.Radio{},
	}
	orchestrator.SetRadioManager(mockRadioManager)

	ctx := context.Background()

	// Test getRadioModelAndBand with invalid radio
	_, _, err := orchestrator.getRadioModelAndBand(ctx, "invalid-radio", mockRadioManager)
	if err == nil {
		t.Error("Expected error for invalid radio")
	}
}

// TestGetRadioModelAndBandWithNilManager tests getRadioModelAndBand with nil radio manager
func TestGetRadioModelAndBandWithNilManager(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	orchestrator := NewOrchestrator(nil, cfg)
	orchestrator.SetRadioManager(nil) // No radio manager

	ctx := context.Background()

	// Test getRadioModelAndBand with nil manager
	_, _, err := orchestrator.getRadioModelAndBand(ctx, "radio-01", nil)
	if err == nil {
		t.Error("Expected error for nil radio manager")
	}
}
