// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/server/event"
)

// TaskManager manages task lifecycle during request execution.
// It handles task retrieval, saving, and updating based on events from agents.
type TaskManager interface {
	// GetTask retrieves the current task from memory or storage.
	GetTask(ctx context.Context) (*a2a.Task, error)

	// SaveTaskEvent processes a task-related event and persists the updated task state.
	SaveTaskEvent(ctx context.Context, event event.Event) error

	// EnsureTask guarantees that a task exists in memory, creating it if necessary.
	EnsureTask(ctx context.Context, event event.Event) (*a2a.Task, error)

	// Process handles an incoming event, updating task state if it's task-related.
	Process(ctx context.Context, event event.Event) error

	// GetTaskID returns the task ID this manager is associated with.
	GetTaskID() string

	// GetContextID returns the context ID this manager is associated with.
	GetContextID() string

	// Close shuts down the manager and releases resources.
	Close() error
}

// TaskManagerConfig holds configuration for creating a TaskManager.
type TaskManagerConfig struct {
	TaskID         string
	ContextID      string
	Store          TaskStore
	InitialMessage *a2a.Message
}

// defaultTaskManager is the default implementation of TaskManager.
type defaultTaskManager struct {
	taskID         string
	contextID      string
	store          TaskStore
	initialMessage *a2a.Message

	mu     sync.RWMutex
	task   *a2a.Task
	closed bool
}

var _ TaskManager = (*defaultTaskManager)(nil)

// NewTaskManager creates a new TaskManager with the given configuration.
func NewTaskManager(config TaskManagerConfig) (TaskManager, error) {
	if config.TaskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}
	if config.ContextID == "" {
		return nil, fmt.Errorf("context ID cannot be empty")
	}
	if config.Store == nil {
		return nil, fmt.Errorf("task store cannot be nil")
	}

	return &defaultTaskManager{
		taskID:         config.TaskID,
		contextID:      config.ContextID,
		store:          config.Store,
		initialMessage: config.InitialMessage,
	}, nil
}

// GetTask retrieves the current task from memory or storage.
func (m *defaultTaskManager) GetTask(ctx context.Context) (*a2a.Task, error) {
	m.mu.RLock()
	if m.task != nil {
		task := m.task
		m.mu.RUnlock()
		return task, nil
	}
	m.mu.RUnlock()

	// Try to load from storage
	task, err := m.store.Get(ctx, m.taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task from store: %w", err)
	}

	// Cache the task
	m.mu.Lock()
	m.task = task
	m.mu.Unlock()

	return task, nil
}

// SaveTaskEvent processes a task-related event and persists the updated task state.
func (m *defaultTaskManager) SaveTaskEvent(ctx context.Context, event event.Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// Ensure we have a task
	task, err := m.EnsureTask(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to ensure task: %w", err)
	}

	// Update task based on event type
	eventType := event.EventType()
	switch eventType {
	case "task_status_update":
		err = m.updateTaskStatus(task, event)
	case "task_artifact_update":
		err = m.updateTaskArtifact(task, event)
	case "task":
		if taskEvent, ok := event.EventData().(*a2a.Task); ok {
			task = taskEvent
		}
	default:
		// Not a task-related event, ignore
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Persist the updated task
	if err := m.store.Save(ctx, task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	// Update cached task
	m.mu.Lock()
	m.task = task
	m.mu.Unlock()

	return nil
}

// EnsureTask guarantees that a task exists in memory, creating it if necessary.
func (m *defaultTaskManager) EnsureTask(ctx context.Context, event event.Event) (*a2a.Task, error) {
	// Try to get existing task first
	task, err := m.GetTask(ctx)
	if err == nil {
		return task, nil
	}

	// Check if it's a "not found" error
	if !isTaskNotFoundError(err) {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Create new task
	task, err = m.createNewTask(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("failed to create new task: %w", err)
	}

	// Save the new task
	if err := m.store.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to save new task: %w", err)
	}

	// Cache the task
	m.mu.Lock()
	m.task = task
	m.mu.Unlock()

	return task, nil
}

// Process handles an incoming event, updating task state if it's task-related.
func (m *defaultTaskManager) Process(ctx context.Context, event event.Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return fmt.Errorf("task manager is closed")
	}
	m.mu.RUnlock()

	// Check if this is a task-related event
	if !m.isTaskRelatedEvent(event) {
		return nil
	}

	return m.SaveTaskEvent(ctx, event)
}

// GetTaskID returns the task ID this manager is associated with.
func (m *defaultTaskManager) GetTaskID() string {
	return m.taskID
}

// GetContextID returns the context ID this manager is associated with.
func (m *defaultTaskManager) GetContextID() string {
	return m.contextID
}

// Close shuts down the manager and releases resources.
func (m *defaultTaskManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}

	m.closed = true
	m.task = nil
	return nil
}

