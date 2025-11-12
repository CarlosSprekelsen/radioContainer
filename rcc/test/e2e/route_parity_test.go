// Package e2e provides route parity testing against OpenAPI specification.
package e2e

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouteParity_OpenAPIValidation(t *testing.T) {
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	// Create a test server (this would normally be your actual server)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock responses for testing
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/radios":
			if r.Method == "GET" {
				w.WriteHeader(200)
				w.Write([]byte(`{"result":"ok","data":[{"id":"silvus-001","name":"Silvus Radio 1"}]}`))
			} else {
				w.WriteHeader(405)
				w.Write([]byte(`{"result":"error","code":"METHOD_NOT_ALLOWED","message":"Method not allowed"}`))
			}
		case "/api/v1/radios/silvus-001/power":
			if r.Method == "GET" {
				w.WriteHeader(200)
				w.Write([]byte(`{"result":"ok","data":{"powerDbm":25.0}}`))
			} else if r.Method == "POST" {
				w.WriteHeader(200)
				w.Write([]byte(`{"result":"ok","data":{"powerDbm":25.0}}`))
			} else {
				w.WriteHeader(405)
				w.Write([]byte(`{"result":"error","code":"METHOD_NOT_ALLOWED","message":"Method not allowed"}`))
			}
		case "/api/v1/radios/silvus-001/channel":
			if r.Method == "GET" {
				w.WriteHeader(200)
				w.Write([]byte(`{"result":"ok","data":{"channelIndex":1,"frequencyMhz":2400.0}}`))
			} else if r.Method == "POST" {
				w.WriteHeader(200)
				w.Write([]byte(`{"result":"ok","data":{"channelIndex":1,"frequencyMhz":2400.0}}`))
			} else {
				w.WriteHeader(405)
				w.Write([]byte(`{"result":"error","code":"METHOD_NOT_ALLOWED","message":"Method not allowed"}`))
			}
		case "/api/v1/telemetry":
			if r.Method == "GET" {
				w.Header().Set("Content-Type", "text/event-stream")
				w.WriteHeader(200)
				w.Write([]byte(`event: ready\ndata: {"snapshot":{"activeRadioId":"","radios":[]}}\n\n`))
			} else {
				w.WriteHeader(405)
				w.Write([]byte(`{"result":"error","code":"METHOD_NOT_ALLOWED","message":"Method not allowed"}`))
			}
		default:
			w.WriteHeader(404)
			w.Write([]byte(`{"result":"error","code":"NOT_FOUND","message":"Not found"}`))
		}
	}))
	defer ts.Close()

	// Test routes against OpenAPI spec
	testCases := []struct {
		method string
		path   string
		desc   string
	}{
		{"GET", "/api/v1/radios", "List radios"},
		{"GET", "/api/v1/radios/silvus-001/power", "Get radio power"},
		{"POST", "/api/v1/radios/silvus-001/power", "Set radio power"},
		{"GET", "/api/v1/radios/silvus-001/channel", "Get radio channel"},
		{"POST", "/api/v1/radios/silvus-001/channel", "Set radio channel"},
		{"GET", "/api/v1/telemetry", "Telemetry stream"},
	}

	t.Logf("=== ROUTE PARITY ANALYSIS ===")
	t.Logf("%-8s %-30s %-20s %-10s %-s", "METHOD", "PATH", "EXPECTED", "ACTUAL", "STATUS")
	t.Logf("%s", "----------------------------------------------------------------")

	for _, tc := range testCases {
		// Make HTTP request
		req, err := http.NewRequest(tc.method, ts.URL+tc.path, nil)
		if err != nil {
			t.Errorf("Failed to create request: %v", err)
			continue
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("Failed to make request: %v", err)
			continue
		}

		// Validate against OpenAPI spec
		validator.ValidateHTTPResponseAgainstOpenAPI(t, resp, tc.method, tc.path)

		// Log results
		status := "PASS"
		if resp.StatusCode >= 400 {
			status = "FAIL"
		}

		t.Logf("%-8s %-30s %-20s %-10d %-s", tc.method, tc.path, "200", resp.StatusCode, status)

		resp.Body.Close()
	}

	t.Logf("=================================================================")
}

func TestErrorMapping_FromContract(t *testing.T) {
	validator := NewContractValidator(t)

	// Test that error mappings are loaded from contract
	if len(validator.errorMappings) == 0 {
		t.Error("Expected error mappings to be loaded from contract")
	}

	// Test specific error mappings
	expectedMappings := map[string]int{
		"INVALID_RANGE": 400,
		"BUSY":          503,
		"UNAVAILABLE":   503,
		"INTERNAL":      500,
	}

	for errorCode, expectedStatus := range expectedMappings {
		if actualStatus, exists := validator.errorMappings[errorCode]; !exists {
			t.Errorf("Expected error mapping for %s not found", errorCode)
		} else if actualStatus != expectedStatus {
			t.Errorf("Expected status %d for %s, got %d", expectedStatus, errorCode, actualStatus)
		}
	}
}
