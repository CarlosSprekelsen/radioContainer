package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewMiddleware(t *testing.T) {
	middleware := NewMiddleware()
	if middleware == nil {
		t.Fatal("NewMiddleware() returned nil")
	}
}

func TestExtractBearerToken(t *testing.T) {
	middleware := NewMiddleware()

	tests := []struct {
		name          string
		authHeader    string
		expectError   bool
		expectedToken string
	}{
		{
			name:          "valid bearer token",
			authHeader:    "Bearer test-token",
			expectError:   false,
			expectedToken: "test-token",
		},
		{
			name:        "missing authorization header",
			authHeader:  "",
			expectError: true,
		},
		{
			name:        "invalid format - no bearer",
			authHeader:  "Basic test-token",
			expectError: true,
		},
		{
			name:        "invalid format - no space",
			authHeader:  "Bearertest-token",
			expectError: true,
		},
		{
			name:        "empty token",
			authHeader:  "Bearer ",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if test.authHeader != "" {
				req.Header.Set("Authorization", test.authHeader)
			}

			token, err := middleware.extractBearerToken(req)

			if test.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if token != test.expectedToken {
					t.Errorf("Expected token '%s', got '%s'", test.expectedToken, token)
				}
			}
		})
	}
}

func TestVerifyToken(t *testing.T) {
	middleware := NewMiddleware()

	tests := []struct {
		name           string
		token          string
		expectError    bool
		expectedClaims *Claims
	}{
		{
			name:        "viewer token",
			token:       "viewer-token",
			expectError: false,
			expectedClaims: &Claims{
				Subject: "user-123",
				Roles:   []string{RoleViewer},
				Scopes:  []string{ScopeRead, ScopeTelemetry},
			},
		},
		{
			name:        "controller token",
			token:       "controller-token",
			expectError: false,
			expectedClaims: &Claims{
				Subject: "admin-456",
				Roles:   []string{RoleController},
				Scopes:  []string{ScopeRead, ScopeControl, ScopeTelemetry},
			},
		},
		{
			name:        "invalid token",
			token:       "invalid-token",
			expectError: true,
		},
		{
			name:        "unknown token (defaults to viewer)",
			token:       "unknown-token",
			expectError: false,
			expectedClaims: &Claims{
				Subject: "user-unknown",
				Roles:   []string{RoleViewer},
				Scopes:  []string{ScopeRead, ScopeTelemetry},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims, err := middleware.verifyToken(test.token)

			if test.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if claims == nil {
					t.Fatal("Expected claims, got nil")
				}
				if claims.Subject != test.expectedClaims.Subject {
					t.Errorf("Expected subject '%s', got '%s'", test.expectedClaims.Subject, claims.Subject)
				}
				if len(claims.Roles) != len(test.expectedClaims.Roles) {
					t.Errorf("Expected %d roles, got %d", len(test.expectedClaims.Roles), len(claims.Roles))
				}
				if len(claims.Scopes) != len(test.expectedClaims.Scopes) {
					t.Errorf("Expected %d scopes, got %d", len(test.expectedClaims.Scopes), len(claims.Scopes))
				}
			}
		})
	}
}

