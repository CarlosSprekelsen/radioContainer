// Package auth implements JWT token verification with RS256/PEM/JWKS support.
//
//   - OpenAPI v1 §1.1: "Send Authorization: Bearer <token> header on every request (except /health)"
//   - OpenAPI v1 §1.2: "viewer: read-only (list radios, get state, subscribe to telemetry)"
//   - OpenAPI v1 §1.2: "controller: all viewer privileges plus control actions (select radio, set power, set channel)"
package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// VerifierConfig holds configuration for JWT verification.
type VerifierConfig struct {
	// RS256 configuration
	PublicKeyPEM string
	JWKSURL      string

	// HS256 configuration (for tests only)
	SecretKey string

	// Algorithm preference
	Algorithm string // "RS256" or "HS256"

	// JWKS configuration
	JWKSRefreshInterval time.Duration
	JWKSCacheTimeout    time.Duration
}

// JWK represents a JSON Web Key.
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKSet represents a JSON Web Key Set.
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// JWKSCacheEntry represents a cached JWKS key with timestamp.
type JWKSCacheEntry struct {
	Key       *rsa.PublicKey
	Timestamp time.Time
}

// Verifier handles JWT token verification with support for RS256 and HS256.
type Verifier struct {
	config     VerifierConfig
	publicKey  *rsa.PublicKey
	jwksCache  map[string]*JWKSCacheEntry
	jwksMutex  sync.RWMutex
	lastFetch  time.Time
	httpClient *http.Client
}

