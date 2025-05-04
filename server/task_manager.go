// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/internal/jsonrpc2"
)

// TaskManager is the interface that task managers must implement.
type TaskManager interface {
	// OnSendTask handles a new task.
	OnSendTask(ctx context.Context, req *a2a.SendTaskRequest) (*a2a.SendTaskResponse, error)

	// OnGetTask retrieves a task.
	OnGetTask(ctx context.Context, req *a2a.GetTaskRequest) (*a2a.GetTaskResponse, error)

	// OnCancelTask cancels a task.
	OnCancelTask(ctx context.Context, req *a2a.CancelTaskRequest) (*a2a.CancelTaskResponse, error)

	// OnSendTaskSubscribe starts a streaming task and returns a channel for updates.
	OnSendTaskSubscribe(ctx context.Context, req *a2a.SendTaskStreamingRequest) (<-chan *a2a.SendTaskStreamingResponse, error)

	// OnSetTaskPushNotification configures push notification for a task.
	OnSetTaskPushNotification(ctx context.Context, req *a2a.SetTaskPushNotificationRequest) (*a2a.SetTaskPushNotificationResponse, error)

	// OnGetTaskPushNotification retrieves push notification configuration for a task.
	OnGetTaskPushNotification(ctx context.Context, req *a2a.GetTaskPushNotificationRequest) (*a2a.GetTaskPushNotificationResponse, error)

	// OnResubscribeToTask resubscribes to a task's updates.
	//
	// Return the `<-chan a2a.TaskEvent`, or [a2a.JSONRPCResponse].
	OnResubscribeToTask(ctx context.Context, req *a2a.TaskResubscriptionRequest) (<-chan a2a.TaskEvent, error)
}

// InMemoryTaskManager is an in-memory implementation of TaskManager.
type InMemoryTaskManager struct {
	// Tasks is a map of task ID to task.
	tasks map[string]*a2a.Task

	// TaskMutex protects the tasks map.
	taskMu sync.RWMutex

	// PushNotifications is a map of task ID to push notification config.
	pushNotifications map[string]a2a.TaskPushNotificationConfig

	// PushMutex protects the pushNotifications map.
	pushMu sync.RWMutex

	// Subscribers is a map of task ID to a list of subscriber channels.
	subscribers map[string][]chan a2a.TaskEvent

	// SubMutex protects the subscribers map.
	subMu sync.RWMutex

	// logger is the logger for the task manager.
	logger *slog.Logger

	// tracer is the tracer for the task manager.
	tracer trace.Tracer
}

var _ TaskManager = (*InMemoryTaskManager)(nil)

// NewInMemoryTaskManager creates a new InMemoryTaskManager.
func NewInMemoryTaskManager() *InMemoryTaskManager {
	return &InMemoryTaskManager{
		tasks:             make(map[string]*a2a.Task),
		pushNotifications: make(map[string]a2a.TaskPushNotificationConfig),
		subscribers:       make(map[string][]chan a2a.TaskEvent),
		logger:            slog.Default(),
		tracer:            otel.GetTracerProvider().Tracer("github.com/go-a2a/a2a/server.task_manager"),
	}
}

// WithLogger sets the logger for the TaskManager.
func (tm *InMemoryTaskManager) WithLogger(logger *slog.Logger) *InMemoryTaskManager {
	tm.logger = logger
	return tm
}

// WithTracer sets the tracer for the TaskManager.
func (tm *InMemoryTaskManager) WithTracer(tracer trace.Tracer) *InMemoryTaskManager {
	tm.tracer = tracer
	return tm
}

// OnSendTask handles a new task.
func (tm *InMemoryTaskManager) OnSendTask(ctx context.Context, req *a2a.SendTaskRequest) (*a2a.SendTaskResponse, error) {
	// no-op
	return &a2a.SendTaskResponse{}, nil
}

