// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// ServerOption represents an option for configuring the [Server].
type ServerOption func(*Server)

// WithEndpoint sets the custom endpoint for the [Server].
func (s *Server) WithEndpoint(endpoint string) ServerOption {
	return func(s *Server) {
		s.endpoint = endpoint
	}
}

// WithLogger sets the [*slog.Logger] for the [Server].
func (s *Server) WithLogger(logger *slog.Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithTracer sets the [trace.Tracer] for the [Server].
func (s *Server) WithTracer(tracer trace.Tracer) ServerOption {
	return func(s *Server) {
		s.tracer = tracer
	}
}
