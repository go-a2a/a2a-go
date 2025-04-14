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

// Server is a server for the A2A protocol.
type Server struct {
	// host is the host to bind to.
	host string

	// port is the port to listen on.
	port int

	// endpoint is the endpoint to expose the API on.
	endpoint string

	// taskManager is the task manager to use.
	taskManager TaskManager

	// agentCard is the agent card for the server.
	agentCard a2a.AgentCard

	// logger is the logger to use.
	logger *slog.Logger

	// tracer is the tracer to use.
	tracer trace.Tracer

	// server is the HTTP server.
	server *http.Server
}

var _ a2a.Server = (*Server)(nil)

// NewServer creates a new [Server].
func NewServer(host string, port int, endpoint string, agentCard a2a.AgentCard, taskManager TaskManager) *Server {
	return &Server{
		host:        host,
		port:        port,
		endpoint:    endpoint,
		taskManager: taskManager,
		agentCard:   agentCard,
		logger:      slog.Default(),
		tracer:      otel.GetTracerProvider().Tracer("github.com/go-a2a/a2a"),
	}
}

// WithLogger sets the logger for the A2AServer.
func (s *Server) WithLogger(logger *slog.Logger) *Server {
	s.logger = logger
	return s
}

// WithTracer sets the tracer for the A2AServer.
func (s *Server) WithTracer(tracer trace.Tracer) *Server {
	s.tracer = tracer
	return s
}

// Start starts the A2AServer.
func (s *Server) Start(ctx context.Context) error {
	if s.agentCard.Name == "" || s.agentCard.URL == "" || s.agentCard.Version == "" {
		return fmt.Errorf("agent card must have name, URL, and version")
	}

	if s.taskManager == nil {
		return fmt.Errorf("task manager cannot be nil")
	}

	mux := http.NewServeMux()

	// Main A2A endpoint
	mux.HandleFunc(s.endpoint, s.processRequest)

	// Agent card endpoint - serves the agent's metadata
	mux.HandleFunc("/", s.handleAgentCardRequest)

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
func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	return s.server.Shutdown(ctx)
}

// handleAgentCardRequest handles requests for the agent card.
func (s *Server) handleAgentCardRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
func (s *Server) processRequest(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "a2a.server.processRequest")
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
		s.logger.ErrorContext(ctx, "method execution failed", "method", req.Method, "error", err)
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
func (s *Server) writeError(w http.ResponseWriter, code int, message string) {
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
func (s *Server) handleSendTask(ctx context.Context, params json.RawMessage) (*a2a.Task, error) {
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
func (s *Server) handleGetTask(ctx context.Context, params json.RawMessage) (*a2a.Task, error) {
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
func (s *Server) handleCancelTask(ctx context.Context, params json.RawMessage) (*a2a.Task, error) {
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
func (s *Server) handleSetTaskPushNotification(ctx context.Context, params json.RawMessage) (*a2a.TaskPushNotificationConfig, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.handleSetTaskPushNotification")
	defer span.End()

	var config a2a.TaskPushNotificationConfig
	if err := sonic.ConfigFastest.Unmarshal(params, &config); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", config.TaskID))

	result, err := s.taskManager.OnSetTaskPushNotification(ctx, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to set push notification: %w", err)
	}

	return result, nil
}

// handleGetTaskPushNotification handles the tasks/pushNotification/get method.
func (s *Server) handleGetTaskPushNotification(ctx context.Context, params json.RawMessage) (*a2a.TaskPushNotificationConfig, error) {
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
func (s *Server) handleTaskResubscription(ctx context.Context, params json.RawMessage) (*a2a.TaskStatusUpdateEvent, error) {
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
func (s *Server) handleSendTaskStreaming(ctx context.Context, w http.ResponseWriter, id string, params json.RawMessage) (*a2a.TaskStatusUpdateEvent, error) {
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
func (s *Server) HandleSendTask(ctx context.Context, req a2a.Request) (*a2a.SendTaskResponse, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.HandleSendTask")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	sendReq, ok := req.(a2a.SendTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected SendTaskRequest but got %T", req)
	}

	task, err := s.taskManager.OnSendTask(ctx, &sendReq.Task)
	if err != nil {
		return nil, fmt.Errorf("failed to process task: %w", err)
	}

	return &a2a.SendTaskResponse{
		Task:      *task,
		RequestID: task.ID,
	}, nil
}

// HandleGetTask processes a get task request.
func (s *Server) HandleGetTask(ctx context.Context, req a2a.Request) (*a2a.GetTaskResponse, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.HandleGetTask")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	getReq, ok := req.(a2a.GetTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected GetTaskRequest but got %T", req)
	}

	task, err := s.taskManager.OnGetTask(ctx, getReq.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &a2a.GetTaskResponse{
		Task:      *task,
		RequestID: getReq.RequestID,
	}, nil
}

// HandleCancelTask processes a cancel task request.
func (s *Server) HandleCancelTask(ctx context.Context, req a2a.Request) (*a2a.CancelTaskResponse, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.HandleCancelTask")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	cancelReq, ok := req.(a2a.CancelTaskRequest)
	if !ok {
		return nil, fmt.Errorf("expected CancelTaskRequest but got %T", req)
	}

	task, err := s.taskManager.OnCancelTask(ctx, cancelReq.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel task: %w", err)
	}

	return &a2a.CancelTaskResponse{
		Task:      *task,
		RequestID: cancelReq.RequestID,
	}, nil
}

// HandleSetTaskPushNotification processes a set task push notification request.
func (s *Server) HandleSetTaskPushNotification(ctx context.Context, req a2a.Request) (*a2a.SetTaskPushNotificationResponse, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.HandleSetTaskPushNotification")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	pushReq, ok := req.(*a2a.SetTaskPushNotificationRequest)
	if !ok {
		return nil, fmt.Errorf("expected SetTaskPushNotificationRequest but got %T", req)
	}

	config := (*a2a.TaskPushNotificationConfig)(pushReq)
	result, err := s.taskManager.OnSetTaskPushNotification(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to set push notification: %w", err)
	}

	return &a2a.SetTaskPushNotificationResponse{
		Config:    *result,
		RequestID: "",
	}, nil
}

// HandleGetTaskPushNotification processes a get task push notification request.
func (s *Server) HandleGetTaskPushNotification(ctx context.Context, req a2a.Request) (*a2a.GetTaskPushNotificationResponse, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.HandleGetTaskPushNotification")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	getPushReq, ok := req.(a2a.GetTaskPushNotificationRequest)
	if !ok {
		return nil, fmt.Errorf("expected GetTaskPushNotificationRequest but got %T", req)
	}

	config, err := s.taskManager.OnGetTaskPushNotification(ctx, getPushReq.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get push notification: %w", err)
	}

	return &a2a.GetTaskPushNotificationResponse{
		Config:    *config,
		RequestID: getPushReq.RequestID,
	}, nil
}

// HandleTaskResubscription processes a task resubscription request.
func (s *Server) HandleTaskResubscription(ctx context.Context, req a2a.Request) (*a2a.TaskStatusUpdateEvent, error) {
	ctx, span := s.tracer.Start(ctx, "a2a.server.HandleTaskResubscription")
	defer span.End()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resubReq, ok := req.(a2a.TaskResubscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("expected TaskResubscriptionRequest but got %T", req)
	}

	event, err := s.taskManager.OnResubscribeToTask(ctx, resubReq.TaskID, resubReq.HistoryLength)
	if err != nil {
		return nil, fmt.Errorf("failed to resubscribe to task: %w", err)
	}

	return event, nil
}
