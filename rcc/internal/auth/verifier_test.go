package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewVerifier(t *testing.T) {
	tests := []struct {
		name    string
		config  VerifierConfig
		wantErr bool
	}{
		{
			name: "valid RS256 config with PEM",
			config: VerifierConfig{
				Algorithm:    "RS256",
				PublicKeyPEM: generateTestRSAPublicKeyPEM(t),
				JWKSURL:      "",
			},
			wantErr: false,
		},
		{
			name: "valid RS256 config with JWKS (will fail on fetch but config is valid)",
			config: VerifierConfig{
				Algorithm:           "RS256",
				JWKSURL:             "https://example.com/.well-known/jwks.json",
				JWKSRefreshInterval: 1 * time.Hour,
				JWKSCacheTimeout:    24 * time.Hour,
			},
			wantErr: true, // Will fail because JWKS URL doesn't exist
		},
		{
			name: "valid HS256 config",
			config: VerifierConfig{
				Algorithm: "HS256",
				SecretKey: "test-secret-key",
			},
			wantErr: false,
		},
		{
			name: "invalid algorithm",
			config: VerifierConfig{
				Algorithm: "ES256",
			},
			wantErr: true,
		},
		{
			name: "HS256 without secret",
			config: VerifierConfig{
				Algorithm: "HS256",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier, err := NewVerifier(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewVerifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && verifier == nil {
				t.Error("NewVerifier() returned nil verifier")
			}
		})
	}
}

func TestVerifyHS256Token(t *testing.T) {
	config := VerifierConfig{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Create a valid HS256 token
	claims := jwt.MapClaims{
		"sub":    "user-123",
		"roles":  []string{RoleViewer},
		"scopes": []string{ScopeRead, ScopeTelemetry},
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Test valid token
	verifiedClaims, err := verifier.VerifyToken(tokenString)
	if err != nil {
		t.Errorf("VerifyToken() error = %v", err)
		return
	}

	if verifiedClaims.Subject != "user-123" {
		t.Errorf("Expected subject 'user-123', got '%s'", verifiedClaims.Subject)
	}

	if len(verifiedClaims.Roles) != 1 || verifiedClaims.Roles[0] != RoleViewer {
		t.Errorf("Expected roles [%s], got %v", RoleViewer, verifiedClaims.Roles)
	}

	if len(verifiedClaims.Scopes) != 2 {
		t.Errorf("Expected 2 scopes, got %d", len(verifiedClaims.Scopes))
	}
}

func TestVerifyRS256Token(t *testing.T) {
	// Generate test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Extract public key from private key
	publicKey := &privateKey.PublicKey
	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	})

	config := VerifierConfig{
		Algorithm:    "RS256",
		PublicKeyPEM: string(publicKeyPEM),
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Create a valid RS256 token
	claims := jwt.MapClaims{
		"sub":    "admin-456",
		"roles":  []string{RoleController},
		"scopes": []string{ScopeRead, ScopeControl, ScopeTelemetry},
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Test valid token
	verifiedClaims, err := verifier.VerifyToken(tokenString)
	if err != nil {
		t.Errorf("VerifyToken() error = %v", err)
		return
	}

	if verifiedClaims.Subject != "admin-456" {
		t.Errorf("Expected subject 'admin-456', got '%s'", verifiedClaims.Subject)
	}

	if len(verifiedClaims.Roles) != 1 || verifiedClaims.Roles[0] != RoleController {
		t.Errorf("Expected roles [%s], got %v", RoleController, verifiedClaims.Roles)
	}

	if len(verifiedClaims.Scopes) != 3 {
		t.Errorf("Expected 3 scopes, got %d", len(verifiedClaims.Scopes))
	}
}

func TestVerifyTokenErrors(t *testing.T) {
	config := VerifierConfig{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	tests := []struct {
		name        string
		tokenString string
		wantErr     bool
	}{
		{
			name:        "empty token",
			tokenString: "",
			wantErr:     true,
		},
		{
			name:        "invalid token format",
			tokenString: "invalid.token.here",
			wantErr:     true,
		},
		{
			name:        "wrong algorithm",
			tokenString: createTokenWithWrongAlgorithm(t),
			wantErr:     true,
		},
		{
			name:        "expired token",
			tokenString: createExpiredToken(t),
			wantErr:     true,
		},
		{
			name:        "missing claims",
			tokenString: createTokenWithMissingClaims(t),
			wantErr:     true,
		},
		{
			name:        "invalid roles",
			tokenString: createTokenWithInvalidRoles(t),
			wantErr:     true,
		},
		{
			name:        "invalid scopes",
			tokenString: createTokenWithInvalidScopes(t),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := verifier.VerifyToken(tt.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRoles(t *testing.T) {
	config := VerifierConfig{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	tests := []struct {
		name     string
		roles    []string
		expected bool
	}{
		{
			name:     "valid viewer role",
			roles:    []string{RoleViewer},
			expected: true,
		},
		{
			name:     "valid controller role",
			roles:    []string{RoleController},
			expected: true,
		},
		{
			name:     "multiple valid roles",
			roles:    []string{RoleViewer, RoleController},
			expected: true,
		},
		{
			name:     "invalid role",
			roles:    []string{"admin"},
			expected: false,
		},
		{
			name:     "empty roles",
			roles:    []string{},
			expected: false,
		},
		{
			name:     "mixed valid and invalid",
			roles:    []string{RoleViewer, "admin"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := verifier.validateRoles(tt.roles)
			if result != tt.expected {
				t.Errorf("validateRoles() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestValidateScopes(t *testing.T) {
	config := VerifierConfig{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	tests := []struct {
		name     string
		scopes   []string
		expected bool
	}{
		{
			name:     "valid read scope",
			scopes:   []string{ScopeRead},
			expected: true,
		},
		{
			name:     "valid control scope",
			scopes:   []string{ScopeControl},
			expected: true,
		},
		{
			name:     "valid telemetry scope",
			scopes:   []string{ScopeTelemetry},
			expected: true,
		},
		{
			name:     "multiple valid scopes",
			scopes:   []string{ScopeRead, ScopeControl, ScopeTelemetry},
			expected: true,
		},
		{
			name:     "invalid scope",
			scopes:   []string{"admin"},
			expected: false,
		},
		{
			name:     "empty scopes",
			scopes:   []string{},
			expected: false,
		},
		{
			name:     "mixed valid and invalid",
			scopes:   []string{ScopeRead, "admin"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := verifier.validateScopes(tt.scopes)
			if result != tt.expected {
				t.Errorf("validateScopes() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Helper functions for test token creation

func generateTestRSAPublicKeyPEM(t *testing.T) string {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	publicKey := &privateKey.PublicKey
	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	})

	return string(publicKeyPEM)
}

func createTokenWithWrongAlgorithm(t *testing.T) string {
	claims := jwt.MapClaims{
		"sub":    "user-123",
		"roles":  []string{RoleViewer},
		"scopes": []string{ScopeRead},
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("wrong-secret"))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	return tokenString
}

func createExpiredToken(t *testing.T) string {
	claims := jwt.MapClaims{
		"sub":    "user-123",
		"roles":  []string{RoleViewer},
		"scopes": []string{ScopeRead},
		"iat":    time.Now().Add(-2 * time.Hour).Unix(),
		"exp":    time.Now().Add(-1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	return tokenString
}

func createTokenWithMissingClaims(t *testing.T) string {
	claims := jwt.MapClaims{
		"sub": "user-123",
		// Missing roles and scopes
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	return tokenString
}

func createTokenWithInvalidRoles(t *testing.T) string {
	claims := jwt.MapClaims{
		"sub":    "user-123",
		"roles":  []string{"admin"}, // Invalid role
		"scopes": []string{ScopeRead},
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	return tokenString
}

func createTokenWithInvalidScopes(t *testing.T) string {
	claims := jwt.MapClaims{
		"sub":    "user-123",
		"roles":  []string{RoleViewer},
		"scopes": []string{"admin"}, // Invalid scope
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	return tokenString
}
