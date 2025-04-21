// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
)

// A2A RPC method names.
const (
	// MethodTasksSend is the method name for sending a task.
	MethodTasksSend = "tasks/send"
	// MethodTasksGet is the method name for getting a task.
	MethodTasksGet = "tasks/get"
	// MethodTasksCancel is the method name for canceling a task.
	MethodTasksCancel = "tasks/cancel"
	// MethodTasksPushNotificationSet is the method name for setting push notification configuration.
	MethodTasksPushNotificationSet = "tasks/pushNotification/set"
	// MethodTasksPushNotificationGet is the method name for getting push notification configuration.
	MethodTasksPushNotificationGet = "tasks/pushNotification/get"
	// MethodTasksSendSubscribe is the method name for sending a task and subscribing to updates.
	MethodTasksSendSubscribe = "tasks/sendSubscribe"
	// MethodTasksResubscribe is the method name for resubscribing to task updates.
	MethodTasksResubscribe = "tasks/resubscribe"
)

// ID represents the unique identifier for JSON-RPC messages.
type ID struct {
	name   string
	number int32
}

var (
	_ fmt.Formatter    = (*ID)(nil)
	_ json.Marshaler   = (*ID)(nil)
	_ json.Unmarshaler = (*ID)(nil)
)

// NewID returns a new request ID.
func NewID[T string | int32](v T) ID {
	switch v := any(v).(type) {
	case string:
		return ID{name: v}
	case int32:
		return ID{number: v}
	default:
		panic("unreachable")
	}
}

// Format writes the ID to the formatter.
//
// If the rune is q the representation is non ambiguous,
// string forms are quoted, number forms are preceded by a #.
func (id ID) Format(f fmt.State, r rune) {
	numF, strF := `%d`, `%s`
	if r == 'q' {
		numF, strF = `#%d`, `%q`
	}

	switch {
	case id.name != "":
		fmt.Fprintf(f, strF, id.name)
	default:
		fmt.Fprintf(f, numF, id.number)
	}
}

// String returns the string representation of the ID.
func (id ID) String() string {
	return fmt.Sprint(id)
}

// MarshalJSON implements json.Marshaler.
func (id *ID) MarshalJSON() ([]byte, error) {
	if id.name != "" {
		return sonic.ConfigFastest.Marshal(id.name)
	}
	return sonic.ConfigFastest.Marshal(id.number)
}

// UnmarshalJSON implements json.Unmarshaler.
func (id *ID) UnmarshalJSON(data []byte) error {
	*id = ID{}
	if err := sonic.ConfigFastest.Unmarshal(data, &id.number); err == nil {
		return nil
	}
	return sonic.ConfigFastest.Unmarshal(data, &id.name)
}

// JSONRPCMessage is the base structure for all JSON-RPC 2.0 messages.
type JSONRPCMessage struct {
	// JSONRPC version, always "2.0".
	JSONRPC string `json:"jsonrpc"`

	// ID is a unique identifier for the request/response correlation.
	ID ID `json:"id,omitzero"` // string, number, or null
}

// NewJSONRPCMessage creates a new [JSONRPCMessage] with the given id.
func NewJSONRPCMessage(id ID) JSONRPCMessage {
	return JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
	}
}

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPCMessage

	// Method identifies the operation to perform.
	Method string `json:"method"`

	// Params contains parameters for the method.
	Params json.RawMessage `json:"params,omitempty"`
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

// Standard JSON-RPC 2.0 error codes.
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

// A2A specific error codes.
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

// JSONRPCError represents a JSON-RPC 2.0 error.
type JSONRPCError struct {
	// Code is the error code.
	Code int `json:"code"`

	// Message is a short description of the error.
	Message string `json:"message"`

	// Data contains optional additional error details.
	Data any `json:"data,omitempty"`
}

// NewJSONParseError creates a new JSONParseError.
func NewJSONParseError() *JSONRPCError {
	return &JSONRPCError{
		Code:    JSONParseErrorCode,
		Message: "Invalid JSON payload",
	}
}

// NewInvalidRequestError creates a new InvalidRequestError.
func NewInvalidRequestError() *JSONRPCError {
	return &JSONRPCError{
		Code:    InvalidRequestErrorCode,
		Message: "Request payload validation error",
	}
}

// NewMethodNotFoundError creates a new MethodNotFoundError.
func NewMethodNotFoundError() *JSONRPCError {
	return &JSONRPCError{
		Code:    MethodNotFoundErrorCode,
		Message: "Method not found",
	}
}

// NewInvalidParamsError creates a new InvalidParamsError.
func NewInvalidParamsError() *JSONRPCError {
	return &JSONRPCError{
		Code:    InvalidParamsErrorCode,
		Message: "Invalid parameters",
	}
}

// NewInternalError creates a new InternalError.
func NewInternalError() *JSONRPCError {
	return &JSONRPCError{
		Code:    InternalErrorCode,
		Message: "Internal error",
	}
}

// NewTaskNotFoundError creates a new TaskNotFoundError.
func NewTaskNotFoundError() *JSONRPCError {
	return &JSONRPCError{
		Code:    TaskNotFoundErrorCode,
		Message: "Task not found",
	}
}

// NewTaskNotCancelableError creates a new TaskNotCancelableError.
func NewTaskNotCancelableError() *JSONRPCError {
	return &JSONRPCError{
		Code:    TaskNotCancelableErrorCode,
		Message: "Task cannot be canceled",
	}
}

// NewPushNotificationNotSupportedError creates a new PushNotificationNotSupportedError.
func NewPushNotificationNotSupportedError() *JSONRPCError {
	return &JSONRPCError{
		Code:    PushNotificationNotSupportedErrorCode,
		Message: "Push Notification is not supported",
	}
}

// NewUnsupportedOperationError creates a new UnsupportedOperationError.
func NewUnsupportedOperationError() *JSONRPCError {
	return &JSONRPCError{
		Code:    UnsupportedOperationErrorCode,
		Message: "This operation is not supported",
	}
}

// NewContentTypeNotSupportedError creates a new ContentTypeNotSupportedError.
func NewContentTypeNotSupportedError() *JSONRPCError {
	return &JSONRPCError{
		Code:    ContentTypeNotSupportedErrorCode,
		Message: "Content type not supported",
	}
}
