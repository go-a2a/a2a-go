// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-a2a/a2a"
)

// TaskEvent represents an event related to a task.
type TaskEvent interface {
	// GetTaskID returns the task ID that this event is for.
	GetTaskID() string
}

// TaskManager is the interface that task managers must implement.
type TaskManager interface {
	// OnSendTask handles a new task.
	OnSendTask(ctx context.Context, task *a2a.Task) (*a2a.Task, error)

	// OnGetTask retrieves a task.
	OnGetTask(ctx context.Context, taskID string) (*a2a.Task, error)

	// OnCancelTask cancels a task.
	OnCancelTask(ctx context.Context, taskID string) (*a2a.Task, error)

	// OnSetTaskPushNotification configures push notification for a task.
	OnSetTaskPushNotification(ctx context.Context, config *a2a.TaskPushNotificationConfig) (*a2a.TaskPushNotificationConfig, error)

	// OnGetTaskPushNotification retrieves push notification configuration for a task.
	OnGetTaskPushNotification(ctx context.Context, taskID string) (*a2a.TaskPushNotificationConfig, error)

	// OnResubscribeToTask resubscribes to a task's updates.
	OnResubscribeToTask(ctx context.Context, taskID string, historyLength int) (*a2a.TaskStatusUpdateEvent, error)

	// OnSendTaskSubscribe starts a streaming task and returns a channel for updates.
	OnSendTaskSubscribe(ctx context.Context, task a2a.Task) (<-chan TaskEvent, error)
}

// InMemoryTaskManager is an in-memory implementation of TaskManager.
type InMemoryTaskManager struct {
	// Tasks is a map of task ID to task.
	tasks map[string]*a2a.Task

	// TaskMutex protects the tasks map.
	taskMutex sync.RWMutex

	// PushNotifications is a map of task ID to push notification config.
	pushNotifications map[string]a2a.TaskPushNotificationConfig

	// PushMutex protects the pushNotifications map.
	pushMutex sync.RWMutex

	// Subscribers is a map of task ID to a list of subscriber channels.
	subscribers map[string][]chan TaskEvent

	// SubMutex protects the subscribers map.
	subMutex sync.RWMutex

	// Logger is the logger for the task manager.
	Logger *slog.Logger

	// Tracer is the tracer for the task manager.
	Tracer trace.Tracer
}

// NewInMemoryTaskManager creates a new InMemoryTaskManager.
func NewInMemoryTaskManager() *InMemoryTaskManager {
	return &InMemoryTaskManager{
		tasks:             make(map[string]*a2a.Task),
		pushNotifications: make(map[string]a2a.TaskPushNotificationConfig),
		subscribers:       make(map[string][]chan TaskEvent),
		Logger:            slog.Default(),
		Tracer:            otel.GetTracerProvider().Tracer("github.com/go-a2a/a2a/task_manager"),
	}
}

// WithLogger sets the logger for the TaskManager.
func (tm *InMemoryTaskManager) WithLogger(logger *slog.Logger) *InMemoryTaskManager {
	tm.Logger = logger
	return tm
}

// WithTracer sets the tracer for the TaskManager.
func (tm *InMemoryTaskManager) WithTracer(tracer trace.Tracer) *InMemoryTaskManager {
	tm.Tracer = tracer
	return tm
}

// OnSendTask handles a new task.
func (tm *InMemoryTaskManager) OnSendTask(ctx context.Context, task a2a.Task) (*a2a.Task, error) {
	ctx, span := tm.Tracer.Start(ctx, "a2a.task_manager.OnSendTask",
		trace.WithAttributes(attribute.String("a2a.task_id", task.ID)))
	defer span.End()

	if task.ID == "" {
		return nil, errors.New("task ID cannot be empty")
	}

	// Ensure created/modified times are set
	now := time.Now().UTC()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	task.ModifiedAt = now

	// Set initial state if not set
	if task.State == "" {
		task.State = a2a.TaskStateSubmitted
	}

	// Store task
	tm.taskMutex.Lock()
	tm.tasks[task.ID] = &task
	tm.taskMutex.Unlock()

	// Create status update event
	event := a2a.TaskStatusUpdateEvent{
		Task:      task,
		RequestID: task.ID,
	}

	// Notify subscribers
	tm.notifySubscribers(ctx, task.ID, event)

	tm.Logger.InfoContext(ctx, "task created", "task_id", task.ID, "state", task.State)
	return &task, nil
}

