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

const (
	// AgantPath is the path to the agent card.
	AgantPath = "/.well-known/agent.json"
)

// Server represents an A2A server that handles incoming requests and manages tasks.
type Server struct {
	// server is the HTTP server.
	server *http.Server

	// handlers is a list of middleware handlers to apply to the server.
	handlers []func(http.Handler) http.Handler

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
func NewServer(host, port string, agentCard *a2a.AgentCard, taskManager TaskManager, opts ...Option) *Server {
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
	// Handle well-known agent.json
	mux.HandleFunc("GET "+AgantPath, s.agentCardRequestHandler)
	// Handle A2A API requests
	mux.HandleFunc("POST "+s.endpoint, s.requestHandler)

	h := http.Handler(mux)
	if len(s.handlers) > 0 {
		h = s.handlers[len(s.handlers)-1](h)
		for i := len(s.handlers) - 2; i >= 0; i-- {
			h = s.handlers[i](h)
		}
	}

	s.server = &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: h,
	}

	return s
}

// ListenAndServe starts the [Server].
func (s *Server) ListenAndServe(ctx context.Context) error {
	if s.agentCard.Name == "" || s.agentCard.URL == "" || s.agentCard.Version == "" {
		return errors.New("agent card must have name, URL, and version")
	}
	if s.taskManager == nil {
		return errors.New("task manager cannot be nil")
	}

	s.logger.DebugContext(ctx, "starting A2A server", "address", s.server.Addr, "endpoint", s.endpoint)
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

// agentCardRequestHandler handles requests for the agent card.
func (s *Server) agentCardRequestHandler(w http.ResponseWriter, r *http.Request) {
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

	r = r.WithContext(ctx)

	if r.Method != http.MethodPost {
		span.SetAttributes(semconv.RPCJsonrpcErrorCode(a2a.InvalidRequestErrorCode))

		s.writeError(ctx, w, a2a.InvalidRequestErrorCode, "method not allowed")
		return
	}

	var req a2a.JSONRPCRequest
	if err := sonic.ConfigFastest.NewDecoder(r.Body).Decode(&req); err != nil {
		span.SetAttributes(semconv.RPCJsonrpcErrorCode(a2a.JSONParseErrorCode))
		span.SetStatus(codes.Error, err.Error())

		s.writeError(ctx, w, a2a.JSONParseErrorCode, "requestHandler: Invalid JSON payload")
		return
	}

	span.SetAttributes(
		semconv.RPCJsonrpcVersion(req.JSONRPC),
		semconv.RPCJsonrpcRequestID(req.ID.String()),
		attribute.Stringer("a2a.request_id", req.ID),
		attribute.String("a2a.method", req.Method),
	)

	// Handle method
	switch req.Method {
	case a2a.MethodTasksSend:
		s.handleSendTask(w, r, req.Params)
	case a2a.MethodTasksGet:
		s.handleGetTask(w, r, req.Params)
	case a2a.MethodTasksCancel:
		s.handleCancelTask(w, r, req.Params)
	case a2a.MethodTasksPushNotificationSet:
		s.handleSetTaskPushNotification(w, r, req.Params)
	case a2a.MethodTasksPushNotificationGet:
		s.handleGetTaskPushNotification(w, r, req.Params)
	case a2a.MethodTasksSendSubscribe:
		s.handleSendTaskStreaming(w, r, req.ID, req.Params)
	case a2a.MethodTasksResubscribe:
		s.handleTaskResubscription(w, r, req.ID, req.Params)
	default:
		s.writeError(ctx, w, a2a.MethodNotFoundErrorCode, "Method not found")
	}
}

func (s *Server) writeResponse(ctx context.Context, w http.ResponseWriter, id a2a.ID, result any) {
	// Write success response
	resp := &a2a.JSONRPCResponse{
		JSONRPCMessage: a2a.NewJSONRPCMessage(id),
		Result:         result,
	}

	data, err := sonic.ConfigFastest.Marshal(resp)
	if err != nil {
		s.logger.ErrorContext(ctx, "marshal response", slog.Any("error", err))
		s.writeError(ctx, w, a2a.InternalErrorCode, "marshal response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		s.logger.ErrorContext(ctx, "write response", slog.Any("error", err))
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("write response: %w", err).Error())
	}
}

// writeError writes an error response in JSON-RPC format.
func (s *Server) writeError(ctx context.Context, w http.ResponseWriter, code int, message string) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(semconv.RPCJsonrpcErrorCode(a2a.InternalErrorCode))
	span.SetStatus(codes.Error, message)

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
func (s *Server) handleSendTask(w http.ResponseWriter, r *http.Request, params json.RawMessage) {
	ctx, span := s.tracer.Start(r.Context(), "server.handleSendTask")
	defer span.End()

	var req a2a.SendTaskRequest
	if err := sonic.ConfigFastest.Unmarshal(params, &req); err != nil {
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("invalid params: %w", err).Error())
		return
	}

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	resp, err := s.taskManager.OnSendTask(ctx, &req)
	if err != nil {
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("process task: %w", err).Error())
		return
	}

	s.writeResponse(ctx, w, req.ID, resp.Result)
}

