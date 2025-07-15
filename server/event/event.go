// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package event provides event handling for the A2A server.
// This package implements event queues, consumers, and managers
// for asynchronous event processing within the A2A Go SDK.
package event

import (
	"context"
	"fmt"

	"github.com/go-a2a/a2a"
)

// Event represents a unified interface for all event types in the A2A system.
// This interface allows different event types to be handled uniformly
// throughout the event processing pipeline.
type Event interface {
	// EventType returns the type of the event (e.g., "message", "task", "task_status_update", "task_artifact_update")
	EventType() string

	// EventData returns the underlying data of the event
	EventData() any

	// Validate ensures the event is in a valid state
	Validate() error

	// String returns a string representation of the event
	String() string
}

// MessageEvent wraps an a2a.Message as an event.
type MessageEvent struct {
	Message *a2a.Message
}

var _ Event = (*MessageEvent)(nil)

// EventType returns the event type for MessageEvent.
func (me *MessageEvent) EventType() string {
	return "message"
}

// EventData returns the underlying message data.
func (me *MessageEvent) EventData() any {
	return me.Message
}

// Validate ensures the MessageEvent is valid.
func (me *MessageEvent) Validate() error {
	if me.Message == nil {
		return fmt.Errorf("message event message cannot be nil")
	}
	return me.Message.Validate()
}

// String returns a string representation of the MessageEvent.
func (me *MessageEvent) String() string {
	if me.Message == nil {
		return "MessageEvent{Message: nil}"
	}
	return fmt.Sprintf("MessageEvent{ContextID: %s, Content: %.50s...}",
		me.Message.ContextID, me.Message.Content)
}

// TaskEvent wraps an a2a.Task as an event.
type TaskEvent struct {
	Task *a2a.Task
}

var _ Event = (*TaskEvent)(nil)

// EventType returns the event type for TaskEvent.
func (te *TaskEvent) EventType() string {
	return "task"
}

// EventData returns the underlying task data.
func (te *TaskEvent) EventData() any {
	return te.Task
}

// Validate ensures the TaskEvent is valid.
func (te *TaskEvent) Validate() error {
	if te.Task == nil {
		return fmt.Errorf("task event task cannot be nil")
	}
	return te.Task.Validate()
}

// String returns a string representation of the TaskEvent.
func (te *TaskEvent) String() string {
	if te.Task == nil {
		return "TaskEvent{Task: nil}"
	}
	return fmt.Sprintf("TaskEvent{ID: %s, ContextID: %s, Status: %s}",
		te.Task.ID, te.Task.ContextID, te.Task.Status.State)
}

