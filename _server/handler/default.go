// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	a2a "github.com/go-a2a/a2a-go"
	server "github.com/go-a2a/a2a-go/_server"
)

// DefaultRequestHandler provides the default implementation of RequestHandler.
// It coordinates between various components to process A2A requests including
// task management, agent execution, and event streaming.
type DefaultRequestHandler struct {
	agentExecutor               AgentExecutor
	taskStore                   TaskStore
	queueManager                QueueManager
	pushNotifier                PushNotifier
	pushNotificationConfigStore PushNotificationConfigStore
	contextBuilder              server.CallContextBuilder

	// Task execution tracking
	runningTasks map[string]context.CancelFunc
	taskMutex    sync.RWMutex
}

var _ RequestHandler = (*DefaultRequestHandler)(nil)

// DefaultRequestHandlerOption defines a function type for configuring DefaultRequestHandler.
type DefaultRequestHandlerOption func(*DefaultRequestHandler)

// WithQueueManager sets the queue manager for event streaming.
func WithQueueManager(qm QueueManager) DefaultRequestHandlerOption {
	return func(h *DefaultRequestHandler) {
		h.queueManager = qm
	}
}

// WithPushNotifier sets the push notifier for notifications.
func WithPushNotifier(pn PushNotifier) DefaultRequestHandlerOption {
	return func(h *DefaultRequestHandler) {
		h.pushNotifier = pn
	}
}

// WithContextBuilder sets the context builder for request contexts.
func WithContextBuilder(cb server.CallContextBuilder) DefaultRequestHandlerOption {
	return func(h *DefaultRequestHandler) {
		h.contextBuilder = cb
	}
}

// WithPushNotificationConfigStore sets the push notification config store.
func WithPushNotificationConfigStore(store PushNotificationConfigStore) DefaultRequestHandlerOption {
	return func(h *DefaultRequestHandler) {
		h.pushNotificationConfigStore = store
	}
}

// NewDefaultRequestHandler creates a new DefaultRequestHandler with the required dependencies.
func NewDefaultRequestHandler(agentExecutor AgentExecutor, taskStore TaskStore, opts ...DefaultRequestHandlerOption) *DefaultRequestHandler {
	if agentExecutor == nil {
		panic("agentExecutor cannot be nil")
	}
	if taskStore == nil {
		panic("taskStore cannot be nil")
	}

	handler := &DefaultRequestHandler{
		agentExecutor:  agentExecutor,
		taskStore:      taskStore,
		contextBuilder: server.NewDefaultCallContextBuilder(),
		runningTasks:   make(map[string]context.CancelFunc),
	}

	for _, opt := range opts {
		opt(handler)
	}

	return handler
}

// OnGetTask handles requests to retrieve task information.
func (h *DefaultRequestHandler) OnGetTask(ctx context.Context, callCtx *server.ServerCallContext, request *GetTaskRequest) (*GetTaskResponse, error) {
	// Validate the request
	if err := request.Validate(); err != nil {
		return nil, NewInvalidRequestError(err.Error())
	}

	// Retrieve the task from storage
	task, err := h.taskStore.Get(ctx, request.TaskID)
	if err != nil {
		log.Printf("Failed to retrieve task %s: %v", request.TaskID, err)
		return nil, NewTaskNotFoundError(request.TaskID)
	}

	// Validate the task before returning
	if err := task.Validate(); err != nil {
		log.Printf("Task %s is invalid: %v", request.TaskID, err)
		return nil, NewInternalError(fmt.Sprintf("task %s is in invalid state", request.TaskID))
	}

	response := &GetTaskResponse{
		Task: task,
	}

	if err := response.Validate(); err != nil {
		log.Printf("Response validation failed: %v", err)
		return nil, NewInternalError("failed to create valid response")
	}

	return response, nil
}

