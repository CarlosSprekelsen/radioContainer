//
//
package radio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
)

// Radio represents a single radio with its capabilities and current state.
type Radio struct {
	ID           string                    `json:"id"`
	Model        string                    `json:"model"`
	Status       string                    `json:"status"`
	Capabilities *adapter.RadioCapabilities `json:"capabilities"`
	State        *adapter.RadioState       `json:"state"`
	LastSeen     time.Time                 `json:"lastSeen,omitempty"`
}

// RadioList represents the response format for GET /radios.
type RadioList struct {
	ActiveRadioID string  `json:"activeRadioId"`
	Items         []Radio `json:"items"`
}

// Manager manages radio inventory, capabilities, and active selection.
type Manager struct {
	mu            sync.RWMutex
	radios        map[string]*Radio
	activeRadioID string
	adapters      map[string]adapter.IRadioAdapter
}

// NewManager creates a new radio manager.
func NewManager() *Manager {
	return &Manager{
		radios:   make(map[string]*Radio),
		adapters: make(map[string]adapter.IRadioAdapter),
	}
}

// LoadCapabilities loads capabilities from an adapter on startup.
func (m *Manager) LoadCapabilities(radioID string, radioAdapter adapter.IRadioAdapter, timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store adapter
	m.adapters[radioID] = radioAdapter

	// Load capabilities from adapter
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	capabilities, err := radioAdapter.SupportedFrequencyProfiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to load capabilities for radio %s: %w", radioID, err)
	}

	// Get current state
	state, err := radioAdapter.GetState(ctx)
	if err != nil {
		// If we can't get state, create a default offline state
		state = &adapter.RadioState{
			PowerDbm:     0,
			FrequencyMhz: 0,
		}
	}

	// Create radio entry
	radio := &Radio{
		ID:     radioID,
		Model:  m.getModelFromCapabilities(capabilities),
		Status: m.determineStatus(err),
		Capabilities: &adapter.RadioCapabilities{
			MinPowerDbm: m.getMinPowerFromCapabilities(capabilities),
			MaxPowerDbm: m.getMaxPowerFromCapabilities(capabilities),
			Channels:    m.getChannelsFromCapabilities(capabilities, radioAdapter),
		},
		State:    state,
		LastSeen: time.Now(),
	}

	m.radios[radioID] = radio

	// Set as active if it's the first radio
	if m.activeRadioID == "" {
		m.activeRadioID = radioID
	}

	return nil
}

// SetActive sets the active radio with existence check.
func (m *Manager) SetActive(radioID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if radio exists
	if _, exists := m.radios[radioID]; !exists {
		return fmt.Errorf("radio %s not found", radioID)
	}

	m.activeRadioID = radioID
	return nil
}

// GetActive returns the active radio ID.
func (m *Manager) GetActive() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeRadioID
}

// GetActiveRadio returns the active radio object.
func (m *Manager) GetActiveRadio() *Radio {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.activeRadioID == "" {
		return nil
	}
	
	return m.radios[m.activeRadioID]
}

// GetActiveAdapter returns the adapter for the active radio.
func (m *Manager) GetActiveAdapter() (adapter.IRadioAdapter, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.activeRadioID == "" {
		return nil, "", fmt.Errorf("no active radio")
	}
	
	adapter, exists := m.adapters[m.activeRadioID]
	if !exists {
		return nil, "", fmt.Errorf("no adapter for active radio %s", m.activeRadioID)
	}
	
	return adapter, m.activeRadioID, nil
}

// List returns the radio list matching OpenAPI schema.
func (m *Manager) List() *RadioList {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]Radio, 0, len(m.radios))
	for _, radio := range m.radios {
		items = append(items, *radio)
	}

	return &RadioList{
		ActiveRadioID: m.activeRadioID,
		Items:         items,
	}
}

