// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

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
)

const (
	defaultTimeout = 30 * time.Second
	userAgent      = "go-a2a/client " + Version
)

// A2AClient is a client for the A2A protocol.
type A2AClient struct {
	// HttpClient is the HTTP client used for requests.
	HttpClient *http.Client

	// URL is the URL of the A2A server.
	URL string

	// AgentCard is the agent card for the client.
	AgentCard *AgentCard

	// Logger for logging operations.
	Logger *slog.Logger

	// Tracer for OpenTelemetry tracing.
	Tracer trace.Tracer
}

// NewA2AClient creates a new A2AClient with either an AgentCard or a direct URL.
func NewA2AClient(agentCard *AgentCard, url string) (*A2AClient, error) {
	if agentCard == nil && url == "" {
		return nil, fmt.Errorf("either agentCard or url must be provided")
	}

	clientURL := url
	if agentCard != nil {
		clientURL = agentCard.URL
	}

	return &A2AClient{
		HttpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		URL:       clientURL,
		AgentCard: agentCard,
		Logger:    slog.Default(),
		Tracer:    otel.GetTracerProvider().Tracer("github.com/go-a2a/a2a"),
	}, nil
}

// WithHTTPClient sets the HTTP client for the A2AClient.
func (c *A2AClient) WithHTTPClient(httpClient *http.Client) *A2AClient {
	c.HttpClient = httpClient
	return c
}

// WithLogger sets the logger for the A2AClient.
func (c *A2AClient) WithLogger(logger *slog.Logger) *A2AClient {
	c.Logger = logger
	return c
}

// WithTracer sets the tracer for the A2AClient.
func (c *A2AClient) WithTracer(tracer trace.Tracer) *A2AClient {
	c.Tracer = tracer
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
func (c *A2AClient) sendRequest(ctx context.Context, method string, params any, id string) ([]byte, error) {
	ctx, span := c.Tracer.Start(ctx, "a2a.client.sendRequest",
		trace.WithAttributes(
			attribute.String("a2a.method", method),
			attribute.String("a2a.request_id", id),
		))
	defer span.End()

	data, err := makeRequest(method, params, id)
	if err != nil {
		c.Logger.ErrorContext(ctx, "failed to create request", "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewBuffer(data))
	if err != nil {
		c.Logger.ErrorContext(ctx, "failed to create HTTP request", "error", err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.Logger.ErrorContext(ctx, "failed to send HTTP request", "error", err)
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.Logger.ErrorContext(ctx, "HTTP request failed", "status", resp.Status)
		return nil, fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	respBody := make([]byte, 0)
	buf := bytes.NewBuffer(respBody)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		c.Logger.ErrorContext(ctx, "failed to read response body", "error", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return buf.Bytes(), nil
}

// SendTask sends a task to an A2A server.
func (c *A2AClient) SendTask(ctx context.Context, req Request) (*SendTaskResponse, error) {
	ctx, span := c.Tracer.Start(ctx, "a2a.client.SendTask")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	sendReq, ok := req.(SendTaskRequest)
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
		Result  Task            `json:"result"`
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

	return &SendTaskResponse{
		Task:      jsonRPCResp.Result,
		RequestID: reqID,
	}, nil
}

// GetTask retrieves a task from an A2A server.
func (c *A2AClient) GetTask(ctx context.Context, req Request) (*GetTaskResponse, error) {
	ctx, span := c.Tracer.Start(ctx, "a2a.client.GetTask")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	getReq, ok := req.(GetTaskRequest)
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
		Result  Task            `json:"result"`
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

	return &GetTaskResponse{
		Task:      jsonRPCResp.Result,
		RequestID: getReq.RequestID,
	}, nil
}

// CancelTask cancels a task on an A2A server.
func (c *A2AClient) CancelTask(ctx context.Context, req Request) (*CancelTaskResponse, error) {
	ctx, span := c.Tracer.Start(ctx, "a2a.client.CancelTask")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	cancelReq, ok := req.(CancelTaskRequest)
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
		Result  Task            `json:"result"`
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

	return &CancelTaskResponse{
		Task:      jsonRPCResp.Result,
		RequestID: cancelReq.RequestID,
	}, nil
}

// SetTaskPushNotification configures push notification for a task.
func (c *A2AClient) SetTaskPushNotification(ctx context.Context, req Request) (*SetTaskPushNotificationResponse, error) {
	ctx, span := c.Tracer.Start(ctx, "a2a.client.SetTaskPushNotification")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	pushReq, ok := req.(SetTaskPushNotificationRequest)
	if !ok {
		return nil, fmt.Errorf("expected SetTaskPushNotificationRequest but got %T", req)
	}

	span.SetAttributes(attribute.String("a2a.task_id", pushReq.TaskID))

	params := SetTaskPushNotificationRequest{
		TaskID:                 pushReq.TaskID,
		PushNotificationConfig: pushReq.PushNotificationConfig,
	}

	responseData, err := c.sendRequest(ctx, "tasks/pushNotification/set", params, "")
	if err != nil {
		return nil, fmt.Errorf("failed to set task push notification: %w", err)
	}

	var jsonRPCResp struct {
		JsonRPC string                     `json:"jsonrpc"`
		Result  TaskPushNotificationConfig `json:"result"`
		Error   json.RawMessage            `json:"error"`
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

	return &SetTaskPushNotificationResponse{
		Config:    jsonRPCResp.Result,
		RequestID: "",
	}, nil
}

// GetTaskPushNotification retrieves push notification configuration for a task.
func (c *A2AClient) GetTaskPushNotification(ctx context.Context, req Request) (*GetTaskPushNotificationResponse, error) {
	ctx, span := c.Tracer.Start(ctx, "a2a.client.GetTaskPushNotification")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	getPushReq, ok := req.(GetTaskPushNotificationRequest)
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
		JsonRPC string                     `json:"jsonrpc"`
		Result  TaskPushNotificationConfig `json:"result"`
		Error   json.RawMessage            `json:"error"`
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

	return &GetTaskPushNotificationResponse{
		Config:    jsonRPCResp.Result,
		RequestID: getPushReq.RequestID,
	}, nil
}

// ResubscribeTask resubscribes to a task's updates.
func (c *A2AClient) ResubscribeTask(ctx context.Context, req Request) (*TaskStatusUpdateEvent, error) {
	ctx, span := c.Tracer.Start(ctx, "a2a.client.ResubscribeTask")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resubReq, ok := req.(TaskResubscriptionRequest)
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
		JsonRPC string                `json:"jsonrpc"`
		Result  TaskStatusUpdateEvent `json:"result"`
		Error   json.RawMessage       `json:"error"`
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
