// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-a2a/a2a"
)

func TestNewEventConsumer(t *testing.T) {
	queue := NewEventQueue(10)
	consumer := NewEventConsumer(queue)

	if consumer == nil {
		t.Fatalf("Expected non-nil consumer")
	}

	if consumer.Queue() != queue {
		t.Errorf("Expected consumer queue to match original")
	}

	if consumer.IsClosed() {
		t.Errorf("Expected consumer to not be closed initially")
	}

	if consumer.HasAgentTaskError() {
		t.Errorf("Expected consumer to not have agent task error initially")
	}
}

func TestNewEventConsumerWithName(t *testing.T) {
	queue := NewEventQueue(10)
	name := "test-consumer"
	consumer := NewEventConsumerWithName(queue, name)

	if consumer.Name() != name {
		t.Errorf("Expected consumer name '%s', got '%s'", name, consumer.Name())
	}
}

func TestEventConsumerConsumeOne(t *testing.T) {
	queue := NewEventQueue(1)
	consumer := NewEventConsumer(queue)

	message := &a2a.Message{
		Content:   "test message",
		ContextID: "test-context",
	}
	event := NewMessageEvent(message)

	// Enqueue event
	err := queue.EnqueueEvent(t.Context(), event)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Consume event
	ctx := context.Background()
	consumedEvent, err := consumer.ConsumeOne(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if consumedEvent != event {
		t.Errorf("Expected consumed event to match original")
	}
}

func TestEventConsumerConsumeOneEmpty(t *testing.T) {
	queue := NewEventQueue(1)
	consumer := NewEventConsumer(queue)

	ctx := context.Background()
	_, err := consumer.ConsumeOne(ctx)
	if err == nil {
		t.Errorf("Expected error consuming from empty queue")
	}
	if !IsQueueEmptyError(err) {
		t.Errorf("Expected queue empty error, got %v", err)
	}
}

func TestEventConsumerConsumeOneClosed(t *testing.T) {
	queue := NewEventQueue(1)
	consumer := NewEventConsumer(queue)

	err := consumer.Close()
	if err != nil {
		t.Errorf("Expected no error closing consumer, got %v", err)
	}

	ctx := context.Background()
	_, err = consumer.ConsumeOne(ctx)
	if err == nil {
		t.Errorf("Expected error consuming from closed consumer")
	}
	if !IsConsumerClosedError(err) {
		t.Errorf("Expected consumer closed error, got %v", err)
	}
}

func TestEventConsumerConsumeOneWithAgentTaskError(t *testing.T) {
	queue := NewEventQueue(1)
	consumer := NewEventConsumer(queue)

	agentError := errors.New("agent task failed")
	consumer.AgentTaskCallback(agentError)

	ctx := context.Background()
	_, err := consumer.ConsumeOne(ctx)
	if err == nil {
		t.Errorf("Expected error due to agent task error")
	}
	if !IsServerError(err) {
		t.Errorf("Expected server error, got %v", err)
	}
}

func TestEventConsumerConsumeAll(t *testing.T) {
	queue := NewEventQueue(3)
	consumer := NewEventConsumer(queue)

	// Create non-final events (task status updates)
	event1 := NewTaskStatusUpdateEvent("task1", "context1", a2a.TaskStatus{State: a2a.TaskStateRunning}, false, nil)
	event2 := NewTaskStatusUpdateEvent("task2", "context2", a2a.TaskStatus{State: a2a.TaskStateRunning}, false, nil)
	// Create final event
	finalEvent := NewTaskStatusUpdateEvent("task3", "context3", a2a.TaskStatus{State: a2a.TaskStateCompleted}, true, nil)

	// Enqueue events
	queue.EnqueueEvent(t.Context(), event1)
	queue.EnqueueEvent(t.Context(), event2)
	queue.EnqueueEvent(t.Context(), finalEvent)

	// Consume all events
	ctx := context.Background()
	eventChan, errorChan := consumer.ConsumeAll(ctx)

	var consumedEvents []Event
	var consumedErrors []error

	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				goto done
			}
			consumedEvents = append(consumedEvents, event)

		case err, ok := <-errorChan:
			if !ok {
				continue
			}
			if err != nil {
				consumedErrors = append(consumedErrors, err)
			}
		}
	}

