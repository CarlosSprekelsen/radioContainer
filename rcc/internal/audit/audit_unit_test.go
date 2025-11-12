package audit

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestAuditLogger_UnitTests tests the audit logger in isolation without orchestrator dependencies
func TestAuditLogger_UnitTests(t *testing.T) {
	// Create temporary directory for audit logs
	tempDir := t.TempDir()
	auditLogPath := filepath.Join(tempDir, "audit.jsonl")

	// Create audit logger
	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer func() { _ = logger.Close() }()

	// Test 1: Log successful action
	t.Run("LogSuccessfulAction", func(t *testing.T) {
		ctx := context.Background()
		logger.LogAction(ctx, "setPower", "silvus-001", "SUCCESS", 100*time.Millisecond)

		// Wait for log to be written
		time.Sleep(10 * time.Millisecond)

		// Read and verify the log entry
		lines, err := readLastNLines(auditLogPath, 1)
		if err != nil {
			t.Fatalf("Failed to read audit log: %v", err)
		}

		if len(lines) == 0 {
			t.Fatal("No audit log entries found")
		}

		var entry AuditEntry
		if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
			t.Fatalf("Failed to unmarshal audit entry: %v", err)
		}

		// Verify the entry
		if entry.User != "unknown" {
			t.Errorf("Expected user 'unknown', got '%s'", entry.User)
		}
		if entry.RadioID != "silvus-001" {
			t.Errorf("Expected radioId 'silvus-001', got '%s'", entry.RadioID)
		}
		if entry.Action != "setPower" {
			t.Errorf("Expected action 'setPower', got '%s'", entry.Action)
		}
		if entry.Outcome != "SUCCESS" {
			t.Errorf("Expected outcome 'SUCCESS', got '%s'", entry.Outcome)
		}
		if entry.Code != "SUCCESS" {
			t.Errorf("Expected code 'SUCCESS', got '%s'", entry.Code)
		}
		if entry.Params == nil {
			t.Error("Expected params to be present")
		}
		if entry.Timestamp.IsZero() {
			t.Error("Expected timestamp to be set")
		}
	})

	// Test 2: Log error action
	t.Run("LogErrorAction", func(t *testing.T) {
		ctx := context.Background()
		logger.LogAction(ctx, "setChannel", "silvus-001", "INVALID_RANGE", 50*time.Millisecond)

		// Wait for log to be written
		time.Sleep(10 * time.Millisecond)

		// Read and verify the log entry
		lines, err := readLastNLines(auditLogPath, 1)
		if err != nil {
			t.Fatalf("Failed to read audit log: %v", err)
		}

		if len(lines) == 0 {
			t.Fatal("No audit log entries found")
		}

		var entry AuditEntry
		if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
			t.Fatalf("Failed to unmarshal audit entry: %v", err)
		}

		// Verify the entry
		if entry.Outcome != "INVALID_RANGE" {
			t.Errorf("Expected outcome 'INVALID_RANGE', got '%s'", entry.Outcome)
		}
		if entry.Code != "INVALID_RANGE" {
			t.Errorf("Expected code 'INVALID_RANGE', got '%s'", entry.Code)
		}
	})

	// Test 3: Log multiple actions and verify append-only behavior
	t.Run("LogMultipleActions", func(t *testing.T) {
		ctx := context.Background()

		// Log multiple actions
		actions := []struct {
			action  string
			outcome string
		}{
			{"setPower", "SUCCESS"},
			{"setChannel", "UNAVAILABLE"},
			{"selectRadio", "BUSY"},
		}

		for _, action := range actions {
			logger.LogAction(ctx, action.action, "silvus-001", action.outcome, 100*time.Millisecond)
		}

		// Wait for logs to be written
		time.Sleep(50 * time.Millisecond)

		// Read all log entries
		lines, err := readLastNLines(auditLogPath, 5) // Should have at least 5 entries now
		if err != nil {
			t.Fatalf("Failed to read audit log: %v", err)
		}

		if len(lines) < 5 {
			t.Fatalf("Expected at least 5 log entries, got %d", len(lines))
		}

		// Verify the last 3 entries match our actions
		for i, action := range actions {
			entryIndex := len(lines) - 3 + i
			var entry AuditEntry
			if err := json.Unmarshal([]byte(lines[entryIndex]), &entry); err != nil {
				t.Fatalf("Failed to unmarshal audit entry %d: %v", entryIndex, err)
			}

			if entry.Action != action.action {
				t.Errorf("Entry %d: Expected action '%s', got '%s'", entryIndex, action.action, entry.Action)
			}
			if entry.Outcome != action.outcome {
				t.Errorf("Entry %d: Expected outcome '%s', got '%s'", entryIndex, action.outcome, entry.Outcome)
			}
		}
	})
}

// TestAuditLogger_SchemaValidation tests that audit entries have all required fields
func TestAuditLogger_SchemaValidation(t *testing.T) {
	tempDir := t.TempDir()
	auditLogPath := filepath.Join(tempDir, "audit.jsonl")

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer func() { _ = logger.Close() }()

	// Log a sample action
	ctx := context.Background()
	logger.LogAction(ctx, "setPower", "silvus-001", "SUCCESS", 100*time.Millisecond)

	// Wait for log to be written
	time.Sleep(10 * time.Millisecond)

	// Read and validate schema
	lines, err := readLastNLines(auditLogPath, 1)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}

	var entry AuditEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("Failed to unmarshal audit entry: %v", err)
	}

	// Verify all required fields are present and non-empty
	requiredFields := map[string]string{
		"ts":      entry.Timestamp.Format(time.RFC3339Nano),
		"user":    entry.User,
		"radioId": entry.RadioID,
		"action":  entry.Action,
		"outcome": entry.Outcome,
		"code":    entry.Code,
	}

	for field, value := range requiredFields {
		if value == "" {
			t.Errorf("Required field '%s' is empty", field)
		}
	}

	// Verify params field exists (can be empty map)
	if entry.Params == nil {
		t.Error("Params field should be present (can be empty map)")
	}

	// Verify timestamp is recent
	if time.Since(entry.Timestamp) > time.Minute {
		t.Errorf("Timestamp is too old: %v", entry.Timestamp)
	}
}

// readLastNLines reads the last N lines from a file
func readLastNLines(filename string, n int) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:] // Keep only the last n lines
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
