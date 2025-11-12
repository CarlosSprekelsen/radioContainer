package api

import (
	"encoding/json"
	"testing"

	"github.com/radio-control/rcc/internal/adapter"
)

// TestAPIEnvelopeHelpers tests envelope creation and error mapping
func TestAPIEnvelopeHelpers(t *testing.T) {
	tests := []struct {
		name           string
		error          error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "invalid_range",
			error:          adapter.ErrInvalidRange,
			expectedStatus: 400,
			expectedCode:   "INVALID_RANGE",
		},
		{
			name:           "busy",
			error:          adapter.ErrBusy,
			expectedStatus: 503,
			expectedCode:   "BUSY",
		},
		{
			name:           "unavailable",
			error:          adapter.ErrUnavailable,
			expectedStatus: 503,
			expectedCode:   "UNAVAILABLE",
		},
		{
			name:           "internal",
			error:          adapter.ErrInternal,
			expectedStatus: 500,
			expectedCode:   "INTERNAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ToAPIError mapping
			status, body := ToAPIError(tt.error)
			
			if status != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, status)
			}

			// Parse response body
			var response Response
			if err := json.Unmarshal(body, &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Code != tt.expectedCode {
				t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, response.Code)
			}

			// Assert envelope fields
			mustEnvelope(t, map[string]any{
				"result":        response.Result,
				"correlationId": response.CorrelationID,
			})
		})
	}
}

// TestSuccessResponse tests success response creation
func TestSuccessResponse(t *testing.T) {
	data := map[string]interface{}{
		"radioId": "test-radio",
		"power":   20.0,
	}

	response := SuccessResponse(data)

	// Assert envelope fields
	mustEnvelope(t, map[string]any{
		"result":        response.Result,
		"correlationId": response.CorrelationID,
	})

	if response.Result != "ok" {
		t.Errorf("Expected result 'ok', got '%s'", response.Result)
	}

	if response.Data == nil {
		t.Errorf("Expected data field to be present")
	}
}

// TestErrorResponse tests error response creation
func TestErrorResponse(t *testing.T) {
	response := ErrorResponse("TEST_ERROR", "Test error message", nil)

	// Assert envelope fields
	mustEnvelope(t, map[string]any{
		"result":        response.Result,
		"correlationId": response.CorrelationID,
	})

	if response.Result != "error" {
		t.Errorf("Expected result 'error', got '%s'", response.Result)
	}

	if response.Code != "TEST_ERROR" {
		t.Errorf("Expected code 'TEST_ERROR', got '%s'", response.Code)
	}

	if response.Message != "Test error message" {
		t.Errorf("Expected message 'Test error message', got '%s'", response.Message)
	}
}

// TestCorrelationIDGeneration tests that correlation IDs are generated
func TestCorrelationIDGeneration(t *testing.T) {
	// Test multiple calls generate different IDs
	responses := make([]*Response, 5)
	for i := 0; i < 5; i++ {
		responses[i] = SuccessResponse(map[string]interface{}{"test": i})
	}

	// All should have correlation IDs
	for i, resp := range responses {
		if resp.CorrelationID == "" {
			t.Errorf("Response %d missing correlation ID", i)
		}
	}

	// Should be different (very unlikely to be the same)
	seen := make(map[string]bool)
	for _, resp := range responses {
		if seen[resp.CorrelationID] {
			t.Errorf("Duplicate correlation ID found: %s", resp.CorrelationID)
		}
		seen[resp.CorrelationID] = true
	}
}

// Helper functions for unit tests
func mustEnvelope(t *testing.T, got map[string]any) {
	t.Helper()
	if got["result"] == nil {
		t.Fatalf("missing result field")
	}
	if got["correlationId"] == nil || got["correlationId"] == "" {
		t.Fatalf("missing or empty correlationId field")
	}
}

func normalize(m map[string]any, drop ...string) map[string]any {
	rm := map[string]any{}
	for k, v := range m {
		skip := false
		for _, d := range drop {
			if k == d {
				skip = true
				break
			}
		}
		if !skip {
			rm[k] = v
		}
	}
	return rm
}