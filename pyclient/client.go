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

package pyclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/internal/jsonrpc2"
	"github.com/go-a2a/a2a-go/transport"
	"github.com/go-json-experiment/json"
)

// Client is the main A2A client that provides high-level methods for
// interacting with A2A agents. It manages connections, handles retries,
// and provides a Python-client-like API adapted to Go idioms.
type Client struct {
	opts        *options
	cm          *connectionManager
	agentCard   *a2a.AgentCard
	agentCardMu sync.RWMutex
	closed      bool
	closeMu     sync.Mutex
}

// New creates a new A2A client with the given options.
func New(opts ...Option) (*Client, error) {
	// Start with defaults
	o := defaultOptions()

	// Apply options
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	// Validate required fields
	if o.baseURL == "" {
		return nil, &ValidationError{Field: "baseURL", Message: "base URL is required"}
	}

	// Create client
	c := &Client{
		opts: o,
	}

	// Configure HTTP client if needed
	if o.httpClient == http.DefaultClient {
		// Create custom HTTP client with our settings
		o.httpClient = &http.Client{
			Timeout: o.timeout,
			Transport: &http.Transport{
				MaxIdleConns:       o.maxIdleConns,
				MaxConnsPerHost:    o.maxConnsPerHost,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: !o.enableCompression,
			},
		}
	}

	// Add retry interceptor if configured
	if o.retryConfig != nil && o.retryConfig.MaxAttempts > 0 {
		o.interceptors = append([]transport.Interceptor{retryInterceptor(o.retryConfig)}, o.interceptors...)
	}

	// Initialize transport if not provided
	if o.transport == nil {
		// Create HTTP+JSON transport
		httpTransport := &httpJSONTransport{
			client:       c,
			baseURL:      o.baseURL,
			httpClient:   o.httpClient,
			interceptors: o.interceptors,
		}
		o.transport = httpTransport
	}

	// Create connection manager
	c.cm = newConnectionManager(c, o.transport)

	return c, nil
}

// Connect establishes a connection to the A2A agent.
func (c *Client) Connect(ctx context.Context) error {
	c.closeMu.Lock()
	if c.closed {
		c.closeMu.Unlock()
		return ErrClientClosed
	}
	c.closeMu.Unlock()

	// Fetch agent card first
	if err := c.fetchAgentCard(ctx); err != nil {
		return fmt.Errorf("fetch agent card: %w", err)
	}

	// Connect using connection manager
	return c.cm.connect(ctx)
}

// Close closes the client and all associated resources.
func (c *Client) Close() error {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	// Disconnect
	if c.cm != nil {
		return c.cm.disconnect()
	}

	return nil
}

// State returns the current connection state.
func (c *Client) State() ConnectionState {
	if c.cm == nil {
		return ConnectionStateDisconnected
	}
	return c.cm.getState()
}

// WaitForState waits for the client to reach a specific connection state.
func (c *Client) WaitForState(ctx context.Context, state ConnectionState) error {
	if c.cm == nil {
		return ErrNotConnected
	}
	return c.cm.waitForState(ctx, state)
}

// AgentCard returns the cached agent card.
func (c *Client) AgentCard() *a2a.AgentCard {
	c.agentCardMu.RLock()
	defer c.agentCardMu.RUnlock()
	return c.agentCard
}

