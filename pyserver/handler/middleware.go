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
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	a2a "github.com/go-a2a/a2a-go"
)

// LoggingMiddleware logs request and response information.
func LoggingMiddleware() Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, method a2a.Method, params a2a.Params) (a2a.Result, error) {
			start := time.Now()
			
			// Log request
			log.Printf("[REQUEST] Method: %s, Params: %T", method, params)
			
			// Call next handler
			result, err := next.Handle(ctx, method, params)
			
			// Log response
			duration := time.Since(start)
			if err != nil {
				log.Printf("[RESPONSE] Method: %s, Duration: %v, Error: %v", method, duration, err)
			} else {
				log.Printf("[RESPONSE] Method: %s, Duration: %v, Result: %T", method, duration, result)
			}
			
			return result, err
		})
	}
}

// RecoveryMiddleware recovers from panics and returns an error.
func RecoveryMiddleware() Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, method a2a.Method, params a2a.Params) (result a2a.Result, err error) {
			defer func() {
				if r := recover(); r != nil {
					// Log the panic and stack trace
					log.Printf("[PANIC] Method: %s, Error: %v\nStack: %s", method, r, debug.Stack())
					err = fmt.Errorf("internal server error: %v", r)
				}
			}()
			
			return next.Handle(ctx, method, params)
		})
	}
}

// TimeoutMiddleware adds a timeout to request processing.
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, method a2a.Method, params a2a.Params) (a2a.Result, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			
			resultCh := make(chan a2a.Result, 1)
			errCh := make(chan error, 1)
			
			go func() {
				result, err := next.Handle(ctx, method, params)
				if err != nil {
					errCh <- err
				} else {
					resultCh <- result
				}
			}()
			
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("request timeout after %v", timeout)
			case err := <-errCh:
				return nil, err
			case result := <-resultCh:
				return result, nil
			}
		})
	}
}

// ValidationMiddleware validates request parameters.
func ValidationMiddleware() Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, method a2a.Method, params a2a.Params) (a2a.Result, error) {
			// Validate based on method
			switch method {
			case a2a.MethodMessageSend, a2a.MethodMessageStream:
				if params == nil {
					return nil, fmt.Errorf("missing required parameters")
				}
				msgParams, ok := params.(*a2a.MessageSendParams)
				if !ok {
					return nil, fmt.Errorf("invalid parameter type for %s", method)
				}
				if msgParams.Message == nil {
					return nil, fmt.Errorf("message is required")
				}
			case a2a.MethodTasksGet, a2a.MethodTasksCancel:
				if params == nil {
					return nil, fmt.Errorf("missing required parameters")
				}
				taskParams, ok := params.(*a2a.TaskQueryParams)
				if !ok {
					idParams, ok := params.(*a2a.TaskIDParams)
					if !ok {
						return nil, fmt.Errorf("invalid parameter type for %s", method)
					}
					if idParams.ID == "" {
						return nil, fmt.Errorf("task ID is required")
					}
				} else if taskParams.ID == "" {
					return nil, fmt.Errorf("task ID is required")
				}
			}
			
			return next.Handle(ctx, method, params)
		})
	}
}

// AuthenticationMiddleware checks authentication status.
func AuthenticationMiddleware(optional bool) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, method a2a.Method, params a2a.Params) (a2a.Result, error) {
			// Extract session from context
			session, ok := ctx.Value(sessionContextKey).(*Session)
			if !ok || session == nil {
				if !optional {
					return nil, fmt.Errorf("authentication required")
				}
				// Create anonymous session
				session = &Session{
					ID:              "anonymous",
					IsAuthenticated: false,
					Metadata:        make(map[string]any),
				}
			}
			
			// Check if authenticated for protected methods
			if !session.IsAuthenticated && requiresAuth(method) {
				return nil, fmt.Errorf("authentication required for method %s", method)
			}
			
			// Add session to context for downstream handlers
			ctx = context.WithValue(ctx, sessionContextKey, session)
			
			return next.Handle(ctx, method, params)
		})
	}
}

// MetricsMiddleware collects metrics about request processing.
func MetricsMiddleware(collector MetricsCollector) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, method a2a.Method, params a2a.Params) (a2a.Result, error) {
			start := time.Now()
			
			// Call next handler
			result, err := next.Handle(ctx, method, params)
			
			// Record metrics
			duration := time.Since(start)
			status := "success"
			if err != nil {
				status = "error"
			}
			
			collector.RecordRequest(string(method), status, duration)
			
			return result, err
		})
	}
}

// HandlerFunc is an adapter to allow the use of ordinary functions as Handlers.
type HandlerFunc func(context.Context, a2a.Method, a2a.Params) (a2a.Result, error)

// Handle calls f(ctx, method, params).
func (f HandlerFunc) Handle(ctx context.Context, method a2a.Method, params a2a.Params) (a2a.Result, error) {
	return f(ctx, method, params)
}

// SupportedMethods returns nil as HandlerFunc doesn't have predefined methods.
func (f HandlerFunc) SupportedMethods() []a2a.Method {
	return nil
}

// MetricsCollector defines the interface for collecting metrics.
type MetricsCollector interface {
	RecordRequest(method, status string, duration time.Duration)
}

// sessionContextKey is the context key for session data.
type contextKey string

const sessionContextKey contextKey = "session"

// requiresAuth checks if a method requires authentication.
func requiresAuth(method a2a.Method) bool {
	switch method {
	case a2a.MethodAgentAuthenticatedExtendedCard:
		return true
	default:
		return false
	}
}