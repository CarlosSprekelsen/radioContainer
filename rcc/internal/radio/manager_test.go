package radio

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
)

// MockAdapter is a mock implementation of IRadioAdapter for testing.
type MockAdapter struct {
	GetStateFunc                   func(ctx context.Context) (*adapter.RadioState, error)
	SetPowerFunc                   func(ctx context.Context, dBm float64) error
	SetFrequencyFunc               func(ctx context.Context, frequencyMhz float64) error
	ReadPowerActualFunc            func(ctx context.Context) (float64, error)
	SupportedFrequencyProfilesFunc func(ctx context.Context) ([]adapter.FrequencyProfile, error)
}

func (m *MockAdapter) GetState(ctx context.Context) (*adapter.RadioState, error) {
	if m.GetStateFunc != nil {
		return m.GetStateFunc(ctx)
	}
	return &adapter.RadioState{PowerDbm: 30, FrequencyMhz: 2412.0}, nil
}

func (m *MockAdapter) SetPower(ctx context.Context, dBm float64) error {
	if m.SetPowerFunc != nil {
		return m.SetPowerFunc(ctx, dBm)
	}
	return nil
}

func (m *MockAdapter) SetFrequency(ctx context.Context, frequencyMhz float64) error {
	if m.SetFrequencyFunc != nil {
		return m.SetFrequencyFunc(ctx, frequencyMhz)
	}
	return nil
}

func (m *MockAdapter) ReadPowerActual(ctx context.Context) (float64, error) {
	if m.ReadPowerActualFunc != nil {
		return m.ReadPowerActualFunc(ctx)
	}
	return 30.0, nil
}

func (m *MockAdapter) SupportedFrequencyProfiles(ctx context.Context) ([]adapter.FrequencyProfile, error) {
	if m.SupportedFrequencyProfilesFunc != nil {
		return m.SupportedFrequencyProfilesFunc(ctx)
	}
	return []adapter.FrequencyProfile{
		{
			Frequencies: []float64{2412.0, 2417.0, 2422.0},
			Bandwidth:   20.0,
			AntennaMask: 1, // omni
		},
	}, nil
}

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.radios == nil {
		t.Error("Radios map not initialized")
	}

	if manager.adapters == nil {
		t.Error("Adapters map not initialized")
	}

	if manager.activeRadioID != "" {
		t.Errorf("Expected empty active radio ID, got '%s'", manager.activeRadioID)
	}
}

