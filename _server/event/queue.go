// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DefaultMaxQueueSize is the default maximum size for event queues.
const DefaultMaxQueueSize = 1000

// EventQueue is a bounded asynchronous queue for A2A events.
// It supports tapping to create child queues that receive the same events.
type EventQueue struct {
	mu           sync.RWMutex
	queue        chan Event
	children     []*EventQueue
	closed       bool
	maxQueueSize int
	name         string
}

// NewEventQueue creates a new EventQueue with the specified maximum size.
func NewEventQueue(maxQueueSize int) *EventQueue {
	if maxQueueSize <= 0 {
		maxQueueSize = DefaultMaxQueueSize
	}

	return &EventQueue{
		queue:        make(chan Event, maxQueueSize),
		children:     make([]*EventQueue, 0),
		closed:       false,
		maxQueueSize: maxQueueSize,
		name:         fmt.Sprintf("EventQueue-%p", &EventQueue{}),
	}
}

// NewEventQueueWithName creates a new EventQueue with the specified name and maximum size.
func NewEventQueueWithName(name string, maxQueueSize int) *EventQueue {
	eq := NewEventQueue(maxQueueSize)
	eq.name = name
	return eq
}

// EnqueueEvent adds an event to this queue and all its children.
// Returns an error if the queue is closed or if the event is invalid.
func (eq *EventQueue) EnqueueEvent(ctx context.Context, event Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	if err := event.Validate(); err != nil {
		return NewEventProcessingError(event, "event validation failed", err)
	}

	eq.mu.RLock()
	defer eq.mu.RUnlock()

	if eq.closed {
		return ErrQueueClosed
	}

	// Try to send to main queue (non-blocking)
	select {
	case eq.queue <- event:
		// Successfully sent to main queue
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Queue is full
		return ErrQueueFull
	}

	// Send to all child queues
	for _, child := range eq.children {
		// Send to child queue asynchronously to avoid blocking
		go func(childQueue *EventQueue) {
			if err := childQueue.EnqueueEvent(ctx, event); err != nil {
				// Log error but don't fail the main enqueue operation
				// In a real implementation, this would use a proper logger
				fmt.Printf("Failed to enqueue event to child queue: %v\n", err)
			}
		}(child)
	}

	return nil
}

// DequeueEvent retrieves an event from the queue.
// If noWait is true, it returns immediately with ErrQueueEmpty if no event is available.
// If noWait is false, it blocks until an event is available or the context is cancelled.
func (eq *EventQueue) DequeueEvent(ctx context.Context, noWait bool) (Event, error) {
	eq.mu.RLock()
	closed := eq.closed
	queue := eq.queue
	eq.mu.RUnlock()

	if closed {
		// Try to drain any remaining events before returning closed error
		select {
		case event, ok := <-queue:
			if ok {
				return event, nil
			}
			return nil, ErrQueueClosed
		default:
			return nil, ErrQueueClosed
		}
	}

	if noWait {
		select {
		case event := <-queue:
			return event, nil
		default:
			return nil, ErrQueueEmpty
		}
	}

	// Blocking wait with context support
	select {
	case event := <-queue:
		return event, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// DequeueEventWithTimeout retrieves an event from the queue with a timeout.
func (eq *EventQueue) DequeueEventWithTimeout(timeout time.Duration) (Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return eq.DequeueEvent(ctx, false)
}

// Tap creates a new child EventQueue that receives all future events.
// The child queue has the same maximum size as the parent.
func (eq *EventQueue) Tap() *EventQueue {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	if eq.closed {
		// Return a closed queue if parent is closed
		child := NewEventQueue(eq.maxQueueSize)
		child.closed = true
		close(child.queue)
		return child
	}

	child := NewEventQueue(eq.maxQueueSize)
	child.name = fmt.Sprintf("%s-child-%d", eq.name, len(eq.children))
	eq.children = append(eq.children, child)

	return child
}

// Close closes the queue for future push events and propagates the close signal to all child queues.
func (eq *EventQueue) Close() error {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	if eq.closed {
		return nil // Already closed
	}

	eq.closed = true
	close(eq.queue)

	// Close all child queues
	for _, child := range eq.children {
		if err := child.Close(); err != nil {
			// Log error but continue closing other children
			fmt.Printf("Failed to close child queue: %v\n", err)
		}
	}

	return nil
}

// IsClosed checks if the queue is closed.
func (eq *EventQueue) IsClosed() bool {
	eq.mu.RLock()
	defer eq.mu.RUnlock()
	return eq.closed
}

// Len returns the current number of events in the queue.
func (eq *EventQueue) Len() int {
	eq.mu.RLock()
	defer eq.mu.RUnlock()
	return len(eq.queue)
}

// Cap returns the maximum capacity of the queue.
func (eq *EventQueue) Cap() int {
	eq.mu.RLock()
	defer eq.mu.RUnlock()
	return eq.maxQueueSize
}

// Name returns the name of the queue.
func (eq *EventQueue) Name() string {
	eq.mu.RLock()
	defer eq.mu.RUnlock()
	return eq.name
}

// ChildCount returns the number of child queues.
func (eq *EventQueue) ChildCount() int {
	eq.mu.RLock()
	defer eq.mu.RUnlock()
	return len(eq.children)
}

// String returns a string representation of the EventQueue.
func (eq *EventQueue) String() string {
	eq.mu.RLock()
	defer eq.mu.RUnlock()

	return fmt.Sprintf("EventQueue{name: %s, len: %d, cap: %d, children: %d, closed: %t}",
		eq.name, len(eq.queue), eq.maxQueueSize, len(eq.children), eq.closed)
}

// DrainEvents drains all events from the queue and returns them.
// This is useful for cleanup or testing purposes.
func (eq *EventQueue) DrainEvents() []Event {
	eq.mu.RLock()
	defer eq.mu.RUnlock()

	var events []Event
	for {
		select {
		case event := <-eq.queue:
			events = append(events, event)
		default:
			return events
		}
	}
}

// WaitForEvent waits for an event to be available in the queue.
// Returns true if an event is available, false if the queue is closed.
func (eq *EventQueue) WaitForEvent(ctx context.Context) bool {
	eq.mu.RLock()
	closed := eq.closed
	queue := eq.queue
	eq.mu.RUnlock()

	if closed {
		return len(queue) > 0
	}

	select {
	case <-ctx.Done():
		return false
	default:
		// Check if there's an event available without consuming it
		select {
		case event := <-queue:
			// Put the event back (this is a bit hacky but works for the use case)
			select {
			case queue <- event:
				return true
			default:
				// Queue is full, which shouldn't happen since we just took an event
				return true
			}
		default:
			// No event available, wait for one
			select {
			case event := <-queue:
				// Put the event back
				select {
				case queue <- event:
					return true
				default:
					return true
				}
			case <-ctx.Done():
				return false
			}
		}
	}
}

// Peek returns the next event without removing it from the queue.
// Returns nil if no event is available.
func (eq *EventQueue) Peek() Event {
	eq.mu.RLock()
	defer eq.mu.RUnlock()

	select {
	case event := <-eq.queue:
		// Put the event back at the front
		select {
		case eq.queue <- event:
			return event
		default:
			// This shouldn't happen since we just took an event
			return event
		}
	default:
		return nil
	}
}
