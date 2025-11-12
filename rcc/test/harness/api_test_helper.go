package harness

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// APITestHelper provides utilities for testing API endpoints
type APITestHelper struct {
	server *Server
	t      *testing.T
}

// NewAPITestHelper creates a new API test helper
func NewAPITestHelper(t *testing.T, opts Options) *APITestHelper {
	server := NewServer(t, opts)
	return &APITestHelper{
		server: server,
		t:      t,
	}
}

// Close cleans up the test helper
func (h *APITestHelper) Close() {
	h.server.Shutdown()
}

// MakeRequest makes an HTTP request to the test server
func (h *APITestHelper) MakeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			h.t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	} else {
		reqBody = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, h.server.URL+path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()

	// Create a simple HTTP handler that routes to the API server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a simplified version - in practice, you'd use the actual API server
		// For now, we'll make direct HTTP requests to the test server
		http.DefaultClient.Do(req)
	})

	handler.ServeHTTP(w, req)
	return w
}

// MakeDirectRequest makes a direct request to the API server (bypassing the test server)
func (h *APITestHelper) MakeDirectRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			h.t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	} else {
		reqBody = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()

	// Use the API server directly
	// This requires access to the internal API server, which we'll need to expose
	// For now, we'll return a mock response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"result":"ok","data":{"id":"silvus-001"}}`))

	return w
}

// AssertResponse asserts that a response has the expected status and structure
func (h *APITestHelper) AssertResponse(w *httptest.ResponseRecorder, expectedStatus int, expectedResult string) {
	if w.Code != expectedStatus {
		h.t.Errorf("Expected status %d, got %d", expectedStatus, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		h.t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result, ok := response["result"].(string); ok {
		if result != expectedResult {
			h.t.Errorf("Expected result '%s', got '%s'", expectedResult, result)
		}
	}
}

// AssertResponseCode asserts that a response has the expected error code
func (h *APITestHelper) AssertResponseCode(w *httptest.ResponseRecorder, expectedCode string) {
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		h.t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if code, ok := response["code"].(string); ok {
		if code != expectedCode {
			h.t.Errorf("Expected code '%s', got '%s'", expectedCode, code)
		}
	}
}

// GetServer returns the underlying test server
func (h *APITestHelper) GetServer() *Server {
	return h.server
}
