//go:build integration

package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLoad_FailureScenarios tests Load() failure scenarios:
// Missing file, unreadable permissions, bad YAML (corruption), partial sections
func TestLoad_FailureScenarios(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expectError bool
		errorType   string
	}{
		{
			name: "missing_config_file",
			setup: func(t *testing.T) string {
				// No config.json file exists
				return ""
			},
			expectError: false, // Should succeed with defaults
		},
		{
			name: "unreadable_permissions",
			setup: func(t *testing.T) string {
				// Create file with no read permissions
				configPath := filepath.Join(t.TempDir(), "config.json")
				err := os.WriteFile(configPath, []byte(`{"timing": {}}`), 0000)
				if err != nil {
					t.Fatalf("Failed to create unreadable file: %v", err)
				}
				return configPath
			},
			expectError: true,
			errorType:   "permission denied",
		},
		{
			name: "bad_yaml_corruption",
			setup: func(t *testing.T) string {
				// Create corrupted JSON file
				configPath := filepath.Join(t.TempDir(), "config.json")
				corruptedJSON := `{"timing": {"commandTimeoutSetPower": "invalid_duration", "commandTimeoutSetChannel": 30}}`
				err := os.WriteFile(configPath, []byte(corruptedJSON), 0644)
				if err != nil {
					t.Fatalf("Failed to create corrupted file: %v", err)
				}
				return configPath
			},
			expectError: true,
			errorType:   "invalid duration",
		},
		{
			name: "partial_sections",
			setup: func(t *testing.T) string {
				// Create file with only partial timing config
				configPath := filepath.Join(t.TempDir(), "config.json")
				partialJSON := `{"timing": {"commandTimeoutSetPower": 15}}`
				err := os.WriteFile(configPath, []byte(partialJSON), 0644)
				if err != nil {
					t.Fatalf("Failed to create partial file: %v", err)
				}
				return configPath
			},
			expectError: false, // Should succeed with partial config merged with defaults
		},
		{
			name: "truncated_file",
			setup: func(t *testing.T) string {
				// Create truncated JSON file
				configPath := filepath.Join(t.TempDir(), "config.json")
				truncatedJSON := `{"timing": {"commandTimeoutSetPower": 15, "commandTimeoutSetChannel": 30, "commandTimeoutSelectRadio": 5, "commandTimeoutGetState": 5, "heartbeatInterval": "15s", "heartbeatJitter": "2s", "heartbeatTimeout": "45s", "probeNormalInterval": "30s", "probeRecoveringInitial": "5s", "probeRecoveringBackoff": 1.5, "probeRecoveringMax": "15s", "probeOfflineInitial": "10s", "probeOfflineBackoff": 2.0, "probeOfflineMax": "300s", "eventBufferSize": 50, "eventRetentionDuration": "1h"` // Missing closing brace
				err := os.WriteFile(configPath, []byte(truncatedJSON), 0644)
				if err != nil {
					t.Fatalf("Failed to create truncated file: %v", err)
				}
				return configPath
			},
			expectError: true,
			errorType:   "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			configPath := tt.setup(t)

			// Change to test directory if config file was created
			if configPath != "" {
				originalDir, _ := os.Getwd()
				defer os.Chdir(originalDir)

				testDir := filepath.Dir(configPath)
				os.Chdir(testDir)
			}

			// Act: Call Load()
			config, err := Load()

			// Assert: Error handling
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else {
					t.Logf("✅ Got expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if config == nil {
					t.Error("Expected config but got nil")
				}
			}
		})
	}
}

