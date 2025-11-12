package audit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}
	defer func() { _ = logger.Close() }()

	if logger == nil {
		t.Fatal("NewLogger() returned nil")
	}

	// Check that log file was created
	expectedPath := filepath.Join(tempDir, "audit.jsonl")
	if logger.GetFilePath() != expectedPath {
		t.Errorf("Expected file path %s, got %s", expectedPath, logger.GetFilePath())
	}

	// Check that file exists
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("Audit log file was not created")
	}
}

func TestNewLoggerWithExistingDir(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create logger
	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}
	defer func() { _ = logger.Close() }()

	// Create another logger in the same directory
	logger2, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed on existing directory: %v", err)
	}
	defer func() { _ = logger2.Close() }()

	if logger2 == nil {
		t.Fatal("NewLogger() returned nil")
	}
}

func TestLogAction(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}
	defer func() { _ = logger.Close() }()

	// Test logging an action
	ctx := context.Background()
	logger.LogAction(ctx, "setPower", "radio-01", "SUCCESS", 100*time.Millisecond)

	// Read the log file and verify content
	logPath := logger.GetFilePath()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON line
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(lines))
	}

	var entry AuditEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("Failed to unmarshal log entry: %v", err)
	}

	// Verify entry fields
	if entry.Action != "setPower" {
		t.Errorf("Expected action 'setPower', got '%s'", entry.Action)
	}
	if entry.RadioID != "radio-01" {
		t.Errorf("Expected radioId 'radio-01', got '%s'", entry.RadioID)
	}
	if entry.Outcome != "SUCCESS" {
		t.Errorf("Expected outcome 'SUCCESS', got '%s'", entry.Outcome)
	}
	if entry.Code != "SUCCESS" {
		t.Errorf("Expected code 'SUCCESS', got '%s'", entry.Code)
	}
	if entry.User != "unknown" {
		t.Errorf("Expected user 'unknown', got '%s'", entry.User)
	}
}

func TestLogControlAction(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}
	defer func() { _ = logger.Close() }()

	// Test logging a control action with parameters
	ctx := context.Background()
	params := map[string]interface{}{
		"powerDbm":     30,
		"frequencyMhz": 2412.0,
	}

	logger.LogControlAction(ctx, "setPower", "radio-01", params, "SUCCESS", nil)

	// Read the log file and verify content
	logPath := logger.GetFilePath()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON line
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(lines))
	}

	var entry AuditEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("Failed to unmarshal log entry: %v", err)
	}

	// Verify entry fields
	if entry.Action != "setPower" {
		t.Errorf("Expected action 'setPower', got '%s'", entry.Action)
	}
	if entry.RadioID != "radio-01" {
		t.Errorf("Expected radioId 'radio-01', got '%s'", entry.RadioID)
	}
	if entry.Outcome != "SUCCESS" {
		t.Errorf("Expected outcome 'SUCCESS', got '%s'", entry.Outcome)
	}
	if entry.Code != "SUCCESS" {
		t.Errorf("Expected code 'SUCCESS', got '%s'", entry.Code)
	}

	// Verify parameters
	if entry.Params == nil {
		t.Error("Expected parameters, got nil")
	}
	if entry.Params["powerDbm"] != float64(30) {
		t.Errorf("Expected powerDbm 30, got %v", entry.Params["powerDbm"])
	}
	if entry.Params["frequencyMhz"] != 2412.0 {
		t.Errorf("Expected frequencyMhz 2412.0, got %v", entry.Params["frequencyMhz"])
	}
}

func TestLogControlActionWithError(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}
	defer func() { _ = logger.Close() }()

	// Test logging a control action with error
	ctx := context.Background()
	params := map[string]interface{}{
		"powerDbm": 50, // Invalid power
	}

	// Simulate an error
	err = &MockError{Code: "INVALID_RANGE", Message: "Power out of range"}
	logger.LogControlAction(ctx, "setPower", "radio-01", params, "FAILED", err)

	// Read the log file and verify content
	logPath := logger.GetFilePath()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON line
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(lines))
	}

	var entry AuditEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("Failed to unmarshal log entry: %v", err)
	}

	// Verify entry fields
	if entry.Action != "setPower" {
		t.Errorf("Expected action 'setPower', got '%s'", entry.Action)
	}
	if entry.RadioID != "radio-01" {
		t.Errorf("Expected radioId 'radio-01', got '%s'", entry.RadioID)
	}
	if entry.Outcome != "FAILED" {
		t.Errorf("Expected outcome 'FAILED', got '%s'", entry.Outcome)
	}
	if entry.Code != "INVALID_RANGE" {
		t.Errorf("Expected code 'INVALID_RANGE', got '%s'", entry.Code)
	}
}

func TestMultipleLogEntries(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}
	defer func() { _ = logger.Close() }()

	// Log multiple actions
	ctx := context.Background()
	logger.LogAction(ctx, "setPower", "radio-01", "SUCCESS", 100*time.Millisecond)
	logger.LogAction(ctx, "setChannel", "radio-01", "SUCCESS", 200*time.Millisecond)
	logger.LogAction(ctx, "selectRadio", "radio-02", "SUCCESS", 50*time.Millisecond)

	// Read the log file and verify content
	logPath := logger.GetFilePath()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON lines
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 log entries, got %d", len(lines))
	}

	// Verify each entry
	expectedActions := []string{"setPower", "setChannel", "selectRadio"}
	expectedRadioIDs := []string{"radio-01", "radio-01", "radio-02"}

	for i, line := range lines {
		var entry AuditEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("Failed to unmarshal log entry %d: %v", i, err)
		}

		if entry.Action != expectedActions[i] {
			t.Errorf("Entry %d: Expected action '%s', got '%s'", i, expectedActions[i], entry.Action)
		}
		if entry.RadioID != expectedRadioIDs[i] {
			t.Errorf("Entry %d: Expected radioId '%s', got '%s'", i, expectedRadioIDs[i], entry.RadioID)
		}
		if entry.Outcome != "SUCCESS" {
			t.Errorf("Entry %d: Expected outcome 'SUCCESS', got '%s'", i, entry.Outcome)
		}
	}
}