// updateTaskStatus updates the task status based on a status update event.
func (m *defaultTaskManager) updateTaskStatus(task *a2a.Task, evt event.Event) error {
	eventData := evt.EventData()
	statusEvent, ok := eventData.(*event.TaskStatusUpdateEvent)
	if !ok {
		return fmt.Errorf("invalid event data type for status update")
	}

	if statusEvent.TaskID != m.taskID {
		return fmt.Errorf("event task ID %s does not match manager task ID %s", statusEvent.TaskID, m.taskID)
	}

	// Update task status
	task.Status = statusEvent.Status

	// Add message to history if provided
	if message, ok := statusEvent.Metadata["message"].(string); ok && message != "" {
		historyMessage := &a2a.Message{
			ContextID: m.contextID,
			Content:   message,
			Metadata: map[string]any{
				"timestamp":  time.Now().UTC(),
				"event_type": "status_update",
			},
		}
		task.History = append(task.History, historyMessage)
	}

	return nil
}

// updateTaskArtifact updates the task artifacts based on an artifact update event.
func (m *defaultTaskManager) updateTaskArtifact(task *a2a.Task, evt event.Event) error {
	eventData := evt.EventData()
	artifactEvent, ok := eventData.(*event.TaskArtifactUpdateEvent)
	if !ok {
		return fmt.Errorf("invalid event data type for artifact update")
	}

	if artifactEvent.TaskID != m.taskID {
		return fmt.Errorf("event task ID %s does not match manager task ID %s", artifactEvent.TaskID, m.taskID)
	}

	if artifactEvent.Append {
		// Append artifact to existing artifacts
		task.Artifacts = append(task.Artifacts, artifactEvent.Artifact)
	} else {
		// Replace all artifacts with this one
		task.Artifacts = []*a2a.Artifact{artifactEvent.Artifact}
	}

	return nil
}

// createNewTask creates a new task based on the provided event.
func (m *defaultTaskManager) createNewTask(ctx context.Context, event event.Event) (*a2a.Task, error) {
	// Start with a basic task
	task := &a2a.Task{
		ID:        m.taskID,
		ContextID: m.contextID,
		Status:    a2a.TaskStatus{State: a2a.TaskStateSubmitted},
		History:   []*a2a.Message{},
		Artifacts: []*a2a.Artifact{},
	}

	// Add initial message if provided
	if m.initialMessage != nil {
		task.History = append(task.History, m.initialMessage)
	}

	return task, nil
}

// isTaskRelatedEvent checks if an event is related to this task manager.
func (m *defaultTaskManager) isTaskRelatedEvent(evt event.Event) bool {
	eventType := evt.EventType()
	eventData := evt.EventData()

	switch eventType {
	case "task_status_update":
		if statusEvent, ok := eventData.(*event.TaskStatusUpdateEvent); ok {
			return statusEvent.TaskID == m.taskID
		}
	case "task_artifact_update":
		if artifactEvent, ok := eventData.(*event.TaskArtifactUpdateEvent); ok {
			return artifactEvent.TaskID == m.taskID
		}
	case "task":
		if taskData, ok := eventData.(*a2a.Task); ok {
			return taskData.ID == m.taskID
		}
	}
	return false
}

// isTaskNotFoundError checks if an error is a "task not found" error.
func isTaskNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for TaskNotFoundError type
	var notFoundErr a2a.TaskNotFoundError
	if errors.As(err, &notFoundErr) {
		return true
	}

	// Check for error message patterns
	errMsg := err.Error()
	return strings.Contains(errMsg, "task not found") ||
		strings.Contains(errMsg, "record not found")
}
