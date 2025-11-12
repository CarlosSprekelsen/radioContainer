package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
)

// TestMalformedJSONRequests tests API with malformed JSON requests
func TestMalformedJSONRequests(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	rm := radio.NewManager()
	orch := command.NewOrchestrator(hub, cfg)
	server := NewServer(hub, orch, rm, 30*time.Second, 30*time.Second, 120*time.Second)

	tests := []struct {
		name           string
		path           string
		body           string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Malformed JSON in SetPower",
			path:           "/api/v1/radios/radio-01/power",
			body:           `{"powerDbm": 30.0,`, // Missing closing brace
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Invalid JSON syntax in SetChannel",
			path:           "/api/v1/radios/radio-01/channel",
			body:           `{"frequencyMhz": 2412.0, "invalid": }`, // Invalid JSON
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Empty JSON object",
			path:           "/api/v1/radios/radio-01/power",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Non-JSON content",
			path:           "/api/v1/radios/radio-01/power",
			body:           `not json at all`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Null JSON",
			path:           "/api/v1/radios/radio-01/power",
			body:           `null`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Array instead of object",
			path:           "/api/v1/radios/radio-01/power",
			body:           `[30.0]`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Use appropriate handler based on path
			if strings.Contains(tt.path, "/power") {
				server.handleSetPower(w, req, "radio-01")
			} else if strings.Contains(tt.path, "/channel") {
				server.handleSetChannel(w, req, "radio-01")
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify error response format
			var response Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Result != "error" {
				t.Errorf("Expected result 'error', got '%s'", response.Result)
			}

			if response.Code != tt.expectedCode {
				t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, response.Code)
			}

			if response.CorrelationID == "" {
				t.Error("Expected correlation ID to be present")
			}
		})
	}
}

// TestMissingAuthentication tests API with missing authentication
func TestMissingAuthentication(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	rm := radio.NewManager()
	orch := command.NewOrchestrator(hub, cfg)
	server := NewServer(hub, orch, rm, 30*time.Second, 30*time.Second, 120*time.Second)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Missing Authorization header",
			method:         "GET",
			path:           "/api/v1/radios",
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
		},
		{
			name:           "Empty Authorization header",
			method:         "GET",
			path:           "/api/v1/radios",
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
		},
		{
			name:           "Invalid Authorization format",
			method:         "GET",
			path:           "/api/v1/radios",
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			// Set invalid or missing authorization based on test case
			switch tt.name {
			case "Missing Authorization header":
				// No authorization header
			case "Empty Authorization header":
				req.Header.Set("Authorization", "")
			case "Invalid Authorization format":
				req.Header.Set("Authorization", "InvalidFormat token")
			}

			w := httptest.NewRecorder()
			server.handleRadios(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify error response format
			var response Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Result != "error" {
				t.Errorf("Expected result 'error', got '%s'", response.Result)
			}

			if response.Code != tt.expectedCode {
				t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, response.Code)
			}
		})
	}
}

// TestNotFoundRadio tests API with non-existent radio
func TestNotFoundRadio(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	rm := radio.NewManager()
	orch := command.NewOrchestrator(hub, cfg)
	server := NewServer(hub, orch, rm, 30*time.Second, 30*time.Second, 120*time.Second)

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "SetPower on non-existent radio",
			method:         "POST",
			path:           "/api/v1/radios/non-existent-radio/power",
			body:           `{"powerDbm": 30.0}`,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
		{
			name:           "SetChannel on non-existent radio",
			method:         "POST",
			path:           "/api/v1/radios/non-existent-radio/channel",
			body:           `{"frequencyMhz": 2412.0}`,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
		{
			name:           "GetState on non-existent radio",
			method:         "GET",
			path:           "/api/v1/radios/non-existent-radio/state",
			body:           "",
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
		{
			name:           "SelectRadio on non-existent radio",
			method:         "POST",
			path:           "/api/v1/radios/non-existent-radio/select",
			body:           `{}`,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()

			// Use appropriate handler based on path
			if strings.Contains(tt.path, "/power") {
				server.handleSetPower(w, req, "non-existent-radio")
			} else if strings.Contains(tt.path, "/channel") {
				server.handleSetChannel(w, req, "non-existent-radio")
			} else if strings.Contains(tt.path, "/state") {
				server.handleRadioByID(w, req)
			} else if strings.Contains(tt.path, "/select") {
				server.handleSelectRadio(w, req)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify error response format
			var response Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Result != "error" {
				t.Errorf("Expected result 'error', got '%s'", response.Result)
			}

			if response.Code != tt.expectedCode {
				t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, response.Code)
			}
		})
	}
}

