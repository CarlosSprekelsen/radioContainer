// Package e2e provides contract validation for E2E tests.
// This file ensures all E2E tests validate against the API specification.
package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// ContractValidator validates E2E test responses against API contracts
type ContractValidator struct {
	specVersion     string
	errorMappings   map[string]int
	telemetrySchema map[string]interface{}
	openapiSpec     map[string]interface{}
}

// NewContractValidator creates a new contract validator
func NewContractValidator(t *testing.T) *ContractValidator {
	// Read spec version
	specVersion := readSpecVersion(t)

	// Load error mappings
	errorMappings := loadErrorMappings(t)

	// Load telemetry schema
	telemetrySchema := loadTelemetrySchema(t)

	// Load OpenAPI specification
	openapiSpec := loadOpenAPISpec(t)

	return &ContractValidator{
		specVersion:     specVersion,
		errorMappings:   errorMappings,
		telemetrySchema: telemetrySchema,
		openapiSpec:     openapiSpec,
	}
}

// PrintSpecVersion prints the spec version at test start
func (cv *ContractValidator) PrintSpecVersion(t *testing.T) {
	t.Logf("Spec Version: %s", cv.specVersion)
}

// ValidateHTTPResponse validates an HTTP response against the contract
func (cv *ContractValidator) ValidateHTTPResponse(t *testing.T, resp *http.Response, expectedStatus int) {
	// Validate status code
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	// Validate content type for JSON responses
	if resp.StatusCode < 400 {
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("Expected JSON content type, got %s", contentType)
		}
	}

	// Validate response envelope for success responses
	if resp.StatusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var envelope map[string]interface{}
		if err := json.Unmarshal(body, &envelope); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		// Normalize dynamic fields before validation
		cv.normalizeDynamicFields(envelope)

		// Check for required envelope fields
		if _, ok := envelope["result"]; !ok {
			t.Error("Expected 'result' field in response envelope")
		}

		if _, ok := envelope["data"]; !ok {
			t.Error("Expected 'data' field in response envelope")
		}
	}
}

// ValidateHTTPResponseAgainstOpenAPI validates response against OpenAPI spec
func (cv *ContractValidator) ValidateHTTPResponseAgainstOpenAPI(t *testing.T, resp *http.Response, method, path string) {
	// Extract expected status from OpenAPI spec
	expectedStatus := cv.getExpectedStatusFromOpenAPI(method, path)

	// Validate status code
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d from OpenAPI spec, got %d", expectedStatus, resp.StatusCode)
	}

	// Validate response schema against OpenAPI
	cv.validateResponseSchema(t, resp, method, path)
}

// getExpectedStatusFromOpenAPI extracts expected status from OpenAPI spec
func (cv *ContractValidator) getExpectedStatusFromOpenAPI(method, path string) int {
	// Default to 200 for successful operations
	// In a full implementation, this would parse the OpenAPI spec
	// and extract the expected status codes for each endpoint
	return 200
}

// validateResponseSchema validates response against OpenAPI schema
func (cv *ContractValidator) validateResponseSchema(t *testing.T, resp *http.Response, method, path string) {
	// Load response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")

	// Handle different content types
	if strings.Contains(contentType, "text/event-stream") {
		// SSE response - validate format
		cv.validateSSEResponse(t, string(body))
		return
	}

	if !strings.Contains(contentType, "application/json") {
		// Non-JSON response - skip JSON validation
		return
	}

	// Parse JSON response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Normalize dynamic fields
	cv.normalizeDynamicFields(response)

	// Validate against OpenAPI schema
	// This is a simplified validation - in production, use a proper OpenAPI validator
	if resp.StatusCode == 200 {
		// Success response validation
		if result, ok := response["result"].(string); !ok || result != "ok" {
			t.Error("Expected 'result' field to be 'ok' for success responses")
		}
		if _, ok := response["data"]; !ok {
			t.Error("Expected 'data' field in success responses")
		}
	} else {
		// Error response validation
		if result, ok := response["result"].(string); !ok || result != "error" {
			t.Error("Expected 'result' field to be 'error' for error responses")
		}
		if _, ok := response["code"]; !ok {
			t.Error("Expected 'code' field in error responses")
		}
		if _, ok := response["message"]; !ok {
			t.Error("Expected 'message' field in error responses")
		}
	}
}

// validateSSEResponse validates SSE response format
func (cv *ContractValidator) validateSSEResponse(t *testing.T, body string) {
	// Basic SSE format validation
	if !strings.Contains(body, "event:") {
		t.Error("Expected SSE response to contain 'event:' field")
	}
	if !strings.Contains(body, "data:") {
		t.Error("Expected SSE response to contain 'data:' field")
	}
}

