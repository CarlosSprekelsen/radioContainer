package command

import (
	"context"
	"fmt"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
)

// Orchestrator routes validated API intents to the active adapter.
type Orchestrator struct {
	// Active radio adapter
	activeAdapter adapter.IRadioAdapter

	// Telemetry hub for event publishing
	telemetryHub *telemetry.Hub

	// Configuration for validation
	config *config.TimingConfig

	// Audit logger (to be implemented)
	auditLogger AuditLogger

	// Radio manager for channel index resolution
	radioManager RadioManager
}

// Compile-time assertion that radio.Manager implements RadioManager
var _ RadioManager = (*radio.Manager)(nil)

// Compile-time assertion that Orchestrator implements OrchestratorPort
var _ OrchestratorPort = (*Orchestrator)(nil)

// AuditLogger interface for writing audit records.
type AuditLogger interface {
	LogAction(ctx context.Context, action string, radioID string, result string, latency time.Duration)
}

// NewOrchestrator creates a new command orchestrator.
func NewOrchestrator(telemetryHub *telemetry.Hub, timingConfig *config.TimingConfig) *Orchestrator {
	return &Orchestrator{
		telemetryHub: telemetryHub,
		config:       timingConfig,
	}
}

// NewOrchestratorWithRadioManager creates a new command orchestrator with radio manager.
func NewOrchestratorWithRadioManager(telemetryHub *telemetry.Hub, timingConfig *config.TimingConfig, radioManager RadioManager) *Orchestrator {
	return &Orchestrator{
		telemetryHub: telemetryHub,
		config:       timingConfig,
		radioManager: radioManager,
	}
}

// SetActiveAdapter sets the active radio adapter.
func (o *Orchestrator) SetActiveAdapter(adapter adapter.IRadioAdapter) {
	o.activeAdapter = adapter
}

// SetPower sets the transmit power for the active radio in dBm.
func (o *Orchestrator) SetPower(ctx context.Context, radioID string, dBm float64) error {
	start := time.Now()

	// Ensure radio exists via radio manager
	if o.radioManager == nil {
		o.logAudit(ctx, "setPower", radioID, "UNAVAILABLE", time.Since(start))
		return adapter.ErrUnavailable
	}
	if _, err := o.radioManager.GetRadio(radioID); err != nil {
		o.logAudit(ctx, "setPower", radioID, "NOT_FOUND", time.Since(start))
		return ErrNotFound
	}

	// Validate power range
	if err := o.validatePowerRange(dBm); err != nil {
		o.logAudit(ctx, "setPower", radioID, "INVALID_RANGE", time.Since(start))
		return err
	}

	// Check if adapter is available
	if o.activeAdapter == nil {
		o.logAudit(ctx, "setPower", radioID, "UNAVAILABLE", time.Since(start))
		return adapter.ErrUnavailable
	}

	// Execute command with timeout
	timeout := o.config.CommandTimeoutSetPower
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := o.activeAdapter.SetPower(ctx, dBm)
	latency := time.Since(start)

	if err != nil {
		// Map adapter error to normalized code
		normalizedErr := adapter.NormalizeVendorError(err, nil)
		o.logAudit(ctx, "setPower", radioID, "ERROR", latency)

		// Publish fault event
		o.publishFaultEvent(radioID, normalizedErr, "Failed to set power")

		return normalizedErr
	}

	// Log successful action
	o.logAudit(ctx, "setPower", radioID, "SUCCESS", latency)

	// Publish power changed event
	o.publishPowerChangedEvent(radioID, dBm)

	return nil
}

// SetChannel sets the channel for the active radio by frequency or index.
func (o *Orchestrator) SetChannel(ctx context.Context, radioID string, frequencyMhz float64) error {
	start := time.Now()

	// Ensure radio exists via radio manager
	if o.radioManager == nil {
		o.logAudit(ctx, "setChannel", radioID, "UNAVAILABLE", time.Since(start))
		return adapter.ErrUnavailable
	}
	if _, err := o.radioManager.GetRadio(radioID); err != nil {
		o.logAudit(ctx, "setChannel", radioID, "NOT_FOUND", time.Since(start))
		return ErrNotFound
	}

	// Validate frequency range
	if err := o.validateFrequencyRange(frequencyMhz); err != nil {
		o.logAudit(ctx, "setChannel", radioID, "INVALID_RANGE", time.Since(start))
		return err
	}

	// Check if adapter is available
	if o.activeAdapter == nil {
		o.logAudit(ctx, "setChannel", radioID, "UNAVAILABLE", time.Since(start))
		return adapter.ErrUnavailable
	}

	// Execute command with timeout
	timeout := o.config.CommandTimeoutSetChannel
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := o.activeAdapter.SetFrequency(ctx, frequencyMhz)
	latency := time.Since(start)

	if err != nil {
		// Map adapter error to normalized code
		normalizedErr := adapter.NormalizeVendorError(err, nil)
		o.logAudit(ctx, "setChannel", radioID, "ERROR", latency)

		// Publish fault event
		o.publishFaultEvent(radioID, normalizedErr, "Failed to set channel")

		return normalizedErr
	}

	// Log successful action
	o.logAudit(ctx, "setChannel", radioID, "SUCCESS", latency)

	// Publish channel changed event
	o.publishChannelChangedEvent(radioID, frequencyMhz, 0) // channelIndex will be derived later

	return nil
}

