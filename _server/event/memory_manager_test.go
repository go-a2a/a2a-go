// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"testing"
	"time"

	a2a "github.com/go-a2a/a2a-go"
)

func TestNewInMemoryQueueManager(t *testing.T) {
	manager := NewInMemoryQueueManager()

	if manager == nil {
		t.Fatalf("Expected non-nil manager")
	}

	if manager.Count() != 0 {
		t.Errorf("Expected 0 queues initially, got %d", manager.Count())
	}
}

func TestNewInMemoryQueueManagerWithConfig(t *testing.T) {
	config := &QueueManagerConfig{
		DefaultMaxQueueSize: 500,
		AutoCreateQueues:    true,
	}

	manager := NewInMemoryQueueManagerWithConfig(config)

	if manager.Config().DefaultMaxQueueSize != 500 {
		t.Errorf("Expected default max queue size 500, got %d", manager.Config().DefaultMaxQueueSize)
	}

	if !manager.Config().AutoCreateQueues {
		t.Errorf("Expected auto create queues to be true")
	}
}

func TestNewInMemoryQueueManagerWithOptions(t *testing.T) {
	manager := NewInMemoryQueueManagerWithOptions(
		WithDefaultMaxQueueSize(300),
		WithAutoCreateQueues(true),
	)

	if manager.Config().DefaultMaxQueueSize != 300 {
		t.Errorf("Expected default max queue size 300, got %d", manager.Config().DefaultMaxQueueSize)
	}

	if !manager.Config().AutoCreateQueues {
		t.Errorf("Expected auto create queues to be true")
	}
}

func TestInMemoryQueueManagerAdd(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue := NewEventQueue(10)
	taskID := "test-task"

	err := manager.Add(taskID, queue)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if manager.Count() != 1 {
		t.Errorf("Expected 1 queue, got %d", manager.Count())
	}

	if !manager.Exists(taskID) {
		t.Errorf("Expected queue to exist for task ID %s", taskID)
	}
}

func TestInMemoryQueueManagerAddEmptyTaskID(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue := NewEventQueue(10)

	err := manager.Add("", queue)
	if err == nil {
		t.Errorf("Expected error for empty task ID")
	}
}

func TestInMemoryQueueManagerAddNilQueue(t *testing.T) {
	manager := NewInMemoryQueueManager()

	err := manager.Add("test-task", nil)
	if err == nil {
		t.Errorf("Expected error for nil queue")
	}
}

