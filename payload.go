// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

// SendTaskRequest represents a request to initiate or continue a task.
type SendTaskRequest struct {
	JSONRPCMessage

	// Method is always "tasks/send".
	Method string         `json:"method"`
	Params TaskSendParams `json:"params"`
}

// MethodName implements [A2ARequest].
func (*SendTaskRequest) MethodName() string {
	return MethodTasksSend
}

// NewSendTaskRequest creates a new [SendTaskRequest].
func NewSendTaskRequest(id ID, params TaskSendParams) SendTaskRequest {
	return SendTaskRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksSend,
		Params:         params,
	}
}

// SendTaskResponse represents a response to a [SendTaskRequest].
type SendTaskResponse struct {
	JSONRPCResponse

	// Result contains the task if successful.
	Result *Task `json:"result,omitempty"`
}

// NewSendTaskResponse creates a new [SendTaskResponse].
func NewSendTaskResponse(id ID, result *Task) SendTaskResponse {
	return SendTaskResponse{
		JSONRPCResponse: JSONRPCResponse{
			JSONRPCMessage: NewJSONRPCMessage(id),
		},
		Result: result,
	}
}

// SendTaskStreamingRequest represents a request to send a task and subscribe to updates.
type SendTaskStreamingRequest struct {
	JSONRPCMessage

	// Method is always "tasks/sendSubscribe".
	Method string         `json:"method"`
	Params TaskSendParams `json:"params"`
}

// MethodName implements [A2ARequest].
func (*SendTaskStreamingRequest) MethodName() string {
	return MethodTasksSendSubscribe
}

// NewSendTaskStreamingRequest creates a new [SendTaskStreamingRequest].
func NewSendTaskStreamingRequest(id ID, params TaskSendParams) SendTaskStreamingRequest {
	return SendTaskStreamingRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksSendSubscribe,
		Params:         params,
	}
}

// SendTaskStreamingResponse represents a streaming response event for a [SendTaskStreamingRequest].
type SendTaskStreamingResponse struct {
	JSONRPCResponse

	// Result contains either a [TaskStatusUpdateEvent] or [TaskArtifactUpdateEvent].
	Result TaskEvent `json:"result,omitempty"`
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

// MethodName implements [A2ARequest].
func (*GetTaskRequest) MethodName() string {
	return MethodTasksGet
}

// NewGetTaskRequest creates a new [GetTaskRequest].
func NewGetTaskRequest(id ID, params TaskQueryParams) GetTaskRequest {
	return GetTaskRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksGet,
		Params:         params,
	}
}

// GetTaskResponse represents a response to a GetTaskRequest.
type GetTaskResponse struct {
	JSONRPCResponse

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

// MethodName implements [A2ARequest].
func (*CancelTaskRequest) MethodName() string {
	return MethodTasksCancel
}

// NewCancelTaskRequest creates a new [CancelTaskRequest].
func NewCancelTaskRequest(id ID, params TaskIDParams) CancelTaskRequest {
	return CancelTaskRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksCancel,
		Params:         params,
	}
}

// CancelTaskResponse represents a response to a CancelTaskRequest.
type CancelTaskResponse struct {
	JSONRPCResponse

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

// MethodName implements [A2ARequest].
func (*SetTaskPushNotificationRequest) MethodName() string {
	return MethodTasksPushNotificationSet
}

// NewSetTaskPushNotificationRequest creates a new [SetTaskPushNotificationRequest].
func NewSetTaskPushNotificationRequest(id ID, params TaskPushNotificationConfig) SetTaskPushNotificationRequest {
	return SetTaskPushNotificationRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksPushNotificationSet,
		Params:         params,
	}
}

// SetTaskPushNotificationResponse represents a response to a SetTaskPushNotificationRequest.
type SetTaskPushNotificationResponse struct {
	JSONRPCResponse

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

// MethodName implements [A2ARequest].
func (*GetTaskPushNotificationRequest) MethodName() string {
	return MethodTasksPushNotificationGet
}

// NewGetTaskPushNotificationRequest creates a new [GetTaskPushNotificationRequest].
func NewGetTaskPushNotificationRequest(id ID, params TaskIDParams) GetTaskPushNotificationRequest {
	return GetTaskPushNotificationRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksPushNotificationGet,
		Params:         params,
	}
}

// GetTaskPushNotificationResponse represents a response to a [GetTaskPushNotificationRequest].
type GetTaskPushNotificationResponse struct {
	JSONRPCResponse

	// Result contains the push notification config if successful.
	Result *TaskPushNotificationConfig `json:"result,omitempty"`
}

// TaskResubscriptionRequest represents a request to resubscribe to task updates.
type TaskResubscriptionRequest struct {
	JSONRPCMessage

	// Method is always "tasks/resubscribe".
	Method string       `json:"method"`
	Params TaskIDParams `json:"params"`
}

// MethodName implements [A2ARequest].
func (*TaskResubscriptionRequest) MethodName() string {
	return MethodTasksResubscribe
}

// NewTaskResubscriptionRequest creates a new [TaskResubscriptionRequest].
func NewTaskResubscriptionRequest(id ID, params TaskIDParams) TaskResubscriptionRequest {
	return TaskResubscriptionRequest{
		JSONRPCMessage: NewJSONRPCMessage(id),
		Method:         MethodTasksResubscribe,
		Params:         params,
	}
}
