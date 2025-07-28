// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
)

// AuthInterceptor provides authentication interceptor functionality.
type AuthInterceptor struct {
	credentialService  CredentialService
	contextIDExtractor func(context.Context) string
}

// NewAuthInterceptor creates a new authentication interceptor.
func NewAuthInterceptor(credentialService CredentialService) *AuthInterceptor {
	return &AuthInterceptor{
		credentialService:  credentialService,
		contextIDExtractor: defaultContextIDExtractor,
	}
}

// WithContextIDExtractor sets a custom context ID extractor.
func (a *AuthInterceptor) WithContextIDExtractor(extractor func(context.Context) string) *AuthInterceptor {
	a.contextIDExtractor = extractor
	return a
}

// Intercept implements the interceptor interface.
func (a *AuthInterceptor) Intercept(ctx context.Context, req *http.Request, invoker func(context.Context, *http.Request) (*http.Response, error)) (*http.Response, error) {
	// Extract context ID
	contextID := a.contextIDExtractor(ctx)
	if contextID == "" {
		// No context ID, proceed without authentication
		return invoker(ctx, req)
	}

	// Get credentials
	credentials, err := a.credentialService.GetCredentials(ctx, contextID)
	if err != nil {
		// No credentials found, proceed without authentication
		return invoker(ctx, req)
	}

	// Validate credentials
	if err := a.credentialService.ValidateCredentials(ctx, credentials); err != nil {
		// Invalid credentials, proceed without authentication
		return invoker(ctx, req)
	}

	// Apply authentication to request
	if err := a.applyAuthentication(req, credentials); err != nil {
		return nil, fmt.Errorf("failed to apply authentication: %w", err)
	}

	return invoker(ctx, req)
}

// applyAuthentication applies authentication to the HTTP request.
func (a *AuthInterceptor) applyAuthentication(req *http.Request, credentials *Credentials) error {
	switch credentials.Type {
	case CredentialTypeAPIKey:
		return a.applyAPIKeyAuth(req, credentials)
	case CredentialTypeBearer, CredentialTypeJWT, CredentialTypeOAuth2:
		return a.applyBearerAuth(req, credentials)
	case CredentialTypeBasic:
		return a.applyBasicAuth(req, credentials)
	case CredentialTypeNone:
		return nil
	default:
		return fmt.Errorf("unsupported credential type: %s", credentials.Type)
	}
}

// applyAPIKeyAuth applies API key authentication.
func (a *AuthInterceptor) applyAPIKeyAuth(req *http.Request, credentials *Credentials) error {
	if credentials.APIKey == "" {
		return fmt.Errorf("API key is empty")
	}

	// Default to header-based API key
	req.Header.Set("X-API-Key", credentials.APIKey)
	return nil
}