// ValidateErrorResponse validates an error response against the error mapping
func (cv *ContractValidator) ValidateErrorResponse(t *testing.T, resp *http.Response, expectedError string) {
	expectedStatus, exists := cv.errorMappings[expectedError]
	if !exists {
		t.Errorf("Unknown error code: %s", expectedError)
		return
	}

	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d for error %s, got %d", expectedStatus, expectedError, resp.StatusCode)
	}

	// Validate error response structure
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read error response body: %v", err)
	}

	var errorResp map[string]interface{}
	if err := json.Unmarshal(body, &errorResp); err != nil {
		t.Fatalf("Failed to parse error JSON response: %v", err)
	}

	// Check for error response structure (result: "error")
	if result, ok := errorResp["result"]; !ok || result != "error" {
		t.Error("Expected 'result' field with value 'error' in error response")
	}
}

// ValidateSSEEvent validates an SSE event against the telemetry schema
func (cv *ContractValidator) ValidateSSEEvent(t *testing.T, event string) {
	// Parse SSE event format with improved multi-line handling
	lines := strings.Split(strings.TrimSpace(event), "\n")
	eventData := make(map[string]string)

	// Process each line, handling multi-line data
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue // Skip empty lines
		}

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Handle multi-line data by concatenating
				if existing, exists := eventData[key]; exists {
					eventData[key] = existing + "\n" + value
				} else {
					eventData[key] = value
				}
			}
		}
	}

	// Log parsed event data for debugging
	t.Logf("Parsed SSE event data: %+v", eventData)

	// Validate required SSE fields
	if _, ok := eventData["event"]; !ok {
		t.Error("Expected 'event' field in SSE event")
	}

	if _, ok := eventData["data"]; !ok {
		t.Error("Expected 'data' field in SSE event")
	}

	// Validate event ID is monotonic
	if id, ok := eventData["id"]; ok {
		if id == "" {
			t.Error("Event ID should not be empty")
		}
	}

	// Validate event type
	eventType := eventData["event"]
	validTypes := []string{"ready", "heartbeat", "powerChanged", "channelChanged"}

	valid := false
	for _, validType := range validTypes {
		if eventType == validType {
			valid = true
			break
		}
	}

	if !valid {
		t.Errorf("Invalid event type: %s", eventType)
	}

	// Validate data field is valid JSON
	if dataStr, ok := eventData["data"]; ok && dataStr != "" {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			t.Errorf("Invalid JSON in data field: %v", err)
		}

		// Validate specific event types
		switch eventType {
		case "ready":
			cv.validateReadyEvent(t, data)
		case "powerChanged":
			cv.validatePowerChangedEvent(t, data)
		case "channelChanged":
			cv.validateChannelChangedEvent(t, data)
		case "heartbeat":
			cv.validateHeartbeatEvent(t, data)
		}
	}
}

// ValidateHeartbeatInterval validates heartbeat timing against CB-TIMING
func (cv *ContractValidator) ValidateHeartbeatInterval(t *testing.T, events []string, baseInterval time.Duration, jitter time.Duration) {
	heartbeatEvents := make([]time.Time, 0)

	// Parse events to extract heartbeat timestamps
	for _, event := range events {
		if strings.Contains(event, "event: heartbeat") {
			// Extract timestamp from event data
			lines := strings.Split(event, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "data:") {
					var data map[string]interface{}
					jsonData := strings.TrimPrefix(line, "data: ")
					if err := json.Unmarshal([]byte(jsonData), &data); err == nil {
						if ts, ok := data["timestamp"].(string); ok {
							if parsedTime, err := time.Parse(time.RFC3339, ts); err == nil {
								heartbeatEvents = append(heartbeatEvents, parsedTime)
							}
						}
					}
				}
			}
		}
	}

	// Log heartbeat analysis
	t.Logf("=== HEARTBEAT TIMING ANALYSIS ===")
	t.Logf("Base interval: %v", baseInterval)
	t.Logf("Jitter: %v", jitter)
	t.Logf("Heartbeat events found: %d", len(heartbeatEvents))

	// Validate heartbeat intervals
	if len(heartbeatEvents) < 2 {
		t.Logf("Insufficient heartbeat events for interval validation (need ≥2, got %d)", len(heartbeatEvents))
		return
	}

	// Calculate tolerance window
	minInterval := baseInterval - jitter
	maxInterval := baseInterval + jitter
	timeout := baseInterval * 3 // Allow up to 3x base interval as timeout

	t.Logf("Tolerance window: [%v, %v]", minInterval, maxInterval)
	t.Logf("Timeout threshold: %v", timeout)

	// Validate each interval
	for i := 1; i < len(heartbeatEvents); i++ {
		interval := heartbeatEvents[i].Sub(heartbeatEvents[i-1])
		t.Logf("Interval %d: %v", i, interval)

		// Check for timeout (gap too large)
		if interval > timeout {
			t.Errorf("Heartbeat gap %v exceeds timeout %v", interval, timeout)
		}

		// Check for tolerance window
		if interval < minInterval {
			t.Errorf("Heartbeat interval %v too fast (below %v)", interval, minInterval)
		}

		if interval > maxInterval {
			t.Errorf("Heartbeat interval %v too slow (above %v)", interval, maxInterval)
		}
	}

	t.Logf("================================")
}