func TestGetCodeFromResult(t *testing.T) {
	logger := &Logger{}

	tests := []struct {
		result   string
		expected string
	}{
		{"SUCCESS", "SUCCESS"},
		{"INVALID_RANGE", "INVALID_RANGE"},
		{"UNAVAILABLE", "UNAVAILABLE"},
		{"ERROR", "ERROR"},
		{"UNKNOWN", "UNKNOWN"},
	}

	for _, test := range tests {
		t.Run(test.result, func(t *testing.T) {
			code := logger.getCodeFromResult(test.result)
			if code != test.expected {
				t.Errorf("Expected code '%s', got '%s'", test.expected, code)
			}
		})
	}
}

func TestGetCodeFromError(t *testing.T) {
	logger := &Logger{}

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"nil error", nil, "SUCCESS"},
		{"INVALID_RANGE error", &MockError{Code: "INVALID_RANGE"}, "INVALID_RANGE"},
		{"UNAVAILABLE error", &MockError{Code: "UNAVAILABLE"}, "UNAVAILABLE"},
		{"BUSY error", &MockError{Code: "BUSY"}, "BUSY"},
		{"UNAUTHORIZED error", &MockError{Code: "UNAUTHORIZED"}, "UNAUTHORIZED"},
		{"FORBIDDEN error", &MockError{Code: "FORBIDDEN"}, "FORBIDDEN"},
		{"unknown error", &MockError{Code: "UNKNOWN"}, "ERROR"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := logger.getCodeFromError(test.err)
			if code != test.expected {
				t.Errorf("Expected code '%s', got '%s'", test.expected, code)
			}
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	logger := &Logger{}

	// Test with no user context
	ctx := context.Background()
	user := logger.getUserFromContext(ctx)
	if user != "unknown" {
		t.Errorf("Expected user 'unknown', got '%s'", user)
	}

	// Test with user context
	ctxWithUser := context.WithValue(ctx, "claims", map[string]interface{}{
		"sub": "user-123",
	})
	user = logger.getUserFromContext(ctxWithUser)
	if user != "user-123" {
		t.Errorf("Expected user 'user-123', got '%s'", user)
	}
}

func TestGetParamsFromContext(t *testing.T) {
	logger := &Logger{}

	// Test with no params context
	ctx := context.Background()
	params := logger.getParamsFromContext(ctx)
	if params == nil {
		t.Error("Expected empty params map, got nil")
	}
	if len(params) != 0 {
		t.Errorf("Expected empty params map, got %d items", len(params))
	}

	// Test with params context
	expectedParams := map[string]interface{}{
		"powerDbm":     30,
		"frequencyMhz": 2412.0,
	}
	ctxWithParams := context.WithValue(ctx, "params", expectedParams)
	params = logger.getParamsFromContext(ctxWithParams)
	if params == nil {
		t.Error("Expected params, got nil")
	}
	if len(params) != len(expectedParams) {
		t.Errorf("Expected %d params, got %d", len(expectedParams), len(params))
	}
}

func TestClose(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	// Close the logger
	if err := logger.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// Close again should not error
	if err := logger.Close(); err != nil {
		t.Errorf("Close() on already closed logger failed: %v", err)
	}
}

func TestRotate(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}
	defer func() { _ = logger.Close() }()

	// Write some data
	ctx := context.Background()
	logger.LogAction(ctx, "setPower", "radio-01", "SUCCESS", 100*time.Millisecond)

	// Rotate the log
	if err := logger.Rotate(); err != nil {
		t.Errorf("Rotate() failed: %v", err)
	}

	// Write more data to new file
	logger.LogAction(ctx, "setChannel", "radio-01", "SUCCESS", 200*time.Millisecond)

	// Check that both files exist
	logPath := logger.GetFilePath()
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("New log file was not created after rotation")
	}

	// Check that rotated file exists
	rotatedFiles, err := filepath.Glob(logPath + ".*")
	if err != nil {
		t.Errorf("Failed to find rotated files: %v", err)
	}
	if len(rotatedFiles) != 1 {
		t.Errorf("Expected 1 rotated file, found %d", len(rotatedFiles))
	}
}

func TestConcurrentLogging(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}
	defer func() { _ = logger.Close() }()

	// Log concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			ctx := context.Background()
			logger.LogAction(ctx, "setPower", "radio-01", "SUCCESS", 100*time.Millisecond)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Read the log file and verify content
	logPath := logger.GetFilePath()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON lines
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 10 {
		t.Fatalf("Expected 10 log entries, got %d", len(lines))
	}

	// Verify all entries are valid JSON
	for i, line := range lines {
		var entry AuditEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("Failed to unmarshal log entry %d: %v", i, err)
		}
		if entry.Action != "setPower" {
			t.Errorf("Entry %d: Expected action 'setPower', got '%s'", i, entry.Action)
		}
	}
}

// MockError is a test error type
type MockError struct {
	Code    string
	Message string
}

func (e *MockError) Error() string {
	return e.Code + ": " + e.Message
}