// SetChannelByIndex sets the channel for the active radio by channel index.
func (o *Orchestrator) SetChannelByIndex(ctx context.Context, radioID string, channelIndex int, radioManager RadioManager) error {
	start := time.Now()

	// Ensure radio exists via radio manager
	if o.radioManager == nil {
		o.logAudit(ctx, "setChannel", radioID, "UNAVAILABLE", time.Since(start))
		return adapter.ErrUnavailable
	}
	if _, err := o.radioManager.GetRadio(radioID); err != nil {
		o.logAudit(ctx, "setChannel", radioID, "NOT_FOUND", time.Since(start))
		return ErrNotFound
	}

	// Validate channel index bounds (1-based)
	if channelIndex < 1 {
		o.logAudit(ctx, "setChannel", radioID, "INVALID_RANGE", time.Since(start))
		return adapter.ErrInvalidRange
	}

	// Check if adapter is available
	if o.activeAdapter == nil {
		o.logAudit(ctx, "setChannel", radioID, "UNAVAILABLE", time.Since(start))
		return adapter.ErrUnavailable
	}

	// Resolve channel index to frequency via radio manager
	frequencyMhz, err := o.resolveChannelIndex(ctx, radioID, channelIndex, radioManager)
	if err != nil {
		o.logAudit(ctx, "setChannel", radioID, "INVALID_RANGE", time.Since(start))
		return err
	}

	// Validate resolved frequency range
	if err := o.validateFrequencyRange(frequencyMhz); err != nil {
		o.logAudit(ctx, "setChannel", radioID, "INVALID_RANGE", time.Since(start))
		return err
	}

	// Execute command with timeout
	timeout := o.config.CommandTimeoutSetChannel
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err = o.activeAdapter.SetFrequency(ctx, frequencyMhz)
	latency := time.Since(start)

	if err != nil {
		// Map adapter error to normalized code
		normalizedErr := adapter.NormalizeVendorError(err, nil)
		o.logAudit(ctx, "setChannel", radioID, "ERROR", latency)

		// Publish fault event
		o.publishFaultEvent(radioID, normalizedErr, "Failed to set channel")

		return normalizedErr
	}

	// Log successful action
	o.logAudit(ctx, "setChannel", radioID, "SUCCESS", latency)

	// Publish channel changed event with resolved frequency and channel index
	o.publishChannelChangedEvent(radioID, frequencyMhz, channelIndex)

	return nil
}

// SelectRadio selects the active radio for subsequent operations.
func (o *Orchestrator) SelectRadio(ctx context.Context, radioID string) error {
	start := time.Now()

	// Validate radio ID
	if radioID == "" {
		o.logAudit(ctx, "selectRadio", radioID, "BAD_REQUEST", time.Since(start))
		return ErrInvalidParameter
	}

	// Ensure radio exists via radio manager
	if o.radioManager == nil {
		o.logAudit(ctx, "selectRadio", radioID, "UNAVAILABLE", time.Since(start))
		return adapter.ErrUnavailable
	}
	if _, err := o.radioManager.GetRadio(radioID); err != nil {
		o.logAudit(ctx, "selectRadio", radioID, "NOT_FOUND", time.Since(start))
		return ErrNotFound
	}

	// Select the active radio via RadioManager per Architecture §5
	if err := o.radioManager.SetActive(radioID); err != nil {
		o.logAudit(ctx, "selectRadio", radioID, "NOT_FOUND", time.Since(start))
		return ErrNotFound
	}

	// Check if adapter is available
	if o.activeAdapter == nil {
		o.logAudit(ctx, "selectRadio", radioID, "UNAVAILABLE", time.Since(start))
		return adapter.ErrUnavailable
	}

	// Execute command with timeout
	timeout := o.config.CommandTimeoutSelectRadio
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// For now, just validate the adapter is responsive
	_, err := o.activeAdapter.GetState(ctx)
	latency := time.Since(start)

	if err != nil {
		// Map adapter error to normalized code
		normalizedErr := adapter.NormalizeVendorError(err, nil)
		o.logAudit(ctx, "selectRadio", radioID, "ERROR", latency)

		// Publish fault event
		o.publishFaultEvent(radioID, normalizedErr, "Failed to select radio")

		return normalizedErr
	}

	// Log successful action
	o.logAudit(ctx, "selectRadio", radioID, "SUCCESS", latency)

	// Publish state event to confirm selection
	o.publishStateEvent(radioID)

	return nil
}

