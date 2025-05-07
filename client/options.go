// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"net/http"
	"time"
)

// Options represents the configuration options for the A2A client.
type Options struct {
	// HTTPClient is the HTTP client to use for requests.
	// If nil, http.DefaultClient will be used.
	HTTPClient *http.Client

	// Headers are additional HTTP headers to include in requests.
	Headers http.Header

	// Timeout is the default timeout for requests.
	// If zero, no timeout is set.
	Timeout time.Duration

	// RetryConfig configures the retry behavior for failed requests.
	RetryConfig RetryConfig
}

// RetryConfig configures retry behavior for failed requests.
type RetryConfig struct {
	// MaxRetries is the maximum number of retries for a request.
	// If zero, no retries will be attempted.
	MaxRetries int

	// RetryDelay is the base delay between retries.
	// The actual delay will be calculated using exponential backoff.
	RetryDelay time.Duration

	// MaxRetryDelay is the maximum delay between retries.
	MaxRetryDelay time.Duration

	// RetryableStatusCodes is a list of HTTP status codes that should trigger a retry.
	// Common retryable status codes include 429 (Too Many Requests) and 5xx codes.
	RetryableStatusCodes []int
}

// DefaultOptions returns the default client options.
func DefaultOptions() Options {
	return Options{
		Timeout: 30 * time.Second,
		RetryConfig: RetryConfig{
			MaxRetries:           3,
			RetryDelay:           100 * time.Millisecond,
			MaxRetryDelay:        10 * time.Second,
			RetryableStatusCodes: []int{408, 429, 500, 502, 503, 504},
		},
	}
}

// Option is a function that configures a Client.
type Option func(*Options)

// WithHTTPClient sets the HTTP client to use for requests.
func WithHTTPClient(client *http.Client) Option {
	return func(o *Options) {
		o.HTTPClient = client
	}
}

// WithTimeout sets the default timeout for requests.
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithHeaders sets additional HTTP headers to include in requests.
func WithHeaders(headers http.Header) Option {
	return func(o *Options) {
		o.Headers = headers
	}
}

// WithRetryConfig sets the retry configuration.
func WithRetryConfig(config RetryConfig) Option {
	return func(o *Options) {
		o.RetryConfig = config
	}
}

// WithMaxRetries sets the maximum number of retries for a request.
func WithMaxRetries(maxRetries int) Option {
	return func(o *Options) {
		o.RetryConfig.MaxRetries = maxRetries
	}
}

// WithRetryDelay sets the base delay between retries.
func WithRetryDelay(delay time.Duration) Option {
	return func(o *Options) {
		o.RetryConfig.RetryDelay = delay
	}
}

// WithBearerToken sets the Authorization header with a bearer token.
func WithBearerToken(token string) Option {
	return func(o *Options) {
		if o.Headers == nil {
			o.Headers = make(http.Header)
		}
		o.Headers.Set("Authorization", "Bearer "+token)
	}
}
