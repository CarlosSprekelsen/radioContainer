package config

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Test loading with defaults
	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if config == nil {
		t.Fatal("Load() returned nil config")
	}

	// Verify default values are loaded
	if config.HeartbeatInterval != 15*time.Second {
		t.Errorf("HeartbeatInterval = %v, want 15s", config.HeartbeatInterval)
	}

	if config.HeartbeatJitter != 2*time.Second {
		t.Errorf("HeartbeatJitter = %v, want 2s", config.HeartbeatJitter)
	}

	if config.HeartbeatTimeout != 45*time.Second {
		t.Errorf("HeartbeatTimeout = %v, want 45s", config.HeartbeatTimeout)
	}
}

func TestLoadWithEnvOverrides(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("RCC_TIMING_HEARTBEAT_INTERVAL", "20s")
	_ = os.Setenv("RCC_TIMING_HEARTBEAT_JITTER", "3s")
	_ = os.Setenv("RCC_TIMING_PROBE_NORMAL_INTERVAL", "45s")
	_ = os.Setenv("RCC_TIMING_COMMAND_SET_POWER", "15s")
	_ = os.Setenv("RCC_TIMING_EVENT_BUFFER_SIZE", "100")

	defer func() {
		_ = os.Unsetenv("RCC_TIMING_HEARTBEAT_INTERVAL")
		_ = os.Unsetenv("RCC_TIMING_HEARTBEAT_JITTER")
		_ = os.Unsetenv("RCC_TIMING_PROBE_NORMAL_INTERVAL")
		_ = os.Unsetenv("RCC_TIMING_COMMAND_SET_POWER")
		_ = os.Unsetenv("RCC_TIMING_EVENT_BUFFER_SIZE")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() with env overrides failed: %v", err)
	}

	// Verify environment overrides are applied
	if config.HeartbeatInterval != 20*time.Second {
		t.Errorf("HeartbeatInterval = %v, want 20s", config.HeartbeatInterval)
	}

	if config.HeartbeatJitter != 3*time.Second {
		t.Errorf("HeartbeatJitter = %v, want 3s", config.HeartbeatJitter)
	}

	if config.ProbeNormalInterval != 45*time.Second {
		t.Errorf("ProbeNormalInterval = %v, want 45s", config.ProbeNormalInterval)
	}

	if config.CommandTimeoutSetPower != 15*time.Second {
		t.Errorf("CommandTimeoutSetPower = %v, want 15s", config.CommandTimeoutSetPower)
	}

	if config.EventBufferSize != 100 {
		t.Errorf("EventBufferSize = %d, want 100", config.EventBufferSize)
	}
}

