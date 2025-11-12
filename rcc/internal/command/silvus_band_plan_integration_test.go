// Package command provides integration tests for Silvus band plan functionality.
//
//   - PRE-INT-09: "orchestrator.SetChannel consults this when adapter capabilities carry a model that matches"
//go:build integration
// +build integration

package command

import (
	"context"
	"fmt"
	"testing"

	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/radio"
)

// SilvusTestRadioManager implements RadioManager interface for testing.
type SilvusTestRadioManager struct {
	radios map[string]*radio.Radio
}

func (m *SilvusTestRadioManager) GetRadio(radioID string) (*radio.Radio, error) {
	radioObj, exists := m.radios[radioID]
	if !exists {
		return nil, fmt.Errorf("radio %s not found", radioID)
	}
	return radioObj, nil
}

func (m *SilvusTestRadioManager) SetActive(radioID string) error {
	// Mock implementation - just verify radio exists
	if _, exists := m.radios[radioID]; !exists {
		return fmt.Errorf("radio %s not found", radioID)
	}
	return nil
}

// TestOrchestrator_SilvusBandPlanIntegration tests orchestrator integration with Silvus band plans.
func TestOrchestrator_SilvusBandPlanIntegration(t *testing.T) {
	// Create test configuration with Silvus band plan
	cfg := &config.TimingConfig{
		SilvusBandPlan: &config.SilvusBandPlan{
			Models: map[string]map[string][]config.SilvusChannel{
				"Silvus-Scout": {
					"2.4GHz": {
						{ChannelIndex: 1, FrequencyMhz: 2412.0},
						{ChannelIndex: 2, FrequencyMhz: 2417.0},
						{ChannelIndex: 3, FrequencyMhz: 2422.0},
					},
					"5GHz": {
						{ChannelIndex: 1, FrequencyMhz: 5180.0},
						{ChannelIndex: 2, FrequencyMhz: 5200.0},
						{ChannelIndex: 3, FrequencyMhz: 5220.0},
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
		},
	}

	// Create orchestrator with configuration
	orchestrator := &Orchestrator{
		config: cfg,
	}

	// Create mock radio manager with radio data
	radioManager := &SilvusTestRadioManager{
		radios: map[string]*radio.Radio{
			"radio-01": {
				ID:    "radio-01",
				Model: "Silvus-Scout",
			},
			"radio-02": {
				ID:    "radio-02",
				Model: "Silvus-Scout",
			},
			"radio-03": {
				ID:    "radio-03",
				Model: "Silvus-Tactical",
			},
			"radio-04": {
				ID:    "radio-04",
				Model: "Unknown-Model",
			},
		},
	}

	orchestrator.SetRadioManager(radioManager)

	// Test valid channel index resolution
	tests := []struct {
		radioID      string
		channelIndex int
		expectedFreq float64
		description  string
	}{
		{"radio-01", 1, 2412.0, "Scout 2.4GHz channel 1"},
		{"radio-01", 2, 2417.0, "Scout 2.4GHz channel 2"},
		{"radio-01", 3, 2422.0, "Scout 2.4GHz channel 3"},
		{"radio-02", 1, 5180.0, "Scout 5GHz channel 1"},
		{"radio-02", 2, 5200.0, "Scout 5GHz channel 2"},
		{"radio-03", 1, 400.0, "Tactical UHF channel 1"},
		{"radio-03", 2, 410.0, "Tactical UHF channel 2"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			frequency, err := orchestrator.resolveChannelIndex(context.Background(), test.radioID, test.channelIndex, radioManager)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if frequency != test.expectedFreq {
				t.Errorf("Expected frequency %.1f, got %.1f", test.expectedFreq, frequency)
			}
		})
	}

	// Test missing channel index (should return INVALID_RANGE error)
	missingTests := []struct {
		radioID      string
		channelIndex int
		description  string
	}{
		{"radio-01", 0, "Scout 2.4GHz channel 0 (out of range)"},
		{"radio-01", 4, "Scout 2.4GHz channel 4 (not in plan)"},
		{"radio-02", 10, "Scout 5GHz channel 10 (way out of range)"},
		{"radio-03", 5, "Tactical UHF channel 5 (not in plan)"},
	}

	for _, test := range missingTests {
		t.Run(test.description, func(t *testing.T) {
			_, err := orchestrator.resolveChannelIndex(context.Background(), test.radioID, test.channelIndex, radioManager)
			if err == nil {
				t.Errorf("Expected error for missing channel index, got nil")
			}
			// The error should be from Silvus band plan, not radio manager
			// For missing channels, we expect the Silvus band plan to return an error
			if !contains(err.Error(), "not found") && !contains(err.Error(), "channel index") && !contains(err.Error(), "capabilities") {
				t.Errorf("Expected 'not found', 'channel index', or 'capabilities' error, got: %v", err)
			}
		})
	}

	// Test unknown model/band (should fall back to radio manager)
	unknownTests := []struct {
		radioID      string
		channelIndex int
		description  string
	}{
		{"radio-04", 1, "Unknown model should fall back to radio manager"},
	}

	for _, test := range unknownTests {
		t.Run(test.description, func(t *testing.T) {
			_, err := orchestrator.resolveChannelIndex(context.Background(), test.radioID, test.channelIndex, radioManager)
			// This should fail because the radio manager doesn't have channel capabilities
			if err == nil {
				t.Errorf("Expected error for unknown model, got nil")
			}
		})
	}
}

// TestOrchestrator_SilvusBandPlanFallback tests fallback to radio manager when Silvus band plan doesn't have the channel.
func TestOrchestrator_SilvusBandPlanFallback(t *testing.T) {
	// Create configuration with limited Silvus band plan
	cfg := &config.TimingConfig{
		SilvusBandPlan: &config.SilvusBandPlan{
			Models: map[string]map[string][]config.SilvusChannel{
				"Silvus-Scout": {
					"2.4GHz": {
						{ChannelIndex: 1, FrequencyMhz: 2412.0},
						// Missing channels 2 and 3
					},
				},
			},
		},
	}

	orchestrator := &Orchestrator{
		config: cfg,
	}

	// Create mock radio manager with full channel capabilities
	radioManager := &SilvusTestRadioManager{
		radios: map[string]*radio.Radio{
			"radio-01": map[string]*radio.Radio{
				"model": "Silvus-Scout",
				"band":  "2.4GHz",
				"capabilities": map[string]*radio.Radio{
					"channels": []interface{}{
						map[string]*radio.Radio{"index": float64(1), "frequencyMhz": 2412.0},
						map[string]*radio.Radio{"index": float64(2), "frequencyMhz": 2417.0},
						map[string]*radio.Radio{"index": float64(3), "frequencyMhz": 2422.0},
					},
				},
			},
		},
	}

	orchestrator.SetRadioManager(radioManager)

	// Test that channel 1 uses Silvus band plan
	frequency, err := orchestrator.resolveChannelIndex(context.Background(), "radio-01", 1, radioManager)
	if err != nil {
		t.Errorf("Expected no error for channel 1, got: %v", err)
	}
	if frequency != 2412.0 {
		t.Errorf("Expected frequency 2412.0, got %.1f", frequency)
	}

	// Test that channel 2 falls back to radio manager
	frequency, err = orchestrator.resolveChannelIndex(context.Background(), "radio-01", 2, radioManager)
	if err != nil {
		t.Errorf("Expected no error for channel 2, got: %v", err)
	}
	if frequency != 2417.0 {
		t.Errorf("Expected frequency 2417.0, got %.1f", frequency)
	}

	// Test that channel 3 falls back to radio manager
	frequency, err = orchestrator.resolveChannelIndex(context.Background(), "radio-01", 3, radioManager)
	if err != nil {
		t.Errorf("Expected no error for channel 3, got: %v", err)
	}
	if frequency != 2422.0 {
		t.Errorf("Expected frequency 2422.0, got %.1f", frequency)
	}
}

// TestOrchestrator_NoSilvusBandPlan tests behavior when no Silvus band plan is configured.
func TestOrchestrator_NoSilvusBandPlan(t *testing.T) {
	// Create configuration without Silvus band plan
	cfg := &config.TimingConfig{
		SilvusBandPlan: nil,
	}

	orchestrator := &Orchestrator{
		config: cfg,
	}

	// Create mock radio manager
	radioManager := &SilvusTestRadioManager{
		radios: map[string]*radio.Radio{
			"radio-01": map[string]*radio.Radio{
				"model": "Silvus-Scout",
				"band":  "2.4GHz",
				"capabilities": map[string]*radio.Radio{
					"channels": []interface{}{
						map[string]*radio.Radio{"index": float64(1), "frequencyMhz": 2412.0},
						map[string]*radio.Radio{"index": float64(2), "frequencyMhz": 2417.0},
					},
				},
			},
		},
	}

	orchestrator.SetRadioManager(radioManager)

	// Test that it falls back to radio manager
	frequency, err := orchestrator.resolveChannelIndex(context.Background(), "radio-01", 1, radioManager)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if frequency != 2412.0 {
		t.Errorf("Expected frequency 2412.0, got %.1f", frequency)
	}
}

// TestOrchestrator_GetRadioModelAndBand tests model and band extraction.
func TestOrchestrator_GetRadioModelAndBand(t *testing.T) {
	orchestrator := &Orchestrator{}

	radioManager := &SilvusTestRadioManager{
		radios: map[string]*radio.Radio{
			"radio-01": map[string]*radio.Radio{
				"model": "Silvus-Scout",
				"band":  "2.4GHz",
			},
			"radio-02": map[string]*radio.Radio{
				"model": "Silvus-Tactical",
				// No band specified
			},
			"radio-03": map[string]*radio.Radio{
				// No model specified
			},
		},
	}

	// Test valid model and band
	model, band, err := orchestrator.getRadioModelAndBand(context.Background(), "radio-01", radioManager)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if model != "Silvus-Scout" {
		t.Errorf("Expected model 'Silvus-Scout', got '%s'", model)
	}
	if band != "2.4GHz" {
		t.Errorf("Expected band '2.4GHz', got '%s'", band)
	}

	// Test model without band (should default to "default")
	model, band, err = orchestrator.getRadioModelAndBand(context.Background(), "radio-02", radioManager)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if model != "Silvus-Tactical" {
		t.Errorf("Expected model 'Silvus-Tactical', got '%s'", model)
	}
	if band != "default" {
		t.Errorf("Expected band 'default', got '%s'", band)
	}

	// Test missing model
	_, _, err = orchestrator.getRadioModelAndBand(context.Background(), "radio-03", radioManager)
	if err == nil {
		t.Error("Expected error for missing model, got nil")
	}
	if !contains(err.Error(), "no model") {
		t.Errorf("Expected 'no model' error, got: %v", err)
	}

	// Test missing radio
	_, _, err = orchestrator.getRadioModelAndBand(context.Background(), "radio-99", radioManager)
	if err == nil {
		t.Error("Expected error for missing radio, got nil")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// Helper function for string contains check
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
