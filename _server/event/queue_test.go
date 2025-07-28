// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"testing"
	"time"

	a2a "github.com/go-a2a/a2a-go"
)

func TestNewEventQueue(t *testing.T) {
	// Test with valid size
	queue := NewEventQueue(10)
	if queue == nil {
		t.Fatalf("Expected non-nil queue")
	}
	if queue.Cap() != 10 {
		t.Errorf("Expected capacity 10, got %d", queue.Cap())
	}
	if queue.Len() != 0 {
		t.Errorf("Expected length 0, got %d", queue.Len())
	}
	if queue.IsClosed() {
		t.Errorf("Expected queue to not be closed")
	}

	// Test with invalid size
	queue = NewEventQueue(-1)
	if queue.Cap() != DefaultMaxQueueSize {
		t.Errorf("Expected default capacity %d, got %d", DefaultMaxQueueSize, queue.Cap())
	}

	// Test with zero size
	queue = NewEventQueue(0)
	if queue.Cap() != DefaultMaxQueueSize {
		t.Errorf("Expected default capacity %d, got %d", DefaultMaxQueueSize, queue.Cap())
	}
}

func TestNewEventQueueWithName(t *testing.T) {
	name := "test-queue"
	queue := NewEventQueueWithName(name, 10)
	if queue.Name() != name {
		t.Errorf("Expected queue name '%s', got '%s'", name, queue.Name())
	}
}

func TestEventQueueEnqueueDequeue(t *testing.T) {
	queue := NewEventQueue(2)

	message := &a2a.Message{
		Content:   "test message",
		ContextID: "test-context",
	}
	event := NewMessageEvent(message)

	// Test enqueue
	err := queue.EnqueueEvent(t.Context(), event)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if queue.Len() != 1 {
		t.Errorf("Expected length 1, got %d", queue.Len())
	}

	// Test dequeue
	ctx := context.Background()
	dequeuedEvent, err := queue.DequeueEvent(ctx, true)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if dequeuedEvent != event {
		t.Errorf("Expected dequeued event to match original")
	}

	if queue.Len() != 0 {
		t.Errorf("Expected length 0, got %d", queue.Len())
	}
}

func TestEventQueueEnqueueNilEvent(t *testing.T) {
	queue := NewEventQueue(1)

	err := queue.EnqueueEvent(t.Context(), nil)
	if err == nil {
		t.Errorf("Expected error for nil event")
	}
}

func TestEventQueueEnqueueInvalidEvent(t *testing.T) {
	queue := NewEventQueue(1)

	// Create invalid event
	event := &MessageEvent{Message: nil}

	err := queue.EnqueueEvent(t.Context(), event)
	if err == nil {
		t.Errorf("Expected error for invalid event")
	}
}

func TestEventQueueFull(t *testing.T) {
	queue := NewEventQueue(1)

	message := &a2a.Message{
		Content:   "test message",
		ContextID: "test-context",
	}
	event := NewMessageEvent(message)

	// Fill the queue
	err := queue.EnqueueEvent(t.Context(), event)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Try to add another event
	err = queue.EnqueueEvent(t.Context(), event)
	if err == nil {
		t.Errorf("Expected error for full queue")
	}
	if !IsQueueFullError(err) {
		t.Errorf("Expected queue full error, got %v", err)
	}
}

func TestEventQueueDequeueEmpty(t *testing.T) {
	queue := NewEventQueue(1)

	ctx := context.Background()
	_, err := queue.DequeueEvent(ctx, true)
	if err == nil {
		t.Errorf("Expected error for empty queue")
	}
	if !IsQueueEmptyError(err) {
		t.Errorf("Expected queue empty error, got %v", err)
	}
}

func TestEventQueueDequeueBlocking(t *testing.T) {
	queue := NewEventQueue(1)

	message := &a2a.Message{
		Content:   "test message",
		ContextID: "test-context",
	}
	event := NewMessageEvent(message)

	// Start a goroutine to enqueue after delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		queue.EnqueueEvent(t.Context(), event)
	}()

	// Block until event is available
	ctx := context.Background()
	dequeuedEvent, err := queue.DequeueEvent(ctx, false)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if dequeuedEvent != event {
		t.Errorf("Expected dequeued event to match original")
	}
}

func TestEventQueueDequeueTimeout(t *testing.T) {
	queue := NewEventQueue(1)

	dequeuedEvent, err := queue.DequeueEventWithTimeout(100 * time.Millisecond)
	if err == nil {
		t.Errorf("Expected timeout error")
	}
	if dequeuedEvent != nil {
		t.Errorf("Expected nil event on timeout")
	}
}