// NewVerifier creates a new JWT verifier.
func NewVerifier(config VerifierConfig) (*Verifier, error) {
	v := &Verifier{
		config:    config,
		jwksCache: make(map[string]*JWKSCacheEntry),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Initialize based on algorithm
	switch config.Algorithm {
	case "RS256":
		if config.PublicKeyPEM != "" {
			if err := v.loadPublicKeyFromPEM(config.PublicKeyPEM); err != nil {
				return nil, fmt.Errorf("failed to load public key from PEM: %w", err)
			}
		}
		if config.JWKSURL != "" {
			// Fetch initial JWKS
			if err := v.fetchJWKS(); err != nil {
				return nil, fmt.Errorf("failed to fetch initial JWKS: %w", err)
			}
		}
	case "HS256":
		if config.SecretKey == "" {
			return nil, fmt.Errorf("HS256 requires secret key")
		}
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", config.Algorithm)
	}

	return v, nil
}

// VerifyToken verifies a JWT token and returns the claims.
func (v *Verifier) VerifyToken(tokenString string) (*Claims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	switch v.config.Algorithm {
	case "RS256":
		return v.verifyRS256Token(tokenString)
	case "HS256":
		return v.verifyHS256Token(tokenString)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", v.config.Algorithm)
	}
}

// verifyRS256Token verifies a JWT token signed with RS256.
func (v *Verifier) verifyRS256Token(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate algorithm
		if token.Method.Alg() != "RS256" {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			// No kid, use default public key
			if v.publicKey == nil {
				return nil, fmt.Errorf("no public key available")
			}
			return v.publicKey, nil
		}

		// Get key from JWKS cache
		key, err := v.getKeyFromJWKS(kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get key from JWKS: %w", err)
		}

		return key, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return v.extractClaimsFromMap(claims)
}

// verifyHS256Token verifies a JWT token signed with HS256.
func (v *Verifier) verifyHS256Token(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate algorithm
		if token.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(v.config.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return v.extractClaimsFromMap(claims)
}

// extractClaimsFromMap extracts claims from JWT MapClaims.
func (v *Verifier) extractClaimsFromMap(claims *jwt.MapClaims) (*Claims, error) {
	// Extract subject
	sub, ok := (*claims)["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'sub' claim")
	}

	// Extract roles
	roles, err := v.extractStringSlice(claims, "roles")
	if err != nil {
		return nil, fmt.Errorf("missing or invalid 'roles' claim: %w", err)
	}

	// Extract scopes
	scopes, err := v.extractStringSlice(claims, "scopes")
	if err != nil {
		return nil, fmt.Errorf("missing or invalid 'scopes' claim: %w", err)
	}

	// Validate roles
	if !v.validateRoles(roles) {
		return nil, fmt.Errorf("invalid roles: %v", roles)
	}

	// Validate scopes
	if !v.validateScopes(scopes) {
		return nil, fmt.Errorf("invalid scopes: %v", scopes)
	}

	return &Claims{
		Subject: sub,
		Roles:   roles,
		Scopes:  scopes,
	}, nil
}

// extractStringSlice extracts a string slice from claims.
func (v *Verifier) extractStringSlice(claims *jwt.MapClaims, key string) ([]string, error) {
	value, ok := (*claims)[key]
	if !ok {
		return nil, fmt.Errorf("missing claim: %s", key)
	}

	switch val := value.(type) {
	case []string:
		return val, nil
	case []interface{}:
		result := make([]string, len(val))
		for i, item := range val {
			if str, ok := item.(string); ok {
				result[i] = str
			} else {
				return nil, fmt.Errorf("invalid %s claim: not a string", key)
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("invalid %s claim: not a string array", key)
	}
}

// validateRoles validates that all roles are valid.
func (v *Verifier) validateRoles(roles []string) bool {
	validRoles := map[string]bool{
		RoleViewer:     true,
		RoleController: true,
	}

	for _, role := range roles {
		if !validRoles[role] {
			return false
		}
	}

	return len(roles) > 0
}

// validateScopes validates that all scopes are valid.
func (v *Verifier) validateScopes(scopes []string) bool {
	validScopes := map[string]bool{
		ScopeRead:      true,
		ScopeControl:   true,
		ScopeTelemetry: true,
	}

	for _, scope := range scopes {
		if !validScopes[scope] {
			return false
		}
	}

	return len(scopes) > 0
}

// loadPublicKeyFromPEM loads a public key from PEM format.
func (v *Verifier) loadPublicKeyFromPEM(pemData string) error {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("not an RSA public key")
	}

	v.publicKey = rsaPub
	return nil
}

// fetchJWKS fetches the JSON Web Key Set from the configured URL.
func (v *Verifier) fetchJWKS() error {
	if v.config.JWKSURL == "" {
		return fmt.Errorf("JWKS URL not configured")
	}

	resp, err := v.httpClient.Get(v.config.JWKSURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS fetch failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwks JWKSet
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("failed to parse JWKS: %w", err)
	}

	// Cache the keys
	v.jwksMutex.Lock()
	defer v.jwksMutex.Unlock()

	now := time.Now()
	for _, key := range jwks.Keys {
		if key.Kty == "RSA" && key.Use == "sig" && key.Alg == "RS256" {
			pubKey, err := v.jwkToRSAPublicKey(key)
			if err != nil {
				continue // Skip invalid keys
			}
			v.jwksCache[key.Kid] = &JWKSCacheEntry{
				Key:       pubKey,
				Timestamp: now,
			}
		}
	}

	v.lastFetch = time.Now()
	return nil
}

// getKeyFromJWKS gets a public key from the JWKS cache.
func (v *Verifier) getKeyFromJWKS(kid string) (*rsa.PublicKey, error) {
	v.jwksMutex.RLock()
	entry, exists := v.jwksCache[kid]
	v.jwksMutex.RUnlock()

	if exists {
		// Check if cache entry is still valid
		if time.Since(entry.Timestamp) < v.config.JWKSCacheTimeout {
			return entry.Key, nil
		}
		// Entry expired, will need refresh
	}

	// Check if we need to refresh JWKS
	if time.Since(v.lastFetch) > v.config.JWKSRefreshInterval {
		v.jwksMutex.Lock()
		if time.Since(v.lastFetch) > v.config.JWKSRefreshInterval {
			if err := v.fetchJWKS(); err != nil {
				v.jwksMutex.Unlock()
				return nil, fmt.Errorf("failed to refresh JWKS: %w", err)
			}
		}
		v.jwksMutex.Unlock()

		// Try again after refresh
		v.jwksMutex.RLock()
		entry, exists = v.jwksCache[kid]
		v.jwksMutex.RUnlock()

		if exists {
			return entry.Key, nil
		}
	}

	return nil, fmt.Errorf("key not found: %s", kid)
}

// jwkToRSAPublicKey converts a JWK to an RSA public key.
func (v *Verifier) jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode base64url encoded modulus and exponent
	n, err := base64URLDecode(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	e, err := base64URLDecode(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert exponent bytes to int
	var exp int
	for _, b := range e {
		exp = exp<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(n),
		E: exp,
	}, nil
}

// base64URLDecode decodes base64url encoded data.
func base64URLDecode(data string) ([]byte, error) {
	// Add padding if needed
	switch len(data) % 4 {
	case 2:
		data += "=="
	case 3:
		data += "="
	}

	return base64.RawURLEncoding.DecodeString(data)
}
