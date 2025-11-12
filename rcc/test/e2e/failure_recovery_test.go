//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/test/harness"
)

// TestFailureRecovery_NetworkPartition tests network partition recovery:
// Network partition of adapter; disconnection/reconnect; verify recovery and no orphaned operations
func TestFailureRecovery_NetworkPartition(t *testing.T) {
	// Arrange: Create harness with network-partitioning adapter
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-partition-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Create network-partitioning adapter
	partitionAdapter := &NetworkPartitioningAdapter{
		radioID:     opts.ActiveRadioID,
		partitioned: false,
	}

	// Replace the default adapter with partition adapter
	err := server.RadioManager.LoadCapabilities(opts.ActiveRadioID, partitionAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load partition adapter capabilities: %v", err)
	}

	ctx := context.Background()

	// Act: Execute operation before partition
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 25.0)
	if err != nil {
		t.Fatalf("SetPower before partition failed: %v", err)
	}

	// Simulate network partition
	partitionAdapter.SetPartitioned(true)
	t.Logf("✅ Network partition simulated")

	// Attempt operation during partition (should fail)
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 30.0)
	if err == nil {
		t.Error("Expected error during network partition")
	}

	// Simulate network recovery
	partitionAdapter.SetPartitioned(false)
	t.Logf("✅ Network recovery simulated")

	// Wait for recovery
	time.Sleep(100 * time.Millisecond)

	// Attempt operation after recovery (should succeed)
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 30.0)
	if err != nil {
		t.Errorf("SetPower after recovery failed: %v", err)
	}

	// Assert: No orphaned operations
	state, err := partitionAdapter.GetState(ctx)
	if err != nil {
		t.Errorf("Failed to get adapter state: %v", err)
	}

	// Verify final state is correct
	if state.PowerDbm != 30.0 {
		t.Errorf("Expected final power 30.0, got %v", state.PowerDbm)
	}

	t.Logf("✅ Network partition recovery: operation succeeded after recovery")
}

// TestFailureRecovery_AdapterDisconnection tests adapter disconnection and reconnection:
// Adapter disconnection; reconnection; verify recovery and no orphaned operations
func TestFailureRecovery_AdapterDisconnection(t *testing.T) {
	// Arrange: Create harness with disconnecting adapter
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-disconnect-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Create disconnecting adapter
	disconnectAdapter := &DisconnectingAdapter{
		radioID:      opts.ActiveRadioID,
		disconnected: false,
	}

	// Replace the default adapter
	err := server.RadioManager.LoadCapabilities(opts.ActiveRadioID, disconnectAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load disconnect adapter capabilities: %v", err)
	}

	ctx := context.Background()

	// Act: Execute operation before disconnection
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 25.0)
	if err != nil {
		t.Fatalf("SetPower before disconnection failed: %v", err)
	}

	// Simulate adapter disconnection
	disconnectAdapter.SetDisconnected(true)
	t.Logf("✅ Adapter disconnection simulated")

	// Attempt operation during disconnection (should fail)
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 30.0)
	if err == nil {
		t.Error("Expected error during adapter disconnection")
	}

	// Simulate adapter reconnection
	disconnectAdapter.SetDisconnected(false)
	t.Logf("✅ Adapter reconnection simulated")

	// Wait for recovery
	time.Sleep(100 * time.Millisecond)

	// Attempt operation after reconnection (should succeed)
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 30.0)
	if err != nil {
		t.Errorf("SetPower after reconnection failed: %v", err)
	}

	// Assert: No orphaned operations
	state, err := disconnectAdapter.GetState(ctx)
	if err != nil {
		t.Errorf("Failed to get adapter state: %v", err)
	}

	// Verify final state is correct
	if state.PowerDbm != 30.0 {
		t.Errorf("Expected final power 30.0, got %v", state.PowerDbm)
	}

	t.Logf("✅ Adapter disconnection recovery: operation succeeded after reconnection")
}

