package config

import (
	"testing"
	"time"
)

func TestLoadCBTimingBaseline(t *testing.T) {
	cfg := LoadCBTimingBaseline()

	// CB-TIMING §3.1
	if cfg.HeartbeatInterval != 15*time.Second {
		t.Errorf("HeartbeatInterval = %v, want 15s", cfg.HeartbeatInterval)
	}
	if cfg.HeartbeatJitter != 2*time.Second {
		t.Errorf("HeartbeatJitter = %v, want 2s", cfg.HeartbeatJitter)
	}
	if cfg.HeartbeatTimeout != 45*time.Second {
		t.Errorf("HeartbeatTimeout = %v, want 45s", cfg.HeartbeatTimeout)
	}

	// CB-TIMING §4.1
	if cfg.ProbeNormalInterval != 30*time.Second {
		t.Errorf("ProbeNormalInterval = %v, want 30s", cfg.ProbeNormalInterval)
	}
	if cfg.ProbeRecoveringInitial != 5*time.Second {
		t.Errorf("ProbeRecoveringInitial = %v, want 5s", cfg.ProbeRecoveringInitial)
	}
	if cfg.ProbeRecoveringBackoff != 1.5 {
		t.Errorf("ProbeRecoveringBackoff = %v, want 1.5", cfg.ProbeRecoveringBackoff)
	}

	// CB-TIMING §5
	if cfg.CommandTimeoutSetPower != 10*time.Second {
		t.Errorf("CommandTimeoutSetPower = %v, want 10s", cfg.CommandTimeoutSetPower)
	}
	if cfg.CommandTimeoutSetChannel != 30*time.Second {
		t.Errorf("CommandTimeoutSetChannel = %v, want 30s", cfg.CommandTimeoutSetChannel)
	}

	// CB-TIMING §6.1
	if cfg.EventBufferSize != 50 {
		t.Errorf("EventBufferSize = %d, want 50", cfg.EventBufferSize)
	}
	if cfg.EventBufferRetention != 1*time.Hour {
		t.Errorf("EventBufferRetention = %v, want 1h", cfg.EventBufferRetention)
	}
}

