//
//
package config

import (
	"fmt"
	"time"
)

// ValidateTiming enforces CB-TIMING v0.3 validation rules.
func ValidateTiming(config *TimingConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate heartbeat configuration
	if err := validateHeartbeat(config); err != nil {
		return fmt.Errorf("heartbeat validation failed: %w", err)
	}

	// Validate probe configuration
	if err := validateProbes(config); err != nil {
		return fmt.Errorf("probe validation failed: %w", err)
	}

	// Validate command timeouts
	if err := validateCommandTimeouts(config); err != nil {
		return fmt.Errorf("command timeout validation failed: %w", err)
	}

	// Validate event buffer configuration
	if err := validateEventBuffer(config); err != nil {
		return fmt.Errorf("event buffer validation failed: %w", err)
	}

	return nil
}

// validateHeartbeat validates heartbeat timing parameters.
func validateHeartbeat(config *TimingConfig) error {
	// Heartbeat interval must be positive
	if config.HeartbeatInterval <= 0 {
		return fmt.Errorf("heartbeat interval must be positive, got %v", config.HeartbeatInterval)
	}

	// Heartbeat jitter must be positive and ≤ 50% of interval
	maxJitter := config.HeartbeatInterval / 2
	if config.HeartbeatJitter < 0 {
		return fmt.Errorf("heartbeat jitter must be non-negative, got %v", config.HeartbeatJitter)
	}
	if config.HeartbeatJitter > maxJitter {
		return fmt.Errorf("heartbeat jitter %v exceeds 50%% of interval %v", config.HeartbeatJitter, config.HeartbeatInterval)
	}

	// Heartbeat timeout must be ≥ interval
	if config.HeartbeatTimeout < config.HeartbeatInterval {
		return fmt.Errorf("heartbeat timeout %v must be >= interval %v", config.HeartbeatTimeout, config.HeartbeatInterval)
	}

	return nil
}

// validateProbes validates probe timing parameters.
func validateProbes(config *TimingConfig) error {
	// Normal probe interval must be positive
	if config.ProbeNormalInterval <= 0 {
		return fmt.Errorf("probe normal interval must be positive, got %v", config.ProbeNormalInterval)
	}

	// Recovering probe configuration
	if config.ProbeRecoveringInitial <= 0 {
		return fmt.Errorf("probe recovering initial must be positive, got %v", config.ProbeRecoveringInitial)
	}
	if config.ProbeRecoveringBackoff < 1.0 {
		return fmt.Errorf("probe recovering backoff must be >= 1.0, got %v", config.ProbeRecoveringBackoff)
	}
	if config.ProbeRecoveringMax < config.ProbeRecoveringInitial {
		return fmt.Errorf("probe recovering max %v must be >= initial %v", config.ProbeRecoveringMax, config.ProbeRecoveringInitial)
	}

	// Offline probe configuration
	if config.ProbeOfflineInitial <= 0 {
		return fmt.Errorf("probe offline initial must be positive, got %v", config.ProbeOfflineInitial)
	}
	if config.ProbeOfflineBackoff < 1.0 {
		return fmt.Errorf("probe offline backoff must be >= 1.0, got %v", config.ProbeOfflineBackoff)
	}
	if config.ProbeOfflineMax < config.ProbeOfflineInitial {
		return fmt.Errorf("probe offline max %v must be >= initial %v", config.ProbeOfflineMax, config.ProbeOfflineInitial)
	}

	return nil
}

// validateCommandTimeouts validates command timeout parameters.
func validateCommandTimeouts(config *TimingConfig) error {
	// All command timeouts must be positive
	if config.CommandTimeoutSetPower <= 0 {
		return fmt.Errorf("command timeout setPower must be positive, got %v", config.CommandTimeoutSetPower)
	}
	if config.CommandTimeoutSetChannel <= 0 {
		return fmt.Errorf("command timeout setChannel must be positive, got %v", config.CommandTimeoutSetChannel)
	}
	if config.CommandTimeoutSelectRadio <= 0 {
		return fmt.Errorf("command timeout selectRadio must be positive, got %v", config.CommandTimeoutSelectRadio)
	}
	if config.CommandTimeoutGetState <= 0 {
		return fmt.Errorf("command timeout getState must be positive, got %v", config.CommandTimeoutGetState)
	}

	return nil
}

// validateEventBuffer validates event buffer parameters.
func validateEventBuffer(config *TimingConfig) error {
	// Event buffer size must be positive
	if config.EventBufferSize <= 0 {
		return fmt.Errorf("event buffer size must be positive, got %d", config.EventBufferSize)
	}

	// Event buffer retention must be positive
	if config.EventBufferRetention <= 0 {
		return fmt.Errorf("event buffer retention must be positive, got %v", config.EventBufferRetention)
	}

	return nil
}

// ValidateTimingConstraints validates additional timing constraints.
func ValidateTimingConstraints(config *TimingConfig) error {
	// Check that backoff factors are reasonable (not too aggressive)
	if config.ProbeRecoveringBackoff > 10.0 {
		return fmt.Errorf("probe recovering backoff %v is too aggressive (max 10.0)", config.ProbeRecoveringBackoff)
	}
	if config.ProbeOfflineBackoff > 10.0 {
		return fmt.Errorf("probe offline backoff %v is too aggressive (max 10.0)", config.ProbeOfflineBackoff)
	}

	// Check that timeouts are reasonable (not too short or too long)
	minTimeout := 100 * time.Millisecond
	maxTimeout := 5 * time.Minute

	if config.CommandTimeoutSetPower < minTimeout || config.CommandTimeoutSetPower > maxTimeout {
		return fmt.Errorf("command timeout setPower %v is outside reasonable range [%v, %v]",
			config.CommandTimeoutSetPower, minTimeout, maxTimeout)
	}

	if config.CommandTimeoutSetChannel < minTimeout || config.CommandTimeoutSetChannel > maxTimeout {
		return fmt.Errorf("command timeout setChannel %v is outside reasonable range [%v, %v]",
			config.CommandTimeoutSetChannel, minTimeout, maxTimeout)
	}

	if config.CommandTimeoutSelectRadio < minTimeout || config.CommandTimeoutSelectRadio > maxTimeout {
		return fmt.Errorf("command timeout selectRadio %v is outside reasonable range [%v, %v]",
			config.CommandTimeoutSelectRadio, minTimeout, maxTimeout)
	}

	if config.CommandTimeoutGetState < minTimeout || config.CommandTimeoutGetState > maxTimeout {
		return fmt.Errorf("command timeout getState %v is outside reasonable range [%v, %v]",
			config.CommandTimeoutGetState, minTimeout, maxTimeout)
	}

	return nil
}

// ValidateTimingComplete performs complete timing validation including constraints.
func ValidateTimingComplete(config *TimingConfig) error {
	// Basic validation
	if err := ValidateTiming(config); err != nil {
		return err
	}

	// Additional constraints
	if err := ValidateTimingConstraints(config); err != nil {
		return err
	}

	return nil
}