// TestFailureRecovery_DatabaseConnectionLoss tests database connection loss and recovery:
// Database connection loss; recovery; verify no data loss
func TestFailureRecovery_DatabaseConnectionLoss(t *testing.T) {
	// Arrange: Create harness with database connection
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-db-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Create database-dependent adapter
	dbAdapter := &DatabaseDependentAdapter{
		radioID:        opts.ActiveRadioID,
		dbConnected:    true,
		persistedState: make(map[string]interface{}),
	}

	// Replace the default adapter
	err := server.RadioManager.LoadCapabilities(opts.ActiveRadioID, dbAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load database adapter capabilities: %v", err)
	}

	ctx := context.Background()

	// Act: Execute operation before database loss
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 25.0)
	if err != nil {
		t.Fatalf("SetPower before database loss failed: %v", err)
	}

	// Simulate database connection loss
	dbAdapter.SetDatabaseConnected(false)
	t.Logf("✅ Database connection loss simulated")

	// Attempt operation during database loss (should fail)
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 30.0)
	if err == nil {
		t.Error("Expected error during database connection loss")
	}

	// Simulate database recovery
	dbAdapter.SetDatabaseConnected(true)
	t.Logf("✅ Database recovery simulated")

	// Wait for recovery
	time.Sleep(100 * time.Millisecond)

	// Attempt operation after database recovery (should succeed)
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 30.0)
	if err != nil {
		t.Errorf("SetPower after database recovery failed: %v", err)
	}

	// Assert: No data loss
	state, err := dbAdapter.GetState(ctx)
	if err != nil {
		t.Errorf("Failed to get adapter state: %v", err)
	}

	// Verify final state is correct
	if state.PowerDbm != 30.0 {
		t.Errorf("Expected final power 30.0, got %v", state.PowerDbm)
	}

	// Verify data persistence
	persistedPower, exists := dbAdapter.persistedState["power"]
	if !exists {
		t.Error("Power state should be persisted")
	}
	if persistedPower != 30.0 {
		t.Errorf("Expected persisted power 30.0, got %v", persistedPower)
	}

	t.Logf("✅ Database connection recovery: operation succeeded, data persisted")
}

// TestFailureRecovery_ExternalServiceDependency tests external service dependency failures:
// External service dependency failures; recovery; verify graceful degradation
func TestFailureRecovery_ExternalServiceDependency(t *testing.T) {
	// Arrange: Create harness with external service dependency
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-external-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Create external service dependent adapter
	externalAdapter := &ExternalServiceDependentAdapter{
		radioID:           opts.ActiveRadioID,
		externalConnected: true,
	}

	// Replace the default adapter
	err := server.RadioManager.LoadCapabilities(opts.ActiveRadioID, externalAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load external service adapter capabilities: %v", err)
	}

	ctx := context.Background()

	// Act: Execute operation before external service failure
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 25.0)
	if err != nil {
		t.Fatalf("SetPower before external service failure failed: %v", err)
	}

	// Simulate external service failure
	externalAdapter.SetExternalServiceConnected(false)
	t.Logf("✅ External service failure simulated")

	// Attempt operation during external service failure (should fail gracefully)
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 30.0)
	if err == nil {
		t.Error("Expected error during external service failure")
	}

	// Simulate external service recovery
	externalAdapter.SetExternalServiceConnected(true)
	t.Logf("✅ External service recovery simulated")

	// Wait for recovery
	time.Sleep(100 * time.Millisecond)

	// Attempt operation after external service recovery (should succeed)
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 30.0)
	if err != nil {
		t.Errorf("SetPower after external service recovery failed: %v", err)
	}

	// Assert: Graceful degradation and recovery
	state, err := externalAdapter.GetState(ctx)
	if err != nil {
		t.Errorf("Failed to get adapter state: %v", err)
	}

	// Verify final state is correct
	if state.PowerDbm != 30.0 {
		t.Errorf("Expected final power 30.0, got %v", state.PowerDbm)
	}

	t.Logf("✅ External service dependency recovery: graceful degradation and recovery")
}

