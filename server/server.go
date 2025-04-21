// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package server provides an implementation of the A2A protocol server.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/bytedance/sonic"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-a2a/a2a"
)

// Server represents an A2A server that handles incoming requests and manages tasks.
type Server struct {
	// server is the HTTP server.
	server *http.Server

	// endpoint is the endpoint to expose the API on.
	endpoint string

	// agentCard is the agent card for the server.
	agentCard *a2a.AgentCard

	// taskManager is the task manager to use.
	taskManager TaskManager

	// logger is the logger to use.
	logger *slog.Logger

	// tracer is the tracer to use.
	tracer trace.Tracer
}

// NewServer creates a new [Server].
func NewServer(host, port string, agentCard *a2a.AgentCard, taskManager TaskManager, opts ...ServerOption) *Server {
	s := &Server{
		endpoint:    "/",
		agentCard:   agentCard,
		taskManager: taskManager,
		logger:      slog.Default(),
		tracer:      otel.GetTracerProvider().Tracer("github.com/go-a2a/a2a/server"),
	}
	for _, opt := range opts {
		opt(s)
	}

	mux := http.NewServeMux()
	// agent card endpoint - serves the agent's metadata
	mux.HandleFunc("GET /.well-known/agent.json", s.handleAgentCardRequest)
	// main A2A endpoint
	mux.HandleFunc("POST "+s.endpoint, s.requestHandler)

	s.server = &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: mux,
	}

	return s
}

// Start starts the [Server].
func (s *Server) Start(ctx context.Context) error {
	if s.agentCard.Name == "" || s.agentCard.URL == "" || s.agentCard.Version == "" {
		return errors.New("agent card must have name, URL, and version")
	}
	if s.taskManager == nil {
		return errors.New("task manager cannot be nil")
	}

	s.logger.InfoContext(ctx, "starting A2A server", "address", s.server.Addr, "endpoint", s.endpoint)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.ErrorContext(ctx, "server error", slog.Any("error", err))
		}
	}()

	return nil
}

// Shutdown shutdowns the [Server].
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// handleAgentCardRequest handles requests for the agent card.
func (s *Server) handleAgentCardRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	data, err := sonic.ConfigFastest.Marshal(s.agentCard)
	if err != nil {
		s.logger.Error("marshal agent card", slog.Any("error", err))
		http.Error(w, "unable to marshal agent card", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		s.logger.Error("unable to write response", slog.Any("error", err))
	}
}

// requestHandler is the main handler for the A2A API.
func (s *Server) requestHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "server.requestHandler")
	defer span.End()

	if r.Method != http.MethodPost {
		span.SetAttributes(semconv.RPCJsonrpcErrorCode(a2a.InvalidRequestErrorCode))

		s.writeError(ctx, w, a2a.InvalidRequestErrorCode, "method not allowed")
		return
	}

	var req a2a.JSONRPCRequest
	if err := sonic.ConfigFastest.NewDecoder(r.Body).Decode(&req); err != nil {
		span.SetAttributes(semconv.RPCJsonrpcErrorCode(a2a.JSONParseErrorCode))
		span.SetStatus(codes.Error, err.Error())

		s.writeError(ctx, w, a2a.JSONParseErrorCode, "Invalid JSON payload")
		return
	}

	span.SetAttributes(
		semconv.RPCJsonrpcVersion(req.JSONRPC),
		semconv.RPCJsonrpcRequestID(req.ID.String()),
		attribute.Stringer("a2a.request_id", req.ID),
		attribute.String("a2a.method", req.Method),
	)

	// Handle method
	var (
		result any
		err    error
	)
	switch req.Method {
	case a2a.MethodTasksSend:
		result, err = s.handleSendTask(ctx, req.Params)
	case a2a.MethodTasksGet:
		result, err = s.handleGetTask(ctx, req.Params)
	case a2a.MethodTasksCancel:
		result, err = s.handleCancelTask(ctx, req.Params)
	case a2a.MethodTasksPushNotificationSet:
		result, err = s.handleSetTaskPushNotification(ctx, req.Params)
	case a2a.MethodTasksPushNotificationGet:
		result, err = s.handleGetTaskPushNotification(ctx, req.Params)
	case a2a.MethodTasksSendSubscribe:
		// TODO(zchee): handleSendTaskStreaming writes the response directly
		_, err = s.handleSendTaskStreaming(ctx, w, req.ID, req.Params)
		if err != nil {
			s.logger.ErrorContext(ctx, "handle SendTaskStreaming", slog.Any("error", err))
			span.SetAttributes(semconv.RPCJsonrpcErrorCode(a2a.InternalErrorCode))
			span.SetStatus(codes.Error, err.Error())
			s.writeError(ctx, w, a2a.InternalErrorCode, "Internal error: handle SendTaskStreaming")
		}
		return
	case a2a.MethodTasksResubscribe:
		// TODO(zchee): handleTaskResubscription writes the response directly
		_, err = s.handleTaskResubscription(ctx, w, req.ID, req.Params)
		if err != nil {
			s.logger.ErrorContext(ctx, "handle TaskResubscription", slog.Any("error", err))
			s.writeError(ctx, w, a2a.InternalErrorCode, "Internal error: handle TaskResubscription")
		}
		return
	default:
		s.writeError(ctx, w, a2a.MethodNotFoundErrorCode, "Method not found")
		return
	}

	if err != nil {
		s.logger.ErrorContext(ctx, "method execution failed", slog.String("method", req.Method), slog.Any("error", err))
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Sprintf("Internal error: %v", err))
		return
	}

	// Write success response
	resp := a2a.JSONRPCResponse{
		JSONRPCMessage: a2a.NewJSONRPCMessage(req.ID),
		Result:         result,
	}

	data, err := sonic.ConfigFastest.Marshal(&resp)
	if err != nil {
		s.logger.ErrorContext(ctx, "marshal response", slog.Any("error", err))
		s.writeError(ctx, w, a2a.InternalErrorCode, "Internal error: marshal response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		s.logger.ErrorContext(ctx, "write response", slog.Any("error", err))
	}
}

