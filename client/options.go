// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"net/http"
	"time"
)

// ClientOption configures a client.
type ClientOption func(*clientConfig)

// RequestOption configures a request.
type RequestOption func(*requestConfig)

// clientConfig holds the configuration for the client.
type clientConfig struct {
	httpClient   *http.Client
	interceptors []Interceptor
	userAgent    string
	timeout      time.Duration
	retryPolicy  *RetryPolicy
}

// requestConfig holds the configuration for a request.
type requestConfig struct {
	headers map[string]string
	timeout time.Duration
}

// RetryPolicy defines retry behavior for requests.
type RetryPolicy struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// WithHTTPClient sets the HTTP client to use.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *clientConfig) {
		c.httpClient = client
	}
}

// WithInterceptors adds interceptors to the client.
func WithInterceptors(interceptors ...Interceptor) ClientOption {
	return func(c *clientConfig) {
		c.interceptors = append(c.interceptors, interceptors...)
	}
}

// WithUserAgent sets the user agent for requests.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *clientConfig) {
		c.userAgent = userAgent
	}
}

// WithTimeout sets the default timeout for requests.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientConfig) {
		c.timeout = timeout
	}
}

// WithRetryPolicy sets the retry policy for requests.
func WithRetryPolicy(policy *RetryPolicy) ClientOption {
	return func(c *clientConfig) {
		c.retryPolicy = policy
	}
}

// WithHeader adds a header to the request.
func WithHeader(key, value string) RequestOption {
	return func(r *requestConfig) {
		if r.headers == nil {
			r.headers = make(map[string]string)
		}
		r.headers[key] = value
	}
}

// WithHeaders adds multiple headers to the request.
func WithHeaders(headers map[string]string) RequestOption {
	return func(r *requestConfig) {
		if r.headers == nil {
			r.headers = make(map[string]string)
		}
		for k, v := range headers {
			r.headers[k] = v
		}
	}
}

// WithRequestTimeout sets the timeout for the request.
func WithRequestTimeout(timeout time.Duration) RequestOption {
	return func(r *requestConfig) {
		r.timeout = timeout
	}
}

// DefaultRetryPolicy returns a default retry policy.
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// DefaultClientConfig returns a default client configuration.
func defaultClientConfig() *clientConfig {
	return &clientConfig{
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		interceptors: []Interceptor{},
		userAgent:    "go-a2a/1.0",
		timeout:      30 * time.Second,
		retryPolicy:  DefaultRetryPolicy(),
	}
}

// DefaultRequestConfig returns a default request configuration.
func defaultRequestConfig() *requestConfig {
	return &requestConfig{
		headers: make(map[string]string),
		timeout: 30 * time.Second,
	}
}

// applyClientOptions applies the client options to the configuration.
func applyClientOptions(opts ...ClientOption) *clientConfig {
	config := defaultClientConfig()
	for _, opt := range opts {
		opt(config)
	}
	return config
}

// applyRequestOptions applies the request options to the configuration.
func applyRequestOptions(opts ...RequestOption) *requestConfig {
	config := defaultRequestConfig()
	for _, opt := range opts {
		opt(config)
	}
	return config
}
