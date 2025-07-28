// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"testing"

	a2a "github.com/go-a2a/a2a-go"
)

func TestMessageEvent(t *testing.T) {
	message := &a2a.Message{
		Content:   "test message",
		ContextID: "test-context",
	}

	event := NewMessageEvent(message)

	if event.EventType() != "message" {
		t.Errorf("Expected event type 'message', got %s", event.EventType())
	}

	if event.EventData() != message {
		t.Errorf("Expected event data to be the message")
	}

	if err := event.Validate(); err != nil {
		t.Errorf("Expected valid event, got error: %v", err)
	}

	if !IsFinalEvent(event) {
		t.Errorf("Expected MessageEvent to be final")
	}
}

func TestMessageEventInvalid(t *testing.T) {
	event := &MessageEvent{Message: nil}

	if err := event.Validate(); err == nil {
		t.Errorf("Expected validation error for nil message")
	}

	event = &MessageEvent{Message: &a2a.Message{}}
	if err := event.Validate(); err == nil {
		t.Errorf("Expected validation error for empty message content")
	}
}

func TestTaskEvent(t *testing.T) {
	task := &a2a.Task{
		ID:        "test-task",
		ContextID: "test-context",
		Status:    a2a.TaskStatus{State: a2a.TaskStateRunning},
		History:   []*a2a.Message{},
		Artifacts: []*a2a.Artifact{},
	}

	event := NewTaskEvent(task)

	if event.EventType() != "task" {
		t.Errorf("Expected event type 'task', got %s", event.EventType())
	}

	if event.EventData() != task {
		t.Errorf("Expected event data to be the task")
	}

	if err := event.Validate(); err != nil {
		t.Errorf("Expected valid event, got error: %v", err)
	}

	// Running task should not be final
	if IsFinalEvent(event) {
		t.Errorf("Expected running TaskEvent to not be final")
	}

	// Completed task should be final
	task.Status.State = a2a.TaskStateCompleted
	if !IsFinalEvent(event) {
		t.Errorf("Expected completed TaskEvent to be final")
	}
}

func TestTaskStatusUpdateEvent(t *testing.T) {
	event := NewTaskStatusUpdateEvent(
		"test-task",
		"test-context",
		a2a.TaskStatus{State: a2a.TaskStateRunning},
		false,
		map[string]any{"key": "value"},
	)

	if event.EventType() != "task_status_update" {
		t.Errorf("Expected event type 'task_status_update', got %s", event.EventType())
	}

	if err := event.Validate(); err != nil {
		t.Errorf("Expected valid event, got error: %v", err)
	}

	if IsFinalEvent(event) {
		t.Errorf("Expected non-final TaskStatusUpdateEvent to not be final")
	}

	// Test final event
	event.Final = true
	if !IsFinalEvent(event) {
		t.Errorf("Expected final TaskStatusUpdateEvent to be final")
	}
}

func TestTaskStatusUpdateEventInvalid(t *testing.T) {
	event := &TaskStatusUpdateEvent{
		TaskID: "",
		Status: a2a.TaskStatus{State: a2a.TaskStateRunning},
	}

	if err := event.Validate(); err == nil {
		t.Errorf("Expected validation error for empty task ID")
	}

	event = &TaskStatusUpdateEvent{
		TaskID: "test-task",
		Status: a2a.TaskStatus{State: "invalid"},
	}

	if err := event.Validate(); err == nil {
		t.Errorf("Expected validation error for invalid task state")
	}
}

func TestTaskArtifactUpdateEvent(t *testing.T) {
	// Create a proper artifact using the NewTextArtifact function
	artifact, err := a2a.NewTextArtifact("test", "test content", "test artifact")
	if err != nil {
		t.Fatalf("Failed to create artifact: %v", err)
	}

	event := NewTaskArtifactUpdateEvent("test-task", artifact, true)

	if event.EventType() != "task_artifact_update" {
		t.Errorf("Expected event type 'task_artifact_update', got %s", event.EventType())
	}

	if err := event.Validate(); err != nil {
		t.Errorf("Expected valid event, got error: %v", err)
	}

	if IsFinalEvent(event) {
		t.Errorf("Expected TaskArtifactUpdateEvent to not be final")
	}
}

func TestTaskArtifactUpdateEventInvalid(t *testing.T) {
	event := &TaskArtifactUpdateEvent{
		TaskID:   "",
		Artifact: &a2a.Artifact{},
	}

	if err := event.Validate(); err == nil {
		t.Errorf("Expected validation error for empty task ID")
	}

	event = &TaskArtifactUpdateEvent{
		TaskID:   "test-task",
		Artifact: nil,
	}

	if err := event.Validate(); err == nil {
		t.Errorf("Expected validation error for nil artifact")
	}
}

func TestIsTerminalTaskState(t *testing.T) {
	tests := []struct {
		state    a2a.TaskState
		terminal bool
	}{
		{a2a.TaskStateSubmitted, false},
		{a2a.TaskStateRunning, false},
		{a2a.TaskStateCompleted, true},
		{a2a.TaskStateFailed, true},
		{a2a.TaskStateCanceled, true},
	}

	for _, test := range tests {
		if IsTerminalTaskState(test.state) != test.terminal {
			t.Errorf("Expected IsTerminalTaskState(%s) to be %t", test.state, test.terminal)
		}
	}
}

func TestEventFromContext(t *testing.T) {
	message := &a2a.Message{
		Content:   "test message",
		ContextID: "test-context",
	}
	event := NewMessageEvent(message)

	// Test with event in context
	ctx := ContextWithEvent(context.Background(), event)
	retrievedEvent, ok := EventFromContext(ctx)
	if !ok {
		t.Errorf("Expected to retrieve event from context")
	}
	if retrievedEvent != event {
		t.Errorf("Expected retrieved event to match original")
	}

	// Test with nil context
	retrievedEvent, ok = EventFromContext(nil)
	if ok {
		t.Errorf("Expected no event from nil context")
	}
	if retrievedEvent != nil {
		t.Errorf("Expected nil event from nil context")
	}

	// Test with context without event
	ctx = context.Background()
	retrievedEvent, ok = EventFromContext(ctx)
	if ok {
		t.Errorf("Expected no event from context without event")
	}
	if retrievedEvent != nil {
		t.Errorf("Expected nil event from context without event")
	}
}

func TestContextWithEvent(t *testing.T) {
	message := &a2a.Message{
		Content:   "test message",
		ContextID: "test-context",
	}
	event := NewMessageEvent(message)

	// Test with nil context
	ctx := ContextWithEvent(nil, event)
	if ctx == nil {
		t.Errorf("Expected non-nil context")
	}

	retrievedEvent, ok := EventFromContext(ctx)
	if !ok {
		t.Errorf("Expected to retrieve event from context")
	}
	if retrievedEvent != event {
		t.Errorf("Expected retrieved event to match original")
	}

	// Test with existing context
	baseCtx := context.Background()
	ctx = ContextWithEvent(baseCtx, event)
	if ctx == nil {
		t.Errorf("Expected non-nil context")
	}

	retrievedEvent, ok = EventFromContext(ctx)
	if !ok {
		t.Errorf("Expected to retrieve event from context")
	}
	if retrievedEvent != event {
		t.Errorf("Expected retrieved event to match original")
	}
}
