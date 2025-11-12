//
//
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Claims represents the parsed token claims.
type Claims struct {
	Subject string   `json:"sub"`
	Roles   []string `json:"roles"`
	Scopes  []string `json:"scopes"`
}

// ContextKey is used for storing claims in request context.
type ContextKey string

const (
	ClaimsKey ContextKey = "claims"
)

// Role constants per OpenAPI v1 §1.2
const (
	RoleViewer     = "viewer"
	RoleController = "controller"
)

// Scope constants per OpenAPI v1 §1.2
const (
	ScopeRead      = "read"
	ScopeControl   = "control"
	ScopeTelemetry = "telemetry"
)

// Middleware handles authentication and authorization.
type Middleware struct {
	verifier *Verifier
}

// NewMiddleware creates a new auth middleware.
func NewMiddleware() *Middleware {
	return &Middleware{}
}

// NewMiddlewareWithVerifier creates a new auth middleware with a JWT verifier.
func NewMiddlewareWithVerifier(verifier *Verifier) *Middleware {
	return &Middleware{
		verifier: verifier,
	}
}

// RequireAuth creates middleware that requires authentication.
func (m *Middleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health endpoint
		if r.URL.Path == "/api/v1/health" {
			next(w, r)
			return
		}

		// Extract bearer token
		token, err := m.extractBearerToken(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED",
				"Authentication required", nil)
			return
		}

		// Verify token and extract claims
		claims, err := m.verifyToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED",
				"Invalid token", nil)
			return
		}

		// Store claims in context
		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}

// RequireScope creates middleware that requires specific scopes.
func (m *Middleware) RequireScope(requiredScopes ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims := m.getClaimsFromContext(r.Context())
			if claims == nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED",
					"Authentication required", nil)
				return
			}

			// Check if user has required scopes
			if !m.hasRequiredScopes(claims, requiredScopes) {
				writeError(w, http.StatusForbidden, "FORBIDDEN",
					"Insufficient permissions", nil)
				return
			}

			next(w, r)
		}
	}
}

// RequireRole creates middleware that requires specific roles.
func (m *Middleware) RequireRole(requiredRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims := m.getClaimsFromContext(r.Context())
			if claims == nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED",
					"Authentication required", nil)
				return
			}

			// Check if user has required roles
			if !m.hasRequiredRoles(claims, requiredRoles) {
				writeError(w, http.StatusForbidden, "FORBIDDEN",
					"Insufficient permissions", nil)
				return
			}

			next(w, r)
		}
	}
}

// extractBearerToken extracts the bearer token from the Authorization header.
func (m *Middleware) extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	// Check for Bearer prefix
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", fmt.Errorf("empty token")
	}

	return token, nil
}

// verifyToken verifies the token and returns claims.
func (m *Middleware) verifyToken(token string) (*Claims, error) {
	// Use real verifier if available
	if m.verifier != nil {
		return m.verifier.VerifyToken(token)
	}

	// Fallback to mock implementation for backward compatibility
	// This should only be used in tests
	switch token {
	case "viewer-token":
		return &Claims{
			Subject: "user-123",
			Roles:   []string{RoleViewer},
			Scopes:  []string{ScopeRead, ScopeTelemetry},
		}, nil
	case "controller-token":
		return &Claims{
			Subject: "admin-456",
			Roles:   []string{RoleController},
			Scopes:  []string{ScopeRead, ScopeControl, ScopeTelemetry},
		}, nil
	case "invalid-token":
		return nil, fmt.Errorf("token verification failed")
	default:
		// Default to viewer for unknown tokens (test mode only)
		return &Claims{
			Subject: "user-unknown",
			Roles:   []string{RoleViewer},
			Scopes:  []string{ScopeRead, ScopeTelemetry},
		}, nil
	}
}

// hasRequiredScopes checks if the user has all required scopes.
func (m *Middleware) hasRequiredScopes(claims *Claims, requiredScopes []string) bool {
	if claims == nil {
		return false
	}

	for _, required := range requiredScopes {
		found := false
		for _, scope := range claims.Scopes {
			if scope == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// hasRequiredRoles checks if the user has any of the required roles.
func (m *Middleware) hasRequiredRoles(claims *Claims, requiredRoles []string) bool {
	if claims == nil {
		return false
	}

	// If no roles are required, return true (no requirements)
	if len(requiredRoles) == 0 {
		return true
	}

	for _, required := range requiredRoles {
		for _, role := range claims.Roles {
			if role == required {
				return true
			}
		}
	}

	return false
}

// getClaimsFromContext extracts claims from the request context.
func (m *Middleware) getClaimsFromContext(ctx context.Context) *Claims {
	claims, ok := ctx.Value(ClaimsKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// GetClaimsFromRequest extracts claims from the request context.
// This is a helper function for use in handlers.
func GetClaimsFromRequest(r *http.Request) *Claims {
	claims, ok := r.Context().Value(ClaimsKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// IsViewer checks if the user has viewer role.
func (m *Middleware) IsViewer(claims *Claims) bool {
	return m.hasRequiredRoles(claims, []string{RoleViewer})
}

// IsController checks if the user has controller role.
func (m *Middleware) IsController(claims *Claims) bool {
	return m.hasRequiredRoles(claims, []string{RoleController})
}

// CanRead checks if the user can perform read operations.
func (m *Middleware) CanRead(claims *Claims) bool {
	return m.hasRequiredScopes(claims, []string{ScopeRead})
}

// CanControl checks if the user can perform control operations.
func (m *Middleware) CanControl(claims *Claims) bool {
	return m.hasRequiredScopes(claims, []string{ScopeControl})
}

// CanAccessTelemetry checks if the user can access telemetry.
func (m *Middleware) CanAccessTelemetry(claims *Claims) bool {
	return m.hasRequiredScopes(claims, []string{ScopeTelemetry})
}

// writeError writes an error response in the API format.
func writeError(w http.ResponseWriter, status int, code, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]interface{}{
		"result":        "error",
		"code":          code,
		"message":       message,
		"correlationId": generateCorrelationID(),
	}

	if details != nil {
		response["details"] = details
	}

	_ = json.NewEncoder(w).Encode(response)
}

// generateCorrelationID generates a simple correlation ID for request tracking.
func generateCorrelationID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