func TestInMemoryQueueManagerAddDuplicate(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue1 := NewEventQueue(10)
	queue2 := NewEventQueue(10)
	taskID := "test-task"

	err := manager.Add(taskID, queue1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = manager.Add(taskID, queue2)
	if err == nil {
		t.Errorf("Expected error for duplicate task ID")
	}

	if !IsTaskQueueExistsError(err) {
		t.Errorf("Expected TaskQueueExistsError, got %v", err)
	}
}

func TestInMemoryQueueManagerGet(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue := NewEventQueue(10)
	taskID := "test-task"

	manager.Add(taskID, queue)

	retrievedQueue := manager.Get(taskID)
	if retrievedQueue != queue {
		t.Errorf("Expected retrieved queue to match original")
	}
}

func TestInMemoryQueueManagerGetNonExistent(t *testing.T) {
	manager := NewInMemoryQueueManager()

	retrievedQueue := manager.Get("non-existent")
	if retrievedQueue != nil {
		t.Errorf("Expected nil for non-existent queue")
	}
}

func TestInMemoryQueueManagerGetEmptyTaskID(t *testing.T) {
	manager := NewInMemoryQueueManager()

	retrievedQueue := manager.Get("")
	if retrievedQueue != nil {
		t.Errorf("Expected nil for empty task ID")
	}
}

func TestInMemoryQueueManagerGetWithAutoCreate(t *testing.T) {
	manager := NewInMemoryQueueManagerWithOptions(WithAutoCreateQueues(true))

	taskID := "test-task"
	retrievedQueue := manager.Get(taskID)

	if retrievedQueue == nil {
		t.Errorf("Expected queue to be auto-created")
	}

	if !manager.Exists(taskID) {
		t.Errorf("Expected queue to exist after auto-creation")
	}
}

func TestInMemoryQueueManagerTap(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue := NewEventQueue(10)
	taskID := "test-task"

	manager.Add(taskID, queue)

	tappedQueue := manager.Tap(taskID)
	if tappedQueue == nil {
		t.Errorf("Expected non-nil tapped queue")
	}

	if queue.ChildCount() != 1 {
		t.Errorf("Expected 1 child queue, got %d", queue.ChildCount())
	}
}

func TestInMemoryQueueManagerTapNonExistent(t *testing.T) {
	manager := NewInMemoryQueueManager()

	tappedQueue := manager.Tap("non-existent")
	if tappedQueue != nil {
		t.Errorf("Expected nil for non-existent queue")
	}
}

func TestInMemoryQueueManagerClose(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue := NewEventQueue(10)
	taskID := "test-task"

	manager.Add(taskID, queue)

	err := manager.Close(taskID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if manager.Count() != 0 {
		t.Errorf("Expected 0 queues after close, got %d", manager.Count())
	}

	if queue.IsClosed() == false {
		t.Errorf("Expected queue to be closed")
	}
}

func TestInMemoryQueueManagerCloseNonExistent(t *testing.T) {
	manager := NewInMemoryQueueManager()

	err := manager.Close("non-existent")
	if err == nil {
		t.Errorf("Expected error for non-existent queue")
	}

	if !IsNoTaskQueueError(err) {
		t.Errorf("Expected NoTaskQueueError, got %v", err)
	}
}

func TestInMemoryQueueManagerCloseEmptyTaskID(t *testing.T) {
	manager := NewInMemoryQueueManager()

	err := manager.Close("")
	if err == nil {
		t.Errorf("Expected error for empty task ID")
	}
}

func TestInMemoryQueueManagerCreateOrTap(t *testing.T) {
	manager := NewInMemoryQueueManager()

	taskID := "test-task"

	// First call should create new queue
	queue1 := manager.CreateOrTap(taskID)
	if queue1 == nil {
		t.Errorf("Expected non-nil queue")
	}

	if manager.Count() != 1 {
		t.Errorf("Expected 1 queue, got %d", manager.Count())
	}

	// Second call should tap existing queue
	queue2 := manager.CreateOrTap(taskID)
	if queue2 == nil {
		t.Errorf("Expected non-nil tapped queue")
	}

	if queue1.ChildCount() != 1 {
		t.Errorf("Expected 1 child queue, got %d", queue1.ChildCount())
	}
}

func TestInMemoryQueueManagerCreateOrTapEmptyTaskID(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue := manager.CreateOrTap("")
	if queue != nil {
		t.Errorf("Expected nil for empty task ID")
	}
}

func TestInMemoryQueueManagerExists(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue := NewEventQueue(10)
	taskID := "test-task"

	if manager.Exists(taskID) {
		t.Errorf("Expected queue to not exist initially")
	}

	manager.Add(taskID, queue)

	if !manager.Exists(taskID) {
		t.Errorf("Expected queue to exist after add")
	}
}

func TestInMemoryQueueManagerList(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue1 := NewEventQueue(10)
	queue2 := NewEventQueue(10)

	manager.Add("task1", queue1)
	manager.Add("task2", queue2)

	taskIDs := manager.List()
	if len(taskIDs) != 2 {
		t.Errorf("Expected 2 task IDs, got %d", len(taskIDs))
	}

	// List should be sorted
	if taskIDs[0] != "task1" || taskIDs[1] != "task2" {
		t.Errorf("Expected sorted task IDs, got %v", taskIDs)
	}
}

func TestInMemoryQueueManagerCloseAll(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue1 := NewEventQueue(10)
	queue2 := NewEventQueue(10)

	manager.Add("task1", queue1)
	manager.Add("task2", queue2)

	err := manager.CloseAll()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if manager.Count() != 0 {
		t.Errorf("Expected 0 queues after close all, got %d", manager.Count())
	}
}

func TestInMemoryQueueManagerGetStats(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue1 := NewEventQueue(10)
	queue2 := NewEventQueue(10)

	manager.Add("task1", queue1)
	manager.Add("task2", queue2)

	// Add events to queues
	message := &a2a.Message{Content: "test", ContextID: "test-context"}
	event := NewMessageEvent(message)

	queue1.EnqueueEvent(t.Context(), event)
	queue2.EnqueueEvent(t.Context(), event)

	// Update the stats manually since we added events after creating the queues
	manager.updateMetrics()

	stats := manager.GetStats()

	if stats.TotalQueues != 2 {
		t.Errorf("Expected 2 total queues, got %d", stats.TotalQueues)
	}

	if stats.ActiveQueues != 2 {
		t.Errorf("Expected 2 active queues, got %d", stats.ActiveQueues)
	}

	if stats.TotalEvents != 2 {
		t.Errorf("Expected 2 total events, got %d", stats.TotalEvents)
	}
}

func TestInMemoryQueueManagerGetQueueInfo(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue := NewEventQueue(10)
	taskID := "test-task"

	manager.Add(taskID, queue)

	info, err := manager.GetQueueInfo(taskID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if info.TaskID != taskID {
		t.Errorf("Expected task ID %s, got %s", taskID, info.TaskID)
	}

	if info.Queue != queue {
		t.Errorf("Expected queue to match original")
	}
}

func TestInMemoryQueueManagerGetQueueInfoNonExistent(t *testing.T) {
	manager := NewInMemoryQueueManager()

	_, err := manager.GetQueueInfo("non-existent")
	if err == nil {
		t.Errorf("Expected error for non-existent queue")
	}

	if !IsNoTaskQueueError(err) {
		t.Errorf("Expected NoTaskQueueError, got %v", err)
	}
}

func TestInMemoryQueueManagerFilter(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue1 := NewEventQueue(10)
	queue2 := NewEventQueue(20)

	manager.Add("task1", queue1)
	manager.Add("task2", queue2)

	// Filter by capacity
	filtered := manager.Filter(QueueCapacityFilter(15))

	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered queue, got %d", len(filtered))
	}

	if filtered["task2"] != queue2 {
		t.Errorf("Expected task2 to be in filtered results")
	}
}

func TestInMemoryQueueManagerFindQueues(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue1 := NewEventQueue(10)
	queue2 := NewEventQueue(10)

	manager.Add("active1", queue1)
	manager.Add("active2", queue2)

	// Find active queues with prefix
	found := manager.FindQueues(ActiveQueueFilter, TaskIDPrefixFilter("active"))

	if len(found) != 2 {
		t.Errorf("Expected 2 found queues, got %d", len(found))
	}
}

func TestInMemoryQueueManagerCountQueues(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue1 := NewEventQueue(10)
	queue2 := NewEventQueue(10)

	manager.Add("task1", queue1)
	manager.Add("task2", queue2)

	// Add events to make queues non-empty
	message := &a2a.Message{Content: "test", ContextID: "test-context"}
	event := NewMessageEvent(message)

	queue1.EnqueueEvent(t.Context(), event)

	count := manager.CountQueues(NonEmptyQueueFilter)

	if count != 1 {
		t.Errorf("Expected 1 non-empty queue, got %d", count)
	}
}

func TestInMemoryQueueManagerBatchAdd(t *testing.T) {
	manager := NewInMemoryQueueManager()

	operations := []QueueOperation{
		{Type: "add", TaskID: "task1", Queue: NewEventQueue(10)},
		{Type: "add", TaskID: "task2", Queue: NewEventQueue(10)},
	}

	results := manager.BatchAdd(operations)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Error != nil {
			t.Errorf("Expected no error for batch add, got %v", result.Error)
		}
	}

	if manager.Count() != 2 {
		t.Errorf("Expected 2 queues after batch add, got %d", manager.Count())
	}
}

func TestInMemoryQueueManagerBatchGet(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue1 := NewEventQueue(10)
	queue2 := NewEventQueue(10)

	manager.Add("task1", queue1)
	manager.Add("task2", queue2)

	results := manager.BatchGet([]string{"task1", "task2", "non-existent"})

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	if results[0].Queue != queue1 {
		t.Errorf("Expected first result to match queue1")
	}

	if results[1].Queue != queue2 {
		t.Errorf("Expected second result to match queue2")
	}

	if results[2].Error == nil {
		t.Errorf("Expected error for non-existent queue")
	}
}

func TestInMemoryQueueManagerBatchClose(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue1 := NewEventQueue(10)
	queue2 := NewEventQueue(10)

	manager.Add("task1", queue1)
	manager.Add("task2", queue2)

	results := manager.BatchClose([]string{"task1", "task2"})

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Error != nil {
			t.Errorf("Expected no error for batch close, got %v", result.Error)
		}
	}

	if manager.Count() != 0 {
		t.Errorf("Expected 0 queues after batch close, got %d", manager.Count())
	}
}

