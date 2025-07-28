// Copyright 2025 The Go A2A Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"

	a2a "github.com/go-a2a/a2a-go"
)

// CardResolver defines the interface for resolving agent cards.
type CardResolver interface {
	GetAgentCard(ctx context.Context, relativePath string, opts ...RequestOption) (*a2a.AgentCard, error)
}

// cardResolver implements CardResolver using HTTP requests.
type cardResolver struct {
	httpClient    *http.Client
	baseURL       string
	agentCardPath string
	interceptors  []Interceptor
}

var _ CardResolver = (*cardResolver)(nil)

// NewCardResolver creates a new HTTP-based card resolver.
func NewCardResolver(baseURL string, opts ...ClientOption) *cardResolver {
	config := applyClientOptions(opts...)

	resolver := &cardResolver{
		httpClient:    config.httpClient,
		baseURL:       strings.TrimRight(baseURL, "/"),
		agentCardPath: strings.TrimLeft(a2a.AgentCardWellKnownPath, "/"),
		interceptors:  config.interceptors,
	}

	return resolver
}

// GetAgentCard fetches an agent card from a specified path relative to the baseURL.
//
// If relative_card_path is None, it defaults to the resolver's configured
// agent_card_path (for the public agent card).
func (r *cardResolver) GetAgentCard(ctx context.Context, relativeCardPath string, opts ...RequestOption) (*a2a.AgentCard, error) {
	if relativeCardPath == "" {
		relativeCardPath = r.agentCardPath
	}
	// Ensure the path starts with a slash
	if !strings.HasPrefix(relativeCardPath, "/") {
		relativeCardPath = "/" + relativeCardPath
	}

	// Remove leading slash from relative path if present
	relativeCardPath = strings.TrimLeft(relativeCardPath, "/")

	targetURL := fmt.Sprintf("%s/%s", r.baseURL, relativeCardPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, NewConfigurationError("create request", err)
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
	invoker = chainInterceptors(r.interceptors, invoker)

	resp, err := invoker(ctx, req)
	if err != nil {
		return nil, NewNetworkError("fetch agent card", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHTTPError(resp.StatusCode, fmt.Sprintf("fetch agent card from %s", targetURL), nil)
	}

	var agentCard a2a.AgentCard
	dec := jsontext.NewDecoder(resp.Body)
	if err := json.UnmarshalDecode(dec, &agentCard, json.DefaultOptionsV2()); err != nil {
		return nil, NewJSONError("decode agent card", err)
	}

	return &agentCard, nil
}

// GetPublicAgentCard fetches the public agent card using the default path.
func (r *cardResolver) GetPublicAgentCard(ctx context.Context, opts ...RequestOption) (*a2a.AgentCard, error) {
	return r.GetAgentCard(ctx, "", opts...)
}

// CreateClientFromAgentCardURL creates a client from an agent card URL.
func CreateClientFromAgentCardURL(ctx context.Context, baseURL string, opts ...ClientOption) (Client, error) {
	resolver := NewCardResolver(baseURL, opts...)

	agentCard, err := resolver.GetPublicAgentCard(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch agent card: %w", err)
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

// Client defines the interface for A2A clients.
type Client interface {
	// SendMessage sends a non-streaming message request to the agent.
	SendMessage(ctx context.Context, req *SendMessageRequest, opts ...RequestOption) (*SendMessageResponse, error)

	// SendMessageStreaming sends a streaming message request to the agent.
	SendMessageStreaming(ctx context.Context, req *SendStreamingMessageRequest, opts ...RequestOption) (<-chan *SendStreamingMessageResponse, error)

	// GetTask retrieves the current state and history of a specific task.
	GetTask(ctx context.Context, req *GetTaskRequest, opts ...RequestOption) (*GetTaskResponse, error)

	// CancelTask requests the agent to cancel a specific task.
	CancelTask(ctx context.Context, req *CancelTaskRequest, opts ...RequestOption) (*CancelTaskResponse, error)

	// SetTaskCallback sets or updates the push notification configuration for a specific task.
	SetTaskCallback(ctx context.Context, req *SetTaskPushNotificationConfigRequest, opts ...RequestOption) (*SetTaskPushNotificationConfigResponse, error)

	// GetTaskCallback retrieves the push notification configuration for a specific task.
	GetTaskCallback(ctx context.Context, req *GetTaskPushNotificationConfigRequest, opts ...RequestOption) (*GetTaskPushNotificationConfigResponse, error)

	// Close closes the client and releases any resources.
	Close() error
}

// HTTPClient implements the Client interface using HTTP requests.
type HTTPClient struct {
	httpClient   *http.Client
	agentCard    *a2a.AgentCard
	url          string
	interceptors []Interceptor
	userAgent    string
}

var _ Client = (*HTTPClient)(nil)

// NewHTTPClient creates a new HTTP client.
func NewHTTPClient(opts ...ClientOption) *HTTPClient {
	config := applyClientOptions(opts...)

	client := &HTTPClient{
		httpClient:   config.httpClient,
		interceptors: config.interceptors,
		userAgent:    config.userAgent,
	}

	return client
}

// NewHTTPClientWithURL creates a new HTTP client with a direct URL.
func NewHTTPClientWithURL(url string, opts ...ClientOption) *HTTPClient {
	client := NewHTTPClient(opts...)
	client.url = url
	return client
}

// NewHTTPClientWithAgentCard creates a new HTTP client with an agent card.
func NewHTTPClientWithAgentCard(agentCard *a2a.AgentCard, opts ...ClientOption) *HTTPClient {
	client := NewHTTPClient(opts...)
	client.agentCard = agentCard
	client.url = agentCard.URL
	return client
}

// func (c *HTTPClient) GetClientFromAgentCardURL(httpClient *http.Client, baseURL, agentCardPath string) Client {
// 	client := NewHTTPClient()
// 	agentCard := NewCardResolver(baseURL)
// 	client.
// }

// SendMessage sends a non-streaming message request to the agent.
func (c *HTTPClient) SendMessage(ctx context.Context, req *SendMessageRequest, opts ...RequestOption) (*SendMessageResponse, error) {
	// Convert to JSON-RPC format
	rpcReq := req.toJSONRPC()

	payload, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, NewJSONError("marshal request", err)
	}

	respData, err := c.sendRequest(ctx, payload, opts...)
	if err != nil {
		return nil, err
	}

	var response SendMessageResponse
	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, NewJSONError("unmarshal response", err)
	}

	return &response, nil
}

// SendMessageStreaming sends a streaming message request to the agent.
func (c *HTTPClient) SendMessageStreaming(ctx context.Context, req *SendStreamingMessageRequest, opts ...RequestOption) (<-chan *SendStreamingMessageResponse, error) {
	req.ensureID()

	// Convert to JSON-RPC format (similar to SendMessage)
	rpcReq := &a2a.SendMessageRequest{
		Method: "message/stream",
		Params: &a2a.MessageSendParams{
			Message: req.Message,
		},
		ID: req.ID,
	}

	payload, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, NewJSONError("marshal request", err)
	}

	return c.sendStreamingRequest(ctx, payload, opts...)
}

// GetTask retrieves the current state and history of a specific task.
func (c *HTTPClient) GetTask(ctx context.Context, req *GetTaskRequest, opts ...RequestOption) (*GetTaskResponse, error) {
	req.ensureID()

	// Convert to JSON-RPC format
	rpcReq := req.toJSONRPC()

	payload, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, NewJSONError("marshal request", err)
	}

	respData, err := c.sendRequest(ctx, payload, opts...)
	if err != nil {
		return nil, err
	}

	var response GetTaskResponse
	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, NewJSONError("unmarshal response", err)
	}

	return &response, nil
}

