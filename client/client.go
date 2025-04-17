// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package client implements the A2A client functionality.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	sse "github.com/r3labs/sse/v2"

	"github.com/go-a2a/a2a"
)

// A2AClient is a client for interacting with A2A servers.
type A2AClient struct {
	// HTTPClient is the HTTP client used for making requests.
	HTTPClient *http.Client
	// BaseURL is the base URL of the A2A server.
	BaseURL string
	// DefaultTimeout is the default timeout for requests.
	DefaultTimeout time.Duration
}

// NewA2AClient creates a new A2A client with the given base URL.
func NewA2AClient(baseURL string) *A2AClient {
	return &A2AClient{
		HTTPClient:     &http.Client{},
		BaseURL:        baseURL,
		DefaultTimeout: 60 * time.Second,
	}
}

// newJSONRPCRequest creates a new JSON-RPC request.
func newJSONRPCRequest(method string, params any) (*a2a.JSONRPCRequest, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	return &a2a.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      uuid.New().String(),
		Method:  method,
		Params:  paramsJSON,
	}, nil
}

// sendRequest sends a JSON-RPC request to the A2A server and returns the response.
func (c *A2AClient) sendRequest(ctx context.Context, req *a2a.JSONRPCRequest, result any) error {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(reqJSON))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("received non-OK response: %d %s: %s", resp.StatusCode, resp.Status, string(bodyBytes))
	}

	var jsonResp a2a.JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if jsonResp.Error != nil {
		return fmt.Errorf("received error response: %d %s", jsonResp.Error.Code, jsonResp.Error.Message)
	}

	resultJSON, err := json.Marshal(jsonResp.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result for decoding: %w", err)
	}

	if err := json.Unmarshal(resultJSON, result); err != nil {
		return fmt.Errorf("failed to decode result: %w", err)
	}

	return nil
}

// GetAgentCard retrieves the agent card from the A2A server.
func (c *A2AClient) GetAgentCard(ctx context.Context) (*a2a.AgentCard, error) {
	req, err := newJSONRPCRequest("agent/getCard", nil)
	if err != nil {
		return nil, err
	}

	var card a2a.AgentCard
	if err := c.sendRequest(ctx, req, &card); err != nil {
		return nil, err
	}

	return &card, nil
}