// handleGetTask handles the tasks/get method.
func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request, params json.RawMessage) {
	ctx, span := s.tracer.Start(r.Context(), "server.handleGetTask")
	defer span.End()

	var req a2a.GetTaskRequest
	if err := sonic.ConfigFastest.Unmarshal(params, &req); err != nil {
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("invalid params: %w", err).Error())
		return
	}

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	resp, err := s.taskManager.OnGetTask(ctx, &req)
	if err != nil {
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("get task: %w", err).Error())
		return
	}

	s.writeResponse(ctx, w, req.ID, resp.Result)
}

// handleCancelTask handles the tasks/cancel method.
func (s *Server) handleCancelTask(w http.ResponseWriter, r *http.Request, params json.RawMessage) {
	ctx, span := s.tracer.Start(r.Context(), "server.handleCancelTask")
	defer span.End()

	var req a2a.CancelTaskRequest
	if err := sonic.ConfigFastest.Unmarshal(params, &req); err != nil {
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("invalid params: %w", err).Error())
		return
	}

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	resp, err := s.taskManager.OnCancelTask(ctx, &req)
	if err != nil {
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("cancel task: %w", err).Error())
		return
	}

	s.writeResponse(ctx, w, req.ID, resp.Result)
}

// handleSetTaskPushNotification handles the tasks/pushNotification/set method.
func (s *Server) handleSetTaskPushNotification(w http.ResponseWriter, r *http.Request, params json.RawMessage) {
	ctx, span := s.tracer.Start(r.Context(), "server.handleSetTaskPushNotification")
	defer span.End()

	var req a2a.SetTaskPushNotificationRequest
	if err := sonic.ConfigFastest.Unmarshal(params, &req); err != nil {
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("invalid params: %w", err).Error())
		return
	}

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	resp, err := s.taskManager.OnSetTaskPushNotification(ctx, &req)
	if err != nil {
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("set push notification: %w", err).Error())
		return
	}

	s.writeResponse(ctx, w, req.ID, resp.Result)
}

// handleGetTaskPushNotification handles the tasks/pushNotification/get method.
func (s *Server) handleGetTaskPushNotification(w http.ResponseWriter, r *http.Request, params json.RawMessage) {
	ctx, span := s.tracer.Start(r.Context(), "server.handleGetTaskPushNotification")
	defer span.End()

	var req a2a.GetTaskPushNotificationRequest
	if err := sonic.ConfigFastest.Unmarshal(params, &req); err != nil {
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("invalid params: %w", err).Error())
		return
	}

	span.SetAttributes(attribute.String("a2a.task_id", req.Params.ID))

	resp, err := s.taskManager.OnGetTaskPushNotification(ctx, &req)
	if err != nil {
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("get push notification: %w", err).Error())
		return
	}

	s.writeResponse(ctx, w, req.ID, resp.Result)
}

// handleSendTaskStreaming handles the tasks/sendSubscribe method.
func (s *Server) handleSendTaskStreaming(w http.ResponseWriter, r *http.Request, id a2a.ID, params json.RawMessage) {
	ctx, span := s.tracer.Start(r.Context(), "server.handleSendTaskStreaming")
	defer span.End()

	var req a2a.SendTaskStreamingRequest
	if err := sonic.ConfigFastest.Unmarshal(params, &req); err != nil {
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("invalid params: %w", err).Error())
		return
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
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("subscribe to task: %w", err).Error())
		return
	}

	// Flush the headers to the client
	flusher, ok := w.(http.Flusher)
	if !ok {
		s.writeError(ctx, w, a2a.InternalErrorCode, "streaming not supported by response writer")
		return
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
}

// handleTaskResubscription handles the tasks/resubscribe method.
func (s *Server) handleTaskResubscription(w http.ResponseWriter, r *http.Request, id a2a.ID, params json.RawMessage) {
	ctx, span := s.tracer.Start(r.Context(), "server.handleTaskResubscription")
	defer span.End()

	var req a2a.TaskResubscriptionRequest
	if err := sonic.ConfigFastest.Unmarshal(params, &req); err != nil {
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("invalid params: %w", err).Error())
		return
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
		s.writeError(ctx, w, a2a.InternalErrorCode, fmt.Errorf("subscribe to task: %w", err).Error())
		return
	}

	// Flush the headers to the client
	flusher, ok := w.(http.Flusher)
	if !ok {
		s.writeError(ctx, w, a2a.InternalErrorCode, "streaming not supported by response writer")
		return
	}

	events, ok := ch.(<-chan a2a.TaskEvent)
	if !ok {
		return
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
}
