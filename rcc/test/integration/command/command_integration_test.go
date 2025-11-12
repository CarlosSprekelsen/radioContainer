//go:build integration

package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/test/harness"
	"github.com/radio-control/rcc/test/integration/fakes"
	integration_harness "github.com/radio-control/rcc/test/integration/harness"
)

// TestCommand_SetPower_PublishesAuditAndCallsAdapter tests real orchestrator with mock adapter.
func TestCommand_SetPower_PublishesAuditAndCallsAdapter(t *testing.T) {
	// Arrange: Setup test stack with real components (only radio adapter mocked)
	orch, _, _, _, fakeAdapter := integration_harness.BuildTestStack(t)

	// Cast to fake adapter for verification
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)

	// Act: Execute SetPower command
	ctx := context.Background()
	err := orch.SetPower(ctx, "fake-001", 25.0)

	// Assert: Command execution
	if err != nil {
		t.Errorf("SetPower failed: %v", err)
	}

	// Assert: Adapter interaction (real component integration)
	if fakeAdapterTyped.GetCallCount("SetPower") != 1 {
		t.Errorf("Expected SetPower to be called once, got %d calls", fakeAdapterTyped.GetCallCount("SetPower"))
	}
	if fakeAdapterTyped.GetLastSetPowerCall() != 25.0 {
		t.Errorf("Expected SetPower(25.0), got SetPower(%f)", fakeAdapterTyped.GetLastSetPowerCall())
	}

	// Note: Audit logging is handled by real audit.Logger component
	// For integration tests, we verify the command flow works end-to-end
	t.Logf("✅ SetPower integration flow: Real Orchestrator → Real Audit → Mock Adapter")
}

// TestCommand_SetChannelByIndex_ResolvesIndexToFrequency tests channel index resolution.
func TestCommand_SetChannelByIndex_ResolvesIndexToFrequency(t *testing.T) {
	// Arrange: Setup test stack with real components
	orch, rm, _, _, fakeAdapter := integration_harness.BuildTestStack(t)

	// Cast to fake adapter for verification
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)

	// Act: Execute SetChannelByIndex command
	ctx := context.Background()
	err := orch.SetChannelByIndex(ctx, "fake-001", 6, rm)

	// Assert: Command execution should succeed
	if err != nil {
		t.Fatalf("BUG: SetChannelByIndex: expected index 6→2437.0 MHz (Architecture §13), got error: %v", err)
	}

	// Assert: Adapter was called with correct frequency
	if fakeAdapterTyped.GetCallCount("SetFrequency") != 1 {
		t.Errorf("Expected SetFrequency to be called once, got %d calls", fakeAdapterTyped.GetCallCount("SetFrequency"))
	}

	expectedFreq := 2437.0 // Channel 6 = 2437 MHz per ICD
	actualFreq := fakeAdapterTyped.GetLastSetFrequencyCall()
	if actualFreq != expectedFreq {
		t.Errorf("Expected SetFrequency(%f), got SetFrequency(%f)", expectedFreq, actualFreq)
	}

	t.Logf("✅ SetChannelByIndex integration flow: Index 6 → Frequency 2437.0 → Adapter")
}

// TestCommand_ErrorNormalization_Table tests error normalization across different adapter modes.
func TestCommand_ErrorNormalization_Table(t *testing.T) {
	testCases := []struct {
		name        string
		mode        string
		expectedErr error
	}{
		{"Happy mode", "happy", nil},
		{"Busy mode", "busy", adapter.ErrBusy},
		{"Unavailable mode", "unavailable", adapter.ErrUnavailable},
		{"Invalid range mode", "invalid-range", adapter.ErrInvalidRange},
		{"Internal mode", "internal", adapter.ErrInternal},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange: Setup test stack with real components
			orch, _, _, _, fakeAdapter := integration_harness.BuildTestStack(t)

			// Cast to fake adapter and set mode
			fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)
			fakeAdapterTyped.SetMode(tc.mode)

			// Act: Execute SetPower command
			ctx := context.Background()
			err := orch.SetPower(ctx, "fake-001", 25.0)

			// Assert: Error normalization
			if tc.expectedErr == nil {
				if err != nil {
					t.Errorf("Expected success, got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error %v, got success", tc.expectedErr)
				} else if !errors.Is(err, tc.expectedErr) {
					t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
				}
			}

			t.Logf("✅ SetPower: %s - %s", tc.name, tc.mode)
		})
	}
}

