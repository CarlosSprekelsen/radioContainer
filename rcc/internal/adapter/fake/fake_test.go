// Package fake provides a fake radio adapter implementation for testing.
package fake

import (
	"context"
	"testing"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/adaptertest"
)

// TestFakeAdapterConformance runs the complete conformance test suite on the fake adapter.
func TestFakeAdapterConformance(t *testing.T) {
	// Define capabilities for the fake adapter
	capabilities := adaptertest.Capabilities{
		MinPowerDbm:      0,
		MaxPowerDbm:      39,
		ValidFrequencies: []float64{2412.0, 2417.0, 2422.0, 2427.0, 2432.0},
		Channels: []adapter.Channel{
			{Index: 1, FrequencyMhz: 2412.0},
			{Index: 2, FrequencyMhz: 2417.0},
			{Index: 3, FrequencyMhz: 2422.0},
			{Index: 4, FrequencyMhz: 2427.0},
			{Index: 5, FrequencyMhz: 2432.0},
		},
		ExpectedErrors: adaptertest.ErrorExpectations{
			InvalidRangeKeywords: []string{"INVALID_RANGE", "OUT_OF_RANGE", "INVALID_PARAMETER"},
			BusyKeywords:         []string{"BUSY", "RETRY", "RATE_LIMIT"},
			UnavailableKeywords:  []string{"UNAVAILABLE", "OFFLINE", "NOT_READY"},
			InternalKeywords:     []string{"INTERNAL", "UNKNOWN", "ERROR"},
		},
	}

	// Run conformance tests
	adaptertest.RunConformance(t, func() adapter.IRadioAdapter {
		return NewFakeAdapter("fake-radio-01")
	}, capabilities)
}

// TestFakeAdapterBasicFunctionality tests basic functionality of the fake adapter.
func TestFakeAdapterBasicFunctionality(t *testing.T) {
	adapter := NewFakeAdapter("test-radio")
	ctx := context.Background()

	// Test basic state
	state, err := adapter.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}

	if state.PowerDbm != 20.0 {
		t.Errorf("Expected power 20.0, got %f", state.PowerDbm)
	}

	if state.FrequencyMhz != 2412.0 {
		t.Errorf("Expected frequency 2412.0, got %f", state.FrequencyMhz)
	}

	// Test power setting
	err = adapter.SetPower(ctx, 30.0)
	if err != nil {
		t.Fatalf("SetPower failed: %v", err)
	}

	state, err = adapter.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState after SetPower failed: %v", err)
	}

	if state.PowerDbm != 30.0 {
		t.Errorf("Expected power 30.0 after SetPower, got %f", state.PowerDbm)
	}

	// Test frequency setting
	err = adapter.SetFrequency(ctx, 2417.0)
	if err != nil {
		t.Fatalf("SetFrequency failed: %v", err)
	}

	state, err = adapter.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState after SetFrequency failed: %v", err)
	}

	if state.FrequencyMhz != 2417.0 {
		t.Errorf("Expected frequency 2417.0 after SetFrequency, got %f", state.FrequencyMhz)
	}
}

// TestFakeAdapterErrorSimulation tests error simulation functionality.
func TestFakeAdapterErrorSimulation(t *testing.T) {
	adapter := NewFakeAdapter("test-radio")
	ctx := context.Background()

	// Test INVALID_RANGE error simulation
	adapter.SetErrorSimulation("INVALID_RANGE")

	_, err := adapter.GetState(ctx)
	if err == nil {
		t.Error("Expected error when error simulation is enabled")
	}

	if err.Error() != "INVALID_RANGE: simulated range error" {
		t.Errorf("Expected INVALID_RANGE error, got: %v", err)
	}

	// Test BUSY error simulation
	adapter.SetErrorSimulation("BUSY")

	err = adapter.SetPower(ctx, 20)
	if err == nil {
		t.Error("Expected error when error simulation is enabled")
	}

	if err.Error() != "BUSY: simulated busy error" {
		t.Errorf("Expected BUSY error, got: %v", err)
	}

	// Disable error simulation
	adapter.DisableErrorSimulation()

	_, err = adapter.GetState(ctx)
	if err != nil {
		t.Errorf("Expected no error when error simulation is disabled, got: %v", err)
	}
}

// TestFakeAdapterValidation tests input validation.
func TestFakeAdapterValidation(t *testing.T) {
	adapter := NewFakeAdapter("test-radio")
	ctx := context.Background()

	// Test invalid power range
	err := adapter.SetPower(ctx, -1)
	if err == nil {
		t.Error("Expected error for invalid power (-1)")
	}

	err = adapter.SetPower(ctx, 100)
	if err == nil {
		t.Error("Expected error for invalid power (100)")
	}

	// Test invalid frequency
	err = adapter.SetFrequency(ctx, 0)
	if err == nil {
		t.Error("Expected error for invalid frequency (0)")
	}

	err = adapter.SetFrequency(ctx, -100)
	if err == nil {
		t.Error("Expected error for invalid frequency (-100)")
	}

	err = adapter.SetFrequency(ctx, 10000)
	if err == nil {
		t.Error("Expected error for invalid frequency (10000)")
	}
}

