// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package handler provides request handlers for the A2A protocol server.
// This package implements the core request handling logic, including
// task management, message processing, and protocol-specific adapters.
package handler

import (
	"context"

	a2a "github.com/go-a2a/a2a-go"
	server "github.com/go-a2a/a2a-go/_server"
)

// RequestHandler defines the interface for handling A2A protocol requests.
// This interface abstracts the core request processing logic from protocol-specific
// concerns, allowing the same handler to work with different transport protocols
// (JSON-RPC, gRPC, etc.).
type RequestHandler interface {
	// OnGetTask handles requests to retrieve task information.
	// It returns the current state of the specified task or an error if the task
	// cannot be found or accessed.
	OnGetTask(ctx context.Context, callCtx *server.ServerCallContext, request *GetTaskRequest) (*GetTaskResponse, error)

	// OnCancelTask handles requests to cancel a running task.
	// It attempts to cancel the specified task and returns the cancellation result.
	OnCancelTask(ctx context.Context, callCtx *server.ServerCallContext, request *CancelTaskRequest) (*CancelTaskResponse, error)

	// OnMessageSend handles requests to send a message and process it.
	// This is the primary method for task execution and message processing.
	OnMessageSend(ctx context.Context, params *a2a.MessageSendParams, callCtx *server.ServerCallContext) (*MessageSendResponse, error)

	// OnMessageSendStream handles requests to send a message with streaming response.
	// This method supports real-time streaming of task execution results.
	OnMessageSendStream(ctx context.Context, callCtx *server.ServerCallContext, request *MessageSendStreamRequest) (*MessageSendStreamResponse, error)

	// OnSetTaskPushNotificationConfig handles requests to set push notification configuration for a task.
	// This method allows clients to configure push notifications for specific tasks.
	OnSetTaskPushNotificationConfig(ctx context.Context, callCtx *server.ServerCallContext, request *SetTaskPushNotificationConfigRequest) (*SetTaskPushNotificationConfigResponse, error)

	// OnGetTaskPushNotificationConfig handles requests to get push notification configuration for a task.
	// This method retrieves the current push notification configuration for a specific task.
	OnGetTaskPushNotificationConfig(ctx context.Context, callCtx *server.ServerCallContext, request *GetTaskPushNotificationConfigRequest) (*GetTaskPushNotificationConfigResponse, error)

	// OnListTaskPushNotificationConfig handles requests to list push notification configurations for a task.
	// This method retrieves all push notification configurations for a specific task.
	OnListTaskPushNotificationConfig(ctx context.Context, callCtx *server.ServerCallContext, request *ListTaskPushNotificationConfigRequest) (*ListTaskPushNotificationConfigResponse, error)

	// OnDeleteTaskPushNotificationConfig handles requests to delete a push notification configuration.
	// This method removes a specific push notification configuration for a task.
	OnDeleteTaskPushNotificationConfig(ctx context.Context, callCtx *server.ServerCallContext, request *DeleteTaskPushNotificationConfigRequest) (*DeleteTaskPushNotificationConfigResponse, error)

	// OnResubscribeToTask handles requests to re-subscribe to a task's event stream.
	// This method allows clients to re-attach to a running streaming task's events.
	OnResubscribeToTask(ctx context.Context, callCtx *server.ServerCallContext, request *ResubscribeToTaskRequest) (*ResubscribeToTaskResponse, error)
}

// GetTaskRequest represents a request to retrieve task information.
type GetTaskRequest struct {
	TaskID        string `json:"task_id"`
	HistoryLength *int   `json:"history_length,omitempty"`
}

// Validate ensures the GetTaskRequest is valid.
func (r *GetTaskRequest) Validate() error {
	if r.TaskID == "" {
		return NewValidationError("task_id", "task ID cannot be empty")
	}
	return nil
}

// GetTaskResponse represents the response containing task information.
type GetTaskResponse struct {
	Task *a2a.Task `json:"task"`
}

// Validate ensures the GetTaskResponse is valid.
func (r *GetTaskResponse) Validate() error {
	if r.Task == nil {
		return NewValidationError("task", "task cannot be nil")
	}
	return r.Task.Validate()
}

// CancelTaskRequest represents a request to cancel a task.
type CancelTaskRequest struct {
	TaskID string `json:"task_id"`
}

// Validate ensures the CancelTaskRequest is valid.
func (r *CancelTaskRequest) Validate() error {
	if r.TaskID == "" {
		return NewValidationError("task_id", "task ID cannot be empty")
	}
	return nil
}