// GetState retrieves the current state of the active radio.
func (o *Orchestrator) GetState(ctx context.Context, radioID string) (*adapter.RadioState, error) {
	start := time.Now()

	// Ensure radio exists via radio manager
	if o.radioManager == nil {
		o.logAudit(ctx, "getState", radioID, "UNAVAILABLE", time.Since(start))
		return nil, adapter.ErrUnavailable
	}
	if _, err := o.radioManager.GetRadio(radioID); err != nil {
		o.logAudit(ctx, "getState", radioID, "NOT_FOUND", time.Since(start))
		return nil, ErrNotFound
	}

	// Check if adapter is available
	if o.activeAdapter == nil {
		o.logAudit(ctx, "getState", radioID, "UNAVAILABLE", time.Since(start))
		return nil, adapter.ErrUnavailable
	}

	// Execute command with timeout
	timeout := o.config.CommandTimeoutGetState
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	state, err := o.activeAdapter.GetState(ctx)
	latency := time.Since(start)

	if err != nil {
		// Map adapter error to normalized code
		normalizedErr := adapter.NormalizeVendorError(err, nil)
		o.logAudit(ctx, "getState", radioID, "ERROR", latency)

		// Publish fault event
		o.publishFaultEvent(radioID, normalizedErr, "Failed to get state")

		return nil, normalizedErr
	}

	// Log successful action
	o.logAudit(ctx, "getState", radioID, "SUCCESS", latency)

	return state, nil
}

// validatePowerRange validates the power range.
func (o *Orchestrator) validatePowerRange(dBm float64) error {
	if dBm < 0 || dBm > 39 {
		return adapter.ErrInvalidRange
	}
	return nil
}

// validateFrequencyRange validates the frequency range.
func (o *Orchestrator) validateFrequencyRange(frequencyMhz float64) error {
	// Basic frequency validation - more sophisticated validation will be added later
	// with derived channel maps
	if frequencyMhz <= 0 {
		return adapter.ErrInvalidRange
	}

	// Check against reasonable frequency ranges (will be enhanced with channel maps)
	if frequencyMhz < 100 || frequencyMhz > 6000 {
		return adapter.ErrInvalidRange
	}

	return nil
}

// publishPowerChangedEvent publishes a power changed event.
func (o *Orchestrator) publishPowerChangedEvent(radioID string, powerDbm float64) {
	if o.telemetryHub == nil {
		return // Skip if no telemetry hub
	}

	event := telemetry.Event{
		Type: "powerChanged",
		Data: map[string]interface{}{
			"radioId":  radioID,
			"powerDbm": powerDbm,
			"ts":       time.Now().UTC().Format(time.RFC3339),
		},
	}

	if err := o.telemetryHub.PublishRadio(radioID, event); err != nil {
		// Publish fault event for telemetry failure
		o.publishFaultEvent(radioID, err, "Failed to publish power changed event")
	}
}

// publishChannelChangedEvent publishes a channel changed event.
func (o *Orchestrator) publishChannelChangedEvent(radioID string, frequencyMhz float64, channelIndex int) {
	if o.telemetryHub == nil {
		return // Skip if no telemetry hub
	}

	event := telemetry.Event{
		Type: "channelChanged",
		Data: map[string]interface{}{
			"radioId":      radioID,
			"frequencyMhz": frequencyMhz,
			"channelIndex": channelIndex,
			"ts":           time.Now().UTC().Format(time.RFC3339),
		},
	}

	if err := o.telemetryHub.PublishRadio(radioID, event); err != nil {
		// Publish fault event for telemetry failure
		o.publishFaultEvent(radioID, err, "Failed to publish channel changed event")
	}
}

// publishStateEvent publishes a state event.
func (o *Orchestrator) publishStateEvent(radioID string) {
	if o.telemetryHub == nil {
		return // Skip if no telemetry hub
	}

	event := telemetry.Event{
		Type: "state",
		Data: map[string]interface{}{
			"radioId": radioID,
			"status":  "online",
			"ts":      time.Now().UTC().Format(time.RFC3339),
		},
	}

	if err := o.telemetryHub.PublishRadio(radioID, event); err != nil {
		// Publish fault event for telemetry failure
		o.publishFaultEvent(radioID, err, "Failed to publish state event")
	}
}

