// Copyright 2025 The Go A2A Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Package handler provides request handling abstractions for A2A server implementations.
package handler

import (
	"context"
	"net/http"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/pyserver/agent_execution"
)

// Handler defines the core interface for all A2A request handlers.
// Handlers process incoming requests and generate appropriate responses.
type Handler interface {
	// Handle processes an A2A request and returns a response.
	Handle(ctx context.Context, method a2a.Method, params a2a.Params) (a2a.Result, error)

	// SupportedMethods returns the A2A methods this handler supports.
	SupportedMethods() []a2a.Method
}

// StreamingHandler extends Handler with streaming capabilities for SSE and WebSocket.
type StreamingHandler interface {
	Handler

	// HandleStream processes a streaming request and returns a stream channel.
	HandleStream(ctx context.Context, method a2a.Method, params a2a.Params) (<-chan a2a.SendStreamingMessageResponse, error)
}

// RequestContext encapsulates all context needed for processing a request.
type RequestContext struct {
	// Method is the A2A method being invoked.
	Method a2a.Method

	// Params contains the request parameters.
	Params a2a.Params

	// Session provides access to session-specific data.
	Session *Session

	// Headers contains the HTTP headers from the request.
	Headers http.Header

	// RequestID is a unique identifier for this request.
	RequestID string
}

// Session represents a client session with state management.
type Session struct {
	// ID is the unique session identifier.
	ID string

	// UserID is the authenticated user identifier, if any.
	UserID string

	// Metadata contains arbitrary session metadata.
	Metadata map[string]any

	// IsAuthenticated indicates if the session is authenticated.
	IsAuthenticated bool
}

// ResponseWriter defines the interface for writing responses.
type ResponseWriter interface {
	// WriteResult writes a successful result.
	WriteResult(result a2a.Result) error

	// WriteError writes an error response.
	WriteError(err error) error

	// WriteStreamingResult writes a streaming result.
	WriteStreamingResult(result a2a.SendStreamingMessageResponse) error

	// Flush flushes any buffered data.
	Flush() error
}

// Middleware defines a function that wraps a Handler.
type Middleware func(Handler) Handler

// StreamingMiddleware defines a function that wraps a StreamingHandler.
type StreamingMiddleware func(StreamingHandler) StreamingHandler

// Router manages routing of A2A methods to appropriate handlers.
type Router interface {
	// Register registers a handler for specific methods.
	Register(handler Handler, methods ...a2a.Method) error

	// RegisterStreaming registers a streaming handler for specific methods.
	RegisterStreaming(handler StreamingHandler, methods ...a2a.Method) error

	// Route finds the appropriate handler for a method.
	Route(method a2a.Method) (Handler, error)

	// RouteStreaming finds the appropriate streaming handler for a method.
	RouteStreaming(method a2a.Method) (StreamingHandler, error)

	// Use adds middleware to the router.
	Use(middleware ...Middleware)

	// UseStreaming adds streaming middleware to the router.
	UseStreaming(middleware ...StreamingMiddleware)
}

// AgentHandler wraps an AgentExecutor to implement the Handler interface.
type AgentHandler struct {
	executor agent_execution.AgentExecutor
}

// NewAgentHandler creates a new handler that delegates to an AgentExecutor.
func NewAgentHandler(executor agent_execution.AgentExecutor) *AgentHandler {
	return &AgentHandler{
		executor: executor,
	}
}

// Handle implements the Handler interface.
func (h *AgentHandler) Handle(ctx context.Context, method a2a.Method, params a2a.Params) (a2a.Result, error) {
	// This will be implemented to bridge between Handler and AgentExecutor
	// For now, return a placeholder
	return nil, nil
}

// SupportedMethods returns the methods this handler supports.
func (h *AgentHandler) SupportedMethods() []a2a.Method {
	return []a2a.Method{
		a2a.MethodMessageSend,
		a2a.MethodTasksGet,
		a2a.MethodTasksCancel,
	}
}

// Chain applies a list of middleware to a handler.
func Chain(h Handler, middleware ...Middleware) Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

// ChainStreaming applies a list of streaming middleware to a streaming handler.
func ChainStreaming(h StreamingHandler, middleware ...StreamingMiddleware) StreamingHandler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}