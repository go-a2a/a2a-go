// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

// CredentialService defines the interface for managing credentials.
type CredentialService interface {
	// GetCredentials retrieves credentials for a given context.
	GetCredentials(ctx context.Context, contextID string) (*Credentials, error)

	// StoreCredentials stores credentials for a given context.
	StoreCredentials(ctx context.Context, contextID string, credentials *Credentials) error

	// DeleteCredentials deletes credentials for a given context.
	DeleteCredentials(ctx context.Context, contextID string) error

	// ValidateCredentials validates if credentials are still valid.
	ValidateCredentials(ctx context.Context, credentials *Credentials) error
}

// Credentials represents authentication credentials.
type Credentials struct {
	Type         CredentialType `json:"type"`
	AccessToken  string         `json:"access_token,omitempty"`
	TokenType    string         `json:"token_type,omitempty"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
	RefreshToken string         `json:"refresh_token,omitempty"`
	Scope        string         `json:"scope,omitempty"`
	APIKey       string         `json:"api_key,omitempty"`
	Username     string         `json:"username,omitempty"`
	Password     string         `json:"password,omitempty"`
	Certificate  string         `json:"certificate,omitempty"`
	PrivateKey   string         `json:"private_key,omitempty"`
}

// CredentialType represents the type of credentials.
type CredentialType string

// Credential types.
const (
	CredentialTypeNone        CredentialType = "none"
	CredentialTypeAPIKey      CredentialType = "api_key"
	CredentialTypeBearer      CredentialType = "bearer"
	CredentialTypeBasic       CredentialType = "basic"
	CredentialTypeJWT         CredentialType = "jwt"
	CredentialTypeOAuth2      CredentialType = "oauth2"
	CredentialTypeCertificate CredentialType = "certificate"
)

// IsExpired checks if the credentials are expired.
func (c *Credentials) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// IsValid checks if the credentials are valid.
func (c *Credentials) IsValid() bool {
	if c.Type == CredentialTypeNone {
		return true
	}

	if c.IsExpired() {
		return false
	}

	switch c.Type {
	case CredentialTypeAPIKey:
		return c.APIKey != ""
	case CredentialTypeBearer, CredentialTypeOAuth2:
		return c.AccessToken != ""
	case CredentialTypeBasic:
		return c.Username != "" && c.Password != ""
	case CredentialTypeJWT:
		return c.AccessToken != ""
	case CredentialTypeCertificate:
		return c.Certificate != "" && c.PrivateKey != ""
	default:
		return false
	}
}

// ToAuthHeader converts credentials to an Authorization header value.
func (c *Credentials) ToAuthHeader() (string, error) {
	if !c.IsValid() {
		return "", fmt.Errorf("credentials are not valid")
	}

	switch c.Type {
	case CredentialTypeBearer, CredentialTypeOAuth2:
		return fmt.Sprintf("Bearer %s", c.AccessToken), nil
	case CredentialTypeJWT:
		return fmt.Sprintf("Bearer %s", c.AccessToken), nil
	case CredentialTypeBasic:
		// Basic auth would need base64 encoding
		return "", fmt.Errorf("basic auth not implemented in ToAuthHeader")
	default:
		return "", fmt.Errorf("unsupported credential type for auth header: %s", c.Type)
	}
}

// JWTCredentials represents JWT-specific credentials.
type JWTCredentials struct {
	*Credentials
	Token jwt.Token
}

// NewJWTCredentials creates new JWT credentials from a JWT token.
func NewJWTCredentials(token jwt.Token) (*JWTCredentials, error) {
	// Extract token string - this is a placeholder implementation
	// In a real implementation, you would serialize the token properly
	tokenString := "placeholder_token_string"

	// Get expiration time
	var expiresAt *time.Time
	if exp, ok := token.Expiration(); ok && !exp.IsZero() {
		expiresAt = &exp
	}

	creds := &Credentials{
		Type:        CredentialTypeJWT,
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
	}

	return &JWTCredentials{
		Credentials: creds,
		Token:       token,
	}, nil
}

// ParseJWTCredentials parses JWT credentials from a token string.
func ParseJWTCredentials(tokenString string) (*JWTCredentials, error) {
	token, err := jwt.Parse([]byte(tokenString), jwt.WithValidate(false))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	// Get expiration time
	var expiresAt *time.Time
	if exp, ok := token.Expiration(); ok && !exp.IsZero() {
		expiresAt = &exp
	}

	creds := &Credentials{
		Type:        CredentialTypeJWT,
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
	}

	return &JWTCredentials{
		Credentials: creds,
		Token:       token,
	}, nil
}

// ValidateJWT validates a JWT token.
func ValidateJWT(tokenString string, keyFunc func(token jwt.Token) (any, error)) error {
	token, err := jwt.Parse([]byte(tokenString), jwt.WithValidate(true))
	if err != nil {
		return fmt.Errorf("failed to parse and validate JWT token: %w", err)
	}

	// Additional validation can be added here
	if exp, ok := token.Expiration(); ok && exp.Before(time.Now()) {
		return fmt.Errorf("JWT token is expired")
	}

	return nil
}

// CreateAPIKeyCredentials creates API key credentials.
func CreateAPIKeyCredentials(apiKey string) *Credentials {
	return &Credentials{
		Type:   CredentialTypeAPIKey,
		APIKey: apiKey,
	}
}

// CreateBearerCredentials creates bearer token credentials.
func CreateBearerCredentials(accessToken string, expiresAt *time.Time) *Credentials {
	return &Credentials{
		Type:        CredentialTypeBearer,
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
	}
}

// CreateBasicCredentials creates basic authentication credentials.
func CreateBasicCredentials(username, password string) *Credentials {
	return &Credentials{
		Type:     CredentialTypeBasic,
		Username: username,
		Password: password,
	}
}

// CreateOAuth2Credentials creates OAuth2 credentials.
func CreateOAuth2Credentials(accessToken, refreshToken string, expiresAt *time.Time, scope string) *Credentials {
	return &Credentials{
		Type:         CredentialTypeOAuth2,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    expiresAt,
		Scope:        scope,
	}
}