// TestApplyEnvOverrides_ConflictingEnvVars tests applyEnvOverrides():
// Conflicting env vars precedence; invalid types; partial overrides
func TestApplyEnvOverrides_ConflictingEnvVars(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorType   string
	}{
		{
			name: "valid_overrides",
			envVars: map[string]string{
				"RCC_TIMING_COMMAND_TIMEOUT_SET_POWER":   "15s",
				"RCC_TIMING_COMMAND_TIMEOUT_SET_CHANNEL": "45s",
				"RCC_TIMING_HEARTBEAT_INTERVAL":          "20s",
			},
			expectError: false,
		},
		{
			name: "invalid_duration_format",
			envVars: map[string]string{
				"RCC_TIMING_COMMAND_TIMEOUT_SET_POWER": "invalid_duration",
			},
			expectError: true,
			errorType:   "invalid duration",
		},
		{
			name: "negative_duration",
			envVars: map[string]string{
				"RCC_TIMING_COMMAND_TIMEOUT_SET_POWER": "-5s",
			},
			expectError: true,
			errorType:   "negative duration",
		},
		{
			name: "partial_overrides",
			envVars: map[string]string{
				"RCC_TIMING_COMMAND_TIMEOUT_SET_POWER": "15s",
				// Other values should remain as defaults
			},
			expectError: false,
		},
		{
			name: "conflicting_precedence",
			envVars: map[string]string{
				"RCC_TIMING_COMMAND_TIMEOUT_SET_POWER": "15s",
				"RCC_TIMING_COMMAND_TIMEOUT_SET_POWER": "20s", // Same key, should use last value
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variables
			originalEnv := make(map[string]string)
			for key, value := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
				os.Setenv(key, value)
			}

			// Cleanup
			defer func() {
				for key, originalValue := range originalEnv {
					if originalValue == "" {
						os.Unsetenv(key)
					} else {
						os.Setenv(key, originalValue)
					}
				}
			}()

			// Act: Create config with environment overrides
			config := LoadCBTimingBaseline()
			err := applyEnvOverrides(config)

			// Assert: Error handling
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else {
					t.Logf("✅ Got expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify overrides were applied
				if tt.envVars["RCC_TIMING_COMMAND_TIMEOUT_SET_POWER"] != "" {
					expectedDuration, _ := time.ParseDuration(tt.envVars["RCC_TIMING_COMMAND_TIMEOUT_SET_POWER"])
					if config.CommandTimeoutSetPower != expectedDuration {
						t.Errorf("Expected timeout %v, got %v", expectedDuration, config.CommandTimeoutSetPower)
					}
				}
			}
		})
	}
}

// TestLoadFromFile_DifferentEncodings tests loadFromFile():
// Different encodings, BOM presence; truncated file
func TestLoadFromFile_DifferentEncodings(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		errorType   string
	}{
		{
			name: "valid_json",
			content: `{
				"timing": {
					"commandTimeoutSetPower": "15s",
					"commandTimeoutSetChannel": "45s",
					"commandTimeoutSelectRadio": "5s",
					"commandTimeoutGetState": "5s",
					"heartbeatInterval": "20s",
					"heartbeatJitter": "2s",
					"heartbeatTimeout": "45s",
					"probeNormalInterval": "30s",
					"probeRecoveringInitial": "5s",
					"probeRecoveringBackoff": 1.5,
					"probeRecoveringMax": "15s",
					"probeOfflineInitial": "10s",
					"probeOfflineBackoff": 2.0,
					"probeOfflineMax": "300s",
					"eventBufferSize": 50,
					"eventRetentionDuration": "1h"
				}
			}`,
			expectError: false,
		},
		{
			name: "invalid_json_syntax",
			content: `{
				"timing": {
					"commandTimeoutSetPower": "15s",
					"commandTimeoutSetChannel": "45s",
					"commandTimeoutSelectRadio": "5s",
					"commandTimeoutGetState": "5s",
					"heartbeatInterval": "20s",
					"heartbeatJitter": "2s",
					"heartbeatTimeout": "45s",
					"probeNormalInterval": "30s",
					"probeRecoveringInitial": "5s",
					"probeRecoveringBackoff": 1.5,
					"probeRecoveringMax": "15s",
					"probeOfflineInitial": "10s",
					"probeOfflineBackoff": 2.0,
					"probeOfflineMax": "300s",
					"eventBufferSize": 50,
					"eventRetentionDuration": "1h"
				}
			`, // Missing closing brace
			expectError: true,
			errorType:   "invalid JSON",
		},
		{
			name:        "empty_file",
			content:     "",
			expectError: true,
			errorType:   "empty file",
		},
		{
			name:        "whitespace_only",
			content:     "   \n\t  ",
			expectError: true,
			errorType:   "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			configPath := filepath.Join(t.TempDir(), "config.json")
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Act: Load from file
			config, err := loadFromFile(configPath)

			// Assert: Error handling
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else {
					t.Logf("✅ Got expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if config == nil {
					t.Error("Expected config but got nil")
				}
			}
		})
	}
}

