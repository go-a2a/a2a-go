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

// Package event provides asynchronous event handling for A2A server implementations.
//
// This package is a Go port of the Python a2a.server.events module, providing
// a producer-consumer pattern for handling real-time communication between
// A2A agents and clients through asynchronous event streams.
//
// # Architecture
//
// The package consists of three main components:
//
//   - EventQueue: A bounded queue that buffers events with support for creating
//     child queues (tap mechanism) that receive copies of all enqueued events.
//
//   - EventConsumer: Consumes events from an EventQueue, handling final event
//     detection and exception propagation from agent tasks.
//
//   - Event types: Message, Task, TaskStatusUpdateEvent, and TaskArtifactUpdateEvent
//     from the main a2a package are used as event types.
//
// # Usage
//
// Basic producer-consumer pattern:
//
//	queue, err := event.NewEventQueue(100)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer queue.Close()
//
//	// Producer
//	go func() {
//	    event := &a2a.TaskStatusUpdateEvent{...}
//	    if err := queue.EnqueueEvent(ctx, event); err != nil {
//	        log.Printf("Failed to enqueue: %v", err)
//	    }
//	}()
//
//	// Consumer
//	consumer := event.NewEventConsumer(queue)
//	for evt := range consumer.ConsumeAll(ctx) {
//	    switch e := evt.(type) {
//	    case *a2a.Message:
//	        // Handle message
//	    case *a2a.TaskStatusUpdateEvent:
//	        // Handle status update
//	    }
//	}
//
// # Tap Mechanism
//
// The tap mechanism allows creating child queues that receive copies of all
// events enqueued to the parent queue:
//
//	parentQueue, _ := event.NewEventQueue(100)
//	childQueue, _ := parentQueue.Tap()
//
//	// Events enqueued to parentQueue will also be sent to childQueue
//	parentQueue.EnqueueEvent(ctx, event)
//
// This is useful for scenarios where multiple consumers need to process the
// same stream of events independently.
//
// # Final Events
//
// The system automatically detects final events that signal the end of a stream:
//   - TaskStatusUpdateEvent with Final=true
//   - Any Message
//   - Task with terminal state (completed, canceled, failed, rejected)
//
// When a final event is detected, the consumer closes the queue and stops consuming.
package event