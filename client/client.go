// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package client provides an implementation of the A2A protocol client.
package client

import (
	"bytes"
	"context"
	"encoding/json"
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
func NewClient(uri string, opts ...ClientOption) (*Client, error) {
	c := &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		url:    uri,
		logger: slog.Default(),
		tracer: otel.GetTracerProvider().Tracer("github.com/go-a2a/a2a"),
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
	ctx, span := c.tracer.Start(ctx, "a2a.client.sendRequest",
		trace.WithAttributes(
			attribute.String("a2a.request_id", id),
			attribute.String("a2a.method", method),
		))
	defer span.End()

	request := &a2a.JSONRPCRequest{
		JSONRPCMessage: a2a.NewJSONRPCMessage(id),
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
		c.logger.ErrorContext(ctx, "HTTP request", slog.String("status", resp.Status))
		return nil, fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to read response body", "error", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

var null = json.RawMessage("null")

// handleRPCError processes any JSON-RPC error response and returns an appropriate error.
func handleRPCError(data json.RawMessage) error {
	if len(data) == 0 || bytes.Equal(data, null) {
		return nil
	}

	var rpcErr a2a.JSONRPCError
	if err := json.Unmarshal(data, &rpcErr); err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	return fmt.Errorf("RPC error: [%d] %s", rpcErr.Code, rpcErr.Message)
}

// SendTask sends a task to an A2A server.
func (c *Client) SendTask(ctx context.Context, req a2a.A2ARequest) (*a2a.Task, error) {
	ctx, span := c.tracer.Start(ctx, "a2a.client.SendTask")
	defer span.End()

	sendReq, ok := req.(*a2a.SendTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected SendTaskRequest but got %T", req)
	}

	taskID := sendReq.Params.ID
	span.SetAttributes(attribute.String("a2a.task_id", taskID))

	data, err := c.sendRequest(ctx, a2a.MethodTasksSend, taskID, sendReq.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to send task: %w", err)
	}

	var resp a2a.SendTaskResponse
	if err := sonic.ConfigFastest.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if err := handleRPCError(json.RawMessage(resp.Error.Message)); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

// SendTaskStreaming sends a task and subscribes to streaming updates.
// It returns a channel that will receive task events as they occur.
func (c *Client) SendTaskStreaming(ctx context.Context, req a2a.A2ARequest) (<-chan a2a.TaskEvent, error) {
	ctx, span := c.tracer.Start(ctx, "a2a.client.SendTaskStreaming")
	defer span.End()

	streamReq, ok := req.(*a2a.SendTaskStreamingRequest)
	if !ok {
		return nil, fmt.Errorf("expected SendTaskStreamingRequest but got %T", req)
	}

	taskID := streamReq.Params.ID
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

// GetTask retrieves a task from an A2A server.
func (c *Client) GetTask(ctx context.Context, req a2a.A2ARequest) (*a2a.Task, error) {
	ctx, span := c.tracer.Start(ctx, "a2a.client.GetTask")
	defer span.End()

	getReq, ok := req.(*a2a.GetTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected GetTaskRequest but got %T", req)
	}

	span.SetAttributes(attribute.String("a2a.task_id", getReq.Params.ID))

	params := map[string]string{
		"id": getReq.Params.ID,
	}

	data, err := c.sendRequest(ctx, getReq.Params.ID, a2a.MethodTasksGet, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	var jsonRPCResp struct {
		JSONRPC string          `json:"jsonrpc"`
		Result  a2a.Task        `json:"result"`
		Error   json.RawMessage `json:"error"`
	}

	if err := sonic.ConfigFastest.Unmarshal(data, &jsonRPCResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if err := handleRPCError(jsonRPCResp.Error); err != nil {
		return nil, err
	}

	return &jsonRPCResp.Result, nil
}

// CancelTask cancels a task on an A2A server.
func (c *Client) CancelTask(ctx context.Context, req a2a.A2ARequest) (*a2a.Task, error) {
	ctx, span := c.tracer.Start(ctx, "a2a.client.CancelTask")
	defer span.End()

	cancelReq, ok := req.(*a2a.CancelTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected CancelTaskRequest but got %T", req)
	}

	span.SetAttributes(attribute.String("a2a.task_id", cancelReq.Params.ID))

	params := map[string]string{
		"id": cancelReq.Params.ID,
	}

	data, err := c.sendRequest(ctx, cancelReq.Params.ID, a2a.MethodTasksCancel, params)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel task: %w", err)
	}

	var jsonRPCResp struct {
		JSONRPC string          `json:"jsonrpc"`
		Result  a2a.Task        `json:"result"`
		Error   json.RawMessage `json:"error"`
	}

	if err := sonic.ConfigFastest.Unmarshal(data, &jsonRPCResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if err := handleRPCError(jsonRPCResp.Error); err != nil {
		return nil, err
	}

	return &jsonRPCResp.Result, nil
}

// SetTaskPushNotification configures push notification for a task.
func (c *Client) SetTaskPushNotification(ctx context.Context, req a2a.A2ARequest) (*a2a.TaskPushNotificationConfig, error) {
	ctx, span := c.tracer.Start(ctx, "a2a.client.SetTaskPushNotification")
	defer span.End()

	pushReq, ok := req.(*a2a.SetTaskPushNotificationRequest)
	if !ok {
		return nil, fmt.Errorf("expected SetTaskPushNotificationRequest but got %T", req)
	}

	span.SetAttributes(attribute.String("a2a.task_id", pushReq.Params.ID))

	params := a2a.TaskPushNotificationConfig{
		ID:                     pushReq.Params.ID,
		PushNotificationConfig: pushReq.Params.PushNotificationConfig,
	}

	data, err := c.sendRequest(ctx, "", a2a.MethodTasksPushNotificationSet, params)
	if err != nil {
		return nil, fmt.Errorf("failed to set task push notification: %w", err)
	}

	var jsonRPCResp struct {
		JSONRPC string                         `json:"jsonrpc"`
		Result  a2a.TaskPushNotificationConfig `json:"result"`
		Error   json.RawMessage                `json:"error"`
	}

	if err := sonic.ConfigFastest.Unmarshal(data, &jsonRPCResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if err := handleRPCError(jsonRPCResp.Error); err != nil {
		return nil, err
	}

	return &jsonRPCResp.Result, nil
}

// GetTaskPushNotification retrieves push notification configuration for a task.
func (c *Client) GetTaskPushNotification(ctx context.Context, req a2a.A2ARequest) (*a2a.TaskPushNotificationConfig, error) {
	ctx, span := c.tracer.Start(ctx, "a2a.client.GetTaskPushNotification")
	defer span.End()

	getPushReq, ok := req.(*a2a.GetTaskPushNotificationRequest)
	if !ok {
		return nil, fmt.Errorf("expected GetTaskPushNotificationRequest but got %T", req)
	}

	span.SetAttributes(attribute.String("a2a.task_id", getPushReq.Params.ID))

	params := map[string]any{
		"id":       getPushReq.Params.ID,
		"metadata": getPushReq.Params.Metadata,
	}

	data, err := c.sendRequest(ctx, getPushReq.Params.ID, a2a.MethodTasksPushNotificationGet, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get task push notification: %w", err)
	}

	var resp a2a.GetTaskPushNotificationResponse
	if err := sonic.ConfigFastest.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if err := handleRPCError(resp.Result); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

// ResubscribeTask resubscribes to a task's updates.
func (c *Client) ResubscribeTask(ctx context.Context, req a2a.A2ARequest) (*a2a.TaskStatusUpdateEvent, error) {
	ctx, span := c.tracer.Start(ctx, "a2a.client.ResubscribeTask")
	defer span.End()

	resubReq, ok := req.(*a2a.TaskResubscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("expected TaskResubscriptionRequest but got %T", req)
	}

	span.SetAttributes(attribute.String("a2a.task_id", resubReq.Params.ID))

	params := map[string]any{
		"id":            resubReq.Params.ID,
		"historyLength": resubReq.Params.HistoryLength,
		"metadata":      resubReq.Params.Metadata,
	}

	data, err := c.sendRequest(ctx, "", a2a.MethodTasksResubscribe, params)
	if err != nil {
		return nil, fmt.Errorf("failed to resubscribe to task: %w", err)
	}

	var resp struct {
		JSONRPC string                    `json:"jsonrpc"`
		Result  a2a.TaskStatusUpdateEvent `json:"result"`
		Error   json.RawMessage           `json:"error"`
	}
	if err := sonic.ConfigFastest.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if err := handleRPCError(resp.Error); err != nil {
		return nil, err
	}

	return &resp.Result, nil
}
