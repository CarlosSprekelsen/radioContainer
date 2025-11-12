package config

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestLoadWithCorruptedConfigFile tests loading with corrupted config.json
func TestLoadWithCorruptedConfigFile(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create corrupted config.json
	corruptedConfig := []byte(`{
		"heartbeatInterval": "invalid-duration",
		"heartbeatJitter": "2s",
		"heartbeatTimeout": "45s"
	}`)
	if err := os.WriteFile("config.json", corruptedConfig, 0644); err != nil {
		t.Fatalf("Failed to write corrupted config: %v", err)
	}

	// Test loading should fail
	_, err := Load()
	if err == nil {
		t.Error("Expected error for corrupted config file, but got success")
	}

	// Verify error message contains relevant information
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestLoadWithInvalidJSON tests loading with invalid JSON
func TestLoadWithInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create invalid JSON
	invalidJSON := []byte(`{
		"heartbeatInterval": "15s",
		"heartbeatJitter": "2s",
		"heartbeatTimeout": "45s"
		// Missing closing brace
	}`)
	if err := os.WriteFile("config.json", invalidJSON, 0644); err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Expected error for invalid JSON, but got success")
	}
}

// TestLoadWithPartialConfig tests loading with partial config that has zero values
func TestLoadWithPartialConfig(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create partial config with only some fields
	partialConfig := map[string]interface{}{
		"heartbeatInterval": "20s",
		"heartbeatJitter":   "3s",
		// Other fields omitted (should use defaults)
	}

	configJSON, err := json.Marshal(partialConfig)
	if err != nil {
		t.Fatalf("Failed to marshal partial config: %v", err)
	}

	if err := os.WriteFile("config.json", configJSON, 0644); err != nil {
		t.Fatalf("Failed to write partial config: %v", err)
	}

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed with partial config: %v", err)
	}

	// Verify that specified values are used
	if config.HeartbeatInterval != 20*time.Second {
		t.Errorf("Expected HeartbeatInterval 20s, got %v", config.HeartbeatInterval)
	}

	if config.HeartbeatJitter != 3*time.Second {
		t.Errorf("Expected HeartbeatJitter 3s, got %v", config.HeartbeatJitter)
	}

	// Verify that defaults are used for omitted fields
	if config.HeartbeatTimeout != 45*time.Second {
		t.Errorf("Expected default HeartbeatTimeout 45s, got %v", config.HeartbeatTimeout)
	}
}

// TestLoadWithInvalidEnvDurations tests loading with invalid environment variable durations
func TestLoadWithInvalidEnvDurations(t *testing.T) {
	// Set invalid environment variables
	invalidDurations := []string{
		"RCC_TIMING_HEARTBEAT_INTERVAL",
		"RCC_TIMING_HEARTBEAT_JITTER",
		"RCC_TIMING_COMMAND_SET_POWER",
		"RCC_TIMING_COMMAND_SET_CHANNEL",
	}

	for _, envVar := range invalidDurations {
		t.Run("Invalid_"+envVar, func(t *testing.T) {
			// Set invalid duration
			os.Setenv(envVar, "invalid-duration")
			defer os.Unsetenv(envVar)

			// Load should still succeed (invalid env vars are ignored)
			config, err := Load()
			if err != nil {
				t.Fatalf("Load() failed with invalid env var %s: %v", envVar, err)
			}

			if config == nil {
				t.Fatal("Load() returned nil config")
			}
		})
	}
}

// TestLoadWithBandPlanJSONErrors tests loading with corrupted band plan JSON
func TestLoadWithBandPlanJSONErrors(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create corrupted band plan JSON
	corruptedBandPlan := []byte(`{
		"channels": [
			{
				"index": 1,
				"frequencyMhz": 2412,
				"name": "Channel 1"
			},
			{
				"index": 2,
				"frequencyMhz": "invalid-frequency", // Invalid type
				"name": "Channel 2"
			}
		]
	}`)
	if err := os.WriteFile("silvus-band-plan.json", corruptedBandPlan, 0644); err != nil {
		t.Fatalf("Failed to write corrupted band plan: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Expected error for corrupted band plan JSON, but got success")
	}
}

// TestLoadWithInvalidBandPlanJSON tests loading with invalid band plan JSON syntax
func TestLoadWithInvalidBandPlanJSON(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create invalid JSON syntax
	invalidBandPlan := []byte(`{
		"channels": [
			{
				"index": 1,
				"frequencyMhz": 2412,
				"name": "Channel 1"
			}
			// Missing comma and closing bracket
		]
	}`)
	if err := os.WriteFile("silvus-band-plan.json", invalidBandPlan, 0644); err != nil {
		t.Fatalf("Failed to write invalid band plan JSON: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Expected error for invalid band plan JSON, but got success")
	}
}

