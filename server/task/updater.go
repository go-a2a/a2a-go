// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/server/event"
)

// TaskUpdater provides the interface for agents to publish task-related events.
// It allows agents to update task status, add artifacts, and manage task lifecycle
// while ensuring thread safety and preventing updates to terminal states.
type TaskUpdater interface {
	// UpdateStatus updates the task status with an optional message.
	// If final is true, the task is marked as terminal and no further updates are allowed.
	UpdateStatus(ctx context.Context, state a2a.TaskState, message string, final bool) error

	// AddArtifact adds an artifact to the task.
	// If append is true, the artifact is appended to existing artifacts.
	// If append is false, the artifact replaces existing artifacts.
	AddArtifact(ctx context.Context, artifact *a2a.Artifact, append bool) error

	// Convenience methods for common status transitions
	Submit(ctx context.Context, message string) error
	StartWork(ctx context.Context, message string) error
	Complete(ctx context.Context, message string) error
	Failed(ctx context.Context, message string) error
	Reject(ctx context.Context, message string) error
	Cancel(ctx context.Context, message string) error
	RequiresInput(ctx context.Context, message string) error
	RequiresAuth(ctx context.Context, message string) error

	// GetTaskID returns the task ID this updater is associated with.
	GetTaskID() string

	// GetContextID returns the context ID this updater is associated with.
	GetContextID() string

	// IsTerminal returns true if the task is in a terminal state.
	IsTerminal() bool

	// Close shuts down the updater and releases resources.
	Close() error
}

// TaskUpdaterConfig holds configuration for creating a TaskUpdater.
type TaskUpdaterConfig struct {
	TaskID    string
	ContextID string
	Queue     *event.EventQueue
}

// defaultTaskUpdater is the default implementation of TaskUpdater.
type defaultTaskUpdater struct {
	taskID    string
	contextID string
	queue     *event.EventQueue

	mu       sync.RWMutex
	terminal bool
	closed   bool
}

var _ TaskUpdater = (*defaultTaskUpdater)(nil)

// NewTaskUpdater creates a new TaskUpdater with the given configuration.
func NewTaskUpdater(config TaskUpdaterConfig) (TaskUpdater, error) {
	if config.TaskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}
	if config.ContextID == "" {
		return nil, fmt.Errorf("context ID cannot be empty")
	}
	if config.Queue == nil {
		return nil, fmt.Errorf("event queue cannot be nil")
	}

	return &defaultTaskUpdater{
		taskID:    config.TaskID,
		contextID: config.ContextID,
		queue:     config.Queue,
	}, nil
}

// UpdateStatus updates the task status with an optional message.
func (u *defaultTaskUpdater) UpdateStatus(ctx context.Context, state a2a.TaskState, message string, final bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.closed {
		return fmt.Errorf("task updater is closed")
	}

	if u.terminal {
		return fmt.Errorf("cannot update task in terminal state")
	}

	// Mark as terminal if this is a final update or if state is terminal
	if final || a2a.IsTerminalTaskState(state) {
		u.terminal = true
	}

	// Create status update event
	status := a2a.TaskStatus{State: state}
	metadata := make(map[string]any)
	if message != "" {
		metadata["message"] = message
	}
	metadata["timestamp"] = time.Now().UTC()

	statusEvent := event.NewTaskStatusUpdateEvent(u.taskID, u.contextID, status, u.terminal, metadata)

	// Publish the event
	if err := u.queue.EnqueueEvent(ctx, statusEvent); err != nil {
		return fmt.Errorf("failed to publish status update event: %w", err)
	}

	return nil
}

// AddArtifact adds an artifact to the task.
func (u *defaultTaskUpdater) AddArtifact(ctx context.Context, artifact *a2a.Artifact, append bool) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.closed {
		return fmt.Errorf("task updater is closed")
	}

	if u.terminal {
		return fmt.Errorf("cannot add artifact to task in terminal state")
	}

	if artifact == nil {
		return fmt.Errorf("artifact cannot be nil")
	}

	if err := artifact.Validate(); err != nil {
		return fmt.Errorf("artifact validation failed: %w", err)
	}

	// Create artifact update event
	artifactEvent := event.NewTaskArtifactUpdateEvent(u.taskID, artifact, append)

	// Publish the event
	if err := u.queue.EnqueueEvent(ctx, artifactEvent); err != nil {
		return fmt.Errorf("failed to publish artifact update event: %w", err)
	}

	return nil
}

// Submit marks the task as submitted.
func (u *defaultTaskUpdater) Submit(ctx context.Context, message string) error {
	return u.UpdateStatus(ctx, a2a.TaskStateSubmitted, message, false)
}

// StartWork marks the task as running.
func (u *defaultTaskUpdater) StartWork(ctx context.Context, message string) error {
	return u.UpdateStatus(ctx, a2a.TaskStateRunning, message, false)
}

// Complete marks the task as completed.
func (u *defaultTaskUpdater) Complete(ctx context.Context, message string) error {
	return u.UpdateStatus(ctx, a2a.TaskStateCompleted, message, true)
}

// Failed marks the task as failed.
func (u *defaultTaskUpdater) Failed(ctx context.Context, message string) error {
	return u.UpdateStatus(ctx, a2a.TaskStateFailed, message, true)
}

// Reject marks the task as rejected.
func (u *defaultTaskUpdater) Reject(ctx context.Context, message string) error {
	return u.UpdateStatus(ctx, a2a.TaskStateRejected, message, true)
}

// Cancel marks the task as canceled.
func (u *defaultTaskUpdater) Cancel(ctx context.Context, message string) error {
	return u.UpdateStatus(ctx, a2a.TaskStateCanceled, message, true)
}

// RequiresInput marks the task as requiring input.
func (u *defaultTaskUpdater) RequiresInput(ctx context.Context, message string) error {
	return u.UpdateStatus(ctx, a2a.TaskStateInputRequired, message, false)
}

// RequiresAuth marks the task as requiring authentication.
func (u *defaultTaskUpdater) RequiresAuth(ctx context.Context, message string) error {
	return u.UpdateStatus(ctx, a2a.TaskStateAuthRequired, message, false)
}

// GetTaskID returns the task ID this updater is associated with.
func (u *defaultTaskUpdater) GetTaskID() string {
	return u.taskID
}

// GetContextID returns the context ID this updater is associated with.
func (u *defaultTaskUpdater) GetContextID() string {
	return u.contextID
}

// IsTerminal returns true if the task is in a terminal state.
func (u *defaultTaskUpdater) IsTerminal() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.terminal
}

// Close shuts down the updater and releases resources.
func (u *defaultTaskUpdater) Close() error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.closed {
		return nil
	}

	u.closed = true
	return nil
}
