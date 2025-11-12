package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/command"
)

// testError is a simple error type for testing
type testError struct {
	Message string
}

func (e *testError) Error() string {
	return e.Message
}

// TestErrorCodeConsistency tests that error codes are consistent with API v1 specification
func TestErrorCodeConsistency(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "INVALID_RANGE maps to 400",
			err:            adapter.ErrInvalidRange,
			expectedCode:   "INVALID_RANGE",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "BUSY maps to 503",
			err:            adapter.ErrBusy,
			expectedCode:   "BUSY",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "UNAVAILABLE maps to 503",
			err:            adapter.ErrUnavailable,
			expectedCode:   "UNAVAILABLE",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "INTERNAL maps to 500",
			err:            adapter.ErrInternal,
			expectedCode:   "INTERNAL",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "NOT_FOUND maps to 404",
			err:            command.ErrNotFound,
			expectedCode:   "NOT_FOUND",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "UNAUTHORIZED maps to 401",
			err:            ErrUnauthorizedError,
			expectedCode:   "UNAUTHORIZED",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "FORBIDDEN maps to 403",
			err:            ErrForbiddenError,
			expectedCode:   "FORBIDDEN",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, responseBody := ToAPIError(tt.err)

			if statusCode != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, statusCode)
			}

			// Parse response body
			var response Response
			if err := json.Unmarshal(responseBody, &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Code != tt.expectedCode {
				t.Errorf("Expected error code %s, got %s", tt.expectedCode, response.Code)
			}

			if response.Result != "error" {
				t.Errorf("Expected result 'error', got %s", response.Result)
			}

			if response.CorrelationID == "" {
				t.Error("Expected correlation ID to be present")
			}
		})
	}
}

// TestVendorErrorMapping tests that vendor errors are properly mapped
func TestVendorErrorMapping(t *testing.T) {
	vendorErr := &adapter.VendorError{
		Code:     adapter.ErrInvalidRange,
		Original: &adapter.VendorError{Code: adapter.ErrInvalidRange, Original: nil},
		Details: map[string]interface{}{
			"parameter": "powerDbm",
			"min":       0,
			"max":       39,
		},
	}

	statusCode, responseBody := ToAPIError(vendorErr)

	if statusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, statusCode)
	}

	var response Response
	if err := json.Unmarshal(responseBody, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != "INVALID_RANGE" {
		t.Errorf("Expected error code INVALID_RANGE, got %s", response.Code)
	}

	if response.Details == nil {
		t.Error("Expected details to be preserved from vendor error")
	}
}

// TestAPIErrorPreservation tests that API errors are preserved correctly
func TestAPIErrorPreservation(t *testing.T) {
	apiErr := NewAPIError("CUSTOM_ERROR", "Custom error message", http.StatusBadRequest, map[string]interface{}{
		"field": "value",
	})

	statusCode, responseBody := ToAPIError(apiErr)

	if statusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, statusCode)
	}

	var response Response
	if err := json.Unmarshal(responseBody, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != "CUSTOM_ERROR" {
		t.Errorf("Expected error code CUSTOM_ERROR, got %s", response.Code)
	}

	if response.Message != "Custom error message" {
		t.Errorf("Expected message 'Custom error message', got %s", response.Message)
	}
}

// TestUnknownErrorHandling tests that unknown errors are handled gracefully
func TestUnknownErrorHandling(t *testing.T) {
	unknownErr := &testError{Message: "Unknown error"}

	statusCode, responseBody := ToAPIError(unknownErr)

	if statusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, statusCode)
	}

	var response Response
	if err := json.Unmarshal(responseBody, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != "INTERNAL" {
		t.Errorf("Expected error code INTERNAL, got %s", response.Code)
	}

	if response.Details == nil {
		t.Error("Expected details to contain original error")
	}
}

// TestErrorResponseFormat tests that error responses follow the unified envelope format
func TestErrorResponseFormat(t *testing.T) {
	_, responseBody := ToAPIError(adapter.ErrInvalidRange)

	var response Response
	if err := json.Unmarshal(responseBody, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate unified envelope format
	if response.Result != "error" {
		t.Errorf("Expected result 'error', got %s", response.Result)
	}

	if response.Code == "" {
		t.Error("Expected error code to be present")
	}

	if response.Message == "" {
		t.Error("Expected error message to be present")
	}

	if response.CorrelationID == "" {
		t.Error("Expected correlation ID to be present")
	}

	// Data should be omitted for error responses
	if response.Data != nil {
		t.Error("Expected data to be omitted for error responses")
	}
}

// TestWriteStandardError tests the WriteStandardError function
func TestWriteStandardError(t *testing.T) {
	tests := []struct {
		name           string
		errorResponse  *Response
		expectedStatus int
	}{
		{
			name:           "INVALID_RANGE returns 400",
			errorResponse:  ErrInvalidRange,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "UNAUTHORIZED returns 401",
			errorResponse:  ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "FORBIDDEN returns 403",
			errorResponse:  ErrForbidden,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "NOT_FOUND returns 404",
			errorResponse:  ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "BUSY returns 503",
			errorResponse:  ErrBusy,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "UNAVAILABLE returns 503",
			errorResponse:  ErrUnavailable,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "INTERNAL returns 500",
			errorResponse:  ErrInternal,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteStandardError(w, tt.errorResponse)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify content type
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json; charset=utf-8" {
				t.Errorf("Expected content type 'application/json; charset=utf-8', got %s", contentType)
			}

			// Verify response format
			var response Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Result != "error" {
				t.Errorf("Expected result 'error', got %s", response.Result)
			}
		})
	}
}

// TestCorrelationIDUniqueness tests that correlation IDs are unique
func TestCorrelationIDUniqueness(t *testing.T) {
	// Generate multiple correlation IDs
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateCorrelationID()
		if ids[id] {
			t.Errorf("Duplicate correlation ID generated: %s", id)
		}
		ids[id] = true
	}
}

// TestMarshalErrorResponseFallback tests the fallback error response
func TestMarshalErrorResponseFallback(t *testing.T) {
	// This test would require mocking json.Marshal to fail
	// For now, we'll just test that the function doesn't panic
	response := marshalErrorResponse("TEST", "Test message", nil)

	if len(response) == 0 {
		t.Error("Expected non-empty response")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(response, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if parsed["result"] != "error" {
		t.Errorf("Expected result 'error', got %v", parsed["result"])
	}
}
