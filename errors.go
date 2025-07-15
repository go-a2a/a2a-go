// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package a2a provides error utilities for the Agent-to-Agent (A2A) protocol.
// This file contains error handling utilities that mirror the Python A2A error utils,
// converted to idiomatic Go code with proper error wrapping and unwrapping support.
package a2a

import (
	"errors"
	"fmt"
)

// Additional error codes for A2A protocol utilities.
const (
	// ErrorCodeMethodNotImplemented indicates that a method is not implemented.
	// This follows the JSON-RPC 2.0 specification pattern for custom error codes.
	ErrorCodeMethodNotImplemented = -32004

	// ErrorCodeUnsupportedOperation indicates that an operation is not supported.
	// This is used for operations that are not implemented or not available.
	ErrorCodeUnsupportedOperation = -32005
)

// ServerError represents a server-side error that wraps other A2A errors.
// It provides a way to propagate A2A-specific errors across application layers
// while maintaining error chain information and supporting Go's error unwrapping.
//
// ServerError is commonly used to:
//   - Wrap specific A2A errors for transport across layers
//   - Convert errors to protocol-specific responses (JSON-RPC, gRPC)
//   - Maintain error context and chain information
//
// Example usage:
//
//	err := TaskNotFoundError{TaskID: "task-123"}
//	serverErr := NewServerError(err)
//	// Later, unwrap to get the original error:
//	var taskErr TaskNotFoundError
//	if errors.As(serverErr, &taskErr) {
//		fmt.Printf("Task ID: %s", taskErr.TaskID)
//	}
type ServerError struct {
	// cause is the underlying A2A error that caused this server error
	cause A2AError
	// msg is an optional additional message context
	msg string
}

// NewServerError creates a new ServerError wrapping the provided A2A error.
func NewServerError(cause A2AError) *ServerError {
	return &ServerError{
		cause: cause,
	}
}

// NewServerErrorWithMessage creates a new ServerError with an additional message.
func NewServerErrorWithMessage(cause A2AError, msg string) *ServerError {
	return &ServerError{
		cause: cause,
		msg:   msg,
	}
}

// Error implements the error interface.
func (e *ServerError) Error() string {
	if e.msg != "" {
		return fmt.Sprintf("server error: %s: %s", e.msg, e.cause.Error())
	}
	return fmt.Sprintf("server error: %s", e.cause.Error())
}

// Unwrap implements Go's error unwrapping interface.
// This allows errors.Is() and errors.As() to work with the wrapped error.
func (e *ServerError) Unwrap() error {
	return e.cause
}

// Cause returns the underlying A2A error that caused this server error.
func (e *ServerError) Cause() A2AError {
	return e.cause
}

// Code returns the error code of the underlying A2A error.
func (e *ServerError) Code() int {
	return e.cause.Code()
}

// Message returns the error message of the underlying A2A error.
func (e *ServerError) Message() string {
	return e.cause.Message()
}

// Is implements error identity checking for Go's errors.Is() function.
func (e *ServerError) Is(target error) bool {
	if target == nil {
		return false
	}

	// Check if target is also a ServerError
	if se, ok := target.(*ServerError); ok {
		return errors.Is(e.cause, se.cause)
	}

	// Check if target matches the underlying cause
	return errors.Is(e.cause, target)
}

// As implements error type assertion for Go's errors.As() function.
func (e *ServerError) As(target any) bool {
	// Check if target is a ServerError pointer
	if se, ok := target.(**ServerError); ok {
		*se = e
		return true
	}

	// Check if target matches the underlying cause
	return errors.As(e.cause, target)
}

// MethodNotImplementedError represents an error when a method is not implemented.
// This is different from MethodNotFoundError, which indicates a method doesn't exist.
// MethodNotImplementedError indicates that a method exists but is not implemented
// in a particular context or implementation.
//
// This error is commonly used in:
//   - Abstract base implementations that need to be overridden
//   - Interface implementations that don't support all methods
//   - Feature flags or conditional implementations
//
// Example usage:
//
//	func (h *BaseHandler) OnMessageSendStream(ctx context.Context, req *Request) error {
//		return &MethodNotImplementedError{Method: "OnMessageSendStream"}
//	}
type MethodNotImplementedError struct {
	Method string
}

// Error implements the error interface.
func (e *MethodNotImplementedError) Error() string {
	return fmt.Sprintf("method not implemented: %s", e.Method)
}

// Code returns the error code.
func (e *MethodNotImplementedError) Code() int {
	return ErrorCodeMethodNotImplemented
}

// Message returns the error message.
func (e *MethodNotImplementedError) Message() string {
	return "Method not implemented"
}

