//
//
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Response represents the unified envelope format.
type Response struct {
	Result        string      `json:"result"`
	Data          interface{} `json:"data,omitempty"`
	Code          string      `json:"code,omitempty"`
	Message       string      `json:"message,omitempty"`
	Details       interface{} `json:"details,omitempty"`
	CorrelationID string      `json:"correlationId"`
}

// SuccessResponse creates a success response.
func SuccessResponse(data interface{}) *Response {
	return &Response{
		Result:        "ok",
		Data:          data,
		CorrelationID: generateCorrelationID(),
	}
}

// ErrorResponse creates an error response.
func ErrorResponse(code, message string, details interface{}) *Response {
	return &Response{
		Result:        "error",
		Code:          code,
		Message:       message,
		Details:       details,
		CorrelationID: generateCorrelationID(),
	}
}

// WriteSuccess writes a success response to the HTTP response writer.
func WriteSuccess(w http.ResponseWriter, data interface{}) {
	response := SuccessResponse(data)
	writeResponse(w, http.StatusOK, response)
}

// WriteError writes an error response to the HTTP response writer.
func WriteError(w http.ResponseWriter, statusCode int, code, message string, details interface{}) {
	response := ErrorResponse(code, message, details)
	writeResponse(w, statusCode, response)
}

// WriteNotImplemented writes a 501 Not Implemented response.
func WriteNotImplemented(w http.ResponseWriter, endpoint string) {
	WriteError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED",
		fmt.Sprintf("Endpoint %s is not yet implemented", endpoint), nil)
}

// writeResponse writes a JSON response to the HTTP response writer.
func writeResponse(w http.ResponseWriter, statusCode int, response *Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback to plain text if JSON encoding fails
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal server error: %v", err)
	}
}

// generateCorrelationID generates a unique correlation ID.
func generateCorrelationID() string {
	// Simple correlation ID using timestamp and random number
	// In production, use a proper UUID library
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// Standard error responses per OpenAPI v1 §2.2
var (
	ErrInvalidRange = ErrorResponse("INVALID_RANGE", "Invalid parameter range", nil)
	ErrUnauthorized = ErrorResponse("UNAUTHORIZED", "Authentication required", nil)
	ErrForbidden    = ErrorResponse("FORBIDDEN", "Insufficient permissions", nil)
	ErrNotFound     = ErrorResponse("NOT_FOUND", "Resource not found", nil)
	ErrBusy         = ErrorResponse("BUSY", "Service busy, retry with backoff", nil)
	ErrUnavailable  = ErrorResponse("UNAVAILABLE", "Service unavailable", nil)
	ErrInternal     = ErrorResponse("INTERNAL", "Internal server error", nil)
)

// WriteStandardError writes a standard error response.
func WriteStandardError(w http.ResponseWriter, err *Response) {
	var statusCode int

	switch err.Code {
	case "INVALID_RANGE":
		statusCode = http.StatusBadRequest
	case "UNAUTHORIZED":
		statusCode = http.StatusUnauthorized
	case "FORBIDDEN":
		statusCode = http.StatusForbidden
	case "NOT_FOUND":
		statusCode = http.StatusNotFound
	case "BUSY", "UNAVAILABLE":
		statusCode = http.StatusServiceUnavailable
	case "INTERNAL":
		statusCode = http.StatusInternalServerError
	default:
		statusCode = http.StatusInternalServerError
	}

	writeResponse(w, statusCode, err)
}