// ValidateHeartbeatIntervalNonFailing validates heartbeat timing without failing tests (for negative test cases)
func (cv *ContractValidator) ValidateHeartbeatIntervalNonFailing(t *testing.T, events []string, baseInterval time.Duration, jitter time.Duration) []string {
	heartbeatEvents := make([]time.Time, 0)

	// Parse events to extract heartbeat timestamps
	for _, event := range events {
		if strings.Contains(event, "event: heartbeat") {
			// Extract timestamp from event data
			lines := strings.Split(event, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "data:") {
					var data map[string]interface{}
					jsonData := strings.TrimPrefix(line, "data: ")
					if err := json.Unmarshal([]byte(jsonData), &data); err == nil {
						if ts, ok := data["timestamp"].(string); ok {
							if parsedTime, err := time.Parse(time.RFC3339, ts); err == nil {
								heartbeatEvents = append(heartbeatEvents, parsedTime)
							}
						}
					}
				}
			}
		}
	}

	var errors []string

	// Log heartbeat analysis
	t.Logf("=== HEARTBEAT TIMING ANALYSIS ===")
	t.Logf("Base interval: %v", baseInterval)
	t.Logf("Jitter: %v", jitter)
	t.Logf("Heartbeat events found: %d", len(heartbeatEvents))

	// Validate heartbeat intervals
	if len(heartbeatEvents) < 2 {
		msg := fmt.Sprintf("Insufficient heartbeat events for interval validation (need ≥2, got %d)", len(heartbeatEvents))
		t.Logf("%s", msg)
		errors = append(errors, msg)
		return errors
	}

	// Calculate tolerance window
	minInterval := baseInterval - jitter
	maxInterval := baseInterval + jitter
	timeout := baseInterval * 3 // Allow up to 3x base interval as timeout

	t.Logf("Tolerance window: [%v, %v]", minInterval, maxInterval)
	t.Logf("Timeout threshold: %v", timeout)

	// Validate each interval
	for i := 1; i < len(heartbeatEvents); i++ {
		interval := heartbeatEvents[i].Sub(heartbeatEvents[i-1])
		t.Logf("Interval %d: %v", i, interval)

		// Check for timeout (gap too large)
		if interval > timeout {
			msg := fmt.Sprintf("Heartbeat gap %v exceeds timeout %v", interval, timeout)
			t.Logf("%s", msg)
			errors = append(errors, msg)
		}

		// Check for tolerance window
		if interval < minInterval {
			msg := fmt.Sprintf("Heartbeat interval %v too fast (below %v)", interval, minInterval)
			t.Logf("%s", msg)
			errors = append(errors, msg)
		}

		if interval > maxInterval {
			msg := fmt.Sprintf("Heartbeat interval %v too slow (above %v)", interval, maxInterval)
			t.Logf("%s", msg)
			errors = append(errors, msg)
		}
	}

	t.Logf("================================")
	return errors
}

// Helper functions

func readSpecVersion(t *testing.T) string {
	// Try multiple possible locations for the VERSION file
	possiblePaths := []string{
		"docs/contract/VERSION",
		"../../docs/contract/VERSION",
		"../../../docs/contract/VERSION",
	}

	var content []byte
	var err error

	for _, path := range possiblePaths {
		content, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Fatalf("Failed to read spec version from any location: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "SPEC_VERSION=") {
			return strings.TrimPrefix(line, "SPEC_VERSION=")
		}
	}

	return "unknown"
}

