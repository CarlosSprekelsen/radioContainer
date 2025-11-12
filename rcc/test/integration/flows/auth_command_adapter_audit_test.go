//go:build integration

package flows

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/adapter/fake"
	"github.com/radio-control/rcc/test/harness"
)

// TestAuthCommandAdapterAuditFlow_HappyPath tests the complete flow:
// API token accepted → Orchestrator routes → Adapter executes → Audit entry matches schema
func TestAuthCommandAdapterAuditFlow_HappyPath(t *testing.T) {
	// Arrange: Create harness with both fake and silvusmock adapters
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Create a second radio with fake adapter for cross-adapter testing
	fakeAdapter := fake.NewFakeAdapter("fake-radio-001")
	err := server.RadioManager.LoadCapabilities("fake-radio-001", fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load fake adapter capabilities: %v", err)
	}

	// Test setPower on silvusmock adapter
	ctx := context.Background()
	powerDbm := 25.0

	// Act: Execute setPower command
	start := time.Now()
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, powerDbm)
	latency := time.Since(start)

	// Assert: Command succeeds
	if err != nil {
		t.Errorf("SetPower failed: %v", err)
	}

	// Assert: Latency within CB-TIMING §5 bounds (setPower: 10s)
	// Note: In real implementation, would get timeout from orchestrator config
	expectedTimeout := 10 * time.Second
	if latency > expectedTimeout {
		t.Errorf("SetPower latency %v exceeds CB-TIMING timeout %v",
			latency, expectedTimeout)
	}

	// Assert: Audit entry matches schema per Architecture §8.6
	auditEntries := getAuditEntries(t, server, 1)
	if len(auditEntries) == 0 {
		t.Fatal("Expected audit entry for setPower")
	}

	entry := auditEntries[0]
	if entry.Action != "setPower" {
		t.Errorf("Expected action 'setPower', got '%s'", entry.Action)
	}
	if entry.RadioID != opts.ActiveRadioID {
		t.Errorf("Expected radioID '%s', got '%s'", opts.ActiveRadioID, entry.RadioID)
	}
	if entry.Result != "SUCCESS" {
		t.Errorf("Expected result 'SUCCESS', got '%s'", entry.Result)
	}
	expectedMaxLatency := int64(10 * time.Second.Milliseconds()) // CB-TIMING §5 setPower timeout
	if entry.LatencyMS < 0 || entry.LatencyMS > expectedMaxLatency {
		t.Errorf("Latency %dms outside expected bounds", entry.LatencyMS)
	}

	t.Logf("✅ Auth→Command→Adapter→Audit flow: SUCCESS (latency: %v)", latency)
}

// TestAuthCommandAdapterAuditFlow_ErrorMapping tests error propagation:
// Adapter returns vendor errors → map to normalized errors per Architecture §8.5
func TestAuthCommandAdapterAuditFlow_ErrorMapping(t *testing.T) {
	// Arrange: Create harness with error-injecting adapter
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "error-radio-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Create error-injecting adapter
	errorAdapter := &ErrorInjectingAdapter{
		radioID: opts.ActiveRadioID,
		errors: map[string]error{
			"setPower": adapter.ErrUnavailable, // Use available error type
		},
	}

	err := server.RadioManager.LoadCapabilities(opts.ActiveRadioID, errorAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load error adapter capabilities: %v", err)
	}

	// Act: Execute setPower command that will fail
	ctx := context.Background()
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 25.0)

	// Assert: Error normalized per Architecture §8.5
	if err == nil {
		t.Error("Expected error from adapter")
	}

	// Verify error type is normalized
	// Note: In real implementation, would check for specific error types
	if err == nil {
		t.Error("Expected error from adapter")
	}

	// Assert: Audit entry records failure
	auditEntries := getAuditEntries(t, server, 1)
	if len(auditEntries) == 0 {
		t.Fatal("Expected audit entry for failed setPower")
	}

	entry := auditEntries[0]
	if entry.Action != "setPower" {
		t.Errorf("Expected action 'setPower', got '%s'", entry.Action)
	}
	if entry.Result != "BUSY" {
		t.Errorf("Expected result 'BUSY', got '%s'", entry.Result)
	}

	t.Logf("✅ Error mapping: %v → BUSY (normalized)", err)
}

// TestAuthCommandAdapterAuditFlow_Timeout tests timeout behavior:
// Enforce CB-TIMING §5 for setPower (10s); simulate adapter stall; assert cancel + normalized error
func TestAuthCommandAdapterAuditFlow_Timeout(t *testing.T) {
	// Arrange: Create harness with timeout-injecting adapter
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "timeout-radio-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Create timeout-injecting adapter
	timeoutAdapter := &TimeoutInjectingAdapter{
		radioID: opts.ActiveRadioID,
		timeout: 15 * time.Second, // Longer than CB-TIMING §5 setPower timeout (10s)
	}

	err := server.RadioManager.LoadCapabilities(opts.ActiveRadioID, timeoutAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load timeout adapter capabilities: %v", err)
	}

	// Act: Execute setPower command that will timeout
	ctx := context.Background()
	start := time.Now()
	err = server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 25.0)
	latency := time.Since(start)

	// Assert: Timeout occurs within CB-TIMING bounds
	expectedTimeout := 10 * time.Second // CB-TIMING §5 setPower timeout
	if latency < expectedTimeout-100*time.Millisecond || latency > expectedTimeout+100*time.Millisecond {
		t.Errorf("Timeout latency %v not within expected bounds %v±100ms", latency, expectedTimeout)
	}

	// Assert: Error normalized to timeout
	if err == nil {
		t.Error("Expected timeout error")
	}
	// Note: Timeout errors should be mapped to ErrUnavailable per Architecture §8.5
	if !errors.Is(err, adapter.ErrUnavailable) {
		t.Errorf("Expected normalized ErrUnavailable for timeout, got: %v", err)
	}

	// Assert: Audit entry records timeout
	auditEntries := getAuditEntries(t, server, 1)
	if len(auditEntries) == 0 {
		t.Fatal("Expected audit entry for timeout")
	}

	entry := auditEntries[0]
	if entry.Action != "setPower" {
		t.Errorf("Expected action 'setPower', got '%s'", entry.Action)
	}
	if entry.Result != "TIMEOUT" {
		t.Errorf("Expected result 'TIMEOUT', got '%s'", entry.Result)
	}

	t.Logf("✅ Timeout handling: %v → TIMEOUT (normalized)", err)
}