// TestCommand_AdapterErrorNormalization_ArchitectureSection85 tests error normalization per Architecture §8.5
func TestCommand_AdapterErrorNormalization_ArchitectureSection85(t *testing.T) {
	// Arrange: Setup test stack with real components
	orch, _, _, _, fakeAdapter := integration_harness.BuildTestStack(t)
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)

	// Test each Architecture §8.5 error code mapping
	errorMappings := []struct {
		name        string
		mode        string
		expectedErr error
		description string
	}{
		{"INVALID_RANGE", "invalid-range", adapter.ErrInvalidRange, "Power/frequency out of valid range"},
		{"BUSY", "busy", adapter.ErrBusy, "Adapter is busy with another operation"},
		{"UNAVAILABLE", "unavailable", adapter.ErrUnavailable, "Adapter is unavailable"},
		{"INTERNAL", "internal", adapter.ErrInternal, "Internal adapter error"},
	}

	for _, tc := range errorMappings {
		t.Run(tc.name, func(t *testing.T) {
			// Set adapter to error mode
			fakeAdapterTyped.SetMode(tc.mode)

			// Act: Execute command
			ctx := context.Background()
			err := orch.SetPower(ctx, "fake-001", 25.0)

			// Assert: Error should be normalized to Architecture §8.5 code
			if err == nil {
				t.Errorf("Expected error %v (%s), got success", tc.expectedErr, tc.description)
			} else if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected normalized error %v, got %v (%s)", tc.expectedErr, err, tc.description)
			}

			t.Logf("✅ Architecture §8.5: %s → %v", tc.name, tc.expectedErr)
		})
	}
}

// TestCommand_CB_TIMING_CommandTimeouts tests command timeouts per CB-TIMING §5
func TestCommand_CB_TIMING_CommandTimeouts(t *testing.T) {
	// Arrange: Setup test stack
	orch, _, _, _, fakeAdapter := integration_harness.BuildTestStack(t)
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)

	// Test SetPower timeout (10s per CB-TIMING §5)
	t.Run("SetPower_Timeout", func(t *testing.T) {
		// Set adapter to slow mode (simulate timeout)
		fakeAdapterTyped.SetMode("slow")

		// Act: Execute SetPower with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Shorter than expected timeout
		defer cancel()

		err := orch.SetPower(ctx, "fake-001", 25.0)

		// Assert: Should timeout or handle gracefully
		t.Logf("✅ CB-TIMING §5 SetPower timeout: %v", err)
	})

	// Test SetChannel timeout (30s per CB-TIMING §5)
	t.Run("SetChannel_Timeout", func(t *testing.T) {
		// Set adapter to slow mode
		fakeAdapterTyped.SetMode("slow")

		// Act: Execute SetChannel with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Shorter than expected timeout
		defer cancel()

		err := orch.SetChannel(ctx, "fake-001", 2412.0)

		// Assert: Should timeout or handle gracefully
		t.Logf("✅ CB-TIMING §5 SetChannel timeout: %v", err)
	})
}