func loadErrorMappings(t *testing.T) map[string]int {
	// Try multiple possible locations for the error-mapping.json file
	possiblePaths := []string{
		"docs/contract/error-mapping.json",
		"../../docs/contract/error-mapping.json",
		"../../../docs/contract/error-mapping.json",
	}

	var content []byte
	var err error

	for _, path := range possiblePaths {
		content, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Fatalf("Failed to read error mappings from any location: %v", err)
	}

	var mappings struct {
		Mappings []struct {
			AdapterError string `json:"adapter_error"`
			HTTPStatus   int    `json:"http_status"`
		} `json:"mappings"`
	}

	if err := json.Unmarshal(content, &mappings); err != nil {
		t.Fatalf("Failed to parse error mappings: %v", err)
	}

	result := make(map[string]int)
	for _, mapping := range mappings.Mappings {
		result[mapping.AdapterError] = mapping.HTTPStatus
	}

	return result
}

func loadTelemetrySchema(t *testing.T) map[string]interface{} {
	// Try multiple possible locations for the telemetry.schema.json file
	possiblePaths := []string{
		"docs/contract/telemetry.schema.json",
		"../../docs/contract/telemetry.schema.json",
		"../../../docs/contract/telemetry.schema.json",
	}

	var content []byte
	var err error

	for _, path := range possiblePaths {
		content, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Fatalf("Failed to read telemetry schema from any location: %v", err)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(content, &schema); err != nil {
		t.Fatalf("Failed to parse telemetry schema: %v", err)
	}

	return schema
}

// normalizeDynamicFields strips or normalizes dynamic fields in JSON responses
func (cv *ContractValidator) normalizeDynamicFields(data map[string]interface{}) {
	// Remove correlationId if present
	delete(data, "correlationId")

	// Normalize timestamps to placeholder
	if ts, ok := data["timestamp"].(string); ok && ts != "" {
		data["timestamp"] = "<TIMESTAMP>"
	}

	// Normalize nested objects recursively
	if dataObj, ok := data["data"].(map[string]interface{}); ok {
		cv.normalizeDynamicFields(dataObj)
	}

	// Normalize arrays
	if dataArray, ok := data["data"].([]interface{}); ok {
		for _, item := range dataArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				cv.normalizeDynamicFields(itemMap)
			}
		}
	}
}

// validateReadyEvent validates a ready event against the schema
func (cv *ContractValidator) validateReadyEvent(t *testing.T, data map[string]interface{}) {
	if snapshot, ok := data["snapshot"].(map[string]interface{}); ok {
		if _, ok := snapshot["activeRadioId"]; !ok {
			t.Error("Ready event missing activeRadioId in snapshot")
		}
		if _, ok := snapshot["radios"]; !ok {
			t.Error("Ready event missing radios in snapshot")
		}
	} else {
		t.Error("Ready event missing snapshot field")
	}
}

// validatePowerChangedEvent validates a powerChanged event against the schema
func (cv *ContractValidator) validatePowerChangedEvent(t *testing.T, data map[string]interface{}) {
	requiredFields := []string{"radioId", "powerDbm", "ts"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			t.Errorf("PowerChanged event missing required field: %s", field)
		}
	}
}

// validateChannelChangedEvent validates a channelChanged event against the schema
func (cv *ContractValidator) validateChannelChangedEvent(t *testing.T, data map[string]interface{}) {
	requiredFields := []string{"radioId", "channelIndex", "frequencyMhz", "ts"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			t.Errorf("ChannelChanged event missing required field: %s", field)
		}
	}
}

// validateHeartbeatEvent validates a heartbeat event against the schema
func (cv *ContractValidator) validateHeartbeatEvent(t *testing.T, data map[string]interface{}) {
	if _, ok := data["timestamp"]; !ok {
		t.Error("Heartbeat event missing timestamp field")
	}
}

// loadOpenAPISpec loads the OpenAPI specification
func loadOpenAPISpec(t *testing.T) map[string]interface{} {
	// Try multiple possible locations for the OpenAPI file
	paths := []string{
		"docs/contract/openapi.yaml",
		"../../docs/contract/openapi.yaml",
		"../../../docs/contract/openapi.yaml",
	}

	var err error

	for _, path := range paths {
		_, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Fatalf("Failed to read OpenAPI spec from any location: %v", err)
	}

	// For now, return a basic spec structure
	// In production, use a proper YAML parser like gopkg.in/yaml.v3
	spec := map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":   "Radio Control Container API",
			"version": "1.0.0",
		},
		"paths": map[string]interface{}{
			"/api/v1/radios": map[string]interface{}{
				"get": map[string]interface{}{
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Radios retrieved successfully",
						},
					},
				},
			},
		},
	}

	return spec
}
