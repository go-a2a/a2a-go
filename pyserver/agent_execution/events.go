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

package agent_execution

import (
	"time"

	"github.com/go-a2a/a2a-go"
)

// Event type constants.
const (
	EventTypeTaskStatus   = "task-status"
	EventTypeTaskArtifact = "task-artifact"
	EventTypeMessage      = "message"
	EventTypeFinal        = "final"
)

// BaseEvent provides common fields for all event types.
type BaseEvent struct {
	// EventType identifies the type of event.
	EventType string
	// Timestamp when the event was created.
	Timestamp time.Time
}

// GetEventType returns the type of event.
func (e BaseEvent) GetEventType() string {
	return e.EventType
}

// TaskStatusUpdateEvent represents a change in task status.
type TaskStatusUpdateEvent struct {
	BaseEvent
	// TaskID is the ID of the task being updated.
	TaskID string
	// ContextID is the context ID associated with the task.
	ContextID string
	// Status contains the new status information.
	Status *TaskStatus
	// Final indicates if this is the terminal update for the task.
	Final bool
}

// TaskStatus represents the status of a task at a point in time.
type TaskStatus struct {
	// State is the current state of the task.
	State a2a.TaskState
	// Message provides additional human-readable status information.
	Message string
	// Timestamp is when this status was set.
	Timestamp time.Time
}

// NewTaskStatusUpdateEvent creates a new task status update event.
func NewTaskStatusUpdateEvent(taskID, contextID string, state a2a.TaskState, message string) *TaskStatusUpdateEvent {
	now := time.Now()
	return &TaskStatusUpdateEvent{
		BaseEvent: BaseEvent{
			EventType: EventTypeTaskStatus,
			Timestamp: now,
		},
		TaskID:    taskID,
		ContextID: contextID,
		Status: &TaskStatus{
			State:     state,
			Message:   message,
			Timestamp: now,
		},
		Final: state == a2a.TaskStateCompleted || state == a2a.TaskStateFailed ||
			state == a2a.TaskStateCanceled || state == a2a.TaskStateRejected,
	}
}

// TaskArtifactUpdateEvent represents a new or updated artifact.
type TaskArtifactUpdateEvent struct {
	BaseEvent
	// TaskID is the ID of the task that produced the artifact.
	TaskID string
	// ContextID is the context ID associated with the task.
	ContextID string
	// Artifact contains the artifact data.
	Artifact *a2a.Artifact
	// Append indicates whether to append parts to existing artifact.
	Append bool
	// LastChunk indicates if this is the final update for the artifact.
	LastChunk bool
}

// NewTaskArtifactUpdateEvent creates a new task artifact update event.
func NewTaskArtifactUpdateEvent(taskID, contextID string, artifact *a2a.Artifact) *TaskArtifactUpdateEvent {
	return &TaskArtifactUpdateEvent{
		BaseEvent: BaseEvent{
			EventType: EventTypeTaskArtifact,
			Timestamp: time.Now(),
		},
		TaskID:    taskID,
		ContextID: contextID,
		Artifact:  artifact,
	}
}

// MessageEvent represents a message from the agent.
type MessageEvent struct {
	BaseEvent
	// Message contains the message content.
	Message *a2a.Message
	// ContextID is the context ID associated with the message.
	ContextID string
	// Final indicates if this is the final message for the task.
	Final bool
}

// NewMessageEvent creates a new message event.
func NewMessageEvent(message *a2a.Message, contextID string) *MessageEvent {
	return &MessageEvent{
		BaseEvent: BaseEvent{
			EventType: EventTypeMessage,
			Timestamp: time.Now(),
		},
		Message:   message,
		ContextID: contextID,
	}
}

// FinalEvent represents the final event in a stream.
type FinalEvent struct {
	BaseEvent
	// TaskID is the ID of the completed task.
	TaskID string
	// ContextID is the context ID associated with the task.
	ContextID string
}

// NewFinalEvent creates a new final event.
func NewFinalEvent(taskID, contextID string) *FinalEvent {
	return &FinalEvent{
		BaseEvent: BaseEvent{
			EventType: EventTypeFinal,
			Timestamp: time.Now(),
		},
		TaskID:    taskID,
		ContextID: contextID,
	}
}

// Helper functions for creating common events.

// CreateWorkingStatusEvent creates a task status event indicating the task is working.
func CreateWorkingStatusEvent(taskID, contextID string) *TaskStatusUpdateEvent {
	return NewTaskStatusUpdateEvent(taskID, contextID, a2a.TaskStateWorking, "Task is being processed")
}

// CreateCompletedStatusEvent creates a task status event indicating the task is completed.
func CreateCompletedStatusEvent(taskID, contextID string) *TaskStatusUpdateEvent {
	return NewTaskStatusUpdateEvent(taskID, contextID, a2a.TaskStateCompleted, "Task completed successfully")
}

// CreateFailedStatusEvent creates a task status event indicating the task has failed.
func CreateFailedStatusEvent(taskID, contextID string, errorMessage string) *TaskStatusUpdateEvent {
	return NewTaskStatusUpdateEvent(taskID, contextID, a2a.TaskStateFailed, errorMessage)
}

// CreateInputRequiredStatusEvent creates a task status event indicating input is required.
func CreateInputRequiredStatusEvent(taskID, contextID string, prompt string) *TaskStatusUpdateEvent {
	return NewTaskStatusUpdateEvent(taskID, contextID, a2a.TaskStateInputRequired, prompt)
}