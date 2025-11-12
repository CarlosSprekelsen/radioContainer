package harness

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestFixedHandleSelectRadio tests the select radio endpoint with proper setup
func TestFixedHandleSelectRadio(t *testing.T) {
	opts := DefaultOptions()
	opts.ActiveRadioID = "silvus-001" // Use the radio that exists in the harness
	server := NewServer(t, opts)
	defer server.Shutdown()

	// Test POST /radios/select with existing radio
	req := httptest.NewRequest("POST", server.URL+"/api/v1/radios/select", strings.NewReader(`{"radioId":"silvus-001"}`))
	req.Header.Set("Content-Type", "application/json")

	// Make the actual HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// The test should now pass because we have a real radio loaded
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result, ok := response["result"].(string); ok {
		if result != "ok" {
			t.Errorf("Expected result 'ok', got '%s'", result)
		}
	}
}

// TestFixedHandleRadioByID tests the radio by ID endpoint with proper setup
func TestFixedHandleRadioByID(t *testing.T) {
	opts := DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	server := NewServer(t, opts)
	defer server.Shutdown()

	// Test GET /radios/{id} with existing radio
	req := httptest.NewRequest("GET", server.URL+"/api/v1/radios/silvus-001", nil)

	// Make the actual HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should now return 200 instead of 404
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result, ok := response["result"].(string); ok {
		if result != "ok" {
			t.Errorf("Expected result 'ok', got '%s'", result)
		}
	}
}

// TestFixedHandleGetPower tests the get power endpoint with proper setup
func TestFixedHandleGetPower(t *testing.T) {
	opts := DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	server := NewServer(t, opts)
	defer server.Shutdown()

	// Test GET /radios/{id}/power with existing radio
	req := httptest.NewRequest("GET", server.URL+"/api/v1/radios/silvus-001/power", nil)

	// Make the actual HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should now return 200 instead of 500
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result, ok := response["result"].(string); ok {
		if result != "ok" {
			t.Errorf("Expected result 'ok', got '%s'", result)
		}
	}
}

// TestFixedHandleSetPower tests the set power endpoint with proper setup
func TestFixedHandleSetPower(t *testing.T) {
	opts := DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	server := NewServer(t, opts)
	defer server.Shutdown()

	// Test POST /radios/{id}/power with existing radio
	req := httptest.NewRequest("POST", server.URL+"/api/v1/radios/silvus-001/power", strings.NewReader(`{"powerDbm": 25}`))
	req.Header.Set("Content-Type", "application/json")

	// Make the actual HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should now return 200 instead of 500
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result, ok := response["result"].(string); ok {
		if result != "ok" {
			t.Errorf("Expected result 'ok', got '%s'", result)
		}
	}
}

// TestFixedHandleGetChannel tests the get channel endpoint with proper setup
func TestFixedHandleGetChannel(t *testing.T) {
	opts := DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	server := NewServer(t, opts)
	defer server.Shutdown()

	// Test GET /radios/{id}/channel with existing radio
	req := httptest.NewRequest("GET", server.URL+"/api/v1/radios/silvus-001/channel", nil)

	// Make the actual HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should now return 200 instead of 500
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result, ok := response["result"].(string); ok {
		if result != "ok" {
			t.Errorf("Expected result 'ok', got '%s'", result)
		}
	}
}

// TestFixedHandleSetChannel tests the set channel endpoint with proper setup
func TestFixedHandleSetChannel(t *testing.T) {
	opts := DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	server := NewServer(t, opts)
	defer server.Shutdown()

	// Test POST /radios/{id}/channel with existing radio and valid frequency
	req := httptest.NewRequest("POST", server.URL+"/api/v1/radios/silvus-001/channel", strings.NewReader(`{"frequencyMhz": 2437.0}`))
	req.Header.Set("Content-Type", "application/json")

	// Make the actual HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should now return 200 instead of 500
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result, ok := response["result"].(string); ok {
		if result != "ok" {
			t.Errorf("Expected result 'ok', got '%s'", result)
		}
	}
}
