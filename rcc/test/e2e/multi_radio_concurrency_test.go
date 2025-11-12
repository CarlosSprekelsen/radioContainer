//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/adapter/fake"
	"github.com/radio-control/rcc/internal/adapter/silvusmock"
	"github.com/radio-control/rcc/test/harness"
)

// TestMultiRadioConcurrency_ConcurrentOperations tests concurrent operations across ≥2 radios:
// Resource contention; verify no cross-radio leakage; audit integrity under load
func TestMultiRadioConcurrency_ConcurrentOperations(t *testing.T) {
	// Arrange: Create harness with multiple radios
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-multi-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Add second radio with fake adapter
	fakeAdapter := fake.NewFakeAdapter("fake-multi-002")
	err := server.RadioManager.LoadCapabilities("fake-multi-002", fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load fake adapter capabilities: %v", err)
	}

	// Add third radio with silvusmock adapter
	silvusAdapter2 := silvusmock.NewSilvusMock("silvus-multi-003", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 6, FrequencyMhz: 2437.0},
		{Index: 11, FrequencyMhz: 2462.0},
	})
	err = server.RadioManager.LoadCapabilities("silvus-multi-003", silvusAdapter2, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load second silvusmock adapter capabilities: %v", err)
	}

	ctx := context.Background()

	// Act: Execute concurrent operations across all radios
	done := make(chan error, 9) // 3 radios × 3 operations each

	// Radio 1 (silvusmock): setPower, setChannel, getState
	go func() {
		err := server.Orchestrator.SetPower(ctx, "silvus-multi-001", 25.0)
		done <- err
	}()
	go func() {
		err := server.Orchestrator.SetChannel(ctx, "silvus-multi-001", 2412.0)
		done <- err
	}()
	go func() {
		_, err := server.Orchestrator.GetState(ctx, "silvus-multi-001")
		done <- err
	}()

	// Radio 2 (fake): setPower, setChannel, getState
	go func() {
		err := fakeAdapter.SetPower(ctx, 30.0)
		done <- err
	}()
	go func() {
		err := fakeAdapter.SetChannel(ctx, 2437.0)
		done <- err
	}()
	go func() {
		_, err := fakeAdapter.GetState(ctx)
		done <- err
	}()

	// Radio 3 (silvusmock): setPower, setChannel, getState
	go func() {
		err := server.Orchestrator.SetPower(ctx, "silvus-multi-003", 20.0)
		done <- err
	}()
	go func() {
		err := server.Orchestrator.SetChannel(ctx, "silvus-multi-003", 2462.0)
		done <- err
	}()
	go func() {
		_, err := server.Orchestrator.GetState(ctx, "silvus-multi-003")
		done <- err
	}()

	// Wait for all operations
	var errors []error
	for i := 0; i < 9; i++ {
		if err := <-done; err != nil {
			errors = append(errors, err)
		}
	}

	// Assert: All operations succeed (no cross-radio interference)
	if len(errors) > 0 {
		t.Errorf("Concurrent operations failed: %v", errors)
	}

	// Assert: No cross-radio leakage - each radio maintains independent state
	radio1State, err := server.SilvusAdapter.GetState(ctx)
	if err != nil {
		t.Errorf("Failed to get radio 1 state: %v", err)
	}

	radio2State, err := fakeAdapter.GetState(ctx)
	if err != nil {
		t.Errorf("Failed to get radio 2 state: %v", err)
	}

	radio3State, err := silvusAdapter2.GetState(ctx)
	if err != nil {
		t.Errorf("Failed to get radio 3 state: %v", err)
	}

	// Verify independent power states
	if radio1State.PowerDbm == radio2State.PowerDbm {
		t.Error("Radio 1 and Radio 2 should have different power states")
	}
	if radio1State.PowerDbm == radio3State.PowerDbm {
		t.Error("Radio 1 and Radio 3 should have different power states")
	}
	if radio2State.PowerDbm == radio3State.PowerDbm {
		t.Error("Radio 2 and Radio 3 should have different power states")
	}

	// Verify independent frequency states
	if radio1State.Frequency == radio2State.Frequency {
		t.Error("Radio 1 and Radio 2 should have different frequencies")
	}
	if radio1State.Frequency == radio3State.Frequency {
		t.Error("Radio 1 and Radio 3 should have different frequencies")
	}

	t.Logf("✅ Multi-radio concurrency: 3 radios, 9 operations, no cross-radio leakage")
}