func TestEventQueueDequeueContextCancelled(t *testing.T) {
	queue := NewEventQueue(1)

	ctx, cancel := context.WithCancel(context.Background())

	// Start a goroutine to cancel context after delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// Block until context is cancelled
	_, err := queue.DequeueEvent(ctx, false)
	if err == nil {
		t.Errorf("Expected context cancellation error")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

func TestEventQueueTap(t *testing.T) {
	queue := NewEventQueue(2)

	child := queue.Tap()
	if child == nil {
		t.Fatalf("Expected non-nil child queue")
	}

	if queue.ChildCount() != 1 {
		t.Errorf("Expected 1 child, got %d", queue.ChildCount())
	}

	message := &a2a.Message{
		Content:   "test message",
		ContextID: "test-context",
	}
	event := NewMessageEvent(message)

	// Enqueue to parent
	err := queue.EnqueueEvent(t.Context(), event)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Wait for child to receive event
	time.Sleep(50 * time.Millisecond)

	// Check parent queue
	ctx := context.Background()
	parentEvent, err := queue.DequeueEvent(ctx, true)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if parentEvent != event {
		t.Errorf("Expected parent event to match original")
	}

	// Check child queue
	childEvent, err := child.DequeueEvent(ctx, true)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if childEvent.EventType() != event.EventType() {
		t.Errorf("Expected child event type to match parent")
	}
}

func TestEventQueueClose(t *testing.T) {
	queue := NewEventQueue(1)

	if queue.IsClosed() {
		t.Errorf("Expected queue to not be closed initially")
	}

	err := queue.Close()
	if err != nil {
		t.Errorf("Expected no error closing queue, got %v", err)
	}

	if !queue.IsClosed() {
		t.Errorf("Expected queue to be closed")
	}

	// Test that enqueue fails after close
	message := &a2a.Message{
		Content:   "test message",
		ContextID: "test-context",
	}
	event := NewMessageEvent(message)

	err = queue.EnqueueEvent(t.Context(), event)
	if err == nil {
		t.Errorf("Expected error enqueueing to closed queue")
	}
	if !IsQueueClosedError(err) {
		t.Errorf("Expected queue closed error, got %v", err)
	}
}

func TestEventQueueCloseWithChildren(t *testing.T) {
	queue := NewEventQueue(1)
	child := queue.Tap()

	err := queue.Close()
	if err != nil {
		t.Errorf("Expected no error closing queue, got %v", err)
	}

	if !queue.IsClosed() {
		t.Errorf("Expected parent queue to be closed")
	}

	if !child.IsClosed() {
		t.Errorf("Expected child queue to be closed")
	}
}

func TestEventQueueTapAfterClose(t *testing.T) {
	queue := NewEventQueue(1)

	err := queue.Close()
	if err != nil {
		t.Errorf("Expected no error closing queue, got %v", err)
	}

	child := queue.Tap()
	if child == nil {
		t.Fatalf("Expected non-nil child queue")
	}

	if !child.IsClosed() {
		t.Errorf("Expected child queue to be closed")
	}
}

func TestEventQueueDequeueAfterClose(t *testing.T) {
	queue := NewEventQueue(1)

	message := &a2a.Message{
		Content:   "test message",
		ContextID: "test-context",
	}
	event := NewMessageEvent(message)

	// Enqueue before close
	err := queue.EnqueueEvent(t.Context(), event)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Close queue
	err = queue.Close()
	if err != nil {
		t.Errorf("Expected no error closing queue, got %v", err)
	}

	// Should still be able to dequeue existing events
	ctx := context.Background()
	dequeuedEvent, err := queue.DequeueEvent(ctx, true)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if dequeuedEvent != event {
		t.Errorf("Expected dequeued event to match original")
	}

	// Next dequeue should fail
	_, err = queue.DequeueEvent(ctx, true)
	if err == nil {
		t.Errorf("Expected error dequeuing from closed empty queue")
	}
	if !IsQueueClosedError(err) {
		t.Errorf("Expected queue closed error, got %v", err)
	}
}

func TestEventQueueDrainEvents(t *testing.T) {
	queue := NewEventQueue(3)

	message := &a2a.Message{
		Content:   "test message",
		ContextID: "test-context",
	}
	event := NewMessageEvent(message)

	// Enqueue multiple events
	for range 3 {
		err := queue.EnqueueEvent(t.Context(), event)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	}

	// Drain events
	events := queue.DrainEvents()
	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}

	if queue.Len() != 0 {
		t.Errorf("Expected queue to be empty after drain, got length %d", queue.Len())
	}
}

func TestEventQueuePeek(t *testing.T) {
	queue := NewEventQueue(1)

	// Peek empty queue
	peekedEvent := queue.Peek()
	if peekedEvent != nil {
		t.Errorf("Expected nil event when peeking empty queue")
	}

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

	// Peek should return the event without removing it
	peekedEvent = queue.Peek()
	if peekedEvent != event {
		t.Errorf("Expected peeked event to match original")
	}

	if queue.Len() != 1 {
		t.Errorf("Expected queue length to remain 1 after peek, got %d", queue.Len())
	}
}

func TestEventQueueString(t *testing.T) {
	queue := NewEventQueueWithName("test-queue", 10)

	str := queue.String()
	if str == "" {
		t.Errorf("Expected non-empty string representation")
	}

	// String should contain key information
	if !contains(str, "test-queue") {
		t.Errorf("Expected string to contain queue name")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && containsMiddle(s, substr)
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
