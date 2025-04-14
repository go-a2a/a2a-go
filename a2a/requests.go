// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

// A2A RPC method names
const (
	// MethodTasksSend is the method name for sending a task.
	MethodTasksSend = "tasks/send"
	// MethodTasksSendSubscribe is the method name for sending a task and subscribing to updates.
	MethodTasksSendSubscribe = "tasks/sendSubscribe"
	// MethodTasksGet is the method name for getting a task.
	MethodTasksGet = "tasks/get"
	// MethodTasksCancel is the method name for canceling a task.
	MethodTasksCancel = "tasks/cancel"
	// MethodTasksPushNotificationSet is the method name for setting push notification configuration.
	MethodTasksPushNotificationSet = "tasks/pushNotification/set"
	// MethodTasksPushNotificationGet is the method name for getting push notification configuration.
	MethodTasksPushNotificationGet = "tasks/pushNotification/get"
	// MethodTasksResubscribe is the method name for resubscribing to task updates.
	MethodTasksResubscribe = "tasks/resubscribe"
)

// NewJSONRPCMessage creates a new JSONRPCMessage with the given ID.
func NewJSONRPCMessage(id any) JSONRPCMessage {
	return JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
	}
}

// SendTaskRequest represents a request to initiate or continue a task.
type SendTaskRequest struct {
	JSONRPCMessage
	// Method is always "tasks/send".
	Method string         `json:"method"`
	Params TaskSendParams `json:"params"`
}

// NewSendTaskRequest creates a new SendTaskRequest.
func NewSendTaskRequest(id any, params TaskSendParams) SendTaskRequest {
	return SendTaskRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksSend,
		Params:         params,
	}
}

// SendTaskResponse represents a response to a SendTaskRequest.
type SendTaskResponse struct {
	JSONRPCMessage
	// Result contains the task if successful.
	Result *Task `json:"result,omitempty"`
	// Error contains error details if the request failed.
	Error *JSONRPCError `json:"error,omitempty"`
}

// SendTaskStreamingRequest represents a request to send a task and subscribe to updates.
type SendTaskStreamingRequest struct {
	JSONRPCMessage
	// Method is always "tasks/sendSubscribe".
	Method string         `json:"method"`
	Params TaskSendParams `json:"params"`
}

// NewSendTaskStreamingRequest creates a new SendTaskStreamingRequest.
func NewSendTaskStreamingRequest(id any, params TaskSendParams) SendTaskStreamingRequest {
	return SendTaskStreamingRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksSendSubscribe,
		Params:         params,
	}
}

// SendTaskStreamingResponse represents a streaming response event for a SendTaskStreamingRequest.
type SendTaskStreamingResponse struct {
	JSONRPCMessage
	// Result contains either a TaskStatusUpdateEvent or TaskArtifactUpdateEvent.
	Result any `json:"result,omitempty"`
	// Error contains error details if the request failed.
	Error *JSONRPCError `json:"error,omitempty"`
}

// GetTaskRequest represents a request to retrieve the current state of a task.
type GetTaskRequest struct {
	JSONRPCMessage
	// Method is always "tasks/get".
	Method string          `json:"method"`
	Params TaskQueryParams `json:"params"`
}

// NewGetTaskRequest creates a new GetTaskRequest.
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
	// Error contains error details if the request failed.
	Error *JSONRPCError `json:"error,omitempty"`
}

// CancelTaskRequest represents a request to cancel a running task.
type CancelTaskRequest struct {
	JSONRPCMessage
	// Method is always "tasks/cancel".
	Method string       `json:"method"`
	Params TaskIdParams `json:"params"`
}

// NewCancelTaskRequest creates a new CancelTaskRequest.
func NewCancelTaskRequest(id any, params TaskIdParams) CancelTaskRequest {
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
	// Error contains error details if the request failed.
	Error *JSONRPCError `json:"error,omitempty"`
}

// SetTaskPushNotificationRequest represents a request to set or update push notification configuration.
type SetTaskPushNotificationRequest struct {
	JSONRPCMessage
	// Method is always "tasks/pushNotification/set".
	Method string                     `json:"method"`
	Params TaskPushNotificationConfig `json:"params"`
}

// NewSetTaskPushNotificationRequest creates a new SetTaskPushNotificationRequest.
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
	// Error contains error details if the request failed.
	Error *JSONRPCError `json:"error,omitempty"`
}

// GetTaskPushNotificationRequest represents a request to retrieve push notification configuration.
type GetTaskPushNotificationRequest struct {
	JSONRPCMessage
	// Method is always "tasks/pushNotification/get".
	Method string       `json:"method"`
	Params TaskIdParams `json:"params"`
}

// NewGetTaskPushNotificationRequest creates a new GetTaskPushNotificationRequest.
func NewGetTaskPushNotificationRequest(id any, params TaskIdParams) GetTaskPushNotificationRequest {
	return GetTaskPushNotificationRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksPushNotificationGet,
		Params:         params,
	}
}

// GetTaskPushNotificationResponse represents a response to a GetTaskPushNotificationRequest.
type GetTaskPushNotificationResponse struct {
	JSONRPCMessage
	// Result contains the push notification config if successful.
	Result *TaskPushNotificationConfig `json:"result,omitempty"`
	// Error contains error details if the request failed.
	Error *JSONRPCError `json:"error,omitempty"`
}

// TaskResubscriptionRequest represents a request to resubscribe to task updates.
type TaskResubscriptionRequest struct {
	JSONRPCMessage
	// Method is always "tasks/resubscribe".
	Method string          `json:"method"`
	Params TaskQueryParams `json:"params"`
}

// NewTaskResubscriptionRequest creates a new TaskResubscriptionRequest.
func NewTaskResubscriptionRequest(id any, params TaskQueryParams) TaskResubscriptionRequest {
	return TaskResubscriptionRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksResubscribe,
		Params:         params,
	}
}
