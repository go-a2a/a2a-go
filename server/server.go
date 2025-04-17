// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	"github.com/bytedance/sonic"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-a2a/a2a"
)

// Server represents an A2A server.
type Server interface {
	HandleSendTask(ctx context.Context, req a2a.A2ARequest) (*a2a.SendTaskResponse, error)
	HandleGetTask(ctx context.Context, req a2a.A2ARequest) (*a2a.GetTaskResponse, error)
	HandleCancelTask(ctx context.Context, req a2a.A2ARequest) (*a2a.CancelTaskResponse, error)
	HandleSetTaskPushNotification(ctx context.Context, req a2a.A2ARequest) (*a2a.SetTaskPushNotificationResponse, error)
	HandleGetTaskPushNotification(ctx context.Context, req a2a.A2ARequest) (*a2a.GetTaskPushNotificationResponse, error)
	HandleTaskResubscription(ctx context.Context, req a2a.A2ARequest) (*a2a.TaskStatusUpdateEvent, error)
}

// A2AServer is a server for the A2A protocol.
type A2AServer struct {
	// host is the host to bind to.
	host string

	// port is the port to listen on.
	port int

	// endpoint is the endpoint to expose the API on.
	endpoint string

	// agentCard is the agent card for the server.
	agentCard a2a.AgentCard

	// taskManager is the task manager to use.
	taskManager TaskManager

	// logger is the logger to use.
	logger *slog.Logger

	// tracer is the tracer to use.
	tracer trace.Tracer

	// server is the HTTP server.
	server *http.Server
}

var _ Server = (*A2AServer)(nil)

// ServerOption represents an option for configuring the [A2AServer].
type ServerOption func(*A2AServer)

// WithLogger sets the logger for the A2AServer.
func (s *A2AServer) WithLogger(logger *slog.Logger) ServerOption {
	return func(s *A2AServer) {
		s.logger = logger
	}
}

// WithTracer sets the tracer for the A2AServer.
func (s *A2AServer) WithTracer(tracer trace.Tracer) ServerOption {
	return func(s *A2AServer) {
		s.tracer = tracer
	}
}

// NewServer creates a new [A2AServer].
func NewServer(host string, port int, endpoint string, agentCard a2a.AgentCard, taskManager TaskManager, opts ...ServerOption) Server {
	s := &A2AServer{
		host:        host,
		port:        port,
		endpoint:    endpoint,
		agentCard:   agentCard,
		taskManager: taskManager,
		logger:      slog.Default(),
		tracer:      otel.GetTracerProvider().Tracer("github.com/go-a2a/a2a/server"),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Start starts the [Server].
func (s *A2AServer) Start(ctx context.Context) error {
	if s.agentCard.Name == "" || s.agentCard.URL == "" || s.agentCard.Version == "" {
		return fmt.Errorf("agent card must have name, URL, and version")
	}

	if s.taskManager == nil {
		return fmt.Errorf("task manager cannot be nil")
	}

	mux := http.NewServeMux()
	// Agent card endpoint - serves the agent's metadata
	mux.HandleFunc("GET /.well-known/agent.json", s.handleAgentCardRequest)
	// Main A2A endpoint
	mux.HandleFunc("POST "+s.endpoint, s.processRequest)

	addr := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.logger.InfoContext(ctx, "starting A2A server", "address", addr, "endpoint", s.endpoint)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.ErrorContext(ctx, "server error", "error", err)
		}
	}()

	return nil
}

// Stop stops the A2AServer.
func (s *A2AServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	return s.server.Shutdown(ctx)
}

// handleAgentCardRequest handles requests for the agent card.
func (s *A2AServer) handleAgentCardRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	data, err := sonic.ConfigFastest.Marshal(s.agentCard)
	if err != nil {
		s.logger.Error("failed to marshal agent card", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		s.logger.Error("failed to write response", "error", err)
	}
}

// processRequest is the main handler for the A2A API.
func (s *A2AServer) processRequest(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "a2a.server.processRequest")
	defer span.End()

	if r.Method != http.MethodPost {
		s.writeError(w, -32600, "method not allowed")
		return
	}

	var req a2a.JSONRPCRequest
	if err := sonic.ConfigFastest.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, -32700, "Invalid JSON payload")
		return
	}

	span.SetAttributes(
		attribute.Stringer("a2a.request_id", req.ID),
		attribute.String("a2a.method", req.Method),
	)

	// Handle method
	var (
		result any
		err    error
	)

	switch req.Method {
	case "tasks/send":
		result, err = s.handleSendTask(ctx, req.Params)
	case "tasks/get":
		result, err = s.handleGetTask(ctx, req.Params)
	case "tasks/cancel":
		result, err = s.handleCancelTask(ctx, req.Params)
	case "tasks/pushNotification/set":
		result, err = s.handleSetTaskPushNotification(ctx, req.Params)
	case "tasks/pushNotification/get":
		result, err = s.handleGetTaskPushNotification(ctx, req.Params)
	case "tasks/resubscribe":
		result, err = s.handleTaskResubscription(ctx, req.Params)
	case "tasks/sendSubscribe":
		// TODO(zchee): handleSendTaskStreaming writes the response directly
		result, err = s.handleSendTaskStreaming(ctx, w, req.ID.String(), req.Params)
		return
	default:
		s.writeError(w, -32601, "Method not found")
		return
	}

	if err != nil {
		s.logger.ErrorContext(ctx, "method execution failed", "method", req.Method, "error", err)
		s.writeError(w, -32603, fmt.Sprintf("Internal error: %v", err))
		return
	}

	// Write success response
	resp := a2a.JSONRPCResponse{
		JSONRPCMessage: a2a.NewJSONRPCMessage(req.ID),
		Result:         result,
	}

	data, err := sonic.ConfigFastest.Marshal(resp)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to marshal response", "error", err)
		s.writeError(w, -32603, "Internal error serializing response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to write response", "error", err)
	}
}

