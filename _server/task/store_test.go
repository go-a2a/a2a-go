// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package task

import (
	"context"
	"fmt"
	"testing"
	"time"

	a2a "github.com/go-a2a/a2a-go"
)

func TestInMemoryTaskStore(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryTaskStore()

	// Test Initialize
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create test task
	task := &a2a.Task{
		ID:        "test-task-1",
		ContextID: "test-context-1",
		Status:    a2a.TaskStatus{State: a2a.TaskStateSubmitted},
		History:   []*a2a.Message{},
		Artifacts: []*a2a.Artifact{},
	}

	// Test Save
	if err := store.Save(ctx, task); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Test Get
	retrievedTask, err := store.Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrievedTask.ID != task.ID {
		t.Errorf("Expected task ID %s, got %s", task.ID, retrievedTask.ID)
	}

	if retrievedTask.ContextID != task.ContextID {
		t.Errorf("Expected context ID %s, got %s", task.ContextID, retrievedTask.ContextID)
	}

	if retrievedTask.Status.State != task.Status.State {
		t.Errorf("Expected status %s, got %s", task.Status.State, retrievedTask.Status.State)
	}

	// Test Get non-existent task
	_, err = store.Get(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error for non-existent task")
	}

	// Test Count
	count, err := store.Count(ctx, "")
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	// Test List
	tasks, err := store.List(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	// Test Delete
	if err := store.Delete(ctx, task.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify task is deleted
	_, err = store.Get(ctx, task.ID)
	if err == nil {
		t.Error("Expected error for deleted task")
	}

	// Test Close
	if err := store.Close(ctx); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestInMemoryTaskStoreWithHistory(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryTaskStore()

	// Create task with history
	task := &a2a.Task{
		ID:        "test-task-2",
		ContextID: "test-context-2",
		Status:    a2a.TaskStatus{State: a2a.TaskStateRunning},
		History: []*a2a.Message{
			{
				ContextID: "test-context-2",
				Content:   "Initial message",
				Metadata: map[string]any{
					"timestamp": time.Now().UTC(),
				},
			},
		},
		Artifacts: []*a2a.Artifact{
			{
				ArtifactID:  "artifact-1",
				Name:        "test-artifact",
				Description: "Test artifact",
				Parts: []*a2a.PartWrapper{
					a2a.NewPartWrapper(&a2a.TextPart{
						Kind: "text",
						Text: "Test content",
					}),
				},
			},
		},
	}

	// Save task
	if err := store.Save(ctx, task); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Retrieve and verify
	retrievedTask, err := store.Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(retrievedTask.History) != 1 {
		t.Errorf("Expected 1 history message, got %d", len(retrievedTask.History))
	}

	if retrievedTask.History[0].Content != "Initial message" {
		t.Errorf("Expected message content 'Initial message', got '%s'", retrievedTask.History[0].Content)
	}

	if len(retrievedTask.Artifacts) != 1 {
		t.Errorf("Expected 1 artifact, got %d", len(retrievedTask.Artifacts))
	}

	if retrievedTask.Artifacts[0].Name != "test-artifact" {
		t.Errorf("Expected artifact name 'test-artifact', got '%s'", retrievedTask.Artifacts[0].Name)
	}
}

func TestInMemoryTaskStoreFiltering(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryTaskStore()

	// Create tasks with different context IDs
	task1 := &a2a.Task{
		ID:        "task-1",
		ContextID: "context-1",
		Status:    a2a.TaskStatus{State: a2a.TaskStateSubmitted},
		History:   []*a2a.Message{},
		Artifacts: []*a2a.Artifact{},
	}

	task2 := &a2a.Task{
		ID:        "task-2",
		ContextID: "context-2",
		Status:    a2a.TaskStatus{State: a2a.TaskStateRunning},
		History:   []*a2a.Message{},
		Artifacts: []*a2a.Artifact{},
	}

	task3 := &a2a.Task{
		ID:        "task-3",
		ContextID: "context-1",
		Status:    a2a.TaskStatus{State: a2a.TaskStateCompleted},
		History:   []*a2a.Message{},
		Artifacts: []*a2a.Artifact{},
	}

	// Save tasks
	if err := store.Save(ctx, task1); err != nil {
		t.Fatalf("Save task1 failed: %v", err)
	}
	if err := store.Save(ctx, task2); err != nil {
		t.Fatalf("Save task2 failed: %v", err)
	}
	if err := store.Save(ctx, task3); err != nil {
		t.Fatalf("Save task3 failed: %v", err)
	}

	// Test count all
	count, err := store.Count(ctx, "")
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// Test count by context
	count1, err := store.Count(ctx, "context-1")
	if err != nil {
		t.Fatalf("Count context-1 failed: %v", err)
	}
	if count1 != 2 {
		t.Errorf("Expected count 2 for context-1, got %d", count1)
	}

	count2, err := store.Count(ctx, "context-2")
	if err != nil {
		t.Fatalf("Count context-2 failed: %v", err)
	}
	if count2 != 1 {
		t.Errorf("Expected count 1 for context-2, got %d", count2)
	}

	// Test list by context
	tasks1, err := store.List(ctx, "context-1", 10, 0)
	if err != nil {
		t.Fatalf("List context-1 failed: %v", err)
	}
	if len(tasks1) != 2 {
		t.Errorf("Expected 2 tasks for context-1, got %d", len(tasks1))
	}

	tasks2, err := store.List(ctx, "context-2", 10, 0)
	if err != nil {
		t.Fatalf("List context-2 failed: %v", err)
	}
	if len(tasks2) != 1 {
		t.Errorf("Expected 1 task for context-2, got %d", len(tasks2))
	}

	// Test list with limit and offset
	tasksLimited, err := store.List(ctx, "", 2, 1)
	if err != nil {
		t.Fatalf("List with limit failed: %v", err)
	}
	if len(tasksLimited) != 2 {
		t.Errorf("Expected 2 tasks with limit, got %d", len(tasksLimited))
	}
}

func TestInMemoryTaskStoreValidation(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryTaskStore()

	// Test saving nil task
	err := store.Save(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil task")
	}

	// Test saving invalid task (empty ID)
	invalidTask := &a2a.Task{
		ID:        "",
		ContextID: "test-context",
		Status:    a2a.TaskStatus{State: a2a.TaskStateSubmitted},
		History:   []*a2a.Message{},
		Artifacts: []*a2a.Artifact{},
	}

	err = store.Save(ctx, invalidTask)
	if err == nil {
		t.Error("Expected error for invalid task")
	}

	// Test getting task with empty ID
	_, err = store.Get(ctx, "")
	if err == nil {
		t.Error("Expected error for empty task ID")
	}

	// Test deleting task with empty ID
	err = store.Delete(ctx, "")
	if err == nil {
		t.Error("Expected error for empty task ID")
	}
}

func TestInMemoryTaskStoreConcurrency(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryTaskStore()

	const numGoroutines = 10
	const numTasks = 100

	// Create tasks concurrently
	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numTasks; j++ {
				task := &a2a.Task{
					ID:        fmt.Sprintf("task-%d-%d", id, j),
					ContextID: fmt.Sprintf("context-%d", id),
					Status:    a2a.TaskStatus{State: a2a.TaskStateSubmitted},
					History:   []*a2a.Message{},
					Artifacts: []*a2a.Artifact{},
				}

				if err := store.Save(ctx, task); err != nil {
					t.Errorf("Save failed: %v", err)
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify total count
	count, err := store.Count(ctx, "")
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	expectedCount := int64(numGoroutines * numTasks)
	if count != expectedCount {
		t.Errorf("Expected count %d, got %d", expectedCount, count)
	}
}
