package silvusmock

import (
	"context"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
)

// TestReadPowerActualAccuracy tests ReadPowerActual accuracy bounds
func TestReadPowerActualAccuracy(t *testing.T) {
	// Create SilvusMock with known power setting
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 2, FrequencyMhz: 2417.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Set a specific power value
	expectedPower := 25.5
	silvus.SetPower(context.Background(), expectedPower)

	// Test ReadPowerActual returns the exact value
	actualPower, err := silvus.ReadPowerActual(context.Background())
	if err != nil {
		t.Fatalf("ReadPowerActual failed: %v", err)
	}

	if actualPower != expectedPower {
		t.Errorf("Expected power %f, got %f", expectedPower, actualPower)
	}
}

// TestReadPowerActualWithContextCancellation tests ReadPowerActual with context cancellation
func TestReadPowerActualWithContextCancellation(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Test ReadPowerActual with cancelled context
	_, err := silvus.ReadPowerActual(ctx)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

// TestReadPowerActualWithFaultInjection tests ReadPowerActual with fault injection
func TestReadPowerActualWithFaultInjection(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Test with BUSY fault injection
	silvus.SetFaultMode("ReturnBusy")
	_, err := silvus.ReadPowerActual(context.Background())
	if err == nil {
		t.Error("Expected error for BUSY fault injection")
	}

	// Test with UNAVAILABLE fault injection
	silvus.SetFaultMode("ReturnUnavailable")
	_, err = silvus.ReadPowerActual(context.Background())
	if err == nil {
		t.Error("Expected error for UNAVAILABLE fault injection")
	}

	// Test with INVALID_RANGE fault injection
	silvus.SetFaultMode("ReturnInvalidRange")
	_, err = silvus.ReadPowerActual(context.Background())
	if err == nil {
		t.Error("Expected error for INVALID_RANGE fault injection")
	}
}

// TestReadPowerActualConcurrency tests ReadPowerActual with concurrent access
func TestReadPowerActualConcurrency(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Set initial power
	silvus.SetPower(context.Background(), 30.0)

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			power, err := silvus.ReadPowerActual(context.Background())
			if err != nil {
				t.Errorf("ReadPowerActual failed: %v", err)
			}
			if power != 30.0 {
				t.Errorf("Expected power 30.0, got %f", power)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestSetBandPlanValidation tests SetBandPlan validation
func TestSetBandPlanValidation(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 2, FrequencyMhz: 2417.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Test setting a new band plan
	newBandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 5180.0},
		{Index: 2, FrequencyMhz: 5200.0},
		{Index: 3, FrequencyMhz: 5220.0},
	}

	silvus.SetBandPlan(newBandPlan)

	// Verify the band plan was set
	retrievedBandPlan := silvus.GetBandPlan()
	if len(retrievedBandPlan) != len(newBandPlan) {
		t.Errorf("Expected %d channels, got %d", len(newBandPlan), len(retrievedBandPlan))
	}

	// Verify channel details
	for i, channel := range newBandPlan {
		if retrievedBandPlan[i].Index != channel.Index {
			t.Errorf("Expected index %d, got %d", channel.Index, retrievedBandPlan[i].Index)
		}
		if retrievedBandPlan[i].FrequencyMhz != channel.FrequencyMhz {
			t.Errorf("Expected frequency %f, got %f", channel.FrequencyMhz, retrievedBandPlan[i].FrequencyMhz)
		}
	}
}

// TestSetBandPlanWithEmptyPlan tests SetBandPlan with empty plan
func TestSetBandPlanWithEmptyPlan(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Test setting empty band plan
	emptyBandPlan := []adapter.Channel{}
	silvus.SetBandPlan(emptyBandPlan)

	// Verify the band plan was set
	retrievedBandPlan := silvus.GetBandPlan()
	if len(retrievedBandPlan) != 0 {
		t.Errorf("Expected empty band plan, got %d channels", len(retrievedBandPlan))
	}
}

// TestSetBandPlanWithNilPlan tests SetBandPlan with nil plan
func TestSetBandPlanWithNilPlan(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Test setting nil band plan
	silvus.SetBandPlan(nil)

	// Verify the band plan was set to nil
	retrievedBandPlan := silvus.GetBandPlan()
	if retrievedBandPlan != nil {
		t.Error("Expected nil band plan")
	}
}

// TestSetBandPlanConcurrency tests SetBandPlan with concurrent access
func TestSetBandPlanConcurrency(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Test concurrent band plan updates
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(index int) {
			newBandPlan := []adapter.Channel{
				{Index: 1, FrequencyMhz: 2412.0 + float64(index)},
				{Index: 2, FrequencyMhz: 2417.0 + float64(index)},
			}
			silvus.SetBandPlan(newBandPlan)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify the band plan is in a consistent state
	retrievedBandPlan := silvus.GetBandPlan()
	if retrievedBandPlan == nil {
		t.Error("Expected non-nil band plan")
	}
}

// TestSetBandPlanFrequencyValidation tests SetBandPlan frequency validation
func TestSetBandPlanFrequencyValidation(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Test setting band plan with duplicate frequencies
	duplicateFreqBandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 2, FrequencyMhz: 2412.0}, // Duplicate frequency
	}

	silvus.SetBandPlan(duplicateFreqBandPlan)

	// Verify the band plan was set (no validation in current implementation)
	retrievedBandPlan := silvus.GetBandPlan()
	if len(retrievedBandPlan) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(retrievedBandPlan))
	}
}