// fetchAgentCard fetches and caches the agent card.
func (c *Client) fetchAgentCard(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.opts.baseURL+a2a.AgentCardWellKnownPath, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Apply interceptors
	invoker := func(ctx context.Context, req *http.Request) (*http.Response, error) {
		return c.opts.httpClient.Do(req.WithContext(ctx))
	}
	invoker = chainInterceptors(c.opts.interceptors, invoker)

	resp, err := invoker(ctx, req)
	if err != nil {
		return fmt.Errorf("fetch agent card: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("fetch agent card: HTTP %d: %s", resp.StatusCode, body)
	}

	var card a2a.AgentCard
	if err := json.UnmarshalRead(resp.Body, &card); err != nil {
		return fmt.Errorf("decode agent card: %w", err)
	}

	c.agentCardMu.Lock()
	c.agentCard = &card
	c.agentCardMu.Unlock()

	return nil
}

// SendMessage sends a message to the agent and waits for a response.
func (c *Client) SendMessage(ctx context.Context, params *a2a.MessageSendParams) (a2a.MessageOrTask, error) {
	session, err := c.cm.getSession()
	if err != nil {
		return nil, err
	}

	// Log outgoing message if callback is set
	if c.opts.onMessage != nil {
		c.opts.onMessage("send", params)
	}

	return session.SendMessage(ctx, params)
}

// SendStreamMessage sends a message and returns a stream for real-time updates.
func (c *Client) SendStreamMessage(ctx context.Context, params *a2a.MessageSendParams) (*Stream, error) {
	// Check if streaming is supported
	card := c.AgentCard()
	if card == nil {
		return nil, ErrNoAgentCard
	}
	if card.Capabilities == nil || !card.Capabilities.Streaming {
		return nil, NewError(ErrCodeUnsupportedOperation, "agent does not support streaming")
	}

	// Build streaming request
	url := c.opts.baseURL + "/message/stream"
	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// Apply authentication
	if err := c.applyAuth(req); err != nil {
		return nil, fmt.Errorf("apply auth: %w", err)
	}

	// Apply interceptors
	invoker := func(ctx context.Context, req *http.Request) (*http.Response, error) {
		return c.opts.httpClient.Do(req.WithContext(ctx))
	}
	invoker = chainInterceptors(c.opts.interceptors, invoker)

	resp, err := invoker(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("send stream request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stream request failed: HTTP %d: %s", resp.StatusCode, body)
	}

	// Extract task ID from response headers or initial data
	taskID := resp.Header.Get("X-Task-ID")
	if taskID == "" {
		// Try to extract from first event
		// For now, use a placeholder
		taskID = "unknown"
	}

	return newStream(c, taskID, resp), nil
}

// GetTask retrieves information about a task.
func (c *Client) GetTask(ctx context.Context, params *a2a.TaskQueryParams) (*a2a.Task, error) {
	session, err := c.cm.getSession()
	if err != nil {
		return nil, err
	}

	return session.GetTask(ctx, params)
}

// CancelTask requests cancellation of a task.
func (c *Client) CancelTask(ctx context.Context, params *a2a.TaskIDParams) (*a2a.Task, error) {
	session, err := c.cm.getSession()
	if err != nil {
		return nil, err
	}

	return session.CancelTask(ctx, params)
}

// ResubscribeTask resubscribes to a task's event stream.
func (c *Client) ResubscribeTask(ctx context.Context, params *a2a.TaskIDParams) (*Stream, error) {
	// Check if streaming is supported
	card := c.AgentCard()
	if card == nil {
		return nil, ErrNoAgentCard
	}
	if card.Capabilities == nil || !card.Capabilities.Streaming {
		return nil, NewError(ErrCodeUnsupportedOperation, "agent does not support streaming")
	}

	// Build resubscribe request
	url := c.opts.baseURL + "/tasks/resubscribe"
	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// Apply authentication
	if err := c.applyAuth(req); err != nil {
		return nil, fmt.Errorf("apply auth: %w", err)
	}

	// Apply interceptors
	invoker := func(ctx context.Context, req *http.Request) (*http.Response, error) {
		return c.opts.httpClient.Do(req.WithContext(ctx))
	}
	invoker = chainInterceptors(c.opts.interceptors, invoker)

	resp, err := invoker(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("resubscribe request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("resubscribe failed: HTTP %d: %s", resp.StatusCode, body)
	}

	return newStream(c, params.ID, resp), nil
}

// applyAuth applies authentication to a request.
func (c *Client) applyAuth(req *http.Request) error {
	// Use auth provider if available
	if c.opts.authProvider != nil {
		token, err := c.opts.authProvider.GetToken()
		if err != nil {
			return fmt.Errorf("get auth token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}

	// Use static token if configured
	if c.opts.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.opts.authToken)
		return nil
	}

	// No auth configured
	return nil
}

// chainInterceptors chains multiple interceptors together.
func chainInterceptors(interceptors []transport.Interceptor, invoker transport.Invoker) transport.Invoker {
	if len(interceptors) == 0 {
		return invoker
	}

	// Build the chain from right to left
	for i := len(interceptors) - 1; i >= 0; i-- {
		interceptor := interceptors[i]
		next := invoker
		invoker = func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return interceptor(ctx, req, next)
		}
	}

	return invoker
}

// httpJSONTransport implements transport.Transport for HTTP+JSON.
type httpJSONTransport struct {
	client       *Client
	baseURL      string
	httpClient   *http.Client
	interceptors []transport.Interceptor
}

// Connect implements transport.Transport.
func (t *httpJSONTransport) Connect(ctx context.Context) (transport.Connection, error) {
	// For HTTP+JSON, we don't maintain a persistent connection
	// Instead, we return a connection that makes HTTP requests
	return &httpJSONConnection{
		transport: t,
		sessionID: generateSessionID(),
	}, nil
}

// httpJSONConnection implements transport.Connection for HTTP+JSON.
type httpJSONConnection struct {
	transport *httpJSONTransport
	sessionID string
}

// SessionID implements transport.Connection.
func (c *httpJSONConnection) SessionID() string {
	return c.sessionID
}

// Read implements transport.Connection.
func (c *httpJSONConnection) Read(ctx context.Context) (jsonrpc2.Message, error) {
	// HTTP+JSON doesn't support server-initiated messages
	return nil, fmt.Errorf("HTTP+JSON transport does not support server-initiated messages")
}

// Write implements transport.Connection.
func (c *httpJSONConnection) Write(ctx context.Context, msg jsonrpc2.Message) error {
	// Convert JSON-RPC message to HTTP request
	// This would be implemented based on the A2A HTTP+JSON protocol spec
	return fmt.Errorf("HTTP+JSON write not yet implemented")
}

// Close implements transport.Connection.
func (c *httpJSONConnection) Close() error {
	// Nothing to close for HTTP+JSON
	return nil
}

// generateSessionID generates a unique session ID.
func generateSessionID() string {
	// Simple implementation, should use UUID in production
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}
