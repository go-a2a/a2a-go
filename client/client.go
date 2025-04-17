// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package client provides a client implementation for the Google A2A protocol.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-a2a/a2a"
)

// Common errors that may occur during client operations
var (
	ErrInvalidResponse    = errors.New("invalid response from server")
	ErrRequestFailed      = errors.New("request failed")
	ErrConnectionFailed   = errors.New("connection failed")
	ErrTimeout            = errors.New("request timed out")
	ErrInvalidServerURL   = errors.New("invalid server URL")
	ErrTaskNotFound       = errors.New("task not found")
	ErrServerError        = errors.New("server error")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// A2AClient is a client for the A2A protocol
type A2AClient struct {
	// ServerURL is the URL of the A2A server
	ServerURL string

	// HTTPClient is the HTTP client to use for requests
	HTTPClient *http.Client

	// Timeout is the request timeout
	Timeout time.Duration

	// DefaultSessionID is the default session ID to use for requests
	DefaultSessionID string

	// AgentCard is the agent card of the server
	AgentCard *a2a.AgentCard

	// RequestHeaders are additional headers to include in requests
	RequestHeaders map[string]string
}

// ClientOptions contains options for creating a new A2A client
type ClientOptions struct {
	ServerURL      string
	HTTPClient     *http.Client
	Timeout        time.Duration
	DefaultSession string
	RequestHeaders map[string]string
}

// DefaultTimeout is the default request timeout
const DefaultTimeout = 30 * time.Second

// NewClient creates a new A2A client
func NewClient(serverURL string, opts ...func(*ClientOptions)) (*A2AClient, error) {
	if serverURL == "" {
		return nil, ErrInvalidServerURL
	}

	options := ClientOptions{
		ServerURL:  serverURL,
		Timeout:    DefaultTimeout,
		HTTPClient: http.DefaultClient,
	}

	// Apply options
	for _, opt := range opts {
		opt(&options)
	}

	client := &A2AClient{
		ServerURL:        serverURL,
		HTTPClient:       options.HTTPClient,
		Timeout:          options.Timeout,
		DefaultSessionID: options.DefaultSession,
		RequestHeaders:   options.RequestHeaders,
	}

	return client, nil
}

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.Timeout = timeout
	}
}

// WithHTTPClient sets the HTTP client
func WithHTTPClient(httpClient *http.Client) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.HTTPClient = httpClient
	}
}

// WithDefaultSession sets the default session ID
func WithDefaultSession(sessionID string) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.DefaultSession = sessionID
	}
}

// WithRequestHeaders sets additional request headers
func WithRequestHeaders(headers map[string]string) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.RequestHeaders = headers
	}
}

// makeRPCRequest sends a JSON-RPC request to the server
func (c *A2AClient) makeRPCRequest(ctx context.Context, method string, params any) (*a2a.JsonRpcResponse, error) {
	request := a2a.JsonRpcRequest{
		JsonRpc: "2.0",
		ID:      rand.Int(),
		Method:  method,
		Params:  params,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.ServerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for key, value := range c.RequestHeaders {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: %s", ErrRequestFailed, resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var response a2a.JsonRpcResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	// Check for JSON-RPC error
	if response.Error != nil {
		return &response, fmt.Errorf("%w: %s (code: %d)", ErrServerError, response.Error.Message, response.Error.Code)
	}

	return &response, nil
}

// FetchAgentCard retrieves the agent card from the server
func (c *A2AClient) FetchAgentCard(ctx context.Context) (*a2a.AgentCard, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	params := a2a.AgentCardRequest{}
	resp, err := c.makeRPCRequest(ctx, "agent/card", params)
	if err != nil {
		return nil, err
	}

	var card a2a.AgentCard
	b, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal response result", ErrInvalidResponse)
	}

	if err := json.Unmarshal(b, &card); err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal agent card", ErrInvalidResponse)
	}

	c.AgentCard = &card
	return &card, nil
}

// GetTask retrieves information about a task
func (c *A2AClient) GetTask(ctx context.Context, taskID string) (*a2a.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	params := a2a.TasksGetRequest{
		ID: taskID,
	}

	resp, err := c.makeRPCRequest(ctx, "tasks/get", params)
	if err != nil {
		return nil, err
	}

	var task a2a.Task
	b, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal response result", ErrInvalidResponse)
	}

	if err := json.Unmarshal(b, &task); err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal task", ErrInvalidResponse)
	}

	return &task, nil
}

// SendTask sends a new task or continues an existing task
func (c *A2AClient) SendTask(ctx context.Context, taskOpts TaskOptions) (*a2a.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	// Generate a task ID if not provided
	if taskOpts.ID == "" {
		taskOpts.ID = generateTaskID()
	}

	// Use default session ID if not provided
	if taskOpts.SessionID == "" {
		taskOpts.SessionID = c.DefaultSessionID
	}

	params := a2a.TasksSendRequest{
		ID:                  taskOpts.ID,
		SessionID:           taskOpts.SessionID,
		AcceptedOutputModes: taskOpts.AcceptedOutputModes,
		Message:             taskOpts.Message,
		ParentTask:          taskOpts.ParentTask,
		PrevTasks:           taskOpts.PrevTasks,
		Metadata:            taskOpts.Metadata,
	}

	resp, err := c.makeRPCRequest(ctx, "tasks/send", params)
	if err != nil {
		return nil, err
	}

	var task a2a.Task
	b, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal response result", ErrInvalidResponse)
	}

	if err := json.Unmarshal(b, &task); err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal task", ErrInvalidResponse)
	}

	return &task, nil
}

