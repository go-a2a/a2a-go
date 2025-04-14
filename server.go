// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bytedance/sonic"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// A2AServer is a server for the A2A protocol.
type A2AServer struct {
	// Host is the host to bind to.
	Host string

	// Port is the port to listen on.
	Port int

	// Endpoint is the endpoint to expose the API on.
	Endpoint string

	// TaskManager is the task manager to use.
	TaskManager TaskManager

	// AgentCard is the agent card for the server.
	AgentCard AgentCard

	// Logger is the logger to use.
	Logger *slog.Logger

	// Tracer is the tracer to use.
	Tracer trace.Tracer

	// Server is the HTTP server.
	Server *http.Server
}

// NewA2AServer creates a new A2AServer.
func NewA2AServer(host string, port int, endpoint string, agentCard AgentCard, taskManager TaskManager) *A2AServer {
	return &A2AServer{
		Host:        host,
		Port:        port,
		Endpoint:    endpoint,
		TaskManager: taskManager,
		AgentCard:   agentCard,
		Logger:      slog.Default(),
		Tracer:      otel.GetTracerProvider().Tracer("github.com/go-a2a/a2a"),
	}
}

// WithLogger sets the logger for the A2AServer.
func (s *A2AServer) WithLogger(logger *slog.Logger) *A2AServer {
	s.Logger = logger
	return s
}

// WithTracer sets the tracer for the A2AServer.
func (s *A2AServer) WithTracer(tracer trace.Tracer) *A2AServer {
	s.Tracer = tracer
	return s
}

// Start starts the A2AServer.
func (s *A2AServer) Start(ctx context.Context) error {
	if s.AgentCard.Name == "" || s.AgentCard.URL == "" || s.AgentCard.Version == "" {
		return fmt.Errorf("agent card must have name, URL, and version")
	}

	if s.TaskManager == nil {
		return fmt.Errorf("task manager cannot be nil")
	}

	mux := http.NewServeMux()

	// Main A2A endpoint
	mux.HandleFunc(s.Endpoint, s.processRequest)

	// Agent card endpoint - serves the agent's metadata
	mux.HandleFunc("/", s.handleAgentCardRequest)

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	s.Server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.Logger.InfoContext(ctx, "starting A2A server", "address", addr, "endpoint", s.Endpoint)

	go func() {
		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.Logger.ErrorContext(ctx, "server error", "error", err)
		}
	}()

	return nil
}

// Stop stops the A2AServer.
func (s *A2AServer) Stop(ctx context.Context) error {
	if s.Server == nil {
		return nil
	}

	return s.Server.Shutdown(ctx)
}

// handleAgentCardRequest handles requests for the agent card.
func (s *A2AServer) handleAgentCardRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	data, err := sonic.ConfigFastest.Marshal(s.AgentCard)
	if err != nil {
		s.Logger.Error("failed to marshal agent card", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		s.Logger.Error("failed to write response", "error", err)
	}
}

