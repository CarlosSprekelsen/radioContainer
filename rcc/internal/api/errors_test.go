package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/command"
)

func TestToAPIError(t *testing.T) {
	tests := []struct {
		name           string
		inputError     error
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name:           "nil error returns OK",
			inputError:     nil,
			expectedStatus: http.StatusOK,
			expectedCode:   "",
			expectedMsg:    "",
		},
		{
			name:           "INVALID_RANGE maps to HTTP 400",
			inputError:     adapter.ErrInvalidRange,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_RANGE",
			expectedMsg:    "Parameter value is outside the allowed range",
		},
		{
			name:           "BUSY maps to HTTP 503",
			inputError:     adapter.ErrBusy,
			expectedStatus: http.StatusServiceUnavailable,
			expectedCode:   "BUSY",
			expectedMsg:    "Service is busy, please retry with backoff",
		},
		{
			name:           "UNAVAILABLE maps to HTTP 503",
			inputError:     adapter.ErrUnavailable,
			expectedStatus: http.StatusServiceUnavailable,
			expectedCode:   "UNAVAILABLE",
			expectedMsg:    "Service is temporarily unavailable",
		},
		{
			name:           "INTERNAL maps to HTTP 500",
			inputError:     adapter.ErrInternal,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL",
			expectedMsg:    "Internal server error",
		},
		{
			name:           "command.ErrNotFound maps to HTTP 404",
			inputError:     command.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
			expectedMsg:    "Resource not found",
		},
		{
			name:           "command.ErrInvalidParameter maps to HTTP 400",
			inputError:     command.ErrInvalidParameter,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
			expectedMsg:    "Malformed or missing required parameter",
		},
		{
			name:           "ErrUnauthorizedError maps to HTTP 401",
			inputError:     ErrUnauthorizedError,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
			expectedMsg:    "Authentication required",
		},
		{
			name:           "ErrForbiddenError maps to HTTP 403",
			inputError:     ErrForbiddenError,
			expectedStatus: http.StatusForbidden,
			expectedCode:   "FORBIDDEN",
			expectedMsg:    "Insufficient permissions",
		},
		{
			name:           "ErrNotFoundError maps to HTTP 404",
			inputError:     ErrNotFoundError,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
			expectedMsg:    "Resource not found",
		},
		{
			name:           "unknown error maps to HTTP 500",
			inputError:     errors.New("unknown error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL",
			expectedMsg:    "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := ToAPIError(tt.inputError)

			if status != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, status)
			}

			if tt.inputError == nil {
				if body != nil {
					t.Errorf("Expected nil body for nil error, got %v", body)
				}
				return
			}

			// Parse JSON response
			var response map[string]interface{}
			if err := json.Unmarshal(body, &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response["result"] != "error" {
				t.Errorf("Expected result 'error', got %v", response["result"])
			}

			if response["code"] != tt.expectedCode {
				t.Errorf("Expected code %q, got %q", tt.expectedCode, response["code"])
			}

			if response["message"] != tt.expectedMsg {
				t.Errorf("Expected message %q, got %q", tt.expectedMsg, response["message"])
			}

			// Check correlation ID is present
			if response["correlationId"] == nil || response["correlationId"] == "" {
				t.Errorf("Expected correlationId to be present")
			}
		})
	}
}

func TestToAPIErrorWithVendorError(t *testing.T) {
	// Test with wrapped vendor error
	vendorErr := &adapter.VendorError{
		Code:     adapter.ErrInvalidRange,
		Original: errors.New("TX_POWER_OUT_OF_RANGE"),
		Details:  map[string]interface{}{"power": 50.0},
	}

	status, body := ToAPIError(vendorErr)

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["code"] != "INVALID_RANGE" {
		t.Errorf("Expected code INVALID_RANGE, got %v", response["code"])
	}

	if response["message"] != "Parameter value is outside the allowed range" {
		t.Errorf("Expected specific message, got %v", response["message"])
	}

	// Check that details are preserved
	if response["details"] == nil {
		t.Errorf("Expected details to be preserved")
	}
}

func TestToAPIErrorWithAPIError(t *testing.T) {
	// Test with existing API error
	apiErr := &APIError{
		Code:       "CUSTOM_ERROR",
		Message:    "Custom error message",
		StatusCode: http.StatusTeapot,
		Details:    map[string]interface{}{"custom": "details"},
	}

	status, body := ToAPIError(apiErr)

	if status != http.StatusTeapot {
		t.Errorf("Expected status %d, got %d", http.StatusTeapot, status)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["code"] != "CUSTOM_ERROR" {
		t.Errorf("Expected code CUSTOM_ERROR, got %v", response["code"])
	}

	if response["message"] != "Custom error message" {
		t.Errorf("Expected message 'Custom error message', got %v", response["message"])
	}
}

func TestMapAdapterError(t *testing.T) {
	tests := []struct {
		name           string
		adapterErr     error
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "ErrInvalidRange maps to INVALID_RANGE/400",
			adapterErr:     adapter.ErrInvalidRange,
			expectedCode:   "INVALID_RANGE",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ErrBusy maps to BUSY/503",
			adapterErr:     adapter.ErrBusy,
			expectedCode:   "BUSY",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "ErrUnavailable maps to UNAVAILABLE/503",
			adapterErr:     adapter.ErrUnavailable,
			expectedCode:   "UNAVAILABLE",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "ErrInternal maps to INTERNAL/500",
			adapterErr:     adapter.ErrInternal,
			expectedCode:   "INTERNAL",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "unknown error maps to INTERNAL/500",
			adapterErr:     errors.New("unknown"),
			expectedCode:   "INTERNAL",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, status := mapAdapterError(tt.adapterErr)

			if code != tt.expectedCode {
				t.Errorf("Expected code %q, got %q", tt.expectedCode, code)
			}

			if status != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, status)
			}
		})
	}
}

