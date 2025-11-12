package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// MockServer creates a test server with auth middleware
func createTestServerWithAuth() (*testServer, *Middleware) {
	// Create auth middleware with mock verifier
	authMiddleware := NewMiddleware()

	// Create test server (simplified for testing)
	server := &testServer{
		authMiddleware: authMiddleware,
	}

	return server, authMiddleware
}

// testServer is a simplified server for testing
type testServer struct {
	authMiddleware *Middleware
}

func TestAPIEndpointAuthentication(t *testing.T) {
	server, _ := createTestServerWithAuth()

	tests := []struct {
		name           string
		method         string
		path           string
		authHeader     string
		expectedStatus int
		description    string
	}{
		// Health endpoint (no auth required)
		{
			name:           "health endpoint no auth",
			method:         "GET",
			path:           "/api/v1/health",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			description:    "Health endpoint should work without authentication",
		},

		// Protected endpoints without auth
		{
			name:           "capabilities no auth",
			method:         "GET",
			path:           "/api/v1/capabilities",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			description:    "Capabilities endpoint should require authentication",
		},
		{
			name:           "radios no auth",
			method:         "GET",
			path:           "/api/v1/radios",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			description:    "Radios endpoint should require authentication",
		},
		{
			name:           "select radio no auth",
			method:         "POST",
			path:           "/api/v1/radios/select",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			description:    "Select radio endpoint should require authentication",
		},
		{
			name:           "telemetry no auth",
			method:         "GET",
			path:           "/api/v1/telemetry",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			description:    "Telemetry endpoint should require authentication",
		},

		// Protected endpoints with valid viewer token
		{
			name:           "capabilities with viewer token",
			method:         "GET",
			path:           "/api/v1/capabilities",
			authHeader:     "Bearer viewer-token",
			expectedStatus: http.StatusOK,
			description:    "Capabilities endpoint should work with viewer token",
		},
		{
			name:           "radios with viewer token",
			method:         "GET",
			path:           "/api/v1/radios",
			authHeader:     "Bearer viewer-token",
			expectedStatus: http.StatusOK,
			description:    "Radios endpoint should work with viewer token",
		},
		{
			name:           "telemetry with viewer token",
			method:         "GET",
			path:           "/api/v1/telemetry",
			authHeader:     "Bearer viewer-token",
			expectedStatus: http.StatusOK,
			description:    "Telemetry endpoint should work with viewer token",
		},

		// Control endpoints with viewer token (should fail)
		{
			name:           "select radio with viewer token",
			method:         "POST",
			path:           "/api/v1/radios/select",
			authHeader:     "Bearer viewer-token",
			expectedStatus: http.StatusForbidden,
			description:    "Select radio should fail with viewer token (no control scope)",
		},

		// Control endpoints with controller token
		{
			name:           "select radio with controller token",
			method:         "POST",
			path:           "/api/v1/radios/select",
			authHeader:     "Bearer controller-token",
			expectedStatus: http.StatusOK,
			description:    "Select radio should work with controller token",
		},

		// Invalid tokens
		{
			name:           "capabilities with invalid token",
			method:         "GET",
			path:           "/api/v1/capabilities",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			description:    "Invalid token should return 401",
		},
		{
			name:           "capabilities with malformed header",
			method:         "GET",
			path:           "/api/v1/capabilities",
			authHeader:     "Basic invalid-token",
			expectedStatus: http.StatusUnauthorized,
			description:    "Malformed auth header should return 401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			var body bytes.Buffer
			if tt.method == "POST" {
				_ = json.NewEncoder(&body).Encode(map[string]string{"id": "test-radio"})
			}

			req := httptest.NewRequest(tt.method, tt.path, &body)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Create a simple test handler that simulates the API behavior
			handler := createTestHandler(server, tt.path)
			handler(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Description: %s",
					tt.expectedStatus, w.Code, tt.description)
			}
		})
	}
}

