// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package server provides server-side abstractions for the A2A protocol.
// This package implements context management for server calls, including
// user authentication and request state management.
package server

import (
	"context"
	"fmt"
	"maps"
	"sync"

	"github.com/go-a2a/a2a/auth"
)

// ServerCallContext represents the context for a server call in the A2A protocol.
// It contains information about the incoming call, including the authenticated
// user and any additional state information. This type is thread-safe and can
// be safely used across multiple goroutines.
type ServerCallContext struct {
	user  auth.User
	state map[string]any
	mu    sync.RWMutex
}

// NewServerCallContext creates a new ServerCallContext with the provided user.
// The state map is initialized as empty and can be populated using SetState.
// If user is nil, an UnauthenticatedUser will be used as a safe default.
func NewServerCallContext(user auth.User) *ServerCallContext {
	if user == nil {
		user = auth.UnauthenticatedUser{}
	}
	return &ServerCallContext{
		user:  user,
		state: make(map[string]any),
	}
}

// NewServerCallContextWithState creates a new ServerCallContext with the provided
// user and initial state. The state map is copied to ensure thread safety.
// If user is nil, an UnauthenticatedUser will be used as a safe default.
func NewServerCallContextWithState(user auth.User, state map[string]any) *ServerCallContext {
	if user == nil {
		user = auth.UnauthenticatedUser{}
	}

	scc := &ServerCallContext{
		user:  user,
		state: make(map[string]any),
	}

	// Copy the provided state to ensure thread safety
	maps.Copy(scc.state, state)

	return scc
}

// User returns the authenticated user associated with this call context.
// This method is thread-safe and can be called from multiple goroutines.
func (scc *ServerCallContext) User() auth.User {
	scc.mu.RLock()
	defer scc.mu.RUnlock()
	return scc.user
}

// State returns a copy of the current state map. This method is thread-safe
// and the returned map can be modified without affecting the internal state.
func (scc *ServerCallContext) State() map[string]any {
	scc.mu.RLock()
	defer scc.mu.RUnlock()

	// Return a copy to prevent external modification
	stateCopy := make(map[string]any, len(scc.state))
	maps.Copy(stateCopy, scc.state)
	return stateCopy
}

// SetState sets a value in the context state. This method is thread-safe
// and can be called from multiple goroutines.
func (scc *ServerCallContext) SetState(key string, value any) {
	scc.mu.Lock()
	defer scc.mu.Unlock()
	scc.state[key] = value
}

// GetState retrieves a value from the context state. It returns the value
// and a boolean indicating whether the key was found. This method is thread-safe.
func (scc *ServerCallContext) GetState(key string) (any, bool) {
	scc.mu.RLock()
	defer scc.mu.RUnlock()
	value, ok := scc.state[key]
	return value, ok
}

// DeleteState removes a key from the context state. This method is thread-safe.
func (scc *ServerCallContext) DeleteState(key string) {
	scc.mu.Lock()
	defer scc.mu.Unlock()
	delete(scc.state, key)
}

// CallContextBuilder defines the interface for building ServerCallContext instances
// from different types of requests. Implementations should handle the specifics
// of their transport protocol (HTTP, gRPC, etc.) and extract relevant information
// to populate the context.
type CallContextBuilder interface {
	// Build creates a ServerCallContext from the provided request context and
	// request object. The request parameter type depends on the specific
	// implementation (e.g., *http.Request for HTTP, grpc.ServerStream for gRPC).
	Build(ctx context.Context, request any) (*ServerCallContext, error)
}

// DefaultCallContextBuilder provides a basic implementation of CallContextBuilder
// that creates contexts with unauthenticated users. This can be used as a base
// for more sophisticated builders or as a fallback when no authentication is needed.
type DefaultCallContextBuilder struct{}

var _ CallContextBuilder = (*DefaultCallContextBuilder)(nil)

// NewDefaultCallContextBuilder creates a new DefaultCallContextBuilder instance.
func NewDefaultCallContextBuilder() *DefaultCallContextBuilder {
	return &DefaultCallContextBuilder{}
}

// Build creates a ServerCallContext with an unauthenticated user.
// The request parameter is ignored in this basic implementation.
// More sophisticated implementations might extract user information,
// headers, or other metadata from the request.
func (dcb *DefaultCallContextBuilder) Build(ctx context.Context, request any) (*ServerCallContext, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	user := auth.UnauthenticatedUser{}
	state := make(map[string]any)

	// Store the Go context for potential future use
	state["go_context"] = ctx

	return NewServerCallContextWithState(user, state), nil
}

// Validate ensures the ServerCallContext is in a valid state.
// This method checks that the user is not nil and the state map is initialized.
func (scc *ServerCallContext) Validate() error {
	scc.mu.RLock()
	defer scc.mu.RUnlock()

	if scc.user == nil {
		return fmt.Errorf("server call context user cannot be nil")
	}

	if scc.state == nil {
		return fmt.Errorf("server call context state cannot be nil")
	}

	return nil
}

// String returns a string representation of the ServerCallContext for debugging.
func (scc *ServerCallContext) String() string {
	scc.mu.RLock()
	defer scc.mu.RUnlock()

	return fmt.Sprintf("ServerCallContext{user: %s, authenticated: %t, state_keys: %d}",
		scc.user.UserName(),
		scc.user.IsAuthenticated(),
		len(scc.state))
}
