// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

// JSONRPCMessage is the base structure for all JSON-RPC 2.0 messages.
type JSONRPCMessage struct {
	// JSONRPC version, always "2.0".
	JSONRPC string `json:"jsonrpc"`
	// ID is a unique identifier for the request/response correlation.
	ID any `json:"id,omitempty"` // string, number, or null
}

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPCMessage
	// Method identifies the operation to perform.
	Method string `json:"method"`
	// Params contains parameters for the method.
	Params any `json:"params,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error.
type JSONRPCError struct {
	// Code is the error code.
	Code int `json:"code"`
	// Message is a short description of the error.
	Message string `json:"message"`
	// Data contains optional additional error details.
	Data any `json:"data,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPCMessage
	// Result contains the successful result data (can be null).
	// Mutually exclusive with Error.
	Result any `json:"result,omitempty"`
	// Error contains an error object if the request failed.
	// Mutually exclusive with Result.
	Error *JSONRPCError `json:"error,omitempty"`
}

// Standard JSON-RPC 2.0 error codes
const (
	// JSONParseErrorCode indicates invalid JSON payload.
	JSONParseErrorCode = -32700
	// InvalidRequestErrorCode indicates request payload validation error.
	InvalidRequestErrorCode = -32600
	// MethodNotFoundErrorCode indicates the method does not exist.
	MethodNotFoundErrorCode = -32601
	// InvalidParamsErrorCode indicates invalid method parameters.
	InvalidParamsErrorCode = -32602
	// InternalErrorCode indicates an internal server error.
	InternalErrorCode = -32603
)

// A2A specific error codes
const (
	// TaskNotFoundErrorCode indicates the specified task ID was not found.
	TaskNotFoundErrorCode = -32001
	// TaskNotCancelableErrorCode indicates the task is in a final state and cannot be canceled.
	TaskNotCancelableErrorCode = -32002
	// PushNotificationNotSupportedErrorCode indicates the agent does not support push notifications.
	PushNotificationNotSupportedErrorCode = -32003
	// UnsupportedOperationErrorCode indicates the requested operation is not supported.
	UnsupportedOperationErrorCode = -32004
	// ContentTypeNotSupportedErrorCode indicates a mismatch in supported content types.
	ContentTypeNotSupportedErrorCode = -32005
)

// NewJSONParseError creates a new JSONParseError.
func NewJSONParseError() *JSONRPCError {
	return &JSONRPCError{
		Code:    JSONParseErrorCode,
		Message: "Invalid JSON payload",
		Data:    nil,
	}
}

// NewInvalidRequestError creates a new InvalidRequestError.
func NewInvalidRequestError() *JSONRPCError {
	return &JSONRPCError{
		Code:    InvalidRequestErrorCode,
		Message: "Request payload validation error",
		Data:    nil,
	}
}

// NewMethodNotFoundError creates a new MethodNotFoundError.
func NewMethodNotFoundError() *JSONRPCError {
	return &JSONRPCError{
		Code:    MethodNotFoundErrorCode,
		Message: "Method not found",
		Data:    nil,
	}
}

// NewInvalidParamsError creates a new InvalidParamsError.
func NewInvalidParamsError() *JSONRPCError {
	return &JSONRPCError{
		Code:    InvalidParamsErrorCode,
		Message: "Invalid parameters",
		Data:    nil,
	}
}

// NewInternalError creates a new InternalError.
func NewInternalError() *JSONRPCError {
	return &JSONRPCError{
		Code:    InternalErrorCode,
		Message: "Internal error",
		Data:    nil,
	}
}

// NewTaskNotFoundError creates a new TaskNotFoundError.
func NewTaskNotFoundError() *JSONRPCError {
	return &JSONRPCError{
		Code:    TaskNotFoundErrorCode,
		Message: "Task not found",
		Data:    nil,
	}
}

// NewTaskNotCancelableError creates a new TaskNotCancelableError.
func NewTaskNotCancelableError() *JSONRPCError {
	return &JSONRPCError{
		Code:    TaskNotCancelableErrorCode,
		Message: "Task cannot be canceled",
		Data:    nil,
	}
}

// NewPushNotificationNotSupportedError creates a new PushNotificationNotSupportedError.
func NewPushNotificationNotSupportedError() *JSONRPCError {
	return &JSONRPCError{
		Code:    PushNotificationNotSupportedErrorCode,
		Message: "Push Notification is not supported",
		Data:    nil,
	}
}

// NewUnsupportedOperationError creates a new UnsupportedOperationError.
func NewUnsupportedOperationError() *JSONRPCError {
	return &JSONRPCError{
		Code:    UnsupportedOperationErrorCode,
		Message: "This operation is not supported",
		Data:    nil,
	}
}

// NewContentTypeNotSupportedError creates a new ContentTypeNotSupportedError.
func NewContentTypeNotSupportedError() *JSONRPCError {
	return &JSONRPCError{
		Code:    ContentTypeNotSupportedErrorCode,
		Message: "Content type not supported",
		Data:    nil,
	}
}
