package adapter

import (
	"context"
	"testing"
)

// MockAdapter implements IRadioAdapter for testing.
// This ensures the interface is complete and can be implemented.
type MockAdapter struct {
	AdapterBase
	state    *RadioState
	profiles []FrequencyProfile
	powerDbm float64
}

// NewMockAdapter creates a new mock adapter for testing.
func NewMockAdapter(radioID, model string) *MockAdapter {
	return &MockAdapter{
		AdapterBase: AdapterBase{
			RadioID: radioID,
			Model:   model,
			Status:  "online",
		},
		state: &RadioState{
			PowerDbm:     30.0,
			FrequencyMhz: 2412.0,
		},
		profiles: []FrequencyProfile{
			{
				Frequencies: []float64{2412.0, 2417.0, 2422.0},
				Bandwidth:   20.0,
				AntennaMask: 1,
			},
		},
		powerDbm: 30.0,
	}
}

// GetState implements IRadioAdapter.GetState
func (m *MockAdapter) GetState(ctx context.Context) (*RadioState, error) {
	return m.state, nil
}

// SetPower implements IRadioAdapter.SetPower
func (m *MockAdapter) SetPower(ctx context.Context, dBm float64) error {
	if dBm < 0 || dBm > 39 {
		return ErrInvalidRange
	}
	m.powerDbm = dBm
	m.state.PowerDbm = dBm
	return nil
}

// SetFrequency implements IRadioAdapter.SetFrequency
func (m *MockAdapter) SetFrequency(ctx context.Context, frequencyMhz float64) error {
	if frequencyMhz < 2400.0 || frequencyMhz > 2500.0 {
		return ErrInvalidRange
	}
	m.state.FrequencyMhz = frequencyMhz
	return nil
}

// ReadPowerActual implements IRadioAdapter.ReadPowerActual
func (m *MockAdapter) ReadPowerActual(ctx context.Context) (float64, error) {
	return m.powerDbm, nil
}

// SupportedFrequencyProfiles implements IRadioAdapter.SupportedFrequencyProfiles
func (m *MockAdapter) SupportedFrequencyProfiles(ctx context.Context) ([]FrequencyProfile, error) {
	return m.profiles, nil
}

// TestIRadioAdapterInterface ensures the interface is complete and implementable.
func TestIRadioAdapterInterface(t *testing.T) {
	// This test ensures compile-time checking that MockAdapter implements IRadioAdapter
	var _ IRadioAdapter = (*MockAdapter)(nil)

	// Test that we can create and use the interface
	adapter := NewMockAdapter("test-radio", "Test-Model")

	ctx := context.Background()

	// Test GetState
	state, err := adapter.GetState(ctx)
	if err != nil {
		t.Errorf("GetState failed: %v", err)
	}
	if state == nil {
		t.Error("GetState returned nil state")
	}
	if state.PowerDbm != 30.0 {
		t.Errorf("GetState returned power %f, want 30.0", state.PowerDbm)
	}

	// Test SetPower
	err = adapter.SetPower(ctx, 25.0)
	if err != nil {
		t.Errorf("SetPower failed: %v", err)
	}

	// Test SetPower with invalid range
	err = adapter.SetPower(ctx, 50)
	if err != ErrInvalidRange {
		t.Errorf("SetPower with invalid range returned %v, want ErrInvalidRange", err)
	}

	// Test SetFrequency
	err = adapter.SetFrequency(ctx, 2422.0)
	if err != nil {
		t.Errorf("SetFrequency failed: %v", err)
	}

	// Test SetFrequency with invalid range
	err = adapter.SetFrequency(ctx, 2000.0)
	if err != ErrInvalidRange {
		t.Errorf("SetFrequency with invalid range returned %v, want ErrInvalidRange", err)
	}

	// Test ReadPowerActual
	power, err := adapter.ReadPowerActual(ctx)
	if err != nil {
		t.Errorf("ReadPowerActual failed: %v", err)
	}
	if power != 25.0 {
		t.Errorf("ReadPowerActual returned %f, want 25.0", power)
	}

	// Test SupportedFrequencyProfiles
	profiles, err := adapter.SupportedFrequencyProfiles(ctx)
	if err != nil {
		t.Errorf("SupportedFrequencyProfiles failed: %v", err)
	}
	if len(profiles) != 1 {
		t.Errorf("SupportedFrequencyProfiles returned %d profiles, want 1", len(profiles))
	}
}

// TestAdapterBase ensures the base adapter functionality works.
func TestAdapterBase(t *testing.T) {
	base := &AdapterBase{
		RadioID: "test-radio",
		Model:   "Test-Model",
		Status:  "online",
	}

	if base.GetRadioID() != "test-radio" {
		t.Errorf("GetRadioID returned %s, want test-radio", base.GetRadioID())
	}

	if base.GetModel() != "Test-Model" {
		t.Errorf("GetModel returned %s, want Test-Model", base.GetModel())
	}

	if base.GetStatus() != "online" {
		t.Errorf("GetStatus returned %s, want online", base.GetStatus())
	}

	base.SetStatus("offline")
	if base.GetStatus() != "offline" {
		t.Errorf("GetStatus after SetStatus returned %s, want offline", base.GetStatus())
	}
}

// TestRadioState ensures the RadioState struct works correctly.
func TestRadioState(t *testing.T) {
	state := &RadioState{
		PowerDbm:     30.0,
		FrequencyMhz: 2412.0,
	}

	if state.PowerDbm != 30.0 {
		t.Errorf("PowerDbm = %f, want 30.0", state.PowerDbm)
	}

	if state.FrequencyMhz != 2412.0 {
		t.Errorf("FrequencyMhz = %f, want 2412.0", state.FrequencyMhz)
	}
}

// TestRadioCapabilities ensures the RadioCapabilities struct works correctly.
func TestRadioCapabilities(t *testing.T) {
	capabilities := &RadioCapabilities{
		MinPowerDbm: 0,
		MaxPowerDbm: 39,
		Channels: []Channel{
			{Index: 1, FrequencyMhz: 2412.0},
			{Index: 2, FrequencyMhz: 2417.0},
		},
	}

	if capabilities.MinPowerDbm != 0 {
		t.Errorf("MinPowerDbm = %d, want 0", capabilities.MinPowerDbm)
	}

	if capabilities.MaxPowerDbm != 39 {
		t.Errorf("MaxPowerDbm = %d, want 39", capabilities.MaxPowerDbm)
	}

	if len(capabilities.Channels) != 2 {
		t.Errorf("Channels length = %d, want 2", len(capabilities.Channels))
	}
}

// TestFrequencyProfile ensures the FrequencyProfile struct works correctly.
func TestFrequencyProfile(t *testing.T) {
	profile := &FrequencyProfile{
		Frequencies: []float64{2412.0, 2417.0, 2422.0},
		Bandwidth:   20.0,
		AntennaMask: 1,
	}

	if len(profile.Frequencies) != 3 {
		t.Errorf("Frequencies length = %d, want 3", len(profile.Frequencies))
	}

	if profile.Bandwidth != 20.0 {
		t.Errorf("Bandwidth = %f, want 20.0", profile.Bandwidth)
	}

	if profile.AntennaMask != 1 {
		t.Errorf("AntennaMask = %d, want 1", profile.AntennaMask)
	}
}
