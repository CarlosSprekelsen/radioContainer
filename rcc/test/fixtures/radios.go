package fixtures

import (
	"time"

	"github.com/radio-control/rcc/internal/config"
)

// RadioProfile represents a standardized radio configuration for testing
type RadioProfile struct {
	ID           string
	Type         string
	Model        string
	Capabilities []string
	Channels     []ChannelProfile
	PowerRange   PowerRange
	Status       string
}

type ChannelProfile struct {
	Index     int
	Frequency float64
	Band      string
	Type      string
}

type PowerRange struct {
	Min int
	Max int
}

// StandardSilvusRadio returns a standard Silvus radio profile for testing
func StandardSilvusRadio() RadioProfile {
	return RadioProfile{
		ID:           "silvus-001",
		Type:         "Silvus",
		Model:        "Scorpion",
		Capabilities: []string{"setPower", "setChannel", "getState"},
		Channels:     WiFi24GHzChannels(),
		PowerRange:   PowerRange{Min: 1, Max: 10},
		Status:       "online",
	}
}

// MultiRadioSetup returns a multi-radio configuration for testing
func MultiRadioSetup() []RadioProfile {
	return []RadioProfile{
		StandardSilvusRadio(),
		{
			ID:           "silvus-002",
			Type:         "Silvus",
			Model:        "Scorpion",
			Capabilities: []string{"setPower", "setChannel", "getState"},
			Channels:     UHFChannels(),
			PowerRange:   PowerRange{Min: 1, Max: 10},
			Status:       "online",
		},
		{
			ID:           "generic-001",
			Type:         "Generic",
			Model:        "TestRadio",
			Capabilities: []string{"getState"},
			Channels:     WiFi5GHzChannels(),
			PowerRange:   PowerRange{Min: 1, Max: 5},
			Status:       "offline",
		},
	}
}

// OfflineRadio returns a radio in offline state for testing
func OfflineRadio() RadioProfile {
	radio := StandardSilvusRadio()
	radio.ID = "offline-001"
	radio.Status = "offline"
	return radio
}

// BusyRadio returns a radio in busy state for testing
func BusyRadio() RadioProfile {
	radio := StandardSilvusRadio()
	radio.ID = "busy-001"
	radio.Status = "busy"
	return radio
}

// LoadTestConfig returns a test configuration with deterministic values
func LoadTestConfig() *config.TimingConfig {
	return &config.TimingConfig{
		HeartbeatInterval:         15 * time.Second,
		HeartbeatJitter:           2 * time.Second,
		HeartbeatTimeout:          45 * time.Second,
		CommandTimeoutSetPower:    10 * time.Second,
		CommandTimeoutSetChannel:  30 * time.Second,
		CommandTimeoutSelectRadio: 5 * time.Second,
		CommandTimeoutGetState:    5 * time.Second,
		EventBufferSize:           50,
		EventBufferRetention:      1 * time.Hour,
	}
}