// CancelTaskResponse represents the response to a task cancellation request.
type CancelTaskResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	TaskState string `json:"task_state,omitempty"`
}

// Validate ensures the CancelTaskResponse is valid.
func (r *CancelTaskResponse) Validate() error {
	// Basic validation - all fields are optional or have default values
	return nil
}

// MessageSendRequest represents a request to send a message for processing.
type MessageSendRequest struct {
	Messages []*a2a.Message `json:"messages"`
	TaskID   string         `json:"task_id,omitempty"`
}

// Validate ensures the MessageSendRequest is valid.
func (r *MessageSendRequest) Validate() error {
	if len(r.Messages) == 0 {
		return NewValidationError("messages", "messages cannot be empty")
	}
	for i, msg := range r.Messages {
		if msg == nil {
			return NewValidationError("messages", "message at index %d cannot be nil", i)
		}
		if err := msg.Validate(); err != nil {
			return NewValidationError("messages", "message at index %d is invalid: %v", i, err)
		}
	}
	return nil
}

// MessageSendResponse represents the response to a message send request.
type MessageSendResponse struct {
	Messages []*a2a.Message `json:"messages"`
	TaskID   string         `json:"task_id"`
}

// Validate ensures the MessageSendResponse is valid.
func (r *MessageSendResponse) Validate() error {
	if r.TaskID == "" {
		return NewValidationError("task_id", "task ID cannot be empty")
	}
	if len(r.Messages) == 0 {
		return NewValidationError("messages", "messages cannot be empty")
	}
	for i, msg := range r.Messages {
		if msg == nil {
			return NewValidationError("messages", "message at index %d cannot be nil", i)
		}
		if err := msg.Validate(); err != nil {
			return NewValidationError("messages", "message at index %d is invalid: %v", i, err)
		}
	}
	return nil
}

// MessageSendStreamRequest represents a request to send a message with streaming response.
type MessageSendStreamRequest struct {
	Messages []*a2a.Message `json:"messages"`
	TaskID   string         `json:"task_id,omitempty"`
}

// Validate ensures the MessageSendStreamRequest is valid.
func (r *MessageSendStreamRequest) Validate() error {
	if len(r.Messages) == 0 {
		return NewValidationError("messages", "messages cannot be empty")
	}
	for i, msg := range r.Messages {
		if msg == nil {
			return NewValidationError("messages", "message at index %d cannot be nil", i)
		}
		if err := msg.Validate(); err != nil {
			return NewValidationError("messages", "message at index %d is invalid: %v", i, err)
		}
	}
	return nil
}

// MessageSendStreamResponse represents the response to a streaming message send request.
type MessageSendStreamResponse struct {
	Messages []*a2a.Message      `json:"messages"`
	TaskID   string              `json:"task_id"`
	Stream   <-chan *a2a.Message `json:"-"` // Channel for streaming messages
}

// Validate ensures the MessageSendStreamResponse is valid.
func (r *MessageSendStreamResponse) Validate() error {
	if r.TaskID == "" {
		return NewValidationError("task_id", "task ID cannot be empty")
	}
	if len(r.Messages) == 0 {
		return NewValidationError("messages", "messages cannot be empty")
	}
	for i, msg := range r.Messages {
		if msg == nil {
			return NewValidationError("messages", "message at index %d cannot be nil", i)
		}
		if err := msg.Validate(); err != nil {
			return NewValidationError("messages", "message at index %d is invalid: %v", i, err)
		}
	}
	return nil
}

// SetTaskPushNotificationConfigRequest represents a request to set push notification configuration.
type SetTaskPushNotificationConfigRequest struct {
	TaskID                   string                      `json:"task_id"`
	PushNotificationConfigID string                      `json:"push_notification_config_id,omitempty"`
	PushNotificationConfig   *a2a.PushNotificationConfig `json:"push_notification_config"`
}

// Validate ensures the SetTaskPushNotificationConfigRequest is valid.
func (r *SetTaskPushNotificationConfigRequest) Validate() error {
	if r.TaskID == "" {
		return NewValidationError("task_id", "task ID cannot be empty")
	}
	if r.PushNotificationConfig == nil {
		return NewValidationError("push_notification_config", "push notification config cannot be nil")
	}
	return r.PushNotificationConfig.Validate()
}