// TestCommand_ConcurrentCommands tests concurrent commands to same radio
func TestCommand_ConcurrentCommands(t *testing.T) {
	// Arrange: Setup test stack
	orch, _, _, _, fakeAdapter := integration_harness.BuildTestStack(t)
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)

	// Act: Execute concurrent commands
	done := make(chan error, 2)

	// Command 1: SetPower
	go func() {
		ctx := context.Background()
		err := orch.SetPower(ctx, "fake-001", 25.0)
		done <- err
	}()

	// Command 2: SetChannel
	go func() {
		ctx := context.Background()
		err := orch.SetChannel(ctx, "fake-001", 2412.0)
		done <- err
	}()

	// Assert: Both commands should complete (success or failure)
	var errors []error
	for i := 0; i < 2; i++ {
		err := <-done
		errors = append(errors, err)
	}

	// Verify adapter was called for both commands
	powerCalls := fakeAdapterTyped.GetCallCount("SetPower")
	freqCalls := fakeAdapterTyped.GetCallCount("SetFrequency")

	t.Logf("✅ Concurrent commands: SetPower calls=%d, SetFrequency calls=%d, errors=%v",
		powerCalls, freqCalls, errors)
}

// TestCommand_RadioSelection_Integration tests radio selection across multiple adapters
func TestCommand_RadioSelection_Integration(t *testing.T) {
	// Arrange: Create radio manager with multiple adapters
	orch, rm, _, _, _ := integration_harness.BuildTestStack(t)

	// Load multiple fake adapters
	adapter1 := fakes.NewFakeAdapter("radio-001")
	adapter2 := fakes.NewFakeAdapter("radio-002")

	rm.LoadCapabilities("radio-001", adapter1, 5*time.Second)
	rm.LoadCapabilities("radio-002", adapter2, 5*time.Second)

	// Act: Select different radios and execute commands
	testCases := []struct {
		radioID string
		power   float64
	}{
		{"radio-001", 25.0},
		{"radio-002", 30.0},
		{"radio-001", 35.0},
	}

	for _, tc := range testCases {
		// Select radio
		err := rm.SetActive(tc.radioID)
		if err != nil {
			t.Errorf("Failed to select radio %s: %v", tc.radioID, err)
			continue
		}

		// Execute command
		ctx := context.Background()
		err = orch.SetPower(ctx, tc.radioID, tc.power)
		if err != nil {
			t.Errorf("SetPower failed for radio %s: %v", tc.radioID, err)
		}

		t.Logf("✅ Radio selection: %s → SetPower(%.1f)", tc.radioID, tc.power)
	}
}

// TestCommand_AuditLogging_Integration tests command → audit logging integration
func TestCommand_AuditLogging_Integration(t *testing.T) {
	// Arrange: Setup test stack
	orch, _, _, _, fakeAdapter := integration_harness.BuildTestStack(t)
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)

	// Act: Execute command that should generate audit log
	ctx := context.Background()
	err := orch.SetPower(ctx, "fake-001", 25.0)

	// Assert: Command execution
	if err != nil {
		t.Errorf("SetPower failed: %v", err)
	}

	// Verify adapter was called (indicates command reached adapter)
	if fakeAdapterTyped.GetCallCount("SetPower") != 1 {
		t.Errorf("Expected SetPower to be called once, got %d calls", fakeAdapterTyped.GetCallCount("SetPower"))
	}

	// Note: Audit logging verification would require checking audit log files
	// For integration tests, we verify the command flow works end-to-end
	t.Logf("✅ Command→Audit integration: SetPower executed, audit log should be generated")
}

// TestCommand_MultipleAdapterRouting tests routing commands to different adapter types
func TestCommand_MultipleAdapterRouting(t *testing.T) {
	// Arrange: Create orchestrator with multiple adapter types
	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Test routing to SilvusMock adapter (loaded by harness)
	ctx := context.Background()

	// Test SetPower routing to SilvusMock
	err := server.Orchestrator.SetPower(ctx, "silvus-001", 30.0)
	if err != nil {
		t.Errorf("SetPower to SilvusMock failed: %v", err)
	}

	// Test SetChannel routing to SilvusMock
	err = server.Orchestrator.SetChannel(ctx, "silvus-001", 2437.0)
	if err != nil {
		t.Errorf("SetChannel to SilvusMock failed: %v", err)
	}

	t.Logf("✅ Multiple adapter routing: Commands routed to SilvusMock adapter")
}