// TestLoadWithEnvBandPlan tests loading band plan from environment variable
func TestLoadWithEnvBandPlan(t *testing.T) {
	// Set band plan in environment variable
	bandPlanJSON := `{
		"channels": [
			{
				"index": 1,
				"frequencyMhz": 2412,
				"name": "Channel 1"
			},
			{
				"index": 2,
				"frequencyMhz": 2417,
				"name": "Channel 2"
			}
		]
	}`
	os.Setenv("RCC_SILVUS_BAND_PLAN", bandPlanJSON)
	defer os.Unsetenv("RCC_SILVUS_BAND_PLAN")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed with env band plan: %v", err)
	}

	if config.SilvusBandPlan == nil {
		t.Error("Expected SilvusBandPlan to be loaded from environment")
	}

	// Check that the band plan has the expected structure
	if config.SilvusBandPlan.Models == nil {
		t.Error("Expected SilvusBandPlan.Models to be loaded")
	}
}

// TestLoadWithInvalidEnvBandPlan tests loading with invalid environment band plan
func TestLoadWithInvalidEnvBandPlan(t *testing.T) {
	// Set invalid band plan in environment variable
	os.Setenv("RCC_SILVUS_BAND_PLAN", "invalid-json")
	defer os.Unsetenv("RCC_SILVUS_BAND_PLAN")

	// Load should still succeed (invalid env band plan is ignored)
	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed with invalid env band plan: %v", err)
	}

	if config == nil {
		t.Fatal("Load() returned nil config")
	}
}

// TestLoadWithFilePermissions tests loading with restricted file permissions
func TestLoadWithFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create config file with restricted permissions
	configData := []byte(`{
		"heartbeatInterval": "15s",
		"heartbeatJitter": "2s",
		"heartbeatTimeout": "45s"
	}`)
	configPath := "config.json"
	if err := os.WriteFile(configPath, configData, 0000); err != nil { // No read permissions
		t.Fatalf("Failed to write config with restricted permissions: %v", err)
	}

	// Load should fail due to permission error
	_, err := Load()
	if err == nil {
		t.Error("Expected error for restricted file permissions, but got success")
	}

	// Clean up
	os.Chmod(configPath, 0644)
}

// TestLoadWithLargeConfigFile tests loading with a very large config file
func TestLoadWithLargeConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create a large config file (but still valid JSON)
	largeConfig := map[string]interface{}{
		"heartbeatInterval": "15s",
		"heartbeatJitter":   "2s",
		"heartbeatTimeout":  "45s",
		"largeData":         make([]string, 10000), // Large array
	}

	// Fill the large array
	for i := 0; i < 10000; i++ {
		largeConfig["largeData"].([]string)[i] = "data"
	}

	configJSON, err := json.Marshal(largeConfig)
	if err != nil {
		t.Fatalf("Failed to marshal large config: %v", err)
	}

	if err := os.WriteFile("config.json", configJSON, 0644); err != nil {
		t.Fatalf("Failed to write large config: %v", err)
	}

	// Load should succeed
	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed with large config: %v", err)
	}

	if config == nil {
		t.Fatal("Load() returned nil config")
	}
}

// TestLoadWithConcurrentAccess tests loading with concurrent file access
func TestLoadWithConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create config file
	configData := []byte(`{
		"heartbeatInterval": "15s",
		"heartbeatJitter": "2s",
		"heartbeatTimeout": "45s"
	}`)
	if err := os.WriteFile("config.json", configData, 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Test concurrent loading
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			config, err := Load()
			if err != nil {
				t.Errorf("Concurrent Load() failed: %v", err)
			}
			if config == nil {
				t.Error("Concurrent Load() returned nil config")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestLoadWithMissingFiles tests loading when config files don't exist
func TestLoadWithMissingFiles(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// No config files should exist in temp directory
	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed with missing files: %v", err)
	}

	if config == nil {
		t.Fatal("Load() returned nil config")
	}

	// Should use default values
	if config.HeartbeatInterval != 15*time.Second {
		t.Errorf("Expected default HeartbeatInterval 15s, got %v", config.HeartbeatInterval)
	}
}

// TestLoadWithEmptyConfigFile tests loading with empty config file
func TestLoadWithEmptyConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create empty config file
	if err := os.WriteFile("config.json", []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write empty config: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Expected error for empty config file, but got success")
	}
}

// TestLoadWithNestedConfig tests loading with deeply nested config structure
func TestLoadWithNestedConfig(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create nested config
	nestedConfig := map[string]interface{}{
		"heartbeatInterval": "15s",
		"heartbeatJitter":   "2s",
		"heartbeatTimeout":  "45s",
		"nested": map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": "deep-value",
				},
			},
		},
	}

	configJSON, err := json.Marshal(nestedConfig)
	if err != nil {
		t.Fatalf("Failed to marshal nested config: %v", err)
	}

	if err := os.WriteFile("config.json", configJSON, 0644); err != nil {
		t.Fatalf("Failed to write nested config: %v", err)
	}

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed with nested config: %v", err)
	}

	if config == nil {
		t.Fatal("Load() returned nil config")
	}
}
