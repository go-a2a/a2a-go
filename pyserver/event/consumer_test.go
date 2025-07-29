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

package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-a2a/a2a-go"
	"github.com/google/go-cmp/cmp"
)

func TestEventConsumer_ConsumeOne(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queue, err := NewEventQueue(10)
	if err != nil {
		t.Fatalf("NewEventQueue() error = %v", err)
	}
	defer queue.Close()

	consumer := NewEventConsumer(queue)

	// Test consume on empty queue - should return ErrQueueEmpty
	_, err = consumer.ConsumeOne(ctx)
	if !errors.Is(err, ErrQueueEmpty) {
		t.Errorf("ConsumeOne() on empty queue error = %v, want %v", err, ErrQueueEmpty)
	}

	// Add event and consume
	event := &a2a.Message{
		MessageID: "test-msg-1",
		TaskID:    "test-task-1",
		Role:      a2a.RoleAgent,
		Kind:      a2a.MessageEventKind,
	}
	if err := queue.EnqueueEvent(ctx, event); err != nil {
		t.Fatalf("EnqueueEvent() error = %v", err)
	}

	consumed, err := consumer.ConsumeOne(ctx)
	if err != nil {
		t.Errorf("ConsumeOne() error = %v", err)
	}

	if diff := cmp.Diff(event, consumed); diff != "" {
		t.Errorf("consumed event mismatch (-want +got):\n%s", diff)
	}
}

func TestEventConsumer_ConsumeAll(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queue, err := NewEventQueue(10)
	if err != nil {
		t.Fatalf("NewEventQueue() error = %v", err)
	}
	defer queue.Close()

	consumer := NewEventConsumer(queue)
	consumer.SetTimeout(100 * time.Millisecond) // Short timeout for testing

	// Create events
	events := []Event{
		&a2a.TaskStatusUpdateEvent{
			TaskID:    "test-task-1",
			ContextID: "test-context-1",
			Kind:      a2a.StatusUpdateEventKind,
			Status: a2a.TaskStatus{
				State: a2a.TaskStateWorking,
			},
			Final: false,
		},
		&a2a.TaskArtifactUpdateEvent{
			TaskID:    "test-task-1",
			ContextID: "test-context-1",
			Kind:      a2a.ArtifactUpdateEventKind,
			Artifact: &a2a.Artifact{
				ArtifactID: "artifact-1",
				Name:       "Test Artifact",
			},
		},
		&a2a.TaskStatusUpdateEvent{
			TaskID:    "test-task-1",
			ContextID: "test-context-1",
			Kind:      a2a.StatusUpdateEventKind,
			Status: a2a.TaskStatus{
				State: a2a.TaskStateCompleted,
			},
			Final: true, // Final event
		},
	}

	// Start consuming in background
	consumedEvents := make([]Event, 0)
	done := make(chan struct{})
	
	go func() {
		defer close(done)
		for event := range consumer.ConsumeAll(ctx) {
			consumedEvents = append(consumedEvents, event)
		}
	}()

	// Enqueue events
	for _, event := range events {
		if err := queue.EnqueueEvent(ctx, event); err != nil {
			t.Fatalf("EnqueueEvent() error = %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Small delay between events
	}

	// Wait for consumption to complete
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("ConsumeAll() timed out")
	}

	// Verify all events were consumed
	if len(consumedEvents) != len(events) {
		t.Errorf("consumed %d events, want %d", len(consumedEvents), len(events))
	}

	// Verify queue was closed after final event
	if !queue.IsClosed() {
		t.Error("queue should be closed after final event")
	}
}

func TestEventConsumer_FinalEventTypes(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		event    Event
		isFinal  bool
	}{
		"final: TaskStatusUpdateEvent with Final=true": {
			event: &a2a.TaskStatusUpdateEvent{
				TaskID: "test-task-1",
				Kind:   a2a.StatusUpdateEventKind,
				Final:  true,
			},
			isFinal: true,
		},
		"final: Message": {
			event: &a2a.Message{
				MessageID: "test-msg-1",
				TaskID:    "test-task-1",
				Kind:      a2a.MessageEventKind,
			},
			isFinal: true,
		},
		"final: Task with completed state": {
			event: &a2a.Task{
				ID:   "test-task-1",
				Kind: a2a.TaskEventKind,
				Status: a2a.TaskStatus{
					State: a2a.TaskStateCompleted,
				},
			},
			isFinal: true,
		},
		"final: Task with canceled state": {
			event: &a2a.Task{
				ID:   "test-task-1",
				Kind: a2a.TaskEventKind,
				Status: a2a.TaskStatus{
					State: a2a.TaskStateCanceled,
				},
			},
			isFinal: true,
		},
		"not final: TaskStatusUpdateEvent with Final=false": {
			event: &a2a.TaskStatusUpdateEvent{
				TaskID: "test-task-1",
				Kind:   a2a.StatusUpdateEventKind,
				Final:  false,
			},
			isFinal: false,
		},
		"not final: Task with working state": {
			event: &a2a.Task{
				ID:   "test-task-1",
				Kind: a2a.TaskEventKind,
				Status: a2a.TaskStatus{
					State: a2a.TaskStateWorking,
				},
			},
			isFinal: false,
		},
		"not final: TaskArtifactUpdateEvent": {
			event: &a2a.TaskArtifactUpdateEvent{
				TaskID: "test-task-1",
				Kind:   a2a.ArtifactUpdateEventKind,
			},
			isFinal: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsFinalEvent(tt.event)
			if got != tt.isFinal {
				t.Errorf("IsFinalEvent() = %v, want %v", got, tt.isFinal)
			}
		})
	}
}

