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
	"sync"
	"testing"
	"time"

	"github.com/go-a2a/a2a-go"
	"github.com/google/go-cmp/cmp"
)

func TestNewEventQueue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		maxSize     int
		wantMaxSize int
		wantErr     error
	}{
		"success: default size": {
			maxSize:     0,
			wantMaxSize: DefaultMaxQueueSize,
			wantErr:     nil,
		},
		"success: custom size": {
			maxSize:     100,
			wantMaxSize: 100,
			wantErr:     nil,
		},
		"error: negative size": {
			maxSize: -1,
			wantErr: ErrInvalidQueueSize,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			queue, err := NewEventQueue(tt.maxSize)
			
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewEventQueue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if err == nil {
				if queue.maxSize != tt.wantMaxSize {
					t.Errorf("queue.maxSize = %v, want %v", queue.maxSize, tt.wantMaxSize)
				}
				if queue.Size() != 0 {
					t.Errorf("new queue should be empty, got size %d", queue.Size())
				}
			}
		})
	}
}

func TestEventQueue_EnqueueDequeue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queue, err := NewEventQueue(10)
	if err != nil {
		t.Fatalf("NewEventQueue() error = %v", err)
	}
	defer queue.Close()

	// Test enqueue
	event := &a2a.Message{
		MessageID: "test-msg-1",
		TaskID:    "test-task-1",
		Role:      a2a.RoleAgent,
		Kind:      a2a.MessageEventKind,
		Parts: []a2a.Part{
			a2a.NewTextPart("Hello, world!"),
		},
	}

	if err := queue.EnqueueEvent(ctx, event); err != nil {
		t.Errorf("EnqueueEvent() error = %v", err)
	}

	if queue.Size() != 1 {
		t.Errorf("queue.Size() = %d, want 1", queue.Size())
	}

	// Test dequeue (blocking)
	dequeued, err := queue.DequeueEvent(ctx, false)
	if err != nil {
		t.Errorf("DequeueEvent() error = %v", err)
	}

	if diff := cmp.Diff(event, dequeued); diff != "" {
		t.Errorf("dequeued event mismatch (-want +got):\n%s", diff)
	}

	if queue.Size() != 0 {
		t.Errorf("queue.Size() = %d after dequeue, want 0", queue.Size())
	}
}

func TestEventQueue_NoWaitDequeue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queue, err := NewEventQueue(10)
	if err != nil {
		t.Fatalf("NewEventQueue() error = %v", err)
	}
	defer queue.Close()

	// Test no-wait dequeue on empty queue
	_, err = queue.DequeueEvent(ctx, true)
	if !errors.Is(err, ErrQueueEmpty) {
		t.Errorf("DequeueEvent(noWait=true) error = %v, want %v", err, ErrQueueEmpty)
	}

	// Add event and test no-wait dequeue
	event := &a2a.TaskStatusUpdateEvent{
		TaskID:    "test-task-1",
		ContextID: "test-context-1",
		Kind:      a2a.StatusUpdateEventKind,
		Status: a2a.TaskStatus{
			State: a2a.TaskStateWorking,
		},
	}

	if err := queue.EnqueueEvent(ctx, event); err != nil {
		t.Fatalf("EnqueueEvent() error = %v", err)
	}

	dequeued, err := queue.DequeueEvent(ctx, true)
	if err != nil {
		t.Errorf("DequeueEvent(noWait=true) error = %v", err)
	}

	if diff := cmp.Diff(event, dequeued); diff != "" {
		t.Errorf("dequeued event mismatch (-want +got):\n%s", diff)
	}
}

func TestEventQueue_Close(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queue, err := NewEventQueue(10)
	if err != nil {
		t.Fatalf("NewEventQueue() error = %v", err)
	}

	// Add an event
	event := &a2a.Message{
		MessageID: "test-msg-1",
		TaskID:    "test-task-1",
		Role:      a2a.RoleAgent,
		Kind:      a2a.MessageEventKind,
	}
	if err := queue.EnqueueEvent(ctx, event); err != nil {
		t.Fatalf("EnqueueEvent() error = %v", err)
	}

	// Close the queue
	if err := queue.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if !queue.IsClosed() {
		t.Error("queue.IsClosed() = false, want true")
	}

	// Enqueue should fail after close
	if err := queue.EnqueueEvent(ctx, event); !errors.Is(err, ErrQueueClosed) {
		t.Errorf("EnqueueEvent() after close error = %v, want %v", err, ErrQueueClosed)
	}

	// Can still dequeue existing events
	dequeued, err := queue.DequeueEvent(ctx, true)
	if err != nil {
		t.Errorf("DequeueEvent() error = %v", err)
	}
	if dequeued == nil {
		t.Error("should be able to dequeue existing events after close")
	}

	// After all events consumed, dequeue should return ErrQueueClosed
	_, err = queue.DequeueEvent(ctx, true)
	if !errors.Is(err, ErrQueueClosed) {
		t.Errorf("DequeueEvent() on closed empty queue error = %v, want %v", err, ErrQueueClosed)
	}
}