// GetRadio returns a specific radio by ID.
func (m *Manager) GetRadio(radioID string) (*Radio, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	radio, exists := m.radios[radioID]
	if !exists {
		return nil, fmt.Errorf("radio %s not found", radioID)
	}

	// Ensure capabilities are loaded from adapter if missing
	if radio.Capabilities == nil && m.adapters[radioID] != nil {
		radioAdapter := m.adapters[radioID]
		// Try to get channels directly from adapter if it supports it
		if bandPlanAdapter, ok := radioAdapter.(interface{ GetBandPlan() []adapter.Channel }); ok {
			channels := bandPlanAdapter.GetBandPlan()
			radio.Capabilities = &adapter.RadioCapabilities{
				Channels: channels,
			}
		}
	}

	return radio, nil
}

// UpdateState updates the state of a radio.
func (m *Manager) UpdateState(radioID string, state *adapter.RadioState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	radio, exists := m.radios[radioID]
	if !exists {
		return fmt.Errorf("radio %s not found", radioID)
	}

	radio.State = state
	radio.LastSeen = time.Now()
	radio.Status = "online"

	return nil
}

// UpdateStatus updates the status of a radio.
func (m *Manager) UpdateStatus(radioID string, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	radio, exists := m.radios[radioID]
	if !exists {
		return fmt.Errorf("radio %s not found", radioID)
	}

	radio.Status = status
	radio.LastSeen = time.Now()

	return nil
}

// RemoveRadio removes a radio from the inventory.
func (m *Manager) RemoveRadio(radioID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.radios[radioID]; !exists {
		return fmt.Errorf("radio %s not found", radioID)
	}

	delete(m.radios, radioID)
	delete(m.adapters, radioID)

	// If this was the active radio, clear active selection
	if m.activeRadioID == radioID {
		m.activeRadioID = ""
	}

	return nil
}

// RefreshCapabilities refreshes capabilities for a radio.
func (m *Manager) RefreshCapabilities(radioID string, timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	radioAdapter, exists := m.adapters[radioID]
	if !exists {
		return fmt.Errorf("no adapter for radio %s", radioID)
	}

	radio, exists := m.radios[radioID]
	if !exists {
		return fmt.Errorf("radio %s not found", radioID)
	}

	// Load updated capabilities
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	capabilities, err := radioAdapter.SupportedFrequencyProfiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh capabilities for radio %s: %w", radioID, err)
	}

	// Update capabilities
	radio.Capabilities.Channels = m.getChannelsFromCapabilities(capabilities, radioAdapter)
	radio.LastSeen = time.Now()

	return nil
}

// Helper methods for capability processing

func (m *Manager) getModelFromCapabilities(capabilities []adapter.FrequencyProfile) string {
	// Default model name - in real implementation, this would come from adapter
	return "Unknown-Radio"
}

func (m *Manager) getMinPowerFromCapabilities(capabilities []adapter.FrequencyProfile) int {
	// Default minimum power - in real implementation, this would come from adapter
	return 0
}

func (m *Manager) getMaxPowerFromCapabilities(capabilities []adapter.FrequencyProfile) int {
	// Default maximum power - in real implementation, this would come from adapter
	return 39
}

func (m *Manager) getChannelsFromCapabilities(capabilities []adapter.FrequencyProfile, radioAdapter adapter.IRadioAdapter) []adapter.Channel {
	// Try to get channels directly from adapter if it supports it (e.g., SilvusMock)
	if bandPlanAdapter, ok := radioAdapter.(interface{ GetBandPlan() []adapter.Channel }); ok {
		return bandPlanAdapter.GetBandPlan()
	}
	
	// Fallback: Convert frequency profiles to channels
	// In real implementation, this would derive channels from frequency profiles
	channels := make([]adapter.Channel, 0)
	for i, profile := range capabilities {
		for j, freq := range profile.Frequencies {
			channels = append(channels, adapter.Channel{
				Index:        i*len(profile.Frequencies) + j + 1, // 1-based indexing
				FrequencyMhz: freq,
			})
		}
	}
	return channels
}

func (m *Manager) determineStatus(err error) string {
	if err != nil {
		return "offline"
	}
	return "online"
}
