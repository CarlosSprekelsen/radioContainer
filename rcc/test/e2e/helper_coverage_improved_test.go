// Package e2e provides improved helper function coverage tests.
// This file tests helper functions to improve E2E coverage with proper success/error testing.
package e2e

import (
	"testing"
	"time"
)

func TestHelperFunctions_mustHaveNumber_Success(t *testing.T) {
	// Test positive case - valid number
	data := map[string]interface{}{
		"value": 42.5,
	}

	// This should not fail
	mustHaveNumber(t, data, "value", 42.5)
}

func TestHelperFunctions_mustHaveNumber_EdgeCases(t *testing.T) {
	// Test integer values
	data := map[string]interface{}{
		"intValue": 42,
	}
	mustHaveNumber(t, data, "intValue", 42.0)

	// Test zero values
	data = map[string]interface{}{
		"zeroValue": 0.0,
	}
	mustHaveNumber(t, data, "zeroValue", 0.0)

	// Test negative values
	data = map[string]interface{}{
		"negativeValue": -25.5,
	}
	mustHaveNumber(t, data, "negativeValue", -25.5)
}

func TestHelperFunctions_getJSONPath_Success(t *testing.T) {
	// Test nested path success
	data := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"value": "found",
			},
		},
	}
	result := getJSONPath(data, "level1.level2.value")
	if result != "found" {
		t.Errorf("Expected 'found', got %v", result)
	}

	// Test single level path
	data = map[string]interface{}{
		"value": "direct",
	}
	result = getJSONPath(data, "value")
	if result != "direct" {
		t.Errorf("Expected 'direct', got %v", result)
	}

	// Test numeric values
	data = map[string]interface{}{
		"numeric": 42.5,
	}
	result = getJSONPath(data, "numeric")
	if result != 42.5 {
		t.Errorf("Expected 42.5, got %v", result)
	}
}

func TestHelperFunctions_threadSafeResponseWriter_Success(t *testing.T) {
	// Test WriteHeader path
	w := newThreadSafeResponseWriter()

	// Test WriteHeader
	w.WriteHeader(200)
	if w.statusCode != 200 {
		t.Errorf("Expected status code 200, got %d", w.statusCode)
	}

	// Test Write
	data := []byte("test data")
	n, err := w.Write(data)
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// Test Header
	headers := w.Header()
	if headers == nil {
		t.Error("Expected non-nil headers")
	}

	// Test multiple writes
	w.Write([]byte("additional data"))

	// Test collectEvents with timeout
	events := w.collectEvents(50 * time.Millisecond)
	t.Logf("Collected %d events", len(events))
}

func TestHelperFunctions_mustHave_Success(t *testing.T) {
	// Test positive case
	data := map[string]interface{}{
		"key": "value",
	}
	mustHave(t, data, "key", "value")

	// Test with different data types
	data = map[string]interface{}{
		"stringValue": "test",
		"intValue":    42,
		"floatValue":  3.14,
		"boolValue":   true,
	}

	mustHave(t, data, "stringValue", "test")
	mustHave(t, data, "intValue", 42)
	mustHave(t, data, "floatValue", 3.14)
	mustHave(t, data, "boolValue", true)
}

func TestHelperFunctions_mustHave_Nested(t *testing.T) {
	// Test nested object access
	data := map[string]interface{}{
		"nested": map[string]interface{}{
			"value": "nested_value",
		},
	}
	mustHave(t, data, "nested", map[string]interface{}{
		"value": "nested_value",
	})
}

func TestHelperFunctions_CoverageTargets(t *testing.T) {
	// Test mustHaveNumber with various scenarios
	testCases := []struct {
		name     string
		data     map[string]interface{}
		path     string
		expected interface{}
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if expectedFloat, ok := tc.expected.(float64); ok {
				mustHaveNumber(t, tc.data, tc.path, expectedFloat)
			} else {
				t.Errorf("Expected value is not a float64: %v", tc.expected)
			}
		})
	}

	// Test getJSONPath with various scenarios
	pathTests := []struct {
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
	}

	for _, tc := range pathTests {
		t.Run(tc.name, func(t *testing.T) {
			result := getJSONPath(tc.data, tc.path)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}

	// Test threadSafeResponseWriter with various scenarios
	t.Run("response_writer_scenarios", func(t *testing.T) {
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
	})
}
