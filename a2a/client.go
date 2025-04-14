// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	json "github.com/bytedance/sonic"
)

// Client represents an A2A API client.
type Client struct {
	// BaseURL is the base URL for the A2A API.
	BaseURL *url.URL
	// HTTPClient is the HTTP client used for API requests.
	HTTPClient *http.Client
	// RequestTimeout is the timeout for API requests.
	RequestTimeout time.Duration
	// UserAgent is the user agent to use for API requests.
	UserAgent string
}

// NewClient creates a new A2A client with the given baseURL.
func NewClient(baseURL string) (*Client, error) {
	// Parse the base URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Create the client
	return &Client{
		BaseURL:        u,
		HTTPClient:     http.DefaultClient,
		RequestTimeout: 30 * time.Second,
		UserAgent:      "go-a2a/" + Version,
	}, nil
}

// GetAgentCard fetches the agent card from the server.
func (c *Client) GetAgentCard(ctx context.Context) (*AgentCard, error) {
	// Create the request URL
	u := *c.BaseURL
	u.Path = "/.well-known/agent.json"

	// Make the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the headers
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response
	var card AgentCard
	if err := json.ConfigFastest.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &card, nil
}

// SendRequest sends a JSON-RPC request to the server.
func (c *Client) SendRequest(ctx context.Context, request any, response any) error {
	// Marshal the request to JSON
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL.String(), bytes.NewReader(requestBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set the headers
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response
	if err := json.ConfigFastest.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// SendTask sends a task to the server.
func (c *Client) SendTask(ctx context.Context, params TaskSendParams) (*Task, error) {
	// Create the request
	request := NewSendTaskRequest(generateID(), params)

	// Send the request
	var response SendTaskResponse
	if err := c.SendRequest(ctx, request, &response); err != nil {
		return nil, err
	}

	// Check for JSON-RPC error
	if response.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error: %d - %s", response.Error.Code, response.Error.Message)
	}

	return response.Result, nil
}

// GetTask retrieves a task from the server.
func (c *Client) GetTask(ctx context.Context, id string, historyLength *int) (*Task, error) {
	// Create the request parameters
	params := TaskQueryParams{
		ID:            id,
		HistoryLength: historyLength,
	}

	// Create the request
	request := NewGetTaskRequest(generateID(), params)

	// Send the request
	var response GetTaskResponse
	if err := c.SendRequest(ctx, request, &response); err != nil {
		return nil, err
	}

	// Check for JSON-RPC error
	if response.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error: %d - %s", response.Error.Code, response.Error.Message)
	}

	return response.Result, nil
}

// CancelTask cancels a task on the server.
func (c *Client) CancelTask(ctx context.Context, id string) (*Task, error) {
	// Create the request parameters
	params := TaskIdParams{
		ID: id,
	}

	// Create the request
	request := NewCancelTaskRequest(generateID(), params)

	// Send the request
	var response CancelTaskResponse
	if err := c.SendRequest(ctx, request, &response); err != nil {
		return nil, err
	}

	// Check for JSON-RPC error
	if response.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error: %d - %s", response.Error.Code, response.Error.Message)
	}

	return response.Result, nil
}

// SetTaskPushNotification sets push notification configuration for a task.
func (c *Client) SetTaskPushNotification(ctx context.Context, taskID string, config PushNotificationConfig) (*TaskPushNotificationConfig, error) {
	// Create the request parameters
	params := TaskPushNotificationConfig{
		ID:                     taskID,
		PushNotificationConfig: config,
	}

	// Create the request
	request := NewSetTaskPushNotificationRequest(generateID(), params)

	// Send the request
	var response SetTaskPushNotificationResponse
	if err := c.SendRequest(ctx, request, &response); err != nil {
		return nil, err
	}

	// Check for JSON-RPC error
	if response.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error: %d - %s", response.Error.Code, response.Error.Message)
	}

	return response.Result, nil
}

// GetTaskPushNotification retrieves push notification configuration for a task.
func (c *Client) GetTaskPushNotification(ctx context.Context, taskID string) (*TaskPushNotificationConfig, error) {
	// Create the request parameters
	params := TaskIdParams{
		ID: taskID,
	}

	// Create the request
	request := NewGetTaskPushNotificationRequest(generateID(), params)

	// Send the request
	var response GetTaskPushNotificationResponse
	if err := c.SendRequest(ctx, request, &response); err != nil {
		return nil, err
	}

	// Check for JSON-RPC error
	if response.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error: %d - %s", response.Error.Code, response.Error.Message)
	}

	return response.Result, nil
}

// SendTaskStreamingResponseReader reads streaming response events from a reader.
type SendTaskStreamingResponseReader struct {
	reader json.Decoder
}

// NewSendTaskStreamingResponseReader creates a new SendTaskStreamingResponseReader.
func NewSendTaskStreamingResponseReader(reader io.Reader) *SendTaskStreamingResponseReader {
	return &SendTaskStreamingResponseReader{
		reader: json.ConfigFastest.NewDecoder(reader),
	}
}

// ReadEvent reads a streaming response event from the reader.
func (r *SendTaskStreamingResponseReader) ReadEvent() (*SendTaskStreamingResponse, error) {
	var response SendTaskStreamingResponse
	if err := r.reader.Decode(&response); err != nil {
		return nil, err
	}

	// Check for JSON-RPC error
	if response.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error: %d - %s", response.Error.Code, response.Error.Message)
	}

	// Parse the result into the appropriate event type
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	// Try to parse as a TaskStatusUpdateEvent
	var statusEvent TaskStatusUpdateEvent
	if err := json.Unmarshal(resultBytes, &statusEvent); err == nil {
		if statusEvent.ID != "" {
			response.Result = statusEvent
			return &response, nil
		}
	}

	// Try to parse as a TaskArtifactUpdateEvent
	var artifactEvent TaskArtifactUpdateEvent
	if err := json.Unmarshal(resultBytes, &artifactEvent); err == nil {
		if artifactEvent.ID != "" {
			response.Result = artifactEvent
			return &response, nil
		}
	}

	return nil, fmt.Errorf("unknown event type")
}

// Generate a unique ID for JSON-RPC requests.
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
