// Package config provides tests for Silvus band plan configuration.
//
//   - PRE-INT-09: "Unit tests for mapping using SilvusBandPlan"
//   - PRE-INT-09: "Negative test: missing index → INVALID_RANGE"
package config

import (
	"os"
	"testing"
)

// TestSilvusBandPlan_ChannelMapping tests channel mapping functionality.
func TestSilvusBandPlan_ChannelMapping(t *testing.T) {
	// Create test band plan
	bandPlan := &SilvusBandPlan{
		Models: map[string]map[string][]SilvusChannel{
			"Silvus-Scout": {
				"2.4GHz": {
					{ChannelIndex: 1, FrequencyMhz: 2412.0},
					{ChannelIndex: 2, FrequencyMhz: 2417.0},
					{ChannelIndex: 3, FrequencyMhz: 2422.0},
					{ChannelIndex: 4, FrequencyMhz: 2427.0},
					{ChannelIndex: 5, FrequencyMhz: 2432.0},
				},
				"5GHz": {
					{ChannelIndex: 1, FrequencyMhz: 5180.0},
					{ChannelIndex: 2, FrequencyMhz: 5200.0},
					{ChannelIndex: 3, FrequencyMhz: 5220.0},
					{ChannelIndex: 4, FrequencyMhz: 5240.0},
					{ChannelIndex: 5, FrequencyMhz: 5260.0},
				},
			},
			"Silvus-Tactical": {
				"UHF": {
					{ChannelIndex: 1, FrequencyMhz: 400.0},
					{ChannelIndex: 2, FrequencyMhz: 410.0},
					{ChannelIndex: 3, FrequencyMhz: 420.0},
				},
			},
		},
	}

	// Test valid channel mappings
	tests := []struct {
		model        string
		band         string
		channelIndex int
		expectedFreq float64
		description  string
	}{
		{"Silvus-Scout", "2.4GHz", 1, 2412.0, "Scout 2.4GHz channel 1"},
		{"Silvus-Scout", "2.4GHz", 3, 2422.0, "Scout 2.4GHz channel 3"},
		{"Silvus-Scout", "5GHz", 1, 5180.0, "Scout 5GHz channel 1"},
		{"Silvus-Scout", "5GHz", 5, 5260.0, "Scout 5GHz channel 5"},
		{"Silvus-Tactical", "UHF", 2, 410.0, "Tactical UHF channel 2"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			frequency, err := bandPlan.GetSilvusChannelFrequency(test.model, test.band, test.channelIndex)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if frequency != test.expectedFreq {
				t.Errorf("Expected frequency %.1f, got %.1f", test.expectedFreq, frequency)
			}
		})
	}

	// Test reverse mapping (frequency to channel index)
	reverseTests := []struct {
		model           string
		band            string
		frequency       float64
		expectedChannel int
		description     string
	}{
		{"Silvus-Scout", "2.4GHz", 2412.0, 1, "Scout 2.4GHz frequency 2412.0"},
		{"Silvus-Scout", "2.4GHz", 2422.0, 3, "Scout 2.4GHz frequency 2422.0"},
		{"Silvus-Scout", "5GHz", 5180.0, 1, "Scout 5GHz frequency 5180.0"},
		{"Silvus-Tactical", "UHF", 410.0, 2, "Tactical UHF frequency 410.0"},
	}

	for _, test := range reverseTests {
		t.Run(test.description, func(t *testing.T) {
			channelIndex, err := bandPlan.GetSilvusChannelIndex(test.model, test.band, test.frequency)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if channelIndex != test.expectedChannel {
				t.Errorf("Expected channel index %d, got %d", test.expectedChannel, channelIndex)
			}
		})
	}
}

