// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

var _ a2a.Client = (*Client)(nil)

// NewClient creates a new [Client] with either an [*a2a.AgentCard] or a direct URL.
func NewClient(agentCard *a2a.AgentCard, url string) (*Client, error) {
	if agentCard == nil && url == "" {
		return nil, fmt.Errorf("either agentCard or url must be provided")
	}

	clientURL := url
	if agentCard != nil {
		clientURL = agentCard.URL
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		url:       clientURL,
		agentCard: agentCard,
		logger:    slog.Default(),
		tracer:    otel.GetTracerProvider().Tracer("github.com/go-a2a/a2a"),
	}, nil
}

// WithHTTPClient sets the HTTP client for the A2AClient.
func (c *Client) WithHTTPClient(httpClient *http.Client) *Client {
	c.httpClient = httpClient
	return c
}

// WithLogger sets the logger for the A2AClient.
func (c *Client) WithLogger(logger *slog.Logger) *Client {
	c.logger = logger
	return c
}

// WithTracer sets the tracer for the A2AClient.
func (c *Client) WithTracer(tracer trace.Tracer) *Client {
	c.tracer = tracer
	return c
}

// makeRequest creates a JSON-RPC request from a method and params.
func makeRequest(method string, params any, id string) ([]byte, error) {
	req := struct {
		JsonRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params"`
		ID      string `json:"id,omitempty"`
	}{
		JsonRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}

	return sonic.ConfigFastest.Marshal(req)
}

// sendRequest makes an HTTP request to the A2A server.
func (c *Client) sendRequest(ctx context.Context, method string, params any, id string) ([]byte, error) {
	ctx, span := c.tracer.Start(ctx, "a2a.client.sendRequest",
		trace.WithAttributes(
			attribute.String("a2a.method", method),
			attribute.String("a2a.request_id", id),
		))
	defer span.End()

	data, err := makeRequest(method, params, id)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create request", "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(data))
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create HTTP request", "error", err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to send HTTP request", "error", err)
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.ErrorContext(ctx, "HTTP request failed", "status", resp.Status)
		return nil, fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	respBody := make([]byte, 0)
	buf := bytes.NewBuffer(respBody)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to read response body", "error", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return buf.Bytes(), nil
}