// SetTaskPushNotificationConfigResponse represents the response to setting push notification configuration.
type SetTaskPushNotificationConfigResponse struct {
	TaskID                   string                      `json:"task_id"`
	PushNotificationConfigID string                      `json:"push_notification_config_id"`
	PushNotificationConfig   *a2a.PushNotificationConfig `json:"push_notification_config"`
}

// Validate ensures the SetTaskPushNotificationConfigResponse is valid.
func (r *SetTaskPushNotificationConfigResponse) Validate() error {
	if r.TaskID == "" {
		return NewValidationError("task_id", "task ID cannot be empty")
	}
	if r.PushNotificationConfigID == "" {
		return NewValidationError("push_notification_config_id", "push notification config ID cannot be empty")
	}
	if r.PushNotificationConfig == nil {
		return NewValidationError("push_notification_config", "push notification config cannot be nil")
	}
	return r.PushNotificationConfig.Validate()
}

// GetTaskPushNotificationConfigRequest represents a request to get push notification configuration.
type GetTaskPushNotificationConfigRequest struct {
	TaskID                   string `json:"task_id"`
	PushNotificationConfigID string `json:"push_notification_config_id,omitempty"`
}

// Validate ensures the GetTaskPushNotificationConfigRequest is valid.
func (r *GetTaskPushNotificationConfigRequest) Validate() error {
	if r.TaskID == "" {
		return NewValidationError("task_id", "task ID cannot be empty")
	}
	return nil
}

// GetTaskPushNotificationConfigResponse represents the response to getting push notification configuration.
type GetTaskPushNotificationConfigResponse struct {
	TaskID                   string                      `json:"task_id"`
	PushNotificationConfigID string                      `json:"push_notification_config_id"`
	PushNotificationConfig   *a2a.PushNotificationConfig `json:"push_notification_config"`
}

// Validate ensures the GetTaskPushNotificationConfigResponse is valid.
func (r *GetTaskPushNotificationConfigResponse) Validate() error {
	if r.TaskID == "" {
		return NewValidationError("task_id", "task ID cannot be empty")
	}
	if r.PushNotificationConfigID == "" {
		return NewValidationError("push_notification_config_id", "push notification config ID cannot be empty")
	}
	if r.PushNotificationConfig == nil {
		return NewValidationError("push_notification_config", "push notification config cannot be nil")
	}
	return r.PushNotificationConfig.Validate()
}

// ListTaskPushNotificationConfigResponse represents the response to listing push notification configurations.
type ListTaskPushNotificationConfigResponse struct {
	TaskID                  string                        `json:"task_id"`
	PushNotificationConfigs []*a2a.PushNotificationConfig `json:"push_notification_configs"`
}

// Validate ensures the ListTaskPushNotificationConfigResponse is valid.
func (r *ListTaskPushNotificationConfigResponse) Validate() error {
	if r.TaskID == "" {
		return NewValidationError("task_id", "task ID cannot be empty")
	}
	for i, config := range r.PushNotificationConfigs {
		if config == nil {
			return NewValidationError("push_notification_configs", "push notification config at index %d cannot be nil", i)
		}
		if err := config.Validate(); err != nil {
			return NewValidationError("push_notification_configs", "push notification config at index %d is invalid: %v", i, err)
		}
	}
	return nil
}

// DeleteTaskPushNotificationConfigRequest represents a request to delete push notification configuration.
type DeleteTaskPushNotificationConfigRequest struct {
	TaskID                   string `json:"task_id"`
	PushNotificationConfigID string `json:"push_notification_config_id"`
}

// Validate ensures the DeleteTaskPushNotificationConfigRequest is valid.
func (r *DeleteTaskPushNotificationConfigRequest) Validate() error {
	if r.TaskID == "" {
		return NewValidationError("task_id", "task ID cannot be empty")
	}
	if r.PushNotificationConfigID == "" {
		return NewValidationError("push_notification_config_id", "push notification config ID cannot be empty")
	}
	return nil
}

// ResubscribeToTaskRequest represents a request to re-subscribe to a task's event stream.
type ResubscribeToTaskRequest struct {
	TaskID string `json:"task_id"`
}

// Validate ensures the ResubscribeToTaskRequest is valid.
func (r *ResubscribeToTaskRequest) Validate() error {
	if r.TaskID == "" {
		return NewValidationError("task_id", "task ID cannot be empty")
	}
	return nil
}

// ResubscribeToTaskResponse represents the response to re-subscribing to a task.
type ResubscribeToTaskResponse struct {
	TaskID   string              `json:"task_id"`
	Messages []*a2a.Message      `json:"messages"`
	Stream   <-chan *a2a.Message `json:"-"` // Channel for streaming events
}