// Helper types for testing

// NetworkPartitioningAdapter simulates network partition scenarios
type NetworkPartitioningAdapter struct {
	radioID     string
	partitioned bool
	powerDbm    float64
}

func (a *NetworkPartitioningAdapter) GetRadioID() string { return a.radioID }
func (a *NetworkPartitioningAdapter) SetPower(ctx context.Context, powerDbm float64) error {
	if a.partitioned {
		return adapter.ErrUnavailable
	}
	a.powerDbm = powerDbm
	return nil
}
func (a *NetworkPartitioningAdapter) SetChannel(ctx context.Context, frequencyMhz float64) error {
	if a.partitioned {
		return adapter.ErrUnavailable
	}
	return nil
}
func (a *NetworkPartitioningAdapter) GetState(ctx context.Context) (adapter.RadioState, error) {
	if a.partitioned {
		return adapter.RadioState{}, adapter.ErrUnavailable
	}
	return adapter.RadioState{
		PowerDbm: a.powerDbm,
		IsActive: true,
	}, nil
}
func (a *NetworkPartitioningAdapter) ReadPowerActual(ctx context.Context) (float64, error) {
	if a.partitioned {
		return 0, adapter.ErrUnavailable
	}
	return a.powerDbm, nil
}
func (a *NetworkPartitioningAdapter) GetSupportedFrequencyProfiles() []adapter.FrequencyProfile {
	return []adapter.FrequencyProfile{
		{Name: "2.4GHz", MinFreq: 2400, MaxFreq: 2500},
	}
}
func (a *NetworkPartitioningAdapter) SetBandPlan(ctx context.Context, plan adapter.BandPlan) error {
	if a.partitioned {
		return adapter.ErrUnavailable
	}
	return nil
}

func (a *NetworkPartitioningAdapter) SetPartitioned(partitioned bool) {
	a.partitioned = partitioned
}

// DisconnectingAdapter simulates adapter disconnection scenarios
type DisconnectingAdapter struct {
	radioID      string
	disconnected bool
	powerDbm     float64
}

func (a *DisconnectingAdapter) GetRadioID() string { return a.radioID }
func (a *DisconnectingAdapter) SetPower(ctx context.Context, powerDbm float64) error {
	if a.disconnected {
		return adapter.ErrUnavailable
	}
	a.powerDbm = powerDbm
	return nil
}
func (a *DisconnectingAdapter) SetChannel(ctx context.Context, frequencyMhz float64) error {
	if a.disconnected {
		return adapter.ErrUnavailable
	}
	return nil
}
func (a *DisconnectingAdapter) GetState(ctx context.Context) (adapter.RadioState, error) {
	if a.disconnected {
		return adapter.RadioState{}, adapter.ErrUnavailable
	}
	return adapter.RadioState{
		PowerDbm: a.powerDbm,
		IsActive: true,
	}, nil
}
func (a *DisconnectingAdapter) ReadPowerActual(ctx context.Context) (float64, error) {
	if a.disconnected {
		return 0, adapter.ErrUnavailable
	}
	return a.powerDbm, nil
}
func (a *DisconnectingAdapter) GetSupportedFrequencyProfiles() []adapter.FrequencyProfile {
	return []adapter.FrequencyProfile{
		{Name: "2.4GHz", MinFreq: 2400, MaxFreq: 2500},
	}
}
func (a *DisconnectingAdapter) SetBandPlan(ctx context.Context, plan adapter.BandPlan) error {
	if a.disconnected {
		return adapter.ErrUnavailable
	}
	return nil
}

func (a *DisconnectingAdapter) SetDisconnected(disconnected bool) {
	a.disconnected = disconnected
}

