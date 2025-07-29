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
	"errors"
	"sync"
	"time"
)

// DefaultEventTimeout is the default timeout for polling events.
const DefaultEventTimeout = 2 * time.Second

// EventConsumer consumes events from an EventQueue, handling final event detection
// and exception propagation from agent tasks.
// This corresponds to Python's EventConsumer class.
type EventConsumer struct {
	queue     *EventQueue
	timeout   time.Duration
	mu        sync.RWMutex
	exception error
}

// NewEventConsumer creates a new event consumer for the given queue.
func NewEventConsumer(queue *EventQueue) *EventConsumer {
	return &EventConsumer{
		queue:   queue,
		timeout: DefaultEventTimeout,
	}
}

// ConsumeOne attempts to consume a single event from the queue in non-blocking mode.
// Returns ErrQueueEmpty if the queue is empty.
func (c *EventConsumer) ConsumeOne(ctx context.Context) (Event, error) {
	event, err := c.queue.DequeueEvent(ctx, true)
	if err != nil {
		return nil, err
	}

	c.queue.TaskDone()
	return event, nil
}

// ConsumeAll returns a channel that yields events as they become available.
// The channel is closed when a final event is received or the queue is closed.
// This corresponds to Python's async generator consume_all method.
func (c *EventConsumer) ConsumeAll(ctx context.Context) <-chan Event {
	events := make(chan Event)

	go func() {
		defer close(events)

		for {
			// Check if an exception was set by agent task callback
			c.mu.RLock()
			if c.exception != nil {
				c.mu.RUnlock()
				// In Go, we can't "re-raise" the exception like Python,
				// but we can log it and stop consuming
				return
			}
			c.mu.RUnlock()

			// Create a timeout context for this iteration
			timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
			event, err := c.queue.DequeueEvent(timeoutCtx, false)
			cancel()

			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					// Timeout - continue polling
					continue
				}
				if errors.Is(err, ErrQueueClosed) && c.queue.IsClosed() {
					// Queue is closed and empty
					return
				}
				// Other errors - stop consuming
				return
			}

			// Successfully got an event
			c.queue.TaskDone()

			// Check if it's a final event
			if IsFinalEvent(event) {
				// Send the final event before closing
				select {
				case events <- event:
				case <-ctx.Done():
					return
				}
				// Close the queue and stop consuming
				_ = c.queue.Close()
				return
			}

			// Send the event
			select {
			case events <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return events
}

// SetAgentTaskError sets an error from the agent task execution.
// This corresponds to Python's agent_task_callback method.
func (c *EventConsumer) SetAgentTaskError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.exception = err
}

// GetError returns any error set by the agent task.
func (c *EventConsumer) GetError() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.exception
}

// SetTimeout sets the timeout for event polling.
func (c *EventConsumer) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timeout = timeout
}