// publishFaultEvent publishes a fault event.
func (o *Orchestrator) publishFaultEvent(radioID string, err error, message string) {
	if o.telemetryHub == nil {
		return // Skip if no telemetry hub
	}

	event := telemetry.Event{
		Type: "fault",
		Data: map[string]interface{}{
			"radioId": radioID,
			"code":    err.Error(),
			"message": message,
			"ts":      time.Now().UTC().Format(time.RFC3339),
		},
	}

	if err := o.telemetryHub.PublishRadio(radioID, event); err != nil {
		// Silently log telemetry failure to avoid infinite recursion
		// This is a fault event itself, so we don't publish another fault
	}
}

// logAudit logs an audit record for a command action.
func (o *Orchestrator) logAudit(ctx context.Context, action, radioID, result string, latency time.Duration) {
	if o.auditLogger != nil {
		o.auditLogger.LogAction(ctx, action, radioID, result, latency)
	}
}

// SetAuditLogger sets the audit logger.
func (o *Orchestrator) SetAuditLogger(logger AuditLogger) {
	o.auditLogger = logger
}

// SetRadioManager sets the radio manager for channel index resolution.
func (o *Orchestrator) SetRadioManager(radioManager RadioManager) {
	o.radioManager = radioManager
}

// resolveChannelIndex resolves a channel index to frequency via radio manager or Silvus band plan.
func (o *Orchestrator) resolveChannelIndex(ctx context.Context, radioID string, channelIndex int, radioManager RadioManager) (float64, error) {
	// First, try to resolve using Silvus band plan if available
	if o.config != nil && o.config.SilvusBandPlan != nil {
		// Try to get model and band from radio manager
		model, band, err := o.getRadioModelAndBand(ctx, radioID, radioManager)
		if err == nil {
			frequency, err := o.config.SilvusBandPlan.GetSilvusChannelFrequency(model, band, channelIndex)
			if err == nil {
				return frequency, nil
			}
			// If Silvus band plan doesn't have the channel, fall back to radio manager
		}
	}

	// Fall back to radio manager resolution
	return o.resolveChannelIndexFromRadioManager(ctx, radioID, channelIndex, radioManager)
}

// getRadioModelAndBand extracts model and band information from radio manager.
func (o *Orchestrator) getRadioModelAndBand(ctx context.Context, radioID string, radioManager RadioManager) (string, string, error) {
	// Use the provided radio manager or fall back to the orchestrator's radio manager
	var manager RadioManager
	if radioManager != nil {
		manager = radioManager
	} else {
		manager = o.radioManager
	}

	if manager == nil {
		return "", "", fmt.Errorf("no radio manager available")
	}

	// Get radio from manager
	radio, err := manager.GetRadio(radioID)
	if err != nil {
		return "", "", fmt.Errorf("radio %s not found: %w", radioID, err)
	}

	// Extract model and band from radio data
	model := radio.Model

	// Default band if not specified in radio
	band := "default"

	return model, band, nil
}

// resolveChannelIndexFromRadioManager resolves a channel index to frequency via radio manager (legacy method).
func (o *Orchestrator) resolveChannelIndexFromRadioManager(ctx context.Context, radioID string, channelIndex int, radioManager RadioManager) (float64, error) {
	// Use the provided radio manager or fall back to the orchestrator's radio manager
	var manager RadioManager
	if radioManager != nil {
		manager = radioManager
	} else {
		manager = o.radioManager
	}

	if manager == nil {
		return 0, fmt.Errorf("no radio manager available for channel index resolution")
	}

	// Get radio from manager
	radio, err := manager.GetRadio(radioID)
	if err != nil {
		return 0, fmt.Errorf("radio %s not found: %w", radioID, err)
	}

	// Extract capabilities from radio
	capabilities := radio.Capabilities
	if capabilities == nil {
		return 0, fmt.Errorf("radio %s has no capabilities", radioID)
	}

	channels := capabilities.Channels
	if len(channels) == 0 {
		return 0, fmt.Errorf("radio %s has no channels", radioID)
	}

	// Find channel with matching index
	for _, channel := range channels {
		if channel.Index == channelIndex {
			return channel.FrequencyMhz, nil
		}
	}

	return 0, &adapter.VendorError{
		Code:     adapter.ErrInvalidRange,
		Original: fmt.Errorf("channel index %d not found for radio %s", channelIndex, radioID),
		Details: map[string]interface{}{
			"radioID":           radioID,
			"requestedIndex":    channelIndex,
			"availableChannels": len(channels),
		},
	}
}
