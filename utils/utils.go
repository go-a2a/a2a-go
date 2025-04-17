// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package utils provides utility functions for the A2A protocol.
package utils

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/go-a2a/a2a"
)

// JSONWebKey represents a JSON Web Key.
type JSONWebKey struct {
	// KID is the key ID.
	KID string `json:"kid"`
	// KTY is the key type.
	KTY string `json:"kty"`
	// ALG is the algorithm.
	ALG string `json:"alg"`
	// USE is the key usage.
	USE string `json:"use"`
	// CRV is the curve name.
	CRV string `json:"crv,omitempty"`
	// X is the x-coordinate.
	X string `json:"x,omitempty"`
	// Y is the y-coordinate.
	Y string `json:"y,omitempty"`
	// N is the modulus value.
	N string `json:"n,omitempty"`
	// E is the exponent value.
	E string `json:"e,omitempty"`
}

// JSONWebKeySet represents a set of JSON Web Keys.
type JSONWebKeySet struct {
	// Keys is the list of keys.
	Keys []JSONWebKey `json:"keys"`
}

// KeyPair holds a private key and its corresponding public JWK.
type KeyPair struct {
	// PrivateKey is the private key.
	PrivateKey *ecdsa.PrivateKey
	// PublicJWK is the public key in JWK format.
	PublicJWK JSONWebKey
}

// KeyManager manages keys for JWT signing and verification.
type KeyManager struct {
	mu       sync.RWMutex
	keyPairs map[string]*KeyPair
	jwks     JSONWebKeySet
}

// NewKeyManager creates a new key manager.
func NewKeyManager() *KeyManager {
	return &KeyManager{
		keyPairs: make(map[string]*KeyPair),
		jwks: JSONWebKeySet{
			Keys: []JSONWebKey{},
		},
	}
}

// GenerateKeyPair generates a new ECDSA key pair and adds it to the manager.
func (m *KeyManager) GenerateKeyPair(kid string) (*KeyPair, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate a new ECDSA private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Convert the public key to JWK format
	x := base64.RawURLEncoding.EncodeToString(privateKey.X.Bytes())
	y := base64.RawURLEncoding.EncodeToString(privateKey.Y.Bytes())

	jwk := JSONWebKey{
		KID: kid,
		KTY: "EC",
		ALG: "ES256",
		USE: "sig",
		CRV: "P-256",
		X:   x,
		Y:   y,
	}

	keyPair := &KeyPair{
		PrivateKey: privateKey,
		PublicJWK:  jwk,
	}

	// Add the key pair to the manager
	m.keyPairs[kid] = keyPair
	m.jwks.Keys = append(m.jwks.Keys, jwk)

	return keyPair, nil
}

// GetKeyPair returns a key pair by key ID.
func (m *KeyManager) GetKeyPair(kid string) (*KeyPair, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keyPair, ok := m.keyPairs[kid]
	if !ok {
		return nil, fmt.Errorf("key pair not found: %s", kid)
	}

	return keyPair, nil
}

// GetJWKS returns the JSON Web Key Set.
func (m *KeyManager) GetJWKS() JSONWebKeySet {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.jwks
}

// SignJWT signs a JWT with a key pair.
func (m *KeyManager) SignJWT(kid string, claims jwt.Claims, expiration time.Duration) (string, error) {
	keyPair, err := m.GetKeyPair(kid)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = kid

	return token.SignedString(keyPair.PrivateKey)
}

// CreateJWKSHandler creates an HTTP handler for serving the JWKS.
func (m *KeyManager) CreateJWKSHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwks := m.GetJWKS()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}
}

// JWTVerifier verifies JWTs signed with a JWK.
type JWTVerifier struct {
	// jwksURL is the URL of the JWKS endpoint.
	jwksURL string
	// client is the HTTP client used for fetching the JWKS.
	client *http.Client
	// jwks is the cached JWKS.
	jwks *JSONWebKeySet
	// keyMap is a map of key IDs to parsed keys.
	keyMap map[string]any
	// mu protects the cache.
	mu sync.RWMutex
	// lastFetch is the time of the last JWKS fetch.
	lastFetch time.Time
	// cacheDuration is the duration for which the JWKS cache is valid.
	cacheDuration time.Duration
}

// NewJWTVerifier creates a new JWT verifier.
func NewJWTVerifier(jwksURL string) *JWTVerifier {
	return &JWTVerifier{
		jwksURL:       jwksURL,
		client:        &http.Client{Timeout: 10 * time.Second},
		keyMap:        make(map[string]any),
		cacheDuration: 1 * time.Hour,
	}
}

// fetchJWKS fetches the JWKS from the endpoint.
func (v *JWTVerifier) fetchJWKS() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Check if we need to refresh the cache
	if v.jwks != nil && time.Since(v.lastFetch) < v.cacheDuration {
		return nil
	}

	// Fetch the JWKS
	resp, err := v.client.Get(v.jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response: %d %s", resp.StatusCode, resp.Status)
	}

	// Decode the JWKS
	var jwks JSONWebKeySet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	// Parse the keys
	keyMap := make(map[string]any)
	for _, jwk := range jwks.Keys {
		var key any
		var err error

		switch jwk.KTY {
		case "EC":
			key, err = v.parseECKey(&jwk)
		case "RSA":
			key, err = v.parseRSAKey(&jwk)
		default:
			return fmt.Errorf("unsupported key type: %s", jwk.KTY)
		}

		if err != nil {
			return fmt.Errorf("failed to parse key: %w", err)
		}

		keyMap[jwk.KID] = key
	}

	// Update the cache
	v.jwks = &jwks
	v.keyMap = keyMap
	v.lastFetch = time.Now()

	return nil
}