done:
	if len(consumedEvents) != 3 {
		t.Errorf("Expected 3 consumed events, got %d", len(consumedEvents))
	}

	if len(consumedErrors) != 0 {
		t.Errorf("Expected no errors, got %v", consumedErrors)
	}

	// Check that the last event caused consumption to stop (since messages are final)
	if consumedEvents[0] != event1 {
		t.Errorf("Expected first event to match")
	}
}

func TestEventConsumerConsumeAllWithCallback(t *testing.T) {
	queue := NewEventQueue(3)
	consumer := NewEventConsumer(queue)

	// Create non-final events
	event1 := NewTaskStatusUpdateEvent("task1", "context1", a2a.TaskStatus{State: a2a.TaskStateRunning}, false, nil)
	event2 := NewTaskStatusUpdateEvent("task2", "context2", a2a.TaskStatus{State: a2a.TaskStateRunning}, false, nil)
	// Create final event to ensure consumption stops
	finalEvent := NewTaskStatusUpdateEvent("task3", "context3", a2a.TaskStatus{State: a2a.TaskStateCompleted}, true, nil)

	// Enqueue events
	queue.EnqueueEvent(t.Context(), event1)
	queue.EnqueueEvent(t.Context(), event2)
	queue.EnqueueEvent(t.Context(), finalEvent)

	// Consume with callback
	ctx := context.Background()
	var consumedEvents []Event

	err := consumer.ConsumeAllWithCallback(ctx, func(event Event) bool {
		consumedEvents = append(consumedEvents, event)
		return len(consumedEvents) < 3 // Stop after 3 events
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(consumedEvents) != 3 {
		t.Errorf("Expected 3 consumed events, got %d", len(consumedEvents))
	}
}

func TestEventConsumerConsumeUntil(t *testing.T) {
	queue := NewEventQueue(3)
	consumer := NewEventConsumer(queue)

	// Create non-final events
	task1 := &a2a.Task{
		ID:        "task1",
		ContextID: "test-context",
		Status:    a2a.TaskStatus{State: a2a.TaskStateRunning},
		History:   []*a2a.Message{},
		Artifacts: []*a2a.Artifact{},
	}

	task2 := &a2a.Task{
		ID:        "task2",
		ContextID: "test-context",
		Status:    a2a.TaskStatus{State: a2a.TaskStateRunning},
		History:   []*a2a.Message{},
		Artifacts: []*a2a.Artifact{},
	}

	// Create final event
	finalTask := &a2a.Task{
		ID:        "final-task",
		ContextID: "test-context",
		Status:    a2a.TaskStatus{State: a2a.TaskStateCompleted},
		History:   []*a2a.Message{},
		Artifacts: []*a2a.Artifact{},
	}

	event1 := NewTaskEvent(task1)
	event2 := NewTaskEvent(task2)
	finalEvent := NewTaskEvent(finalTask)

	// Enqueue events
	queue.EnqueueEvent(t.Context(), event1)
	queue.EnqueueEvent(t.Context(), event2)
	queue.EnqueueEvent(t.Context(), finalEvent)

	// Consume until final event
	ctx := context.Background()
	consumedEvents, err := consumer.ConsumeUntil(ctx, IsFinalEvent)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(consumedEvents) != 3 {
		t.Errorf("Expected 3 consumed events, got %d", len(consumedEvents))
	}

	if !IsFinalEvent(consumedEvents[len(consumedEvents)-1]) {
		t.Errorf("Expected last event to be final")
	}
}

func TestEventConsumerConsumeWithTimeout(t *testing.T) {
	queue := NewEventQueue(1)
	consumer := NewEventConsumer(queue)

	message := &a2a.Message{Content: "test message", ContextID: "test-context"}
	event := NewMessageEvent(message)

	// Enqueue event
	queue.EnqueueEvent(t.Context(), event)

	// Consume with timeout
	consumedEvents, err := consumer.ConsumeWithTimeout(100 * time.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(consumedEvents) != 1 {
		t.Errorf("Expected 1 consumed event, got %d", len(consumedEvents))
	}

	if consumedEvents[0] != event {
		t.Errorf("Expected consumed event to match original")
	}
}

func TestEventConsumerAgentTaskCallback(t *testing.T) {
	queue := NewEventQueue(1)
	consumer := NewEventConsumer(queue)

	if consumer.HasAgentTaskError() {
		t.Errorf("Expected no agent task error initially")
	}

	agentError := errors.New("agent task failed")
	consumer.AgentTaskCallback(agentError)

	if !consumer.HasAgentTaskError() {
		t.Errorf("Expected agent task error after callback")
	}

	if consumer.GetAgentTaskError() != agentError {
		t.Errorf("Expected agent task error to match")
	}

	consumer.ClearAgentTaskError()

	if consumer.HasAgentTaskError() {
		t.Errorf("Expected no agent task error after clear")
	}
}

func TestEventConsumerClose(t *testing.T) {
	queue := NewEventQueue(1)
	consumer := NewEventConsumer(queue)

	if consumer.IsClosed() {
		t.Errorf("Expected consumer to not be closed initially")
	}

	err := consumer.Close()
	if err != nil {
		t.Errorf("Expected no error closing consumer, got %v", err)
	}

	if !consumer.IsClosed() {
		t.Errorf("Expected consumer to be closed")
	}

	// Closing again should not error
	err = consumer.Close()
	if err != nil {
		t.Errorf("Expected no error closing consumer again, got %v", err)
	}
}

func TestEventConsumerWaitForEvent(t *testing.T) {
	queue := NewEventQueue(1)
	consumer := NewEventConsumer(queue)

	message := &a2a.Message{Content: "test message", ContextID: "test-context"}
	event := NewMessageEvent(message)

	// Enqueue event
	queue.EnqueueEvent(t.Context(), event)

	// Wait for event
	ctx := context.Background()
	hasEvent := consumer.WaitForEvent(ctx)

	if !hasEvent {
		t.Errorf("Expected event to be available")
	}

	// Test with closed consumer
	consumer.Close()
	hasEvent = consumer.WaitForEvent(ctx)

	if hasEvent {
		t.Errorf("Expected no event to be available from closed consumer")
	}
}

func TestEventConsumerPeek(t *testing.T) {
	queue := NewEventQueue(1)
	consumer := NewEventConsumer(queue)

	// Peek empty consumer
	peekedEvent := consumer.Peek()
	if peekedEvent != nil {
		t.Errorf("Expected nil event when peeking empty consumer")
	}

	message := &a2a.Message{Content: "test message", ContextID: "test-context"}
	event := NewMessageEvent(message)

	// Enqueue event
	queue.EnqueueEvent(t.Context(), event)

	// Peek should return the event
	peekedEvent = consumer.Peek()
	if peekedEvent != event {
		t.Errorf("Expected peeked event to match original")
	}

	// Test with closed consumer
	consumer.Close()
	peekedEvent = consumer.Peek()
	if peekedEvent != nil {
		t.Errorf("Expected nil event when peeking closed consumer")
	}
}

func TestEventConsumerString(t *testing.T) {
	queue := NewEventQueueWithName("test-queue", 10)
	consumer := NewEventConsumerWithName(queue, "test-consumer")

	str := consumer.String()
	if str == "" {
		t.Errorf("Expected non-empty string representation")
	}
}

// Helper function to check if consumer closed error
func IsConsumerClosedError(err error) bool {
	return errors.Is(err, ErrConsumerClosed)
}
