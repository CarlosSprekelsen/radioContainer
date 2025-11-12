package config

import (
	"fmt"
	"time"
)

// TimingConfig maps CB-TIMING v0.3 structure.
type TimingConfig struct {
	// CB-TIMING §3.1 Heartbeat Configuration
	HeartbeatInterval time.Duration
	HeartbeatJitter   time.Duration
	HeartbeatTimeout  time.Duration

	// CB-TIMING §4.1 Probe States & Cadences
	ProbeNormalInterval    time.Duration
	ProbeRecoveringInitial time.Duration
	ProbeRecoveringBackoff float64
	ProbeRecoveringMax     time.Duration
	ProbeOfflineInitial    time.Duration
	ProbeOfflineBackoff    float64
	ProbeOfflineMax        time.Duration

	// CB-TIMING §5 Command Timeout Classes
	CommandTimeoutSetPower    time.Duration
	CommandTimeoutSetChannel  time.Duration
	CommandTimeoutSelectRadio time.Duration
	CommandTimeoutGetState    time.Duration

	// CB-TIMING §6.1 Event Buffer Configuration
	EventBufferSize      int
	EventBufferRetention time.Duration

	// PRE-INT-09: Silvus Band Plan Configuration
	SilvusBandPlan *SilvusBandPlan
}

// SilvusBandPlan represents Silvus radio band plan configuration.
type SilvusBandPlan struct {
	// Band plans organized by model and band
	Models map[string]map[string][]SilvusChannel `json:"models"`
}

// SilvusChannel represents a single channel in a Silvus band plan.
type SilvusChannel struct {
	ChannelIndex int     `json:"channelIndex"`
	FrequencyMhz float64 `json:"frequencyMhz"`
}

// LoadCBTimingBaseline returns CB-TIMING v0.3 baseline values.
func LoadCBTimingBaseline() *TimingConfig {
	return &TimingConfig{
		// CB-TIMING §3.1: Heartbeat interval 15s, jitter ±2s, timeout 45s
		HeartbeatInterval: 15 * time.Second, // CB-TIMING §3.1
		HeartbeatJitter:   2 * time.Second,  // CB-TIMING §3.1
		HeartbeatTimeout:  45 * time.Second, // CB-TIMING §3.1

		// CB-TIMING §4.1: Normal 30s, Recovering 5s/1.5x/15s, Offline 10s/2.0x/300s
		ProbeNormalInterval:    30 * time.Second,  // CB-TIMING §4.1
		ProbeRecoveringInitial: 5 * time.Second,   // CB-TIMING §4.1
		ProbeRecoveringBackoff: 1.5,               // CB-TIMING §4.1
		ProbeRecoveringMax:     15 * time.Second,  // CB-TIMING §4.1
		ProbeOfflineInitial:    10 * time.Second,  // CB-TIMING §4.1
		ProbeOfflineBackoff:    2.0,               // CB-TIMING §4.1
		ProbeOfflineMax:        300 * time.Second, // CB-TIMING §4.1

		// CB-TIMING §5: setPower 10s, setChannel 30s, selectRadio 5s, getState 5s
		CommandTimeoutSetPower:    10 * time.Second, // CB-TIMING §5
		CommandTimeoutSetChannel:  30 * time.Second, // CB-TIMING §5
		CommandTimeoutSelectRadio: 5 * time.Second,  // CB-TIMING §5
		CommandTimeoutGetState:    5 * time.Second,  // CB-TIMING §5

		// CB-TIMING §6.1: 50 events, 1 hour retention
		EventBufferSize:      50,            // CB-TIMING §6.1
		EventBufferRetention: 1 * time.Hour, // CB-TIMING §6.1
	}
}

// GetSilvusChannelFrequency returns the frequency for a given channel index in a Silvus band plan.
func (sbp *SilvusBandPlan) GetSilvusChannelFrequency(model, band string, channelIndex int) (float64, error) {
	if sbp == nil || sbp.Models == nil {
		return 0, fmt.Errorf("no Silvus band plan configured")
	}

	modelBands, exists := sbp.Models[model]
	if !exists {
		return 0, fmt.Errorf("model %s not found in band plan", model)
	}

	channels, exists := modelBands[band]
	if !exists {
		return 0, fmt.Errorf("band %s not found for model %s", band, model)
	}

	for _, channel := range channels {
		if channel.ChannelIndex == channelIndex {
			return channel.FrequencyMhz, nil
		}
	}

	return 0, fmt.Errorf("channel index %d not found in model %s band %s", channelIndex, model, band)
}

// GetSilvusChannelIndex returns the channel index for a given frequency in a Silvus band plan.
func (sbp *SilvusBandPlan) GetSilvusChannelIndex(model, band string, frequencyMhz float64) (int, error) {
	if sbp == nil || sbp.Models == nil {
		return 0, fmt.Errorf("no Silvus band plan configured")
	}

	modelBands, exists := sbp.Models[model]
	if !exists {
		return 0, fmt.Errorf("model %s not found in band plan", model)
	}

	channels, exists := modelBands[band]
	if !exists {
		return 0, fmt.Errorf("band %s not found for model %s", band, model)
	}

	for _, channel := range channels {
		if channel.FrequencyMhz == frequencyMhz {
			return channel.ChannelIndex, nil
		}
	}

	return 0, fmt.Errorf("frequency %.1f MHz not found in model %s band %s", frequencyMhz, model, band)
}

// HasModelBand checks if a model and band combination exists in the band plan.
func (sbp *SilvusBandPlan) HasModelBand(model, band string) bool {
	if sbp == nil || sbp.Models == nil {
		return false
	}

	modelBands, exists := sbp.Models[model]
	if !exists {
		return false
	}

	_, exists = modelBands[band]
	return exists
}

// GetAvailableModels returns a list of available models in the band plan.
func (sbp *SilvusBandPlan) GetAvailableModels() []string {
	if sbp == nil || sbp.Models == nil {
		return []string{}
	}

	models := make([]string, 0, len(sbp.Models))
	for model := range sbp.Models {
		models = append(models, model)
	}
	return models
}

// GetAvailableBands returns a list of available bands for a given model.
func (sbp *SilvusBandPlan) GetAvailableBands(model string) []string {
	if sbp == nil || sbp.Models == nil {
		return []string{}
	}

	modelBands, exists := sbp.Models[model]
	if !exists {
		return []string{}
	}

	bands := make([]string, 0, len(modelBands))
	for band := range modelBands {
		bands = append(bands, band)
	}
	return bands
}
