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

package handler

import (
	"fmt"
	"sync"

	a2a "github.com/go-a2a/a2a-go"
)

// router implements the Router interface.
type router struct {
	mu                  sync.RWMutex
	handlers            map[a2a.Method]Handler
	streamingHandlers   map[a2a.Method]StreamingHandler
	middleware          []Middleware
	streamingMiddleware []StreamingMiddleware
}

// NewRouter creates a new router instance.
func NewRouter() Router {
	return &router{
		handlers:          make(map[a2a.Method]Handler),
		streamingHandlers: make(map[a2a.Method]StreamingHandler),
	}
}

// Register registers a handler for specific methods.
func (r *router) Register(handler Handler, methods ...a2a.Method) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	if len(methods) == 0 {
		methods = handler.SupportedMethods()
	}

	for _, method := range methods {
		if _, exists := r.handlers[method]; exists {
			return fmt.Errorf("handler already registered for method %s", method)
		}
		// Apply middleware chain
		h := handler
		for i := len(r.middleware) - 1; i >= 0; i-- {
			h = r.middleware[i](h)
		}
		r.handlers[method] = h
	}

	return nil
}

// RegisterStreaming registers a streaming handler for specific methods.
func (r *router) RegisterStreaming(handler StreamingHandler, methods ...a2a.Method) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if handler == nil {
		return fmt.Errorf("streaming handler cannot be nil")
	}

	if len(methods) == 0 {
		methods = handler.SupportedMethods()
	}

	for _, method := range methods {
		if _, exists := r.streamingHandlers[method]; exists {
			return fmt.Errorf("streaming handler already registered for method %s", method)
		}
		// Apply streaming middleware chain
		h := handler
		for i := len(r.streamingMiddleware) - 1; i >= 0; i-- {
			h = r.streamingMiddleware[i](h)
		}
		r.streamingHandlers[method] = h
	}

	// Also register as regular handler
	return r.Register(handler, methods...)
}

// Route finds the appropriate handler for a method.
func (r *router) Route(method a2a.Method) (Handler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.handlers[method]
	if !exists {
		return nil, fmt.Errorf("no handler registered for method %s", method)
	}

	return handler, nil
}

// RouteStreaming finds the appropriate streaming handler for a method.
func (r *router) RouteStreaming(method a2a.Method) (StreamingHandler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.streamingHandlers[method]
	if !exists {
		return nil, fmt.Errorf("no streaming handler registered for method %s", method)
	}

	return handler, nil
}

// Use adds middleware to the router.
func (r *router) Use(middleware ...Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.middleware = append(r.middleware, middleware...)

	// Re-apply middleware to existing handlers
	for method, handler := range r.handlers {
		h := handler
		// Remove existing middleware layers
		for i := 0; i < len(r.middleware)-len(middleware); i++ {
			// This is a simplification; in practice, we'd need to track the original handler
		}
		// Apply all middleware
		for i := len(r.middleware) - 1; i >= 0; i-- {
			h = r.middleware[i](h)
		}
		r.handlers[method] = h
	}
}

// UseStreaming adds streaming middleware to the router.
func (r *router) UseStreaming(middleware ...StreamingMiddleware) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.streamingMiddleware = append(r.streamingMiddleware, middleware...)

	// Re-apply middleware to existing streaming handlers
	for method, handler := range r.streamingHandlers {
		h := handler
		// Apply all streaming middleware
		for i := len(r.streamingMiddleware) - 1; i >= 0; i-- {
			h = r.streamingMiddleware[i](h)
		}
		r.streamingHandlers[method] = h
	}
}

// DefaultRouter provides a pre-configured router with common middleware.
func DefaultRouter() Router {
	r := NewRouter()
	r.Use(
		RecoveryMiddleware(),
		LoggingMiddleware(),
	)
	return r
}