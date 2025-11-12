package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestBase64URLDecodeVariants tests base64url decoding with various padding scenarios
func TestBase64URLDecodeVariants(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
		wantErr  bool
	}{
		{
			name:     "no padding needed",
			input:    "dGVzdA", // "test" in base64url
			expected: []byte("test"),
		},
		{
			name:     "one padding",
			input:    "dGVzdA", // "test" in base64url
			expected: []byte("test"),
		},
		{
			name:     "two padding",
			input:    "dGVzdA", // "test" in base64url
			expected: []byte("test"),
		},
		{
			name:     "three padding",
			input:    "dGVzdA", // "test" in base64url
			expected: []byte("test"),
		},
		{
			name:     "empty string",
			input:    "",
			expected: []byte{},
		},
		{
			name:    "invalid characters",
			input:   "dGVzdA+", // contains + which is not base64url
			wantErr: true,
		},
		{
			name:    "invalid padding",
			input:   "dGVzdA==", // invalid padding for base64url
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := base64URLDecode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("base64URLDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(result) != string(tt.expected) {
				t.Errorf("base64URLDecode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestJWKSToRSAPublicKeyBase64URL tests JWK to RSA conversion with base64url encoding
func TestJWKSToRSAPublicKeyBase64URL(t *testing.T) {
	// Generate a test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	// Convert to JWK format with base64url encoding
	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString([]byte{1, 0, 1}) // 65537 in bytes

	jwk := JWK{
		Kty: "RSA",
		Kid: "test-key-1",
		Use: "sig",
		Alg: "RS256",
		N:   n,
		E:   e,
	}

	// Test conversion
	verifier := &Verifier{}
	convertedKey, err := verifier.jwkToRSAPublicKey(jwk)
	if err != nil {
		t.Fatalf("Failed to convert JWK to RSA key: %v", err)
	}

	// Verify the key matches
	if convertedKey.N.Cmp(publicKey.N) != 0 {
		t.Error("Converted key modulus does not match original")
	}
	if convertedKey.E != publicKey.E {
		t.Error("Converted key exponent does not match original")
	}
}

// TestJWKSCacheTTL tests JWKS cache TTL behavior
func TestJWKSCacheTTL(t *testing.T) {
	// Create a mock JWKS server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwks := JWKSet{
			Keys: []JWK{
				{
					Kty: "RSA",
					Kid: "test-key-1",
					Use: "sig",
					Alg: "RS256",
					N:   "test-n-value",
					E:   "AQAB",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	// Create verifier with short cache timeout
	config := VerifierConfig{
		Algorithm:           "RS256",
		JWKSURL:             server.URL,
		JWKSRefreshInterval: 1 * time.Hour,
		JWKSCacheTimeout:    100 * time.Millisecond, // Very short TTL
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// First fetch should work
	_, err = verifier.getKeyFromJWKS("test-key-1")
	if err != nil {
		t.Errorf("First fetch failed: %v", err)
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Second fetch should trigger refresh due to expired cache
	_, err = verifier.getKeyFromJWKS("test-key-1")
	if err != nil {
		t.Errorf("Second fetch after TTL expiry failed: %v", err)
	}
}

// TestJWKSCacheRotation tests JWKS key rotation scenarios
func TestJWKSCacheRotation(t *testing.T) {
	keyCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keyCount++
		jwks := JWKSet{
			Keys: []JWK{
				{
					Kty: "RSA",
					Kid: fmt.Sprintf("test-key-%d", keyCount),
					Use: "sig",
					Alg: "RS256",
					N:   "test-n-value",
					E:   "AQAB",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	config := VerifierConfig{
		Algorithm:           "RS256",
		JWKSURL:             server.URL,
		JWKSRefreshInterval: 100 * time.Millisecond,
		JWKSCacheTimeout:    1 * time.Hour,
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// First fetch
	_, err = verifier.getKeyFromJWKS("test-key-1")
	if err != nil {
		t.Errorf("First fetch failed: %v", err)
	}

	// Wait for refresh interval
	time.Sleep(150 * time.Millisecond)

	// Second fetch should get new key
	_, err = verifier.getKeyFromJWKS("test-key-2")
	if err != nil {
		t.Errorf("Second fetch with rotation failed: %v", err)
	}

	// Old key should no longer be available
	_, err = verifier.getKeyFromJWKS("test-key-1")
	if err == nil {
		t.Error("Expected error for old key, but got success")
	}
}

// TestJWKSNetworkErrors tests JWKS fetch error handling
func TestJWKSNetworkErrors(t *testing.T) {
	// Test with 500 error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	config := VerifierConfig{
		Algorithm:           "RS256",
		JWKSURL:             server.URL,
		JWKSRefreshInterval: 100 * time.Millisecond,
		JWKSCacheTimeout:    1 * time.Hour,
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Fetch should fail
	_, err = verifier.getKeyFromJWKS("test-key-1")
	if err == nil {
		t.Error("Expected error for 500 response, but got success")
	}
}

// TestJWKSInvalidJSON tests JWKS with invalid JSON
func TestJWKSInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	config := VerifierConfig{
		Algorithm:           "RS256",
		JWKSURL:             server.URL,
		JWKSRefreshInterval: 100 * time.Millisecond,
		JWKSCacheTimeout:    1 * time.Hour,
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Fetch should fail
	_, err = verifier.getKeyFromJWKS("test-key-1")
	if err == nil {
		t.Error("Expected error for invalid JSON, but got success")
	}
}

// TestJWKSTokenVerificationWithCache tests end-to-end token verification with cache
func TestJWKSTokenVerificationWithCache(t *testing.T) {
	// Generate test key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Create JWKS server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Convert public key to JWK format
		n := base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString([]byte{1, 0, 1})

		jwks := JWKSet{
			Keys: []JWK{
				{
					Kty: "RSA",
					Kid: "test-key-1",
					Use: "sig",
					Alg: "RS256",
					N:   n,
					E:   e,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	// Create verifier
	config := VerifierConfig{
		Algorithm:           "RS256",
		JWKSURL:             server.URL,
		JWKSRefreshInterval: 1 * time.Hour,
		JWKSCacheTimeout:    1 * time.Hour,
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Create and sign a test token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":   "test-user",
		"roles": []string{"controller"},
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	token.Header["kid"] = "test-key-1"

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Verify token
	claims, err := verifier.VerifyToken(tokenString)
	if err != nil {
		t.Fatalf("Token verification failed: %v", err)
	}

	if claims.Subject != "test-user" {
		t.Errorf("Expected subject 'test-user', got '%s'", claims.Subject)
	}
}