// TaskStatusUpdateEvent represents a task status update event.
type TaskStatusUpdateEvent struct {
	TaskID    string         `json:"task_id"`
	ContextID string         `json:"context_id,omitempty"`
	Status    a2a.TaskStatus `json:"status"`
	Final     bool           `json:"final"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

var _ Event = (*TaskStatusUpdateEvent)(nil)

// EventType returns the event type for TaskStatusUpdateEvent.
func (tsue *TaskStatusUpdateEvent) EventType() string {
	return "task_status_update"
}

// EventData returns the underlying status update data.
func (tsue *TaskStatusUpdateEvent) EventData() any {
	return tsue
}

// Validate ensures the TaskStatusUpdateEvent is valid.
func (tsue *TaskStatusUpdateEvent) Validate() error {
	if tsue.TaskID == "" {
		return fmt.Errorf("task status update event task ID cannot be empty")
	}
	return tsue.Status.Validate()
}

// String returns a string representation of the TaskStatusUpdateEvent.
func (tsue *TaskStatusUpdateEvent) String() string {
	return fmt.Sprintf("TaskStatusUpdateEvent{TaskID: %s, Status: %s, Final: %t}",
		tsue.TaskID, tsue.Status.State, tsue.Final)
}

// TaskArtifactUpdateEvent represents a task artifact update event.
type TaskArtifactUpdateEvent struct {
	TaskID   string        `json:"task_id"`
	Artifact *a2a.Artifact `json:"artifact"`
	Append   bool          `json:"append"`
}

var _ Event = (*TaskArtifactUpdateEvent)(nil)

// EventType returns the event type for TaskArtifactUpdateEvent.
func (taue *TaskArtifactUpdateEvent) EventType() string {
	return "task_artifact_update"
}

// EventData returns the underlying artifact update data.
func (taue *TaskArtifactUpdateEvent) EventData() any {
	return taue
}

// Validate ensures the TaskArtifactUpdateEvent is valid.
func (taue *TaskArtifactUpdateEvent) Validate() error {
	if taue.TaskID == "" {
		return fmt.Errorf("task artifact update event task ID cannot be empty")
	}
	if taue.Artifact == nil {
		return fmt.Errorf("task artifact update event artifact cannot be nil")
	}
	return taue.Artifact.Validate()
}

// String returns a string representation of the TaskArtifactUpdateEvent.
func (taue *TaskArtifactUpdateEvent) String() string {
	artifactName := "nil"
	if taue.Artifact != nil {
		artifactName = taue.Artifact.Name
	}
	return fmt.Sprintf("TaskArtifactUpdateEvent{TaskID: %s, Artifact: %s, Append: %t}",
		taue.TaskID, artifactName, taue.Append)
}

// NewMessageEvent creates a new MessageEvent.
func NewMessageEvent(message *a2a.Message) *MessageEvent {
	return &MessageEvent{Message: message}
}

// NewTaskEvent creates a new TaskEvent.
func NewTaskEvent(task *a2a.Task) *TaskEvent {
	return &TaskEvent{Task: task}
}

// NewTaskStatusUpdateEvent creates a new TaskStatusUpdateEvent.
func NewTaskStatusUpdateEvent(taskID, contextID string, status a2a.TaskStatus, final bool, metadata map[string]any) *TaskStatusUpdateEvent {
	return &TaskStatusUpdateEvent{
		TaskID:    taskID,
		ContextID: contextID,
		Status:    status,
		Final:     final,
		Metadata:  metadata,
	}
}

// NewTaskArtifactUpdateEvent creates a new TaskArtifactUpdateEvent.
func NewTaskArtifactUpdateEvent(taskID string, artifact *a2a.Artifact, append bool) *TaskArtifactUpdateEvent {
	return &TaskArtifactUpdateEvent{
		TaskID:   taskID,
		Artifact: artifact,
		Append:   append,
	}
}

// IsFinalEvent determines if an event represents a final state.
// Final events are:
// - TaskStatusUpdateEvent with Final=true
// - TaskEvent with terminal status (completed, failed, canceled)
// - MessageEvent (always considered final)
func IsFinalEvent(event Event) bool {
	if event == nil {
		return false
	}

	switch e := event.(type) {
	case *TaskStatusUpdateEvent:
		return e.Final
	case *TaskEvent:
		if e.Task == nil {
			return false
		}
		return e.Task.Status.State == a2a.TaskStateCompleted ||
			e.Task.Status.State == a2a.TaskStateFailed ||
			e.Task.Status.State == a2a.TaskStateCanceled
	case *MessageEvent:
		return true
	case *TaskArtifactUpdateEvent:
		return false
	default:
		return false
	}
}

// IsTerminalTaskState checks if a task state is terminal.
func IsTerminalTaskState(state a2a.TaskState) bool {
	return state == a2a.TaskStateCompleted ||
		state == a2a.TaskStateFailed ||
		state == a2a.TaskStateCanceled ||
		state == a2a.TaskStateRejected
}

// EventFromContext extracts event context information if available.
func EventFromContext(ctx context.Context) (Event, bool) {
	if ctx == nil {
		return nil, false
	}

	event, ok := ctx.Value("event").(Event)
	return event, ok
}

// ContextWithEvent creates a context with an event.
func ContextWithEvent(ctx context.Context, event Event) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, "event", event)
}
