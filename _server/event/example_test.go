// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package event_test

import (
	"context"
	"fmt"
	"time"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/server/event"
)

// ExampleEventQueue demonstrates basic usage of EventQueue.
func ExampleEventQueue() {
	// Create a new event queue with capacity 10
	queue := event.NewEventQueue(10)

	// Create a message event
	message := &a2a.Message{
		Content:   "Hello, world!",
		ContextID: "example-context",
	}
	messageEvent := event.NewMessageEvent(message)

	ctx := context.Background()

	// Enqueue the event
	err := queue.EnqueueEvent(ctx, messageEvent)
	if err != nil {
		fmt.Printf("Error enqueueing event: %v\n", err)
		return
	}

	// Dequeue the event
	dequeuedEvent, err := queue.DequeueEvent(ctx, true)
	if err != nil {
		fmt.Printf("Error dequeuing event: %v\n", err)
		return
	}

	fmt.Printf("Dequeued event type: %s\n", dequeuedEvent.EventType())

	// Close the queue
	queue.Close()

	// Output:
	// Dequeued event type: message
}

// ExampleEventConsumer demonstrates basic usage of EventConsumer.
func ExampleEventConsumer() {
	// Create a queue and consumer
	queue := event.NewEventQueue(10)
	consumer := event.NewEventConsumer(queue)

	// Create events
	event1 := event.NewTaskStatusUpdateEvent("task1", "context1",
		a2a.TaskStatus{State: a2a.TaskStateRunning}, false, nil)
	event2 := event.NewTaskStatusUpdateEvent("task2", "context2",
		a2a.TaskStatus{State: a2a.TaskStateCompleted}, true, nil)

	ctx := context.Background()

	// Enqueue events
	queue.EnqueueEvent(ctx, event1)
	queue.EnqueueEvent(ctx, event2)

	// Consume all events
	eventChan, errorChan := consumer.ConsumeAll(ctx)

	for {
		select {
		case evt, ok := <-eventChan:
			if !ok {
				fmt.Println("All events consumed")
				return
			}
			fmt.Printf("Consumed event: %s\n", evt.EventType())

		case err, ok := <-errorChan:
			if !ok {
				continue
			}
			if err != nil {
				fmt.Printf("Error consuming event: %v\n", err)
				return
			}
		}
	}

	// Output:
	// Consumed event: task_status_update
	// Consumed event: task_status_update
	// All events consumed
}

// ExampleInMemoryQueueManager demonstrates basic usage of InMemoryQueueManager.
func ExampleInMemoryQueueManager() {
	// Create a queue manager
	manager := event.NewInMemoryQueueManager()

	// Create and add a queue
	queue := event.NewEventQueue(10)
	err := manager.Add("task1", queue)
	if err != nil {
		fmt.Printf("Error adding queue: %v\n", err)
		return
	}

	// Get the queue
	retrievedQueue := manager.Get("task1")
	if retrievedQueue == nil {
		fmt.Println("Queue not found")
		return
	}

	// Create a tapped queue
	tappedQueue := manager.Tap("task1")
	if tappedQueue == nil {
		fmt.Println("Failed to tap queue")
		return
	}

	ctx := context.Background()

	// Enqueue an event
	message := &a2a.Message{Content: "test", ContextID: "test-context"}
	messageEvent := event.NewMessageEvent(message)
	retrievedQueue.EnqueueEvent(ctx, messageEvent)

	// Wait for async propagation to tapped queue
	time.Sleep(10 * time.Millisecond)

	// Both queues should receive the event

	// Consume from original queue
	originalEvent, err := retrievedQueue.DequeueEvent(ctx, true)
	if err != nil {
		fmt.Printf("Error dequeuing from original queue: %v\n", err)
		return
	}
	fmt.Printf("Original queue event: %s\n", originalEvent.EventType())

	// Consume from tapped queue
	tappedEvent, err := tappedQueue.DequeueEvent(ctx, true)
	if err != nil {
		fmt.Printf("Error dequeuing from tapped queue: %v\n", err)
		return
	}
	fmt.Printf("Tapped queue event: %s\n", tappedEvent.EventType())

	// Close the queue
	manager.Close("task1")

	fmt.Printf("Queue count: %d\n", manager.Count())

	// Output:
	// Original queue event: message
	// Tapped queue event: message
	// Queue count: 0
}

// ExampleEventQueue_tapping demonstrates the tapping functionality.
func ExampleEventQueue_tapping() {
	// Create a parent queue
	parent := event.NewEventQueue(10)

	// Create child queues by tapping
	child1 := parent.Tap()
	child2 := parent.Tap()

	// Create an event
	statusEvent := event.NewTaskStatusUpdateEvent("task1", "context1",
		a2a.TaskStatus{State: a2a.TaskStateRunning}, false, nil)

	ctx := context.Background()

	// Enqueue to parent - all children will receive it
	parent.EnqueueEvent(ctx, statusEvent)

	// Give a moment for async propagation
	time.Sleep(10 * time.Millisecond)

	// Consume from parent
	parentEvent, _ := parent.DequeueEvent(ctx, true)
	fmt.Printf("Parent received: %s\n", parentEvent.EventType())

	// Consume from child1
	child1Event, _ := child1.DequeueEvent(ctx, true)
	fmt.Printf("Child1 received: %s\n", child1Event.EventType())

	// Consume from child2
	child2Event, _ := child2.DequeueEvent(ctx, true)
	fmt.Printf("Child2 received: %s\n", child2Event.EventType())

	// Output:
	// Parent received: task_status_update
	// Child1 received: task_status_update
	// Child2 received: task_status_update
}

// ExampleEventConsumer_callback demonstrates callback-based event consumption.
func ExampleEventConsumer_callback() {
	// Create queue and consumer
	queue := event.NewEventQueue(10)
	consumer := event.NewEventConsumer(queue)

	// Create events
	events := []event.Event{
		event.NewTaskStatusUpdateEvent("task1", "context1",
			a2a.TaskStatus{State: a2a.TaskStateRunning}, false, nil),
		event.NewTaskStatusUpdateEvent("task2", "context2",
			a2a.TaskStatus{State: a2a.TaskStateRunning}, false, nil),
		event.NewTaskStatusUpdateEvent("task3", "context3",
			a2a.TaskStatus{State: a2a.TaskStateCompleted}, true, nil),
	}

	ctx := context.Background()

	// Enqueue events
	for _, evt := range events {
		queue.EnqueueEvent(ctx, evt)
	}

	// Consume with callback
	eventCount := 0

	err := consumer.ConsumeAllWithCallback(ctx, func(evt event.Event) bool {
		eventCount++
		fmt.Printf("Event %d: %s\n", eventCount, evt.EventType())
		return true // Continue consuming
	})
	if err != nil {
		fmt.Printf("Error consuming events: %v\n", err)
		return
	}

	fmt.Printf("Total events consumed: %d\n", eventCount)

	// Output:
	// Event 1: task_status_update
	// Event 2: task_status_update
	// Event 3: task_status_update
	// Total events consumed: 3
}