// TestMultiRadioConcurrency_AuditIntegrity tests audit integrity under load:
// Verify audit entries are created for all operations and maintain integrity
func TestMultiRadioConcurrency_AuditIntegrity(t *testing.T) {
	// Arrange: Create harness with multiple radios
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-audit-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Add second radio
	fakeAdapter := fake.NewFakeAdapter("fake-audit-002")
	err := server.RadioManager.LoadCapabilities("fake-audit-002", fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load fake adapter capabilities: %v", err)
	}

	ctx := context.Background()

	// Act: Execute operations that should generate audit entries
	operations := []struct {
		radioID string
		action  string
		params  map[string]interface{}
	}{
		{"silvus-audit-001", "setPower", map[string]interface{}{"power": 25.0}},
		{"silvus-audit-001", "setChannel", map[string]interface{}{"frequency": 2412.0}},
		{"fake-audit-002", "setPower", map[string]interface{}{"power": 30.0}},
		{"fake-audit-002", "setChannel", map[string]interface{}{"frequency": 2437.0}},
	}

	for _, op := range operations {
		switch op.action {
		case "setPower":
			err := server.Orchestrator.SetPower(ctx, op.radioID, op.params["power"].(float64))
			if err != nil {
				t.Errorf("SetPower failed for %s: %v", op.radioID, err)
			}
		case "setChannel":
			err := server.Orchestrator.SetChannel(ctx, op.radioID, op.params["frequency"].(float64))
			if err != nil {
				t.Errorf("SetChannel failed for %s: %v", op.radioID, err)
			}
		}
	}

	// Assert: Audit entries were created for all operations
	auditEntries := getAuditEntries(t, server, len(operations))
	if len(auditEntries) != len(operations) {
		t.Errorf("Expected %d audit entries, got %d", len(operations), len(auditEntries))
	}

	// Verify audit entry schema per Architecture §8.6
	for i, entry := range auditEntries {
		if entry.Timestamp.IsZero() {
			t.Errorf("Audit entry %d: timestamp should not be zero", i)
		}
		if entry.RadioID == "" {
			t.Errorf("Audit entry %d: radioID should not be empty", i)
		}
		if entry.Action == "" {
			t.Errorf("Audit entry %d: action should not be empty", i)
		}
		if entry.Result == "" {
			t.Errorf("Audit entry %d: result should not be empty", i)
		}
		if entry.LatencyMS < 0 {
			t.Errorf("Audit entry %d: latency should not be negative", i)
		}
	}

	// Verify all radio IDs are present in audit entries
	radioIDs := make(map[string]bool)
	for _, entry := range auditEntries {
		radioIDs[entry.RadioID] = true
	}
	if !radioIDs["silvus-audit-001"] {
		t.Error("Missing audit entry for silvus-audit-001")
	}
	if !radioIDs["fake-audit-002"] {
		t.Error("Missing audit entry for fake-audit-002")
	}

	t.Logf("✅ Audit integrity: %d entries, all radios represented", len(auditEntries))
}

// TestMultiRadioConcurrency_ResourceContention tests resource contention scenarios:
// Multiple radios competing for resources; verify proper isolation
func TestMultiRadioConcurrency_ResourceContention(t *testing.T) {
	// Arrange: Create harness with multiple radios
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-contention-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Add second radio
	fakeAdapter := fake.NewFakeAdapter("fake-contention-002")
	err := server.RadioManager.LoadCapabilities("fake-contention-002", fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load fake adapter capabilities: %v", err)
	}

	ctx := context.Background()

	// Act: Execute operations that could cause resource contention
	done := make(chan error, 6)

	// Radio 1: Multiple rapid operations
	go func() {
		err := server.Orchestrator.SetPower(ctx, "silvus-contention-001", 25.0)
		done <- err
	}()
	go func() {
		err := server.Orchestrator.SetChannel(ctx, "silvus-contention-001", 2412.0)
		done <- err
	}()
	go func() {
		err := server.Orchestrator.SetPower(ctx, "silvus-contention-001", 30.0)
		done <- err
	}()

	// Radio 2: Multiple rapid operations
	go func() {
		err := fakeAdapter.SetPower(ctx, 20.0)
		done <- err
	}()
	go func() {
		err := fakeAdapter.SetChannel(ctx, 2437.0)
		done <- err
	}()
	go func() {
		err := fakeAdapter.SetPower(ctx, 35.0)
		done <- err
	}()

	// Wait for all operations
	var errors []error
	for i := 0; i < 6; i++ {
		if err := <-done; err != nil {
			errors = append(errors, err)
		}
	}

	// Assert: All operations succeed despite resource contention
	if len(errors) > 0 {
		t.Errorf("Resource contention caused failures: %v", errors)
	}

	// Assert: Final states are correct (last operation wins)
	radio1State, err := server.SilvusAdapter.GetState(ctx)
	if err != nil {
		t.Errorf("Failed to get radio 1 state: %v", err)
	}

	radio2State, err := fakeAdapter.GetState(ctx)
	if err != nil {
		t.Errorf("Failed to get radio 2 state: %v", err)
	}

	// Verify final power states
	if radio1State.PowerDbm != 30.0 {
		t.Errorf("Radio 1 final power should be 30.0, got %v", radio1State.PowerDbm)
	}
	if radio2State.PowerDbm != 35.0 {
		t.Errorf("Radio 2 final power should be 35.0, got %v", radio2State.PowerDbm)
	}

	t.Logf("✅ Resource contention: 6 operations, no interference, final states correct")
}

// Helper functions

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