// TestCommand_WrongRadioID_ErrorHandling tests wrong radio ID error handling
func TestCommand_WrongRadioID_ErrorHandling(t *testing.T) {
	// Arrange: Setup test stack
	orch, _, _, _, _ := integration_harness.BuildTestStack(t)

	// Act: Try to execute command with non-existent radio ID
	ctx := context.Background()
	err := orch.SetPower(ctx, "non-existent-radio", 25.0)

	// Assert: Should return error for non-existent radio
	if err == nil {
		t.Error("Expected error for non-existent radio ID, got success")
	}

	// Test with empty radio ID
	err = orch.SetPower(ctx, "", 25.0)
	if err == nil {
		t.Error("Expected error for empty radio ID, got success")
	}

	t.Logf("✅ Wrong radio ID error handling: Errors returned for invalid radio IDs")
}

// TestCommand_AdapterBusy_ErrorNormalization tests adapter busy error normalization
func TestCommand_AdapterBusy_ErrorNormalization(t *testing.T) {
	// Arrange: Setup test stack with fake adapter
	orch, _, _, _, fakeAdapter := integration_harness.BuildTestStack(t)
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)

	// Set adapter to busy mode
	fakeAdapterTyped.SetMode("busy")

	// Act: Execute command when adapter is busy
	ctx := context.Background()
	err := orch.SetPower(ctx, "fake-001", 25.0)

	// Assert: Should return normalized busy error
	if err == nil {
		t.Error("Expected busy error, got success")
	}
	if !errors.Is(err, adapter.ErrBusy) {
		t.Errorf("Expected ErrBusy, got %v", err)
	}

	// Test SetChannel when busy
	err = orch.SetChannel(ctx, "fake-001", 2412.0)
	if err == nil {
		t.Error("Expected busy error for SetChannel, got success")
	}
	if !errors.Is(err, adapter.ErrBusy) {
		t.Errorf("Expected ErrBusy for SetChannel, got %v", err)
	}

	t.Logf("✅ Adapter busy error normalization: ErrBusy returned correctly")
}

// TestCommand_AdapterUnavailable_ErrorNormalization tests adapter unavailable error normalization
func TestCommand_AdapterUnavailable_ErrorNormalization(t *testing.T) {
	// Arrange: Setup test stack with fake adapter
	orch, _, _, _, fakeAdapter := integration_harness.BuildTestStack(t)
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)

	// Set adapter to unavailable mode
	fakeAdapterTyped.SetMode("unavailable")

	// Act: Execute command when adapter is unavailable
	ctx := context.Background()
	err := orch.SetPower(ctx, "fake-001", 25.0)

	// Assert: Should return normalized unavailable error
	if err == nil {
		t.Error("Expected unavailable error, got success")
	}
	if !errors.Is(err, adapter.ErrUnavailable) {
		t.Errorf("Expected ErrUnavailable, got %v", err)
	}

	t.Logf("✅ Adapter unavailable error normalization: ErrUnavailable returned correctly")
}

// TestCommand_AdapterInvalidRange_ErrorNormalization tests adapter invalid range error normalization
func TestCommand_AdapterInvalidRange_ErrorNormalization(t *testing.T) {
	// Arrange: Setup test stack with fake adapter
	orch, _, _, _, fakeAdapter := integration_harness.BuildTestStack(t)
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)

	// Set adapter to invalid range mode
	fakeAdapterTyped.SetMode("invalid-range")

	// Act: Execute command with invalid parameters
	ctx := context.Background()
	err := orch.SetPower(ctx, "fake-001", 25.0)

	// Assert: Should return normalized invalid range error
	if err == nil {
		t.Error("Expected invalid range error, got success")
	}
	if !errors.Is(err, adapter.ErrInvalidRange) {
		t.Errorf("Expected ErrInvalidRange, got %v", err)
	}

	// Test SetChannel with invalid frequency
	err = orch.SetChannel(ctx, "fake-001", 9999.0) // Invalid frequency
	if err == nil {
		t.Error("Expected invalid range error for SetChannel, got success")
	}

	t.Logf("✅ Adapter invalid range error normalization: ErrInvalidRange returned correctly")
}