// CancelTask requests the agent to cancel a specific task.
func (c *HTTPClient) CancelTask(ctx context.Context, req *CancelTaskRequest, opts ...RequestOption) (*CancelTaskResponse, error) {
	req.ensureID()

	// Convert to JSON-RPC format
	rpcReq := req.toJSONRPC()

	payload, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, NewJSONError("marshal request", err)
	}

	respData, err := c.sendRequest(ctx, payload, opts...)
	if err != nil {
		return nil, err
	}

	var response CancelTaskResponse
	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, NewJSONError("unmarshal response", err)
	}

	return &response, nil
}

// SetTaskCallback sets or updates the push notification configuration for a specific task.
func (c *HTTPClient) SetTaskCallback(ctx context.Context, req *SetTaskPushNotificationConfigRequest, opts ...RequestOption) (*SetTaskPushNotificationConfigResponse, error) {
	req.ensureID()

	// Create JSON-RPC request
	rpcReq := map[string]any{
		"method": "tasks/pushNotificationConfig/set",
		"params": map[string]any{
			"task_id": req.TaskID,
			"config":  req.Config,
		},
		"id": req.ID,
	}

	payload, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, NewJSONError("marshal request", err)
	}

	respData, err := c.sendRequest(ctx, payload, opts...)
	if err != nil {
		return nil, err
	}

	var response SetTaskPushNotificationConfigResponse
	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, NewJSONError("unmarshal response", err)
	}

	return &response, nil
}

