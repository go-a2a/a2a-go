// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EventConsumer consumes events from an EventQueue.
// It provides both single event consumption and streaming capabilities.
type EventConsumer struct {
	mu           sync.RWMutex
	queue        *EventQueue
	agentTaskErr error
	closed       bool
	name         string
}

// NewEventConsumer creates a new EventConsumer for the given EventQueue.
func NewEventConsumer(queue *EventQueue) *EventConsumer {
	if queue == nil {
		panic("queue cannot be nil")
	}

	return &EventConsumer{
		queue:        queue,
		agentTaskErr: nil,
		closed:       false,
		name:         fmt.Sprintf("EventConsumer-%s", queue.Name()),
	}
}

// NewEventConsumerWithName creates a new EventConsumer with a custom name.
func NewEventConsumerWithName(queue *EventQueue, name string) *EventConsumer {
	consumer := NewEventConsumer(queue)
	consumer.name = name
	return consumer
}

// ConsumeOne retrieves a single event from the queue non-blocking.
// Returns ErrQueueEmpty if the queue is empty.
// Returns ErrConsumerClosed if the consumer is closed.
func (ec *EventConsumer) ConsumeOne(ctx context.Context) (Event, error) {
	ec.mu.RLock()
	closed := ec.closed
	agentTaskErr := ec.agentTaskErr
	ec.mu.RUnlock()

	if closed {
		return nil, ErrConsumerClosed
	}

	// Check for agent task error first
	if agentTaskErr != nil {
		return nil, NewServerErrorWithCause("agent task failed", agentTaskErr)
	}

	return ec.queue.DequeueEvent(ctx, true)
}

// ConsumeAll streams all events from the queue until a final event is received or the queue is closed.
// It returns a channel that will receive events as they become available.
// The channel will be closed when consumption is complete or an error occurs.
func (ec *EventConsumer) ConsumeAll(ctx context.Context) (<-chan Event, <-chan error) {
	eventChan := make(chan Event, 10) // Buffer to prevent blocking
	errorChan := make(chan error, 1)

	go func() {
		defer close(eventChan)
		defer close(errorChan)

		for {
			ec.mu.RLock()
			closed := ec.closed
			agentTaskErr := ec.agentTaskErr
			ec.mu.RUnlock()

			if closed {
				return
			}

			// Check for agent task error
			if agentTaskErr != nil {
				select {
				case errorChan <- NewServerErrorWithCause("agent task failed", agentTaskErr):
				case <-ctx.Done():
				}
				return
			}

			// Try to get an event
			event, err := ec.queue.DequeueEvent(ctx, false)
			if err != nil {
				if IsQueueClosedError(err) {
					// Queue is closed, stop consuming
					return
				}
				if ctx.Err() != nil {
					// Context cancelled
					return
				}

				// Other error, send it and continue
				select {
				case errorChan <- err:
				case <-ctx.Done():
					return
				}
				continue
			}

			// Send the event
			select {
			case eventChan <- event:
			case <-ctx.Done():
				return
			}

			// Check if this is a final event
			if IsFinalEvent(event) {
				return
			}
		}
	}()

	return eventChan, errorChan
}

// ConsumeAllWithCallback consumes all events and calls the provided callback for each event.
// The callback should return true to continue consuming, false to stop.
func (ec *EventConsumer) ConsumeAllWithCallback(ctx context.Context, callback func(Event) bool) error {
	eventChan, errorChan := ec.ConsumeAll(ctx)

	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				// Channel closed, consumption complete
				return nil
			}

			if !callback(event) {
				// Callback requested to stop
				return nil
			}

		case err, ok := <-errorChan:
			if !ok {
				// Error channel closed
				continue
			}

			if err != nil {
				return err
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// ConsumeUntil consumes events until a condition is met or an error occurs.
func (ec *EventConsumer) ConsumeUntil(ctx context.Context, condition func(Event) bool) ([]Event, error) {
	var events []Event

	eventChan, errorChan := ec.ConsumeAll(ctx)

	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				// Channel closed
				return events, nil
			}

			events = append(events, event)

			if condition(event) {
				return events, nil
			}

		case err, ok := <-errorChan:
			if !ok {
				// Error channel closed
				continue
			}

			if err != nil {
				return events, err
			}

		case <-ctx.Done():
			return events, ctx.Err()
		}
	}
}

// ConsumeWithTimeout consumes events for a specified duration.
func (ec *EventConsumer) ConsumeWithTimeout(timeout time.Duration) ([]Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return ec.ConsumeUntil(ctx, func(event Event) bool {
		return IsFinalEvent(event)
	})
}

// AgentTaskCallback stores an error from the agent's execution task.
// This error will be re-raised during consumption if it occurs.
func (ec *EventConsumer) AgentTaskCallback(err error) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.agentTaskErr = err
}

// Close closes the consumer and prevents further consumption.
func (ec *EventConsumer) Close() error {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if ec.closed {
		return nil
	}

	ec.closed = true
	return nil
}

// IsClosed returns true if the consumer is closed.
func (ec *EventConsumer) IsClosed() bool {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.closed
}

// Queue returns the underlying EventQueue.
func (ec *EventConsumer) Queue() *EventQueue {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.queue
}

// Name returns the name of the consumer.
func (ec *EventConsumer) Name() string {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.name
}

// HasAgentTaskError returns true if an agent task error has been set.
func (ec *EventConsumer) HasAgentTaskError() bool {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.agentTaskErr != nil
}

// GetAgentTaskError returns the current agent task error, if any.
func (ec *EventConsumer) GetAgentTaskError() error {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.agentTaskErr
}

// ClearAgentTaskError clears the current agent task error.
func (ec *EventConsumer) ClearAgentTaskError() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.agentTaskErr = nil
}

// String returns a string representation of the EventConsumer.
func (ec *EventConsumer) String() string {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	return fmt.Sprintf("EventConsumer{name: %s, closed: %t, hasAgentTaskError: %t, queue: %s}",
		ec.name, ec.closed, ec.agentTaskErr != nil, ec.queue.String())
}

// WaitForEvent waits for an event to be available in the underlying queue.
func (ec *EventConsumer) WaitForEvent(ctx context.Context) bool {
	ec.mu.RLock()
	closed := ec.closed
	queue := ec.queue
	ec.mu.RUnlock()

	if closed {
		return false
	}

	return queue.WaitForEvent(ctx)
}

// Peek returns the next event without consuming it.
func (ec *EventConsumer) Peek() Event {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if ec.closed {
		return nil
	}

	return ec.queue.Peek()
}
