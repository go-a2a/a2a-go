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
	"fmt"
	"net/http"
	"time"

	"github.com/go-a2a/a2a-go/transport"
)

// Option configures a Client.
type Option func(*options) error

// options holds all configuration for a Client.
type options struct {
	// Core configuration
	baseURL    string
	httpClient *http.Client
	transport  transport.Transport

	// Connection settings
	timeout           time.Duration
	connectTimeout    time.Duration
	keepAliveInterval time.Duration
	maxIdleConns      int
	maxConnsPerHost   int

	// Retry configuration
	retryConfig *RetryConfig

	// Authentication
	authToken    string
	authProvider AuthProvider

	// Interceptors
	interceptors []transport.Interceptor

	// Callbacks
	onConnectionStateChange ConnectionStateCallback
	onMessage               MessageCallback

	// Feature flags
	enableAutoReconnect bool
	enableCompression   bool
	enableDebugLogging  bool

	// Stream configuration
	streamBufferSize  int
	streamReadTimeout time.Duration
}

// defaultOptions returns default client options.
func defaultOptions() *options {
	return &options{
		timeout:             30 * time.Second,
		connectTimeout:      10 * time.Second,
		keepAliveInterval:   30 * time.Second,
		maxIdleConns:        100,
		maxConnsPerHost:     10,
		retryConfig:         DefaultRetryConfig(),
		enableAutoReconnect: true,
		enableCompression:   true,
		streamBufferSize:    1024,
		streamReadTimeout:   5 * time.Minute,
		httpClient:          http.DefaultClient,
	}
}

// WithBaseURL sets the base URL for the A2A agent.
func WithBaseURL(url string) Option {
	return func(o *options) error {
		if url == "" {
			return &ValidationError{Field: "baseURL", Message: "base URL cannot be empty"}
		}
		o.baseURL = url
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(o *options) error {
		if client == nil {
			return &ValidationError{Field: "httpClient", Message: "HTTP client cannot be nil"}
		}
		o.httpClient = client
		return nil
	}
}

// WithTransport sets a custom transport.
func WithTransport(t transport.Transport) Option {
	return func(o *options) error {
		if t == nil {
			return &ValidationError{Field: "transport", Message: "transport cannot be nil"}
		}
		o.transport = t
		return nil
	}
}

// WithTimeout sets the default timeout for requests.
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) error {
		if timeout <= 0 {
			return &ValidationError{Field: "timeout", Message: "timeout must be positive"}
		}
		o.timeout = timeout
		return nil
	}
}

// WithConnectTimeout sets the connection timeout.
func WithConnectTimeout(timeout time.Duration) Option {
	return func(o *options) error {
		if timeout <= 0 {
			return &ValidationError{Field: "connectTimeout", Message: "connect timeout must be positive"}
		}
		o.connectTimeout = timeout
		return nil
	}
}

// WithRetryConfig sets custom retry configuration.
func WithRetryConfig(config *RetryConfig) Option {
	return func(o *options) error {
		if config == nil {
			return &ValidationError{Field: "retryConfig", Message: "retry config cannot be nil"}
		}
		if config.MaxAttempts < 0 {
			return &ValidationError{Field: "retryConfig.MaxAttempts", Message: "max attempts must be non-negative"}
		}
		o.retryConfig = config
		return nil
	}
}

// WithAuthToken sets the authentication token.
func WithAuthToken(token string) Option {
	return func(o *options) error {
		o.authToken = token
		return nil
	}
}

// WithAuthProvider sets a custom authentication provider.
func WithAuthProvider(provider AuthProvider) Option {
	return func(o *options) error {
		if provider == nil {
			return &ValidationError{Field: "authProvider", Message: "auth provider cannot be nil"}
		}
		o.authProvider = provider
		return nil
	}
}

// WithInterceptor adds an interceptor to the client.
func WithInterceptor(interceptor transport.Interceptor) Option {
	return func(o *options) error {
		if interceptor == nil {
			return &ValidationError{Field: "interceptor", Message: "interceptor cannot be nil"}
		}
		o.interceptors = append(o.interceptors, interceptor)
		return nil
	}
}

// WithInterceptors sets multiple interceptors.
func WithInterceptors(interceptors ...transport.Interceptor) Option {
	return func(o *options) error {
		for i, interceptor := range interceptors {
			if interceptor == nil {
				return &ValidationError{
					Field:   "interceptor",
					Message: fmt.Sprintf("interceptor at index %d cannot be nil", i),
				}
			}
		}
		o.interceptors = append(o.interceptors, interceptors...)
		return nil
	}
}

// WithConnectionStateCallback sets a callback for connection state changes.
func WithConnectionStateCallback(callback ConnectionStateCallback) Option {
	return func(o *options) error {
		o.onConnectionStateChange = callback
		return nil
	}
}

// WithMessageCallback sets a callback for messages.
func WithMessageCallback(callback MessageCallback) Option {
	return func(o *options) error {
		o.onMessage = callback
		return nil
	}
}

// WithAutoReconnect enables or disables automatic reconnection.
func WithAutoReconnect(enable bool) Option {
	return func(o *options) error {
		o.enableAutoReconnect = enable
		return nil
	}
}

// WithCompression enables or disables compression.
func WithCompression(enable bool) Option {
	return func(o *options) error {
		o.enableCompression = enable
		return nil
	}
}

// WithDebugLogging enables or disables debug logging.
func WithDebugLogging(enable bool) Option {
	return func(o *options) error {
		o.enableDebugLogging = enable
		return nil
	}
}

// WithStreamBufferSize sets the buffer size for streaming responses.
func WithStreamBufferSize(size int) Option {
	return func(o *options) error {
		if size <= 0 {
			return &ValidationError{Field: "streamBufferSize", Message: "stream buffer size must be positive"}
		}
		o.streamBufferSize = size
		return nil
	}
}

// WithStreamReadTimeout sets the read timeout for streaming responses.
func WithStreamReadTimeout(timeout time.Duration) Option {
	return func(o *options) error {
		if timeout <= 0 {
			return &ValidationError{Field: "streamReadTimeout", Message: "stream read timeout must be positive"}
		}
		o.streamReadTimeout = timeout
		return nil
	}
}

// WithMaxIdleConns sets the maximum number of idle connections.
func WithMaxIdleConns(n int) Option {
	return func(o *options) error {
		if n < 0 {
			return &ValidationError{Field: "maxIdleConns", Message: "max idle connections must be non-negative"}
		}
		o.maxIdleConns = n
		return nil
	}
}

// WithMaxConnsPerHost sets the maximum number of connections per host.
func WithMaxConnsPerHost(n int) Option {
	return func(o *options) error {
		if n <= 0 {
			return &ValidationError{Field: "maxConnsPerHost", Message: "max connections per host must be positive"}
		}
		o.maxConnsPerHost = n
		return nil
	}
}

// AuthProvider provides authentication credentials.
type AuthProvider interface {
	// GetToken returns the current authentication token.
	GetToken() (string, error)
	// RefreshToken refreshes the authentication token if needed.
	RefreshToken() error
}
