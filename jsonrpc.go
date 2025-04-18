// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"encoding/json"
	"strconv"
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
	any
}

func (id ID) String() string {
	switch id := id.any.(type) {
	case string:
		return id
	case float64:
		return strconv.FormatFloat(id, 'f', 0, 64)
	default:
		panic("unreachable")
	}
}

// JSONRPCMessage is the base structure for all JSON-RPC 2.0 messages.
type JSONRPCMessage struct {
	// JSONRPC version, always "2.0".
	JSONRPC string `json:"jsonrpc"`
	// ID is a unique identifier for the request/response correlation.
	ID ID `json:"id,omitzero"` // string, number, or null
}

// NewJSONRPCMessage creates a new [JSONRPCMessage] with the given id.
func NewJSONRPCMessage(id any) JSONRPCMessage {
	return JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      ID{id},
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

// A2ARequest represents a request to the A2A API.
type A2ARequest interface {
	MethodName() string
}

// SendTaskRequest represents a request to initiate or continue a task.
type SendTaskRequest struct {
	JSONRPCMessage

	// Method is always "tasks/send".
	Method string         `json:"method"`
	Params TaskSendParams `json:"params"`
}

var _ A2ARequest = (*SendTaskRequest)(nil)

// MethodName implements [A2ARequest].
func (*SendTaskRequest) MethodName() string {
	return MethodTasksSend
}

// NewSendTaskRequest creates a new [SendTaskRequest].
func NewSendTaskRequest(id any, params TaskSendParams) SendTaskRequest {
	return SendTaskRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksSend,
		Params:         params,
	}
}

// SendTaskResponse represents a response to a [SendTaskRequest].
type SendTaskResponse struct {
	JSONRPCMessage

	// Result contains the task if successful.
	Result *Task `json:"result,omitempty"`
	// Error contains error details if the request failed.
	Error *JSONRPCError `json:"error,omitempty"`
}

// NewSendTaskResponse creates a new [SendTaskResponse].
func NewSendTaskResponse(id any, result *Task) SendTaskResponse {
	return SendTaskResponse{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Result:         result,
	}
}

// SendTaskStreamingRequest represents a request to send a task and subscribe to updates.
type SendTaskStreamingRequest struct {
	JSONRPCMessage

	// Method is always "tasks/sendSubscribe".
	Method string         `json:"method"`
	Params TaskSendParams `json:"params"`
}

var _ A2ARequest = (*SendTaskStreamingRequest)(nil)

// MethodName implements [A2ARequest].
func (*SendTaskStreamingRequest) MethodName() string {
	return MethodTasksSendSubscribe
}

// NewSendTaskStreamingRequest creates a new [SendTaskStreamingRequest].
func NewSendTaskStreamingRequest(id any, params TaskSendParams) SendTaskStreamingRequest {
	return SendTaskStreamingRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksSendSubscribe,
		Params:         params,
	}
}

// SendTaskStreamingResponse represents a streaming response event for a [SendTaskStreamingRequest].
type SendTaskStreamingResponse struct {
	JSONRPCMessage

	// Result contains either a [TaskStatusUpdateEvent] or [TaskArtifactUpdateEvent].
	Result any `json:"result,omitempty"`
}

// GetTaskRequest represents a request to retrieve the current state of a task.
type GetTaskRequest struct {
	JSONRPCMessage

	// Method is always "tasks/get".
	Method string          `json:"method"`
	Params TaskQueryParams `json:"params"`
}

var _ A2ARequest = (*GetTaskRequest)(nil)

// MethodName implements [A2ARequest].
func (*GetTaskRequest) MethodName() string {
	return MethodTasksGet
}

// NewGetTaskRequest creates a new [GetTaskRequest].
func NewGetTaskRequest(id any, params TaskQueryParams) GetTaskRequest {
	return GetTaskRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksGet,
		Params:         params,
	}
}

// GetTaskResponse represents a response to a GetTaskRequest.
type GetTaskResponse struct {
	JSONRPCMessage

	// Result contains the task if successful.
	Result *Task `json:"result,omitempty"`
}

