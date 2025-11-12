// Package e2e provides shared helper functions for end-to-end tests.
package e2e

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

// HTTP helper functions
func httpGetJSON(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s returned status %d", url, resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	return result
}

func httpPostJSON200(t *testing.T, url string, payload map[string]any) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST %s returned status %d", url, resp.StatusCode)
	}
}

func httpGetWithStatus(t *testing.T, url string) *http.Response {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s failed: %v", url, err)
	}
	return resp
}

func httpPostWithStatus(t *testing.T, url, payload string) *http.Response {
	resp, err := http.Post(url, "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	return resp
}

// JSON assertion helpers
func mustHave(t *testing.T, data map[string]interface{}, path string, expected interface{}) {
	actual := getJSONPath(data, path)

	// Handle map comparison safely
	if expectedMap, ok := expected.(map[string]interface{}); ok {
		if actualMap, ok := actual.(map[string]interface{}); ok {
			// Compare maps by converting to JSON strings
			expectedJSON, _ := json.Marshal(expectedMap)
			actualJSON, _ := json.Marshal(actualMap)
			if string(expectedJSON) != string(actualJSON) {
				t.Errorf("Expected %s to be %v, got %v", path, expected, actual)
			}
			return
		}
	}

	// For non-map types, use direct comparison
	if actual != expected {
		t.Errorf("Expected %s to be %v, got %v", path, expected, actual)
	}
}

func mustHaveNumber(t *testing.T, data map[string]interface{}, path string, expected float64) {
	actual := getJSONPath(data, path)
	var num float64
	switch v := actual.(type) {
	case float64:
		num = v
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	case float32:
		num = float64(v)
	default:
		t.Errorf("Expected %s to be a number, got %T: %v", path, actual, actual)
		return
	}
	if num != expected {
		t.Errorf("Expected %s to be %v, got %v", path, expected, num)
	}
}

func getJSONPath(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			return current[part]
		}

		// Handle array indices (e.g., "items[0]")
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			// Extract array name and index
			openBracket := strings.Index(part, "[")
			closeBracket := strings.Index(part, "]")
			if openBracket > 0 && closeBracket > openBracket {
				arrayName := part[:openBracket]
				indexStr := part[openBracket+1 : closeBracket]

				// Get the array
				if array, ok := current[arrayName].([]interface{}); ok {
					// Parse index
					if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 && index < len(array) {
						if i == len(parts)-1 {
							return array[index]
						}
						// Convert to map for next iteration
						if next, ok := array[index].(map[string]interface{}); ok {
							current = next
						} else {
							return nil
						}
					} else {
						return nil
					}
				} else {
					return nil
				}
			} else {
				return nil
			}
		} else {
			// Handle regular map access
			if next, ok := current[part].(map[string]interface{}); ok {
				current = next
			} else {
				return nil
			}
		}
	}

	return nil
}

// Thread-safe response writer for SSE testing
type threadSafeResponseWriter struct {
	events     chan string
	headers    http.Header
	statusCode int
}

func newThreadSafeResponseWriter() *threadSafeResponseWriter {
	return &threadSafeResponseWriter{
		events:     make(chan string, 100),
		headers:    make(http.Header),
		statusCode: 200,
	}
}

func (w *threadSafeResponseWriter) Header() http.Header {
	return w.headers
}

func (w *threadSafeResponseWriter) Write(data []byte) (int, error) {
	select {
	case w.events <- string(data):
		return len(data), nil
	default:
		return len(data), nil
	}
}

func (w *threadSafeResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *threadSafeResponseWriter) collectEvents(timeout time.Duration) []string {
	var events []string
	timeoutChan := time.After(timeout)

	for {
		select {
		case event := <-w.events:
			events = append(events, event)
		case <-timeoutChan:
			return events
		}
	}
}
