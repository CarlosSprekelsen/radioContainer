// Package adapter defines IRadioAdapter interface from Architecture §5.
//
//   - Architecture §8.5: "Error normalization to INVALID_RANGE, BUSY, UNAVAILABLE, INTERNAL"
package adapter

import (
	"context"
)

// RadioState represents the current state of a radio.
type RadioState struct {
	PowerDbm     float64 `json:"powerDbm"`
	FrequencyMhz float64 `json:"frequencyMhz"`
}

// RadioCapabilities represents the capabilities of a radio.
type RadioCapabilities struct {
	MinPowerDbm int       `json:"minPowerDbm"`
	MaxPowerDbm int       `json:"maxPowerDbm"`
	Channels    []Channel `json:"channels"`
}

// Channel represents a single channel mapping.
type Channel struct {
	Index        int     `json:"index"`
	FrequencyMhz float64 `json:"frequencyMhz"`
}

// FrequencyProfile represents a supported frequency profile.
type FrequencyProfile struct {
	Frequencies []float64 `json:"frequencies"`
	Bandwidth   float64   `json:"bandwidth"`
	AntennaMask int       `json:"antenna_mask"`
}

// IRadioAdapter defines the stable southbound adapter contract.
type IRadioAdapter interface {
	// GetState returns the current radio state.
	GetState(ctx context.Context) (*RadioState, error)

	// SetPower sets the transmit power in dBm.
	// Params: dBm (0-39, accuracy 10-39)
	SetPower(ctx context.Context, dBm float64) error

	// SetFrequency sets the transmit frequency in MHz.
	// Params: frequencyMhz (0.1 MHz resolution)
	// Side-effect: soft boot - driver/services reboot
	SetFrequency(ctx context.Context, frequencyMhz float64) error

	// ReadPowerActual reads the current power setting.
	ReadPowerActual(ctx context.Context) (float64, error)

	// SupportedFrequencyProfiles returns allowed frequency/bandwidth/antenna combinations.
	SupportedFrequencyProfiles(ctx context.Context) ([]FrequencyProfile, error)
}

// AdapterBase provides common functionality for adapter implementations.
type AdapterBase struct {
	// RadioID identifies the radio this adapter controls
	RadioID string

	// Model identifies the radio model
	Model string

	// Status indicates the current radio status
	Status string
}

// GetRadioID returns the radio identifier.
func (a *AdapterBase) GetRadioID() string {
	return a.RadioID
}

// GetModel returns the radio model.
func (a *AdapterBase) GetModel() string {
	return a.Model
}

// GetStatus returns the radio status.
func (a *AdapterBase) GetStatus() string {
	return a.Status
}

// SetStatus updates the radio status.
func (a *AdapterBase) SetStatus(status string) {
	a.Status = status
}