// writeError writes an error response in JSON-RPC format.
func (s *Server) writeError(ctx context.Context, w http.ResponseWriter, code int, message string) {
	span := trace.SpanFromContext(ctx)
	_ = span

	resp := &a2a.JSONRPCResponse{
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
		s.logger.Error("marshal error response", slog.Any("error", err))
		http.Error(w, "Marshal error response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // error is in the JSON-RPC response, not HTTP
	_, err = w.Write(data)
	if err != nil {
		s.logger.Error("write error response", slog.Any("error", err))
	}
}

// handleSendTask handles the tasks/send method.
func (s *Server) handleSendTask(ctx context.Context, params json.RawMessage) (*a2a.Task, error) {
	ctx, span := s.tracer.Start(ctx, "server.handleSendTask")
	defer span.End()

	var req a2a.SendTaskRequest
	if err := sonic.ConfigFastest.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	resp, err := s.taskManager.OnSendTask(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("process task: %w", err)
	}

	return resp.Result, nil
}

// handleGetTask handles the tasks/get method.
func (s *Server) handleGetTask(ctx context.Context, params json.RawMessage) (*a2a.Task, error) {
	ctx, span := s.tracer.Start(ctx, "server.handleGetTask")
	defer span.End()

	var req a2a.GetTaskRequest
	if err := sonic.ConfigFastest.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	resp, err := s.taskManager.OnGetTask(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	return resp.Result, nil
}

// handleCancelTask handles the tasks/cancel method.
func (s *Server) handleCancelTask(ctx context.Context, params json.RawMessage) (*a2a.Task, error) {
	ctx, span := s.tracer.Start(ctx, "server.handleCancelTask")
	defer span.End()

	var req a2a.CancelTaskRequest
	if err := sonic.ConfigFastest.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	resp, err := s.taskManager.OnCancelTask(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("cancel task: %w", err)
	}

	return resp.Result, nil
}

// handleSetTaskPushNotification handles the tasks/pushNotification/set method.
func (s *Server) handleSetTaskPushNotification(ctx context.Context, params json.RawMessage) (*a2a.TaskPushNotificationConfig, error) {
	ctx, span := s.tracer.Start(ctx, "server.handleSetTaskPushNotification")
	defer span.End()

	var req a2a.SetTaskPushNotificationRequest
	if err := sonic.ConfigFastest.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	resp, err := s.taskManager.OnSetTaskPushNotification(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("set push notification: %w", err)
	}

	return resp.Result, nil
}

// handleGetTaskPushNotification handles the tasks/pushNotification/get method.
func (s *Server) handleGetTaskPushNotification(ctx context.Context, params json.RawMessage) (*a2a.TaskPushNotificationConfig, error) {
	ctx, span := s.tracer.Start(ctx, "server.handleGetTaskPushNotification")
	defer span.End()

	var req a2a.GetTaskPushNotificationRequest
	if err := sonic.ConfigFastest.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	resp, err := s.taskManager.OnGetTaskPushNotification(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("get push notification: %w", err)
	}

	return resp.Result, nil
}

// handleSendTaskStreaming handles the tasks/sendSubscribe method.
func (s *Server) handleSendTaskStreaming(ctx context.Context, w http.ResponseWriter, id a2a.ID, params json.RawMessage) (*a2a.TaskStatusUpdateEvent, error) {
	ctx, span := s.tracer.Start(ctx, "server.handleSendTaskStreaming")
	defer span.End()

	var req a2a.SendTaskStreamingRequest
	if err := sonic.ConfigFastest.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	// Create a channel for the task events
	eventsCh, err := s.taskManager.OnSendTaskSubscribe(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to task: %w", err)
	}

	// Flush the headers to the client
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported by response writer")
	}

	// Begin streaming events
	for event := range eventsCh {
		// Marshal event to JSON
		resp := &a2a.JSONRPCResponse{
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

// handleTaskResubscription handles the tasks/resubscribe method.
func (s *Server) handleTaskResubscription(ctx context.Context, w http.ResponseWriter, id a2a.ID, params json.RawMessage) (any, error) {
	ctx, span := s.tracer.Start(ctx, "server.handleTaskResubscription")
	defer span.End()

	var req a2a.TaskResubscriptionRequest
	if err := sonic.ConfigFastest.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	// Create a channel for the task events
	ch, err := s.taskManager.OnResubscribeToTask(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to task: %w", err)
	}

	// Flush the headers to the client
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported by response writer")
	}

	events, ok := ch.(<-chan a2a.TaskEvent)
	if !ok {
		return nil, nil
	}

	// Begin streaming events
	for event := range events {
		// Marshal event to JSON
		resp := &a2a.JSONRPCResponse{
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
