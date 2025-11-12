//
//
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Load merges defaults from LoadCBTimingBaseline() + env overrides (RCC_TIMING_*) + optional config.json.
func Load() (*TimingConfig, error) {
	// Start with CB-TIMING v0.3 baseline
	config := LoadCBTimingBaseline()

	// Apply environment variable overrides
	if err := applyEnvOverrides(config); err != nil {
		return nil, fmt.Errorf("failed to apply environment overrides: %w", err)
	}

	// Try to load from config.json if it exists
	if _, err := os.Stat("config.json"); err == nil {
		fileConfig, err := loadFromFile("config.json")
		if err != nil {
			return nil, fmt.Errorf("failed to load config.json: %w", err)
		}

		// Merge file config with current config
		config = mergeTimingConfigs(config, fileConfig)
	}

	// Try to load Silvus band plan from silvus-band-plan.json if it exists
	if _, err := os.Stat("silvus-band-plan.json"); err == nil {
		bandPlan, err := loadSilvusBandPlanFromFile("silvus-band-plan.json")
		if err != nil {
			return nil, fmt.Errorf("failed to load silvus-band-plan.json: %w", err)
		}
		config.SilvusBandPlan = bandPlan
	}

	// Validate the final configuration
	if err := ValidateTiming(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// applyEnvOverrides applies RCC_TIMING_* environment variables to the config.
func applyEnvOverrides(config *TimingConfig) error {
	// Heartbeat configuration
	if val := os.Getenv("RCC_TIMING_HEARTBEAT_INTERVAL"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.HeartbeatInterval = duration
		}
	}

	if val := os.Getenv("RCC_TIMING_HEARTBEAT_JITTER"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.HeartbeatJitter = duration
		}
	}

	if val := os.Getenv("RCC_TIMING_HEARTBEAT_TIMEOUT"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.HeartbeatTimeout = duration
		}
	}

	// Probe configuration
	if val := os.Getenv("RCC_TIMING_PROBE_NORMAL_INTERVAL"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.ProbeNormalInterval = duration
		}
	}

	if val := os.Getenv("RCC_TIMING_PROBE_RECOVERING_INITIAL"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.ProbeRecoveringInitial = duration
		}
	}

	if val := os.Getenv("RCC_TIMING_PROBE_RECOVERING_BACKOFF"); val != "" {
		if factor, err := strconv.ParseFloat(val, 64); err == nil {
			config.ProbeRecoveringBackoff = factor
		}
	}

	if val := os.Getenv("RCC_TIMING_PROBE_RECOVERING_MAX"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.ProbeRecoveringMax = duration
		}
	}

	if val := os.Getenv("RCC_TIMING_PROBE_OFFLINE_INITIAL"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.ProbeOfflineInitial = duration
		}
	}

	if val := os.Getenv("RCC_TIMING_PROBE_OFFLINE_BACKOFF"); val != "" {
		if factor, err := strconv.ParseFloat(val, 64); err == nil {
			config.ProbeOfflineBackoff = factor
		}
	}

	if val := os.Getenv("RCC_TIMING_PROBE_OFFLINE_MAX"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.ProbeOfflineMax = duration
		}
	}

	// Command timeouts
	if val := os.Getenv("RCC_TIMING_COMMAND_SET_POWER"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.CommandTimeoutSetPower = duration
		}
	}

	if val := os.Getenv("RCC_TIMING_COMMAND_SET_CHANNEL"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.CommandTimeoutSetChannel = duration
		}
	}

	if val := os.Getenv("RCC_TIMING_COMMAND_SELECT_RADIO"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.CommandTimeoutSelectRadio = duration
		}
	}

	if val := os.Getenv("RCC_TIMING_COMMAND_GET_STATE"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.CommandTimeoutGetState = duration
		}
	}

	// Event buffer configuration
	if val := os.Getenv("RCC_TIMING_EVENT_BUFFER_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			config.EventBufferSize = size
		}
	}

	if val := os.Getenv("RCC_TIMING_EVENT_BUFFER_RETENTION"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.EventBufferRetention = duration
		}
	}

	// Load Silvus band plan from environment variable
	if val := os.Getenv("RCC_SILVUS_BAND_PLAN"); val != "" {
		bandPlan, err := loadSilvusBandPlanFromJSON(val)
		if err == nil {
			config.SilvusBandPlan = bandPlan
		}
	}

	return nil
}