func TestInMemoryQueueManagerAddAsync(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue := NewEventQueue(10)
	taskID := "test-task"

	ctx := context.Background()
	errChan := manager.AddAsync(ctx, taskID, queue)

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	case <-time.After(time.Second):
		t.Errorf("Expected result within timeout")
	}

	if !manager.Exists(taskID) {
		t.Errorf("Expected queue to exist after async add")
	}
}

func TestInMemoryQueueManagerGetAsync(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue := NewEventQueue(10)
	taskID := "test-task"

	manager.Add(taskID, queue)

	ctx := context.Background()
	queueChan := manager.GetAsync(ctx, taskID)

	select {
	case retrievedQueue := <-queueChan:
		if retrievedQueue != queue {
			t.Errorf("Expected retrieved queue to match original")
		}
	case <-time.After(time.Second):
		t.Errorf("Expected result within timeout")
	}
}

func TestInMemoryQueueManagerCloseAsync(t *testing.T) {
	manager := NewInMemoryQueueManager()

	queue := NewEventQueue(10)
	taskID := "test-task"

	manager.Add(taskID, queue)

	ctx := context.Background()
	errChan := manager.CloseAsync(ctx, taskID)

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	case <-time.After(time.Second):
		t.Errorf("Expected result within timeout")
	}

	if manager.Exists(taskID) {
		t.Errorf("Expected queue to not exist after async close")
	}
}