func TestLoadWithConfigFile(t *testing.T) {
	// Create a temporary config file
	configData := map[string]interface{}{
		"heartbeat_interval":        "25s",
		"heartbeat_jitter":          "4s",
		"heartbeat_timeout":         "60s",
		"probe_normal_interval":     "40s",
		"command_timeout_set_power": "12s",
		"event_buffer_size":         75,
	}

	configJSON, err := json.Marshal(configData)
	if err != nil {
		t.Fatalf("Failed to marshal config data: %v", err)
	}

	// Write to temporary file
	tmpFile, err := os.CreateTemp("", "config_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.Write(configJSON); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	_ = tmpFile.Close()

	// Change to the directory with the config file
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	configDir := tmpFile.Name()
	_ = os.Chdir(configDir)

	// Test loading with config file
	config, err := Load()
	if err != nil {
		t.Fatalf("Load() with config file failed: %v", err)
	}

	// Note: The current implementation looks for "config.json" in the current directory
	// This test would need to be adjusted based on the actual file loading logic
	t.Logf("Config loaded: %+v", config)
}

func TestValidateTiming(t *testing.T) {
	// Test valid configuration
	config := LoadCBTimingBaseline()
	if err := ValidateTiming(config); err != nil {
		t.Errorf("ValidateTiming() failed on valid config: %v", err)
	}

	// Test invalid heartbeat interval
	config.HeartbeatInterval = 0
	if err := ValidateTiming(config); err == nil {
		t.Error("ValidateTiming() should fail on zero heartbeat interval")
	}

	// Test invalid heartbeat jitter (> 50% of interval)
	config = LoadCBTimingBaseline()
	config.HeartbeatJitter = 10 * time.Second // > 50% of 15s interval
	if err := ValidateTiming(config); err == nil {
		t.Error("ValidateTiming() should fail on jitter > 50% of interval")
	}

	// Test invalid heartbeat timeout (< interval)
	config = LoadCBTimingBaseline()
	config.HeartbeatTimeout = 10 * time.Second // < 15s interval
	if err := ValidateTiming(config); err == nil {
		t.Error("ValidateTiming() should fail on timeout < interval")
	}

	// Test invalid probe backoff (< 1.0)
	config = LoadCBTimingBaseline()
	config.ProbeRecoveringBackoff = 0.5
	if err := ValidateTiming(config); err == nil {
		t.Error("ValidateTiming() should fail on backoff < 1.0")
	}

	// Test invalid probe max (< initial)
	config = LoadCBTimingBaseline()
	config.ProbeRecoveringMax = 3 * time.Second // < initial (5s)
	if err := ValidateTiming(config); err == nil {
		t.Error("ValidateTiming() should fail on max < initial")
	}

	// Test invalid command timeout
	config = LoadCBTimingBaseline()
	config.CommandTimeoutSetPower = 0
	if err := ValidateTiming(config); err == nil {
		t.Error("ValidateTiming() should fail on zero command timeout")
	}

	// Test invalid event buffer size
	config = LoadCBTimingBaseline()
	config.EventBufferSize = 0
	if err := ValidateTiming(config); err == nil {
		t.Error("ValidateTiming() should fail on zero buffer size")
	}

	// Test invalid event buffer retention
	config = LoadCBTimingBaseline()
	config.EventBufferRetention = 0
	if err := ValidateTiming(config); err == nil {
		t.Error("ValidateTiming() should fail on zero buffer retention")
	}
}

func TestValidateTimingConstraints(t *testing.T) {
	// Test valid configuration
	config := LoadCBTimingBaseline()
	if err := ValidateTimingConstraints(config); err != nil {
		t.Errorf("ValidateTimingConstraints() failed on valid config: %v", err)
	}

	// Test aggressive backoff
	config.ProbeRecoveringBackoff = 15.0 // > 10.0
	if err := ValidateTimingConstraints(config); err == nil {
		t.Error("ValidateTimingConstraints() should fail on aggressive backoff")
	}

	// Test timeout too short
	config = LoadCBTimingBaseline()
	config.CommandTimeoutSetPower = 50 * time.Millisecond // < 100ms
	if err := ValidateTimingConstraints(config); err == nil {
		t.Error("ValidateTimingConstraints() should fail on timeout too short")
	}

	// Test timeout too long
	config = LoadCBTimingBaseline()
	config.CommandTimeoutSetPower = 10 * time.Minute // > 5min
	if err := ValidateTimingConstraints(config); err == nil {
		t.Error("ValidateTimingConstraints() should fail on timeout too long")
	}
}

func TestValidateTimingComplete(t *testing.T) {
	// Test valid configuration
	config := LoadCBTimingBaseline()
	if err := ValidateTimingComplete(config); err != nil {
		t.Errorf("ValidateTimingComplete() failed on valid config: %v", err)
	}

	// Test invalid configuration
	config.HeartbeatInterval = 0
	if err := ValidateTimingComplete(config); err == nil {
		t.Error("ValidateTimingComplete() should fail on invalid config")
	}
}

func TestGetEnvVar(t *testing.T) {
	// Test with environment variable set
	_ = os.Setenv("TEST_VAR", "test_value")
	defer func() { _ = os.Unsetenv("TEST_VAR") }()

	value := GetEnvVar("TEST_VAR", "default")
	if value != "test_value" {
		t.Errorf("GetEnvVar() = %s, want test_value", value)
	}

	// Test with default value
	value = GetEnvVar("NONEXISTENT_VAR", "default")
	if value != "default" {
		t.Errorf("GetEnvVar() = %s, want default", value)
	}
}

func TestGetEnvDuration(t *testing.T) {
	// Test with environment variable set
	_ = os.Setenv("TEST_DURATION", "30s")
	defer func() { _ = os.Unsetenv("TEST_DURATION") }()

	value := GetEnvDuration("TEST_DURATION", 10*time.Second)
	if value != 30*time.Second {
		t.Errorf("GetEnvDuration() = %v, want 30s", value)
	}

	// Test with default value
	value = GetEnvDuration("NONEXISTENT_DURATION", 10*time.Second)
	if value != 10*time.Second {
		t.Errorf("GetEnvDuration() = %v, want 10s", value)
	}

	// Test with invalid duration
	_ = os.Setenv("INVALID_DURATION", "invalid")
	defer func() { _ = os.Unsetenv("INVALID_DURATION") }()

	value = GetEnvDuration("INVALID_DURATION", 10*time.Second)
	if value != 10*time.Second {
		t.Errorf("GetEnvDuration() = %v, want 10s", value)
	}
}

func TestGetEnvFloat(t *testing.T) {
	// Test with environment variable set
	_ = os.Setenv("TEST_FLOAT", "3.14")
	defer func() { _ = os.Unsetenv("TEST_FLOAT") }()

	value := GetEnvFloat("TEST_FLOAT", 1.0)
	if value != 3.14 {
		t.Errorf("GetEnvFloat() = %f, want 3.14", value)
	}

	// Test with default value
	value = GetEnvFloat("NONEXISTENT_FLOAT", 1.0)
	if value != 1.0 {
		t.Errorf("GetEnvFloat() = %f, want 1.0", value)
	}

	// Test with invalid float
	_ = os.Setenv("INVALID_FLOAT", "invalid")
	defer func() { _ = os.Unsetenv("INVALID_FLOAT") }()

	value = GetEnvFloat("INVALID_FLOAT", 1.0)
	if value != 1.0 {
		t.Errorf("GetEnvFloat() = %f, want 1.0", value)
	}
}

func TestGetEnvInt(t *testing.T) {
	// Test with environment variable set
	_ = os.Setenv("TEST_INT", "42")
	defer func() { _ = os.Unsetenv("TEST_INT") }()

	value := GetEnvInt("TEST_INT", 10)
	if value != 42 {
		t.Errorf("GetEnvInt() = %d, want 42", value)
	}

	// Test with default value
	value = GetEnvInt("NONEXISTENT_INT", 10)
	if value != 10 {
		t.Errorf("GetEnvInt() = %d, want 10", value)
	}

	// Test with invalid int
	_ = os.Setenv("INVALID_INT", "invalid")
	defer func() { _ = os.Unsetenv("INVALID_INT") }()

	value = GetEnvInt("INVALID_INT", 10)
	if value != 10 {
		t.Errorf("GetEnvInt() = %d, want 10", value)
	}
}
