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

func TestInMemoryQueueManager_GetOrCreate(t *testing.T) {
	t.Parallel()

	manager := NewInMemoryQueueManager(100)

	// Get queue for new task
	queue1, err := manager.Get("task-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if queue1 == nil {
		t.Fatal("Get() returned nil queue")
	}

	// Get same queue again
	queue2, err := manager.Get("task-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Should be the same instance
	if queue1 != queue2 {
		t.Error("Get() should return same queue instance for same task ID")
	}

	// Get queue for different task
	queue3, err := manager.Get("task-2")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Should be different instance
	if queue3 == queue1 {
		t.Error("Get() should return different queue instance for different task ID")
	}

	// Verify manager size
	if manager.Size() != 2 {
		t.Errorf("manager.Size() = %d, want 2", manager.Size())
	}
}

func TestInMemoryQueueManager_Tap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewInMemoryQueueManager(100)

	// Get parent queue
	parent, err := manager.Get("task-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Tap the queue
	child, err := manager.Tap("task-1")
	if err != nil {
		t.Fatalf("Tap() error = %v", err)
	}

	// Enqueue to parent
	event := &a2a.Message{
		MessageID: "msg-1",
		TaskID:    "task-1",
		Kind:      a2a.MessageEventKind,
	}
	if err := parent.EnqueueEvent(ctx, event); err != nil {
		t.Fatalf("EnqueueEvent() error = %v", err)
	}

	// Wait for propagation
	// Note: In real implementation, might need better synchronization
	ctx = context.Background()
	
	// Both should have the event
	parentEvent, err := parent.DequeueEvent(ctx, true)
	if err != nil {
		t.Errorf("parent.DequeueEvent() error = %v", err)
	}

	childEvent, err := child.DequeueEvent(ctx, true)
	if err != nil {
		t.Errorf("child.DequeueEvent() error = %v", err)
	}

	if parentEvent == nil || childEvent == nil {
		t.Error("both parent and child should have received the event")
	}
}

func TestInMemoryQueueManager_Close(t *testing.T) {
	t.Parallel()

	manager := NewInMemoryQueueManager(100)

	// Create queues
	queue1, _ := manager.Get("task-1")
	queue2, _ := manager.Get("task-2")

	// Close one queue
	if err := manager.Close("task-1"); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Queue should be closed
	if !queue1.IsClosed() {
		t.Error("queue1 should be closed")
	}

	// Other queue should still be open
	if queue2.IsClosed() {
		t.Error("queue2 should not be closed")
	}

	// Manager should have one queue left
	if manager.Size() != 1 {
		t.Errorf("manager.Size() = %d, want 1", manager.Size())
	}

	// Closing non-existent queue should not error
	if err := manager.Close("non-existent"); err != nil {
		t.Errorf("Close() on non-existent queue error = %v", err)
	}
}

func TestInMemoryQueueManager_CloseAll(t *testing.T) {
	t.Parallel()

	manager := NewInMemoryQueueManager(100)

	// Create multiple queues
	queues := make([]*EventQueue, 3)
	for i := 0; i < 3; i++ {
		q, _ := manager.Get("task-" + string(rune('0'+i)))
		queues[i] = q
	}

	// Close all
	if err := manager.CloseAll(); err != nil {
		t.Errorf("CloseAll() error = %v", err)
	}

	// All queues should be closed
	for i, q := range queues {
		if !q.IsClosed() {
			t.Errorf("queue[%d] should be closed", i)
		}
	}

	// Manager should be empty
	if manager.Size() != 0 {
		t.Errorf("manager.Size() = %d, want 0", manager.Size())
	}
}

func TestInMemoryQueueManager_TaskIDs(t *testing.T) {
	t.Parallel()

	manager := NewInMemoryQueueManager(100)

	// Create queues
	taskIDs := []string{"task-1", "task-2", "task-3"}
	for _, id := range taskIDs {
		_, err := manager.Get(id)
		if err != nil {
			t.Fatalf("Get(%s) error = %v", id, err)
		}
	}

	// Get task IDs
	gotIDs := manager.TaskIDs()

	// Should have all task IDs (order may differ)
	if len(gotIDs) != len(taskIDs) {
		t.Errorf("TaskIDs() returned %d IDs, want %d", len(gotIDs), len(taskIDs))
	}

	// Check all expected IDs are present
	idMap := make(map[string]bool)
	for _, id := range gotIDs {
		idMap[id] = true
	}

	for _, expectedID := range taskIDs {
		if !idMap[expectedID] {
			t.Errorf("TaskIDs() missing ID: %s", expectedID)
		}
	}
}

func TestInMemoryQueueManager_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	manager := NewInMemoryQueueManager(100)
	const numGoroutines = 10
	const numTasksPerGoroutine = 5

	var wg sync.WaitGroup

	// Concurrent queue creation
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numTasksPerGoroutine; j++ {
				taskID := "task-" + string(rune('0'+goroutineID)) + "-" + string(rune('0'+j))
				_, err := manager.Get(taskID)
				if err != nil {
					t.Errorf("goroutine %d: Get(%s) error = %v", goroutineID, taskID, err)
				}
			}
		}(i)
	}

	// Concurrent tapping
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			// Tap some existing queues
			for j := 0; j < 3; j++ {
				taskID := "task-0-" + string(rune('0'+j))
				_, err := manager.Tap(taskID)
				if err != nil {
					// Might fail if queue doesn't exist yet, which is ok
					continue
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify we have the expected number of queues
	expectedQueues := numGoroutines * numTasksPerGoroutine
	if manager.Size() != expectedQueues {
		t.Errorf("manager.Size() = %d, want %d", manager.Size(), expectedQueues)
	}
}

func TestInMemoryQueueManager_DefaultSize(t *testing.T) {
	t.Parallel()

	// Test with 0 size (should use default)
	manager := NewInMemoryQueueManager(0)
	queue, err := manager.Get("task-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Queue should have default capacity
	if queue.Capacity() != DefaultMaxQueueSize {
		t.Errorf("queue.Capacity() = %d, want %d", queue.Capacity(), DefaultMaxQueueSize)
	}

	// Test with negative size (should use default)
	manager2 := NewInMemoryQueueManager(-1)
	queue2, err := manager2.Get("task-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if queue2.Capacity() != DefaultMaxQueueSize {
		t.Errorf("queue.Capacity() = %d, want %d", queue2.Capacity(), DefaultMaxQueueSize)
	}
}