// TestSilvusBandPlan_MissingIndex tests negative cases for missing channel indices.
func TestSilvusBandPlan_MissingIndex(t *testing.T) {
	bandPlan := &SilvusBandPlan{
		Models: map[string]map[string][]SilvusChannel{
			"Silvus-Scout": {
				"2.4GHz": {
					{ChannelIndex: 1, FrequencyMhz: 2412.0},
					{ChannelIndex: 2, FrequencyMhz: 2417.0},
					{ChannelIndex: 3, FrequencyMhz: 2422.0},
				},
			},
		},
	}

	// Test missing channel indices
	missingTests := []struct {
		model        string
		band         string
		channelIndex int
		description  string
	}{
		{"Silvus-Scout", "2.4GHz", 0, "Channel index 0 (out of range)"},
		{"Silvus-Scout", "2.4GHz", 4, "Channel index 4 (not in plan)"},
		{"Silvus-Scout", "2.4GHz", 10, "Channel index 10 (way out of range)"},
		{"Silvus-Scout", "5GHz", 1, "Channel index 1 in non-existent band"},
		{"Silvus-Tactical", "2.4GHz", 1, "Channel index 1 in non-existent model"},
	}

	for _, test := range missingTests {
		t.Run(test.description, func(t *testing.T) {
			_, err := bandPlan.GetSilvusChannelFrequency(test.model, test.band, test.channelIndex)
			if err == nil {
				t.Errorf("Expected error for missing channel index, got nil")
			}
			if !contains(err.Error(), "not found") {
				t.Errorf("Expected 'not found' error, got: %v", err)
			}
		})
	}

	// Test missing model
	_, err := bandPlan.GetSilvusChannelFrequency("NonExistentModel", "2.4GHz", 1)
	if err == nil {
		t.Error("Expected error for missing model, got nil")
	}
	if !contains(err.Error(), "model") {
		t.Errorf("Expected model error, got: %v", err)
	}

	// Test missing band
	_, err = bandPlan.GetSilvusChannelFrequency("Silvus-Scout", "NonExistentBand", 1)
	if err == nil {
		t.Error("Expected error for missing band, got nil")
	}
	if !contains(err.Error(), "band") {
		t.Errorf("Expected band error, got: %v", err)
	}
}

// TestSilvusBandPlan_HelperMethods tests helper methods for band plan management.
func TestSilvusBandPlan_HelperMethods(t *testing.T) {
	bandPlan := &SilvusBandPlan{
		Models: map[string]map[string][]SilvusChannel{
			"Silvus-Scout": {
				"2.4GHz": {{ChannelIndex: 1, FrequencyMhz: 2412.0}},
				"5GHz":   {{ChannelIndex: 1, FrequencyMhz: 5180.0}},
			},
			"Silvus-Tactical": {
				"UHF": {{ChannelIndex: 1, FrequencyMhz: 400.0}},
			},
		},
	}

	// Test HasModelBand
	if !bandPlan.HasModelBand("Silvus-Scout", "2.4GHz") {
		t.Error("Expected Silvus-Scout 2.4GHz to exist")
	}
	if bandPlan.HasModelBand("Silvus-Scout", "NonExistentBand") {
		t.Error("Expected Silvus-Scout NonExistentBand to not exist")
	}
	if bandPlan.HasModelBand("NonExistentModel", "2.4GHz") {
		t.Error("Expected NonExistentModel 2.4GHz to not exist")
	}

	// Test GetAvailableModels
	models := bandPlan.GetAvailableModels()
	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}
	if !containsSlice(models, "Silvus-Scout") {
		t.Error("Expected Silvus-Scout in available models")
	}
	if !containsSlice(models, "Silvus-Tactical") {
		t.Error("Expected Silvus-Tactical in available models")
	}

	// Test GetAvailableBands
	scoutBands := bandPlan.GetAvailableBands("Silvus-Scout")
	if len(scoutBands) != 2 {
		t.Errorf("Expected 2 bands for Silvus-Scout, got %d", len(scoutBands))
	}
	if !containsSlice(scoutBands, "2.4GHz") {
		t.Error("Expected 2.4GHz in Scout bands")
	}
	if !containsSlice(scoutBands, "5GHz") {
		t.Error("Expected 5GHz in Scout bands")
	}

	// Test GetAvailableBands for non-existent model
	nonExistentBands := bandPlan.GetAvailableBands("NonExistentModel")
	if len(nonExistentBands) != 0 {
		t.Errorf("Expected 0 bands for non-existent model, got %d", len(nonExistentBands))
	}
}

