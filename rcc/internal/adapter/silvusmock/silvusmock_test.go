// Package silvusmock provides tests for the SilvusMock adapter.
//
//   - PRE-INT-08: "adaptertest.RunConformance passes with SilvusMock"
//   - PRE-INT-08: "API e2e test: register SilvusMock, exercise select/power/channel, observe telemetry + audit"
package silvusmock

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/adaptertest"
)

// TestSilvusMock_BasicFunctionality tests basic SilvusMock functionality.
func TestSilvusMock_BasicFunctionality(t *testing.T) {
	mock := NewSilvusMock("test-radio-01", nil)
	ctx := context.Background()

	// Test initial state
	state, err := mock.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if state.PowerDbm != 20 {
		t.Errorf("Expected initial power 20, got %f", state.PowerDbm)
	}
	if state.FrequencyMhz != 2412.0 {
		t.Errorf("Expected initial frequency 2412.0, got %f", state.FrequencyMhz)
	}

	// Test SetPower
	err = mock.SetPower(ctx, 25)
	if err != nil {
		t.Fatalf("SetPower failed: %v", err)
	}

	state, err = mock.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState after SetPower failed: %v", err)
	}
	if state.PowerDbm != 25 {
		t.Errorf("Expected power 25, got %f", state.PowerDbm)
	}

	// Test SetFrequency
	err = mock.SetFrequency(ctx, 2417.0)
	if err != nil {
		t.Fatalf("SetFrequency failed: %v", err)
	}

	state, err = mock.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState after SetFrequency failed: %v", err)
	}
	if state.FrequencyMhz != 2417.0 {
		t.Errorf("Expected frequency 2417.0, got %f", state.FrequencyMhz)
	}
}

// TestSilvusMock_FaultInjection tests fault injection modes.
func TestSilvusMock_FaultInjection(t *testing.T) {
	mock := NewSilvusMock("test-radio-02", nil)
	ctx := context.Background()

	// Test ReturnBusy fault
	mock.SetFaultMode("ReturnBusy")
	err := mock.SetPower(ctx, 25)
	if err == nil {
		t.Error("Expected busy error, got nil")
	}
	if err.Error() != "BUSY: SilvusMock simulated busy error for SetPower" {
		t.Errorf("Expected busy error message, got: %v", err)
	}

	// Test ReturnUnavailable fault
	mock.SetFaultMode("ReturnUnavailable")
	_, err = mock.GetState(ctx)
	if err == nil {
		t.Error("Expected unavailable error, got nil")
	}
	if err.Error() != "UNAVAILABLE: SilvusMock simulated unavailable error for GetState" {
		t.Errorf("Expected unavailable error message, got: %v", err)
	}

	// Test ReturnInvalidRange fault
	mock.SetFaultMode("ReturnInvalidRange")
	err = mock.SetFrequency(ctx, 2417.0)
	if err == nil {
		t.Error("Expected invalid range error, got nil")
	}
	if err.Error() != "INVALID_RANGE: SilvusMock simulated invalid range error for SetFrequency" {
		t.Errorf("Expected invalid range error message, got: %v", err)
	}

	// Test clearing fault mode
	mock.ClearFaultMode()
	err = mock.SetPower(ctx, 25)
	if err != nil {
		t.Errorf("Expected no error after clearing fault mode, got: %v", err)
	}
}

// TestSilvusMock_BandPlan tests band plan functionality.
func TestSilvusMock_BandPlan(t *testing.T) {
	// Create custom band plan
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2400.0},
		{Index: 2, FrequencyMhz: 2405.0},
		{Index: 3, FrequencyMhz: 2410.0},
	}

	mock := NewSilvusMock("test-radio-03", bandPlan)
	ctx := context.Background()

	// Test that band plan is set correctly
	retrievedPlan := mock.GetBandPlan()
	if len(retrievedPlan) != 3 {
		t.Errorf("Expected band plan length 3, got %d", len(retrievedPlan))
	}

	// Test channel mapping
	err := mock.SetFrequency(ctx, 2405.0)
	if err != nil {
		t.Fatalf("SetFrequency failed: %v", err)
	}

	// Check that channel index was updated
	_, _, channelIndex := mock.GetCurrentState()
	if channelIndex != 2 {
		t.Errorf("Expected channel index 2, got %d", channelIndex)
	}

	// Test frequency profiles
	profiles, err := mock.SupportedFrequencyProfiles(ctx)
	if err != nil {
		t.Fatalf("SupportedFrequencyProfiles failed: %v", err)
	}
	if len(profiles) != 1 {
		t.Errorf("Expected 1 frequency profile, got %d", len(profiles))
	}
	if len(profiles[0].Frequencies) != 3 {
		t.Errorf("Expected 3 frequencies in profile, got %d", len(profiles[0].Frequencies))
	}
}

