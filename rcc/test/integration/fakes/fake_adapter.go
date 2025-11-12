//go:build integration

package fakes

import (
	"context"
	"fmt"
	"sync"

	"github.com/radio-control/rcc/internal/adapter"
)

// FakeAdapter implements adapter.IRadioAdapter with configurable behavior modes.
type FakeAdapter struct {
	adapter.AdapterBase

	// State
	mu           sync.RWMutex
	powerDbm     float64
	frequencyMhz float64
	bandPlan     []adapter.Channel

	// Behavior modes
	mode string // "happy", "busy", "invalid-range", "unavailable", "internal"

	// Call tracking for assertions
	lastSetPowerCall     float64
	lastSetFrequencyCall float64
	callCount            map[string]int
}

// Compile-time assertion that FakeAdapter implements adapter.IRadioAdapter
var _ adapter.IRadioAdapter = (*FakeAdapter)(nil)

// NewFakeAdapter creates a new fake adapter in "happy" mode.
func NewFakeAdapter(radioID string) *FakeAdapter {
	return &FakeAdapter{
		AdapterBase: adapter.AdapterBase{
			RadioID: radioID,
			Model:   "FakeAdapter-Test",
			Status:  "online",
		},
		powerDbm:     20.0,
		frequencyMhz: 2412.0,
		bandPlan: []adapter.Channel{
			{Index: 1, FrequencyMhz: 2412.0},
			{Index: 6, FrequencyMhz: 2437.0},
			{Index: 11, FrequencyMhz: 2462.0},
		},
		mode:      "happy",
		callCount: make(map[string]int),
	}
}

// WithMode sets the behavior mode for the adapter.
func (f *FakeAdapter) WithMode(mode string) *FakeAdapter {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.mode = mode
	return f
}

// WithInitial sets the initial state of the adapter.
func (f *FakeAdapter) WithInitial(powerDbm float64, freqMhz float64, band []adapter.Channel) *FakeAdapter {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.powerDbm = powerDbm
	f.frequencyMhz = freqMhz
	if band != nil {
		f.bandPlan = band
	}
	return f
}

// GetState returns the current radio state.
func (f *FakeAdapter) GetState(ctx context.Context) (*adapter.RadioState, error) {
	f.mu.Lock()
	f.callCount["GetState"]++
	f.mu.Unlock()

	if err := f.checkMode("GetState"); err != nil {
		return nil, err
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	return &adapter.RadioState{
		PowerDbm:     f.powerDbm,
		FrequencyMhz: f.frequencyMhz,
	}, nil
}

// SetPower sets the transmit power in dBm.
func (f *FakeAdapter) SetPower(ctx context.Context, dBm float64) error {
	f.mu.Lock()
	f.callCount["SetPower"]++
	f.lastSetPowerCall = dBm
	f.mu.Unlock()

	if err := f.checkMode("SetPower"); err != nil {
		return err
	}

	// Validate power range
	if dBm < 0 || dBm > 39 {
		return fmt.Errorf("INVALID_RANGE: power %f is outside valid range [0, 39]", dBm)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.powerDbm = dBm
	return nil
}

// SetFrequency sets the transmit frequency in MHz.
func (f *FakeAdapter) SetFrequency(ctx context.Context, frequencyMhz float64) error {
	f.mu.Lock()
	f.callCount["SetFrequency"]++
	f.lastSetFrequencyCall = frequencyMhz
	f.mu.Unlock()

	if err := f.checkMode("SetFrequency"); err != nil {
		return err
	}

	// Validate frequency
	if frequencyMhz <= 0 {
		return fmt.Errorf("INVALID_RANGE: frequency %.1f is invalid", frequencyMhz)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.frequencyMhz = frequencyMhz
	return nil
}

// ReadPowerActual reads the current power setting.
func (f *FakeAdapter) ReadPowerActual(ctx context.Context) (float64, error) {
	f.mu.Lock()
	f.callCount["ReadPowerActual"]++
	f.mu.Unlock()

	if err := f.checkMode("ReadPowerActual"); err != nil {
		return 0, err
	}

	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.powerDbm, nil
}

// SupportedFrequencyProfiles returns allowed frequency/bandwidth/antenna combinations.
func (f *FakeAdapter) SupportedFrequencyProfiles(ctx context.Context) ([]adapter.FrequencyProfile, error) {
	f.mu.Lock()
	f.callCount["SupportedFrequencyProfiles"]++
	f.mu.Unlock()

	if err := f.checkMode("SupportedFrequencyProfiles"); err != nil {
		return nil, err
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	// Extract frequencies from band plan
	frequencies := make([]float64, len(f.bandPlan))
	for i, channel := range f.bandPlan {
		frequencies[i] = channel.FrequencyMhz
	}

	return []adapter.FrequencyProfile{
		{
			Frequencies: frequencies,
			Bandwidth:   20.0,
			AntennaMask: 1,
		},
	}, nil
}

// checkMode checks if a fault should be injected for the current operation.
func (f *FakeAdapter) checkMode(operation string) error {
	f.mu.RLock()
	mode := f.mode
	f.mu.RUnlock()

	switch mode {
	case "busy":
		return fmt.Errorf("BUSY: FakeAdapter simulated busy error for %s", operation)
	case "unavailable":
		return fmt.Errorf("UNAVAILABLE: FakeAdapter simulated unavailable error for %s", operation)
	case "invalid-range":
		return fmt.Errorf("INVALID_RANGE: FakeAdapter simulated invalid range error for %s", operation)
	case "internal":
		return fmt.Errorf("INTERNAL: FakeAdapter simulated internal error for %s", operation)
	default:
		return nil
	}
}

// GetCallCount returns the number of times a method was called.
func (f *FakeAdapter) GetCallCount(method string) int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.callCount[method]
}

// GetLastSetPowerCall returns the last power value passed to SetPower.
func (f *FakeAdapter) GetLastSetPowerCall() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.lastSetPowerCall
}

// GetLastSetFrequencyCall returns the last frequency value passed to SetFrequency.
func (f *FakeAdapter) GetLastSetFrequencyCall() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.lastSetFrequencyCall
}

// GetBandPlan returns the current band plan.
func (f *FakeAdapter) GetBandPlan() []adapter.Channel {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result := make([]adapter.Channel, len(f.bandPlan))
	copy(result, f.bandPlan)
	return result
}

// GetCurrentState returns the current internal state (for testing).
func (f *FakeAdapter) GetCurrentState() (float64, float64) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.powerDbm, f.frequencyMhz
}

// SetMode sets the behavior mode for testing different error conditions.
func (f *FakeAdapter) SetMode(mode string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.mode = mode
}