// applyBearerAuth applies Bearer token authentication.
func (a *AuthInterceptor) applyBearerAuth(req *http.Request, credentials *Credentials) error {
	if credentials.AccessToken == "" {
		return fmt.Errorf("access token is empty")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", credentials.AccessToken))
	return nil
}

// applyBasicAuth applies Basic authentication.
func (a *AuthInterceptor) applyBasicAuth(req *http.Request, credentials *Credentials) error {
	if credentials.Username == "" || credentials.Password == "" {
		return fmt.Errorf("username or password is empty")
	}

	auth := credentials.Username + ":" + credentials.Password
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", encoded))
	return nil
}

// defaultContextIDExtractor extracts context ID from the request context.
func defaultContextIDExtractor(ctx context.Context) string {
	// Try to extract from context values
	if contextID, ok := ctx.Value(contextIDContextKey).(string); ok {
		return contextID
	}

	// Try to extract from client call context
	if callCtx, ok := ctx.Value(clientCallContextKey("client_call_context")).(*ClientCallContext); ok {
		if contextID, ok := callCtx.Metadata["context_id"].(string); ok {
			return contextID
		}
	}

	return ""
}

// contextIDKey is used for context values.
type contextIDKey string

// clientCallContextKey is used for client call context.
type clientCallContextKey string

// Context keys
const (
	contextIDContextKey contextIDKey = "context_id"
)

// ClientCallContext represents the client call context (mirrors the one in parent package).
type ClientCallContext struct {
	Method    string
	URL       string
	Headers   map[string]string
	Metadata  map[string]any
	AgentCard any
}

// WithContextID adds a context ID to the context.
func WithContextID(ctx context.Context, contextID string) context.Context {
	return context.WithValue(ctx, contextIDContextKey, contextID)
}

// GetContextID retrieves the context ID from the context.
func GetContextID(ctx context.Context) (string, bool) {
	contextID, ok := ctx.Value(contextIDContextKey).(string)
	return contextID, ok
}

// StaticCredentialInterceptor provides static credential authentication.
type StaticCredentialInterceptor struct {
	credentials *Credentials
}

// NewStaticCredentialInterceptor creates a new static credential interceptor.
func NewStaticCredentialInterceptor(credentials *Credentials) *StaticCredentialInterceptor {
	return &StaticCredentialInterceptor{
		credentials: credentials,
	}
}

// Intercept implements the interceptor interface.
func (s *StaticCredentialInterceptor) Intercept(ctx context.Context, req *http.Request, invoker func(context.Context, *http.Request) (*http.Response, error)) (*http.Response, error) {
	if s.credentials == nil || !s.credentials.IsValid() {
		return invoker(ctx, req)
	}

	interceptor := &AuthInterceptor{
		credentialService: NewStaticCredentialService(s.credentials),
		contextIDExtractor: func(ctx context.Context) string {
			return "static"
		},
	}

	return interceptor.Intercept(ctx, req, invoker)
}

// StaticCredentialService provides static credentials.
type StaticCredentialService struct {
	credentials *Credentials
}

var _ CredentialService = (*StaticCredentialService)(nil)

// NewStaticCredentialService creates a new static credential service.
func NewStaticCredentialService(credentials *Credentials) *StaticCredentialService {
	return &StaticCredentialService{
		credentials: credentials,
	}
}

// GetCredentials returns the static credentials.
func (s *StaticCredentialService) GetCredentials(ctx context.Context, contextID string) (*Credentials, error) {
	if s.credentials == nil {
		return nil, fmt.Errorf("no static credentials configured")
	}
	return s.credentials, nil
}

// StoreCredentials does nothing for static credentials.
func (s *StaticCredentialService) StoreCredentials(ctx context.Context, contextID string, credentials *Credentials) error {
	return fmt.Errorf("cannot store credentials in static credential service")
}

// DeleteCredentials does nothing for static credentials.
func (s *StaticCredentialService) DeleteCredentials(ctx context.Context, contextID string) error {
	return fmt.Errorf("cannot delete credentials in static credential service")
}

// ValidateCredentials validates the static credentials.
func (s *StaticCredentialService) ValidateCredentials(ctx context.Context, credentials *Credentials) error {
	if credentials == nil {
		return fmt.Errorf("credentials cannot be nil")
	}

	if !credentials.IsValid() {
		return fmt.Errorf("credentials are not valid")
	}

	return nil
}

// MultiAuthInterceptor chains multiple authentication interceptors.
type MultiAuthInterceptor struct {
	interceptors []func(context.Context, *http.Request, func(context.Context, *http.Request) (*http.Response, error)) (*http.Response, error)
}

// NewMultiAuthInterceptor creates a new multi-auth interceptor.
func NewMultiAuthInterceptor(interceptors ...func(context.Context, *http.Request, func(context.Context, *http.Request) (*http.Response, error)) (*http.Response, error)) *MultiAuthInterceptor {
	return &MultiAuthInterceptor{
		interceptors: interceptors,
	}
}

// Intercept implements the interceptor interface by chaining multiple interceptors.
func (m *MultiAuthInterceptor) Intercept(ctx context.Context, req *http.Request, invoker func(context.Context, *http.Request) (*http.Response, error)) (*http.Response, error) {
	// Chain interceptors from right to left
	currentInvoker := invoker
	for i := len(m.interceptors) - 1; i >= 0; i-- {
		interceptor := m.interceptors[i]
		nextInvoker := currentInvoker
		currentInvoker = func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return interceptor(ctx, req, nextInvoker)
		}
	}

	return currentInvoker(ctx, req)
}
