// Package silvusmock provides a Silvus-like mock adapter for testing and development.
//
//   - PRE-INT-08: "Simulate Silvus-like behavior now"
//
package silvusmock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
)

// SilvusMock implements IRadioAdapter with Silvus-like behavior for testing.
type SilvusMock struct {
	adapter.AdapterBase

	// In-memory state
	mu              sync.RWMutex
	powerDbm        float64
	frequencyMhz    float64
	channelIndex    int
	bandPlan        []adapter.Channel
	lastCommandTime time.Time

	// Fault injection modes
	faultMode string // "ReturnBusy", "ReturnUnavailable", "ReturnInvalidRange", ""

	// Configuration
	minPower   int
	maxPower   int
	validFreqs []float64
}

// NewSilvusMock creates a new SilvusMock adapter.
func NewSilvusMock(radioID string, bandPlan []adapter.Channel) *SilvusMock {
	// Default band plan if none provided
	if bandPlan == nil {
		bandPlan = []adapter.Channel{
			{Index: 1, FrequencyMhz: 2412.0},
			{Index: 2, FrequencyMhz: 2417.0},
			{Index: 3, FrequencyMhz: 2422.0},
			{Index: 4, FrequencyMhz: 2427.0},
			{Index: 5, FrequencyMhz: 2432.0},
		}
	}

	// Extract valid frequencies from band plan
	validFreqs := make([]float64, len(bandPlan))
	for i, channel := range bandPlan {
		validFreqs[i] = channel.FrequencyMhz
	}

	return &SilvusMock{
		AdapterBase: adapter.AdapterBase{
			RadioID: radioID,
			Model:   "SilvusMock-Test",
			Status:  "online",
		},
		powerDbm:        20,     // Default power
		frequencyMhz:    2412.0, // Default frequency
		channelIndex:    1,      // Default channel
		bandPlan:        bandPlan,
		lastCommandTime: time.Now(),
		minPower:        0,
		maxPower:        39,
		validFreqs:      validFreqs,
		faultMode:       "", // No faults by default
	}
}

// GetState returns the current radio state.
func (s *SilvusMock) GetState(ctx context.Context) (*adapter.RadioState, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Check for fault injection
	if err := s.checkFaultMode("GetState"); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return &adapter.RadioState{
		PowerDbm:     s.powerDbm,
		FrequencyMhz: s.frequencyMhz,
	}, nil
}

// SetPower sets the transmit power in dBm.
func (s *SilvusMock) SetPower(ctx context.Context, dBm float64) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Check for fault injection
	if err := s.checkFaultMode("SetPower"); err != nil {
		return err
	}

	// Validate power range
	if dBm < float64(s.minPower) || dBm > float64(s.maxPower) {
		return fmt.Errorf("INVALID_RANGE: power %f is outside valid range [%d, %d]", dBm, s.minPower, s.maxPower)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.powerDbm = dBm
	s.lastCommandTime = time.Now()
	return nil
}

// SetFrequency sets the transmit frequency in MHz.
func (s *SilvusMock) SetFrequency(ctx context.Context, frequencyMhz float64) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Check for fault injection
	if err := s.checkFaultMode("SetFrequency"); err != nil {
		return err
	}

	// Validate frequency
	if frequencyMhz <= 0 {
		return fmt.Errorf("INVALID_RANGE: frequency %.1f is invalid", frequencyMhz)
	}

	// Check if frequency is in valid range
	if frequencyMhz < 100 || frequencyMhz > 6000 {
		return fmt.Errorf("INVALID_RANGE: frequency %.1f is outside valid range [100, 6000]", frequencyMhz)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.frequencyMhz = frequencyMhz
	s.lastCommandTime = time.Now()

	// Update channel index based on frequency
	s.updateChannelIndex(frequencyMhz)
	return nil
}

// ReadPowerActual reads the current power setting.
func (s *SilvusMock) ReadPowerActual(ctx context.Context) (float64, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	// Check for fault injection
	if err := s.checkFaultMode("ReadPowerActual"); err != nil {
		return 0, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.powerDbm, nil
}

// SupportedFrequencyProfiles returns allowed frequency/bandwidth/antenna combinations.
func (s *SilvusMock) SupportedFrequencyProfiles(ctx context.Context) ([]adapter.FrequencyProfile, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Check for fault injection
	if err := s.checkFaultMode("SupportedFrequencyProfiles"); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return frequency profiles based on band plan
	return []adapter.FrequencyProfile{
		{
			Frequencies: s.validFreqs,
			Bandwidth:   20.0,
			AntennaMask: 1,
		},
	}, nil
}

// Fault injection methods

// SetFaultMode sets the fault injection mode.
func (s *SilvusMock) SetFaultMode(mode string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.faultMode = mode
}

// ClearFaultMode clears the fault injection mode.
func (s *SilvusMock) ClearFaultMode() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.faultMode = ""
}

// checkFaultMode checks if a fault should be injected for the current operation.
func (s *SilvusMock) checkFaultMode(operation string) error {
	s.mu.RLock()
	mode := s.faultMode
	s.mu.RUnlock()

	switch mode {
	case "ReturnBusy":
		return fmt.Errorf("BUSY: SilvusMock simulated busy error for %s", operation)
	case "ReturnUnavailable":
		return fmt.Errorf("UNAVAILABLE: SilvusMock simulated unavailable error for %s", operation)
	case "ReturnInvalidRange":
		return fmt.Errorf("INVALID_RANGE: SilvusMock simulated invalid range error for %s", operation)
	default:
		return nil
	}
}

// updateChannelIndex updates the channel index based on the current frequency.
func (s *SilvusMock) updateChannelIndex(frequencyMhz float64) {
	for _, channel := range s.bandPlan {
		if channel.FrequencyMhz == frequencyMhz {
			s.channelIndex = channel.Index
			return
		}
	}
	// If no exact match, set to 0 (unknown channel)
	s.channelIndex = 0
}

// Helper methods for testing

// GetCurrentState returns the current internal state (for testing).
func (s *SilvusMock) GetCurrentState() (float64, float64, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.powerDbm, s.frequencyMhz, s.channelIndex
}

// SetCurrentState sets the current internal state (for testing).
func (s *SilvusMock) SetCurrentState(power float64, frequency float64, channel int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.powerDbm = power
	s.frequencyMhz = frequency
	s.channelIndex = channel
	s.lastCommandTime = time.Now()
}

// GetBandPlan returns the current band plan.
func (s *SilvusMock) GetBandPlan() []adapter.Channel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bandPlan
}

// SetBandPlan sets a new band plan.
func (s *SilvusMock) SetBandPlan(bandPlan []adapter.Channel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bandPlan = bandPlan

	// Update valid frequencies
	s.validFreqs = make([]float64, len(bandPlan))
	for i, channel := range bandPlan {
		s.validFreqs[i] = channel.FrequencyMhz
	}
}

// GetLastCommandTime returns the time of the last command.
func (s *SilvusMock) GetLastCommandTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastCommandTime
}

// SimulateSilvusBehavior simulates Silvus-specific behavior patterns.
func (s *SilvusMock) SimulateSilvusBehavior() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Simulate Silvus-specific initialization
	s.Status = "online"
	s.lastCommandTime = time.Now()
}