// CancelTaskRequest represents a request to cancel a running task.
type CancelTaskRequest struct {
	JSONRPCMessage

	// Method is always "tasks/cancel".
	Method string       `json:"method"`
	Params TaskIDParams `json:"params"`
}

var _ A2ARequest = (*CancelTaskRequest)(nil)

// MethodName implements [A2ARequest].
func (*CancelTaskRequest) MethodName() string {
	return MethodTasksCancel
}

// NewCancelTaskRequest creates a new [CancelTaskRequest].
func NewCancelTaskRequest(id any, params TaskIDParams) CancelTaskRequest {
	return CancelTaskRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksCancel,
		Params:         params,
	}
}

// CancelTaskResponse represents a response to a CancelTaskRequest.
type CancelTaskResponse struct {
	JSONRPCMessage

	// Result contains the updated task if successful.
	Result *Task `json:"result,omitempty"`
}

// SetTaskPushNotificationRequest represents a request to set or update push notification configuration.
type SetTaskPushNotificationRequest struct {
	JSONRPCMessage

	// Method is always "tasks/pushNotification/set".
	Method string                     `json:"method"`
	Params TaskPushNotificationConfig `json:"params"`
}

var _ A2ARequest = (*SetTaskPushNotificationRequest)(nil)

// MethodName implements [A2ARequest].
func (*SetTaskPushNotificationRequest) MethodName() string {
	return MethodTasksPushNotificationSet
}

// NewSetTaskPushNotificationRequest creates a new [SetTaskPushNotificationRequest].
func NewSetTaskPushNotificationRequest(id any, params TaskPushNotificationConfig) SetTaskPushNotificationRequest {
	return SetTaskPushNotificationRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksPushNotificationSet,
		Params:         params,
	}
}

// SetTaskPushNotificationResponse represents a response to a SetTaskPushNotificationRequest.
type SetTaskPushNotificationResponse struct {
	JSONRPCMessage

	// Result contains the confirmed config if successful.
	Result *TaskPushNotificationConfig `json:"result,omitempty"`
}

// GetTaskPushNotificationRequest represents a request to retrieve push notification configuration.
type GetTaskPushNotificationRequest struct {
	JSONRPCMessage

	// Method is always "tasks/pushNotification/get".
	Method string       `json:"method"`
	Params TaskIDParams `json:"params"`
}

var _ A2ARequest = (*GetTaskPushNotificationRequest)(nil)

// MethodName implements [A2ARequest].
func (*GetTaskPushNotificationRequest) MethodName() string {
	return MethodTasksPushNotificationGet
}

// NewGetTaskPushNotificationRequest creates a new [GetTaskPushNotificationRequest].
func NewGetTaskPushNotificationRequest(id any, params TaskIDParams) GetTaskPushNotificationRequest {
	return GetTaskPushNotificationRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksPushNotificationGet,
		Params:         params,
	}
}

// GetTaskPushNotificationResponse represents a response to a [GetTaskPushNotificationRequest].
type GetTaskPushNotificationResponse struct {
	JSONRPCMessage

	// Result contains the push notification config if successful.
	Result *TaskPushNotificationConfig `json:"result,omitempty"`
}

// TaskResubscriptionRequest represents a request to resubscribe to task updates.
type TaskResubscriptionRequest struct {
	JSONRPCMessage

	// Method is always "tasks/resubscribe".
	Method string          `json:"method"`
	Params TaskQueryParams `json:"params"`
}

var _ A2ARequest = (*TaskResubscriptionRequest)(nil)

// MethodName implements [A2ARequest].
func (*TaskResubscriptionRequest) MethodName() string {
	return MethodTasksResubscribe
}

// NewTaskResubscriptionRequest creates a new [TaskResubscriptionRequest].
func NewTaskResubscriptionRequest(id any, params TaskQueryParams) TaskResubscriptionRequest {
	return TaskResubscriptionRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksResubscribe,
		Params:         params,
	}
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
