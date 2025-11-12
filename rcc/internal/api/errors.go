//
//
package api

import (
    "encoding/json"
    "errors"
    "fmt"
    "net/http"

    "github.com/radio-control/rcc/internal/adapter"
    "github.com/radio-control/rcc/internal/command"
)

// APIError represents an API-layer error with HTTP status code.
type APIError struct {
	Code       string
	Message    string
	Details    interface{}
	StatusCode int
}

// API error codes for transport/security/lookup conditions
var (
	ErrBadRequest        = errors.New("BAD_REQUEST")
	ErrUnauthorizedError = errors.New("UNAUTHORIZED")
	ErrForbiddenError    = errors.New("FORBIDDEN")
	ErrNotFoundError     = errors.New("NOT_FOUND")
)

// ToAPIError converts an error to an API error with HTTP status code and JSON body.
func ToAPIError(err error) (int, []byte) {
	if err == nil {
		return http.StatusOK, nil
	}

	var apiErr *APIError
	var vendorErr *adapter.VendorError

	// Check if it's already an API error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode, marshalErrorResponse(apiErr.Code, apiErr.Message, apiErr.Details)
	}

	// Check if it's a vendor error from adapter
	if errors.As(err, &vendorErr) {
		// Map adapter error to API error
		code, statusCode := mapAdapterError(vendorErr.Code)
		message := getErrorMessage(vendorErr.Code, vendorErr.Original)
		return statusCode, marshalErrorResponse(code, message, vendorErr.Details)
	}

	// Check for adapter error codes
	if errors.Is(err, adapter.ErrInvalidRange) {
		return http.StatusBadRequest, marshalErrorResponse("INVALID_RANGE", getErrorMessage(adapter.ErrInvalidRange, err), nil)
	}
	if errors.Is(err, adapter.ErrBusy) {
		return http.StatusServiceUnavailable, marshalErrorResponse("BUSY", getErrorMessage(adapter.ErrBusy, err), nil)
	}
	if errors.Is(err, adapter.ErrUnavailable) {
		return http.StatusServiceUnavailable, marshalErrorResponse("UNAVAILABLE", getErrorMessage(adapter.ErrUnavailable, err), nil)
	}
	if errors.Is(err, adapter.ErrInternal) {
		return http.StatusInternalServerError, marshalErrorResponse("INTERNAL", getErrorMessage(adapter.ErrInternal, err), nil)
	}

	// Check for API-layer errors
	if errors.Is(err, command.ErrNotFound) {
		return http.StatusNotFound, marshalErrorResponse("NOT_FOUND", "Resource not found", nil)
	}
    if errors.Is(err, command.ErrInvalidParameter) {
        return http.StatusBadRequest, marshalErrorResponse("BAD_REQUEST", "Malformed or missing required parameter", nil)
    }
	if errors.Is(err, ErrUnauthorizedError) {
		return http.StatusUnauthorized, marshalErrorResponse("UNAUTHORIZED", "Authentication required", nil)
	}
	if errors.Is(err, ErrForbiddenError) {
		return http.StatusForbidden, marshalErrorResponse("FORBIDDEN", "Insufficient permissions", nil)
	}
	if errors.Is(err, ErrNotFoundError) {
		return http.StatusNotFound, marshalErrorResponse("NOT_FOUND", "Resource not found", nil)
	}

	// Default to internal server error for unknown errors
	return http.StatusInternalServerError, marshalErrorResponse("INTERNAL", "Internal server error", map[string]interface{}{
		"original": err.Error(),
	})
}

// mapAdapterError maps adapter error codes to API error codes and HTTP status codes.
func mapAdapterError(adapterErr error) (string, int) {
	switch {
	case errors.Is(adapterErr, adapter.ErrInvalidRange):
		return "INVALID_RANGE", http.StatusBadRequest
	case errors.Is(adapterErr, adapter.ErrBusy):
		return "BUSY", http.StatusServiceUnavailable
	case errors.Is(adapterErr, adapter.ErrUnavailable):
		return "UNAVAILABLE", http.StatusServiceUnavailable
	case errors.Is(adapterErr, adapter.ErrInternal):
		return "INTERNAL", http.StatusInternalServerError
	default:
		return "INTERNAL", http.StatusInternalServerError
	}
}

// getErrorMessage returns a user-friendly error message for the given error.
func getErrorMessage(code error, original error) string {
	switch {
	case errors.Is(code, adapter.ErrInvalidRange):
		return "Parameter value is outside the allowed range"
	case errors.Is(code, adapter.ErrBusy):
		return "Service is busy, please retry with backoff"
	case errors.Is(code, adapter.ErrUnavailable):
		return "Service is temporarily unavailable"
	case errors.Is(code, adapter.ErrInternal):
		return "Internal server error"
	case errors.Is(code, ErrUnauthorizedError):
		return "Authentication required"
	case errors.Is(code, ErrForbiddenError):
		return "Insufficient permissions"
	case errors.Is(code, ErrNotFoundError):
		return "Resource not found"
	default:
		if original != nil {
			return original.Error()
		}
		return "Unknown error"
	}
}

// marshalErrorResponse creates a JSON error response with correlation ID.
func marshalErrorResponse(code, message string, details interface{}) []byte {
	response := Response{
		Result:        "error",
		Code:          code,
		Message:       message,
		Details:       details,
		CorrelationID: generateCorrelationID(),
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		// Fallback error response if marshaling fails
		fallback := map[string]interface{}{
			"result":        "error",
			"code":          "INTERNAL",
			"message":       "Failed to marshal error response",
			"correlationId": generateCorrelationID(),
		}
		jsonBytes, _ := json.Marshal(fallback)
		return jsonBytes
	}

	return jsonBytes
}

// NewAPIError creates a new API error.
func NewAPIError(code string, message string, statusCode int, details interface{}) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: statusCode,
	}
}

// Error implements the error interface for APIError.
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