func TestHasRequiredScopes(t *testing.T) {
	middleware := NewMiddleware()

	viewerClaims := &Claims{
		Subject: "user-123",
		Roles:   []string{RoleViewer},
		Scopes:  []string{ScopeRead, ScopeTelemetry},
	}

	controllerClaims := &Claims{
		Subject: "admin-456",
		Roles:   []string{RoleController},
		Scopes:  []string{ScopeRead, ScopeControl, ScopeTelemetry},
	}

	tests := []struct {
		name           string
		claims         *Claims
		requiredScopes []string
		expected       bool
	}{
		{
			name:           "viewer has read scope",
			claims:         viewerClaims,
			requiredScopes: []string{ScopeRead},
			expected:       true,
		},
		{
			name:           "viewer has telemetry scope",
			claims:         viewerClaims,
			requiredScopes: []string{ScopeTelemetry},
			expected:       true,
		},
		{
			name:           "viewer lacks control scope",
			claims:         viewerClaims,
			requiredScopes: []string{ScopeControl},
			expected:       false,
		},
		{
			name:           "controller has all scopes",
			claims:         controllerClaims,
			requiredScopes: []string{ScopeRead, ScopeControl, ScopeTelemetry},
			expected:       true,
		},
		{
			name:           "controller has control scope",
			claims:         controllerClaims,
			requiredScopes: []string{ScopeControl},
			expected:       true,
		},
		{
			name:           "nil claims",
			claims:         nil,
			requiredScopes: []string{ScopeRead},
			expected:       false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := middleware.hasRequiredScopes(test.claims, test.requiredScopes)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestHasRequiredRoles(t *testing.T) {
	middleware := NewMiddleware()

	viewerClaims := &Claims{
		Subject: "user-123",
		Roles:   []string{RoleViewer},
		Scopes:  []string{ScopeRead, ScopeTelemetry},
	}

	controllerClaims := &Claims{
		Subject: "admin-456",
		Roles:   []string{RoleController},
		Scopes:  []string{ScopeRead, ScopeControl, ScopeTelemetry},
	}

	tests := []struct {
		name          string
		claims        *Claims
		requiredRoles []string
		expected      bool
	}{
		{
			name:          "viewer has viewer role",
			claims:        viewerClaims,
			requiredRoles: []string{RoleViewer},
			expected:      true,
		},
		{
			name:          "viewer lacks controller role",
			claims:        viewerClaims,
			requiredRoles: []string{RoleController},
			expected:      false,
		},
		{
			name:          "controller has controller role",
			claims:        controllerClaims,
			requiredRoles: []string{RoleController},
			expected:      true,
		},
		{
			name:          "controller has either role",
			claims:        controllerClaims,
			requiredRoles: []string{RoleViewer, RoleController},
			expected:      true,
		},
		{
			name:          "nil claims",
			claims:        nil,
			requiredRoles: []string{RoleViewer},
			expected:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := middleware.hasRequiredRoles(test.claims, test.requiredRoles)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestRequireAuth(t *testing.T) {
	middleware := NewMiddleware()

	// Test handler that checks for claims in context
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaimsFromRequest(r)
		if claims == nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("No claims in context"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}

	// Simple test handler for health endpoint (no claims required)
	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}

	tests := []struct {
		name           string
		authHeader     string
		path           string
		expectedStatus int
	}{
		{
			name:           "valid viewer token",
			authHeader:     "Bearer viewer-token",
			path:           "/api/v1/radios",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid controller token",
			authHeader:     "Bearer controller-token",
			path:           "/api/v1/radios",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing auth header",
			authHeader:     "",
			path:           "/api/v1/radios",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			path:           "/api/v1/radios",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "health endpoint skips auth",
			authHeader:     "",
			path:           "/api/v1/health",
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", test.path, nil)
			if test.authHeader != "" {
				req.Header.Set("Authorization", test.authHeader)
			}
			w := httptest.NewRecorder()

			// Use health handler for health endpoint, test handler for others
			var handler http.HandlerFunc
			if test.path == "/api/v1/health" {
				handler = middleware.RequireAuth(healthHandler)
			} else {
				handler = middleware.RequireAuth(testHandler)
			}
			handler(w, req)

			if w.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, w.Code)
			}
		})
	}
}

func TestRequireScope(t *testing.T) {
	middleware := NewMiddleware()

	// Test handler
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}

	tests := []struct {
		name           string
		authHeader     string
		requiredScopes []string
		expectedStatus int
	}{
		{
			name:           "viewer with read scope",
			authHeader:     "Bearer viewer-token",
			requiredScopes: []string{ScopeRead},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "viewer without control scope",
			authHeader:     "Bearer viewer-token",
			requiredScopes: []string{ScopeControl},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "controller with control scope",
			authHeader:     "Bearer controller-token",
			requiredScopes: []string{ScopeControl},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "controller with multiple scopes",
			authHeader:     "Bearer controller-token",
			requiredScopes: []string{ScopeRead, ScopeControl, ScopeTelemetry},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no auth header",
			authHeader:     "",
			requiredScopes: []string{ScopeRead},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if test.authHeader != "" {
				req.Header.Set("Authorization", test.authHeader)
			}
			w := httptest.NewRecorder()

			// Chain auth and scope middleware
			handler := middleware.RequireAuth(middleware.RequireScope(test.requiredScopes...)(testHandler))
			handler(w, req)

			if w.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, w.Code)
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	middleware := NewMiddleware()

	// Test handler
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}

	tests := []struct {
		name           string
		authHeader     string
		requiredRoles  []string
		expectedStatus int
	}{
		{
			name:           "viewer with viewer role",
			authHeader:     "Bearer viewer-token",
			requiredRoles:  []string{RoleViewer},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "viewer without controller role",
			authHeader:     "Bearer viewer-token",
			requiredRoles:  []string{RoleController},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "controller with controller role",
			authHeader:     "Bearer controller-token",
			requiredRoles:  []string{RoleController},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "controller with either role",
			authHeader:     "Bearer controller-token",
			requiredRoles:  []string{RoleViewer, RoleController},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no auth header",
			authHeader:     "",
			requiredRoles:  []string{RoleViewer},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if test.authHeader != "" {
				req.Header.Set("Authorization", test.authHeader)
			}
			w := httptest.NewRecorder()

			// Chain auth and role middleware
			handler := middleware.RequireAuth(middleware.RequireRole(test.requiredRoles...)(testHandler))
			handler(w, req)

			if w.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetClaimsFromRequest(t *testing.T) {
	middleware := NewMiddleware()

	// Test with claims in context
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer viewer-token")

	// Process through auth middleware to add claims to context
	w := httptest.NewRecorder()
	handler := middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaimsFromRequest(r)
		if claims == nil {
			t.Error("Expected claims, got nil")
		}
		if claims.Subject != "user-123" {
			t.Errorf("Expected subject 'user-123', got '%s'", claims.Subject)
		}
		if !strings.Contains(strings.Join(claims.Roles, ","), RoleViewer) {
			t.Errorf("Expected viewer role, got %v", claims.Roles)
		}
	})
	handler(w, req)

	// Test without claims in context
	req2 := httptest.NewRequest("GET", "/test", nil)
	claims := GetClaimsFromRequest(req2)
	if claims != nil {
		t.Error("Expected nil claims, got non-nil")
	}
}

func TestRoleAndScopeHelpers(t *testing.T) {
	middleware := NewMiddleware()

	viewerClaims := &Claims{
		Subject: "user-123",
		Roles:   []string{RoleViewer},
		Scopes:  []string{ScopeRead, ScopeTelemetry},
	}

	controllerClaims := &Claims{
		Subject: "admin-456",
		Roles:   []string{RoleController},
		Scopes:  []string{ScopeRead, ScopeControl, ScopeTelemetry},
	}

	// Test role helpers
	if !middleware.IsViewer(viewerClaims) {
		t.Error("Expected viewer claims to be viewer")
	}
	if middleware.IsController(viewerClaims) {
		t.Error("Expected viewer claims to not be controller")
	}
	if !middleware.IsController(controllerClaims) {
		t.Error("Expected controller claims to be controller")
	}

	// Test scope helpers
	if !middleware.CanRead(viewerClaims) {
		t.Error("Expected viewer to be able to read")
	}
	if middleware.CanControl(viewerClaims) {
		t.Error("Expected viewer to not be able to control")
	}
	if !middleware.CanControl(controllerClaims) {
		t.Error("Expected controller to be able to control")
	}
	if !middleware.CanAccessTelemetry(viewerClaims) {
		t.Error("Expected viewer to be able to access telemetry")
	}
	if !middleware.CanAccessTelemetry(controllerClaims) {
		t.Error("Expected controller to be able to access telemetry")
	}

	// Test with nil claims
	if middleware.IsViewer(nil) {
		t.Error("Expected nil claims to not be viewer")
	}
	if middleware.IsController(nil) {
		t.Error("Expected nil claims to not be controller")
	}
	if middleware.CanRead(nil) {
		t.Error("Expected nil claims to not be able to read")
	}
	if middleware.CanControl(nil) {
		t.Error("Expected nil claims to not be able to control")
	}
	if middleware.CanAccessTelemetry(nil) {
		t.Error("Expected nil claims to not be able to access telemetry")
	}
}

func TestContextKey(t *testing.T) {
	// Test that context key is properly defined
	if ClaimsKey != "claims" {
		t.Errorf("Expected ClaimsKey to be 'claims', got '%s'", ClaimsKey)
	}
}