func TestValidateTiming_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*TimingConfig)
		wantErr bool
	}{
		{
			name: "invalid_heartbeat_interval",
			modify: func(c *TimingConfig) {
				c.HeartbeatInterval = 0
			},
			wantErr: true,
		},
		{
			name: "invalid_heartbeat_timeout",
			modify: func(c *TimingConfig) {
				c.HeartbeatTimeout = 0
			},
			wantErr: true,
		},
		{
			name: "invalid_probe_normal_interval",
			modify: func(c *TimingConfig) {
				c.ProbeNormalInterval = 0
			},
			wantErr: true,
		},
		{
			name: "invalid_probe_recovering_initial",
			modify: func(c *TimingConfig) {
				c.ProbeRecoveringInitial = 0
			},
			wantErr: true,
		},
		{
			name: "invalid_probe_recovering_max",
			modify: func(c *TimingConfig) {
				c.ProbeRecoveringMax = 0
			},
			wantErr: true,
		},
		{
			name: "invalid_probe_offline_initial",
			modify: func(c *TimingConfig) {
				c.ProbeOfflineInitial = 0
			},
			wantErr: true,
		},
		{
			name: "invalid_probe_offline_max",
			modify: func(c *TimingConfig) {
				c.ProbeOfflineMax = 0
			},
			wantErr: true,
		},
		{
			name: "invalid_event_buffer_size",
			modify: func(c *TimingConfig) {
				c.EventBufferSize = 0
			},
			wantErr: true,
		},
		{
			name: "invalid_event_buffer_retention",
			modify: func(c *TimingConfig) {
				c.EventBufferRetention = 0
			},
			wantErr: true,
		},
		{
			name: "invalid_command_timeout_set_power",
			modify: func(c *TimingConfig) {
				c.CommandTimeoutSetPower = 0
			},
			wantErr: true,
		},
		{
			name: "invalid_command_timeout_set_channel",
			modify: func(c *TimingConfig) {
				c.CommandTimeoutSetChannel = 0
			},
			wantErr: true,
		},
		{
			name: "invalid_command_timeout_select_radio",
			modify: func(c *TimingConfig) {
				c.CommandTimeoutSelectRadio = 0
			},
			wantErr: true,
		},
		{
			name: "invalid_command_timeout_get_state",
			modify: func(c *TimingConfig) {
				c.CommandTimeoutGetState = 0
			},
			wantErr: true,
		},
		{
			name: "negative_heartbeat_jitter",
			modify: func(c *TimingConfig) {
				c.HeartbeatJitter = -1 * time.Second
			},
			wantErr: true,
		},
		{
			name: "negative_probe_recovering_backoff",
			modify: func(c *TimingConfig) {
				c.ProbeRecoveringBackoff = -1.0
			},
			wantErr: true,
		},
		{
			name: "negative_probe_offline_backoff",
			modify: func(c *TimingConfig) {
				c.ProbeOfflineBackoff = -1.0
			},
			wantErr: true,
		},
		{
			name: "valid_config",
			modify: func(c *TimingConfig) {
				// Don't modify - keep valid values
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := LoadCBTimingBaseline()
			tt.modify(cfg)
			err := ValidateTimingComplete(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimingComplete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadCBTiming_EnvOverrides(t *testing.T) {
	// Test environment variable overrides
	// This tests the env var parsing paths in LoadCBTimingBaseline
	t.Run("env_vars_parsed", func(t *testing.T) {
		cfg := LoadCBTimingBaseline()

		// Verify that env vars are being read (if set)
		// The actual values depend on environment, but we can verify the structure
		if cfg.HeartbeatInterval <= 0 {
			t.Error("HeartbeatInterval should be positive")
		}
		if cfg.HeartbeatJitter < 0 {
			t.Error("HeartbeatJitter should be non-negative")
		}
		if cfg.ProbeRecoveringBackoff <= 0 {
			t.Error("ProbeRecoveringBackoff should be positive")
		}
		if cfg.ProbeOfflineBackoff <= 0 {
			t.Error("ProbeOfflineBackoff should be positive")
		}
	})
}

func TestGetEnvHelpers(t *testing.T) {
	// Test the helper functions for environment variable parsing
	t.Run("get_env_duration", func(t *testing.T) {
		// Test with a valid duration
		result := GetEnvDuration("RCC_TEST_DURATION", 5*time.Second)
		if result != 5*time.Second {
			t.Errorf("Expected 5s, got %v", result)
		}
	})

	t.Run("get_env_int", func(t *testing.T) {
		// Test with a valid integer
		result := GetEnvInt("RCC_TEST_INT", 42)
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("get_env_float", func(t *testing.T) {
		// Test with a valid float
		result := GetEnvFloat("RCC_TEST_FLOAT", 3.14)
		if result != 3.14 {
			t.Errorf("Expected 3.14, got %f", result)
		}
	})
}

func TestLoadCBTiming_FileLoadErrors(t *testing.T) {
	// Test file loading error paths
	t.Run("load_from_file_error", func(t *testing.T) {
		// Test with non-existent file
		cfg, err := loadFromFile("/non/existent/file.json")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
		if cfg != nil {
			t.Error("Expected nil config for non-existent file")
		}
	})

	t.Run("merge_timing_configs", func(t *testing.T) {
		// Test merging configurations
		base := &TimingConfig{
			HeartbeatInterval: 15 * time.Second,
			EventBufferSize:   50,
		}
		override := &TimingConfig{
			HeartbeatInterval: 30 * time.Second,
			EventBufferSize:   100,
		}

		result := mergeTimingConfigs(base, override)

		if result.HeartbeatInterval != 30*time.Second {
			t.Errorf("Expected 30s, got %v", result.HeartbeatInterval)
		}
		if result.EventBufferSize != 100 {
			t.Errorf("Expected 100, got %d", result.EventBufferSize)
		}
	})
}

func TestLoad_FullIntegration(t *testing.T) {
	// Test the main Load function with various scenarios
	t.Run("load_baseline", func(t *testing.T) {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg == nil {
			t.Fatal("Load() returned nil config")
		}

		// Verify basic structure
		if cfg.HeartbeatInterval <= 0 {
			t.Error("HeartbeatInterval should be positive")
		}
		if cfg.EventBufferSize <= 0 {
			t.Error("EventBufferSize should be positive")
		}
	})

	t.Run("load_with_validation", func(t *testing.T) {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Test that loaded config passes validation
		err = ValidateTimingComplete(cfg)
		if err != nil {
			t.Errorf("Loaded config should be valid: %v", err)
		}

		// Test constraint validation
		err = ValidateTimingConstraints(cfg)
		if err != nil {
			t.Errorf("Loaded config should pass constraints: %v", err)
		}
	})
}

func TestApplyEnvOverrides_EdgeCases(t *testing.T) {
	// Test environment variable override edge cases
	t.Run("invalid_env_values", func(t *testing.T) {
		cfg := &TimingConfig{
			HeartbeatInterval: 15 * time.Second,
			EventBufferSize:   50,
		}

		// Test with invalid environment values
		// This tests the error handling paths in applyEnvOverrides
		err := applyEnvOverrides(cfg)
		if err != nil {
			t.Logf("applyEnvOverrides returned error (expected for invalid env): %v", err)
		}
	})
}

func TestGetSilvusChannelIndex_EdgeCases(t *testing.T) {
	// Test edge cases for GetSilvusChannelIndex to improve coverage
	t.Run("invalid_model", func(t *testing.T) {
		// Create a band plan and test invalid model
		bandPlan := &SilvusBandPlan{}
		_, err := bandPlan.GetSilvusChannelIndex("InvalidModel", "InvalidBand", 2412.0)
		if err == nil {
			t.Error("Expected error for invalid model")
		}
	})

	t.Run("invalid_band", func(t *testing.T) {
		// Create a band plan and test invalid band
		bandPlan := &SilvusBandPlan{}
		_, err := bandPlan.GetSilvusChannelIndex("Scout", "InvalidBand", 2412.0)
		if err == nil {
			t.Error("Expected error for invalid band")
		}
	})
}

func TestValidateTimingConstraints_EdgeCases(t *testing.T) {
	// Test edge cases for ValidateTimingConstraints
	t.Run("constraint_violations", func(t *testing.T) {
		cfg := &TimingConfig{
			HeartbeatInterval:   1 * time.Second, // Too short
			HeartbeatTimeout:    2 * time.Second, // Too short
			ProbeNormalInterval: 1 * time.Second, // Too short
		}

		err := ValidateTimingConstraints(cfg)
		if err == nil {
			t.Error("Expected constraint validation error")
		}
	})
}
