// Package e2e provides negative test cases for the Radio Control Container API.
// This file implements black-box testing using only HTTP and contract validation.
package e2e

import (
	"testing"

	"github.com/radio-control/rcc/test/harness"
)

func TestE2E_ErrorHandling(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	// Create test harness with seeded state
	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Evidence: Seeded state via HTTP contract
	t.Logf("=== TEST EVIDENCE ===")
	radios := httpGetJSON(t, server.URL+"/api/v1/radios")
	mustHave(t, radios, "result", "ok")
	if d, ok := radios["data"].(map[string]any); ok {
		if id, ok := d["activeRadioId"].(string); ok {
			t.Logf("Active Radio ID: %s", id)
		}
	}
	t.Logf("===================")

	// Test invalid radio ID
	resp := httpGetWithStatus(t, server.URL+"/api/v1/radios/invalid-radio-id")
	validator.ValidateErrorResponse(t, resp, "NOT_FOUND")

	// Test power out of range
	payload := `{"powerDbm": 100}`
	resp = httpPostWithStatus(t, server.URL+"/api/v1/radios/silvus-001/power", payload)
	validator.ValidateErrorResponse(t, resp, "INVALID_RANGE")

	// Test channel out of range
	payload = `{"frequencyMhz": 10000.0}`
	resp = httpPostWithStatus(t, server.URL+"/api/v1/radios/silvus-001/channel", payload)
	validator.ValidateErrorResponse(t, resp, "INVALID_RANGE")

	// Audit logs are server-side only per Architecture §8.6; no E2E access

	t.Log("✅ Error handling working correctly")
}

func TestE2E_InvalidJSON(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Test malformed JSON
	payload := `{"powerDbm": invalid}`
	resp := httpPostWithStatus(t, server.URL+"/api/v1/radios/silvus-001/power", payload)
	validator.ValidateHTTPResponse(t, resp, 400)

	// Test out-of-range power value (valid JSON but invalid business logic)
	payload = `{"powerDbm": 100}`
	resp = httpPostWithStatus(t, server.URL+"/api/v1/radios/silvus-001/power", payload)
	validator.ValidateHTTPResponse(t, resp, 400)

	t.Log("✅ Invalid JSON handling working correctly")
}

func TestE2E_RadioNotFound(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Test operations on non-existent radio
	payload := `{"powerDbm": 10.0}`
	resp := httpPostWithStatus(t, server.URL+"/api/v1/radios/non-existent/power", payload)
	validator.ValidateErrorResponse(t, resp, "NOT_FOUND")

	// Test getting state of non-existent radio
	resp = httpGetWithStatus(t, server.URL+"/api/v1/radios/non-existent/power")
	validator.ValidateErrorResponse(t, resp, "NOT_FOUND")

	t.Log("✅ Radio not found handling working correctly")
}

func TestE2E_AdapterBusy(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// BUSY injection is not available via black-box API; skip until test-only fault profile exists
	t.Skipf("BUSY fault requires black-box trigger; see DRIFT-LEDGER")

	// Test with valid power to trigger busy response
	payload := `{"powerDbm": 10.0}`
	resp := httpPostWithStatus(t, server.URL+"/api/v1/radios/silvus-001/power", payload)
	validator.ValidateErrorResponse(t, resp, "BUSY")

	// Audit logs are server-side only per Architecture §8.6; no E2E access

	t.Log("✅ Adapter busy handling working correctly")
}
