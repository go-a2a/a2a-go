// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/internal/jsonrpc2"
)

// Client is an A2A protocol client for communicating with A2A-compatible agents.
type Client struct {
	// baseURL is the base URL of the A2A server.
	baseURL string

	// httpClient is the HTTP client used for communication.
	httpClient *http.Client

	// transport is the JSON-RPC transport.
	transport *Transport

	// agentCard is the cached agent card, if fetched.
	agentCard   *a2a.AgentCard
	agentCardMu sync.RWMutex

	// reqID is a counter for generating unique request IDs.
	reqID int64

	// options contains client configuration options.
	options Options
}

// New creates a new A2A client with the given base URL and options.
func New(baseURL string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}

	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	transport := NewTransport(baseURL, httpClient)
	transport.SetHeaders(options.Headers)

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		transport:  transport,
		options:    options,
	}, nil
}

// nextID generates a unique request ID.
func (c *Client) nextID() jsonrpc2.ID {
	id := atomic.AddInt64(&c.reqID, 1)
	return jsonrpc2.Int64ID(id)
}

// GetAgentCard fetches the agent card from the server.
func (c *Client) GetAgentCard(ctx context.Context) (*a2a.AgentCard, error) {
	c.agentCardMu.RLock()
	if c.agentCard != nil {
		card := c.agentCard
		c.agentCardMu.RUnlock()
		return card, nil
	}
	c.agentCardMu.RUnlock()

	card, err := FetchAgentCard(ctx, c.baseURL, c.httpClient)
	if err != nil {
		return nil, fmt.Errorf("fetching agent card: %w", err)
	}

	c.agentCardMu.Lock()
	c.agentCard = card
	c.agentCardMu.Unlock()

	return card, nil
}

// SendTask sends a task to the agent and returns the result.
func (c *Client) SendTask(ctx context.Context, params a2a.TaskSendParams) (*a2a.Task, error) {
	req := a2a.NewSendTaskRequest(c.nextID(), params)

	var resp a2a.SendTaskResponse
	if err := c.transport.Call(ctx, req, &resp); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

// GetTask fetches a task from the agent.
func (c *Client) GetTask(ctx context.Context, params a2a.TaskQueryParams) (*a2a.Task, error) {
	req := a2a.NewGetTaskRequest(c.nextID(), params)

	var resp a2a.GetTaskResponse
	if err := c.transport.Call(ctx, req, &resp); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

// CancelTask cancels a task.
func (c *Client) CancelTask(ctx context.Context, params a2a.TaskIdParams) (*a2a.Task, error) {
	req := a2a.NewCancelTaskRequest(c.nextID(), params)

	var resp a2a.CancelTaskResponse
	if err := c.transport.Call(ctx, req, &resp); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

// SetTaskPushNotification sets a push notification for a task.
func (c *Client) SetTaskPushNotification(ctx context.Context, params a2a.TaskPushNotificationConfig) (*a2a.TaskPushNotificationConfig, error) {
	req := a2a.NewSetTaskPushNotificationRequest(c.nextID(), params)

	var resp a2a.SetTaskPushNotificationResponse
	if err := c.transport.Call(ctx, req, &resp); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

// GetTaskPushNotification gets the push notification configuration for a task.
func (c *Client) GetTaskPushNotification(ctx context.Context, params a2a.TaskIdParams) (*a2a.TaskPushNotificationConfig, error) {
	req := a2a.NewGetTaskPushNotificationRequest(c.nextID(), params)

	var resp a2a.GetTaskPushNotificationResponse
	if err := c.transport.Call(ctx, req, &resp); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

// Close releases any resources used by the client.
func (c *Client) Close() error {
	// Currently no resources to close
	return nil
}
