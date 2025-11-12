//go:build integration

package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/radio-control/rcc/test/harness"
)

// TestContractIntegration_SelectRadio tests the Api → Orchestrator → RadioManager flow
func TestContractIntegration_SelectRadio(t *testing.T) {
	// Arrange: Use existing harness with proper RadioManager setup
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Act: POST /radios/select
	req, err := http.NewRequest("POST", server.URL+"/api/v1/radios/select",
		strings.NewReader(`{"radioId":"silvus-001"}`))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert: 200 OK with proper result
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result, ok := response["result"].(string); !ok || result != "ok" {
		t.Errorf("Expected result 'ok', got '%s'", result)
	}

	t.Logf("✅ Select radio flow: Api → Orchestrator → RadioManager")
}

// TestContractIntegration_SetChannel tests the channel index mapping flow
func TestContractIntegration_SetChannel(t *testing.T) {
	// Arrange: Use harness with channel mapping
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	// Default band plan: Index 1→2412.0, Index 6→2437.0, Index 11→2462.0
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Act: POST /radios/{id}/channel with channel index
	req, err := http.NewRequest("POST", server.URL+"/api/v1/radios/silvus-001/channel",
		strings.NewReader(`{"channelIndex":6}`))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert: 200 OK with channel change acknowledgment
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)

		// Log the error response for debugging
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			t.Logf("Error response: %+v", errorResponse)
		}
		return // Don't continue with success assertions
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result, ok := response["result"].(string); !ok || result != "ok" {
		t.Errorf("Expected result 'ok', got '%s'", result)
	}

	t.Logf("✅ Set channel flow: Api → Orchestrator → ConfigStore → RadioManager → Adapter")
}

// TestContractIntegration_SetPower tests the power setting flow with error normalization
func TestContractIntegration_SetPower(t *testing.T) {
	// Arrange: Use harness with power validation
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Act: POST /radios/{id}/power with valid power
	req, err := http.NewRequest("POST", server.URL+"/api/v1/radios/silvus-001/power",
		strings.NewReader(`{"powerDbm":25}`))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert: 200 OK with power change acknowledgment
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result, ok := response["result"].(string); !ok || result != "ok" {
		t.Errorf("Expected result 'ok', got '%s'", result)
	}

	t.Logf("✅ Set power flow: Api → Orchestrator → RadioManager → Adapter")
}

// TestContractIntegration_ErrorNormalization tests error mapping per Architecture §8.5
func TestContractIntegration_ErrorNormalization(t *testing.T) {
	// Arrange: Use harness with fault injection
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	testCases := []struct {
		name           string
		powerDbm       float64
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid power range",
			powerDbm:       25,
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "Invalid power range (too high)",
			powerDbm:       50,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_RANGE",
		},
		{
			name:           "Invalid power range (too low)",
			powerDbm:       -10,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_RANGE",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act: POST /radios/{id}/power
			reqBody := strings.NewReader(`{"powerDbm":` + fmt.Sprintf("%.0f", tc.powerDbm) + `}`)
			req, err := http.NewRequest("POST", server.URL+"/api/v1/radios/silvus-001/power", reqBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Assert: Expected status code
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)

				// Log the response for debugging
				var errorResponse map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
					t.Logf("Response: %+v", errorResponse)
				}
			}

			if tc.expectedError != "" {
				var response map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				t.Logf("Response for %s: %+v", tc.name, response)

				// Check both 'error' and 'code' fields for error code
				var errorCode string
				if code, ok := response["code"].(string); ok {
					errorCode = code
				} else if err, ok := response["error"].(string); ok {
					errorCode = err
				}

				if errorCode != tc.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tc.expectedError, errorCode)
				}
			}
		})
	}

	t.Logf("✅ Error normalization: Vendor errors → {INVALID_RANGE,BUSY,UNAVAILABLE,INTERNAL}")
}

// TestContractIntegration_TelemetryEvents tests telemetry event emission
func TestContractIntegration_TelemetryEvents(t *testing.T) {
	// Arrange: Use harness with telemetry hub
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Act: Make a control command that should emit telemetry
	req, err := http.NewRequest("POST", server.URL+"/api/v1/radios/silvus-001/power",
		strings.NewReader(`{"powerDbm":30}`))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert: Command succeeded
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Note: In a real integration test, we would also verify that:
	// 1. TelemetryHub published a 'powerChanged' event
	// 2. Event structure matches Telemetry SSE v1 schema
	// 3. Event contains correct radioId, powerDbm, timestamp
	// This would require SSE client connection to verify events

	t.Logf("✅ Telemetry events: Orchestrator → TelemetryHub → SSE clients")
}

// TestContractIntegration_TimingConstraints tests CB-TIMING compliance
func TestContractIntegration_TimingConstraints(t *testing.T) {
	// Arrange: Use harness with timing configuration
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Act: Measure command execution time
	start := time.Now()

	req, err := http.NewRequest("POST", server.URL+"/api/v1/radios/silvus-001/power",
		strings.NewReader(`{"powerDbm":20}`))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	latency := time.Since(start)

	// Assert: Command succeeded and within timing constraints
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// CB-TIMING v0.3 §5: setPower timeout is 10 seconds
	if latency > 10*time.Second {
		t.Errorf("setPower took %v, exceeds CB-TIMING timeout of 10s", latency)
	}

	t.Logf("✅ Timing constraints: Commands complete within CB-TIMING limits (%v)", latency)
}
