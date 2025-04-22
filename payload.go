// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
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

// SendTaskRequest represents a request to initiate or continue a task.
type SendTaskRequest struct {
	JSONRPCRequest

	// Method is always "tasks/send".
	Params TaskSendParams `json:"params"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *SendTaskRequest) UnmarshalJSON(data []byte) error {
	var m map[string]any
	if err := sonic.ConfigFastest.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshal to map[string]any: %w", err)
	}

	r.Method = MethodTasksSend
	r.JSONRPCMessage = JSONRPCMessage{
		JSONRPC: "2.0",
	}
	if id, ok := m["id"].(string); ok {
		r.JSONRPCMessage.ID = NewID(id)
	}

	paramsData, err := sonic.ConfigFastest.Marshal(m["params"])
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}

	var rr TaskSendParams
	if err := sonic.ConfigFastest.Unmarshal(paramsData, &rr); err != nil {
		return fmt.Errorf("unmarshal to TaskSendParams: %w", err)
	}
	r.Params = rr

	return nil
}

// NewSendTaskRequest creates a new [SendTaskRequest].
func NewSendTaskRequest(id ID, params TaskSendParams) *SendTaskRequest {
	return &SendTaskRequest{
		JSONRPCRequest: JSONRPCRequest{
			JSONRPCMessage: NewJSONRPCMessage(id),
			Method:         MethodTasksSend,
		},
		Params: params,
	}
}

// SendTaskResponse represents a response to a [SendTaskRequest].
type SendTaskResponse struct {
	JSONRPCResponse

	// Result contains the task if successful.
	Result *Task `json:"result,omitempty"`
}

// GetTaskRequest represents a request to retrieve the current state of a task.
type GetTaskRequest struct {
	JSONRPCRequest

	Params TaskQueryParams `json:"params"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *GetTaskRequest) UnmarshalJSON(data []byte) error {
	var m map[string]any
	if err := sonic.ConfigFastest.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshal to map[string]any: %w", err)
	}

	r.Method = MethodTasksGet
	r.JSONRPCMessage = JSONRPCMessage{
		JSONRPC: "2.0",
	}
	if id, ok := m["id"].(string); ok {
		r.JSONRPCMessage.ID = NewID(id)
	}

	paramsData, err := sonic.ConfigFastest.Marshal(m["params"])
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}

	var rr TaskQueryParams
	if err := sonic.ConfigFastest.Unmarshal(paramsData, &rr); err != nil {
		return fmt.Errorf("unmarshal to TaskSendParams: %w", err)
	}
	r.Params = rr

	return nil
}

// NewGetTaskRequest creates a new [GetTaskRequest].
func NewGetTaskRequest(id ID, params TaskQueryParams) *GetTaskRequest {
	return &GetTaskRequest{
		JSONRPCRequest: JSONRPCRequest{
			JSONRPCMessage: NewJSONRPCMessage(id),
			Method:         MethodTasksGet,
		},
		Params: params,
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
	JSONRPCRequest

	Params TaskIDParams `json:"params"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *CancelTaskRequest) UnmarshalJSON(data []byte) error {
	var m map[string]any
	if err := sonic.ConfigFastest.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshal to map[string]any: %w", err)
	}

	r.Method = MethodTasksCancel
	r.JSONRPCMessage = JSONRPCMessage{
		JSONRPC: "2.0",
	}
	if id, ok := m["id"].(string); ok {
		r.JSONRPCMessage.ID = NewID(id)
	}

	paramsData, err := sonic.ConfigFastest.Marshal(m["params"])
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}

	var rr TaskIDParams
	if err := sonic.ConfigFastest.Unmarshal(paramsData, &rr); err != nil {
		return fmt.Errorf("unmarshal to TaskSendParams: %w", err)
	}
	r.Params = rr

	return nil
}

// NewCancelTaskRequest creates a new [CancelTaskRequest].
func NewCancelTaskRequest(id ID, params TaskIDParams) *CancelTaskRequest {
	return &CancelTaskRequest{
		JSONRPCRequest: JSONRPCRequest{
			JSONRPCMessage: NewJSONRPCMessage(id),
			Method:         MethodTasksCancel,
		},
		Params: params,
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
	JSONRPCRequest

	Params TaskPushNotificationConfig `json:"params"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *SetTaskPushNotificationRequest) UnmarshalJSON(data []byte) error {
	var m map[string]any
	if err := sonic.ConfigFastest.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshal to map[string]any: %w", err)
	}

	r.Method = MethodTasksPushNotificationSet
	r.JSONRPCMessage = JSONRPCMessage{
		JSONRPC: "2.0",
	}
	if id, ok := m["id"].(string); ok {
		r.JSONRPCMessage.ID = NewID(id)
	}

	paramsData, err := sonic.ConfigFastest.Marshal(m["params"])
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}

	var rr TaskPushNotificationConfig
	if err := sonic.ConfigFastest.Unmarshal(paramsData, &rr); err != nil {
		return fmt.Errorf("unmarshal to TaskSendParams: %w", err)
	}
	r.Params = rr

	return nil
}