func TestGetErrorMessage(t *testing.T) {
	tests := []struct {
		name        string
		code        error
		original    error
		expectedMsg string
	}{
		{
			name:        "ErrInvalidRange returns range message",
			code:        adapter.ErrInvalidRange,
			original:    nil,
			expectedMsg: "Parameter value is outside the allowed range",
		},
		{
			name:        "ErrBusy returns busy message",
			code:        adapter.ErrBusy,
			original:    nil,
			expectedMsg: "Service is busy, please retry with backoff",
		},
		{
			name:        "ErrUnavailable returns unavailable message",
			code:        adapter.ErrUnavailable,
			original:    nil,
			expectedMsg: "Service is temporarily unavailable",
		},
		{
			name:        "ErrInternal returns internal message",
			code:        adapter.ErrInternal,
			original:    nil,
			expectedMsg: "Internal server error",
		},
		{
			name:        "unknown code with original returns original",
			code:        errors.New("UNKNOWN"),
			original:    errors.New("original error"),
			expectedMsg: "original error",
		},
		{
			name:        "unknown code without original returns unknown",
			code:        errors.New("UNKNOWN"),
			original:    nil,
			expectedMsg: "Unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := getErrorMessage(tt.code, tt.original)

			if msg != tt.expectedMsg {
				t.Errorf("Expected message %q, got %q", tt.expectedMsg, msg)
			}
		})
	}
}

func TestMarshalErrorResponse(t *testing.T) {
	response := marshalErrorResponse("TEST_ERROR", "Test message", map[string]interface{}{"key": "value"})

	var parsed map[string]interface{}
	if err := json.Unmarshal(response, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	expectedFields := []string{"result", "code", "message", "details", "correlationId"}
	for _, field := range expectedFields {
		if parsed[field] == nil {
			t.Errorf("Expected field %q to be present", field)
		}
	}

	if parsed["result"] != "error" {
		t.Errorf("Expected result 'error', got %v", parsed["result"])
	}

	if parsed["code"] != "TEST_ERROR" {
		t.Errorf("Expected code 'TEST_ERROR', got %v", parsed["code"])
	}

	if parsed["message"] != "Test message" {
		t.Errorf("Expected message 'Test message', got %v", parsed["message"])
	}

	// Check correlation ID format (should be non-empty string)
	correlationID, ok := parsed["correlationId"].(string)
	if !ok || correlationID == "" {
		t.Errorf("Expected correlationId to be non-empty string, got %v", parsed["correlationId"])
	}
}

func TestNewAPIError(t *testing.T) {
	details := map[string]interface{}{"test": true}
	apiErr := NewAPIError("TEST_CODE", "Test message", http.StatusBadRequest, details)

	if apiErr.Code != "TEST_CODE" {
		t.Errorf("Expected code 'TEST_CODE', got %q", apiErr.Code)
	}

	if apiErr.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got %q", apiErr.Message)
	}

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, apiErr.StatusCode)
	}

	// Compare details by string representation to avoid map comparison issues
	expectedStr := ""
	if details != nil {
		expectedStr = fmt.Sprintf("%v", details)
	}
	actualStr := ""
	if apiErr.Details != nil {
		actualStr = fmt.Sprintf("%v", apiErr.Details)
	}
	if expectedStr != actualStr {
		t.Errorf("Expected details %q, got %q", expectedStr, actualStr)
	}

	// Test Error() method
	expectedErrorMsg := "TEST_CODE: Test message"
	if apiErr.Error() != expectedErrorMsg {
		t.Errorf("Expected error message %q, got %q", expectedErrorMsg, apiErr.Error())
	}
}

func TestErrorResponseJSONSchema(t *testing.T) {
	// Test that error response matches OpenAPI schema structure
	_, body := ToAPIError(adapter.ErrInvalidRange)

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify required fields per OpenAPI spec
	requiredFields := map[string]string{
		"result":        "error",
		"code":          "INVALID_RANGE",
		"message":       "Parameter value is outside the allowed range",
		"correlationId": "", // Just check it exists
	}

	for field, expectedValue := range requiredFields {
		actualValue, exists := response[field]
		if !exists {
			t.Errorf("Required field %q missing from response", field)
			continue
		}

		if expectedValue != "" && actualValue != expectedValue {
			t.Errorf("Field %q: expected %q, got %v", field, expectedValue, actualValue)
		}
	}

	// Verify optional details field structure
	if response["details"] != nil {
		// Details should be an object (map) or null
		if _, ok := response["details"].(map[string]interface{}); !ok && response["details"] != nil {
			t.Errorf("Details should be an object or null, got %T", response["details"])
		}
	}
}
