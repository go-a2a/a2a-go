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

package client

import (
	"github.com/google/uuid"

	a2a "github.com/go-a2a/a2a-go"
)

// SendMessageRequest represents a request to send a message to an agent.
type SendMessageRequest struct {
	ID            string                        `json:"id,omitempty"`
	Message       *a2a.Message                  `json:"message"`
	Configuration *a2a.MessageSendConfiguration `json:"configuration,omitempty"`
}

// SendMessageResponse represents the response from sending a message.
type SendMessageResponse struct {
	ID     string    `json:"id"`
	Result any       `json:"result,omitempty"` // Can be Task or Message
	Error  *RPCError `json:"error,omitempty"`
}

// SendStreamingMessageRequest represents a request to send a streaming message.
type SendStreamingMessageRequest struct {
	ID            string                        `json:"id,omitempty"`
	Message       *a2a.Message                  `json:"message"`
	Configuration *a2a.MessageSendConfiguration `json:"configuration,omitempty"`
}

// SendStreamingMessageResponse represents a streaming response from sending a message.
type SendStreamingMessageResponse struct {
	ID     string    `json:"id"`
	Result any       `json:"result,omitempty"` // Can be Task, Message, or events
	Error  *RPCError `json:"error,omitempty"`
}

// GetTaskRequest represents a request to get a task.
type GetTaskRequest struct {
	ID            string `json:"id,omitempty"`
	TaskID        string `json:"task_id"`
	HistoryLength int32  `json:"history_length,omitempty"`
}

// GetTaskResponse represents the response from getting a task.
type GetTaskResponse struct {
	ID     string    `json:"id"`
	Result *a2a.Task `json:"result,omitempty"`
	Error  *RPCError `json:"error,omitempty"`
}

// CancelTaskRequest represents a request to cancel a task.
type CancelTaskRequest struct {
	ID     string `json:"id,omitempty"`
	TaskID string `json:"task_id"`
}

// CancelTaskResponse represents the response from canceling a task.
type CancelTaskResponse struct {
	ID     string    `json:"id"`
	Result *a2a.Task `json:"result,omitempty"`
	Error  *RPCError `json:"error,omitempty"`
}

// SetTaskPushNotificationConfigRequest represents a request to set task push notification config.
type SetTaskPushNotificationConfigRequest struct {
	ID     string                          `json:"id,omitempty"`
	TaskID string                          `json:"task_id"`
	Config *a2a.TaskPushNotificationConfig `json:"config"`
}

// SetTaskPushNotificationConfigResponse represents the response from setting task push notification config.
type SetTaskPushNotificationConfigResponse struct {
	ID     string    `json:"id"`
	Result bool      `json:"result,omitempty"`
	Error  *RPCError `json:"error,omitempty"`
}

// GetTaskPushNotificationConfigRequest represents a request to get task push notification config.
type GetTaskPushNotificationConfigRequest struct {
	ID     string `json:"id,omitempty"`
	TaskID string `json:"task_id"`
}

// GetTaskPushNotificationConfigResponse represents the response from getting task push notification config.
type GetTaskPushNotificationConfigResponse struct {
	ID     string                          `json:"id"`
	Result *a2a.TaskPushNotificationConfig `json:"result,omitempty"`
	Error  *RPCError                       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error implements the error interface.
func (e *RPCError) Error() string {
	return e.Message
}

// ensureID ensures the request has an ID, generating one if necessary.
func (r *SendMessageRequest) ensureID() {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
}

// ensureID ensures the request has an ID, generating one if necessary.
func (r *SendStreamingMessageRequest) ensureID() {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
}

// ensureID ensures the request has an ID, generating one if necessary.
func (r *GetTaskRequest) ensureID() {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
}

// ensureID ensures the request has an ID, generating one if necessary.
func (r *CancelTaskRequest) ensureID() {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
}

// ensureID ensures the request has an ID, generating one if necessary.
func (r *SetTaskPushNotificationConfigRequest) ensureID() {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
}

// ensureID ensures the request has an ID, generating one if necessary.
func (r *GetTaskPushNotificationConfigRequest) ensureID() {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
}

// toJSONRPC converts the client request to a JSON-RPC request.
func (r *SendMessageRequest) toJSONRPC() *a2a.SendMessageRequest {
	return &a2a.SendMessageRequest{
		Method: "message/send",
		Params: &a2a.MessageSendParams{
			Message: r.Message,
		},
		ID: r.ID,
	}
}

// toJSONRPC converts the client request to a JSON-RPC request.
func (r *GetTaskRequest) toJSONRPC() *a2a.GetTaskRequest {
	return &a2a.GetTaskRequest{
		Method: "tasks/get",
		Params: &a2a.TaskQueryParams{
			ID: r.TaskID,
		},
		ID: r.ID,
	}
}

// toJSONRPC converts the client request to a JSON-RPC request.
func (r *CancelTaskRequest) toJSONRPC() *a2a.CancelTaskRequest {
	return &a2a.CancelTaskRequest{
		Method: "tasks/cancel",
		Params: a2a.TaskIDParams{
			ID: r.TaskID,
		},
		ID: r.ID,
	}
}

// StreamingEvent represents a streaming event type.
type StreamingEvent interface {
	EventType() string
}

// TaskStatusUpdateEvent represents a task status update event.
type TaskStatusUpdateEvent struct {
	TaskID    string         `json:"task_id"`
	ContextID string         `json:"context_id,omitempty"`
	Status    a2a.TaskStatus `json:"status"`
	Final     bool           `json:"final"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

var _ StreamingEvent = (*TaskStatusUpdateEvent)(nil)

// EventType returns the event type.
func (e *TaskStatusUpdateEvent) EventType() string {
	return "task_status_update"
}

// TaskArtifactUpdateEvent represents a task artifact update event.
type TaskArtifactUpdateEvent struct {
	Artifact *a2a.Artifact `json:"artifact"`
	Append   bool          `json:"append"`
}

var _ StreamingEvent = (*TaskArtifactUpdateEvent)(nil)

// EventType returns the event type.
func (e *TaskArtifactUpdateEvent) EventType() string {
	return "task_artifact_update"
}
