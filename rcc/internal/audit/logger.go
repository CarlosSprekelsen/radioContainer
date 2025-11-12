//
//
package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AuditEntry represents a single audit log entry.
type AuditEntry struct {
	Timestamp time.Time              `json:"ts"`
	User      string                 `json:"user"`
	RadioID   string                 `json:"radioId"`
	Action    string                 `json:"action"`
	Params    map[string]interface{} `json:"params"`
	Outcome   string                 `json:"outcome"`
	Code      string                 `json:"code"`
}

// Logger implements the audit logging functionality.
type Logger struct {
	mu       sync.Mutex
	filePath string
	file     *os.File
}

// NewLogger creates a new audit logger.
func NewLogger(logDir string) (*Logger, error) {
	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create audit log file path
	filePath := filepath.Join(logDir, "audit.jsonl")

	// Open file for append-only writing
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	return &Logger{
		filePath: filePath,
		file:     file,
	}, nil
}

// LogAction logs an audit record for a command action.
func (l *Logger) LogAction(ctx context.Context, action, radioID, result string, latency time.Duration) {
	// Extract user from context (if available)
	user := l.getUserFromContext(ctx)
	
	// Create audit entry
	entry := AuditEntry{
		Timestamp: time.Now().UTC(),
		User:      user,
		RadioID:   radioID,
		Action:    action,
		Params:    l.getParamsFromContext(ctx),
		Outcome:   result,
		Code:      l.getCodeFromResult(result),
	}

	// Write to log file
	l.writeEntry(entry)
}

// LogControlAction logs a control action with detailed parameters.
func (l *Logger) LogControlAction(ctx context.Context, action, radioID string, params map[string]interface{}, outcome string, err error) {
	// Extract user from context (if available)
	user := l.getUserFromContext(ctx)
	
	// Determine result code
	code := "SUCCESS"
	if err != nil {
		code = l.getCodeFromError(err)
	}

	// Create audit entry
	entry := AuditEntry{
		Timestamp: time.Now().UTC(),
		User:      user,
		RadioID:   radioID,
		Action:    action,
		Params:    params,
		Outcome:   outcome,
		Code:      code,
	}

	// Write to log file
	l.writeEntry(entry)
}

// writeEntry writes an audit entry to the log file.
func (l *Logger) writeEntry(entry AuditEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Marshal entry to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Log error to stderr if JSON marshaling fails
		fmt.Fprintf(os.Stderr, "Failed to marshal audit entry: %v\n", err)
		return
	}

	// Write JSON line to file
	if _, err := l.file.Write(append(jsonData, '\n')); err != nil {
		// Log error to stderr if file write fails
		fmt.Fprintf(os.Stderr, "Failed to write audit entry: %v\n", err)
		return
	}

	// Flush to ensure data is written to disk
	if err := l.file.Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to sync audit log: %v\n", err)
	}
}

// getUserFromContext extracts user information from the request context.
func (l *Logger) getUserFromContext(ctx context.Context) string {
	// Try to get user from auth claims
	// This would be populated by the auth middleware
	if claims, ok := ctx.Value("claims").(map[string]interface{}); ok {
		if subject, ok := claims["sub"].(string); ok {
			return subject
		}
	}
	
	// Default to "unknown" if no user context
	return "unknown"
}

// getParamsFromContext extracts parameters from the request context.
func (l *Logger) getParamsFromContext(ctx context.Context) map[string]interface{} {
	// Try to get parameters from context
	if params, ok := ctx.Value("params").(map[string]interface{}); ok {
		return params
	}
	
	// Return empty map if no parameters
	return make(map[string]interface{})
}

// getCodeFromResult maps result strings to standardized codes.
func (l *Logger) getCodeFromResult(result string) string {
	switch result {
	case "SUCCESS":
		return "SUCCESS"
	case "INVALID_RANGE":
		return "INVALID_RANGE"
	case "UNAVAILABLE":
		return "UNAVAILABLE"
	case "ERROR":
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// getCodeFromError maps error types to standardized codes.
func (l *Logger) getCodeFromError(err error) string {
	if err == nil {
		return "SUCCESS"
	}
	
	// Check for specific error types
	errStr := err.Error()
	if contains(errStr, "INVALID_RANGE") {
		return "INVALID_RANGE"
	}
	if contains(errStr, "UNAVAILABLE") {
		return "UNAVAILABLE"
	}
	if contains(errStr, "BUSY") {
		return "BUSY"
	}
	if contains(errStr, "UNAUTHORIZED") {
		return "UNAUTHORIZED"
	}
	if contains(errStr, "FORBIDDEN") {
		return "FORBIDDEN"
	}
	
	// Default to ERROR for unknown errors
	return "ERROR"
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Close closes the audit logger and its file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}

// GetFilePath returns the path to the audit log file.
func (l *Logger) GetFilePath() string {
	return l.filePath
}

// Rotate rotates the audit log file.
// This is a placeholder for future log rotation functionality.
func (l *Logger) Rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Close current file
	if l.file != nil {
		if err := l.file.Close(); err != nil {
			return fmt.Errorf("failed to close current log file: %w", err)
		}
	}
	
	// Create new file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	newFilePath := fmt.Sprintf("%s.%s", l.filePath, timestamp)
	
	// Rename current file
	if err := os.Rename(l.filePath, newFilePath); err != nil {
		return fmt.Errorf("failed to rename log file: %w", err)
	}
	
	// Open new file
	file, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new log file: %w", err)
	}
	
	l.file = file
	return nil
}