// Validate ensures the ResubscribeToTaskResponse is valid.
func (r *ResubscribeToTaskResponse) Validate() error {
	if r.TaskID == "" {
		return NewValidationError("task_id", "task ID cannot be empty")
	}
	for i, msg := range r.Messages {
		if msg == nil {
			return NewValidationError("messages", "message at index %d cannot be nil", i)
		}
		if err := msg.Validate(); err != nil {
			return NewValidationError("messages", "message at index %d is invalid: %v", i, err)
		}
	}
	return nil
}

// AgentExecutor defines the interface for executing agent business logic.
type AgentExecutor interface {
	// Execute runs the agent with the given messages and returns the result.
	Execute(ctx context.Context, messages []*a2a.Message) (*ExecutionResult, error)

	// ExecuteStream runs the agent with streaming output.
	ExecuteStream(ctx context.Context, messages []*a2a.Message) (<-chan *a2a.Message, error)

	// Cancel cancels a running execution by task ID.
	Cancel(ctx context.Context, taskID string) error
}

// ExecutionResult represents the result of agent execution.
type ExecutionResult struct {
	Messages []*a2a.Message `json:"messages"`
	TaskID   string         `json:"task_id"`
}

// Validate ensures the ExecutionResult is valid.
func (r *ExecutionResult) Validate() error {
	if r.TaskID == "" {
		return NewValidationError("task_id", "task ID cannot be empty")
	}
	if len(r.Messages) == 0 {
		return NewValidationError("messages", "messages cannot be empty")
	}
	for i, msg := range r.Messages {
		if msg == nil {
			return NewValidationError("messages", "message at index %d cannot be nil", i)
		}
		if err := msg.Validate(); err != nil {
			return NewValidationError("messages", "message at index %d is invalid: %v", i, err)
		}
	}
	return nil
}

// TaskStore defines the interface for task persistence.
type TaskStore interface {
	// Get retrieves a task by its ID.
	Get(ctx context.Context, taskID string) (*a2a.Task, error)

	// Create creates a new task.
	Create(ctx context.Context, task *a2a.Task) error

	// Update updates an existing task.
	Update(ctx context.Context, task *a2a.Task) error

	// Delete removes a task by its ID.
	Delete(ctx context.Context, taskID string) error
}

// QueueManager defines the interface for managing event queues.
type QueueManager interface {
	// CreateQueue creates a new event queue for a task.
	CreateQueue(ctx context.Context, taskID string) error

	// GetQueue retrieves the event queue for a task.
	GetQueue(ctx context.Context, taskID string) (<-chan *a2a.Message, error)

	// PublishEvent publishes an event to the task's queue.
	PublishEvent(ctx context.Context, taskID string, message *a2a.Message) error

	// CloseQueue closes the event queue for a task.
	CloseQueue(ctx context.Context, taskID string) error
}

// PushNotifier defines the interface for push notifications.
type PushNotifier interface {
	// SendNotification sends a push notification.
	SendNotification(ctx context.Context, notification *PushNotification) error

	// ConfigureNotification configures push notification settings.
	ConfigureNotification(ctx context.Context, config *a2a.PushNotificationConfig) error
}

// PushNotificationConfigStore defines the interface for push notification configuration storage.
type PushNotificationConfigStore interface {
	// SetInfo sets push notification configuration for a task.
	SetInfo(ctx context.Context, taskID string, config *a2a.PushNotificationConfig) error

	// GetInfo retrieves push notification configuration for a task.
	GetInfo(ctx context.Context, taskID string) ([]*a2a.PushNotificationConfig, error)

	// DeleteInfo deletes a specific push notification configuration.
	DeleteInfo(ctx context.Context, taskID, configID string) error
}

// PushNotification represents a push notification.
type PushNotification struct {
	TaskID  string       `json:"task_id"`
	Message *a2a.Message `json:"message"`
	Target  string       `json:"target"`
}

// Validate ensures the PushNotification is valid.
func (p *PushNotification) Validate() error {
	if p.TaskID == "" {
		return NewValidationError("task_id", "task ID cannot be empty")
	}
	if p.Message == nil {
		return NewValidationError("message", "message cannot be nil")
	}
	if err := p.Message.Validate(); err != nil {
		return NewValidationError("message", "message is invalid: %v", err)
	}
	if p.Target == "" {
		return NewValidationError("target", "target cannot be empty")
	}
	return nil
}