// SendTask sends a task to an A2A server.
func (c *Client) SendTask(ctx context.Context, req a2a.Request) (*a2a.SendTaskResponse, error) {
	ctx, span := c.tracer.Start(ctx, "a2a.client.SendTask")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	sendReq, ok := req.(a2a.SendTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected SendTaskRequest but got %T", req)
	}

	reqID := sendReq.Task.ID
	span.SetAttributes(attribute.String("a2a.task_id", reqID))

	responseData, err := c.sendRequest(ctx, "tasks/send", sendReq.Task, reqID)
	if err != nil {
		return nil, fmt.Errorf("failed to send task: %w", err)
	}

	var jsonRPCResp struct {
		JsonRPC string          `json:"jsonrpc"`
		Result  a2a.Task        `json:"result"`
		Error   json.RawMessage `json:"error"`
	}

	if err := sonic.ConfigFastest.Unmarshal(responseData, &jsonRPCResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(jsonRPCResp.Error) > 0 && string(jsonRPCResp.Error) != "null" {
		var rpcError struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := sonic.ConfigFastest.Unmarshal(jsonRPCResp.Error, &rpcError); err != nil {
			return nil, fmt.Errorf("failed to parse error: %w", err)
		}
		return nil, fmt.Errorf("RPC error: [%d] %s", rpcError.Code, rpcError.Message)
	}

	return &a2a.SendTaskResponse{
		Task:      jsonRPCResp.Result,
		RequestID: reqID,
	}, nil
}

// GetTask retrieves a task from an A2A server.
func (c *Client) GetTask(ctx context.Context, req a2a.Request) (*a2a.GetTaskResponse, error) {
	ctx, span := c.tracer.Start(ctx, "a2a.client.GetTask")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	getReq, ok := req.(a2a.GetTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected GetTaskRequest but got %T", req)
	}

	span.SetAttributes(attribute.String("a2a.task_id", getReq.TaskID))

	params := map[string]string{
		"id": getReq.TaskID,
	}

	responseData, err := c.sendRequest(ctx, "tasks/get", params, getReq.RequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	var jsonRPCResp struct {
		JsonRPC string          `json:"jsonrpc"`
		Result  a2a.Task        `json:"result"`
		Error   json.RawMessage `json:"error"`
	}

	if err := sonic.ConfigFastest.Unmarshal(responseData, &jsonRPCResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(jsonRPCResp.Error) > 0 && string(jsonRPCResp.Error) != "null" {
		var rpcError struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := sonic.ConfigFastest.Unmarshal(jsonRPCResp.Error, &rpcError); err != nil {
			return nil, fmt.Errorf("failed to parse error: %w", err)
		}
		return nil, fmt.Errorf("RPC error: [%d] %s", rpcError.Code, rpcError.Message)
	}

	return &a2a.GetTaskResponse{
		Task:      jsonRPCResp.Result,
		RequestID: getReq.RequestID,
	}, nil
}

// CancelTask cancels a task on an A2A server.
func (c *Client) CancelTask(ctx context.Context, req a2a.Request) (*a2a.CancelTaskResponse, error) {
	ctx, span := c.tracer.Start(ctx, "a2a.client.CancelTask")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	cancelReq, ok := req.(a2a.CancelTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected CancelTaskRequest but got %T", req)
	}

	span.SetAttributes(attribute.String("a2a.task_id", cancelReq.TaskID))

	params := map[string]string{
		"id": cancelReq.TaskID,
	}

	responseData, err := c.sendRequest(ctx, "tasks/cancel", params, cancelReq.RequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel task: %w", err)
	}

	var jsonRPCResp struct {
		JsonRPC string          `json:"jsonrpc"`
		Result  a2a.Task        `json:"result"`
		Error   json.RawMessage `json:"error"`
	}

	if err := sonic.ConfigFastest.Unmarshal(responseData, &jsonRPCResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(jsonRPCResp.Error) > 0 && string(jsonRPCResp.Error) != "null" {
		var rpcError struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := sonic.ConfigFastest.Unmarshal(jsonRPCResp.Error, &rpcError); err != nil {
			return nil, fmt.Errorf("failed to parse error: %w", err)
		}
		return nil, fmt.Errorf("RPC error: [%d] %s", rpcError.Code, rpcError.Message)
	}

	return &a2a.CancelTaskResponse{
		Task:      jsonRPCResp.Result,
		RequestID: cancelReq.RequestID,
	}, nil
}

// SetTaskPushNotification configures push notification for a task.
func (c *Client) SetTaskPushNotification(ctx context.Context, req a2a.Request) (*a2a.SetTaskPushNotificationResponse, error) {
	ctx, span := c.tracer.Start(ctx, "a2a.client.SetTaskPushNotification")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	pushReq, ok := req.(a2a.SetTaskPushNotificationRequest)
	if !ok {
		return nil, fmt.Errorf("expected SetTaskPushNotificationRequest but got %T", req)
	}

	span.SetAttributes(attribute.String("a2a.task_id", pushReq.TaskID))

	params := a2a.SetTaskPushNotificationRequest{
		TaskID:                 pushReq.TaskID,
		PushNotificationConfig: pushReq.PushNotificationConfig,
	}

	responseData, err := c.sendRequest(ctx, "tasks/pushNotification/set", params, "")
	if err != nil {
		return nil, fmt.Errorf("failed to set task push notification: %w", err)
	}

	var jsonRPCResp struct {
		JsonRPC string                         `json:"jsonrpc"`
		Result  a2a.TaskPushNotificationConfig `json:"result"`
		Error   json.RawMessage                `json:"error"`
	}

	if err := sonic.ConfigFastest.Unmarshal(responseData, &jsonRPCResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(jsonRPCResp.Error) > 0 && string(jsonRPCResp.Error) != "null" {
		var rpcError struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := sonic.ConfigFastest.Unmarshal(jsonRPCResp.Error, &rpcError); err != nil {
			return nil, fmt.Errorf("failed to parse error: %w", err)
		}
		return nil, fmt.Errorf("RPC error: [%d] %s", rpcError.Code, rpcError.Message)
	}

	return &a2a.SetTaskPushNotificationResponse{
		Config:    jsonRPCResp.Result,
		RequestID: "",
	}, nil
}

// GetTaskPushNotification retrieves push notification configuration for a task.
func (c *Client) GetTaskPushNotification(ctx context.Context, req a2a.Request) (*a2a.GetTaskPushNotificationResponse, error) {
	ctx, span := c.tracer.Start(ctx, "a2a.client.GetTaskPushNotification")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	getPushReq, ok := req.(a2a.GetTaskPushNotificationRequest)
	if !ok {
		return nil, fmt.Errorf("expected GetTaskPushNotificationRequest but got %T", req)
	}

	span.SetAttributes(attribute.String("a2a.task_id", getPushReq.TaskID))

	params := map[string]any{
		"id":       getPushReq.TaskID,
		"metadata": getPushReq.Metadata,
	}

	responseData, err := c.sendRequest(ctx, "tasks/pushNotification/get", params, getPushReq.RequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task push notification: %w", err)
	}

	var jsonRPCResp struct {
		JsonRPC string                         `json:"jsonrpc"`
		Result  a2a.TaskPushNotificationConfig `json:"result"`
		Error   json.RawMessage                `json:"error"`
	}

	if err := sonic.ConfigFastest.Unmarshal(responseData, &jsonRPCResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(jsonRPCResp.Error) > 0 && string(jsonRPCResp.Error) != "null" {
		var rpcError struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := sonic.ConfigFastest.Unmarshal(jsonRPCResp.Error, &rpcError); err != nil {
			return nil, fmt.Errorf("failed to parse error: %w", err)
		}
		return nil, fmt.Errorf("RPC error: [%d] %s", rpcError.Code, rpcError.Message)
	}

	return &a2a.GetTaskPushNotificationResponse{
		Config:    jsonRPCResp.Result,
		RequestID: getPushReq.RequestID,
	}, nil
}

// ResubscribeTask resubscribes to a task's updates.
func (c *Client) ResubscribeTask(ctx context.Context, req a2a.Request) (*a2a.TaskStatusUpdateEvent, error) {
	ctx, span := c.tracer.Start(ctx, "a2a.client.ResubscribeTask")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resubReq, ok := req.(a2a.TaskResubscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("expected TaskResubscriptionRequest but got %T", req)
	}

	span.SetAttributes(attribute.String("a2a.task_id", resubReq.TaskID))

	params := map[string]any{
		"id":            resubReq.TaskID,
		"historyLength": resubReq.HistoryLength,
		"metadata":      resubReq.Metadata,
	}

	responseData, err := c.sendRequest(ctx, "tasks/resubscribe", params, "")
	if err != nil {
		return nil, fmt.Errorf("failed to resubscribe to task: %w", err)
	}

	var jsonRPCResp struct {
		JsonRPC string                    `json:"jsonrpc"`
		Result  a2a.TaskStatusUpdateEvent `json:"result"`
		Error   json.RawMessage           `json:"error"`
	}

	if err := sonic.ConfigFastest.Unmarshal(responseData, &jsonRPCResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(jsonRPCResp.Error) > 0 && string(jsonRPCResp.Error) != "null" {
		var rpcError struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := sonic.ConfigFastest.Unmarshal(jsonRPCResp.Error, &rpcError); err != nil {
			return nil, fmt.Errorf("failed to parse error: %w", err)
		}
		return nil, fmt.Errorf("RPC error: [%d] %s", rpcError.Code, rpcError.Message)
	}

	return &jsonRPCResp.Result, nil
}
