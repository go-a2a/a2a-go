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
	"errors"
	"fmt"
	"net"
	"net/url"

	a2a "github.com/go-a2a/a2a-go"
)

// Error codes that map to Python client exceptions.
const (
	ErrCodeConnectionFailed     = -1
	ErrCodeTimeout              = -2
	ErrCodeInvalidConfiguration = -3
	ErrCodeAuthenticationFailed = -4
	ErrCodeStreamError          = -5

	// A2A protocol error codes
	ErrCodeTaskNotFound                 = -32001
	ErrCodeTaskNotCancelable            = -32002
	ErrCodePushNotificationNotSupported = -32003
	ErrCodeUnsupportedOperation         = -32004
	ErrCodeContentTypeNotSupported      = -32005
	ErrCodeInvalidAgentResponse         = -32006
	ErrCodeMethodNotImplemented         = -32099
)

// Common errors.
var (
	// ErrClientClosed is returned when operations are attempted on a closed client.
	ErrClientClosed = errors.New("client is closed")

	// ErrNotConnected is returned when operations require an active connection.
	ErrNotConnected = errors.New("client is not connected")

	// ErrStreamClosed is returned when reading from a closed stream.
	ErrStreamClosed = errors.New("stream is closed")

	// ErrInvalidResponse is returned when the server response is malformed.
	ErrInvalidResponse = errors.New("invalid server response")

	// ErrNoAgentCard is returned when agent card is not available.
	ErrNoAgentCard = errors.New("agent card not available")
)

// Error represents a client error with code and message.
type Error struct {
	Code    int64
	Message string
	Cause   error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("pyclient error %d: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("pyclient error %d: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is implements errors.Is for Error.
func (e *Error) Is(target error) bool {
	var targetErr *Error
	if errors.As(target, &targetErr) {
		return e.Code == targetErr.Code
	}
	return false
}

// NewError creates a new Error with the given code and message.
func NewError(code int64, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// NewErrorWithCause creates a new Error with the given code, message, and cause.
func NewErrorWithCause(code int64, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// ConnectionError represents a connection-related error.
type ConnectionError struct {
	Operation string
	URL       string
	Err       error
}

// Error implements the error interface.
func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error during %s to %s: %v", e.Operation, e.URL, e.Err)
}

// Unwrap returns the underlying error.
func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// TimeoutError represents a timeout error.
type TimeoutError struct {
	Operation string
	Duration  string
}

// Error implements the error interface.
func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout during %s after %s", e.Operation, e.Duration)
}

// Timeout implements net.Error.
func (e *TimeoutError) Timeout() bool {
	return true
}

// Temporary implements net.Error.
func (e *TimeoutError) Temporary() bool {
	return true
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for field %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// IsRetryableError determines if an error should trigger a retry.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Temporary() {
		return true
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Temporary() {
		return true
	}

	// Check for A2A errors that might be retryable
	var a2aErr *a2a.Error
	if errors.As(err, &a2aErr) {
		switch a2aErr.Code {
		case -32000: // Server overloaded
			return true
		}
	}

	// Check for connection errors
	var connErr *ConnectionError
	if errors.As(err, &connErr) {
		return true
	}

	return false
}

// ConvertA2AError converts an A2A protocol error to a client Error.
func ConvertA2AError(err *a2a.Error) *Error {
	return &Error{
		Code:    err.Code,
		Message: err.Message,
		Cause:   err,
	}
}

// WrapError wraps an error with additional context.
func WrapError(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}
