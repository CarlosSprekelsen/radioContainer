package command

import (
	"context"
	"errors"
	"testing"

	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/telemetry"
)

// FailingTelemetryHub wraps a real telemetry.Hub but can simulate publish failures
type FailingTelemetryHub struct {
	*telemetry.Hub
	shouldFail bool
	events     []telemetry.Event
}

func NewFailingTelemetryHub(cfg *config.TimingConfig, shouldFail bool) *FailingTelemetryHub {
	return &FailingTelemetryHub{
		Hub:        telemetry.NewHub(cfg),
		shouldFail: shouldFail,
		events:     make([]telemetry.Event, 0),
	}
}

func (f *FailingTelemetryHub) Publish(event telemetry.Event) error {
	if f.shouldFail {
		return errors.New("telemetry publish failed")
	}
	f.events = append(f.events, event)
	return f.Hub.Publish(event)
}

func (f *FailingTelemetryHub) PublishRadio(radioID string, event telemetry.Event) error {
	if f.shouldFail {
		return errors.New("telemetry publish failed")
	}
	event.Radio = radioID
	f.events = append(f.events, event)
	return f.Hub.PublishRadio(radioID, event)
}

// TestTelemetryPublishErrorHandling tests that telemetry publish errors are handled gracefully
func TestTelemetryPublishErrorHandling(t *testing.T) {
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

// TestTelemetryPublishErrorHandlingChannel tests channel event telemetry publish errors
func TestTelemetryPublishErrorHandlingChannel(t *testing.T) {
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

// TestTelemetryPublishErrorHandlingState tests state event telemetry publish errors
func TestTelemetryPublishErrorHandlingState(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	telemetryHub := NewFailingTelemetryHub(cfg, true)
	defer telemetryHub.Stop()
	auditLogger := &MockAuditLogger{}

	orchestrator := NewOrchestrator(telemetryHub.Hub, cfg)
	orchestrator.SetAuditLogger(auditLogger)
	orchestrator.SetActiveAdapter(&MockAdapter{})

	ctx := context.Background()

	// Test GetState with telemetry publish failure
	_, err := orchestrator.GetState(ctx, "radio-01")
	if err != nil {
		t.Errorf("GetState should not fail due to telemetry publish error: %v", err)
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

	if faultEvent.Data["message"] != "Failed to publish state event" {
		t.Errorf("Expected fault message about state event, got %s", faultEvent.Data["message"])
	}
}

// TestTelemetryPublishErrorWithNilHub tests that nil telemetry hub doesn't cause panics
func TestTelemetryPublishErrorWithNilHub(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	auditLogger := &MockAuditLogger{}

	orchestrator := NewOrchestrator(nil, cfg)
	orchestrator.SetAuditLogger(auditLogger)
	orchestrator.SetActiveAdapter(&MockAdapter{})

	ctx := context.Background()

	// Test SetPower with nil telemetry hub
	err := orchestrator.SetPower(ctx, "radio-01", 30.0)
	if err != nil {
		t.Errorf("SetPower should not fail with nil telemetry hub: %v", err)
	}

	// Test SetChannel with nil telemetry hub
	err = orchestrator.SetChannel(ctx, "radio-01", 2412.0)
	if err != nil {
		t.Errorf("SetChannel should not fail with nil telemetry hub: %v", err)
	}

	// Test GetState with nil telemetry hub
	_, err = orchestrator.GetState(ctx, "radio-01")
	if err != nil {
		t.Errorf("GetState should not fail with nil telemetry hub: %v", err)
	}
}
