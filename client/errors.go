// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
)

// Standard JSON-RPC 2.0 error codes
const (
	// ErrorCodeParse is returned when the server encountered a parse error (-32700)
	ErrorCodeParse = -32700
	// ErrorCodeInvalidRequest is returned when the request was invalid (-32600)
	ErrorCodeInvalidRequest = -32600
	// ErrorCodeMethodNotFound is returned when the method doesn't exist (-32601)
	ErrorCodeMethodNotFound = -32601
	// ErrorCodeInvalidParams is returned when the parameters are invalid (-32602)
	ErrorCodeInvalidParams = -32602
	// ErrorCodeInternal is returned when there was an internal server error (-32603)
	ErrorCodeInternal = -32603
)

// A2A-specific error codes
const (
	// ErrorCodeTaskNotFound is returned when the specified task ID was not found (-32001)
	ErrorCodeTaskNotFound = -32001
	// ErrorCodeTaskNotCancelable is returned when the task is in a final state and cannot be canceled (-32002)
	ErrorCodeTaskNotCancelable = -32002
	// ErrorCodePushNotificationNotSupported is returned when the agent does not support push notifications (-32003)
	ErrorCodePushNotificationNotSupported = -32003
	// ErrorCodeUnsupportedOperation is returned when the requested operation is not supported (-32004)
	ErrorCodeUnsupportedOperation = -32004
	// ErrorCodeContentTypeNotSupported is returned when there's a mismatch in supported content types (-32005)
	ErrorCodeContentTypeNotSupported = -32005
)

// RPCError represents a JSON-RPC 2.0 error.
type RPCError struct {
	// Code is the error code
	Code int `json:"code"`
	// Message is the error message
	Message string `json:"message"`
	// Data is optional additional information about the error
	Data any `json:"data,omitzero"`
}

// Error implements the error interface.
func (e *RPCError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("rpc error: code = %d, message = %s, data = %v", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("rpc error: code = %d, message = %s", e.Code, e.Message)
}

// NewRPCError creates a new RPCError.
func NewRPCError(code int, message string, data any) *RPCError {
	return &RPCError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// IsRPCError checks if an error is an RPCError with the specified code.
func IsRPCError(err error, code int) bool {
	if err == nil {
		return false
	}
	if rpcErr, ok := err.(*RPCError); ok {
		return rpcErr.Code == code
	}
	return false
}

// IsTaskNotFoundError checks if an error is due to a task not being found.
func IsTaskNotFoundError(err error) bool {
	return IsRPCError(err, ErrorCodeTaskNotFound)
}

// IsTaskNotCancelableError checks if an error is due to a task not being cancelable.
func IsTaskNotCancelableError(err error) bool {
	return IsRPCError(err, ErrorCodeTaskNotCancelable)
}

// IsPushNotificationNotSupportedError checks if an error is due to push notifications not being supported.
func IsPushNotificationNotSupportedError(err error) bool {
	return IsRPCError(err, ErrorCodePushNotificationNotSupported)
}

// IsUnsupportedOperationError checks if an error is due to an unsupported operation.
func IsUnsupportedOperationError(err error) bool {
	return IsRPCError(err, ErrorCodeUnsupportedOperation)
}

// IsContentTypeNotSupportedError checks if an error is due to a content type not being supported.
func IsContentTypeNotSupportedError(err error) bool {
	return IsRPCError(err, ErrorCodeContentTypeNotSupported)
}
