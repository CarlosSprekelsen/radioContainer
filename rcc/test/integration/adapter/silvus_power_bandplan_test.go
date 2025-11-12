//go:build integration

package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/test/harness"
)

// TestSilvusPowerBandplan_ReadPowerActual tests ReadPowerActual() accuracy:
// Accuracy window vs setpoint; failure modes surfaced
func TestSilvusPowerBandplan_ReadPowerActual(t *testing.T) {
	// Arrange: Create harness with silvusmock adapter
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-power-test-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	ctx := context.Background()

	// Test cases for power accuracy
	tests := []struct {
		name        string
		setPower    float64
		expectedMin float64
		expectedMax float64
		expectError bool
		errorType   string
	}{
		{
			name:        "valid_power_range",
			setPower:    25.0,
			expectedMin: 23.0, // 2dB tolerance
			expectedMax: 27.0,
			expectError: false,
		},
		{
			name:        "high_power",
			setPower:    30.0,
			expectedMin: 28.0,
			expectedMax: 32.0,
			expectError: false,
		},
		{
			name:        "low_power",
			setPower:    10.0,
			expectedMin: 8.0,
			expectedMax: 12.0,
			expectError: false,
		},
		{
			name:        "invalid_power_negative",
			setPower:    -5.0,
			expectError: true,
			errorType:   "invalid range",
		},
		{
			name:        "invalid_power_too_high",
			setPower:    50.0,
			expectError: true,
			errorType:   "invalid range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act: Set power and read actual
			err := server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, tt.setPower)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else {
					t.Logf("✅ Got expected error: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("SetPower failed: %v", err)
			}

			// Read actual power
			actualPower, err := server.SilvusAdapter.ReadPowerActual(ctx)
			if err != nil {
				t.Fatalf("ReadPowerActual failed: %v", err)
			}

			// Assert: Power within expected accuracy window
			if actualPower < tt.expectedMin || actualPower > tt.expectedMax {
				t.Errorf("Actual power %v outside expected range [%v, %v]",
					actualPower, tt.expectedMin, tt.expectedMax)
			}

			t.Logf("✅ Power accuracy: set %v, actual %v (within [%v, %v])",
				tt.setPower, actualPower, tt.expectedMin, tt.expectedMax)
		})
	}
}

// TestSilvusPowerBandplan_SetBandPlan tests SetBandPlan():
// Valid/invalid bandplan per model; verify adapter → radio state and Orchestrator events
func TestSilvusPowerBandplan_SetBandPlan(t *testing.T) {
	// Arrange: Create harness with silvusmock adapter
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-bandplan-test-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	ctx := context.Background()

	// Test cases for band plan configuration
	tests := []struct {
		name        string
		bandPlan    adapter.BandPlan
		expectError bool
		errorType   string
	}{
		{
			name: "valid_2_4ghz_plan",
			bandPlan: adapter.BandPlan{
				Channels: []adapter.Channel{
					{Index: 1, FrequencyMhz: 2412.0},
					{Index: 6, FrequencyMhz: 2437.0},
					{Index: 11, FrequencyMhz: 2462.0},
				},
			},
			expectError: false,
		},
		{
			name: "valid_5ghz_plan",
			bandPlan: adapter.BandPlan{
				Channels: []adapter.Channel{
					{Index: 36, FrequencyMhz: 5180.0},
					{Index: 40, FrequencyMhz: 5200.0},
					{Index: 44, FrequencyMhz: 5220.0},
				},
			},
			expectError: false,
		},
		{
			name: "mixed_band_plan",
			bandPlan: adapter.BandPlan{
				Channels: []adapter.Channel{
					{Index: 1, FrequencyMhz: 2412.0},
					{Index: 6, FrequencyMhz: 2437.0},
					{Index: 36, FrequencyMhz: 5180.0},
					{Index: 40, FrequencyMhz: 5200.0},
				},
			},
			expectError: false,
		},
		{
			name: "invalid_frequency",
			bandPlan: adapter.BandPlan{
				Channels: []adapter.Channel{
					{Index: 1, FrequencyMhz: 1000.0}, // Invalid frequency
				},
			},
			expectError: true,
			errorType:   "invalid frequency",
		},
		{
			name: "empty_band_plan",
			bandPlan: adapter.BandPlan{
				Channels: []adapter.Channel{},
			},
			expectError: true,
			errorType:   "empty band plan",
		},
		{
			name: "duplicate_channels",
			bandPlan: adapter.BandPlan{
				Channels: []adapter.Channel{
					{Index: 1, FrequencyMhz: 2412.0},
					{Index: 1, FrequencyMhz: 2412.0}, // Duplicate
				},
			},
			expectError: true,
			errorType:   "duplicate channels",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act: Set band plan
			err := server.SilvusAdapter.SetBandPlan(ctx, tt.bandPlan)

			// Assert: Error handling
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else {
					t.Logf("✅ Got expected error: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("SetBandPlan failed: %v", err)
			}

			// Assert: Band plan was applied
			appliedPlan := server.SilvusAdapter.GetBandPlan()
			if len(appliedPlan.Channels) != len(tt.bandPlan.Channels) {
				t.Errorf("Expected %d channels, got %d",
					len(tt.bandPlan.Channels), len(appliedPlan.Channels))
			}

			// Verify channel frequencies match
			for i, expectedChannel := range tt.bandPlan.Channels {
				if i >= len(appliedPlan.Channels) {
					t.Errorf("Missing channel %d", i)
					continue
				}
				actualChannel := appliedPlan.Channels[i]
				if actualChannel.FrequencyMhz != expectedChannel.FrequencyMhz {
					t.Errorf("Channel %d: expected frequency %v, got %v",
						i, expectedChannel.FrequencyMhz, actualChannel.FrequencyMhz)
				}
			}

			t.Logf("✅ Band plan applied: %d channels", len(appliedPlan.Channels))
		})
	}
}