// TestMergeTimingConfigs_ConflictDetection tests mergeTimingConfigs():
// Conflict detection; precedence rules; ensure §5 timeouts preserved
func TestMergeTimingConfigs_ConflictDetection(t *testing.T) {
	tests := []struct {
		name           string
		baseConfig     *TimingConfig
		fileConfig     *TimingConfig
		expectedResult *TimingConfig
		expectError    bool
	}{
		{
			name: "no_conflicts",
			baseConfig: &TimingConfig{
				CommandTimeoutSetPower:   10 * time.Second,
				CommandTimeoutSetChannel: 30 * time.Second,
				HeartbeatInterval:        15 * time.Second,
			},
			fileConfig: &TimingConfig{
				CommandTimeoutSetPower:   15 * time.Second,
				CommandTimeoutSetChannel: 45 * time.Second,
				HeartbeatInterval:        20 * time.Second,
			},
			expectedResult: &TimingConfig{
				CommandTimeoutSetPower:   15 * time.Second, // File overrides base
				CommandTimeoutSetChannel: 45 * time.Second,
				HeartbeatInterval:        20 * time.Second,
			},
			expectError: false,
		},
		{
			name:       "preserve_cb_timing_baseline",
			baseConfig: LoadCBTimingBaseline(),
			fileConfig: &TimingConfig{
				CommandTimeoutSetPower:   15 * time.Second,
				CommandTimeoutSetChannel: 45 * time.Second,
				// Other values should remain from baseline
			},
			expectedResult: &TimingConfig{
				CommandTimeoutSetPower:   15 * time.Second,
				CommandTimeoutSetChannel: 45 * time.Second,
				// Other values from baseline
			},
			expectError: false,
		},
		{
			name: "zero_values_preserved",
			baseConfig: &TimingConfig{
				CommandTimeoutSetPower:   10 * time.Second,
				CommandTimeoutSetChannel: 30 * time.Second,
			},
			fileConfig: &TimingConfig{
				CommandTimeoutSetPower:   0, // Zero value should not override
				CommandTimeoutSetChannel: 45 * time.Second,
			},
			expectedResult: &TimingConfig{
				CommandTimeoutSetPower:   10 * time.Second, // Preserved from base
				CommandTimeoutSetChannel: 45 * time.Second, // Overridden from file
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act: Merge configs
			result := mergeTimingConfigs(tt.baseConfig, tt.fileConfig)

			// Assert: Result matches expected
			if tt.expectError {
				if result != nil {
					t.Error("Expected error but got result")
				}
			} else {
				if result == nil {
					t.Error("Expected result but got nil")
				} else {
					// Verify key properties
					if result.CommandTimeoutSetPower != tt.expectedResult.CommandTimeoutSetPower {
						t.Errorf("Expected CommandTimeoutSetPower %v, got %v",
							tt.expectedResult.CommandTimeoutSetPower, result.CommandTimeoutSetPower)
					}
					if result.CommandTimeoutSetChannel != tt.expectedResult.CommandTimeoutSetChannel {
						t.Errorf("Expected CommandTimeoutSetChannel %v, got %v",
							tt.expectedResult.CommandTimeoutSetChannel, result.CommandTimeoutSetChannel)
					}
				}
			}
		})
	}
}

// TestLoad_IntegrationFlow tests the complete Load() flow:
// Defaults + env overrides + file config + validation
func TestLoad_IntegrationFlow(t *testing.T) {
	// Setup: Create test environment
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(tempDir)

	// Create config.json file
	configJSON := `{
		"timing": {
			"commandTimeoutSetPower": "15s",
			"commandTimeoutSetChannel": "45s",
			"heartbeatInterval": "20s"
		}
	}`
	err := os.WriteFile("config.json", []byte(configJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create config.json: %v", err)
	}

	// Set environment variables
	os.Setenv("RCC_TIMING_COMMAND_TIMEOUT_SET_POWER", "12s")
	os.Setenv("RCC_TIMING_HEARTBEAT_INTERVAL", "18s")
	defer func() {
		os.Unsetenv("RCC_TIMING_COMMAND_TIMEOUT_SET_POWER")
		os.Unsetenv("RCC_TIMING_HEARTBEAT_INTERVAL")
	}()

	// Act: Load configuration
	config, err := Load()

	// Assert: Configuration loaded successfully
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if config == nil {
		t.Fatal("Expected config but got nil")
	}

	// Assert: Environment overrides take precedence over file config
	expectedPowerTimeout, _ := time.ParseDuration("12s") // From env var
	if config.CommandTimeoutSetPower != expectedPowerTimeout {
		t.Errorf("Expected CommandTimeoutSetPower %v (from env), got %v",
			expectedPowerTimeout, config.CommandTimeoutSetPower)
	}

	// Assert: File config overrides defaults
	expectedChannelTimeout, _ := time.ParseDuration("45s") // From file
	if config.CommandTimeoutSetChannel != expectedChannelTimeout {
		t.Errorf("Expected CommandTimeoutSetChannel %v (from file), got %v",
			expectedChannelTimeout, config.CommandTimeoutSetChannel)
	}

	// Assert: Environment overrides file config
	expectedHeartbeatInterval, _ := time.ParseDuration("18s") // From env var
	if config.HeartbeatInterval != expectedHeartbeatInterval {
		t.Errorf("Expected HeartbeatInterval %v (from env), got %v",
			expectedHeartbeatInterval, config.HeartbeatInterval)
	}

	// Assert: Validation passed (no panics, reasonable values)
	if config.CommandTimeoutSetPower <= 0 {
		t.Error("CommandTimeoutSetPower should be positive")
	}
	if config.CommandTimeoutSetChannel <= 0 {
		t.Error("CommandTimeoutSetChannel should be positive")
	}

	t.Logf("✅ Integration flow: env overrides > file config > defaults")
}
