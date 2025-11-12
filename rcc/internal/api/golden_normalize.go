package api

import (
	"encoding/json"
	"testing"
)

// normalizeResponse normalizes API responses for stable golden test comparison
// by removing or fixing dynamic fields like correlationId, timestamp, eventId
func normalizeResponse(t *testing.T, raw []byte) map[string]any {
	var response map[string]any
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatalf("Failed to unmarshal response for normalization: %v", err)
	}

	// Remove or normalize dynamic fields
	delete(response, "correlationId")
	delete(response, "timestamp")
	delete(response, "eventId")
	delete(response, "lastSeen")

	// If there's a data field with nested dynamic content, normalize it too
	if data, ok := response["data"].(map[string]any); ok {
		normalizeDataField(data)
	}

	return response
}

// normalizeDataField normalizes nested data fields that may contain dynamic content
func normalizeDataField(data map[string]any) {
	// Remove common dynamic fields from data
	delete(data, "lastSeen")
	delete(data, "timestamp")
	delete(data, "eventId")

	// If there's a state field, normalize it
	if state, ok := data["state"].(map[string]any); ok {
		delete(state, "lastSeen")
		delete(state, "timestamp")
	}

	// If there's a capabilities field, normalize it
	if capabilities, ok := data["capabilities"].(map[string]any); ok {
		delete(capabilities, "lastSeen")
		delete(capabilities, "timestamp")
	}
}

// assertCorrelationIdPresent asserts that correlationId exists and is non-empty
// without comparing the actual value (for tests that need to verify presence)
func assertCorrelationIdPresent(t *testing.T, response map[string]any, testName string) {
	if correlationId, exists := response["correlationId"]; !exists {
		t.Errorf("Expected correlationId to be present in response for %s", testName)
	} else if correlationIdStr, ok := correlationId.(string); !ok || correlationIdStr == "" {
		t.Errorf("Expected correlationId to be non-empty string for %s, got: %v", testName, correlationId)
	}
}