// TestSilvusPowerBandplan_Concurrency tests concurrency with mixed adapter set:
// Verify Command→RadioManager routing and error isolation
func TestSilvusPowerBandplan_Concurrency(t *testing.T) {
	// Arrange: Create harness with multiple adapters
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-concurrent-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Add fake adapter for comparison
	fakeAdapter := &FakeAdapterForConcurrency{radioID: "fake-concurrent-001"}
	err := server.RadioManager.LoadCapabilities("fake-concurrent-001", fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load fake adapter capabilities: %v", err)
	}

	ctx := context.Background()

	// Act: Execute concurrent operations on different adapters
	done := make(chan error, 4)

	// Operation 1: setPower on silvusmock adapter
	go func() {
		err := server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 25.0)
		done <- err
	}()

	// Operation 2: setPower on fake adapter
	go func() {
		err := server.Orchestrator.SetPower(ctx, "fake-concurrent-001", 30.0)
		done <- err
	}()

	// Operation 3: setChannel on silvusmock adapter
	go func() {
		err := server.Orchestrator.SetChannel(ctx, opts.ActiveRadioID, 2412.0)
		done <- err
	}()

	// Operation 4: setChannel on fake adapter
	go func() {
		err := server.Orchestrator.SetChannel(ctx, "fake-concurrent-001", 2437.0)
		done <- err
	}()

	// Wait for all operations
	var errors []error
	for i := 0; i < 4; i++ {
		if err := <-done; err != nil {
			errors = append(errors, err)
		}
	}

	// Assert: All operations succeed (no cross-adapter interference)
	if len(errors) > 0 {
		t.Errorf("Concurrent operations failed: %v", errors)
	}

	// Assert: Adapters maintain independent state
	silvusState, err := server.SilvusAdapter.GetState(ctx)
	if err != nil {
		t.Errorf("Failed to get silvusmock state: %v", err)
	}

	fakeState, err := fakeAdapter.GetState(ctx)
	if err != nil {
		t.Errorf("Failed to get fake adapter state: %v", err)
	}

	// Verify states are different (no cross-contamination)
	if silvusState.PowerDbm == fakeState.PowerDbm {
		t.Error("Adapters should maintain independent power states")
	}

	t.Logf("✅ Concurrency: silvusmock power %v, fake power %v (independent)",
		silvusState.PowerDbm, fakeState.PowerDbm)
}