// UnsupportedOperationError represents an error when an operation is not supported.
// This is used for operations that are not implemented or not available in the
// current context or configuration.
//
// Example usage:
//
//	if pushConfigStore == nil {
//		return &UnsupportedOperationError{Operation: "push notifications"}
//	}
type UnsupportedOperationError struct {
	Operation string
}

// Error implements the error interface.
func (e *UnsupportedOperationError) Error() string {
	return fmt.Sprintf("unsupported operation: %s", e.Operation)
}

// Code returns the error code.
func (e *UnsupportedOperationError) Code() int {
	return ErrorCodeUnsupportedOperation
}

// Message returns the error message.
func (e *UnsupportedOperationError) Message() string {
	return "Unsupported operation"
}

// Helper functions for common error operations

// IsServerError checks if an error is a ServerError.
func IsServerError(err error) bool {
	var serverErr *ServerError
	return errors.As(err, &serverErr)
}

// AsServerError attempts to convert an error to a ServerError.
// It returns the ServerError and true if successful, nil and false otherwise.
func AsServerError(err error) (*ServerError, bool) {
	var serverErr *ServerError
	if errors.As(err, &serverErr) {
		return serverErr, true
	}
	return nil, false
}

// WrapA2AError wraps an A2A error in a ServerError with an optional message.
// If the error is already a ServerError, it returns the error as-is.
// If the error is not an A2A error, it wraps it in an InternalError first.
func WrapA2AError(err error, msg string) *ServerError {
	if err == nil {
		return nil
	}

	// If it's already a ServerError, return as-is
	if se, ok := err.(*ServerError); ok {
		return se
	}

	// If it's an A2A error, wrap it
	if a2aErr, ok := err.(A2AError); ok {
		if msg != "" {
			return NewServerErrorWithMessage(a2aErr, msg)
		}
		return NewServerError(a2aErr)
	}

	// If it's not an A2A error, wrap it in an InternalError first
	internalErr := &InternalError{Msg: err.Error()}
	if msg != "" {
		return NewServerErrorWithMessage(internalErr, msg)
	}
	return NewServerError(internalErr)
}

// IsMethodNotImplementedError checks if an error is a MethodNotImplementedError.
func IsMethodNotImplementedError(err error) bool {
	var methodErr *MethodNotImplementedError
	return errors.As(err, &methodErr)
}

// AsMethodNotImplementedError attempts to convert an error to a MethodNotImplementedError.
// It returns the MethodNotImplementedError and true if successful, nil and false otherwise.
func AsMethodNotImplementedError(err error) (*MethodNotImplementedError, bool) {
	var methodErr *MethodNotImplementedError
	if errors.As(err, &methodErr) {
		return methodErr, true
	}
	return nil, false
}

// IsUnsupportedOperationError checks if an error is an UnsupportedOperationError.
func IsUnsupportedOperationError(err error) bool {
	var opErr *UnsupportedOperationError
	return errors.As(err, &opErr)
}

// AsUnsupportedOperationError attempts to convert an error to an UnsupportedOperationError.
// It returns the UnsupportedOperationError and true if successful, nil and false otherwise.
func AsUnsupportedOperationError(err error) (*UnsupportedOperationError, bool) {
	var opErr *UnsupportedOperationError
	if errors.As(err, &opErr) {
		return opErr, true
	}
	return nil, false
}

// GetA2AErrorCode extracts the error code from any A2A error, including wrapped ones.
// It returns the error code and true if successful, 0 and false otherwise.
func GetA2AErrorCode(err error) (int, bool) {
	// Check if it's directly an A2A error
	if a2aErr, ok := err.(A2AError); ok {
		return a2aErr.Code(), true
	}

	// Check if it's a ServerError (which implements A2AError)
	if serverErr, ok := err.(*ServerError); ok {
		return serverErr.Code(), true
	}

	// Try to unwrap and check again
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		return GetA2AErrorCode(unwrapped)
	}

	return 0, false
}

// GetA2AErrorMessage extracts the error message from any A2A error, including wrapped ones.
// It returns the error message and true if successful, empty string and false otherwise.
func GetA2AErrorMessage(err error) (string, bool) {
	// Check if it's directly an A2A error
	if a2aErr, ok := err.(A2AError); ok {
		return a2aErr.Message(), true
	}

	// Check if it's a ServerError (which implements A2AError)
	if serverErr, ok := err.(*ServerError); ok {
		return serverErr.Message(), true
	}

	// Try to unwrap and check again
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		return GetA2AErrorMessage(unwrapped)
	}

	return "", false
}
