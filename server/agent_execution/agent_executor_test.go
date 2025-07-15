// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package agent_execution

import (
	"context"
	"testing"
	"time"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/server"
	"github.com/go-a2a/a2a/server/event"
)

func TestDummyAgentExecutor_Execute(t *testing.T) {
	executor := NewDummyAgentExecutor()
	eventQueue := event.NewEventQueue(10)
	defer eventQueue.Close()

	// Create a test request context
	params := &a2a.MessageSendParams{
		Message: &a2a.Message{
			Content: "test message",
		},
	}
	callContext := server.NewServerCallContext(nil)
	requestContext := NewRequestContext(params, "test-task", "test-context", nil, callContext)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test execution
	err := executor.Execute(ctx, requestContext, eventQueue)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify events were published
	event1, err := eventQueue.DequeueEvent(ctx, true)
	if err != nil {
		t.Fatalf("Failed to dequeue first event: %v", err)
	}

	if event1.EventType() != "task_status_update" {
		t.Errorf("Expected task_status_update event, got %s", event1.EventType())
	}

	event2, err := eventQueue.DequeueEvent(ctx, true)
	if err != nil {
		t.Fatalf("Failed to dequeue second event: %v", err)
	}

	if event2.EventType() != "task_status_update" {
		t.Errorf("Expected task_status_update event, got %s", event2.EventType())
	}
}

func TestDummyAgentExecutor_Cancel(t *testing.T) {
	executor := NewDummyAgentExecutor()
	eventQueue := event.NewEventQueue(10)
	defer eventQueue.Close()

	// Create a test request context
	params := &a2a.MessageSendParams{
		Message: &a2a.Message{
			Content: "test message",
		},
	}
	callContext := server.NewServerCallContext(nil)
	requestContext := NewRequestContext(params, "test-task", "test-context", nil, callContext)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test cancellation
	err := executor.Cancel(ctx, requestContext, eventQueue)
	if err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}

	// Verify cancellation event was published
	ev, err := eventQueue.DequeueEvent(ctx, true)
	if err != nil {
		t.Fatalf("Failed to dequeue event: %v", err)
	}

	if ev.EventType() != "task_status_update" {
		t.Errorf("Expected task_status_update event, got %s", ev.EventType())
	}

	// Check that it's a cancellation event
	if statusUpdateEvent, ok := ev.(*event.TaskStatusUpdateEvent); ok {
		if statusUpdateEvent.Status.State != a2a.TaskStateCanceled {
			t.Errorf("Expected canceled state, got %s", statusUpdateEvent.Status.State)
		}
	} else {
		t.Error("Event is not a TaskStatusUpdate event")
	}
}

func TestDummyAgentExecutor_ErrorHandling(t *testing.T) {
	executor := NewDummyAgentExecutor()
	ctx := context.Background()

	// Test with nil request context
	err := executor.Execute(ctx, nil, event.NewEventQueue(10))
	if err == nil {
		t.Error("Expected error with nil request context")
	}

	// Test with nil event queue
	params := &a2a.MessageSendParams{
		Message: &a2a.Message{
			Content: "test message",
		},
	}
	callContext := server.NewServerCallContext(nil)
	requestContext := NewRequestContext(params, "test-task", "test-context", nil, callContext)

	err = executor.Execute(ctx, requestContext, nil)
	if err == nil {
		t.Error("Expected error with nil event queue")
	}
}
