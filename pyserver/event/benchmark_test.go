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
	"sync"
	"testing"

	"github.com/go-a2a/a2a-go"
)

func BenchmarkEventQueue_EnqueueDequeue(b *testing.B) {
	ctx := context.Background()
	queue, err := NewEventQueue(1000)
	if err != nil {
		b.Fatalf("NewEventQueue() error = %v", err)
	}
	defer queue.Close()

	event := &a2a.Message{
		MessageID: "bench-msg-1",
		TaskID:    "bench-task-1",
		Role:      a2a.RoleAgent,
		Kind:      a2a.MessageEventKind,
		Parts: []a2a.Part{
			a2a.NewTextPart("Benchmark message"),
		},
	}

	b.ResetTimer()
	for b.Loop() {
		if err := queue.EnqueueEvent(ctx, event); err != nil {
			b.Fatalf("EnqueueEvent() error = %v", err)
		}
		if _, err := queue.DequeueEvent(ctx, true); err != nil {
			b.Fatalf("DequeueEvent() error = %v", err)
		}
	}
}

func BenchmarkEventQueue_ConcurrentEnqueueDequeue(b *testing.B) {
	ctx := context.Background()
	queue, err := NewEventQueue(1000)
	if err != nil {
		b.Fatalf("NewEventQueue() error = %v", err)
	}
	defer queue.Close()

	event := &a2a.TaskStatusUpdateEvent{
		TaskID:    "bench-task-1",
		ContextID: "bench-context-1",
		Kind:      a2a.StatusUpdateEventKind,
		Status: a2a.TaskStatus{
			State: a2a.TaskStateWorking,
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := queue.EnqueueEvent(ctx, event); err != nil {
				continue // Queue might be full
			}
			queue.DequeueEvent(ctx, true)
		}
	})
}

func BenchmarkEventQueue_Tap(b *testing.B) {
	ctx := context.Background()
	parent, err := NewEventQueue(1000)
	if err != nil {
		b.Fatalf("NewEventQueue() error = %v", err)
	}
	defer parent.Close()

	// Create some child queues
	children := make([]*EventQueue, 5)
	for i := 0; i < 5; i++ {
		child, err := parent.Tap()
		if err != nil {
			b.Fatalf("Tap() error = %v", err)
		}
		defer child.Close()
		children[i] = child
	}

	event := &a2a.TaskArtifactUpdateEvent{
		TaskID:    "bench-task-1",
		ContextID: "bench-context-1",
		Kind:      a2a.ArtifactUpdateEventKind,
		Artifact: &a2a.Artifact{
			ArtifactID: "artifact-1",
			Name:       "Benchmark Artifact",
		},
	}

	b.ResetTimer()
	for b.Loop() {
		if err := parent.EnqueueEvent(ctx, event); err != nil {
			b.Fatalf("EnqueueEvent() error = %v", err)
		}
		// Dequeue from parent to keep it from filling
		parent.DequeueEvent(ctx, true)
	}
}

func BenchmarkEventConsumer_ConsumeOne(b *testing.B) {
	ctx := context.Background()
	queue, err := NewEventQueue(10000)
	if err != nil {
		b.Fatalf("NewEventQueue() error = %v", err)
	}
	defer queue.Close()

	consumer := NewEventConsumer(queue)

	// Pre-fill queue
	event := &a2a.Message{
		MessageID: "bench-msg-1",
		TaskID:    "bench-task-1",
		Role:      a2a.RoleAgent,
		Kind:      a2a.MessageEventKind,
	}

	for i := 0; i < b.N; i++ {
		if err := queue.EnqueueEvent(ctx, event); err != nil {
			b.Fatalf("EnqueueEvent() error = %v", err)
		}
	}

	b.ResetTimer()
	for b.Loop() {
		if _, err := consumer.ConsumeOne(ctx); err != nil {
			b.Fatalf("ConsumeOne() error = %v", err)
		}
	}
}

func BenchmarkInMemoryQueueManager_Get(b *testing.B) {
	manager := NewInMemoryQueueManager(100)

	// Pre-create some queues
	for i := 0; i < 100; i++ {
		manager.Get("task-" + string(rune(i)))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			taskID := "task-" + string(rune(i%100))
			if _, err := manager.Get(taskID); err != nil {
				b.Fatalf("Get() error = %v", err)
			}
			i++
		}
	})
}

func BenchmarkInMemoryQueueManager_ConcurrentOperations(b *testing.B) {
	manager := NewInMemoryQueueManager(100)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			taskID := "task-" + string(rune(i%1000))
			
			// Mix of operations
			switch i % 4 {
			case 0:
				// Get queue
				queue, err := manager.Get(taskID)
				if err != nil {
					b.Fatalf("Get() error = %v", err)
				}
				// Enqueue event
				event := &a2a.Message{MessageID: "msg-1", Kind: a2a.MessageEventKind}
				queue.EnqueueEvent(ctx, event)
			case 1:
				// Tap queue
				manager.Tap(taskID)
			case 2:
				// Close queue
				manager.Close(taskID)
			case 3:
				// Get task IDs
				manager.TaskIDs()
			}
			i++
		}
	})
}

func BenchmarkIsFinalEvent(b *testing.B) {
	events := []Event{
		&a2a.TaskStatusUpdateEvent{Final: true, Kind: a2a.StatusUpdateEventKind},
		&a2a.TaskStatusUpdateEvent{Final: false, Kind: a2a.StatusUpdateEventKind},
		&a2a.Message{Kind: a2a.MessageEventKind},
		&a2a.Task{Kind: a2a.TaskEventKind, Status: a2a.TaskStatus{State: a2a.TaskStateCompleted}},
		&a2a.Task{Kind: a2a.TaskEventKind, Status: a2a.TaskStatus{State: a2a.TaskStateWorking}},
		&a2a.TaskArtifactUpdateEvent{Kind: a2a.ArtifactUpdateEventKind},
	}

	b.ResetTimer()
	i := 0
	for b.Loop() {
		event := events[i%len(events)]
		_ = IsFinalEvent(event)
		i++
	}
}

func BenchmarkEventQueue_LargeEvents(b *testing.B) {
	ctx := context.Background()
	queue, err := NewEventQueue(100)
	if err != nil {
		b.Fatalf("NewEventQueue() error = %v", err)
	}
	defer queue.Close()

	// Create a large event with many parts
	parts := make([]a2a.Part, 100)
	for i := range parts {
		parts[i] = a2a.NewTextPart("This is a text part with some content that makes the event larger")
	}

	event := &a2a.Message{
		MessageID: "bench-msg-1",
		TaskID:    "bench-task-1",
		Role:      a2a.RoleAgent,
		Kind:      a2a.MessageEventKind,
		Parts:     parts,
		Metadata: map[string]any{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
	}

	// Create a consumer
	consumer := NewEventConsumer(queue)

	// Use a WaitGroup to coordinate producer and consumer
	var wg sync.WaitGroup
	wg.Add(2)

	// Producer
	go func() {
		defer wg.Done()
		for i := 0; i < b.N; i++ {
			if err := queue.EnqueueEvent(ctx, event); err != nil {
				// Queue might be full, which is ok
				continue
			}
		}
	}()

	// Consumer
	consumed := 0
	go func() {
		defer wg.Done()
		for consumed < b.N {
			if _, err := consumer.ConsumeOne(ctx); err == nil {
				consumed++
			}
		}
	}()

	b.ResetTimer()
	wg.Wait()
}