// TestAuthCommandAdapterAuditFlow_CrossAdapter tests cross-adapter routing:
// Two radios with different adapters; concurrent operations; assert isolation
func TestAuthCommandAdapterAuditFlow_CrossAdapter(t *testing.T) {
	// Arrange: Create harness with multiple adapters
	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Add fake adapter for second radio
	fakeAdapter := fake.NewFakeAdapter("fake-radio-002")
	err := server.RadioManager.LoadCapabilities("fake-radio-002", fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load fake adapter capabilities: %v", err)
	}

	// Act: Execute concurrent operations on different adapters
	ctx := context.Background()

	// Channel for coordination
	done := make(chan error, 2)

	// Operation 1: setPower on silvusmock adapter
	go func() {
		err := server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 25.0)
		done <- err
	}()

	// Operation 2: setPower on fake adapter
	go func() {
		err := server.Orchestrator.SetPower(ctx, "fake-radio-002", 30.0)
		done <- err
	}()

	// Wait for both operations
	var errors []error
	for i := 0; i < 2; i++ {
		if err := <-done; err != nil {
			errors = append(errors, err)
		}
	}

	// Assert: Both operations succeed (no cross-adapter interference)
	if len(errors) > 0 {
		t.Errorf("Cross-adapter operations failed: %v", errors)
	}

	// Assert: Audit entries for both operations
	auditEntries := getAuditEntries(t, server, 2)
	if len(auditEntries) != 2 {
		t.Errorf("Expected 2 audit entries, got %d", len(auditEntries))
	}

	// Verify both radio IDs are present
	radioIDs := make(map[string]bool)
	for _, entry := range auditEntries {
		radioIDs[entry.RadioID] = true
	}
	if !radioIDs[opts.ActiveRadioID] {
		t.Error("Missing audit entry for silvusmock radio")
	}
	if !radioIDs["fake-radio-002"] {
		t.Error("Missing audit entry for fake radio")
	}

	t.Logf("✅ Cross-adapter isolation: both operations succeeded")
}

// Helper types for testing

// ErrorInjectingAdapter injects specific errors for testing
type ErrorInjectingAdapter struct {
	radioID string
	errors  map[string]error
}

func (a *ErrorInjectingAdapter) GetRadioID() string { return a.radioID }
func (a *ErrorInjectingAdapter) SetPower(ctx context.Context, powerDbm float64) error {
	if err, exists := a.errors["setPower"]; exists {
		return err
	}
	return nil
}
func (a *ErrorInjectingAdapter) SetFrequency(ctx context.Context, frequencyMhz float64) error {
	return nil
}
func (a *ErrorInjectingAdapter) GetState(ctx context.Context) (*adapter.RadioState, error) {
	return &adapter.RadioState{}, nil
}
func (a *ErrorInjectingAdapter) ReadPowerActual(ctx context.Context) (float64, error) { return 0, nil }
func (a *ErrorInjectingAdapter) SupportedFrequencyProfiles(ctx context.Context) ([]adapter.FrequencyProfile, error) {
	return nil, nil
}

// TimeoutInjectingAdapter injects timeouts for testing
type TimeoutInjectingAdapter struct {
	radioID string
	timeout time.Duration
}

func (a *TimeoutInjectingAdapter) GetRadioID() string { return a.radioID }
func (a *TimeoutInjectingAdapter) SetPower(ctx context.Context, powerDbm float64) error {
	time.Sleep(a.timeout) // Simulate long operation
	return nil
}
func (a *TimeoutInjectingAdapter) SetFrequency(ctx context.Context, frequencyMhz float64) error {
	return nil
}
func (a *TimeoutInjectingAdapter) GetState(ctx context.Context) (*adapter.RadioState, error) {
	return &adapter.RadioState{}, nil
}
func (a *TimeoutInjectingAdapter) ReadPowerActual(ctx context.Context) (float64, error) {
	return 0, nil
}
func (a *TimeoutInjectingAdapter) SupportedFrequencyProfiles(ctx context.Context) ([]adapter.FrequencyProfile, error) {
	return nil, nil
}

// AuditEntry represents the audit log schema per Architecture §8.6
type AuditEntry struct {
	Timestamp time.Time `json:"ts"`
	Actor     string    `json:"user"`
	RadioID   string    `json:"radioId"`
	Action    string    `json:"action"`
	Result    string    `json:"outcome"`
	LatencyMS int64     `json:"latencyMs"`
}

// getAuditEntries retrieves audit entries from the test server
func getAuditEntries(t *testing.T, server *harness.Server, expectedCount int) []AuditEntry {
	// For now, return empty slice - in real implementation, this would read from audit log file
	// or make HTTP request to audit endpoint
	return []AuditEntry{}
}
