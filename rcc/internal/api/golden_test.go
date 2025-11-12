package api

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
)

var update = flag.Bool("update", false, "update golden files")

// TestAPIEndpointsGolden tests API GET endpoints against golden files
func TestAPIEndpointsGolden(t *testing.T) {
	// Set up test server with stable configuration
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	rm := radio.NewManager()
	orch := command.NewOrchestrator(hub, cfg)
	server := NewServer(hub, orch, rm, 30*time.Second, 30*time.Second, 120*time.Second)

	// Set fixed start time for consistent uptime
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	server.startTime = fixedTime

	tests := []struct {
		name       string
		method     string
		path       string
		headers    map[string]string
		goldenFile string
	}{
		{
			name:       "health",
			method:     "GET",
			path:       "/api/v1/health",
			goldenFile: "health.json",
		},
		{
			name:       "capabilities",
			method:     "GET",
			path:       "/api/v1/capabilities",
			goldenFile: "capabilities.json",
		},
		{
			name:       "radios",
			method:     "GET",
			path:       "/api/v1/radios",
			goldenFile: "radios.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Route request to appropriate handler
			server.routeRequest(w, req)

			// Get response body
			body := w.Body.Bytes()

			// Normalize response for stable comparison
			normalized := normalizeAPIResponse(body, tt.name)

			goldenPath := filepath.Join("testdata", "api", tt.goldenFile)

			if *update {
				// Update golden file
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("Failed to create testdata directory: %v", err)
				}
				if err := ioutil.WriteFile(goldenPath, normalized, 0644); err != nil {
					t.Fatalf("Failed to write golden file: %v", err)
				}
				t.Logf("Updated golden file: %s", goldenPath)
				return
			}

			// Read golden file
			golden, err := ioutil.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("Failed to read golden file %s: %v", goldenPath, err)
			}

			// Compare responses
			if string(normalized) != string(golden) {
				t.Errorf("Response doesn't match golden file %s\nExpected:\n%s\nGot:\n%s",
					goldenPath, string(golden), string(normalized))
			}
		})
	}
}

// routeRequest routes a request to the appropriate handler
func (s *Server) routeRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/api/v1/health":
		s.handleHealth(w, r)
	case path == "/api/v1/capabilities":
		s.handleCapabilities(w, r)
	case path == "/api/v1/radios":
		s.handleRadios(w, r)
	case strings.HasPrefix(path, "/api/v1/radios/") && !strings.Contains(path, "/power") && !strings.Contains(path, "/channel"):
		s.handleRadioByID(w, r)
	case strings.HasSuffix(path, "/power"):
		s.handleGetPower(w, r, s.extractRadioID(path))
	case strings.HasSuffix(path, "/channel"):
		s.handleGetChannel(w, r, s.extractRadioID(path))
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Endpoint not found", nil)
	}
}

// normalizeAPIResponse normalizes API responses for stable comparison
func normalizeAPIResponse(body []byte, testName string) []byte {
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return body // Return as-is if not JSON
	}

	// Normalize correlation ID to a fixed value for stable comparison
	response["correlationId"] = "test-correlation-id-12345"

	// Normalize based on test type
	switch testName {
	case "health":
		// Normalize uptime to a fixed value for stable comparison
		if data, ok := response["data"].(map[string]interface{}); ok {
			data["uptimeSec"] = 12345.0 // Fixed uptime for stable comparison
		}
	case "radios", "radio_by_id":
		// Ensure consistent ordering of radio fields
		if data, ok := response["data"].(map[string]interface{}); ok {
			normalizeRadioData(data)
		} else if radios, ok := response["data"].([]interface{}); ok {
			for _, radio := range radios {
				if radioMap, ok := radio.(map[string]interface{}); ok {
					normalizeRadioData(radioMap)
				}
			}
		}
	}

	// Re-marshal with consistent formatting
	normalized, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return body
	}
	return normalized
}

// normalizeRadioData normalizes radio data for stable comparison
func normalizeRadioData(data map[string]interface{}) {
	// Ensure consistent field ordering and types
	if power, ok := data["powerDbm"].(float64); ok {
		data["powerDbm"] = int(power)
	}
	if freq, ok := data["frequencyMhz"].(float64); ok {
		// Round to 1 decimal place for stability
		data["frequencyMhz"] = float64(int(freq*10)) / 10
	}
}

// TestAPIErrorResponsesGolden tests API error responses against golden files
func TestAPIErrorResponsesGolden(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	rm := radio.NewManager()
	orch := command.NewOrchestrator(hub, cfg)
	server := NewServer(hub, orch, rm, 30*time.Second, 30*time.Second, 120*time.Second)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		goldenFile string
	}{
		{
			name:       "invalid_radio_id",
			method:     "GET",
			path:       "/api/v1/radios/nonexistent",
			goldenFile: "error_invalid_radio_id.json",
		},
		{
			name:       "invalid_power_range",
			method:     "POST",
			path:       "/api/v1/radios/radio-01/power",
			body:       `{"powerDbm": 50}`,
			goldenFile: "error_invalid_power_range.json",
		},
		{
			name:       "invalid_channel_range",
			method:     "POST",
			path:       "/api/v1/radios/radio-01/channel",
			body:       `{"frequencyMhz": 100000}`,
			goldenFile: "error_invalid_channel_range.json",
		},
		{
			name:       "missing_radio_id",
			method:     "GET",
			path:       "/api/v1/radios/",
			goldenFile: "error_missing_radio_id.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Route request to appropriate handler
			server.routeRequest(w, req)

			// Get response body
			body := w.Body.Bytes()

			// Normalize response for stable comparison
			normalized := normalizeResponse(t, body)
			normalizedJSON, err := json.MarshalIndent(normalized, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal normalized response: %v", err)
			}

			goldenPath := filepath.Join("testdata", "api", tt.goldenFile)

			if *update {
				// Update golden file with normalized response
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("Failed to create testdata directory: %v", err)
				}
				if err := ioutil.WriteFile(goldenPath, normalizedJSON, 0644); err != nil {
					t.Fatalf("Failed to write golden file: %v", err)
				}
				t.Logf("Updated golden file: %s", goldenPath)
				return
			}

			// Read golden file
			golden, err := ioutil.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("Failed to read golden file %s: %v", goldenPath, err)
			}

			// Compare normalized responses
			if string(normalizedJSON) != string(golden) {
				t.Errorf("Response doesn't match golden file %s\nExpected:\n%s\nGot:\n%s",
					goldenPath, string(golden), string(normalizedJSON))
			}

			// For tests that need to verify correlationId presence, do so on original response
			var originalResponse map[string]any
			if err := json.Unmarshal(body, &originalResponse); err == nil {
				assertCorrelationIdPresent(t, originalResponse, tt.name)
			}
		})
	}
}