// TestSilvusPowerBandplan_ErrorIsolation tests error isolation:
// Verify that errors from one adapter don't affect others
func TestSilvusPowerBandplan_ErrorIsolation(t *testing.T) {
	// Arrange: Create harness with error-injecting adapter
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-error-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Add error-injecting adapter
	errorAdapter := &ErrorInjectingAdapterForConcurrency{radioID: "error-001"}
	err := server.RadioManager.LoadCapabilities("error-001", errorAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load error adapter capabilities: %v", err)
	}

	ctx := context.Background()

	// Act: Execute operations on both adapters
	done := make(chan error, 2)

	// Operation 1: setPower on silvusmock adapter (should succeed)
	go func() {
		err := server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 25.0)
		done <- err
	}()

	// Operation 2: setPower on error adapter (should fail)
	go func() {
		err := server.Orchestrator.SetPower(ctx, "error-001", 30.0)
		done <- err
	}()

	// Wait for both operations
	var errors []error
	for i := 0; i < 2; i++ {
		if err := <-done; err != nil {
			errors = append(errors, err)
		}
	}

	// Assert: One operation succeeds, one fails (error isolation)
	if len(errors) != 1 {
		t.Errorf("Expected 1 error (error adapter), got %d errors: %v", len(errors), errors)
	}

	// Assert: Silvusmock adapter state is unaffected by error adapter
	silvusState, err := server.SilvusAdapter.GetState(ctx)
	if err != nil {
		t.Errorf("Failed to get silvusmock state: %v", err)
	}

	// Verify silvusmock adapter has the expected power
	if silvusState.PowerDbm != 25.0 {
		t.Errorf("Silvusmock power should be 25.0, got %v", silvusState.PowerDbm)
	}

	t.Logf("✅ Error isolation: silvusmock unaffected by error adapter failure")
}

// Helper types for testing

// FakeAdapterForConcurrency implements IRadioAdapter for concurrency testing
type FakeAdapterForConcurrency struct {
	radioID   string
	powerDbm  float64
	frequency float64
}

func (a *FakeAdapterForConcurrency) GetRadioID() string { return a.radioID }
func (a *FakeAdapterForConcurrency) SetPower(ctx context.Context, powerDbm float64) error {
	a.powerDbm = powerDbm
	return nil
}
func (a *FakeAdapterForConcurrency) SetChannel(ctx context.Context, frequencyMhz float64) error {
	a.frequency = frequencyMhz
	return nil
}
func (a *FakeAdapterForConcurrency) GetState(ctx context.Context) (adapter.RadioState, error) {
	return adapter.RadioState{
		PowerDbm:  a.powerDbm,
		Frequency: a.frequency,
		IsActive:  true,
	}, nil
}
func (a *FakeAdapterForConcurrency) ReadPowerActual(ctx context.Context) (float64, error) {
	return a.powerDbm, nil
}
func (a *FakeAdapterForConcurrency) GetSupportedFrequencyProfiles() []adapter.FrequencyProfile {
	return []adapter.FrequencyProfile{
		{Name: "2.4GHz", MinFreq: 2400, MaxFreq: 2500},
	}
}
func (a *FakeAdapterForConcurrency) SetBandPlan(ctx context.Context, plan adapter.BandPlan) error {
	return nil
}

// ErrorInjectingAdapterForConcurrency injects errors for testing
type ErrorInjectingAdapterForConcurrency struct {
	radioID string
}

func (a *ErrorInjectingAdapterForConcurrency) GetRadioID() string { return a.radioID }
func (a *ErrorInjectingAdapterForConcurrency) SetPower(ctx context.Context, powerDbm float64) error {
	return adapter.ErrUnavailable
}
func (a *ErrorInjectingAdapterForConcurrency) SetChannel(ctx context.Context, frequencyMhz float64) error {
	return adapter.ErrUnavailable
}
func (a *ErrorInjectingAdapterForConcurrency) GetState(ctx context.Context) (adapter.RadioState, error) {
	return adapter.RadioState{}, adapter.ErrUnavailable
}
func (a *ErrorInjectingAdapterForConcurrency) ReadPowerActual(ctx context.Context) (float64, error) {
	return 0, adapter.ErrUnavailable
}
func (a *ErrorInjectingAdapterForConcurrency) GetSupportedFrequencyProfiles() []adapter.FrequencyProfile {
	return nil
}
func (a *ErrorInjectingAdapterForConcurrency) SetBandPlan(ctx context.Context, plan adapter.BandPlan) error {
	return adapter.ErrUnavailable
}