// writeError writes an error response in JSON-RPC format.
func (s *A2AServer) writeError(w http.ResponseWriter, code int, message string) {
	resp := a2a.JSONRPCResponse{
		JSONRPCMessage: a2a.JSONRPCMessage{
			JSONRPC: "2.0",
		},
		Error: &a2a.JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
	data, err := sonic.ConfigFastest.Marshal(resp)
	if err != nil {
		s.logger.Error("failed to marshal error response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Error is in the JSON-RPC response, not HTTP
	_, err = w.Write(data)
	if err != nil {
		s.logger.Error("failed to write error response", "error", err)
	}
}

// handleSendTask handles the tasks/send method.
func (s *A2AServer) handleSendTask(ctx context.Context, params json.RawMessage) (*a2a.Task, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.handleSendTask")
	defer span.End()

	var taskParams a2a.Task
	if err := sonic.ConfigFastest.Unmarshal(params, &taskParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", taskParams.ID))

	task, err := s.taskManager.OnSendTask(ctx, &taskParams)
	if err != nil {
		return nil, fmt.Errorf("failed to process task: %w", err)
	}

	return task, nil
}

// handleGetTask handles the tasks/get method.
func (s *A2AServer) handleGetTask(ctx context.Context, params json.RawMessage) (*a2a.Task, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.handleGetTask")
	defer span.End()

	var taskParams struct {
		ID string `json:"id"`
	}
	if err := sonic.ConfigFastest.Unmarshal(params, &taskParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", taskParams.ID))

	task, err := s.taskManager.OnGetTask(ctx, taskParams.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

// handleCancelTask handles the tasks/cancel method.
func (s *A2AServer) handleCancelTask(ctx context.Context, params json.RawMessage) (*a2a.Task, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.handleCancelTask")
	defer span.End()

	var taskParams struct {
		ID string `json:"id"`
	}
	if err := sonic.ConfigFastest.Unmarshal(params, &taskParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", taskParams.ID))

	task, err := s.taskManager.OnCancelTask(ctx, taskParams.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel task: %w", err)
	}

	return task, nil
}

// handleSetTaskPushNotification handles the tasks/pushNotification/set method.
func (s *A2AServer) handleSetTaskPushNotification(ctx context.Context, params json.RawMessage) (*a2a.TaskPushNotificationConfig, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.handleSetTaskPushNotification")
	defer span.End()

	var config a2a.TaskPushNotificationConfig
	if err := sonic.ConfigFastest.Unmarshal(params, &config); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", config.ID))

	result, err := s.taskManager.OnSetTaskPushNotification(ctx, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to set push notification: %w", err)
	}

	return result, nil
}

// handleGetTaskPushNotification handles the tasks/pushNotification/get method.
func (s *A2AServer) handleGetTaskPushNotification(ctx context.Context, params json.RawMessage) (*a2a.TaskPushNotificationConfig, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.handleGetTaskPushNotification")
	defer span.End()

	var taskParams struct {
		ID string `json:"id"`
	}
	if err := sonic.ConfigFastest.Unmarshal(params, &taskParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", taskParams.ID))

	config, err := s.taskManager.OnGetTaskPushNotification(ctx, taskParams.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get push notification: %w", err)
	}

	return config, nil
}

// handleTaskResubscription handles the tasks/resubscribe method.
func (s *A2AServer) handleTaskResubscription(ctx context.Context, params json.RawMessage) (*a2a.TaskStatusUpdateEvent, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.handleTaskResubscription")
	defer span.End()

	var taskParams struct {
		ID            string `json:"id"`
		HistoryLength int    `json:"historyLength,omitempty"`
	}
	if err := sonic.ConfigFastest.Unmarshal(params, &taskParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", taskParams.ID))

	event, err := s.taskManager.OnResubscribeToTask(ctx, taskParams.ID, taskParams.HistoryLength)
	if err != nil {
		return nil, fmt.Errorf("failed to resubscribe to task: %w", err)
	}

	return event, nil
}

// handleSendTaskStreaming handles the tasks/sendSubscribe method.
func (s *A2AServer) handleSendTaskStreaming(ctx context.Context, w http.ResponseWriter, id string, params json.RawMessage) (*a2a.TaskStatusUpdateEvent, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.handleSendTaskStreaming")
	defer span.End()

	var taskParams a2a.Task
	if err := sonic.ConfigFastest.Unmarshal(params, &taskParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", taskParams.ID))

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	// Create a channel for the task events
	eventCh, err := s.taskManager.OnSendTaskSubscribe(ctx, taskParams)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to task: %w", err)
	}

	// Flush the headers to the client
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported by response writer")
	}

	// Begin streaming events
	for event := range eventCh {
		// Marshal event to JSON
		resp := a2a.JSONRPCResponse{
			JSONRPCMessage: a2a.NewJSONRPCMessage(id),
			Result:         event,
		}

		data, err := sonic.ConfigFastest.Marshal(resp)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to marshal event", "error", err)
			continue
		}

		// Write event to the client
		_, err = fmt.Fprintf(w, "data: %s\n\n", data)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to write event", "error", err)
			break
		}

		flusher.Flush()
	}

	return nil, nil
}

// HandleSendTask processes a send task request.
func (s *A2AServer) HandleSendTask(ctx context.Context, req a2a.A2ARequest) (*a2a.SendTaskResponse, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.HandleSendTask")
	defer span.End()

	sendReq, ok := req.(*a2a.SendTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected SendTaskRequest but got %T", req)
	}

	task, err := s.taskManager.OnSendTask(ctx, sendReq.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to process task: %w", err)
	}

	return &a2a.SendTaskResponse{
		JSONRPCMessage: a2a.NewJSONRPCMessage(sendReq.ID),
		Result:         task,
	}, nil
}

// HandleGetTask processes a get task request.
func (s *A2AServer) HandleGetTask(ctx context.Context, req a2a.A2ARequest) (*a2a.GetTaskResponse, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.HandleGetTask")
	defer span.End()

	getReq, ok := req.(a2a.GetTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected GetTaskRequest but got %T", req)
	}

	task, err := s.taskManager.OnGetTask(ctx, getReq.Params.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &a2a.GetTaskResponse{
		JSONRPCMessage: a2a.NewJSONRPCMessage(getReq.ID),
		Result:         task,
	}, nil
}

// HandleCancelTask processes a cancel task request.
func (s *A2AServer) HandleCancelTask(ctx context.Context, req a2a.A2ARequest) (*a2a.CancelTaskResponse, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.HandleCancelTask")
	defer span.End()

	cancelReq, ok := req.(a2a.CancelTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected CancelTaskRequest but got %T", req)
	}

	task, err := s.taskManager.OnCancelTask(ctx, cancelReq.Params.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel task: %w", err)
	}

	return &a2a.CancelTaskResponse{
		JSONRPCMessage: a2a.NewJSONRPCMessage(cancelReq.ID),
		Result:         task,
	}, nil
}

// HandleSetTaskPushNotification processes a set task push notification request.
func (s *A2AServer) HandleSetTaskPushNotification(ctx context.Context, req a2a.A2ARequest) (*a2a.SetTaskPushNotificationResponse, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.HandleSetTaskPushNotification")
	defer span.End()

	pushReq, ok := req.(*a2a.SetTaskPushNotificationRequest)
	if !ok {
		return nil, fmt.Errorf("expected SetTaskPushNotificationRequest but got %T", req)
	}

	result, err := s.taskManager.OnSetTaskPushNotification(ctx, &pushReq.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to set push notification: %w", err)
	}

	return &a2a.SetTaskPushNotificationResponse{
		Result: result,
	}, nil
}

// HandleGetTaskPushNotification processes a get task push notification request.
func (s *A2AServer) HandleGetTaskPushNotification(ctx context.Context, req a2a.A2ARequest) (*a2a.GetTaskPushNotificationResponse, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.HandleGetTaskPushNotification")
	defer span.End()

	getPushReq, ok := req.(a2a.GetTaskPushNotificationRequest)
	if !ok {
		return nil, fmt.Errorf("expected GetTaskPushNotificationRequest but got %T", req)
	}

	config, err := s.taskManager.OnGetTaskPushNotification(ctx, getPushReq.Params.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get push notification: %w", err)
	}

	return &a2a.GetTaskPushNotificationResponse{
		JSONRPCMessage: a2a.NewJSONRPCMessage(getPushReq.ID),
		Result:         config,
	}, nil
}

// HandleTaskResubscription processes a task resubscription request.
func (s *A2AServer) HandleTaskResubscription(ctx context.Context, req a2a.A2ARequest) (*a2a.TaskStatusUpdateEvent, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.HandleTaskResubscription")
	defer span.End()

	resubReq, ok := req.(a2a.TaskResubscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("expected TaskResubscriptionRequest but got %T", req)
	}

	event, err := s.taskManager.OnResubscribeToTask(ctx, resubReq.Params.ID, resubReq.Params.HistoryLength)
	if err != nil {
		return nil, fmt.Errorf("failed to resubscribe to task: %w", err)
	}

	return event, nil
}
