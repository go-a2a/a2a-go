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
	"context"
	"testing"
	"time"
)

func TestChannelEventQueue_PutGet(t *testing.T) {
	tests := map[string]struct {
		capacity int
		events   []Event
		wantErr  bool
	}{
		"success: single event": {
			capacity: 10,
			events: []Event{
				CreateWorkingStatusEvent("task-1", "ctx-1"),
			},
			wantErr: false,
		},
		"success: multiple events": {
			capacity: 10,
			events: []Event{
				CreateWorkingStatusEvent("task-1", "ctx-1"),
				NewTaskArtifactUpdateEvent("task-1", "ctx-1", NewTextArtifact("test", "content")),
				CreateCompletedStatusEvent("task-1", "ctx-1"),
			},
			wantErr: false,
		},
		"success: unbuffered queue": {
			capacity: 0,
			events: []Event{
				CreateWorkingStatusEvent("task-1", "ctx-1"),
			},
			wantErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			queue := NewChannelEventQueue(tc.capacity)
			defer queue.Close()

			// Put events
			for _, event := range tc.events {
				if err := queue.Put(ctx, event); err != nil && !tc.wantErr {
					t.Errorf("Put() error = %v", err)
				}
			}

			// Get events
			for i, expectedEvent := range tc.events {
				gotEvent, err := queue.Get(ctx)
				if err != nil {
					t.Fatalf("Get() error = %v", err)
				}

				// Compare event types
				if gotEvent.GetEventType() != expectedEvent.GetEventType() {
					t.Errorf("Event[%d] type = %q, want %q", i, gotEvent.GetEventType(), expectedEvent.GetEventType())
				}
			}
		})
	}
}

func TestChannelEventQueue_Close(t *testing.T) {
	ctx := t.Context()
	queue := NewChannelEventQueue(10)

	// Add an event
	event := CreateWorkingStatusEvent("task-1", "ctx-1")
	if err := queue.Put(ctx, event); err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	// Close the queue
	if err := queue.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Verify Done channel is closed
	select {
	case <-queue.Done():
		// Expected
	default:
		t.Error("Done channel should be closed")
	}

	// Try to put after close
	if err := queue.Put(ctx, event); err == nil {
		t.Error("Put() after Close() should return error")
	}

	// Try to get after close - in our implementation, Close() closes the channel
	// so Get() will return an error
	if _, err := queue.Get(ctx); err == nil {
		t.Error("Get() after Close() should return error")
	}
}

func TestChannelEventQueue_ContextCancellation(t *testing.T) {
	queue := NewChannelEventQueue(0) // Unbuffered
	defer queue.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// Try to put with cancelled context
	cancel()
	event := CreateWorkingStatusEvent("task-1", "ctx-1")
	if err := queue.Put(ctx, event); err == nil {
		t.Error("Put() with cancelled context should return error")
	}

	// Try to get with cancelled context
	if _, err := queue.Get(ctx); err == nil {
		t.Error("Get() with cancelled context should return error")
	}
}

func TestChannelEventQueue_Len(t *testing.T) {
	queue := NewChannelEventQueue(10)
	defer queue.Close()

	ctx := t.Context()

	// Initially empty
	if queue.Len() != 0 {
		t.Errorf("Len() = %d, want 0", queue.Len())
	}

	// Add events
	events := []Event{
		CreateWorkingStatusEvent("task-1", "ctx-1"),
		CreateCompletedStatusEvent("task-1", "ctx-1"),
	}

	for _, event := range events {
		if err := queue.Put(ctx, event); err != nil {
			t.Fatalf("Put() error = %v", err)
		}
	}

	// Check length
	if queue.Len() != len(events) {
		t.Errorf("Len() = %d, want %d", queue.Len(), len(events))
	}

	// Get one event
	if _, err := queue.Get(ctx); err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Check length decreased
	if queue.Len() != len(events)-1 {
		t.Errorf("Len() after Get() = %d, want %d", queue.Len(), len(events)-1)
	}
}

func TestChannelEventQueue_Concurrent(t *testing.T) {
	queue := NewChannelEventQueue(100)
	defer queue.Close()

	ctx := t.Context()
	numProducers := 5
	numEvents := 10

	// Start producers
	for i := 0; i < numProducers; i++ {
		go func(producerID int) {
			for j := 0; j < numEvents; j++ {
				event := CreateWorkingStatusEvent(
					t.Name(),
					t.Name(),
				)
				if err := queue.Put(ctx, event); err != nil {
					t.Errorf("Producer %d: Put() error = %v", producerID, err)
				}
			}
		}(i)
	}

	// Consume events
	totalExpected := numProducers * numEvents
	consumed := 0
	timeout := time.After(2 * time.Second)

	for consumed < totalExpected {
		select {
		case <-timeout:
			t.Fatalf("Timeout: consumed %d events, expected %d", consumed, totalExpected)
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			if _, err := queue.Get(ctx); err == nil {
				consumed++
			}
			cancel()
		}
	}

	if consumed != totalExpected {
		t.Errorf("Consumed %d events, expected %d", consumed, totalExpected)
	}
}

func BenchmarkChannelEventQueue_Put(b *testing.B) {
	queue := NewChannelEventQueue(1000)
	defer queue.Close()

	ctx := context.Background()
	event := CreateWorkingStatusEvent("bench-task", "bench-ctx")

	b.ResetTimer()
	for range b.N {
		if err := queue.Put(ctx, event); err != nil {
			b.Fatalf("Put() error = %v", err)
		}
	}
}

func BenchmarkChannelEventQueue_Get(b *testing.B) {
	queue := NewChannelEventQueue(1000)
	defer queue.Close()

	ctx := context.Background()
	event := CreateWorkingStatusEvent("bench-task", "bench-ctx")

	// Pre-fill queue
	for i := 0; i < 1000; i++ {
		queue.Put(ctx, event)
	}

	b.ResetTimer()
	for range b.N {
		if _, err := queue.Get(ctx); err != nil {
			// Refill if empty
			queue.Put(ctx, event)
		}
	}
}