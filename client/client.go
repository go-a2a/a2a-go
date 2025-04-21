// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package client provides an implementation of the A2A protocol client.
package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-a2a/a2a"
)

const (
	defaultTimeout = 30 * time.Second
	userAgent      = "go-a2a/client " + a2a.Version
)

// Client is a client for the A2A protocol.
type Client struct {
	// httpClient is the HTTP client used for requests.
	httpClient *http.Client

	// url is the url of the A2A server.
	url string

	// agentCard is the agent card for the client.
	agentCard *a2a.AgentCard

	// logger for logging operations.
	logger *slog.Logger

	// tracer for OpenTelemetry tracing.
	tracer trace.Tracer
}

// NewClient creates a new [Client] with either a direct URL or [*a2a.AgentCard] option.
func NewClient(url string, opts ...Option) (*Client, error) {
	c := &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		url:    url,
		logger: slog.Default(),
		tracer: otel.GetTracerProvider().Tracer("github.com/go-a2a/a2a/client"),
	}
	for _, opt := range opts {
		opt(c)
	}

	if c.url == "" && c.agentCard == nil {
		return nil, errors.New("must provide either agent_card or url")
	}

	if c.agentCard != nil {
		c.url = c.agentCard.URL
	}

	return c, nil
}

// sendRequest makes an HTTP request to the A2A server.
func (c *Client) sendRequest(ctx context.Context, method, id string, payload any) ([]byte, error) {
	ctx, span := c.tracer.Start(ctx, "client.sendRequest",
		trace.WithAttributes(
			attribute.String("a2a.request_id", id),
			attribute.String("a2a.method", method),
		))
	defer span.End()

	request := &a2a.JSONRPCRequest{
		JSONRPCMessage: a2a.NewJSONRPCMessage(a2a.NewID(id)),
		Method:         method,
	}

	// Marshal the payload separately
	params, err := sonic.ConfigFastest.Marshal(payload)
	if err != nil {
		c.logger.ErrorContext(ctx, "marshal params", slog.Any("error", err))
		return nil, fmt.Errorf("marshal params: %w", err)
	}
	request.Params = params

	// Marshal the request
	data, err := sonic.ConfigFastest.Marshal(request)
	if err != nil {
		c.logger.ErrorContext(ctx, "create request", slog.Any("error", err))
		return nil, fmt.Errorf("create request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(data))
	if err != nil {
		c.logger.ErrorContext(ctx, "create HTTP request", slog.Any("error", err))
		return nil, fmt.Errorf("create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "send HTTP request", slog.Any("error", err))
		return nil, fmt.Errorf("send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.ErrorContext(ctx, "HTTP request failed with status", slog.String("status", resp.Status))
		return nil, fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.ErrorContext(ctx, "read response body", "error", err)
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return body, nil
}

// handleRPCError processes any JSON-RPC error response and returns an appropriate error.
func handleRPCError(jerr *a2a.JSONRPCError) error {
	if jerr == nil {
		return nil
	}
	return fmt.Errorf("RPC error: [%d] %s", jerr.Code, jerr.Message)
}

// SendTask sends a task to an A2A server.
func (c *Client) SendTask(ctx context.Context, req a2a.SendTaskRequest) (*a2a.Task, error) {
	ctx, span := c.tracer.Start(ctx, "client.SendTask")
	defer span.End()

	taskID := req.Params.ID
	span.SetAttributes(attribute.String("a2a.task_id", taskID))

	data, err := c.sendRequest(ctx, a2a.MethodTasksSend, taskID, req.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to send task: %w", err)
	}

	var resp a2a.SendTaskResponse
	if err := sonic.ConfigFastest.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if err := handleRPCError(resp.Error); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

// GetTask retrieves a task from an A2A server.
func (c *Client) GetTask(ctx context.Context, req *a2a.GetTaskRequest) (*a2a.Task, error) {
	ctx, span := c.tracer.Start(ctx, "client.GetTask")
	defer span.End()

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	params := map[string]string{
		"id": req.Params.ID,
	}
	data, err := c.sendRequest(ctx, a2a.MethodTasksGet, req.Params.ID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	var resp a2a.GetTaskResponse
	if err := sonic.ConfigFastest.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if err := handleRPCError(resp.Error); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

// SendTaskStreaming sends a task and subscribes to streaming updates.
// It returns a channel that will receive task events as they occur.
func (c *Client) SendTaskStreaming(ctx context.Context, req *a2a.SendTaskStreamingRequest) (<-chan a2a.TaskEvent, error) {
	ctx, span := c.tracer.Start(ctx, "client.SendTaskStreaming")
	defer span.End()

	taskID := req.Params.ID
	span.SetAttributes(attribute.String("a2a.task_id", taskID))

	// Create the events channel with reasonable buffer size
	events := make(chan a2a.TaskEvent, 10)

	// Start a separate goroutine to handle the streaming connection
	go func() {
		defer close(events)

		// Create a context with a default timeout
		streamCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
		defer cancel()

		// Open a streaming connection (implementation would depend on your SSE client)
		// This is a placeholder for the actual implementation
		c.logger.InfoContext(streamCtx, "Starting streaming connection for task", "taskID", taskID)

		// Handle the connection error and report it
		// TODO: Implement actual SSE client connection logic here
	}()

	return events, nil
}

// CancelTask cancels a task on an A2A server.
func (c *Client) CancelTask(ctx context.Context, req *a2a.CancelTaskRequest) (*a2a.Task, error) {
	ctx, span := c.tracer.Start(ctx, "client.CancelTask")
	defer span.End()

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	params := map[string]string{
		"id": req.Params.ID,
	}
	data, err := c.sendRequest(ctx, a2a.MethodTasksCancel, req.Params.ID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel task: %w", err)
	}

	var resp a2a.CancelTaskResponse
	if err := sonic.ConfigFastest.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if err := handleRPCError(resp.Error); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

// SetTaskPushNotification configures push notification for a task.
func (c *Client) SetTaskPushNotification(ctx context.Context, req *a2a.SetTaskPushNotificationRequest) (*a2a.TaskPushNotificationConfig, error) {
	ctx, span := c.tracer.Start(ctx, "client.SetTaskPushNotification")
	defer span.End()

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	params := a2a.TaskPushNotificationConfig{
		ID:                     req.Params.ID,
		PushNotificationConfig: req.Params.PushNotificationConfig,
	}
	data, err := c.sendRequest(ctx, a2a.MethodTasksPushNotificationSet, "", params)
	if err != nil {
		return nil, fmt.Errorf("failed to set task push notification: %w", err)
	}

	var resp a2a.SetTaskPushNotificationResponse
	if err := sonic.ConfigFastest.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if err := handleRPCError(resp.Error); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

// GetTaskPushNotification retrieves push notification configuration for a task.
func (c *Client) GetTaskPushNotification(ctx context.Context, req *a2a.GetTaskPushNotificationRequest) (*a2a.TaskPushNotificationConfig, error) {
	ctx, span := c.tracer.Start(ctx, "client.GetTaskPushNotification")
	defer span.End()

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	params := map[string]any{
		"id":       req.Params.ID,
		"metadata": req.Params.Metadata,
	}
	data, err := c.sendRequest(ctx, a2a.MethodTasksPushNotificationGet, req.Params.ID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get task push notification: %w", err)
	}

	var resp a2a.GetTaskPushNotificationResponse
	if err := sonic.ConfigFastest.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if err := handleRPCError(resp.Error); err != nil {
		return nil, err
	}

	return resp.Result, nil
}