// OnGetTask retrieves a task.
func (tm *InMemoryTaskManager) OnGetTask(ctx context.Context, taskID string) (*a2a.Task, error) {
	ctx, span := tm.Tracer.Start(ctx, "a2a.task_manager.OnGetTask",
		trace.WithAttributes(attribute.String("a2a.task_id", taskID)))
	defer span.End()

	if taskID == "" {
		return nil, errors.New("task ID cannot be empty")
	}

	tm.taskMutex.RLock()
	task, ok := tm.tasks[taskID]
	tm.taskMutex.RUnlock()

	if !ok {
		tm.Logger.InfoContext(ctx, "task not found", "task_id", taskID)
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	tm.Logger.InfoContext(ctx, "task retrieved", "task_id", taskID, "state", task.State)
	return task, nil
}

// OnCancelTask cancels a task.
func (tm *InMemoryTaskManager) OnCancelTask(ctx context.Context, taskID string) (*a2a.Task, error) {
	ctx, span := tm.Tracer.Start(ctx, "a2a.task_manager.OnCancelTask",
		trace.WithAttributes(attribute.String("a2a.task_id", taskID)))
	defer span.End()

	if taskID == "" {
		return nil, errors.New("task ID cannot be empty")
	}

	tm.taskMutex.Lock()
	task, ok := tm.tasks[taskID]
	if !ok {
		tm.taskMutex.Unlock()
		tm.Logger.InfoContext(ctx, "task not found", "task_id", taskID)
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Only allow cancellation of tasks that are not already in terminal states
	if task.State == a2a.TaskStateCompleted || task.State == a2a.TaskStateCanceled || task.State == a2a.TaskStateFailed {
		tm.taskMutex.Unlock()
		tm.Logger.InfoContext(ctx, "task cannot be canceled", "task_id", taskID, "state", task.State)
		return nil, fmt.Errorf("task cannot be canceled: already in state %s", task.State)
	}

	// Update task state
	task.State = a2a.TaskStateCanceled
	task.ModifiedAt = time.Now().UTC()
	tm.taskMutex.Unlock()

	// Create status update event
	event := a2a.TaskStatusUpdateEvent{
		Task:      *task,
		RequestID: taskID,
	}

	// Notify subscribers
	tm.notifySubscribers(ctx, taskID, event)

	tm.Logger.InfoContext(ctx, "task canceled", "task_id", taskID)
	return task, nil
}

// OnSetTaskPushNotification configures push notification for a task.
func (tm *InMemoryTaskManager) OnSetTaskPushNotification(ctx context.Context, config a2a.TaskPushNotificationConfig) (*a2a.TaskPushNotificationConfig, error) {
	ctx, span := tm.Tracer.Start(ctx, "a2a.task_manager.OnSetTaskPushNotification",
		trace.WithAttributes(attribute.String("a2a.task_id", config.TaskID)))
	defer span.End()

	if config.TaskID == "" {
		return nil, errors.New("task ID cannot be empty")
	}

	if config.PushNotificationConfig.URL == "" {
		return nil, errors.New("push notification URL cannot be empty")
	}

	// Verify task exists
	tm.taskMutex.RLock()
	_, ok := tm.tasks[config.TaskID]
	tm.taskMutex.RUnlock()

	if !ok {
		tm.Logger.InfoContext(ctx, "task not found", "task_id", config.TaskID)
		return nil, fmt.Errorf("task not found: %s", config.TaskID)
	}

	// Store push notification config
	tm.pushMutex.Lock()
	tm.pushNotifications[config.TaskID] = config
	tm.pushMutex.Unlock()

	tm.Logger.InfoContext(ctx, "task push notification configured", "task_id", config.TaskID)
	return &config, nil
}

// OnGetTaskPushNotification retrieves push notification configuration for a task.
func (tm *InMemoryTaskManager) OnGetTaskPushNotification(ctx context.Context, taskID string) (*a2a.TaskPushNotificationConfig, error) {
	ctx, span := tm.Tracer.Start(ctx, "a2a.task_manager.OnGetTaskPushNotification",
		trace.WithAttributes(attribute.String("a2a.task_id", taskID)))
	defer span.End()

	if taskID == "" {
		return nil, errors.New("task ID cannot be empty")
	}

	// Verify task exists
	tm.taskMutex.RLock()
	_, ok := tm.tasks[taskID]
	tm.taskMutex.RUnlock()

	if !ok {
		tm.Logger.InfoContext(ctx, "task not found", "task_id", taskID)
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Get push notification config
	tm.pushMutex.RLock()
	config, ok := tm.pushNotifications[taskID]
	tm.pushMutex.RUnlock()

	if !ok {
		tm.Logger.InfoContext(ctx, "push notification not found", "task_id", taskID)
		return nil, fmt.Errorf("push notification not found for task: %s", taskID)
	}

	tm.Logger.InfoContext(ctx, "task push notification retrieved", "task_id", taskID)
	return &config, nil
}

// OnResubscribeToTask resubscribes to a task's updates.
func (tm *InMemoryTaskManager) OnResubscribeToTask(ctx context.Context, taskID string, historyLength int) (*a2a.TaskStatusUpdateEvent, error) {
	ctx, span := tm.Tracer.Start(ctx, "a2a.task_manager.OnResubscribeToTask",
		trace.WithAttributes(attribute.String("a2a.task_id", taskID)))
	defer span.End()

	if taskID == "" {
		return nil, errors.New("task ID cannot be empty")
	}

	// Get task
	tm.taskMutex.RLock()
	task, ok := tm.tasks[taskID]
	tm.taskMutex.RUnlock()

	if !ok {
		tm.Logger.InfoContext(ctx, "task not found", "task_id", taskID)
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Create event
	event := a2a.TaskStatusUpdateEvent{
		Task:      *task,
		RequestID: taskID,
	}

	tm.Logger.InfoContext(ctx, "task resubscribed", "task_id", taskID)
	return &event, nil
}

// OnSendTaskSubscribe starts a streaming task and returns a channel for updates.
func (tm *InMemoryTaskManager) OnSendTaskSubscribe(ctx context.Context, task a2a.Task) (<-chan TaskEvent, error) {
	ctx, span := tm.Tracer.Start(ctx, "a2a.task_manager.OnSendTaskSubscribe",
		trace.WithAttributes(attribute.String("a2a.task_id", task.ID)))
	defer span.End()

	if task.ID == "" {
		return nil, errors.New("task ID cannot be empty")
	}

	// Process the task first
	processedTask, err := tm.OnSendTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to process task: %w", err)
	}

	// Create subscriber channel
	eventCh := make(chan TaskEvent, 10)

	// Add to subscribers
	tm.subMutex.Lock()
	if _, ok := tm.subscribers[task.ID]; !ok {
		tm.subscribers[task.ID] = make([]chan TaskEvent, 0)
	}
	tm.subscribers[task.ID] = append(tm.subscribers[task.ID], eventCh)
	tm.subMutex.Unlock()

	// Send initial event
	initialEvent := a2a.TaskStatusUpdateEvent{
		Task:      *processedTask,
		RequestID: task.ID,
	}

	go func() {
		select {
		case eventCh <- initialEvent:
			// Event sent successfully
		case <-ctx.Done():
			// Context canceled
			tm.Logger.InfoContext(ctx, "context canceled for task subscription", "task_id", task.ID)
			close(eventCh)
		}
	}()

	tm.Logger.InfoContext(ctx, "task subscription created", "task_id", task.ID)
	return eventCh, nil
}

// notifySubscribers sends an event to all subscribers of a task.
func (tm *InMemoryTaskManager) notifySubscribers(ctx context.Context, taskID string, event TaskEvent) {
	tm.subMutex.RLock()
	subs := tm.subscribers[taskID]
	tm.subMutex.RUnlock()

	for _, sub := range subs {
		go func(s chan TaskEvent) {
			select {
			case s <- event:
				// Event sent successfully
			case <-ctx.Done():
				// Context canceled
				tm.Logger.InfoContext(ctx, "context canceled for event delivery", "task_id", taskID)
			default:
				// Channel full, log warning
				tm.Logger.WarnContext(ctx, "subscriber channel full, dropping event", "task_id", taskID)
			}
		}(sub)
	}
}

// UpdateTaskStatus updates a task's status and notifies subscribers.
func (tm *InMemoryTaskManager) UpdateTaskStatus(ctx context.Context, taskID string, state a2a.TaskState) error {
	ctx, span := tm.Tracer.Start(ctx, "a2a.task_manager.UpdateTaskStatus",
		trace.WithAttributes(
			attribute.String("a2a.task_id", taskID),
			attribute.String("a2a.task_state", string(state)),
		))
	defer span.End()

	if taskID == "" {
		return errors.New("task ID cannot be empty")
	}

	// Update task
	tm.taskMutex.Lock()
	task, ok := tm.tasks[taskID]
	if !ok {
		tm.taskMutex.Unlock()
		tm.Logger.InfoContext(ctx, "task not found", "task_id", taskID)
		return fmt.Errorf("task not found: %s", taskID)
	}

	task.State = state
	task.ModifiedAt = time.Now().UTC()
	tm.taskMutex.Unlock()

	// Create event
	event := a2a.TaskStatusUpdateEvent{
		Task:      *task,
		RequestID: taskID,
	}

	// Notify subscribers
	tm.notifySubscribers(ctx, taskID, event)

	tm.Logger.InfoContext(ctx, "task status updated", "task_id", taskID, "state", state)
	return nil
}

// AddTaskMessage adds a message to a task and notifies subscribers.
func (tm *InMemoryTaskManager) AddTaskMessage(ctx context.Context, taskID string, message a2a.Message) error {
	ctx, span := tm.Tracer.Start(ctx, "a2a.task_manager.AddTaskMessage",
		trace.WithAttributes(attribute.String("a2a.task_id", taskID)))
	defer span.End()

	if taskID == "" {
		return errors.New("task ID cannot be empty")
	}

	// Set message created time if not set
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now().UTC()
	}

	// Update task
	tm.taskMutex.Lock()
	task, ok := tm.tasks[taskID]
	if !ok {
		tm.taskMutex.Unlock()
		tm.Logger.InfoContext(ctx, "task not found", "task_id", taskID)
		return fmt.Errorf("task not found: %s", taskID)
	}

	task.Messages = append(task.Messages, message)
	task.ModifiedAt = time.Now().UTC()
	tm.taskMutex.Unlock()

	// Create event
	event := a2a.TaskStatusUpdateEvent{
		Task:      *task,
		RequestID: taskID,
	}

	// Notify subscribers
	tm.notifySubscribers(ctx, taskID, event)

	tm.Logger.InfoContext(ctx, "task message added", "task_id", taskID, "role", message.Role)
	return nil
}

// AddTaskArtifact adds an artifact to a task and notifies subscribers.
func (tm *InMemoryTaskManager) AddTaskArtifact(ctx context.Context, taskID string, artifact a2a.Artifact) error {
	ctx, span := tm.Tracer.Start(ctx, "a2a.task_manager.AddTaskArtifact",
		trace.WithAttributes(attribute.String("a2a.task_id", taskID)))
	defer span.End()

	if taskID == "" {
		return errors.New("task ID cannot be empty")
	}

	// Set artifact created/modified times if not set
	now := time.Now().UTC()
	if artifact.CreatedAt.IsZero() {
		artifact.CreatedAt = now
	}
	artifact.ModifiedAt = now

	// Update task
	tm.taskMutex.Lock()
	task, ok := tm.tasks[taskID]
	if !ok {
		tm.taskMutex.Unlock()
		tm.Logger.InfoContext(ctx, "task not found", "task_id", taskID)
		return fmt.Errorf("task not found: %s", taskID)
	}

	task.Artifacts = append(task.Artifacts, artifact)
	task.ModifiedAt = now
	tm.taskMutex.Unlock()

	// Create event
	event := a2a.TaskStatusUpdateEvent{
		Task:      *task,
		RequestID: taskID,
	}

	// Notify subscribers
	tm.notifySubscribers(ctx, taskID, event)

	tm.Logger.InfoContext(ctx, "task artifact added", "task_id", taskID, "artifact_id", artifact.ID)
	return nil
}

// RemoveTaskSubscriber removes a subscriber channel for a task.
func (tm *InMemoryTaskManager) RemoveTaskSubscriber(ctx context.Context, taskID string, ch chan TaskEvent) {
	ctx, span := tm.Tracer.Start(ctx, "a2a.task_manager.RemoveTaskSubscriber",
		trace.WithAttributes(attribute.String("a2a.task_id", taskID)))
	defer span.End()

	tm.subMutex.Lock()
	defer tm.subMutex.Unlock()

	subs, ok := tm.subscribers[taskID]
	if !ok {
		return
	}

	// Find and remove the channel
	for i, sub := range subs {
		if sub == ch {
			// Remove the channel from the slice
			tm.subscribers[taskID] = slices.Delete(subs, i, i+1)
			close(ch)
			break
		}
	}

	// If no more subscribers, remove the entry
	if len(tm.subscribers[taskID]) == 0 {
		delete(tm.subscribers, taskID)
	}

	tm.Logger.InfoContext(ctx, "task subscriber removed", "task_id", taskID)
}