func TestScopeBasedAuthorization(t *testing.T) {
	server, _ := createTestServerWithAuth()

	tests := []struct {
		name           string
		method         string
		path           string
		authHeader     string
		expectedStatus int
		description    string
	}{
		// Test read scope requirements
		{
			name:           "get radio power with viewer token",
			method:         "GET",
			path:           "/api/v1/radios/test-radio/power",
			authHeader:     "Bearer viewer-token",
			expectedStatus: http.StatusOK,
			description:    "GET power should work with viewer token (read scope)",
		},
		{
			name:           "get radio channel with viewer token",
			method:         "GET",
			path:           "/api/v1/radios/test-radio/channel",
			authHeader:     "Bearer viewer-token",
			expectedStatus: http.StatusOK,
			description:    "GET channel should work with viewer token (read scope)",
		},

		// Test control scope requirements
		{
			name:           "set radio power with viewer token",
			method:         "POST",
			path:           "/api/v1/radios/test-radio/power",
			authHeader:     "Bearer viewer-token",
			expectedStatus: http.StatusForbidden,
			description:    "POST power should fail with viewer token (no control scope)",
		},
		{
			name:           "set radio channel with viewer token",
			method:         "POST",
			path:           "/api/v1/radios/test-radio/channel",
			authHeader:     "Bearer viewer-token",
			expectedStatus: http.StatusForbidden,
			description:    "POST channel should fail with viewer token (no control scope)",
		},
		{
			name:           "set radio power with controller token",
			method:         "POST",
			path:           "/api/v1/radios/test-radio/power",
			authHeader:     "Bearer controller-token",
			expectedStatus: http.StatusOK,
			description:    "POST power should work with controller token (control scope)",
		},
		{
			name:           "set radio channel with controller token",
			method:         "POST",
			path:           "/api/v1/radios/test-radio/channel",
			authHeader:     "Bearer controller-token",
			expectedStatus: http.StatusOK,
			description:    "POST channel should work with controller token (control scope)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			var body bytes.Buffer
			if tt.method == "POST" {
				if tt.path == "/api/v1/radios/test-radio/power" {
					_ = json.NewEncoder(&body).Encode(map[string]int{"powerDbm": 30})
				} else if tt.path == "/api/v1/radios/test-radio/channel" {
					_ = json.NewEncoder(&body).Encode(map[string]int{"channelIndex": 1})
				}
			}

			req := httptest.NewRequest(tt.method, tt.path, &body)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Create a simple test handler that simulates the API behavior
			handler := createTestHandler(server, tt.path)
			handler(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Description: %s",
					tt.expectedStatus, w.Code, tt.description)
			}
		})
	}
}

func TestErrorResponseFormat(t *testing.T) {
	server, _ := createTestServerWithAuth()

	tests := []struct {
		name           string
		authHeader     string
		path           string
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "no auth header",
			authHeader:     "",
			path:           "/api/v1/capabilities",
			expectedCode:   "UNAUTHORIZED",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			path:           "/api/v1/capabilities",
			expectedCode:   "UNAUTHORIZED",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "insufficient permissions",
			authHeader:     "Bearer viewer-token",
			path:           "/api/v1/radios/select",
			expectedCode:   "FORBIDDEN",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			handler := createTestHandler(server, tt.path)
			handler(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check error response format
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
				if code, ok := response["code"].(string); ok {
					if code != tt.expectedCode {
						t.Errorf("Expected error code '%s', got '%s'", tt.expectedCode, code)
					}
				} else {
					t.Error("Expected error response to contain 'code' field")
				}

				// Check for correlation ID
				if _, hasCorrelationID := response["correlationId"]; !hasCorrelationID {
					t.Error("Expected error response to contain 'correlationId' field")
				}
			}
		})
	}
}

// createTestHandler creates a test handler that simulates API behavior
func createTestHandler(server *testServer, path string) http.HandlerFunc {
	// This is a simplified test handler that simulates the API behavior
	// In a real implementation, this would use the actual API handlers

	return func(w http.ResponseWriter, r *http.Request) {
		// Simulate health endpoint (no auth required)
		if path == "/api/v1/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}

		// For other endpoints, we need to apply auth middleware
		// This is a simplified version for testing
		authMiddleware := server.authMiddleware
		if authMiddleware == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Apply auth middleware with scope checking based on path
		var authHandler http.HandlerFunc

		// Determine required scope based on path and method
		switch {
		case strings.Contains(path, "/radios/select") ||
			(strings.Contains(path, "/power") && r.Method == "POST") ||
			(strings.Contains(path, "/channel") && r.Method == "POST"):
			// Control operations require control scope
			authHandler = authMiddleware.RequireAuth(authMiddleware.RequireScope(ScopeControl)(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"result":"ok"}`))
			}))
		case strings.Contains(path, "/telemetry"):
			// Telemetry requires telemetry scope
			authHandler = authMiddleware.RequireAuth(authMiddleware.RequireScope(ScopeTelemetry)(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"result":"ok"}`))
			}))
		default:
			// Read operations require read scope
			authHandler = authMiddleware.RequireAuth(authMiddleware.RequireScope(ScopeRead)(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"result":"ok"}`))
			}))
		}

		authHandler(w, r)
	}
}
