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
	"time"

	a2a "github.com/go-a2a/a2a-go"
)

// ConnectionState represents the state of a client connection.
type ConnectionState int

const (
	// ConnectionStateDisconnected indicates the client is not connected.
	ConnectionStateDisconnected ConnectionState = iota
	// ConnectionStateConnecting indicates the client is establishing a connection.
	ConnectionStateConnecting
	// ConnectionStateConnected indicates the client is connected and ready.
	ConnectionStateConnected
	// ConnectionStateReconnecting indicates the client is attempting to reconnect.
	ConnectionStateReconnecting
	// ConnectionStateClosed indicates the client connection is permanently closed.
	ConnectionStateClosed
)

// String returns the string representation of the connection state.
func (s ConnectionState) String() string {
	switch s {
	case ConnectionStateDisconnected:
		return "disconnected"
	case ConnectionStateConnecting:
		return "connecting"
	case ConnectionStateConnected:
		return "connected"
	case ConnectionStateReconnecting:
		return "reconnecting"
	case ConnectionStateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// StreamEvent represents an event received from a streaming response.
type StreamEvent interface {
	// EventTime returns when the event was received.
	EventTime() time.Time
}

// MessageEvent wraps an A2A message as a stream event.
type MessageEvent struct {
	*a2a.Message
	receivedAt time.Time
}

// EventTime implements StreamEvent.
func (e *MessageEvent) EventTime() time.Time {
	return e.receivedAt
}

// ArtifactUpdateEvent wraps an artifact update as a stream event.
type ArtifactUpdateEvent struct {
	*a2a.TaskArtifactUpdateEvent
	receivedAt time.Time
}

// EventTime implements StreamEvent.
func (e *ArtifactUpdateEvent) EventTime() time.Time {
	return e.receivedAt
}

// StatusUpdateEvent wraps a status update as a stream event.
type StatusUpdateEvent struct {
	*a2a.TaskStatusUpdateEvent
	receivedAt time.Time
}

// EventTime implements StreamEvent.
func (e *StatusUpdateEvent) EventTime() time.Time {
	return e.receivedAt
}

// TaskEvent wraps a task as a stream event.
type TaskEvent struct {
	*a2a.Task
	receivedAt time.Time
}

// EventTime implements StreamEvent.
func (e *TaskEvent) EventTime() time.Time {
	return e.receivedAt
}

// ErrorEvent represents an error that occurred during streaming.
type ErrorEvent struct {
	Err        error
	receivedAt time.Time
}

// EventTime implements StreamEvent.
func (e *ErrorEvent) EventTime() time.Time {
	return e.receivedAt
}

// Error implements error interface.
func (e *ErrorEvent) Error() string {
	return e.Err.Error()
}

// Unwrap returns the underlying error.
func (e *ErrorEvent) Unwrap() error {
	return e.Err
}

// RetryConfig configures retry behavior for failed requests.
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts.
	MaxAttempts int
	// InitialDelay is the initial delay between retries.
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration
	// Multiplier is the exponential backoff multiplier.
	Multiplier float64
	// RetryableErrors defines which errors should trigger a retry.
	RetryableErrors func(error) bool
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		RetryableErrors: func(err error) bool {
			// Default: retry on network errors and 5xx status codes
			return IsRetryableError(err)
		},
	}
}

// ConnectionStateCallback is called when the connection state changes.
type ConnectionStateCallback func(oldState, newState ConnectionState)

// MessageCallback is called when a message is sent or received.
type MessageCallback func(direction string, message any)