// loadFromFile loads timing configuration from a JSON file.
func loadFromFile(filename string) (*TimingConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var config TimingConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// mergeTimingConfigs merges file configuration with current configuration.
// File values take precedence over current values.
func mergeTimingConfigs(current, file *TimingConfig) *TimingConfig {
	// Create a copy of current config
	merged := *current

	// Apply file overrides (only if file values are non-zero)
	if file.HeartbeatInterval != 0 {
		merged.HeartbeatInterval = file.HeartbeatInterval
	}
	if file.HeartbeatJitter != 0 {
		merged.HeartbeatJitter = file.HeartbeatJitter
	}
	if file.HeartbeatTimeout != 0 {
		merged.HeartbeatTimeout = file.HeartbeatTimeout
	}
	if file.ProbeNormalInterval != 0 {
		merged.ProbeNormalInterval = file.ProbeNormalInterval
	}
	if file.ProbeRecoveringInitial != 0 {
		merged.ProbeRecoveringInitial = file.ProbeRecoveringInitial
	}
	if file.ProbeRecoveringBackoff != 0 {
		merged.ProbeRecoveringBackoff = file.ProbeRecoveringBackoff
	}
	if file.ProbeRecoveringMax != 0 {
		merged.ProbeRecoveringMax = file.ProbeRecoveringMax
	}
	if file.ProbeOfflineInitial != 0 {
		merged.ProbeOfflineInitial = file.ProbeOfflineInitial
	}
	if file.ProbeOfflineBackoff != 0 {
		merged.ProbeOfflineBackoff = file.ProbeOfflineBackoff
	}
	if file.ProbeOfflineMax != 0 {
		merged.ProbeOfflineMax = file.ProbeOfflineMax
	}
	if file.CommandTimeoutSetPower != 0 {
		merged.CommandTimeoutSetPower = file.CommandTimeoutSetPower
	}
	if file.CommandTimeoutSetChannel != 0 {
		merged.CommandTimeoutSetChannel = file.CommandTimeoutSetChannel
	}
	if file.CommandTimeoutSelectRadio != 0 {
		merged.CommandTimeoutSelectRadio = file.CommandTimeoutSelectRadio
	}
	if file.CommandTimeoutGetState != 0 {
		merged.CommandTimeoutGetState = file.CommandTimeoutGetState
	}
	if file.EventBufferSize != 0 {
		merged.EventBufferSize = file.EventBufferSize
	}
	if file.EventBufferRetention != 0 {
		merged.EventBufferRetention = file.EventBufferRetention
	}

	return &merged
}

// GetEnvVar returns the value of an environment variable with a default.
func GetEnvVar(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvDuration returns the value of an environment variable as a duration with a default.
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetEnvFloat returns the value of an environment variable as a float64 with a default.
func GetEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

// GetEnvInt returns the value of an environment variable as an int with a default.
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// loadSilvusBandPlanFromJSON loads a Silvus band plan from JSON string.
func loadSilvusBandPlanFromJSON(jsonStr string) (*SilvusBandPlan, error) {
	var bandPlan SilvusBandPlan
	if err := json.Unmarshal([]byte(jsonStr), &bandPlan); err != nil {
		return nil, fmt.Errorf("failed to parse Silvus band plan JSON: %w", err)
	}
	return &bandPlan, nil
}

// loadSilvusBandPlanFromFile loads a Silvus band plan from a JSON file.
func loadSilvusBandPlanFromFile(filename string) (*SilvusBandPlan, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var bandPlan SilvusBandPlan
	if err := json.NewDecoder(file).Decode(&bandPlan); err != nil {
		return nil, fmt.Errorf("failed to decode Silvus band plan from %s: %w", filename, err)
	}
	return &bandPlan, nil
}