// TestSetBandPlanWithNegativeFrequencies tests SetBandPlan with negative frequencies
func TestSetBandPlanWithNegativeFrequencies(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Test setting band plan with negative frequencies
	negativeFreqBandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: -1000.0}, // Negative frequency
		{Index: 2, FrequencyMhz: 0.0},     // Zero frequency
	}

	silvus.SetBandPlan(negativeFreqBandPlan)

	// Verify the band plan was set (no validation in current implementation)
	retrievedBandPlan := silvus.GetBandPlan()
	if len(retrievedBandPlan) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(retrievedBandPlan))
	}
}

// TestSetBandPlanWithLargeDataset tests SetBandPlan with large dataset
func TestSetBandPlanWithLargeDataset(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Create a large band plan
	largeBandPlan := make([]adapter.Channel, 1000)
	for i := 0; i < 1000; i++ {
		largeBandPlan[i] = adapter.Channel{
			Index:        i + 1,
			FrequencyMhz: 2412.0 + float64(i),
		}
	}

	// Test setting large band plan
	silvus.SetBandPlan(largeBandPlan)

	// Verify the band plan was set
	retrievedBandPlan := silvus.GetBandPlan()
	if len(retrievedBandPlan) != 1000 {
		t.Errorf("Expected 1000 channels, got %d", len(retrievedBandPlan))
	}
}

// TestSetBandPlanWithZeroIndex tests SetBandPlan with zero index
func TestSetBandPlanWithZeroIndex(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Test setting band plan with zero index
	zeroIndexBandPlan := []adapter.Channel{
		{Index: 0, FrequencyMhz: 2412.0}, // Zero index
		{Index: 1, FrequencyMhz: 2417.0},
	}

	silvus.SetBandPlan(zeroIndexBandPlan)

	// Verify the band plan was set
	retrievedBandPlan := silvus.GetBandPlan()
	if len(retrievedBandPlan) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(retrievedBandPlan))
	}
}

// TestSetBandPlanWithNegativeIndex tests SetBandPlan with negative index
func TestSetBandPlanWithNegativeIndex(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Test setting band plan with negative index
	negativeIndexBandPlan := []adapter.Channel{
		{Index: -1, FrequencyMhz: 2412.0}, // Negative index
		{Index: 1, FrequencyMhz: 2417.0},
	}

	silvus.SetBandPlan(negativeIndexBandPlan)

	// Verify the band plan was set
	retrievedBandPlan := silvus.GetBandPlan()
	if len(retrievedBandPlan) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(retrievedBandPlan))
	}
}

// TestSetBandPlanWithDuplicateIndex tests SetBandPlan with duplicate index
func TestSetBandPlanWithDuplicateIndex(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Test setting band plan with duplicate index
	duplicateIndexBandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 1, FrequencyMhz: 2417.0}, // Duplicate index
	}

	silvus.SetBandPlan(duplicateIndexBandPlan)

	// Verify the band plan was set
	retrievedBandPlan := silvus.GetBandPlan()
	if len(retrievedBandPlan) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(retrievedBandPlan))
	}
}

// TestSetBandPlanWithExtremeFrequencies tests SetBandPlan with extreme frequencies
func TestSetBandPlanWithExtremeFrequencies(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Test setting band plan with extreme frequencies
	extremeFreqBandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 0.001},      // Very low frequency
		{Index: 2, FrequencyMhz: 999999.999}, // Very high frequency
	}

	silvus.SetBandPlan(extremeFreqBandPlan)

	// Verify the band plan was set
	retrievedBandPlan := silvus.GetBandPlan()
	if len(retrievedBandPlan) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(retrievedBandPlan))
	}
}

// TestSetBandPlanWithMixedDataTypes tests SetBandPlan with mixed data types
func TestSetBandPlanWithMixedDataTypes(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Test setting band plan with mixed data types
	mixedBandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 2, FrequencyMhz: 2417.5},
		{Index: 3, FrequencyMhz: 2422.0},
	}

	silvus.SetBandPlan(mixedBandPlan)

	// Verify the band plan was set
	retrievedBandPlan := silvus.GetBandPlan()
	if len(retrievedBandPlan) != 3 {
		t.Errorf("Expected 3 channels, got %d", len(retrievedBandPlan))
	}
}

// TestSetBandPlanPerformance tests SetBandPlan performance with large datasets
func TestSetBandPlanPerformance(t *testing.T) {
	bandPlan := []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
	}
	silvus := NewSilvusMock("radio-01", bandPlan)

	// Create a large band plan for performance testing
	largeBandPlan := make([]adapter.Channel, 10000)
	for i := 0; i < 10000; i++ {
		largeBandPlan[i] = adapter.Channel{
			Index:        i + 1,
			FrequencyMhz: 2412.0 + float64(i),
		}
	}

	// Measure performance
	start := time.Now()
	silvus.SetBandPlan(largeBandPlan)
	duration := time.Since(start)

	// Verify the band plan was set
	retrievedBandPlan := silvus.GetBandPlan()
	if len(retrievedBandPlan) != 10000 {
		t.Errorf("Expected 10000 channels, got %d", len(retrievedBandPlan))
	}

	// Performance should be reasonable (less than 1 second for 10k channels)
	if duration > time.Second {
		t.Errorf("SetBandPlan took too long: %v", duration)
	}
}