// TestFakeAdapterReadPowerActual tests ReadPowerActual functionality.
func TestFakeAdapterReadPowerActual(t *testing.T) {
	adapter := NewFakeAdapter("test-radio")
	ctx := context.Background()

	// Test initial power reading
	power, err := adapter.ReadPowerActual(ctx)
	if err != nil {
		t.Fatalf("ReadPowerActual failed: %v", err)
	}

	if power != 20.0 {
		t.Errorf("Expected initial power 20.0, got %f", power)
	}

	// Test power reading after setting power
	err = adapter.SetPower(ctx, 35.0)
	if err != nil {
		t.Fatalf("SetPower failed: %v", err)
	}

	power, err = adapter.ReadPowerActual(ctx)
	if err != nil {
		t.Fatalf("ReadPowerActual after SetPower failed: %v", err)
	}

	if power != 35.0 {
		t.Errorf("Expected power 35.0 after SetPower, got %f", power)
	}
}

// TestFakeAdapterSupportedFrequencyProfiles tests SupportedFrequencyProfiles functionality.
func TestFakeAdapterSupportedFrequencyProfiles(t *testing.T) {
	adapter := NewFakeAdapter("test-radio")
	ctx := context.Background()

	profiles, err := adapter.SupportedFrequencyProfiles(ctx)
	if err != nil {
		t.Fatalf("SupportedFrequencyProfiles failed: %v", err)
	}

	// Check that we get expected profiles - the fake adapter returns one profile with validFreqs
	if len(profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(profiles))
	}

	profile := profiles[0]
	if profile.Bandwidth != 20.0 {
		t.Errorf("Expected Bandwidth 20.0, got %f", profile.Bandwidth)
	}
	if profile.AntennaMask != 1 {
		t.Errorf("Expected AntennaMask 1, got %d", profile.AntennaMask)
	}
	// Check that we have some frequencies
	if len(profile.Frequencies) == 0 {
		t.Errorf("Expected at least one frequency, got %d", len(profile.Frequencies))
	}
}

// TestFakeAdapterGetSetCurrentState tests GetCurrentState and SetCurrentState functionality.
func TestFakeAdapterGetSetCurrentState(t *testing.T) {
	adapter := NewFakeAdapter("test-radio")

	// Test initial state
	power, frequency := adapter.GetCurrentState()

	expectedPower := 20.0
	expectedFrequency := 2412.0

	if power != expectedPower {
		t.Errorf("Expected initial power %f, got %f", expectedPower, power)
	}
	if frequency != expectedFrequency {
		t.Errorf("Expected initial frequency %f, got %f", expectedFrequency, frequency)
	}

	// Test setting new state
	newPower := 25.0
	newFrequency := 2420.0

	adapter.SetCurrentState(newPower, newFrequency)

	// Test getting the updated state
	updatedPower, updatedFrequency := adapter.GetCurrentState()

	if updatedPower != newPower {
		t.Errorf("Expected updated power %f, got %f", newPower, updatedPower)
	}
	if updatedFrequency != newFrequency {
		t.Errorf("Expected updated frequency %f, got %f", newFrequency, updatedFrequency)
	}
}

// TestFakeAdapterStatePersistence tests that state changes persist across operations.
func TestFakeAdapterStatePersistence(t *testing.T) {
	adapter := NewFakeAdapter("test-radio")
	ctx := context.Background()

	// Set power and frequency
	err := adapter.SetPower(ctx, 30.0)
	if err != nil {
		t.Fatalf("SetPower failed: %v", err)
	}

	err = adapter.SetFrequency(ctx, 2425.0)
	if err != nil {
		t.Fatalf("SetFrequency failed: %v", err)
	}

	// Verify ReadPowerActual reflects the change
	power, err := adapter.ReadPowerActual(ctx)
	if err != nil {
		t.Fatalf("ReadPowerActual failed: %v", err)
	}
	if power != 30.0 {
		t.Errorf("Expected ReadPowerActual to return 30.0, got %f", power)
	}

	// Verify GetState reflects the changes
	state, err := adapter.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if state.PowerDbm != 30.0 {
		t.Errorf("Expected GetState PowerDbm 30.0, got %f", state.PowerDbm)
	}
	if state.FrequencyMhz != 2425.0 {
		t.Errorf("Expected GetState FrequencyMhz 2425.0, got %f", state.FrequencyMhz)
	}

	// Verify GetCurrentState reflects the changes
	currentPower, currentFrequency := adapter.GetCurrentState()
	if currentPower != 30.0 {
		t.Errorf("Expected GetCurrentState power 30.0, got %f", currentPower)
	}
	if currentFrequency != 2425.0 {
		t.Errorf("Expected GetCurrentState frequency 2425.0, got %f", currentFrequency)
	}
}