// TaskSend sends a task to the A2A server.
func (c *A2AClient) TaskSend(ctx context.Context, params *a2a.TaskSendParams) (*a2a.Task, error) {
	req, err := newJSONRPCRequest("tasks/send", params)
	if err != nil {
		return nil, err
	}

	var task a2a.Task
	if err := c.sendRequest(ctx, req, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// TaskGet retrieves a task from the A2A server.
func (c *A2AClient) TaskGet(ctx context.Context, id string) (*a2a.Task, error) {
	params := a2a.TaskGetParams{ID: id}
	req, err := newJSONRPCRequest("tasks/get", params)
	if err != nil {
		return nil, err
	}

	var task a2a.Task
	if err := c.sendRequest(ctx, req, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// TaskCancel cancels a task on the A2A server.
func (c *A2AClient) TaskCancel(ctx context.Context, id string) (*a2a.Task, error) {
	params := a2a.TaskCancelParams{ID: id}
	req, err := newJSONRPCRequest("tasks/cancel", params)
	if err != nil {
		return nil, err
	}

	var task a2a.Task
	if err := c.sendRequest(ctx, req, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// SetPushNotification configures push notifications for the A2A server.
func (c *A2AClient) SetPushNotification(ctx context.Context, config *a2a.PushNotificationConfig) error {
	params := a2a.SetPushNotificationParams{Config: *config}
	req, err := newJSONRPCRequest("tasks/pushNotification/set", params)
	if err != nil {
		return err
	}

	var result any
	if err := c.sendRequest(ctx, req, &result); err != nil {
		return err
	}

	return nil
}

// TaskSendSubscribeHandler is a handler for streaming events from a task.
type TaskSendSubscribeHandler interface {
	// OnStatusUpdate is called when a task status update is received.
	OnStatusUpdate(event *a2a.TaskStatusUpdateEvent) error
	// OnArtifactUpdate is called when a task artifact update is received.
	OnArtifactUpdate(event *a2a.TaskArtifactUpdateEvent) error
	// OnComplete is called when the task is completed.
	OnComplete(task *a2a.Task) error
	// OnError is called when an error occurs.
	OnError(err error)
}

// DefaultTaskSendSubscribeHandler is a default implementation of TaskSendSubscribeHandler.
type DefaultTaskSendSubscribeHandler struct {
	// Status is the latest task status.
	Status *a2a.TaskStatus
	// Artifacts is a map of artifact indices to artifacts.
	Artifacts map[int]*a2a.Artifact
	// Task is the completed task.
	Task *a2a.Task
	// Err is the last error that occurred.
	Err error
}

// NewDefaultTaskSendSubscribeHandler creates a new default handler.
func NewDefaultTaskSendSubscribeHandler() *DefaultTaskSendSubscribeHandler {
	return &DefaultTaskSendSubscribeHandler{
		Artifacts: make(map[int]*a2a.Artifact),
	}
}

// OnStatusUpdate handles task status updates.
func (h *DefaultTaskSendSubscribeHandler) OnStatusUpdate(event *a2a.TaskStatusUpdateEvent) error {
	h.Status = &event.Status
	return nil
}

// OnArtifactUpdate handles task artifact updates.
func (h *DefaultTaskSendSubscribeHandler) OnArtifactUpdate(event *a2a.TaskArtifactUpdateEvent) error {
	h.Artifacts[event.Artifact.Index] = &event.Artifact
	return nil
}

// OnComplete handles task completion.
func (h *DefaultTaskSendSubscribeHandler) OnComplete(task *a2a.Task) error {
	h.Task = task
	return nil
}

// OnError handles errors.
func (h *DefaultTaskSendSubscribeHandler) OnError(err error) {
	h.Err = err
}

// TaskSendSubscribe sends a task to the A2A server and subscribes to streaming updates.
func (c *A2AClient) TaskSendSubscribe(ctx context.Context, params *a2a.TaskSendParams, handler TaskSendSubscribeHandler) error {
	reqJSON, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	url := fmt.Sprintf("%s/tasks/sendSubscribe", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqJSON))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	client := sse.NewClient(url)
	client.Headers = map[string]string{
		"Content-Type": "application/json",
		"Accept":       "text/event-stream",
	}
	client.ResponseValidator = func(c *sse.Client, resp *http.Response) error {
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("received non-OK response: %d %s", resp.StatusCode, resp.Status)
		}
		return nil
	}

	events := make(chan *sse.Event)
	eventsFunc := func(msg *sse.Event) {
		events <- msg
	}
	errCh := make(chan error)

	go func() {
		if err := client.SubscribeWithContext(ctx, "", eventsFunc); err != nil {
			errCh <- err
		}
	}()

	for {
		select {
		case event := <-events:
			if err := handleSSEEvent(event, handler); err != nil {
				handler.OnError(err)
				return err
			}
		case err := <-errCh:
			handler.OnError(err)
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// handleSSEEvent handles a server-sent event.
func handleSSEEvent(event *sse.Event, handler TaskSendSubscribeHandler) error {
	var jsonRPCEvent a2a.JSONRPCEvent
	if err := json.Unmarshal(event.Data, &jsonRPCEvent); err != nil {
		return fmt.Errorf("failed to decode SSE event data: %w", err)
	}

	paramsJSON, err := json.Marshal(jsonRPCEvent.Params)
	if err != nil {
		return fmt.Errorf("failed to marshal event params: %w", err)
	}

	switch jsonRPCEvent.Method {
	case "task/status":
		var statusEvent a2a.TaskStatusUpdateEvent
		if err := json.Unmarshal(paramsJSON, &statusEvent); err != nil {
			return fmt.Errorf("failed to decode status event: %w", err)
		}
		return handler.OnStatusUpdate(&statusEvent)
	case "task/artifact":
		var artifactEvent a2a.TaskArtifactUpdateEvent
		if err := json.Unmarshal(paramsJSON, &artifactEvent); err != nil {
			return fmt.Errorf("failed to decode artifact event: %w", err)
		}
		return handler.OnArtifactUpdate(&artifactEvent)
	case "task/complete":
		var task a2a.Task
		if err := json.Unmarshal(paramsJSON, &task); err != nil {
			return fmt.Errorf("failed to decode complete event: %w", err)
		}
		return handler.OnComplete(&task)
	default:
		return fmt.Errorf("unknown event method: %s", jsonRPCEvent.Method)
	}
}

// CardResolver is an interface for resolving agent cards.
type CardResolver interface {
	// ResolveCard resolves an agent card from a URL.
	ResolveCard(ctx context.Context, url string) (*a2a.AgentCard, error)
}

// A2ACardResolver is an implementation of CardResolver for A2A servers.
type A2ACardResolver struct {
	// HTTPClient is the HTTP client used for making requests.
	HTTPClient *http.Client
	// DefaultTimeout is the default timeout for requests.
	DefaultTimeout time.Duration
	// Cache is a cache of resolved agent cards.
	Cache map[string]*a2a.AgentCard
}

// NewA2ACardResolver creates a new A2A card resolver.
func NewA2ACardResolver() *A2ACardResolver {
	return &A2ACardResolver{
		HTTPClient:     &http.Client{},
		DefaultTimeout: 30 * time.Second,
		Cache:          make(map[string]*a2a.AgentCard),
	}
}

// ResolveCard resolves an agent card from a URL.
func (r *A2ACardResolver) ResolveCard(ctx context.Context, url string) (*a2a.AgentCard, error) {
	// Check cache first
	if card, ok := r.Cache[url]; ok {
		return card, nil
	}

	// Create a new client for the given URL
	client := NewA2AClient(url)
	client.HTTPClient = r.HTTPClient
	client.DefaultTimeout = r.DefaultTimeout

	// Get the agent card
	card, err := client.GetAgentCard(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.Cache[url] = card

	return card, nil
}

// ClearCache clears the card resolver's cache.
func (r *A2ACardResolver) ClearCache() {
	r.Cache = make(map[string]*a2a.AgentCard)
}