// processRequest is the main handler for the A2A API.
func (s *A2AServer) processRequest(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.Tracer.Start(r.Context(), "a2a.server.processRequest")
	defer span.End()

	if r.Method != http.MethodPost {
		s.writeError(w, -32600, "Method not allowed")
		return
	}

	// Parse request body
	var req struct {
		JsonRPC string          `json:"jsonrpc"`
		ID      string          `json:"id,omitempty"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, -32700, "Invalid JSON payload")
		return
	}

	span.SetAttributes(
		attribute.String("a2a.request_id", req.ID),
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
		result, err = s.handleSendTaskStreaming(ctx, w, req.ID, req.Params)
		return
	default:
		s.writeError(w, -32601, "Method not found")
		return
	}

	if err != nil {
		s.Logger.ErrorContext(ctx, "method execution failed", "method", req.Method, "error", err)
		s.writeError(w, -32603, fmt.Sprintf("Internal error: %v", err))
		return
	}

	// Write success response
	resp := struct {
		JsonRPC string `json:"jsonrpc"`
		ID      string `json:"id,omitempty"`
		Result  any    `json:"result"`
	}{
		JsonRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}

	data, err := sonic.ConfigFastest.Marshal(resp)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to marshal response", "error", err)
		s.writeError(w, -32603, "Internal error serializing response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to write response", "error", err)
	}
}

// writeError writes an error response in JSON-RPC format.
func (s *A2AServer) writeError(w http.ResponseWriter, code int, message string) {
	resp := struct {
		JsonRPC string `json:"jsonrpc"`
		ID      string `json:"id,omitempty"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{
		JsonRPC: "2.0",
		Error: struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	}

	data, err := sonic.ConfigFastest.Marshal(resp)
	if err != nil {
		s.Logger.Error("failed to marshal error response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Error is in the JSON-RPC response, not HTTP
	_, err = w.Write(data)
	if err != nil {
		s.Logger.Error("failed to write error response", "error", err)
	}
}

// handleSendTask handles the tasks/send method.
func (s *A2AServer) handleSendTask(ctx context.Context, params json.RawMessage) (*Task, error) {
	ctx, span := s.Tracer.Start(ctx, "a2a.server.handleSendTask")
	defer span.End()

	var taskParams Task
	if err := sonic.ConfigFastest.Unmarshal(params, &taskParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", taskParams.ID))

	task, err := s.TaskManager.OnSendTask(ctx, &taskParams)
	if err != nil {
		return nil, fmt.Errorf("failed to process task: %w", err)
	}

	return task, nil
}

// handleGetTask handles the tasks/get method.
func (s *A2AServer) handleGetTask(ctx context.Context, params json.RawMessage) (*Task, error) {
	ctx, span := s.Tracer.Start(ctx, "a2a.server.handleGetTask")
	defer span.End()

	var taskParams struct {
		ID string `json:"id"`
	}
	if err := sonic.ConfigFastest.Unmarshal(params, &taskParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", taskParams.ID))

	task, err := s.TaskManager.OnGetTask(ctx, taskParams.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

// handleCancelTask handles the tasks/cancel method.
func (s *A2AServer) handleCancelTask(ctx context.Context, params json.RawMessage) (*Task, error) {
	ctx, span := s.Tracer.Start(ctx, "a2a.server.handleCancelTask")
	defer span.End()

	var taskParams struct {
		ID string `json:"id"`
	}
	if err := sonic.ConfigFastest.Unmarshal(params, &taskParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", taskParams.ID))

	task, err := s.TaskManager.OnCancelTask(ctx, taskParams.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel task: %w", err)
	}

	return task, nil
}

// handleSetTaskPushNotification handles the tasks/pushNotification/set method.
func (s *A2AServer) handleSetTaskPushNotification(ctx context.Context, params json.RawMessage) (*TaskPushNotificationConfig, error) {
	ctx, span := s.Tracer.Start(ctx, "a2a.server.handleSetTaskPushNotification")
	defer span.End()

	var config TaskPushNotificationConfig
	if err := sonic.ConfigFastest.Unmarshal(params, &config); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", config.TaskID))

	result, err := s.TaskManager.OnSetTaskPushNotification(ctx, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to set push notification: %w", err)
	}

	return result, nil
}

// handleGetTaskPushNotification handles the tasks/pushNotification/get method.
func (s *A2AServer) handleGetTaskPushNotification(ctx context.Context, params json.RawMessage) (*TaskPushNotificationConfig, error) {
	ctx, span := s.Tracer.Start(ctx, "a2a.server.handleGetTaskPushNotification")
	defer span.End()

	var taskParams struct {
		ID string `json:"id"`
	}
	if err := sonic.ConfigFastest.Unmarshal(params, &taskParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", taskParams.ID))

	config, err := s.TaskManager.OnGetTaskPushNotification(ctx, taskParams.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get push notification: %w", err)
	}

	return config, nil
}

// handleTaskResubscription handles the tasks/resubscribe method.
func (s *A2AServer) handleTaskResubscription(ctx context.Context, params json.RawMessage) (*TaskStatusUpdateEvent, error) {
	ctx, span := s.Tracer.Start(ctx, "a2a.server.handleTaskResubscription")
	defer span.End()

	var taskParams struct {
		ID            string `json:"id"`
		HistoryLength int    `json:"historyLength,omitempty"`
	}
	if err := sonic.ConfigFastest.Unmarshal(params, &taskParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", taskParams.ID))

	event, err := s.TaskManager.OnResubscribeToTask(ctx, taskParams.ID, taskParams.HistoryLength)
	if err != nil {
		return nil, fmt.Errorf("failed to resubscribe to task: %w", err)
	}

	return event, nil
}

// handleSendTaskStreaming handles the tasks/sendSubscribe method.
func (s *A2AServer) handleSendTaskStreaming(ctx context.Context, w http.ResponseWriter, id string, params json.RawMessage) (*TaskStatusUpdateEvent, error) {
	ctx, span := s.Tracer.Start(ctx, "a2a.server.handleSendTaskStreaming")
	defer span.End()

	var taskParams Task
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
	eventCh, err := s.TaskManager.OnSendTaskSubscribe(ctx, taskParams)
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
		resp := struct {
			JsonRPC string `json:"jsonrpc"`
			ID      string `json:"id,omitempty"`
			Result  any    `json:"result"`
		}{
			JsonRPC: "2.0",
			ID:      id,
			Result:  event,
		}

		data, err := sonic.ConfigFastest.Marshal(resp)
		if err != nil {
			s.Logger.ErrorContext(ctx, "failed to marshal event", "error", err)
			continue
		}

		// Write event to the client
		_, err = fmt.Fprintf(w, "data: %s\n\n", data)
		if err != nil {
			s.Logger.ErrorContext(ctx, "failed to write event", "error", err)
			break
		}

		flusher.Flush()
	}

	return nil, nil
}

// HandleSendTask processes a send task request.
func (s *A2AServer) HandleSendTask(ctx context.Context, req Request) (*SendTaskResponse, error) {
	ctx, span := s.Tracer.Start(ctx, "a2a.server.HandleSendTask")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	sendReq, ok := req.(SendTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected SendTaskRequest but got %T", req)
	}

	task, err := s.TaskManager.OnSendTask(ctx, &sendReq.Task)
	if err != nil {
		return nil, fmt.Errorf("failed to process task: %w", err)
	}

	return &SendTaskResponse{
		Task:      *task,
		RequestID: task.ID,
	}, nil
}

// HandleGetTask processes a get task request.
func (s *A2AServer) HandleGetTask(ctx context.Context, req Request) (*GetTaskResponse, error) {
	ctx, span := s.Tracer.Start(ctx, "a2a.server.HandleGetTask")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	getReq, ok := req.(GetTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected GetTaskRequest but got %T", req)
	}

	task, err := s.TaskManager.OnGetTask(ctx, getReq.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &GetTaskResponse{
		Task:      *task,
		RequestID: getReq.RequestID,
	}, nil
}

// HandleCancelTask processes a cancel task request.
func (s *A2AServer) HandleCancelTask(ctx context.Context, req Request) (*CancelTaskResponse, error) {
	ctx, span := s.Tracer.Start(ctx, "a2a.server.HandleCancelTask")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	cancelReq, ok := req.(CancelTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected CancelTaskRequest but got %T", req)
	}

	task, err := s.TaskManager.OnCancelTask(ctx, cancelReq.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel task: %w", err)
	}

	return &CancelTaskResponse{
		Task:      *task,
		RequestID: cancelReq.RequestID,
	}, nil
}

// HandleSetTaskPushNotification processes a set task push notification request.
func (s *A2AServer) HandleSetTaskPushNotification(ctx context.Context, req Request) (*SetTaskPushNotificationResponse, error) {
	ctx, span := s.Tracer.Start(ctx, "a2a.server.HandleSetTaskPushNotification")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	pushReq, ok := req.(*SetTaskPushNotificationRequest)
	if !ok {
		return nil, fmt.Errorf("expected SetTaskPushNotificationRequest but got %T", req)
	}

	config := (*TaskPushNotificationConfig)(pushReq)
	result, err := s.TaskManager.OnSetTaskPushNotification(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to set push notification: %w", err)
	}

	return &SetTaskPushNotificationResponse{
		Config:    *result,
		RequestID: "",
	}, nil
}

// HandleGetTaskPushNotification processes a get task push notification request.
func (s *A2AServer) HandleGetTaskPushNotification(ctx context.Context, req Request) (*GetTaskPushNotificationResponse, error) {
	ctx, span := s.Tracer.Start(ctx, "a2a.server.HandleGetTaskPushNotification")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	getPushReq, ok := req.(GetTaskPushNotificationRequest)
	if !ok {
		return nil, fmt.Errorf("expected GetTaskPushNotificationRequest but got %T", req)
	}

	config, err := s.TaskManager.OnGetTaskPushNotification(ctx, getPushReq.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get push notification: %w", err)
	}

	return &GetTaskPushNotificationResponse{
		Config:    *config,
		RequestID: getPushReq.RequestID,
	}, nil
}

// HandleTaskResubscription processes a task resubscription request.
func (s *A2AServer) HandleTaskResubscription(ctx context.Context, req Request) (*TaskStatusUpdateEvent, error) {
	ctx, span := s.Tracer.Start(ctx, "a2a.server.HandleTaskResubscription")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resubReq, ok := req.(TaskResubscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("expected TaskResubscriptionRequest but got %T", req)
	}

	event, err := s.TaskManager.OnResubscribeToTask(ctx, resubReq.TaskID, resubReq.HistoryLength)
	if err != nil {
		return nil, fmt.Errorf("failed to resubscribe to task: %w", err)
	}

	return event, nil
}