func TestEventQueue_Tap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	parent, err := NewEventQueue(10)
	if err != nil {
		t.Fatalf("NewEventQueue() error = %v", err)
	}
	defer parent.Close()

	// Create child queue via tap
	child, err := parent.Tap()
	if err != nil {
		t.Fatalf("Tap() error = %v", err)
	}
	defer child.Close()

	// Enqueue event to parent
	event := &a2a.TaskArtifactUpdateEvent{
		TaskID:    "test-task-1",
		ContextID: "test-context-1",
		Kind:      a2a.ArtifactUpdateEventKind,
		Artifact: &a2a.Artifact{
			ArtifactID: "artifact-1",
			Name:       "Test Artifact",
		},
	}

	if err := parent.EnqueueEvent(ctx, event); err != nil {
		t.Fatalf("parent.EnqueueEvent() error = %v", err)
	}

	// Give some time for async propagation to child
	time.Sleep(10 * time.Millisecond)

	// Both parent and child should have the event
	if parent.Size() != 1 {
		t.Errorf("parent.Size() = %d, want 1", parent.Size())
	}
	if child.Size() != 1 {
		t.Errorf("child.Size() = %d, want 1", child.Size())
	}

	// Dequeue from both
	parentEvent, err := parent.DequeueEvent(ctx, true)
	if err != nil {
		t.Errorf("parent.DequeueEvent() error = %v", err)
	}

	childEvent, err := child.DequeueEvent(ctx, true)
	if err != nil {
		t.Errorf("child.DequeueEvent() error = %v", err)
	}

	// Both should have the same event
	if diff := cmp.Diff(parentEvent, childEvent); diff != "" {
		t.Errorf("parent and child events differ (-parent +child):\n%s", diff)
	}
}

func TestEventQueue_MultipleTaps(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	parent, err := NewEventQueue(10)
	if err != nil {
		t.Fatalf("NewEventQueue() error = %v", err)
	}
	defer parent.Close()

	// Create multiple child queues
	const numChildren = 3
	children := make([]*EventQueue, numChildren)
	for i := 0; i < numChildren; i++ {
		child, err := parent.Tap()
		if err != nil {
			t.Fatalf("Tap() error = %v", err)
		}
		defer child.Close()
		children[i] = child
	}

	// Enqueue event to parent
	event := &a2a.Task{
		ID:        "test-task-1",
		ContextID: "test-context-1",
		Kind:      a2a.TaskEventKind,
		Status: a2a.TaskStatus{
			State: a2a.TaskStateCompleted,
		},
	}

	if err := parent.EnqueueEvent(ctx, event); err != nil {
		t.Fatalf("parent.EnqueueEvent() error = %v", err)
	}

	// Give some time for async propagation
	time.Sleep(20 * time.Millisecond)

	// All children should have the event
	for i, child := range children {
		if child.Size() != 1 {
			t.Errorf("child[%d].Size() = %d, want 1", i, child.Size())
		}
	}
}

func TestEventQueue_ClosePropagatesToChildren(t *testing.T) {
	t.Parallel()

	parent, err := NewEventQueue(10)
	if err != nil {
		t.Fatalf("NewEventQueue() error = %v", err)
	}

	// Create child queues
	child1, _ := parent.Tap()
	child2, _ := parent.Tap()

	// Close parent
	if err := parent.Close(); err != nil {
		t.Errorf("parent.Close() error = %v", err)
	}

	// Children should also be closed
	if !child1.IsClosed() {
		t.Error("child1.IsClosed() = false, want true")
	}
	if !child2.IsClosed() {
		t.Error("child2.IsClosed() = false, want true")
	}
}

func TestEventQueue_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queue, err := NewEventQueue(100)
	if err != nil {
		t.Fatalf("NewEventQueue() error = %v", err)
	}
	defer queue.Close()

	const numProducers = 5
	const numConsumers = 3
	const eventsPerProducer = 10

	var wg sync.WaitGroup

	// Producers
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			for j := 0; j < eventsPerProducer; j++ {
				event := &a2a.Message{
					MessageID: "msg-" + string(rune(producerID)) + "-" + string(rune(j)),
					TaskID:    "task-1",
					Role:      a2a.RoleAgent,
					Kind:      a2a.MessageEventKind,
				}
				if err := queue.EnqueueEvent(ctx, event); err != nil {
					t.Errorf("producer %d: EnqueueEvent() error = %v", producerID, err)
				}
			}
		}(i)
	}

	// Consumers
	consumed := make([]Event, 0, numProducers*eventsPerProducer)
	var consumeMu sync.Mutex

	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go func(consumerID int) {
			defer wg.Done()
			for {
				event, err := queue.DequeueEvent(ctx, true)
				if err != nil {
					if errors.Is(err, ErrQueueEmpty) {
						// Expected when queue is empty
						time.Sleep(time.Millisecond)
						continue
					}
					t.Errorf("consumer %d: DequeueEvent() error = %v", consumerID, err)
					return
				}
				
				consumeMu.Lock()
				consumed = append(consumed, event)
				if len(consumed) >= numProducers*eventsPerProducer {
					consumeMu.Unlock()
					return
				}
				consumeMu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Verify all events were consumed
	if len(consumed) != numProducers*eventsPerProducer {
		t.Errorf("consumed %d events, want %d", len(consumed), numProducers*eventsPerProducer)
	}
}