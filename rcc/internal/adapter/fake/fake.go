// Package fake provides a fake radio adapter implementation for testing.
//
//   - RE-INT-03: "Any adapter (including Silvus) must pass a standard test suite"
package fake

import (
	"context"
	"fmt"

	"github.com/radio-control/rcc/internal/adapter"
)

// FakeAdapter implements IRadioAdapter for testing purposes.
type FakeAdapter struct {
	adapter.AdapterBase

	// Current state
	currentPower     float64
	currentFrequency float64

	// Configuration
	minPower   int
	maxPower   int
	validFreqs []float64
	channels   []adapter.Channel

	// Error simulation
	simulateErrors bool
	errorType      string
}

// NewFakeAdapter creates a new fake adapter for testing.
func NewFakeAdapter(radioID string) *FakeAdapter {
	return &FakeAdapter{
		AdapterBase: adapter.AdapterBase{
			RadioID: radioID,
			Model:   "Fake-Radio-Test",
			Status:  "online",
		},
		currentPower:     20,
		currentFrequency: 2412.0,
		minPower:         0,
		maxPower:         39,
		validFreqs:       []float64{2412.0, 2417.0, 2422.0, 2427.0, 2432.0},
		channels: []adapter.Channel{
			{Index: 1, FrequencyMhz: 2412.0},
			{Index: 2, FrequencyMhz: 2417.0},
			{Index: 3, FrequencyMhz: 2422.0},
			{Index: 4, FrequencyMhz: 2427.0},
			{Index: 5, FrequencyMhz: 2432.0},
		},
		simulateErrors: false,
	}
}

// GetState returns the current radio state.
func (f *FakeAdapter) GetState(ctx context.Context) (*adapter.RadioState, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if f.simulateErrors {
		return nil, f.getSimulatedError()
	}

	return &adapter.RadioState{
		PowerDbm:     f.currentPower,
		FrequencyMhz: f.currentFrequency,
	}, nil
}

// SetPower sets the transmit power in dBm.
func (f *FakeAdapter) SetPower(ctx context.Context, dBm float64) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if f.simulateErrors {
		return f.getSimulatedError()
	}

	// Validate power range
	if dBm < float64(f.minPower) || dBm > float64(f.maxPower) {
		return fmt.Errorf("INVALID_RANGE: power %f is outside valid range [%d, %d]", dBm, f.minPower, f.maxPower)
	}

	f.currentPower = dBm
	return nil
}

// SetFrequency sets the transmit frequency in MHz.
func (f *FakeAdapter) SetFrequency(ctx context.Context, frequencyMhz float64) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if f.simulateErrors {
		return f.getSimulatedError()
	}

	// Validate frequency
	if frequencyMhz <= 0 {
		return fmt.Errorf("INVALID_RANGE: frequency %.1f is invalid", frequencyMhz)
	}

	// Check if frequency is in valid range (basic validation)
	if frequencyMhz < 100 || frequencyMhz > 6000 {
		return fmt.Errorf("INVALID_RANGE: frequency %.1f is outside valid range [100, 6000]", frequencyMhz)
	}

	f.currentFrequency = frequencyMhz
	return nil
}

// ReadPowerActual reads the current power setting.
func (f *FakeAdapter) ReadPowerActual(ctx context.Context) (float64, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	if f.simulateErrors {
		return 0, f.getSimulatedError()
	}

	return f.currentPower, nil
}

// SupportedFrequencyProfiles returns allowed frequency/bandwidth/antenna combinations.
func (f *FakeAdapter) SupportedFrequencyProfiles(ctx context.Context) ([]adapter.FrequencyProfile, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if f.simulateErrors {
		return nil, f.getSimulatedError()
	}

	// Return a simple frequency profile
	return []adapter.FrequencyProfile{
		{
			Frequencies: f.validFreqs,
			Bandwidth:   20.0,
			AntennaMask: 1,
		},
	}, nil
}

// Helper methods for testing

// SetErrorSimulation enables error simulation for testing.
func (f *FakeAdapter) SetErrorSimulation(errorType string) {
	f.simulateErrors = true
	f.errorType = errorType
}

// DisableErrorSimulation disables error simulation.
func (f *FakeAdapter) DisableErrorSimulation() {
	f.simulateErrors = false
	f.errorType = ""
}

// getSimulatedError returns a simulated error based on the configured error type.
func (f *FakeAdapter) getSimulatedError() error {
	switch f.errorType {
	case "INVALID_RANGE":
		return fmt.Errorf("INVALID_RANGE: simulated range error")
	case "BUSY":
		return fmt.Errorf("BUSY: simulated busy error")
	case "UNAVAILABLE":
		return fmt.Errorf("UNAVAILABLE: simulated unavailable error")
	case "INTERNAL":
		return fmt.Errorf("INTERNAL: simulated internal error")
	default:
		return fmt.Errorf("INTERNAL: unknown simulated error")
	}
}

// GetCurrentState returns the current internal state (for testing).
func (f *FakeAdapter) GetCurrentState() (float64, float64) {
	return f.currentPower, f.currentFrequency
}

// SetCurrentState sets the current internal state (for testing).
func (f *FakeAdapter) SetCurrentState(power float64, frequency float64) {
	f.currentPower = power
	f.currentFrequency = frequency
}