// TestSilvusMock_Validation tests input validation.
func TestSilvusMock_Validation(t *testing.T) {
	mock := NewSilvusMock("test-radio-04", nil)
	ctx := context.Background()

	// Test invalid power ranges
	invalidPowers := []float64{-1.0, 40.0, 100.0}
	for _, power := range invalidPowers {
		err := mock.SetPower(ctx, power)
		if err == nil {
			t.Errorf("Expected error for power %f, got nil", power)
		}
		if err.Error() != fmt.Sprintf("INVALID_RANGE: power %f is outside valid range [0, 39]", power) {
			t.Errorf("Expected INVALID_RANGE error for power %f, got: %v", power, err)
		}
	}

	// Test invalid frequencies
	invalidFrequencies := []float64{0, -100, 100000}
	for _, freq := range invalidFrequencies {
		err := mock.SetFrequency(ctx, freq)
		if err == nil {
			t.Errorf("Expected error for frequency %f, got nil", freq)
		}
		if !contains(err.Error(), "INVALID_RANGE") {
			t.Errorf("Expected INVALID_RANGE error for frequency %f, got: %v", freq, err)
		}
	}
}

// TestSilvusMock_Concurrency tests thread safety.
func TestSilvusMock_Concurrency(t *testing.T) {
	mock := NewSilvusMock("test-radio-05", nil)
	ctx := context.Background()

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			defer func() { done <- true }()

			// Mix of operations
			if index%2 == 0 {
				_ = mock.SetPower(ctx, float64(20+index))
			} else {
				_ = mock.SetFrequency(ctx, 2412.0+float64(index))
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state is consistent
	state, err := mock.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState after concurrent access failed: %v", err)
	}
	if state.PowerDbm < 0 || state.PowerDbm > 39 {
		t.Errorf("Power out of range after concurrent access: %f", state.PowerDbm)
	}
	if state.FrequencyMhz < 100 || state.FrequencyMhz > 6000 {
		t.Errorf("Frequency out of range after concurrent access: %f", state.FrequencyMhz)
	}
}

// TestSilvusMock_Conformance tests that SilvusMock passes adapter conformance tests.
func TestSilvusMock_Conformance(t *testing.T) {
	// Define capabilities for SilvusMock
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
			InvalidRangeKeywords: []string{"INVALID_RANGE"},
			BusyKeywords:         []string{"BUSY"},
			UnavailableKeywords:  []string{"UNAVAILABLE"},
			InternalKeywords:     []string{"INTERNAL"},
		},
	}

	// Create adapter factory function
	newAdapter := func() adapter.IRadioAdapter {
		return NewSilvusMock("conformance-test", capabilities.Channels)
	}

	// Run conformance tests
	adaptertest.RunConformance(t, newAdapter, capabilities)
}

// TestSilvusMock_StateManagement tests in-memory state management.
func TestSilvusMock_StateManagement(t *testing.T) {
	mock := NewSilvusMock("test-radio-06", nil)
	ctx := context.Background()

	// Test initial state
	power, freq, channel := mock.GetCurrentState()
	if power != 20 {
		t.Errorf("Expected initial power 20, got %f", power)
	}
	if freq != 2412.0 {
		t.Errorf("Expected initial frequency 2412.0, got %f", freq)
	}
	if channel != 1 {
		t.Errorf("Expected initial channel 1, got %d", channel)
	}

	// Test state updates
	mock.SetCurrentState(25, 2417.0, 2)
	power, freq, channel = mock.GetCurrentState()
	if power != 25 {
		t.Errorf("Expected power 25, got %f", power)
	}
	if freq != 2417.0 {
		t.Errorf("Expected frequency 2417.0, got %f", freq)
	}
	if channel != 2 {
		t.Errorf("Expected channel 2, got %d", channel)
	}

	// Test that state persists across operations
	state, err := mock.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if state.PowerDbm != 25 {
		t.Errorf("Expected persisted power 25, got %f", state.PowerDbm)
	}
	if state.FrequencyMhz != 2417.0 {
		t.Errorf("Expected persisted frequency 2417.0, got %f", state.FrequencyMhz)
	}
}

// TestSilvusMock_LastCommandTime tests command timing.
func TestSilvusMock_LastCommandTime(t *testing.T) {
	mock := NewSilvusMock("test-radio-07", nil)
	ctx := context.Background()

	initialTime := mock.GetLastCommandTime()

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Perform an operation
	err := mock.SetPower(ctx, 30)
	if err != nil {
		t.Fatalf("SetPower failed: %v", err)
	}

	// Check that last command time was updated
	updatedTime := mock.GetLastCommandTime()
	if !updatedTime.After(initialTime) {
		t.Error("Expected last command time to be updated")
	}
}

// TestSilvusMock_SimulateSilvusBehavior tests Silvus-specific behavior simulation.
func TestSilvusMock_SimulateSilvusBehavior(t *testing.T) {
	mock := NewSilvusMock("test-radio-08", nil)

	// Test initial status
	if mock.GetStatus() != "online" {
		t.Errorf("Expected initial status 'online', got '%s'", mock.GetStatus())
	}

	// Simulate Silvus behavior
	mock.SimulateSilvusBehavior()

	// Check that status is still online
	if mock.GetStatus() != "online" {
		t.Errorf("Expected status 'online' after simulation, got '%s'", mock.GetStatus())
	}

	// Check that last command time was updated
	lastTime := mock.GetLastCommandTime()
	if lastTime.IsZero() {
		t.Error("Expected last command time to be set")
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
