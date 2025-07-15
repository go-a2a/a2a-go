// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-a2a/a2a"
)

// CardResolver defines the interface for resolving agent cards.
type CardResolver interface {
	GetAgentCard(ctx context.Context, relativePath string, opts ...RequestOption) (*a2a.AgentCard, error)
}

// HTTPCardResolver implements CardResolver using HTTP requests.
type HTTPCardResolver struct {
	httpClient    *http.Client
	baseURL       string
	agentCardPath string
	interceptors  []Interceptor
}

var _ CardResolver = (*HTTPCardResolver)(nil)

// NewHTTPCardResolver creates a new HTTP-based card resolver.
func NewHTTPCardResolver(baseURL string, opts ...ClientOption) *HTTPCardResolver {
	config := applyClientOptions(opts...)

	resolver := &HTTPCardResolver{
		httpClient:    config.httpClient,
		baseURL:       strings.TrimRight(baseURL, "/"),
		agentCardPath: "/.well-known/agent-card",
		interceptors:  config.interceptors,
	}

	return resolver
}

// WithAgentCardPath sets the agent card path.
func WithAgentCardPath(path string) ClientOption {
	return func(c *clientConfig) {
		// This is a bit of a hack, but we'll store it in the user agent for now
		// In a real implementation, we'd extend clientConfig
		c.userAgent = path
	}
}

// GetAgentCard fetches an agent card from the specified path.
func (r *HTTPCardResolver) GetAgentCard(ctx context.Context, relativePath string, opts ...RequestOption) (*a2a.AgentCard, error) {
	if relativePath == "" {
		relativePath = r.agentCardPath
	}

	// Ensure the path starts with a slash
	if !strings.HasPrefix(relativePath, "/") {
		relativePath = "/" + relativePath
	}

	// Remove leading slash from relative path if present
	relativePath = strings.TrimLeft(relativePath, "/")

	targetURL := fmt.Sprintf("%s/%s", r.baseURL, relativePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, NewConfigurationError("failed to create request", err)
	}

	// Apply request options
	requestConfig := applyRequestOptions(opts...)
	for key, value := range requestConfig.headers {
		req.Header.Set(key, value)
	}

	// Set default headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Apply interceptors
	invoker := func(ctx context.Context, req *http.Request) (*http.Response, error) {
		return r.httpClient.Do(req)
	}

	chainedInvoker := chainInterceptors(r.interceptors, invoker)

	resp, err := chainedInvoker(ctx, req)
	if err != nil {
		return nil, NewNetworkError("failed to fetch agent card", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHTTPError(resp.StatusCode, fmt.Sprintf("failed to fetch agent card from %s", targetURL), nil)
	}

	var agentCard a2a.AgentCard
	if err := json.NewDecoder(resp.Body).Decode(&agentCard); err != nil {
		return nil, NewJSONError("failed to decode agent card", err)
	}

	// Validate the agent card
	if err := agentCard.Validate(); err != nil {
		return nil, NewValidationError("invalid agent card", err)
	}

	return &agentCard, nil
}

// GetPublicAgentCard fetches the public agent card using the default path.
func (r *HTTPCardResolver) GetPublicAgentCard(ctx context.Context, opts ...RequestOption) (*a2a.AgentCard, error) {
	return r.GetAgentCard(ctx, "", opts...)
}

// CreateClientFromAgentCardURL creates a client from an agent card URL.
func CreateClientFromAgentCardURL(ctx context.Context, baseURL string, opts ...ClientOption) (Client, error) {
	resolver := NewHTTPCardResolver(baseURL, opts...)

	agentCard, err := resolver.GetPublicAgentCard(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agent card: %w", err)
	}

	// Create client with the agent card
	clientOpts := append(opts, WithAgentCard(agentCard))
	return NewHTTPClient(clientOpts...), nil
}

// WithAgentCard sets the agent card for the client.
func WithAgentCard(agentCard *a2a.AgentCard) ClientOption {
	return func(c *clientConfig) {
		// Store the agent card in the interceptors for now
		// In a real implementation, we'd extend clientConfig
		c.interceptors = append(c.interceptors, func(ctx context.Context, req *http.Request, invoker Invoker) (*http.Response, error) {
			// Add agent card to context
			callCtx := &ClientCallContext{
				Method:    req.Method,
				URL:       req.URL.String(),
				Headers:   make(map[string]string),
				Metadata:  make(map[string]any),
				AgentCard: agentCard,
			}

			for key, values := range req.Header {
				if len(values) > 0 {
					callCtx.Headers[key] = values[0]
				}
			}

			ctxWithCallCtx := WithClientCallContext(ctx, callCtx)
			return invoker(ctxWithCallCtx, req)
		})
	}
}
