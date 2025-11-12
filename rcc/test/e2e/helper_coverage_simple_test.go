// Package e2e provides simple helper function coverage tests.
// This file tests helper functions to improve E2E coverage with focused success testing.
package e2e

import (
	"testing"
	"time"
)

func TestHelperFunctions_mustHaveNumber_Coverage(t *testing.T) {
	// Test various numeric scenarios to improve coverage
	testCases := []struct {
		name     string
		data     map[string]interface{}
		path     string
		expected float64
	}{
		{
			name:     "integer_value",
			data:     map[string]interface{}{"value": 42},
			path:     "value",
			expected: 42.0,
		},
		{
			name:     "float_value",
			data:     map[string]interface{}{"value": 3.14},
			path:     "value",
			expected: 3.14,
		},
		{
			name:     "zero_value",
			data:     map[string]interface{}{"value": 0.0},
			path:     "value",
			expected: 0.0,
		},
		{
			name:     "negative_value",
			data:     map[string]interface{}{"value": -25.5},
			path:     "value",
			expected: -25.5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mustHaveNumber(t, tc.data, tc.path, tc.expected)
		})
	}
}

func TestHelperFunctions_getJSONPath_Coverage(t *testing.T) {
	// Test various path scenarios to improve coverage
	testCases := []struct {
		name     string
		data     map[string]interface{}
		path     string
		expected interface{}
	}{
		{
			name:     "simple_path",
			data:     map[string]interface{}{"key": "value"},
			path:     "key",
			expected: "value",
		},
		{
			name: "nested_path",
			data: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": "nested_value",
				},
			},
			path:     "level1.level2",
			expected: "nested_value",
		},
		{
			name:     "numeric_path",
			data:     map[string]interface{}{"numeric": 42.5},
			path:     "numeric",
			expected: 42.5,
		},
		{
			name:     "boolean_path",
			data:     map[string]interface{}{"flag": true},
			path:     "flag",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getJSONPath(tc.data, tc.path)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestHelperFunctions_threadSafeResponseWriter_Coverage(t *testing.T) {
	// Test WriteHeader path extensively
	w := newThreadSafeResponseWriter()

	// Test different status codes
	statusCodes := []int{200, 201, 400, 404, 500}
	for _, code := range statusCodes {
		w.WriteHeader(code)
		if w.statusCode != code {
			t.Errorf("Expected status code %d, got %d", code, w.statusCode)
		}
	}

	// Test multiple writes
	testData := []string{"first", "second", "third"}
	for _, data := range testData {
		n, err := w.Write([]byte(data))
		if err != nil {
			t.Errorf("Write failed: %v", err)
		}
		if n != len(data) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
		}
	}

	// Test Header method
	headers := w.Header()
	if headers == nil {
		t.Error("Expected non-nil headers")
	}

	// Test collectEvents
	events := w.collectEvents(50 * time.Millisecond)
	t.Logf("Collected %d events", len(events))
}

func TestHelperFunctions_mustHave_Coverage(t *testing.T) {
	// Test various data types to improve coverage
	testCases := []struct {
		name     string
		data     map[string]interface{}
		path     string
		expected interface{}
	}{
		{
			name:     "string_value",
			data:     map[string]interface{}{"key": "value"},
			path:     "key",
			expected: "value",
		},
		{
			name:     "int_value",
			data:     map[string]interface{}{"number": 42},
			path:     "number",
			expected: 42,
		},
		{
			name:     "float_value",
			data:     map[string]interface{}{"float": 3.14},
			path:     "float",
			expected: 3.14,
		},
		{
			name:     "bool_value",
			data:     map[string]interface{}{"flag": true},
			path:     "flag",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mustHave(t, tc.data, tc.path, tc.expected)
		})
	}
}

func TestHelperFunctions_EdgeCases(t *testing.T) {
	// Test edge cases for better coverage

	// Test getJSONPath with missing path
	data := map[string]interface{}{"key": "value"}
	result := getJSONPath(data, "missing")
	if result != nil {
		t.Errorf("Expected nil for missing path, got %v", result)
	}

	// Test threadSafeResponseWriter with empty writes
	w := newThreadSafeResponseWriter()
	n, err := w.Write([]byte{})
	if err != nil {
		t.Errorf("Empty write failed: %v", err)
	}
	if n != 0 {
		t.Errorf("Expected to write 0 bytes, wrote %d", n)
	}

	// Test mustHave with nil values
	data = map[string]interface{}{"nilValue": nil}
	mustHave(t, data, "nilValue", nil)
}
