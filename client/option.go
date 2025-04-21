// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/trace"

	"github.com/go-a2a/a2a"
)

// Option represents an option for configuring the [Client].
type Option func(*Client)

// WithHTTPClient sets the [*http.Client] for the [Client].
func (c *Client) WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithAgentCard sets the agent card for the [Client].
func WithAgentCard(agentCard *a2a.AgentCard) Option {
	return func(c *Client) {
		c.agentCard = agentCard
	}
}

// WithLogger sets the [*slog.Logger] for the [Client].
func WithLogger(logger *slog.Logger) Option {
	return func(s *Client) {
		s.logger = logger
	}
}

// WithTracer sets the [trace.Tracer] for the [Client].
func WithTracer(tracer trace.Tracer) Option {
	return func(s *Client) {
		s.tracer = tracer
	}
}
