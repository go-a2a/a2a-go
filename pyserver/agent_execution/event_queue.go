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
	"errors"
	"sync"
)

// Event represents a base interface for all events that can be sent through the EventQueue.
type Event interface {
	// GetEventType returns the type of event.
	GetEventType() string
}

// EventQueue defines the interface for managing events during agent execution.
// It provides a way for agents to send status updates, messages, and artifacts
// back to the client asynchronously.
type EventQueue interface {
	// Put adds an event to the queue.
	Put(ctx context.Context, event Event) error

	// Get retrieves the next event from the queue, blocking if necessary.
	Get(ctx context.Context) (Event, error)

	// Close closes the event queue and releases any resources.
	Close() error

	// Done returns a channel that's closed when the queue is closed.
	Done() <-chan struct{}

	// Len returns the current number of events in the queue.
	Len() int
}

// ChannelEventQueue implements EventQueue using Go channels.
// This provides a thread-safe, buffered queue for events.
type ChannelEventQueue struct {
	events   chan Event
	done     chan struct{}
	mu       sync.RWMutex
	closed   bool
	capacity int
}

// NewChannelEventQueue creates a new channel-based event queue with the specified capacity.
// If capacity is 0, the queue will be unbuffered.
func NewChannelEventQueue(capacity int) *ChannelEventQueue {
	return &ChannelEventQueue{
		events:   make(chan Event, capacity),
		done:     make(chan struct{}),
		capacity: capacity,
	}
}

// Put adds an event to the queue.
func (q *ChannelEventQueue) Put(ctx context.Context, event Event) error {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.closed {
		return errors.New("event queue is closed")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-q.done:
		return errors.New("event queue is closed")
	case q.events <- event:
		return nil
	}
}

// Get retrieves the next event from the queue, blocking if necessary.
func (q *ChannelEventQueue) Get(ctx context.Context) (Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-q.done:
		return nil, errors.New("event queue is closed")
	case event, ok := <-q.events:
		if !ok {
			return nil, errors.New("event queue is closed")
		}
		return event, nil
	}
}

// Close closes the event queue and releases any resources.
func (q *ChannelEventQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil
	}

	q.closed = true
	close(q.done)
	close(q.events)
	return nil
}

// Done returns a channel that's closed when the queue is closed.
func (q *ChannelEventQueue) Done() <-chan struct{} {
	return q.done
}

// Len returns the current number of events in the queue.
func (q *ChannelEventQueue) Len() int {
	return len(q.events)
}