// NewSetTaskPushNotificationRequest creates a new [SetTaskPushNotificationRequest].
func NewSetTaskPushNotificationRequest(id ID, params TaskPushNotificationConfig) *SetTaskPushNotificationRequest {
	return &SetTaskPushNotificationRequest{
		JSONRPCRequest: JSONRPCRequest{
			JSONRPCMessage: NewJSONRPCMessage(id),
			Method:         MethodTasksPushNotificationSet,
		},
		Params: params,
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
	JSONRPCRequest

	Params TaskIDParams `json:"params"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *GetTaskPushNotificationRequest) UnmarshalJSON(data []byte) error {
	var m map[string]any
	if err := sonic.ConfigFastest.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshal to map[string]any: %w", err)
	}

	r.Method = MethodTasksPushNotificationGet
	r.JSONRPCMessage = JSONRPCMessage{
		JSONRPC: "2.0",
	}
	if id, ok := m["id"].(string); ok {
		r.JSONRPCMessage.ID = NewID(id)
	}

	paramsData, err := sonic.ConfigFastest.Marshal(m["params"])
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}

	var rr TaskIDParams
	if err := sonic.ConfigFastest.Unmarshal(paramsData, &rr); err != nil {
		return fmt.Errorf("unmarshal to TaskSendParams: %w", err)
	}
	r.Params = rr

	return nil
}

// NewGetTaskPushNotificationRequest creates a new [GetTaskPushNotificationRequest].
func NewGetTaskPushNotificationRequest(id ID, params TaskIDParams) *GetTaskPushNotificationRequest {
	return &GetTaskPushNotificationRequest{
		JSONRPCRequest: JSONRPCRequest{
			JSONRPCMessage: NewJSONRPCMessage(id),
			Method:         MethodTasksPushNotificationGet,
		},
		Params: params,
	}
}

// GetTaskPushNotificationResponse represents a response to a [GetTaskPushNotificationRequest].
type GetTaskPushNotificationResponse struct {
	JSONRPCResponse

	// Result contains the push notification config if successful.
	Result *TaskPushNotificationConfig `json:"result,omitempty"`
}

// SendTaskStreamingRequest represents a request to send a task and subscribe to updates.
type SendTaskStreamingRequest struct {
	JSONRPCRequest

	Params TaskSendParams `json:"params"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *SendTaskStreamingRequest) UnmarshalJSON(data []byte) error {
	var m map[string]any
	if err := sonic.ConfigFastest.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshal to map[string]any: %w", err)
	}

	r.Method = MethodTasksSendSubscribe
	r.JSONRPCMessage = JSONRPCMessage{
		JSONRPC: "2.0",
	}
	if id, ok := m["id"].(string); ok {
		r.JSONRPCMessage.ID = NewID(id)
	}

	paramsData, err := sonic.ConfigFastest.Marshal(m["params"])
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}

	var rr TaskSendParams
	if err := sonic.ConfigFastest.Unmarshal(paramsData, &rr); err != nil {
		return fmt.Errorf("unmarshal to TaskSendParams: %w", err)
	}
	r.Params = rr

	return nil
}

// NewSendTaskStreamingRequest creates a new [SendTaskStreamingRequest].
func NewSendTaskStreamingRequest(id ID, params TaskSendParams) *SendTaskStreamingRequest {
	return &SendTaskStreamingRequest{
		JSONRPCRequest: JSONRPCRequest{
			JSONRPCMessage: NewJSONRPCMessage(id),
			Method:         MethodTasksSendSubscribe,
		},
		Params: params,
	}
}

// SendTaskStreamingResponse represents a streaming response event for a [SendTaskStreamingRequest].
type SendTaskStreamingResponse struct {
	JSONRPCResponse

	// Result contains either a [TaskStatusUpdateEvent] or [TaskArtifactUpdateEvent].
	Result TaskEvent `json:"result,omitempty"`
}

// TaskResubscriptionRequest represents a request to resubscribe to task updates.
type TaskResubscriptionRequest struct {
	JSONRPCRequest

	Params TaskIDParams `json:"params"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *TaskResubscriptionRequest) UnmarshalJSON(data []byte) error {
	var m map[string]any
	if err := sonic.ConfigFastest.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshal to map[string]any: %w", err)
	}

	r.Method = MethodTasksResubscribe
	r.JSONRPCMessage = JSONRPCMessage{
		JSONRPC: "2.0",
	}
	if id, ok := m["id"].(string); ok {
		r.JSONRPCMessage.ID = NewID(id)
	}

	paramsData, err := sonic.ConfigFastest.Marshal(m["params"])
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}

	var rr TaskIDParams
	if err := sonic.ConfigFastest.Unmarshal(paramsData, &rr); err != nil {
		return fmt.Errorf("unmarshal to TaskSendParams: %w", err)
	}
	r.Params = rr

	return nil
}

// NewTaskResubscriptionRequest creates a new [TaskResubscriptionRequest].
func NewTaskResubscriptionRequest(id ID, params TaskIDParams) *TaskResubscriptionRequest {
	return &TaskResubscriptionRequest{
		JSONRPCRequest: JSONRPCRequest{
			JSONRPCMessage: NewJSONRPCMessage(id),
			Method:         MethodTasksResubscribe,
		},
		Params: params,
	}
}