func TestInMemoryQueueManagerString(t *testing.T) {
	manager := NewInMemoryQueueManager()

	str := manager.String()
	if str == "" {
		t.Errorf("Expected non-empty string representation")
	}
}

// Test queue filters
func TestQueueFilters(t *testing.T) {
	queue1 := NewEventQueue(10)
	queue2 := NewEventQueue(20)
	queue3 := NewEventQueue(10)

	queue2.Close()

	// Add event to make queue1 non-empty
	message := &a2a.Message{Content: "test", ContextID: "test-context"}
	event := NewMessageEvent(message)
	queue1.EnqueueEvent(t.Context(), event)

	// Test ActiveQueueFilter
	if !ActiveQueueFilter("task1", queue1) {
		t.Errorf("Expected queue1 to be active")
	}

	if ActiveQueueFilter("task2", queue2) {
		t.Errorf("Expected queue2 to not be active")
	}

	// Test ClosedQueueFilter
	if ClosedQueueFilter("task1", queue1) {
		t.Errorf("Expected queue1 to not be closed")
	}

	if !ClosedQueueFilter("task2", queue2) {
		t.Errorf("Expected queue2 to be closed")
	}

	// Test NonEmptyQueueFilter
	if !NonEmptyQueueFilter("task1", queue1) {
		t.Errorf("Expected queue1 to be non-empty")
	}

	if NonEmptyQueueFilter("task3", queue3) {
		t.Errorf("Expected queue3 to be empty")
	}

	// Test TaskIDPrefixFilter
	prefixFilter := TaskIDPrefixFilter("test")
	if !prefixFilter("test-task", queue1) {
		t.Errorf("Expected 'test-task' to match prefix 'test'")
	}

	if prefixFilter("other-task", queue1) {
		t.Errorf("Expected 'other-task' to not match prefix 'test'")
	}

	// Test QueueCapacityFilter
	capacityFilter := QueueCapacityFilter(15)
	if capacityFilter("task1", queue1) {
		t.Errorf("Expected queue1 capacity to be below threshold")
	}

	if !capacityFilter("task2", queue2) {
		t.Errorf("Expected queue2 capacity to be above threshold")
	}
}