func TestLoadCapabilities(t *testing.T) {
	manager := NewManager()
	mockAdapter := &MockAdapter{}

	// Test successful capability loading
	err := manager.LoadCapabilities("radio-01", mockAdapter, 2*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	// Check that radio was added
	radio, exists := manager.radios["radio-01"]
	if !exists {
		t.Fatal("Radio not added to inventory")
	}

	if radio.ID != "radio-01" {
		t.Errorf("Expected radio ID 'radio-01', got '%s'", radio.ID)
	}

	if radio.Status != "online" {
		t.Errorf("Expected status 'online', got '%s'", radio.Status)
	}

	if radio.Capabilities == nil {
		t.Error("Expected capabilities, got nil")
	}

	if radio.State == nil {
		t.Error("Expected state, got nil")
	}

	// Check that adapter was stored by testing behavior
	adapter := manager.adapters["radio-01"]
	if adapter == nil {
		t.Error("Adapter not stored correctly")
	}
	// Test that the adapter works by calling a method
	_, err = adapter.ReadPowerActual(context.Background())
	if err != nil {
		t.Errorf("Stored adapter not working: %v", err)
	}

	// Check that first radio becomes active
	if manager.activeRadioID != "radio-01" {
		t.Errorf("Expected active radio 'radio-01', got '%s'", manager.activeRadioID)
	}
}

func TestLoadCapabilitiesWithError(t *testing.T) {
	manager := NewManager()

	// Mock adapter that returns error
	mockAdapter := &MockAdapter{
		SupportedFrequencyProfilesFunc: func(ctx context.Context) ([]adapter.FrequencyProfile, error) {
			return nil, &MockError{Message: "Adapter error"}
		},
	}

	// Test capability loading with error
	err := manager.LoadCapabilities("radio-01", mockAdapter, 2*time.Second)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Check that radio was not added
	if _, exists := manager.radios["radio-01"]; exists {
		t.Error("Radio should not be added on error")
	}
}

func TestSetActive(t *testing.T) {
	manager := NewManager()
	mockAdapter := &MockAdapter{}

	// Load a radio first
	err := manager.LoadCapabilities("radio-01", mockAdapter, 2*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	// Load another radio
	mockAdapter2 := &MockAdapter{}
	err = manager.LoadCapabilities("radio-02", mockAdapter2, 2*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	// Test setting active radio
	err = manager.SetActive("radio-02")
	if err != nil {
		t.Errorf("SetActive() failed: %v", err)
	}

	if manager.activeRadioID != "radio-02" {
		t.Errorf("Expected active radio 'radio-02', got '%s'", manager.activeRadioID)
	}

	// Test setting non-existent radio
	err = manager.SetActive("radio-99")
	if err == nil {
		t.Error("Expected error for non-existent radio")
	}
}

func TestGetActive(t *testing.T) {
	manager := NewManager()

	// Test with no active radio
	active := manager.GetActive()
	if active != "" {
		t.Errorf("Expected empty active radio, got '%s'", active)
	}

	// Load a radio and test
	mockAdapter := &MockAdapter{}
	err := manager.LoadCapabilities("radio-01", mockAdapter, 2*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	active = manager.GetActive()
	if active != "radio-01" {
		t.Errorf("Expected active radio 'radio-01', got '%s'", active)
	}
}

func TestGetActiveRadio(t *testing.T) {
	manager := NewManager()

	// Test with no active radio
	radio := manager.GetActiveRadio()
	if radio != nil {
		t.Error("Expected nil active radio, got non-nil")
	}

	// Load a radio and test
	mockAdapter := &MockAdapter{}
	err := manager.LoadCapabilities("radio-01", mockAdapter, 2*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	radio = manager.GetActiveRadio()
	if radio == nil {
		t.Fatal("Expected active radio, got nil")
	}

	if radio.ID != "radio-01" {
		t.Errorf("Expected radio ID 'radio-01', got '%s'", radio.ID)
	}
}

func TestGetActiveAdapter(t *testing.T) {
	manager := NewManager()

	// Test with no active radio
	adapter, radioID, err := manager.GetActiveAdapter()
	if err == nil {
		t.Error("Expected error for no active radio")
	}
	if adapter != nil {
		t.Error("Expected nil adapter")
	}
	if radioID != "" {
		t.Error("Expected empty radio ID")
	}

	// Load a radio and test
	mockAdapter := &MockAdapter{}
	err = manager.LoadCapabilities("radio-01", mockAdapter, 2*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	adapter, radioID, err = manager.GetActiveAdapter()
	if err != nil {
		t.Errorf("GetActiveAdapter() failed: %v", err)
	}

	if adapter == nil {
		t.Error("Expected adapter, got nil")
	}
	// Test that the adapter works by calling a method
	_, err = adapter.ReadPowerActual(context.Background())
	if err != nil {
		t.Errorf("Active adapter not working: %v", err)
	}

	if radioID != "radio-01" {
		t.Errorf("Expected radio ID 'radio-01', got '%s'", radioID)
	}
}

func TestList(t *testing.T) {
	manager := NewManager()

	// Test with no radios
	list := manager.List()
	if list.ActiveRadioID != "" {
		t.Errorf("Expected empty active radio ID, got '%s'", list.ActiveRadioID)
	}
	if len(list.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(list.Items))
	}

	// Load some radios
	mockAdapter1 := &MockAdapter{}
	err := manager.LoadCapabilities("radio-01", mockAdapter1, 5*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	mockAdapter2 := &MockAdapter{}
	err = manager.LoadCapabilities("radio-02", mockAdapter2, 5*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	// Test list
	list = manager.List()
	if list.ActiveRadioID != "radio-01" {
		t.Errorf("Expected active radio 'radio-01', got '%s'", list.ActiveRadioID)
	}
	if len(list.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(list.Items))
	}

	// Verify items contain expected radios
	radioIDs := make(map[string]bool)
	for _, item := range list.Items {
		radioIDs[item.ID] = true
	}

	if !radioIDs["radio-01"] {
		t.Error("Expected radio-01 in list")
	}
	if !radioIDs["radio-02"] {
		t.Error("Expected radio-02 in list")
	}
}

func TestGetRadio(t *testing.T) {
	manager := NewManager()

	// Test with non-existent radio
	radio, err := manager.GetRadio("radio-99")
	if err == nil {
		t.Error("Expected error for non-existent radio")
	}
	if radio != nil {
		t.Error("Expected nil radio")
	}

	// Load a radio and test
	mockAdapter := &MockAdapter{}
	err = manager.LoadCapabilities("radio-01", mockAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	radio, err = manager.GetRadio("radio-01")
	if err != nil {
		t.Errorf("GetRadio() failed: %v", err)
	}

	if radio.ID != "radio-01" {
		t.Errorf("Expected radio ID 'radio-01', got '%s'", radio.ID)
	}
}

func TestUpdateState(t *testing.T) {
	manager := NewManager()
	mockAdapter := &MockAdapter{}

	// Load a radio
	err := manager.LoadCapabilities("radio-01", mockAdapter, 2*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	// Update state
	newState := &adapter.RadioState{
		PowerDbm:     35,
		FrequencyMhz: 2422.0,
	}

	err = manager.UpdateState("radio-01", newState)
	if err != nil {
		t.Errorf("UpdateState() failed: %v", err)
	}

	// Verify state was updated
	radio, err := manager.GetRadio("radio-01")
	if err != nil {
		t.Fatalf("GetRadio() failed: %v", err)
	}

	if radio.State.PowerDbm != 35 {
		t.Errorf("Expected power 35, got %f", radio.State.PowerDbm)
	}
	if radio.State.FrequencyMhz != 2422.0 {
		t.Errorf("Expected frequency 2422.0, got %f", radio.State.FrequencyMhz)
	}
	if radio.Status != "online" {
		t.Errorf("Expected status 'online', got '%s'", radio.Status)
	}

	// Test with non-existent radio
	err = manager.UpdateState("radio-99", newState)
	if err == nil {
		t.Error("Expected error for non-existent radio")
	}
}

func TestUpdateStatus(t *testing.T) {
	manager := NewManager()
	mockAdapter := &MockAdapter{}

	// Load a radio
	err := manager.LoadCapabilities("radio-01", mockAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	// Update status
	err = manager.UpdateStatus("radio-01", "offline")
	if err != nil {
		t.Errorf("UpdateStatus() failed: %v", err)
	}

	// Verify status was updated
	radio, err := manager.GetRadio("radio-01")
	if err != nil {
		t.Fatalf("GetRadio() failed: %v", err)
	}

	if radio.Status != "offline" {
		t.Errorf("Expected status 'offline', got '%s'", radio.Status)
	}

	// Test with non-existent radio
	err = manager.UpdateStatus("radio-99", "offline")
	if err == nil {
		t.Error("Expected error for non-existent radio")
	}
}

func TestRemoveRadio(t *testing.T) {
	manager := NewManager()
	mockAdapter := &MockAdapter{}

	// Load a radio
	err := manager.LoadCapabilities("radio-01", mockAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	// Remove radio
	err = manager.RemoveRadio("radio-01")
	if err != nil {
		t.Errorf("RemoveRadio() failed: %v", err)
	}

	// Verify radio was removed
	_, err = manager.GetRadio("radio-01")
	if err == nil {
		t.Error("Expected error for removed radio")
	}

	// Verify active radio was cleared
	if manager.activeRadioID != "" {
		t.Errorf("Expected empty active radio ID, got '%s'", manager.activeRadioID)
	}

	// Test removing non-existent radio
	err = manager.RemoveRadio("radio-99")
	if err == nil {
		t.Error("Expected error for non-existent radio")
	}
}

func TestRefreshCapabilities(t *testing.T) {
	manager := NewManager()
	mockAdapter := &MockAdapter{}

	// Load a radio
	err := manager.LoadCapabilities("radio-01", mockAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	// Refresh capabilities
	err = manager.RefreshCapabilities("radio-01", 5*time.Second)
	if err != nil {
		t.Errorf("RefreshCapabilities() failed: %v", err)
	}

	// Test with non-existent radio
	err = manager.RefreshCapabilities("radio-99", 5*time.Second)
	if err == nil {
		t.Error("Expected error for non-existent radio")
	}
}

func TestMultipleRadios(t *testing.T) {
	manager := NewManager()

	// Load multiple radios
	mockAdapter1 := &MockAdapter{}
	err := manager.LoadCapabilities("radio-01", mockAdapter1, 5*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	mockAdapter2 := &MockAdapter{}
	err = manager.LoadCapabilities("radio-02", mockAdapter2, 2*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities() failed: %v", err)
	}

	// Verify both radios are in inventory
	list := manager.List()
	if len(list.Items) != 2 {
		t.Errorf("Expected 2 radios, got %d", len(list.Items))
	}

	// Test switching active radio
	err = manager.SetActive("radio-02")
	if err != nil {
		t.Errorf("SetActive() failed: %v", err)
	}

	if manager.activeRadioID != "radio-02" {
		t.Errorf("Expected active radio 'radio-02', got '%s'", manager.activeRadioID)
	}

	// Test removing active radio
	err = manager.RemoveRadio("radio-02")
	if err != nil {
		t.Errorf("RemoveRadio() failed: %v", err)
	}

	// Verify active radio was cleared
	if manager.activeRadioID != "" {
		t.Errorf("Expected empty active radio ID after removing active radio")
	}
}

func TestConcurrentAccess(t *testing.T) {
	manager := NewManager()

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			mockAdapter := &MockAdapter{}
			radioID := fmt.Sprintf("radio-%d", i)
			manager.LoadCapabilities(radioID, mockAdapter, 2*time.Second)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all radios were added
	list := manager.List()
	if len(list.Items) != 10 {
		t.Errorf("Expected 10 radios, got %d", len(list.Items))
	}
}

// MockError is a test error type
type MockError struct {
	Message string
}

func (e *MockError) Error() string {
	return e.Message
}