// OnGetTask retrieves a task.
func (tm *InMemoryTaskManager) OnGetTask(ctx context.Context, req *a2a.GetTaskRequest) (*a2a.GetTaskResponse, error) {
	ctx, span := tm.tracer.Start(ctx, "task_manager.OnGetTask",
		trace.WithAttributes(attribute.String("a2a.task_id", req.Params.ID)))
	defer span.End()

	taskID := req.Params.ID
	if taskID == "" {
		return nil, errors.New("task ID cannot be empty")
	}

	tm.taskMu.RLock()
	task, ok := tm.tasks[taskID]
	tm.taskMu.RUnlock()

	if !ok {
		tm.logger.InfoContext(ctx, "task not found", slog.String("task_id", taskID))
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	tm.logger.InfoContext(ctx, "task retrieved", slog.String("task_id", taskID), slog.String("state", string(task.Status.State)))

	resp, err := jsonrpc2.NewResponse(req.ID, task, nil)
	if err != nil {
		return nil, err
	}
	return &a2a.GetTaskResponse{
		Response: resp,
		Result:   task,
	}, nil
}

// OnCancelTask cancels a task.
func (tm *InMemoryTaskManager) OnCancelTask(ctx context.Context, req *a2a.CancelTaskRequest) (*a2a.CancelTaskResponse, error) {
	ctx, span := tm.tracer.Start(ctx, "task_manager.OnCancelTask",
		trace.WithAttributes(attribute.String("a2a.task_id", req.Params.ID)))
	defer span.End()

	taskID := req.Params.ID
	if taskID == "" {
		return nil, errors.New("task ID cannot be empty")
	}

	tm.taskMu.Lock()
	task, ok := tm.tasks[taskID]
	if !ok {
		tm.taskMu.Unlock()
		tm.logger.InfoContext(ctx, "task not found", slog.String("task_id", taskID))
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Only allow cancellation of tasks that are not already in terminal states
	switch state := task.Status.State; state {
	case a2a.TaskStateCompleted, a2a.TaskStateCanceled, a2a.TaskStateFailed:
		tm.taskMu.Unlock()
		tm.logger.InfoContext(ctx, "task cannot be canceled", slog.String("task_id", taskID), slog.String("state", string(state)))
		return nil, fmt.Errorf("task cannot be canceled: already in state %s", state)
	}

	// Update task state
	task.Status.State = a2a.TaskStateCanceled
	task.Status.Timestamp = time.Now().UTC()
	tm.taskMu.Unlock()

	// Create status update event
	event := &a2a.TaskStatusUpdateEvent{
		ID:     taskID,
		Status: task.Status,
	}

	// Notify subscribers
	tm.notifySubscribers(ctx, taskID, event)

	tm.logger.InfoContext(ctx, "task canceled", slog.String("task_id", taskID))

	resp, err := jsonrpc2.NewResponse(req.ID, task, nil)
	if err != nil {
		return nil, err
	}
	return &a2a.CancelTaskResponse{
		Response: resp,
		Result:   task,
	}, nil
}

// OnSendTaskSubscribe starts a streaming task and returns a channel for updates.
func (tm *InMemoryTaskManager) OnSendTaskSubscribe(ctx context.Context, req *a2a.SendTaskStreamingRequest) (<-chan *a2a.SendTaskStreamingResponse, error) {
	// no-op
	return make(chan *a2a.SendTaskStreamingResponse, 0), nil
}

// OnSetTaskPushNotification configures push notification for a task.
func (tm *InMemoryTaskManager) OnSetTaskPushNotification(ctx context.Context, req *a2a.SetTaskPushNotificationRequest) (*a2a.SetTaskPushNotificationResponse, error) {
	ctx, span := tm.tracer.Start(ctx, "task_manager.OnSetTaskPushNotification",
		trace.WithAttributes(attribute.String("a2a.task_id", req.Params.ID)))
	defer span.End()

	task := req.Params

	if task.ID == "" {
		return nil, errors.New("task ID cannot be empty")
	}

	if task.PushNotificationConfig.URL == "" {
		return nil, errors.New("push notification URL cannot be empty")
	}

	// Verify task exists
	tm.taskMu.RLock()
	_, ok := tm.tasks[task.ID]
	tm.taskMu.RUnlock()

	if !ok {
		tm.logger.InfoContext(ctx, "task not found", slog.String("task_id", task.ID))
		return nil, fmt.Errorf("task not found: %s", task.ID)
	}

	// Store push notification config
	tm.pushMu.Lock()
	tm.pushNotifications[task.ID] = task
	tm.pushMu.Unlock()

	tm.logger.InfoContext(ctx, "task push notification configured", slog.String("task_id", task.ID))

	resp, err := jsonrpc2.NewResponse(req.ID, task, nil)
	if err != nil {
		return nil, err
	}
	return &a2a.SetTaskPushNotificationResponse{
		Response: resp,
		Result:   &task,
	}, nil
}

// OnGetTaskPushNotification retrieves push notification configuration for a task.
func (tm *InMemoryTaskManager) OnGetTaskPushNotification(ctx context.Context, req *a2a.GetTaskPushNotificationRequest) (*a2a.GetTaskPushNotificationResponse, error) {
	ctx, span := tm.tracer.Start(ctx, "task_manager.OnGetTaskPushNotification",
		trace.WithAttributes(attribute.String("a2a.task_id", req.Params.ID)))
	defer span.End()

	task := req.Params

	if task.ID == "" {
		return nil, errors.New("task ID cannot be empty")
	}

	// Verify task exists
	tm.taskMu.RLock()
	_, ok := tm.tasks[task.ID]
	tm.taskMu.RUnlock()

	if !ok {
		tm.logger.InfoContext(ctx, "task not found", slog.String("task_id", task.ID))
		return nil, fmt.Errorf("task not found: %s", task.ID)
	}

	// Get push notification config
	tm.pushMu.RLock()
	config, ok := tm.pushNotifications[task.ID]
	tm.pushMu.RUnlock()

	if !ok {
		tm.logger.InfoContext(ctx, "push notification not found", slog.String("task_id", task.ID))
		return nil, fmt.Errorf("push notification not found for task: %s", task.ID)
	}

	tm.logger.InfoContext(ctx, "task push notification retrieved", slog.String("task_id", task.ID))

	resp, err := jsonrpc2.NewResponse(req.ID, &config, nil)
	if err != nil {
		return nil, err
	}
	return &a2a.GetTaskPushNotificationResponse{
		Response: resp,
		Result:   &config,
	}, nil
}

// onResubscribeToTask resubscribes to a task's updates.
// func (tm *InMemoryTaskManager) onResubscribeToTask(ctx context.Context, req *a2a.TaskResubscriptionRequest) (any, error) {
// 	return &a2a.JSONRPCResponse{
// 		JSONRPCMessage: a2a.NewJSONRPCMessage(req.ID),
// 		Error:          a2a.NewUnsupportedOperationError(),
// 	}, nil
// }

// OnResubscribeToTask resubscribes to a task's updates.
func (tm *InMemoryTaskManager) OnResubscribeToTask(ctx context.Context, req *a2a.TaskResubscriptionRequest) (<-chan a2a.TaskEvent, error) {
	ctx, span := tm.tracer.Start(ctx, "task_manager.OnResubscribeToTask",
		trace.WithAttributes(attribute.String("a2a.task_id", req.Params.ID)))
	defer span.End()

	if req.Params.ID == "" {
		return nil, errors.New("task ID cannot be empty")
	}

	// Get task
	tm.taskMu.RLock()
	task, ok := tm.tasks[req.Params.ID]
	tm.taskMu.RUnlock()

	if !ok {
		tm.logger.InfoContext(ctx, "task not found", slog.String("task_id", req.Params.ID))
		return nil, fmt.Errorf("task not found: %s", req.Params.ID)
	}

	// Create event
	event := &a2a.TaskStatusUpdateEvent{
		ID:     task.ID,
		Status: task.Status,
	}

	// notify subscribers
	tm.notifySubscribers(ctx, req.Params.ID, event)

	tm.logger.InfoContext(ctx, "task resubscribed", slog.String("task_id", req.Params.ID))

	return nil, nil
}

// notifySubscribers sends an event to all subscribers of a task.
func (tm *InMemoryTaskManager) notifySubscribers(ctx context.Context, taskID string, event a2a.TaskEvent) {
	tm.subMu.RLock()
	subs := tm.subscribers[taskID]
	tm.subMu.RUnlock()

	for _, sub := range subs {
		go func(s chan a2a.TaskEvent) {
			select {
			case s <- event:
				// Event sent successfully
			case <-ctx.Done():
				// Context canceled
				tm.logger.InfoContext(ctx, "context canceled for event delivery", "task_id", taskID)
			default:
				// Channel full, log warning
				tm.logger.WarnContext(ctx, "subscriber channel full, dropping event", "task_id", taskID)
			}
		}(sub)
	}
}

// UpdateTaskStatus updates a task's status and notifies subscribers.
func (tm *InMemoryTaskManager) UpdateTaskStatus(ctx context.Context, taskID string, status a2a.TaskStatus, artifacts []a2a.Artifact) error {
	ctx, span := tm.tracer.Start(ctx, "task_manager.UpdateTaskStatus",
		trace.WithAttributes(
			attribute.String("a2a.task_id", taskID),
			attribute.String("a2a.task_state", string(status.State)),
		))
	defer span.End()

	if taskID == "" {
		return errors.New("task ID cannot be empty")
	}

	// Update task
	tm.taskMu.Lock()
	task, ok := tm.tasks[taskID]
	if !ok {
		tm.taskMu.Unlock()
		tm.logger.InfoContext(ctx, "task not found", slog.String("task_id", taskID))
		return fmt.Errorf("task not found: %s", taskID)
	}

	task.Status = status
	task.Status.Timestamp = time.Now().UTC()
	tm.taskMu.Unlock()

	// Create event
	event := &a2a.TaskStatusUpdateEvent{
		ID:     taskID,
		Status: status,
	}

	// Notify subscribers
	tm.notifySubscribers(ctx, taskID, event)

	tm.logger.InfoContext(ctx, "task status updated", slog.String("task_id", taskID), slog.String("state", string(status.State)))

	return nil
}
