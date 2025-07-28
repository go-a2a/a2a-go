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

package client

import (
	"context"
	"log/slog"
	"math"
	"net/http"
	"time"
)

// Interceptor defines a middleware function that can intercept and modify requests/responses.
type Interceptor func(ctx context.Context, req *http.Request, invoker Invoker) (*http.Response, error)

// Invoker represents the next handler in the interceptor chain.
type Invoker func(ctx context.Context, req *http.Request) (*http.Response, error)

// ClientCallContext provides context information for client calls.
type ClientCallContext struct {
	Method    string
	URL       string
	Headers   map[string]string
	Metadata  map[string]any
	AgentCard any // Will be typed as *AgentCard when imported
}

// contextKey is used for context values.
type contextKey string

const (
	clientCallContextKey contextKey = "client_call_context"
)

// WithClientCallContext adds a ClientCallContext to the context.
func WithClientCallContext(ctx context.Context, callCtx *ClientCallContext) context.Context {
	return context.WithValue(ctx, clientCallContextKey, callCtx)
}

// GetClientCallContext retrieves the ClientCallContext from the context.
func GetClientCallContext(ctx context.Context) (*ClientCallContext, bool) {
	callCtx, ok := ctx.Value(clientCallContextKey).(*ClientCallContext)
	return callCtx, ok
}

// chainInterceptors chains multiple interceptors together.
func chainInterceptors(interceptors []Interceptor, invoker Invoker) Invoker {
	if len(interceptors) == 0 {
		return invoker
	}

	// Build the chain from right to left
	for i := len(interceptors) - 1; i >= 0; i-- {
		interceptor := interceptors[i]
		next := invoker
		invoker = func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return interceptor(ctx, req, next)
		}
	}

	return invoker
}

// LoggingInterceptor logs requests and responses.
func LoggingInterceptor(logger *slog.Logger) Interceptor {
	return func(ctx context.Context, req *http.Request, invoker Invoker) (*http.Response, error) {
		logger.InfoContext(ctx, "Request", slog.String("method", req.Method), slog.String("req.URL", req.URL.String()))

		resp, err := invoker(ctx, req)
		if err != nil {
			logger.ErrorContext(ctx, "Request failed", slog.Any("error", err))
		} else {
			logger.InfoContext(ctx, "Response", slog.Int("status_code", resp.StatusCode))
		}

		return resp, err
	}
}

// RetryInterceptor retries requests based on the retry policy.
func RetryInterceptor(policy *RetryPolicy) Interceptor {
	return func(ctx context.Context, req *http.Request, invoker Invoker) (*http.Response, error) {
		var lastErr error
		var resp *http.Response

		for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
			resp, lastErr = invoker(ctx, req)

			if lastErr == nil {
				if !shouldRetry(resp.StatusCode) {
					return resp, nil
				}
				// Close the response body if we're retrying
				if resp.Body != nil {
					resp.Body.Close()
				}
			}

			// Don't wait after the last attempt
			if attempt < policy.MaxAttempts-1 {
				delay := calculateDelay(policy, attempt)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(delay):
					// Continue to next attempt
				}
			}
		}

		return resp, lastErr
	}
}

// UserAgentInterceptor adds a user agent header to requests.
func UserAgentInterceptor(userAgent string) Interceptor {
	return func(ctx context.Context, req *http.Request, invoker Invoker) (*http.Response, error) {
		req.Header.Set("User-Agent", userAgent)
		return invoker(ctx, req)
	}
}

// HeaderInterceptor adds custom headers to requests.
func HeaderInterceptor(headers map[string]string) Interceptor {
	return func(ctx context.Context, req *http.Request, invoker Invoker) (*http.Response, error) {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
		return invoker(ctx, req)
	}
}

// shouldRetry determines if a response should be retried based on status code.
func shouldRetry(statusCode int) bool {
	return statusCode >= 500 || statusCode == 408 || statusCode == 429
}

// calculateDelay calculates the delay for the next retry attempt.
func calculateDelay(policy *RetryPolicy, attempt int) time.Duration {
	return min(time.Duration(float64(policy.InitialDelay)*math.Pow(policy.Multiplier, float64(attempt))), policy.MaxDelay)
}

// NoopLogger is a no-op logger implementation.
type NoopLogger struct{}

func (NoopLogger) Infof(format string, args ...any)  {}
func (NoopLogger) Errorf(format string, args ...any) {}
