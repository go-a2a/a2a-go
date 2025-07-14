// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package auth provides authentication abstractions for the A2A protocol.
// This package implements user authentication interfaces and types that can be
// used to represent authenticated and unauthenticated users in A2A contexts.
package auth

// User represents an authenticated or unauthenticated user in the A2A system.
// This interface provides the minimal contract for user authentication status
// and identity information.
type User interface {
	// IsAuthenticated returns true if the user is authenticated, false otherwise.
	IsAuthenticated() bool

	// UserName returns the username of the user. For unauthenticated users,
	// this returns an empty string.
	UserName() string
}

// UnauthenticatedUser represents an unauthenticated user in the A2A system.
// This implements the Null Object pattern, providing safe defaults for
// authentication operations without requiring nil checks.
//
// UnauthenticatedUser is safe to use as a zero value and is immutable.
type UnauthenticatedUser struct{}

// IsAuthenticated always returns false for unauthenticated users.
func (u UnauthenticatedUser) IsAuthenticated() bool {
	return false
}

// UserName always returns an empty string for unauthenticated users.
func (u UnauthenticatedUser) UserName() string {
	return ""
}