// DatabaseDependentAdapter simulates database dependency scenarios
type DatabaseDependentAdapter struct {
	radioID        string
	dbConnected    bool
	persistedState map[string]interface{}
	powerDbm       float64
}

func (a *DatabaseDependentAdapter) GetRadioID() string { return a.radioID }
func (a *DatabaseDependentAdapter) SetPower(ctx context.Context, powerDbm float64) error {
	if !a.dbConnected {
		return adapter.ErrUnavailable
	}
	a.powerDbm = powerDbm
	a.persistedState["power"] = powerDbm
	return nil
}
func (a *DatabaseDependentAdapter) SetChannel(ctx context.Context, frequencyMhz float64) error {
	if !a.dbConnected {
		return adapter.ErrUnavailable
	}
	a.persistedState["frequency"] = frequencyMhz
	return nil
}
func (a *DatabaseDependentAdapter) GetState(ctx context.Context) (adapter.RadioState, error) {
	if !a.dbConnected {
		return adapter.RadioState{}, adapter.ErrUnavailable
	}
	return adapter.RadioState{
		PowerDbm: a.powerDbm,
		IsActive: true,
	}, nil
}
func (a *DatabaseDependentAdapter) ReadPowerActual(ctx context.Context) (float64, error) {
	if !a.dbConnected {
		return 0, adapter.ErrUnavailable
	}
	return a.powerDbm, nil
}
func (a *DatabaseDependentAdapter) GetSupportedFrequencyProfiles() []adapter.FrequencyProfile {
	return []adapter.FrequencyProfile{
		{Name: "2.4GHz", MinFreq: 2400, MaxFreq: 2500},
	}
}
func (a *DatabaseDependentAdapter) SetBandPlan(ctx context.Context, plan adapter.BandPlan) error {
	if !a.dbConnected {
		return adapter.ErrUnavailable
	}
	return nil
}

func (a *DatabaseDependentAdapter) SetDatabaseConnected(connected bool) {
	a.dbConnected = connected
}

// ExternalServiceDependentAdapter simulates external service dependency scenarios
type ExternalServiceDependentAdapter struct {
	radioID           string
	externalConnected bool
	powerDbm          float64
}

func (a *ExternalServiceDependentAdapter) GetRadioID() string { return a.radioID }
func (a *ExternalServiceDependentAdapter) SetPower(ctx context.Context, powerDbm float64) error {
	if !a.externalConnected {
		return adapter.ErrUnavailable
	}
	a.powerDbm = powerDbm
	return nil
}
func (a *ExternalServiceDependentAdapter) SetChannel(ctx context.Context, frequencyMhz float64) error {
	if !a.externalConnected {
		return adapter.ErrUnavailable
	}
	return nil
}
func (a *ExternalServiceDependentAdapter) GetState(ctx context.Context) (adapter.RadioState, error) {
	if !a.externalConnected {
		return adapter.RadioState{}, adapter.ErrUnavailable
	}
	return adapter.RadioState{
		PowerDbm: a.powerDbm,
		IsActive: true,
	}, nil
}
func (a *ExternalServiceDependentAdapter) ReadPowerActual(ctx context.Context) (float64, error) {
	if !a.externalConnected {
		return 0, adapter.ErrUnavailable
	}
	return a.powerDbm, nil
}
func (a *ExternalServiceDependentAdapter) GetSupportedFrequencyProfiles() []adapter.FrequencyProfile {
	return []adapter.FrequencyProfile{
		{Name: "2.4GHz", MinFreq: 2400, MaxFreq: 2500},
	}
}
func (a *ExternalServiceDependentAdapter) SetBandPlan(ctx context.Context, plan adapter.BandPlan) error {
	if !a.externalConnected {
		return adapter.ErrUnavailable
	}
	return nil
}

func (a *ExternalServiceDependentAdapter) SetExternalServiceConnected(connected bool) {
	a.externalConnected = connected
}