// OnCancelTask handles requests to cancel a running task.
func (h *DefaultRequestHandler) OnCancelTask(ctx context.Context, callCtx *server.ServerCallContext, request *CancelTaskRequest) (*CancelTaskResponse, error) {
	// Validate the request
	if err := request.Validate(); err != nil {
		return nil, NewInvalidRequestError(err.Error())
	}

	// Get the task to check its current state
	task, err := h.taskStore.Get(ctx, request.TaskID)
	if err != nil {
		log.Printf("Failed to retrieve task %s for cancellation: %v", request.TaskID, err)
		return nil, NewTaskNotFoundError(request.TaskID)
	}

	// Check if the task can be canceled
	if task.Status.State == a2a.TaskStateCompleted || task.Status.State == a2a.TaskStateCanceled {
		return &CancelTaskResponse{
			Success:   false,
			Message:   fmt.Sprintf("task is already %s", task.Status.State),
			TaskState: string(task.Status.State),
		}, nil
	}

	// Cancel the running execution if it exists
	h.taskMutex.Lock()
	if cancelFunc, exists := h.runningTasks[request.TaskID]; exists {
		cancelFunc()
		delete(h.runningTasks, request.TaskID)
	}
	h.taskMutex.Unlock()

	// Cancel the task in the agent executor
	if err := h.agentExecutor.Cancel(ctx, request.TaskID); err != nil {
		log.Printf("Failed to cancel task %s in agent executor: %v", request.TaskID, err)
		return nil, NewTaskNotCancelableError(request.TaskID, err.Error())
	}

	// Update the task status to canceled
	task.Status.State = a2a.TaskStateCanceled

	if err := h.taskStore.Update(ctx, task); err != nil {
		log.Printf("Failed to update task %s status to canceled: %v", request.TaskID, err)
		return nil, NewStorageError("update", request.TaskID, err.Error())
	}

	return &CancelTaskResponse{
		Success:   true,
		Message:   "task canceled successfully",
		TaskState: string(a2a.TaskStateCanceled),
	}, nil
}

// OnMessageSend handles requests to send a message and process it.
func (h *DefaultRequestHandler) OnMessageSend(ctx context.Context, callCtx *server.ServerCallContext, request *MessageSendRequest) (*MessageSendResponse, error) {
	// Validate the request
	if err := request.Validate(); err != nil {
		return nil, NewInvalidRequestError(err.Error())
	}

	// Generate a task ID if not provided
	taskID := request.TaskID
	if taskID == "" {
		taskID = uuid.New().String()
	}

	// Create or update the task
	task, err := h.getOrCreateTask(ctx, taskID, request.Messages)
	if err != nil {
		return nil, err
	}

	// Create a cancellable context for the execution
	execCtx, cancel := context.WithCancel(ctx)

	// Store the cancel function for potential cancellation
	h.taskMutex.Lock()
	h.runningTasks[taskID] = cancel
	h.taskMutex.Unlock()

	// Clean up the cancel function when done
	defer func() {
		h.taskMutex.Lock()
		delete(h.runningTasks, taskID)
		h.taskMutex.Unlock()
	}()

	// Update task status to running
	task.Status.State = a2a.TaskStateRunning
	if err := h.taskStore.Update(ctx, task); err != nil {
		log.Printf("Failed to update task %s status to running: %v", taskID, err)
		return nil, NewStorageError("update", taskID, err.Error())
	}

	// Execute the agent
	result, err := h.agentExecutor.Execute(execCtx, request.Messages)
	if err != nil {
		// Update task status to failed
		task.Status.State = a2a.TaskStateFailed
		if updateErr := h.taskStore.Update(ctx, task); updateErr != nil {
			log.Printf("Failed to update task %s status to failed: %v", taskID, updateErr)
		}
		return nil, NewExecutionError(taskID, err.Error())
	}

	// Update task with results
	task.Status.State = a2a.TaskStateCompleted
	for _, msg := range result.Messages {
		task.History = append(task.History, msg)
	}

	if err := h.taskStore.Update(ctx, task); err != nil {
		log.Printf("Failed to update task %s with results: %v", taskID, err)
		return nil, NewStorageError("update", taskID, err.Error())
	}

	// Send push notification if configured
	if h.pushNotifier != nil {
		for _, msg := range result.Messages {
			notification := &PushNotification{
				TaskID:  taskID,
				Message: msg,
				Target:  callCtx.User().UserName(),
			}
			if err := h.pushNotifier.SendNotification(ctx, notification); err != nil {
				log.Printf("Failed to send push notification for task %s: %v", taskID, err)
			}
		}
	}

	response := &MessageSendResponse{
		Messages: result.Messages,
		TaskID:   taskID,
	}

	if err := response.Validate(); err != nil {
		log.Printf("Response validation failed: %v", err)
		return nil, NewInternalError("failed to create valid response")
	}

	return response, nil
}