// TestInvalidParameterValues tests API with invalid parameter values
func TestInvalidParameterValues(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	rm := radio.NewManager()
	orch := command.NewOrchestrator(hub, cfg)
	server := NewServer(hub, orch, rm, 30*time.Second, 30*time.Second, 120*time.Second)

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Invalid power value (too high)",
			method:         "POST",
			path:           "/api/v1/radios/radio-01/power",
			body:           `{"powerDbm": 100.0}`, // Too high
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_RANGE",
		},
		{
			name:           "Invalid power value (negative)",
			method:         "POST",
			path:           "/api/v1/radios/radio-01/power",
			body:           `{"powerDbm": -50.0}`, // Negative
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_RANGE",
		},
		{
			name:           "Invalid frequency value (too high)",
			method:         "POST",
			path:           "/api/v1/radios/radio-01/channel",
			body:           `{"frequencyMhz": 100000.0}`, // Too high
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_RANGE",
		},
		{
			name:           "Invalid frequency value (negative)",
			method:         "POST",
			path:           "/api/v1/radios/radio-01/channel",
			body:           `{"frequencyMhz": -1000.0}`, // Negative
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_RANGE",
		},
		{
			name:           "Missing required field (powerDbm)",
			method:         "POST",
			path:           "/api/v1/radios/radio-01/power",
			body:           `{"frequencyMhz": 2412.0}`, // Wrong field
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Missing required field (frequencyMhz)",
			method:         "POST",
			path:           "/api/v1/radios/radio-01/channel",
			body:           `{"powerDbm": 30.0}`, // Wrong field
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()

			// Use appropriate handler based on path
			if strings.Contains(tt.path, "/power") {
				server.handleSetPower(w, req, "non-existent-radio")
			} else if strings.Contains(tt.path, "/channel") {
				server.handleSetChannel(w, req, "non-existent-radio")
			} else if strings.Contains(tt.path, "/state") {
				server.handleRadioByID(w, req)
			} else if strings.Contains(tt.path, "/select") {
				server.handleSelectRadio(w, req)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify error response format
			var response Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Result != "error" {
				t.Errorf("Expected result 'error', got '%s'", response.Result)
			}

			if response.Code != tt.expectedCode {
				t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, response.Code)
			}
		})
	}
}

// TestUnsupportedMethods tests API with unsupported HTTP methods
func TestUnsupportedMethods(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	rm := radio.NewManager()
	orch := command.NewOrchestrator(hub, cfg)
	server := NewServer(hub, orch, rm, 30*time.Second, 30*time.Second, 120*time.Second)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "PUT on SetPower endpoint",
			method:         "PUT",
			path:           "/api/v1/radios/radio-01/power",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE on SetChannel endpoint",
			method:         "DELETE",
			path:           "/api/v1/radios/radio-01/channel",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PATCH on GetState endpoint",
			method:         "PATCH",
			path:           "/api/v1/radios/radio-01/state",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()

			// Use appropriate handler based on path
			if strings.Contains(tt.path, "/power") {
				server.handleSetPower(w, req, "non-existent-radio")
			} else if strings.Contains(tt.path, "/channel") {
				server.handleSetChannel(w, req, "non-existent-radio")
			} else if strings.Contains(tt.path, "/state") {
				server.handleRadioByID(w, req)
			} else if strings.Contains(tt.path, "/select") {
				server.handleSelectRadio(w, req)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestInvalidContentType tests API with invalid content types
func TestInvalidContentType(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	rm := radio.NewManager()
	orch := command.NewOrchestrator(hub, cfg)
	server := NewServer(hub, orch, rm, 30*time.Second, 30*time.Second, 120*time.Second)

	tests := []struct {
		name           string
		contentType    string
		body           string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Text/plain content type",
			contentType:    "text/plain",
			body:           `{"powerDbm": 30.0}`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "XML content type",
			contentType:    "application/xml",
			body:           `{"powerDbm": 30.0}`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Form data content type",
			contentType:    "application/x-www-form-urlencoded",
			body:           `powerDbm=30.0`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "No content type",
			contentType:    "",
			body:           `{"powerDbm": 30.0}`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/radios/radio-01/power", strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()

			server.handleSetPower(w, req, "radio-01")

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify error response format
			var response Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Result != "error" {
				t.Errorf("Expected result 'error', got '%s'", response.Result)
			}

			if response.Code != tt.expectedCode {
				t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, response.Code)
			}
		})
	}
}

// TestLargeRequestBodies tests API with large request bodies
func TestLargeRequestBodies(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	rm := radio.NewManager()
	orch := command.NewOrchestrator(hub, cfg)
	server := NewServer(hub, orch, rm, 30*time.Second, 30*time.Second, 120*time.Second)

	// Create a large JSON payload
	largeData := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		largeData[fmt.Sprintf("field%d", i)] = strings.Repeat("x", 1000)
	}
	largeJSON, _ := json.Marshal(largeData)

	req := httptest.NewRequest("POST", "/api/v1/radios/radio-01/power", bytes.NewReader(largeJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	server.handleSetPower(w, req, "radio-01")

	// Should handle large requests gracefully
	if w.Code == http.StatusRequestEntityTooLarge {
		t.Log("Server correctly rejected large request")
	} else if w.Code == http.StatusBadRequest {
		t.Log("Server correctly rejected malformed large request")
	} else {
		t.Logf("Server handled large request with status %d", w.Code)
	}
}

// TestConcurrentRequests tests API with concurrent requests
func TestConcurrentRequests(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	rm := radio.NewManager()
	orch := command.NewOrchestrator(hub, cfg)
	server := NewServer(hub, orch, rm, 30*time.Second, 30*time.Second, 120*time.Second)

	// Test concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			req := httptest.NewRequest("GET", "/api/v1/radios", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()

			server.handleRadios(w, req)

			// Should handle concurrent requests gracefully
			if w.Code != http.StatusOK && w.Code != http.StatusUnauthorized {
				t.Errorf("Unexpected status code %d for concurrent request %d", w.Code, index)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestTimeoutRequests tests API with timeout scenarios
func TestTimeoutRequests(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	rm := radio.NewManager()
	orch := command.NewOrchestrator(hub, cfg)
	server := NewServer(hub, orch, rm, 1*time.Millisecond, 1*time.Millisecond, 1*time.Millisecond) // Very short timeouts

	req := httptest.NewRequest("GET", "/api/v1/radios", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	server.handleSetPower(w, req, "radio-01")

	// Should handle timeout gracefully
	if w.Code == http.StatusRequestTimeout || w.Code == http.StatusServiceUnavailable {
		t.Log("Server correctly handled timeout")
	} else {
		t.Logf("Server handled timeout with status %d", w.Code)
	}
}
