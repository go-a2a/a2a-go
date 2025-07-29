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
)

// DefaultMaxQueueSize is the default maximum queue size.
const DefaultMaxQueueSize = 1024

// EventQueue manages a bounded queue of events with support for creating child queues
// that receive copies of all enqueued events (tap mechanism).
// This corresponds to Python's EventQueue class.
type EventQueue struct {
	events      chan Event
	maxSize     int
	mu          sync.RWMutex
	closed      bool
	closeOnce   sync.Once
	children    []*EventQueue
	doneSignal  chan struct{}
	taskCounter int
}

// NewEventQueue creates a new event queue with the specified maximum size.
// If maxSize is 0, DefaultMaxQueueSize is used.
func NewEventQueue(maxSize int) (*EventQueue, error) {
	if maxSize < 0 {
		return nil, ErrInvalidQueueSize
	}
	if maxSize == 0 {
		maxSize = DefaultMaxQueueSize
	}

	return &EventQueue{
		events:     make(chan Event, maxSize),
		maxSize:    maxSize,
		children:   make([]*EventQueue, 0),
		doneSignal: make(chan struct{}),
	}, nil
}

// EnqueueEvent adds an event to the queue and propagates it to all child queues.
// Returns ErrQueueClosed if the queue is closed.
func (q *EventQueue) EnqueueEvent(ctx context.Context, event Event) error {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.closed {
		return ErrQueueClosed
	}

	// Try to enqueue to this queue
	select {
	case <-ctx.Done():
		return ctx.Err()
	case q.events <- event:
		// Successfully enqueued, now propagate to children
		for _, child := range q.children {
			// Best effort propagation to children - don't block
			go func(c *EventQueue) {
				_ = c.EnqueueEvent(context.Background(), event)
			}(child)
		}
		return nil
	default:
		// Queue is full
		return ErrQueueClosed
	}
}

// DequeueEvent retrieves an event from the queue.
// If noWait is true, returns immediately with ErrQueueEmpty if queue is empty.
// If noWait is false, blocks until an event is available or context is canceled.
func (q *EventQueue) DequeueEvent(ctx context.Context, noWait bool) (Event, error) {
	if noWait {
		select {
		case event := <-q.events:
			return event, nil
		default:
			q.mu.RLock()
			closed := q.closed
			q.mu.RUnlock()
			if closed && len(q.events) == 0 {
				return nil, ErrQueueClosed
			}
			return nil, ErrQueueEmpty
		}
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case event := <-q.events:
		return event, nil
	case <-q.doneSignal:
		// Check if there are still events in the queue
		select {
		case event := <-q.events:
			return event, nil
		default:
			return nil, ErrQueueClosed
		}
	}
}

// TaskDone signals that a dequeued task is complete.
// This is used for tracking processed items, similar to Python's queue.task_done().
func (q *EventQueue) TaskDone() {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	if q.taskCounter > 0 {
		q.taskCounter--
	}
}

// Tap creates and returns a new EventQueue that will receive all future events
// enqueued to this queue. This implements the tap mechanism from Python.
func (q *EventQueue) Tap() (*EventQueue, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil, ErrQueueClosed
	}

	child, err := NewEventQueue(q.maxSize)
	if err != nil {
		return nil, err
	}

	q.children = append(q.children, child)
	return child, nil
}

// Close closes the queue, preventing future enqueues and causing dequeue operations
// to eventually return ErrQueueClosed when empty.
func (q *EventQueue) Close() error {
	var closeErr error
	
	q.closeOnce.Do(func() {
		q.mu.Lock()
		defer q.mu.Unlock()

		q.closed = true
		close(q.doneSignal)

		// Propagate close to all children
		for _, child := range q.children {
			if err := child.Close(); err != nil && closeErr == nil {
				closeErr = err
			}
		}
	})

	return closeErr
}

// IsClosed returns true if the queue is closed.
func (q *EventQueue) IsClosed() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.closed
}

// Size returns the current number of events in the queue.
func (q *EventQueue) Size() int {
	return len(q.events)
}

// Capacity returns the maximum capacity of the queue.
func (q *EventQueue) Capacity() int {
	return q.maxSize
}