// OnMessageSendStream handles requests to send a message with streaming response.
func (h *DefaultRequestHandler) OnMessageSendStream(ctx context.Context, callCtx *server.ServerCallContext, request *MessageSendStreamRequest) (*MessageSendStreamResponse, error) {
	// Validate the request
	if err := request.Validate(); err != nil {
		return nil, NewInvalidRequestError(err.Error())
	}

	// Generate a task ID if not provided
	taskID := request.TaskID
	if taskID == "" {
		taskID = uuid.New().String()
	}

	// Create or update the task
	task, err := h.getOrCreateTask(ctx, taskID, request.Messages)
	if err != nil {
		return nil, err
	}

	// Create a queue for streaming if available
	if h.queueManager != nil {
		if err := h.queueManager.CreateQueue(ctx, taskID); err != nil {
			log.Printf("Failed to create queue for task %s: %v", taskID, err)
		}
	}

	// Create a cancellable context for the execution
	execCtx, cancel := context.WithCancel(ctx)

	// Store the cancel function for potential cancellation
	h.taskMutex.Lock()
	h.runningTasks[taskID] = cancel
	h.taskMutex.Unlock()

	// Clean up when done
	defer func() {
		h.taskMutex.Lock()
		delete(h.runningTasks, taskID)
		h.taskMutex.Unlock()

		if h.queueManager != nil {
			h.queueManager.CloseQueue(ctx, taskID)
		}
	}()

	// Update task status to running
	task.Status.State = a2a.TaskStateRunning
	if err := h.taskStore.Update(ctx, task); err != nil {
		log.Printf("Failed to update task %s status to running: %v", taskID, err)
		return nil, NewStorageError("update", taskID, err.Error())
	}

	// Execute the agent with streaming
	stream, err := h.agentExecutor.ExecuteStream(execCtx, request.Messages)
	if err != nil {
		// Update task status to failed
		task.Status.State = a2a.TaskStateFailed
		if updateErr := h.taskStore.Update(ctx, task); updateErr != nil {
			log.Printf("Failed to update task %s status to failed: %v", taskID, updateErr)
		}
		return nil, NewExecutionError(taskID, err.Error())
	}

	// Start a goroutine to process the stream and update the task
	go h.processStream(ctx, taskID, stream, task)

	// Get the initial messages from the stream
	var initialMessages []*a2a.Message
	select {
	case msg := <-stream:
		if msg != nil {
			initialMessages = []*a2a.Message{msg}
		}
	case <-time.After(100 * time.Millisecond):
		// No immediate message, continue with empty initial messages
	}

	response := &MessageSendStreamResponse{
		Messages: initialMessages,
		TaskID:   taskID,
		Stream:   stream,
	}

	if err := response.Validate(); err != nil {
		log.Printf("Response validation failed: %v", err)
		return nil, NewInternalError("failed to create valid response")
	}

	return response, nil
}

// processStream processes the streaming messages and updates the task.
func (h *DefaultRequestHandler) processStream(ctx context.Context, taskID string, stream <-chan *a2a.Message, task *a2a.Task) {
	var allMessages []*a2a.Message

	for msg := range stream {
		if msg == nil {
			continue
		}

		allMessages = append(allMessages, msg)
		task.History = append(task.History, msg)

		// Publish to queue if available
		if h.queueManager != nil {
			if err := h.queueManager.PublishEvent(ctx, taskID, msg); err != nil {
				log.Printf("Failed to publish event for task %s: %v", taskID, err)
			}
		}

		// Send push notification if configured
		if h.pushNotifier != nil {
			notification := &PushNotification{
				TaskID:  taskID,
				Message: msg,
				Target:  "stream",
			}
			if err := h.pushNotifier.SendNotification(ctx, notification); err != nil {
				log.Printf("Failed to send push notification for task %s: %v", taskID, err)
			}
		}
	}

	// Update final task status
	task.Status.State = a2a.TaskStateCompleted
	if err := h.taskStore.Update(ctx, task); err != nil {
		log.Printf("Failed to update task %s with final status: %v", taskID, err)
	}
}

// getOrCreateTask retrieves an existing task or creates a new one.
func (h *DefaultRequestHandler) getOrCreateTask(ctx context.Context, taskID string, messages []*a2a.Message) (*a2a.Task, error) {
	// Try to get existing task
	task, err := h.taskStore.Get(ctx, taskID)
	if err == nil {
		// Task exists, append new messages to history
		for _, msg := range messages {
			task.History = append(task.History, msg)
		}
		return task, nil
	}

	// Create new task
	task = &a2a.Task{
		ID:        taskID,
		ContextID: uuid.New().String(),
		Status: a2a.TaskStatus{
			State: a2a.TaskStateSubmitted,
		},
		History:   messages,
		Artifacts: []*a2a.Artifact{},
	}

	if err := h.taskStore.Create(ctx, task); err != nil {
		return nil, NewStorageError("create", taskID, err.Error())
	}

	return task, nil
}

