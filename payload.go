// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"fmt"

	"github.com/bytedance/sonic"

	"github.com/go-a2a/a2a/internal/jsonrpc2"
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
	*jsonrpc2.Request

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
	if id, ok := m["id"].(string); ok {
		var err error
		r.Request.ID, err = jsonrpc2.MakeID(id)
		if err != nil {
			return err
		}
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
func NewSendTaskRequest(id string, params TaskSendParams) *SendTaskRequest {
	jID, _ := jsonrpc2.MakeID(id)
	req, err := jsonrpc2.NewCall(jID, MethodTasksSend, params)
	if err != nil {
		panic(err)
	}
	return &SendTaskRequest{
		Request: req,
		Params:  params,
	}
}

// SendTaskResponse represents a response to a [SendTaskRequest].
type SendTaskResponse struct {
	*jsonrpc2.Response

	// Result contains the task if successful.
	Result *Task `json:"result,omitempty"`
}

// GetTaskRequest represents a request to retrieve the current state of a task.
type GetTaskRequest struct {
	*jsonrpc2.Request

	Params TaskQueryParams `json:"params"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *GetTaskRequest) UnmarshalJSON(data []byte) (err error) {
	var m map[string]any
	if err := sonic.ConfigFastest.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshal to map[string]any: %w", err)
	}

	r.Method = MethodTasksGet
	if id, ok := m["id"].(string); ok {
		r.Request.ID, err = jsonrpc2.MakeID(id)
		if err != nil {
			return err
		}
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
func NewGetTaskRequest(id string, params TaskQueryParams) *GetTaskRequest {
	jID, _ := jsonrpc2.MakeID(id)
	req, err := jsonrpc2.NewCall(jID, MethodTasksGet, params)
	if err != nil {
		panic(err)
	}
	return &GetTaskRequest{
		Request: req,
		Params:  params,
	}
}

// GetTaskResponse represents a response to a GetTaskRequest.
type GetTaskResponse struct {
	*jsonrpc2.Response

	// Result contains the task if successful.
	Result *Task `json:"result,omitempty"`
}

// CancelTaskRequest represents a request to cancel a running task.
type CancelTaskRequest struct {
	*jsonrpc2.Request

	Params TaskIDParams `json:"params"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *CancelTaskRequest) UnmarshalJSON(data []byte) (err error) {
	var m map[string]any
	if err := sonic.ConfigFastest.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshal to map[string]any: %w", err)
	}

	r.Method = MethodTasksCancel
	if id, ok := m["id"].(string); ok {
		r.Request.ID, err = jsonrpc2.MakeID(id)
		if err != nil {
			return err
		}
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
func NewCancelTaskRequest(id string, params TaskIDParams) *CancelTaskRequest {
	jID, _ := jsonrpc2.MakeID(id)
	req, err := jsonrpc2.NewCall(jID, MethodTasksCancel, params)
	if err != nil {
		panic(err)
	}
	return &CancelTaskRequest{
		Request: req,
		Params:  params,
	}
}

// CancelTaskResponse represents a response to a CancelTaskRequest.
type CancelTaskResponse struct {
	*jsonrpc2.Response

	// Result contains the updated task if successful.
	Result *Task `json:"result,omitempty"`
}

// SetTaskPushNotificationRequest represents a request to set or update push notification configuration.
type SetTaskPushNotificationRequest struct {
	*jsonrpc2.Request

	Params TaskPushNotificationConfig `json:"params"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *SetTaskPushNotificationRequest) UnmarshalJSON(data []byte) (err error) {
	var m map[string]any
	if err := sonic.ConfigFastest.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshal to map[string]any: %w", err)
	}

	r.Method = MethodTasksPushNotificationSet
	if id, ok := m["id"].(string); ok {
		r.Request.ID, err = jsonrpc2.MakeID(id)
		if err != nil {
			return err
		}
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
func NewSetTaskPushNotificationRequest(id string, params TaskPushNotificationConfig) *SetTaskPushNotificationRequest {
	jID, _ := jsonrpc2.MakeID(id)
	req, err := jsonrpc2.NewCall(jID, MethodTasksPushNotificationSet, params)
	if err != nil {
		panic(err)
	}
	return &SetTaskPushNotificationRequest{
		Request: req,
		Params:  params,
	}
}

// SetTaskPushNotificationResponse represents a response to a SetTaskPushNotificationRequest.
type SetTaskPushNotificationResponse struct {
	*jsonrpc2.Response

	// Result contains the confirmed config if successful.
	Result *TaskPushNotificationConfig `json:"result,omitempty"`
}

// GetTaskPushNotificationRequest represents a request to retrieve push notification configuration.
type GetTaskPushNotificationRequest struct {
	*jsonrpc2.Request

	Params TaskIDParams `json:"params"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *GetTaskPushNotificationRequest) UnmarshalJSON(data []byte) (err error) {
	var m map[string]any
	if err := sonic.ConfigFastest.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshal to map[string]any: %w", err)
	}

	r.Method = MethodTasksPushNotificationGet
	if id, ok := m["id"].(string); ok {
		r.Request.ID, err = jsonrpc2.MakeID(id)
		if err != nil {
			return err
		}
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
func NewGetTaskPushNotificationRequest(id string, params TaskIDParams) *GetTaskPushNotificationRequest {
	jID, _ := jsonrpc2.MakeID(id)
	req, err := jsonrpc2.NewCall(jID, MethodTasksPushNotificationGet, params)
	if err != nil {
		panic(err)
	}
	return &GetTaskPushNotificationRequest{
		Request: req,
		Params:  params,
	}
}

// GetTaskPushNotificationResponse represents a response to a [GetTaskPushNotificationRequest].
type GetTaskPushNotificationResponse struct {
	*jsonrpc2.Response

	// Result contains the push notification config if successful.
	Result *TaskPushNotificationConfig `json:"result,omitempty"`
}

// SendTaskStreamingRequest represents a request to send a task and subscribe to updates.
type SendTaskStreamingRequest struct {
	*jsonrpc2.Request

	Params TaskSendParams `json:"params"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *SendTaskStreamingRequest) UnmarshalJSON(data []byte) (err error) {
	var m map[string]any
	if err := sonic.ConfigFastest.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshal to map[string]any: %w", err)
	}

	r.Method = MethodTasksSendSubscribe
	if id, ok := m["id"].(string); ok {
		r.Request.ID, err = jsonrpc2.MakeID(id)
		if err != nil {
			return err
		}
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
func NewSendTaskStreamingRequest(id string, params TaskSendParams) *SendTaskStreamingRequest {
	jID, _ := jsonrpc2.MakeID(id)
	req, err := jsonrpc2.NewCall(jID, MethodTasksSendSubscribe, params)
	if err != nil {
		panic(err)
	}
	return &SendTaskStreamingRequest{
		Request: req,
		Params:  params,
	}
}

// SendTaskStreamingResponse represents a streaming response event for a [SendTaskStreamingRequest].
type SendTaskStreamingResponse struct {
	*jsonrpc2.Response

	// Result contains either a [TaskStatusUpdateEvent] or [TaskArtifactUpdateEvent].
	Result TaskEvent `json:"result,omitempty"`
}

// TaskResubscriptionRequest represents a request to resubscribe to task updates.
type TaskResubscriptionRequest struct {
	*jsonrpc2.Request

	Params TaskIDParams `json:"params"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *TaskResubscriptionRequest) UnmarshalJSON(data []byte) (err error) {
	var m map[string]any
	if err := sonic.ConfigFastest.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshal to map[string]any: %w", err)
	}

	r.Method = MethodTasksResubscribe
	if id, ok := m["id"].(string); ok {
		r.Request.ID, err = jsonrpc2.MakeID(id)
		if err != nil {
			return err
		}
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
func NewTaskResubscriptionRequest(id string, params TaskIDParams) *TaskResubscriptionRequest {
	jID, _ := jsonrpc2.MakeID(id)
	req, err := jsonrpc2.NewCall(jID, MethodTasksResubscribe, params)
	if err != nil {
		panic(err)
	}
	return &TaskResubscriptionRequest{
		Request: req,
		Params:  params,
	}
}