// GetTaskCallback retrieves the push notification configuration for a specific task.
func (c *HTTPClient) GetTaskCallback(ctx context.Context, req *GetTaskPushNotificationConfigRequest, opts ...RequestOption) (*GetTaskPushNotificationConfigResponse, error) {
	req.ensureID()

	// Create JSON-RPC request
	rpcReq := map[string]any{
		"method": "tasks/pushNotificationConfig/get",
		"params": map[string]any{
			"task_id": req.TaskID,
		},
		"id": req.ID,
	}

	payload, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, NewJSONError("marshal request", err)
	}

	respData, err := c.sendRequest(ctx, payload, opts...)
	if err != nil {
		return nil, err
	}

	var response GetTaskPushNotificationConfigResponse
	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, NewJSONError("unmarshal response", err)
	}

	return &response, nil
}

// Close closes the client and releases any resources.
func (c *HTTPClient) Close() error {
	// For HTTP client, there's nothing to close
	return nil
}

// sendRequest sends a non-streaming JSON-RPC request to the agent.
func (c *HTTPClient) sendRequest(ctx context.Context, payload []byte, opts ...RequestOption) ([]byte, error) {
	if c.url == "" {
		return nil, NewConfigurationError("client URL is not set", nil)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(payload))
	if err != nil {
		return nil, NewConfigurationError("create request", err)
	}

	// Apply request options
	requestConfig := applyRequestOptions(opts...)
	for key, value := range requestConfig.headers {
		req.Header.Set(key, value)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	// Apply interceptors
	invoker := func(ctx context.Context, req *http.Request) (*http.Response, error) {
		return c.httpClient.Do(req)
	}

	invoker = chainInterceptors(c.interceptors, invoker)

	resp, err := invoker(ctx, req)
	if err != nil {
		return nil, NewNetworkError("send request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHTTPError(resp.StatusCode, fmt.Sprintf("request failed with status %d", resp.StatusCode), nil)
	}

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewNetworkError("read response body", err)
	}

	return respData, nil
}

// sendStreamingRequest sends a streaming JSON-RPC request to the agent.
func (c *HTTPClient) sendStreamingRequest(ctx context.Context, payload []byte, opts ...RequestOption) (<-chan *SendStreamingMessageResponse, error) {
	if c.url == "" {
		return nil, NewConfigurationError("client URL is not set", nil)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(payload))
	if err != nil {
		return nil, NewConfigurationError("create request", err)
	}

	// Apply request options
	requestConfig := applyRequestOptions(opts...)
	for key, value := range requestConfig.headers {
		req.Header.Set(key, value)
	}

	// Set default headers for SSE
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Cache-Control", "no-cache")

	// Apply interceptors
	invoker := func(ctx context.Context, req *http.Request) (*http.Response, error) {
		return c.httpClient.Do(req)
	}
	invoker = chainInterceptors(c.interceptors, invoker)

	resp, err := invoker(ctx, req)
	if err != nil {
		return nil, NewNetworkError("send streaming request", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, NewHTTPError(resp.StatusCode, fmt.Sprintf("streaming request failed with status %d", resp.StatusCode), nil)
	}

	// Create a channel for streaming responses
	ch := make(chan *SendStreamingMessageResponse, 10)

	go func() {
		defer resp.Body.Close()
		defer close(ch)

		if err := c.parseSSEStream(ctx, resp.Body, ch); err != nil {
			// Send error as a response
			select {
			case ch <- &SendStreamingMessageResponse{
				Error: &RPCError{
					Code:    -32603,
					Message: fmt.Sprintf("streaming error: %v", err),
				},
			}:
			case <-ctx.Done():
			}
		}
	}()

	return ch, nil
}

// parseSSEStream parses Server-Sent Events stream.
func (c *HTTPClient) parseSSEStream(ctx context.Context, reader io.Reader, ch chan<- *SendStreamingMessageResponse) error {
	// This is a simplified SSE parser
	// In a real implementation, we'd use a proper SSE library
	dec := jsontext.NewDecoder(reader)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var response SendStreamingMessageResponse
			if err := json.UnmarshalDecode(dec, &response, json.DefaultOptionsV2()); err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("decode SSE event: %w", err)
			}

			select {
			case ch <- &response:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
