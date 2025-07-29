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

package pyclient

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-a2a/a2a-go/transport"
)

// retryableFunc is a function that can be retried.
type retryableFunc func(context.Context) error

// withRetry executes a function with retry logic.
func withRetry(ctx context.Context, config *RetryConfig, operation string, fn retryableFunc) error {
	if config == nil || config.MaxAttempts <= 0 {
		// No retry configured
		return fn(ctx)
	}

	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check context before attempt
		if err := ctx.Err(); err != nil {
			return err
		}

		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if config.RetryableErrors != nil && !config.RetryableErrors(err) {
			return err // Not retryable
		}

		// Check if this was the last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Add jitter to delay (10% variance)
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
		actualDelay := delay + jitter

		// Wait before next attempt
		select {
		case <-time.After(actualDelay):
			// Continue to next attempt
		case <-ctx.Done():
			return ctx.Err()
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	// All attempts failed
	return fmt.Errorf("operation %s failed after %d attempts: %w", operation, config.MaxAttempts, lastErr)
}

// retryInterceptor creates an HTTP interceptor that adds retry logic.
func retryInterceptor(config *RetryConfig) transport.Interceptor {
	return func(ctx context.Context, req *http.Request, invoker transport.Invoker) (*http.Response, error) {
		var resp *http.Response

		err := withRetry(ctx, config, "HTTP request", func(ctx context.Context) error {
			var err error
			resp, err = invoker(ctx, req)
			if err != nil {
				return err
			}

			// Check HTTP status for retryable errors
			if resp.StatusCode >= 500 || resp.StatusCode == 429 {
				// Server error or rate limit - retryable
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			}

			return nil
		})

		return resp, err
	}
}