// TestSilvusBandPlan_NilBandPlan tests behavior with nil band plan.
func TestSilvusBandPlan_NilBandPlan(t *testing.T) {
	var bandPlan *SilvusBandPlan

	// Test all methods with nil band plan
	_, err := bandPlan.GetSilvusChannelFrequency("model", "band", 1)
	if err == nil {
		t.Error("Expected error for nil band plan, got nil")
	}

	_, err = bandPlan.GetSilvusChannelIndex("model", "band", 2412.0)
	if err == nil {
		t.Error("Expected error for nil band plan, got nil")
	}

	if bandPlan.HasModelBand("model", "band") {
		t.Error("Expected false for nil band plan")
	}

	models := bandPlan.GetAvailableModels()
	if len(models) != 0 {
		t.Errorf("Expected empty models list for nil band plan, got %d", len(models))
	}

	bands := bandPlan.GetAvailableBands("model")
	if len(bands) != 0 {
		t.Errorf("Expected empty bands list for nil band plan, got %d", len(bands))
	}
}

// TestSilvusBandPlan_JSONLoading tests loading band plans from JSON.
func TestSilvusBandPlan_JSONLoading(t *testing.T) {
	// Test JSON string loading
	jsonStr := `{
		"models": {
			"Silvus-Scout": {
				"2.4GHz": [
					{"channelIndex": 1, "frequencyMhz": 2412.0},
					{"channelIndex": 2, "frequencyMhz": 2417.0}
				]
			}
		}
	}`

	bandPlan, err := loadSilvusBandPlanFromJSON(jsonStr)
	if err != nil {
		t.Fatalf("Failed to load band plan from JSON: %v", err)
	}

	// Test that the loaded band plan works
	frequency, err := bandPlan.GetSilvusChannelFrequency("Silvus-Scout", "2.4GHz", 1)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if frequency != 2412.0 {
		t.Errorf("Expected frequency 2412.0, got %.1f", frequency)
	}

	// Test invalid JSON
	_, err = loadSilvusBandPlanFromJSON("invalid json")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// TestSilvusBandPlan_FileLoading tests loading band plans from files.
func TestSilvusBandPlan_FileLoading(t *testing.T) {
	// Create a temporary JSON file
	jsonContent := `{
		"models": {
			"Silvus-Test": {
				"TestBand": [
					{"channelIndex": 1, "frequencyMhz": 1000.0},
					{"channelIndex": 2, "frequencyMhz": 2000.0}
				]
			}
		}
	}`

	// Write to temporary file
	tmpFile, err := os.CreateTemp("", "silvus-band-plan-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(jsonContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpFile.Close()

	// Test loading from file
	bandPlan, err := loadSilvusBandPlanFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load band plan from file: %v", err)
	}

	// Test that the loaded band plan works
	frequency, err := bandPlan.GetSilvusChannelFrequency("Silvus-Test", "TestBand", 1)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if frequency != 1000.0 {
		t.Errorf("Expected frequency 1000.0, got %.1f", frequency)
	}

	// Test loading from non-existent file
	_, err = loadSilvusBandPlanFromFile("non-existent-file.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestSilvusBandPlan_EnvironmentVariable tests loading from environment variables.
func TestSilvusBandPlan_EnvironmentVariable(t *testing.T) {
	// Set environment variable
	jsonStr := `{
		"models": {
			"Silvus-Env": {
				"EnvBand": [
					{"channelIndex": 1, "frequencyMhz": 3000.0}
				]
			}
		}
	}`
	_ = os.Setenv("RCC_SILVUS_BAND_PLAN", jsonStr)
	defer func() { _ = os.Unsetenv("RCC_SILVUS_BAND_PLAN") }()

	// Test loading from environment
	bandPlan, err := loadSilvusBandPlanFromJSON(jsonStr)
	if err != nil {
		t.Fatalf("Failed to load band plan from environment: %v", err)
	}

	// Test that the loaded band plan works
	frequency, err := bandPlan.GetSilvusChannelFrequency("Silvus-Env", "EnvBand", 1)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if frequency != 3000.0 {
		t.Errorf("Expected frequency 3000.0, got %.1f", frequency)
	}
}

// Helper function for string slice contains check
func containsSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function for string contains check (for error messages)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
