// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package task

import (
	"context"
	"testing"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/server/event"
)

func TestTaskManagementIntegration(t *testing.T) {
	ctx := context.Background()

	// Create components
	store := NewInMemoryTaskStore()
	queue := event.NewEventQueue(100)

	// Initialize store
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}

	// Create task updater
	updaterConfig := TaskUpdaterConfig{
		TaskID:    "test-task-1",
		ContextID: "test-context-1",
		Queue:     queue,
	}

	updater, err := NewTaskUpdater(updaterConfig)
	if err != nil {
		t.Fatalf("Failed to create task updater: %v", err)
	}

	// Create task manager
	managerConfig := TaskManagerConfig{
		TaskID:    "test-task-1",
		ContextID: "test-context-1",
		Store:     store,
		InitialMessage: &a2a.Message{
			ContextID: "test-context-1",
			Content:   "Initial task message",
		},
	}

	manager, err := NewTaskManager(managerConfig)
	if err != nil {
		t.Fatalf("Failed to create task manager: %v", err)
	}

	// Create result aggregator
	aggregatorConfig := ResultAggregatorConfig{
		TaskManager: manager,
		BufferSize:  10,
	}

	aggregator, err := NewResultAggregator(aggregatorConfig)
	if err != nil {
		t.Fatalf("Failed to create result aggregator: %v", err)
	}

	// Submit task
	if err := updater.Submit(ctx, "Task submitted"); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	// Start work
	if err := updater.StartWork(ctx, "Starting work"); err != nil {
		t.Fatalf("Failed to start work: %v", err)
	}

	// Add artifact
	textArtifact, err := a2a.NewTextArtifact("test-result", "This is a test result", "Test artifact")
	if err != nil {
		t.Fatalf("Failed to create text artifact: %v", err)
	}

	if err := updater.AddArtifact(ctx, textArtifact, true); err != nil {
		t.Fatalf("Failed to add artifact: %v", err)
	}

	// Complete task
	if err := updater.Complete(ctx, "Task completed successfully"); err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}

	// Process events through the queue
	eventsChan := make(chan event.Event, 10)
	go func() {
		for {
			evt, err := queue.DequeueEvent(ctx, false)
			if err != nil {
				return
			}
			eventsChan <- evt

			// Stop after receiving a final event
			if event.IsFinalEvent(evt) {
				close(eventsChan)
				return
			}
		}
	}()

	// Use result aggregator to process all events
	finalTask, err := aggregator.ConsumeAll(ctx, eventsChan)
	if err != nil {
		t.Fatalf("Failed to consume all events: %v", err)
	}

	// Verify final task state
	if finalTask.Status.State != a2a.TaskStateCompleted {
		t.Errorf("Expected task state %s, got %s", a2a.TaskStateCompleted, finalTask.Status.State)
	}

	if len(finalTask.Artifacts) != 1 {
		t.Errorf("Expected 1 artifact, got %d", len(finalTask.Artifacts))
	}

	if finalTask.Artifacts[0].Name != "test-result" {
		t.Errorf("Expected artifact name 'test-result', got '%s'", finalTask.Artifacts[0].Name)
	}

	if len(finalTask.History) < 3 {
		t.Errorf("Expected at least 3 history messages, got %d", len(finalTask.History))
	}

	// Check that task was persisted
	persistedTask, err := store.Get(ctx, "test-task-1")
	if err != nil {
		t.Fatalf("Failed to get persisted task: %v", err)
	}

	if persistedTask.Status.State != a2a.TaskStateCompleted {
		t.Errorf("Expected persisted task state %s, got %s", a2a.TaskStateCompleted, persistedTask.Status.State)
	}

	// Verify updater is terminal
	if !updater.IsTerminal() {
		t.Error("Expected updater to be terminal after completion")
	}

	// Verify we can't update terminal task
	err = updater.StartWork(ctx, "Should fail")
	if err == nil {
		t.Error("Expected error when trying to update terminal task")
	}

	// Clean up
	if err := updater.Close(); err != nil {
		t.Errorf("Failed to close updater: %v", err)
	}

	if err := manager.Close(); err != nil {
		t.Errorf("Failed to close manager: %v", err)
	}

	if err := aggregator.Close(); err != nil {
		t.Errorf("Failed to close aggregator: %v", err)
	}

	if err := queue.Close(); err != nil {
		t.Errorf("Failed to close queue: %v", err)
	}

	if err := store.Close(ctx); err != nil {
		t.Errorf("Failed to close store: %v", err)
	}
}

func TestTaskManagementWithFailure(t *testing.T) {
	ctx := context.Background()

	// Create components
	store := NewInMemoryTaskStore()
	queue := event.NewEventQueue(100)

	// Initialize store
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}

	// Create task updater
	updaterConfig := TaskUpdaterConfig{
		TaskID:    "test-task-2",
		ContextID: "test-context-2",
		Queue:     queue,
	}

	updater, err := NewTaskUpdater(updaterConfig)
	if err != nil {
		t.Fatalf("Failed to create task updater: %v", err)
	}

	// Create task manager
	managerConfig := TaskManagerConfig{
		TaskID:    "test-task-2",
		ContextID: "test-context-2",
		Store:     store,
	}

	manager, err := NewTaskManager(managerConfig)
	if err != nil {
		t.Fatalf("Failed to create task manager: %v", err)
	}

	// Submit and then fail task
	if err := updater.Submit(ctx, "Task submitted"); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	if err := updater.StartWork(ctx, "Starting work"); err != nil {
		t.Fatalf("Failed to start work: %v", err)
	}

	if err := updater.Failed(ctx, "Task failed due to error"); err != nil {
		t.Fatalf("Failed to fail task: %v", err)
	}

	// Process events
	eventsChan := make(chan event.Event, 10)
	go func() {
		for {
			evt, err := queue.DequeueEvent(ctx, false)
			if err != nil {
				return
			}
			eventsChan <- evt

			// Stop after receiving a final event
			if event.IsFinalEvent(evt) {
				close(eventsChan)
				return
			}
		}
	}()

	// Process all events
	for evt := range eventsChan {
		if err := manager.Process(ctx, evt); err != nil {
			t.Errorf("Failed to process event: %v", err)
		}
	}

	// Verify final task state
	finalTask, err := manager.GetTask(ctx)
	if err != nil {
		t.Fatalf("Failed to get final task: %v", err)
	}

	if finalTask.Status.State != a2a.TaskStateFailed {
		t.Errorf("Expected task state %s, got %s", a2a.TaskStateFailed, finalTask.Status.State)
	}

	// Verify updater is terminal
	if !updater.IsTerminal() {
		t.Error("Expected updater to be terminal after failure")
	}

	// Clean up
	if err := updater.Close(); err != nil {
		t.Errorf("Failed to close updater: %v", err)
	}

	if err := manager.Close(); err != nil {
		t.Errorf("Failed to close manager: %v", err)
	}

	if err := queue.Close(); err != nil {
		t.Errorf("Failed to close queue: %v", err)
	}

	if err := store.Close(ctx); err != nil {
		t.Errorf("Failed to close store: %v", err)
	}
}
