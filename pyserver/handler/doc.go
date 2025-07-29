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

// Package handler provides request handling infrastructure for A2A servers.
//
// This package implements the core request handling abstractions that bridge
// between transport protocols (HTTP, WebSocket, SSE) and agent executors.
// It provides:
//
//   - Handler interfaces for processing A2A protocol methods
//   - Request routing and method dispatch
//   - Session and context management
//   - Middleware support for cross-cutting concerns
//   - Streaming support for long-running operations
//   - Error handling and response formatting
//
// # Architecture
//
// The handler package follows a layered architecture:
//
//   - Transport Layer: HTTP, WebSocket, SSE handlers
//   - Routing Layer: Method routing and middleware chains
//   - Execution Layer: Agent executors that process requests
//   - Response Layer: Response formatting and streaming
//
// # Usage
//
// Basic handler implementation:
//
//	type MyHandler struct{}
//
//	func (h *MyHandler) Handle(ctx context.Context, method a2a.Method, params a2a.Params) (a2a.Result, error) {
//	    switch method {
//	    case a2a.MethodMessageSend:
//	        // Process message send
//	        return processMessage(ctx, params)
//	    default:
//	        return nil, errors.New("unsupported method")
//	    }
//	}
//
//	func (h *MyHandler) SupportedMethods() []a2a.Method {
//	    return []a2a.Method{a2a.MethodMessageSend}
//	}
//
// Setting up a router with middleware:
//
//	router := handler.NewRouter()
//	router.Use(
//	    handler.LoggingMiddleware(),
//	    handler.AuthenticationMiddleware(),
//	    handler.ValidationMiddleware(),
//	)
//	router.Register(myHandler, a2a.MethodMessageSend)
//
// # Streaming
//
// For long-running operations, handlers can implement StreamingHandler:
//
//	func (h *MyStreamingHandler) HandleStream(ctx context.Context, method a2a.Method, params a2a.Params) (<-chan a2a.SendStreamingMessageResponse, error) {
//	    ch := make(chan a2a.SendStreamingMessageResponse)
//	    go func() {
//	        defer close(ch)
//	        // Send streaming updates
//	        ch <- &a2a.TaskStatusUpdateEvent{...}
//	    }()
//	    return ch, nil
//	}
package handler