// OnSetTaskPushNotificationConfig handles requests to set push notification configuration for a task.
func (h *DefaultRequestHandler) OnSetTaskPushNotificationConfig(ctx context.Context, callCtx *server.ServerCallContext, request *SetTaskPushNotificationConfigRequest) (*SetTaskPushNotificationConfigResponse, error) {
	// Validate the request
	if err := request.Validate(); err != nil {
		return nil, NewInvalidRequestError(err.Error())
	}

	// Check if push notification config store is available
	if h.pushNotificationConfigStore == nil {
		return nil, NewUnsupportedOperationError("push notification config store not configured")
	}

	// Check if task exists
	_, err := h.taskStore.Get(ctx, request.TaskID)
	if err != nil {
		log.Printf("Failed to retrieve task %s: %v", request.TaskID, err)
		return nil, NewTaskNotFoundError(request.TaskID)
	}

	// Generate config ID if not provided
	configID := request.PushNotificationConfigID
	if configID == "" {
		configID = uuid.New().String()
	}

	// Set the config ID
	config := request.PushNotificationConfig
	config.ID = configID

	// Store the configuration
	if err := h.pushNotificationConfigStore.SetInfo(ctx, request.TaskID, config); err != nil {
		log.Printf("Failed to store push notification config for task %s: %v", request.TaskID, err)
		return nil, NewStorageError("store", request.TaskID, err.Error())
	}

	return &SetTaskPushNotificationConfigResponse{
		TaskID:                   request.TaskID,
		PushNotificationConfigID: configID,
		PushNotificationConfig:   config,
	}, nil
}

// OnGetTaskPushNotificationConfig handles requests to get push notification configuration for a task.
func (h *DefaultRequestHandler) OnGetTaskPushNotificationConfig(ctx context.Context, callCtx *server.ServerCallContext, request *GetTaskPushNotificationConfigRequest) (*GetTaskPushNotificationConfigResponse, error) {
	// Validate the request
	if err := request.Validate(); err != nil {
		return nil, NewInvalidRequestError(err.Error())
	}

	// Check if push notification config store is available
	if h.pushNotificationConfigStore == nil {
		return nil, NewUnsupportedOperationError("push notification config store not configured")
	}

	// Check if task exists
	_, err := h.taskStore.Get(ctx, request.TaskID)
	if err != nil {
		log.Printf("Failed to retrieve task %s: %v", request.TaskID, err)
		return nil, NewTaskNotFoundError(request.TaskID)
	}

	// Get the configurations
	configs, err := h.pushNotificationConfigStore.GetInfo(ctx, request.TaskID)
	if err != nil {
		log.Printf("Failed to get push notification configs for task %s: %v", request.TaskID, err)
		return nil, NewStorageError("get", request.TaskID, err.Error())
	}

	if len(configs) == 0 {
		return nil, NewTaskNotFoundError(request.TaskID)
	}

	// If specific config ID is requested, find it
	if request.PushNotificationConfigID != "" {
		for _, config := range configs {
			if config.ID == request.PushNotificationConfigID {
				return &GetTaskPushNotificationConfigResponse{
					TaskID:                   request.TaskID,
					PushNotificationConfigID: config.ID,
					PushNotificationConfig:   config,
				}, nil
			}
		}
		return nil, NewTaskNotFoundError(request.TaskID)
	}

	// Return the first config if no specific ID is requested
	config := configs[0]
	return &GetTaskPushNotificationConfigResponse{
		TaskID:                   request.TaskID,
		PushNotificationConfigID: config.ID,
		PushNotificationConfig:   config,
	}, nil
}