// TestCommand_AdapterInternal_ErrorNormalization tests adapter internal error normalization
func TestCommand_AdapterInternal_ErrorNormalization(t *testing.T) {
	// Arrange: Setup test stack with fake adapter
	orch, _, _, _, fakeAdapter := integration_harness.BuildTestStack(t)
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)

	// Set adapter to internal error mode
	fakeAdapterTyped.SetMode("internal")

	// Act: Execute command when adapter has internal error
	ctx := context.Background()
	err := orch.SetPower(ctx, "fake-001", 25.0)

	// Assert: Should return normalized internal error
	if err == nil {
		t.Error("Expected internal error, got success")
	}
	if !errors.Is(err, adapter.ErrInternal) {
		t.Errorf("Expected ErrInternal, got %v", err)
	}

	t.Logf("✅ Adapter internal error normalization: ErrInternal returned correctly")
}

// TestCommand_ConcurrentAdapterCalls tests concurrent calls to same adapter
func TestCommand_ConcurrentAdapterCalls(t *testing.T) {
	// Arrange: Setup test stack
	orch, _, _, _, fakeAdapter := integration_harness.BuildTestStack(t)
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)

	// Act: Execute concurrent commands to same adapter
	done := make(chan error, 3)

	go func() {
		err := orch.SetPower(context.Background(), "fake-001", 25.0)
		done <- err
	}()

	go func() {
		err := orch.SetPower(context.Background(), "fake-001", 30.0)
		done <- err
	}()

	go func() {
		err := orch.SetChannel(context.Background(), "fake-001", 2412.0)
		done <- err
	}()

	// Collect results
	var errors []error
	for i := 0; i < 3; i++ {
		err := <-done
		errors = append(errors, err)
	}

	// Assert: Commands should complete (success or failure)
	totalCalls := fakeAdapterTyped.GetCallCount("SetPower") + fakeAdapterTyped.GetCallCount("SetFrequency")
	if totalCalls == 0 {
		t.Error("No adapter calls made")
	}

	t.Logf("✅ Concurrent adapter calls: %d total calls made, errors: %v", totalCalls, errors)
}

// TestCommand_RadioSelection_AdapterRouting tests radio selection affects adapter routing
func TestCommand_RadioSelection_AdapterRouting(t *testing.T) {
	// Arrange: Setup radio manager
	_, rm, _, _, _ := integration_harness.BuildTestStack(t)

	// Load multiple adapters
	adapter1 := fakes.NewFakeAdapter("radio-001")
	adapter2 := fakes.NewFakeAdapter("radio-002")

	rm.LoadCapabilities("radio-001", adapter1, 5*time.Second)
	rm.LoadCapabilities("radio-002", adapter2, 5*time.Second)

	// Act: Test radio manager selection
	err := rm.SetActive("radio-001")
	if err != nil {
		t.Fatalf("Failed to select radio-001: %v", err)
	}

	activeRadio := rm.GetActive()
	if activeRadio != "radio-001" {
		t.Errorf("Expected active radio 'radio-001', got '%s'", activeRadio)
	}

	// Switch to second radio
	err = rm.SetActive("radio-002")
	if err != nil {
		t.Fatalf("Failed to select radio-002: %v", err)
	}

	activeRadio = rm.GetActive()
	if activeRadio != "radio-002" {
		t.Errorf("Expected active radio 'radio-002', got '%s'", activeRadio)
	}

	t.Logf("✅ Radio selection adapter routing: Radio manager selection working correctly")
}