// CancelTask cancels a task
func (c *A2AClient) CancelTask(ctx context.Context, taskID string, reason string) error {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	params := a2a.TasksCancelRequest{
		ID:     taskID,
		Reason: reason,
	}

	_, err := c.makeRPCRequest(ctx, "tasks/cancel", params)
	return err
}

// SubscribeTask subscribes to updates for a task
func (c *A2AClient) SubscribeTask(ctx context.Context, taskID string) (<-chan *a2a.Task, error) {
	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(ctx)

	// Check if the server supports streaming
	if c.AgentCard == nil {
		if _, err := c.FetchAgentCard(ctx); err != nil {
			cancel()
			return nil, err
		}
	}

	if !c.AgentCard.Capabilities.Streaming {
		cancel()
		return nil, errors.New("server does not support streaming")
	}

	params := a2a.TasksSubscribeRequest{
		ID: taskID,
	}

	// Create a channel to send task updates
	taskCh := make(chan *a2a.Task)

	// Start a goroutine to handle the streaming response
	go func() {
		defer close(taskCh)
		defer cancel()

		// Create HTTP request
		jsonData, err := json.Marshal(a2a.JsonRpcRequest{
			JsonRpc: "2.0",
			ID:      rand.Int(),
			Method:  "tasks/subscribe",
			Params:  params,
		})
		if err != nil {
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", c.ServerURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		for key, value := range c.RequestHeaders {
			req.Header.Set(key, value)
		}

		// Send request
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return
		}

		// Read response line by line
		decoder := json.NewDecoder(resp.Body)
		for {
			// Check if context is done
			select {
			case <-ctx.Done():
				return
			default:
				// Continue processing
			}

			// Parse response
			var response a2a.JsonRpcResponse
			if err := decoder.Decode(&response); err != nil {
				if err == io.EOF {
					return
				}
				continue
			}

			// Check for JSON-RPC error
			if response.Error != nil {
				continue
			}

			// Parse task
			var task a2a.Task
			b, err := json.Marshal(response.Result)
			if err != nil {
				continue
			}

			if err := json.Unmarshal(b, &task); err != nil {
				continue
			}

			// Send task update
			select {
			case taskCh <- &task:
				// Sent successfully
			case <-ctx.Done():
				return
			}

			// If task is completed, failed, or canceled, stop streaming
			if task.Status.State == a2a.TaskCompleted ||
				task.Status.State == a2a.TaskFailed ||
				task.Status.State == a2a.TaskCanceled {
				return
			}
		}
	}()

	return taskCh, nil
}

// SendTaskAndSubscribe sends a task and subscribes to updates
func (c *A2AClient) SendTaskAndSubscribe(ctx context.Context, taskOpts TaskOptions) (<-chan *a2a.Task, error) {
	// Check if the server supports streaming
	if c.AgentCard == nil {
		if _, err := c.FetchAgentCard(ctx); err != nil {
			return nil, err
		}
	}

	if !c.AgentCard.Capabilities.Streaming {
		return nil, errors.New("server does not support streaming")
	}

	// Generate a task ID if not provided
	if taskOpts.ID == "" {
		taskOpts.ID = generateTaskID()
	}

	// Use default session ID if not provided
	if taskOpts.SessionID == "" {
		taskOpts.SessionID = c.DefaultSessionID
	}

	// Create a channel to send task updates
	taskCh := make(chan *a2a.Task)

	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(ctx)

	params := a2a.TasksSendSubscribeRequest{
		ID:                  taskOpts.ID,
		SessionID:           taskOpts.SessionID,
		AcceptedOutputModes: taskOpts.AcceptedOutputModes,
		Message:             taskOpts.Message,
		ParentTask:          taskOpts.ParentTask,
		PrevTasks:           taskOpts.PrevTasks,
		Metadata:            taskOpts.Metadata,
	}

	// Start a goroutine to handle the streaming response
	go func() {
		defer close(taskCh)
		defer cancel()

		// Create HTTP request
		jsonData, err := json.Marshal(a2a.JsonRpcRequest{
			JsonRpc: "2.0",
			ID:      rand.Int(),
			Method:  "tasks/sendSubscribe",
			Params:  params,
		})
		if err != nil {
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", c.ServerURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		for key, value := range c.RequestHeaders {
			req.Header.Set(key, value)
		}

		// Send request
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return
		}

		// Read response line by line
		decoder := json.NewDecoder(resp.Body)
		for {
			// Check if context is done
			select {
			case <-ctx.Done():
				return
			default:
				// Continue processing
			}

			// Parse response
			var response a2a.JsonRpcResponse
			if err := decoder.Decode(&response); err != nil {
				if err == io.EOF {
					return
				}
				continue
			}

			// Check for JSON-RPC error
			if response.Error != nil {
				continue
			}

			// Parse task
			var task a2a.Task
			b, err := json.Marshal(response.Result)
			if err != nil {
				continue
			}

			if err := json.Unmarshal(b, &task); err != nil {
				continue
			}

			// Send task update
			select {
			case taskCh <- &task:
				// Sent successfully
			case <-ctx.Done():
				return
			}

			// If task is completed, failed, or canceled, stop streaming
			if task.Status.State == a2a.TaskCompleted ||
				task.Status.State == a2a.TaskFailed ||
				task.Status.State == a2a.TaskCanceled {
				return
			}
		}
	}()

	return taskCh, nil
}

// TaskOptions contains options for sending a task
type TaskOptions struct {
	ID                  string
	SessionID           string
	AcceptedOutputModes []string
	Message             a2a.Message
	ParentTask          *a2a.ParentTask
	PrevTasks           []*a2a.ParentTask
	Metadata            any
}

// generateTaskID generates a random task ID
func generateTaskID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 16

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}

	return string(b)
}