// parseECKey parses an EC JWK.
func (v *JWTVerifier) parseECKey(jwk *JSONWebKey) (*ecdsa.PublicKey, error) {
	// Decode the x and y coordinates
	xData, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode x coordinate: %w", err)
	}

	yData, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode y coordinate: %w", err)
	}

	// Convert to big integers
	x := new(big.Int).SetBytes(xData)
	y := new(big.Int).SetBytes(yData)

	// Create the public key
	var curve elliptic.Curve
	switch jwk.CRV {
	case "P-256":
		curve = elliptic.P256()
	case "P-384":
		curve = elliptic.P384()
	case "P-521":
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported curve: %s", jwk.CRV)
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}

// parseRSAKey parses an RSA JWK.
func (v *JWTVerifier) parseRSAKey(jwk *JSONWebKey) (any, error) {
	// For this example, we'll just return a placeholder
	// In a real implementation, this would parse the n and e parameters
	return nil, errors.New("RSA key parsing not implemented")
}

// GetKey returns a key by key ID.
func (v *JWTVerifier) GetKey(kid string) (any, error) {
	// Ensure we have the latest JWKS
	if err := v.fetchJWKS(); err != nil {
		return nil, err
	}

	v.mu.RLock()
	defer v.mu.RUnlock()

	key, ok := v.keyMap[kid]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", kid)
	}

	return key, nil
}

// VerifyJWT verifies a JWT.
func (v *JWTVerifier) VerifyJWT(tokenString string, claims jwt.Claims) error {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid not found in token header")
		}

		// Get the key
		key, err := v.GetKey(kid)
		if err != nil {
			return nil, err
		}

		return key, nil
	})
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return errors.New("token is invalid")
	}

	return nil
}

// AuthMiddleware creates an HTTP middleware for JWT authentication.
func (v *JWTVerifier) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := tokenParts[1]

		// Verify the token
		claims := jwt.MapClaims{}
		if err := v.VerifyJWT(tokenString, claims); err != nil {
			http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		// Set claims in the request context
		ctx := r.Context()
		for key, val := range claims {
			ctx = context.WithValue(ctx, key, val)
		}

		// Call the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// PushNotificationSender sends push notifications to clients.
type PushNotificationSender struct {
	// keyManager is the key manager for signing JWTs.
	keyManager *KeyManager
	// client is the HTTP client used for sending notifications.
	client *http.Client
}

// NewPushNotificationSender creates a new push notification sender.
func NewPushNotificationSender(keyManager *KeyManager) *PushNotificationSender {
	return &PushNotificationSender{
		keyManager: keyManager,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// SendStatusUpdate sends a task status update notification.
func (s *PushNotificationSender) SendStatusUpdate(ctx context.Context, config *a2a.PushNotificationConfig, event *a2a.TaskStatusUpdateEvent) error {
	return s.sendNotification(ctx, config, &a2a.JSONRPCEvent{
		JSONRPC: "2.0",
		Method:  "task/status",
		Params:  event,
	})
}

// SendArtifactUpdate sends a task artifact update notification.
func (s *PushNotificationSender) SendArtifactUpdate(ctx context.Context, config *a2a.PushNotificationConfig, event *a2a.TaskArtifactUpdateEvent) error {
	return s.sendNotification(ctx, config, &a2a.JSONRPCEvent{
		JSONRPC: "2.0",
		Method:  "task/artifact",
		Params:  event,
	})
}

// SendTaskCompletion sends a task completion notification.
func (s *PushNotificationSender) SendTaskCompletion(ctx context.Context, config *a2a.PushNotificationConfig, task *a2a.Task) error {
	return s.sendNotification(ctx, config, &a2a.JSONRPCEvent{
		JSONRPC: "2.0",
		Method:  "task/complete",
		Params:  task,
	})
}

// sendNotification sends a notification to a webhook URL.
func (s *PushNotificationSender) sendNotification(ctx context.Context, config *a2a.PushNotificationConfig, event *a2a.JSONRPCEvent) error {
	if config == nil || config.URL == "" {
		return errors.New("push notification config not set")
	}

	// Marshal the event
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.URL, bytes.NewReader(eventJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication if configured
	if config.Authentication != nil && len(config.Authentication.Schemes) > 0 {
		for _, scheme := range config.Authentication.Schemes {
			if strings.HasPrefix(scheme, "jwt") {
				// Extract the key ID from the scheme
				kidStart := strings.Index(scheme, "+")
				if kidStart == -1 {
					return errors.New("invalid JWT scheme format")
				}
				kid := scheme[kidStart+1:]

				// Sign a JWT
				claims := jwt.MapClaims{
					"iss": "a2a-go",
					"sub": "push-notification",
					"aud": config.URL,
					"exp": time.Now().Add(5 * time.Minute).Unix(),
					"iat": time.Now().Unix(),
					"jti": uuid.New().String(),
				}

				token, err := s.keyManager.SignJWT(kid, claims, 5*time.Minute)
				if err != nil {
					return fmt.Errorf("failed to sign JWT: %w", err)
				}

				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				break
			}
		}
	}

	// Send the request
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("received non-OK response: %d %s: %s", resp.StatusCode, resp.Status, string(bodyBytes))
	}

	return nil
}