func TestEventConsumer_AgentTaskError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queue, err := NewEventQueue(10)
	if err != nil {
		t.Fatalf("NewEventQueue() error = %v", err)
	}
	defer queue.Close()

	consumer := NewEventConsumer(queue)
	consumer.SetTimeout(100 * time.Millisecond)

	// Set agent task error
	agentErr := errors.New("agent task failed")
	consumer.SetAgentTaskError(agentErr)

	// Verify error is stored
	if err := consumer.GetError(); !errors.Is(err, agentErr) {
		t.Errorf("GetError() = %v, want %v", err, agentErr)
	}

	// Start consuming - should stop immediately due to error
	done := make(chan struct{})
	eventCount := 0
	
	go func() {
		defer close(done)
		for range consumer.ConsumeAll(ctx) {
			eventCount++
		}
	}()

	// Wait for consumption to stop
	select {
	case <-done:
		// Success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("ConsumeAll() did not stop after agent error")
	}

	// Should not have consumed any events
	if eventCount != 0 {
		t.Errorf("consumed %d events, want 0 when agent error is set", eventCount)
	}
}

func TestEventConsumer_ContextCancellation(t *testing.T) {
	t.Parallel()

	queue, err := NewEventQueue(10)
	if err != nil {
		t.Fatalf("NewEventQueue() error = %v", err)
	}
	defer queue.Close()

	consumer := NewEventConsumer(queue)
	consumer.SetTimeout(100 * time.Millisecond)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start consuming
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range consumer.ConsumeAll(ctx) {
			// Consume events
		}
	}()

	// Cancel context
	cancel()

	// Consumer should stop
	select {
	case <-done:
		// Success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("ConsumeAll() did not stop after context cancellation")
	}
}

func TestEventConsumer_MultipleConsumers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	parentQueue, err := NewEventQueue(10)
	if err != nil {
		t.Fatalf("NewEventQueue() error = %v", err)
	}
	defer parentQueue.Close()

	// Create child queues via tap
	childQueue1, _ := parentQueue.Tap()
	childQueue2, _ := parentQueue.Tap()

	// Create consumers
	consumer1 := NewEventConsumer(childQueue1)
	consumer2 := NewEventConsumer(childQueue2)
	consumer1.SetTimeout(100 * time.Millisecond)
	consumer2.SetTimeout(100 * time.Millisecond)

	// Collect events from both consumers
	events1 := make([]Event, 0)
	events2 := make([]Event, 0)
	done1 := make(chan struct{})
	done2 := make(chan struct{})

	go func() {
		defer close(done1)
		for event := range consumer1.ConsumeAll(ctx) {
			events1 = append(events1, event)
		}
	}()

	go func() {
		defer close(done2)
		for event := range consumer2.ConsumeAll(ctx) {
			events2 = append(events2, event)
		}
	}()

	// Send events to parent queue
	testEvents := []Event{
		&a2a.TaskStatusUpdateEvent{
			TaskID: "task-1",
			Kind:   a2a.StatusUpdateEventKind,
			Status: a2a.TaskStatus{State: a2a.TaskStateWorking},
		},
		&a2a.Message{
			MessageID: "msg-1",
			TaskID:    "task-1",
			Kind:      a2a.MessageEventKind,
		}, // Final event
	}

	for _, event := range testEvents {
		if err := parentQueue.EnqueueEvent(ctx, event); err != nil {
			t.Fatalf("EnqueueEvent() error = %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Wait for both consumers to finish
	select {
	case <-done1:
	case <-time.After(2 * time.Second):
		t.Fatal("consumer1 timed out")
	}

	select {
	case <-done2:
	case <-time.After(2 * time.Second):
		t.Fatal("consumer2 timed out")
	}

	// Both consumers should have received all events
	if len(events1) != len(testEvents) {
		t.Errorf("consumer1 received %d events, want %d", len(events1), len(testEvents))
	}
	if len(events2) != len(testEvents) {
		t.Errorf("consumer2 received %d events, want %d", len(events2), len(testEvents))
	}
}