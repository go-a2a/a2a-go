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

package handler

import (
	"errors"
	"fmt"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-json-experiment/json"
)

// Common handler errors.
var (
	// ErrMethodNotFound indicates the requested method is not registered.
	ErrMethodNotFound = errors.New("method not found")

	// ErrInvalidParams indicates the request parameters are invalid.
	ErrInvalidParams = errors.New("invalid parameters")

	// ErrUnauthorized indicates the request lacks valid authentication.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates the authenticated user lacks permission.
	ErrForbidden = errors.New("forbidden")

	// ErrTimeout indicates the request timed out.
	ErrTimeout = errors.New("request timeout")

	// ErrStreamingNotSupported indicates streaming is not supported for the method.
	ErrStreamingNotSupported = errors.New("streaming not supported")

	// ErrSessionNotFound indicates the session could not be found.
	ErrSessionNotFound = errors.New("session not found")
)

// HandlerError represents an error that occurred during request handling.
type HandlerError struct {
	// Code is the error code (usually maps to HTTP status or JSON-RPC error code).
	Code int64

	// Message is the error message.
	Message string

	// Details contains additional error details.
	Details map[string]any

	// Cause is the underlying error, if any.
	Cause error
}

// Error implements the error interface.
func (e *HandlerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *HandlerError) Unwrap() error {
	return e.Cause
}

// ToJSONRPCError converts the handler error to a JSON-RPC error.
func (e *HandlerError) ToJSONRPCError() *a2a.Error {
	a2aErr := &a2a.Error{
		Code:    e.Code,
		Message: e.Message,
	}

	if data, err := json.Marshal(e.Details); err == nil {
		a2aErr.Data = data
	}

	return a2aErr
}

// NewHandlerError creates a new handler error.
func NewHandlerError(code int64, message string) *HandlerError {
	return &HandlerError{
		Code:    code,
		Message: message,
	}
}

// WithDetails adds details to the handler error.
func (e *HandlerError) WithDetails(details map[string]any) *HandlerError {
	e.Details = details
	return e
}

// WithCause adds an underlying cause to the handler error.
func (e *HandlerError) WithCause(err error) *HandlerError {
	e.Cause = err
	return e
}

// ErrorToJSONRPC converts a standard error to a JSON-RPC error.
func ErrorToJSONRPC(err error) *a2a.Error {
	if err == nil {
		return nil
	}

	// Check if it's already a JSON-RPC error
	var jsonrpcErr *a2a.Error
	if errors.As(err, &jsonrpcErr) {
		return jsonrpcErr
	}

	// Check if it's a handler error
	var handlerErr *HandlerError
	if errors.As(err, &handlerErr) {
		return handlerErr.ToJSONRPCError()
	}

	// Map common errors
	switch {
	case errors.Is(err, ErrMethodNotFound):
		return &a2a.Error{
			Code:    a2a.ErrMethodNotFound.Code,
			Message: err.Error(),
		}
	case errors.Is(err, ErrInvalidParams):
		return &a2a.Error{
			Code:    a2a.ErrInvalidParams.Code,
			Message: err.Error(),
		}
	case errors.Is(err, ErrUnauthorized):
		return &a2a.Error{
			Code:    -32001, // Custom code for unauthorized
			Message: err.Error(),
		}
	case errors.Is(err, ErrForbidden):
		return &a2a.Error{
			Code:    -32002, // Custom code for forbidden
			Message: err.Error(),
		}
	case errors.Is(err, ErrTimeout):
		return &a2a.Error{
			Code:    -32003, // Custom code for timeout
			Message: err.Error(),
		}
	default:
		// Default to internal error
		return &a2a.Error{
			Code:    a2a.ErrInternal.Code,
			Message: err.Error(),
		}
	}
}

// ErrorToHTTPStatus converts an error to an HTTP status code.
func ErrorToHTTPStatus(err error) int {
	if err == nil {
		return 200 // OK
	}

	// Check if it's a handler error with a code
	var handlerErr *HandlerError
	if errors.As(err, &handlerErr) {
		// If code is already an HTTP status, use it
		if handlerErr.Code >= 100 && handlerErr.Code < 600 {
			return int(handlerErr.Code)
		}
	}

	// Map common errors to HTTP status codes
	switch {
	case errors.Is(err, ErrMethodNotFound):
		return 404 // Not Found
	case errors.Is(err, ErrInvalidParams):
		return 400 // Bad Request
	case errors.Is(err, ErrUnauthorized):
		return 401 // Unauthorized
	case errors.Is(err, ErrForbidden):
		return 403 // Forbidden
	case errors.Is(err, ErrTimeout):
		return 504 // Gateway Timeout
	default:
		return 500 // Internal Server Error
	}
}
