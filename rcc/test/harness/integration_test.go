package harness

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// TestIntegration_SelectRadio tests the select radio endpoint with proper setup
func TestIntegration_SelectRadio(t *testing.T) {
	opts := DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	server := NewServer(t, opts)
	defer server.Shutdown()

	// Test POST /radios/select with existing radio
	req, err := http.NewRequest("POST", server.URL+"/api/v1/radios/select", strings.NewReader(`{"radioId":"silvus-001"}`))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the actual HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("Response status: %d", resp.StatusCode)
	t.Logf("Response headers: %v", resp.Header)

	// The test should now pass because we have a real radio loaded
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	t.Logf("Response body: %+v", response)

	if result, ok := response["result"].(string); ok {
		if result != "ok" {
			t.Errorf("Expected result 'ok', got '%s'", result)
		}
	}
}

// TestIntegration_RadioByID tests the radio by ID endpoint with proper setup
func TestIntegration_RadioByID(t *testing.T) {
	opts := DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	server := NewServer(t, opts)
	defer server.Shutdown()

	// Test GET /radios/{id} with existing radio
	req, err := http.NewRequest("GET", server.URL+"/api/v1/radios/silvus-001", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Make the actual HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("Response status: %d", resp.StatusCode)

	// Should now return 200 instead of 404
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	t.Logf("Response body: %+v", response)

	if result, ok := response["result"].(string); ok {
		if result != "ok" {
			t.Errorf("Expected result 'ok', got '%s'", result)
		}
	}
}