// OnListTaskPushNotificationConfig handles requests to list push notification configurations for a task.
func (h *DefaultRequestHandler) OnListTaskPushNotificationConfig(ctx context.Context, callCtx *server.ServerCallContext, request *ListTaskPushNotificationConfigRequest) (*ListTaskPushNotificationConfigResponse, error) {
	// Validate the request
	if err := request.Validate(); err != nil {
		return nil, NewInvalidRequestError(err.Error())
	}

	// Check if push notification config store is available
	if h.pushNotificationConfigStore == nil {
		return nil, NewUnsupportedOperationError("push notification config store not configured")
	}

	// Check if task exists
	_, err := h.taskStore.Get(ctx, request.ID)
	if err != nil {
		log.Printf("Failed to retrieve task %s: %v", request.ID, err)
		return nil, NewTaskNotFoundError(request.ID)
	}

	// Get the configurations
	configs, err := h.pushNotificationConfigStore.GetInfo(ctx, request.ID)
	if err != nil {
		log.Printf("Failed to get push notification configs for task %s: %v", request.ID, err)
		return nil, NewStorageError("get", request.ID, err.Error())
	}

	return &ListTaskPushNotificationConfigResponse{
		TaskID:                  request.ID,
		PushNotificationConfigs: configs,
	}, nil
}

// OnDeleteTaskPushNotificationConfig handles requests to delete a push notification configuration.
func (h *DefaultRequestHandler) OnDeleteTaskPushNotificationConfig(ctx context.Context, callCtx *server.ServerCallContext, request *DeleteTaskPushNotificationConfigRequest) (*DeleteTaskPushNotificationConfigResponse, error) {
	// Validate the request
	if err := request.Validate(); err != nil {
		return nil, NewInvalidRequestError(err.Error())
	}

	// Check if push notification config store is available
	if h.pushNotificationConfigStore == nil {
		return nil, NewUnsupportedOperationError("push notification config store not configured")
	}

	// Check if task exists
	_, err := h.taskStore.Get(ctx, request.TaskID)
	if err != nil {
		log.Printf("Failed to retrieve task %s: %v", request.TaskID, err)
		return nil, NewTaskNotFoundError(request.TaskID)
	}

	// Delete the configuration
	if err := h.pushNotificationConfigStore.DeleteInfo(ctx, request.TaskID, request.PushNotificationConfigID); err != nil {
		log.Printf("Failed to delete push notification config for task %s: %v", request.TaskID, err)
		return nil, NewStorageError("delete", request.TaskID, err.Error())
	}

	return &DeleteTaskPushNotificationConfigResponse{
		Success: true,
		Message: "push notification configuration deleted successfully",
	}, nil
}

// OnResubscribeToTask handles requests to re-subscribe to a task's event stream.
func (h *DefaultRequestHandler) OnResubscribeToTask(ctx context.Context, callCtx *server.ServerCallContext, request *ResubscribeToTaskRequest) (*ResubscribeToTaskResponse, error) {
	// Validate the request
	if err := request.Validate(); err != nil {
		return nil, NewInvalidRequestError(err.Error())
	}

	// Check if task exists
	task, err := h.taskStore.Get(ctx, request.TaskID)
	if err != nil {
		log.Printf("Failed to retrieve task %s: %v", request.TaskID, err)
		return nil, NewTaskNotFoundError(request.TaskID)
	}

	// Check if the task is in a terminal state
	if task.Status.State == a2a.TaskStateCompleted || task.Status.State == a2a.TaskStateCanceled || task.Status.State == a2a.TaskStateFailed {
		return nil, NewInvalidRequestError(fmt.Sprintf("task %s is in terminal state: %s", request.TaskID, task.Status.State))
	}

	// Get the event queue if available
	if h.queueManager == nil {
		return nil, NewUnsupportedOperationError("queue manager not configured")
	}

	queue, err := h.queueManager.GetQueue(ctx, request.TaskID)
	if err != nil {
		log.Printf("Failed to get queue for task %s: %v", request.TaskID, err)
		return nil, NewTaskNotFoundError(request.TaskID)
	}

	// Return the response with the stream
	return &ResubscribeToTaskResponse{
		TaskID:   request.TaskID,
		Messages: []*a2a.Message{}, // No initial messages for resubscription
		Stream:   queue,
	}, nil
}

// Validate ensures the DefaultRequestHandler is in a valid state.
func (h *DefaultRequestHandler) Validate() error {
	if h.agentExecutor == nil {
		return fmt.Errorf("agent executor cannot be nil")
	}
	if h.taskStore == nil {
		return fmt.Errorf("task store cannot be nil")
	}
	if h.contextBuilder == nil {
		return fmt.Errorf("context builder cannot be nil")
	}
	if h.runningTasks == nil {
		return fmt.Errorf("running tasks map cannot be nil")
	}
	return nil
}
