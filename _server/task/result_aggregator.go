// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/server/event"
)

// ResultAggregator processes event streams and provides different consumption patterns.
// It uses TaskManager to process events and maintain task state while providing
// various ways to consume and respond to events.
type ResultAggregator interface {
	// ConsumeAndEmit processes events from input channel and emits them to output channel.
	// This is used for streaming events with side effects.
	ConsumeAndEmit(ctx context.Context, input <-chan event.Event) (<-chan event.Event, error)

	// ConsumeAll processes all events from the input channel until completion.
	// This is used for blocking consumption until a task is complete.
	ConsumeAll(ctx context.Context, input <-chan event.Event) (*a2a.Task, error)

	// ConsumeAndBreakOnInterrupt processes events until an interruptible state is reached.
	// This is used for handling states like "requires_input" or "requires_auth".
	ConsumeAndBreakOnInterrupt(ctx context.Context, input <-chan event.Event) (*a2a.Task, error)

	// ContinueConsuming continues processing after an interrupt in a background goroutine.
	// This allows resuming event processing after handling an interrupt.
	ContinueConsuming(ctx context.Context, input <-chan event.Event) error

	// GetTaskManager returns the underlying task manager.
	GetTaskManager() TaskManager

	// Close shuts down the aggregator and releases resources.
	Close() error
}

// ResultAggregatorConfig holds configuration for creating a ResultAggregator.
type ResultAggregatorConfig struct {
	TaskManager TaskManager
	BufferSize  int // Buffer size for internal channels
}

// defaultResultAggregator is the default implementation of ResultAggregator.
type defaultResultAggregator struct {
	taskManager TaskManager
	bufferSize  int

	mu     sync.RWMutex
	closed bool
}

var _ ResultAggregator = (*defaultResultAggregator)(nil)

// NewResultAggregator creates a new ResultAggregator with the given configuration.
func NewResultAggregator(config ResultAggregatorConfig) (ResultAggregator, error) {
	if config.TaskManager == nil {
		return nil, fmt.Errorf("task manager cannot be nil")
	}

	bufferSize := config.BufferSize
	if bufferSize <= 0 {
		bufferSize = 100 // Default buffer size
	}

	return &defaultResultAggregator{
		taskManager: config.TaskManager,
		bufferSize:  bufferSize,
	}, nil
}

// ConsumeAndEmit processes events from input channel and emits them to output channel.
func (r *defaultResultAggregator) ConsumeAndEmit(ctx context.Context, input <-chan event.Event) (<-chan event.Event, error) {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return nil, fmt.Errorf("result aggregator is closed")
	}
	r.mu.RUnlock()

	output := make(chan event.Event, r.bufferSize)

	go func() {
		defer close(output)

		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-input:
				if !ok {
					return
				}

				// Process the event
				if err := r.taskManager.Process(ctx, evt); err != nil {
					// Log error but continue processing
					// In a real implementation, this would use structured logging
					fmt.Printf("Error processing event: %v\n", err)
				}

				// Emit the event
				select {
				case output <- evt:
				case <-ctx.Done():
					return
				}

				// Check if this is a final event
				if event.IsFinalEvent(evt) {
					return
				}
			}
		}
	}()

	return output, nil
}

// ConsumeAll processes all events from the input channel until completion.
func (r *defaultResultAggregator) ConsumeAll(ctx context.Context, input <-chan event.Event) (*a2a.Task, error) {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return nil, fmt.Errorf("result aggregator is closed")
	}
	r.mu.RUnlock()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case evt, ok := <-input:
			if !ok {
				// Channel closed, return current task
				return r.taskManager.GetTask(ctx)
			}

			// Process the event
			if err := r.taskManager.Process(ctx, evt); err != nil {
				return nil, fmt.Errorf("error processing event: %w", err)
			}

			// Check if this is a final event
			if event.IsFinalEvent(evt) {
				return r.taskManager.GetTask(ctx)
			}
		}
	}
}

// ConsumeAndBreakOnInterrupt processes events until an interruptible state is reached.
func (r *defaultResultAggregator) ConsumeAndBreakOnInterrupt(ctx context.Context, input <-chan event.Event) (*a2a.Task, error) {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return nil, fmt.Errorf("result aggregator is closed")
	}
	r.mu.RUnlock()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case evt, ok := <-input:
			if !ok {
				// Channel closed, return current task
				return r.taskManager.GetTask(ctx)
			}

			// Process the event
			if err := r.taskManager.Process(ctx, evt); err != nil {
				return nil, fmt.Errorf("error processing event: %w", err)
			}

			// Check if this is an interruptible event
			if r.isInterruptibleEvent(evt) {
				return r.taskManager.GetTask(ctx)
			}

			// Check if this is a final event
			if event.IsFinalEvent(evt) {
				return r.taskManager.GetTask(ctx)
			}
		}
	}
}

// ContinueConsuming continues processing after an interrupt in a background goroutine.
func (r *defaultResultAggregator) ContinueConsuming(ctx context.Context, input <-chan event.Event) error {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return fmt.Errorf("result aggregator is closed")
	}
	r.mu.RUnlock()

	// Start a goroutine to continue consuming
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-input:
				if !ok {
					return
				}

				// Process the event
				if err := r.taskManager.Process(ctx, evt); err != nil {
					// Log error but continue processing
					// In a real implementation, this would use structured logging
					fmt.Printf("Error processing event in background: %v\n", err)
				}

				// Check if this is a final event
				if event.IsFinalEvent(evt) {
					return
				}
			}
		}
	}()

	return nil
}

// GetTaskManager returns the underlying task manager.
func (r *defaultResultAggregator) GetTaskManager() TaskManager {
	return r.taskManager
}

// Close shuts down the aggregator and releases resources.
func (r *defaultResultAggregator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	return r.taskManager.Close()
}

// isInterruptibleEvent checks if an event represents an interruptible state.
func (r *defaultResultAggregator) isInterruptibleEvent(evt event.Event) bool {
	switch e := evt.(type) {
	case *event.TaskStatusUpdateEvent:
		// Check if the task state is interruptible
		return isInterruptibleTaskState(e.Status.State)
	case *event.TaskEvent:
		// Check if the task is in an interruptible state
		if e.Task != nil {
			return isInterruptibleTaskState(e.Task.Status.State)
		}
	}
	return false
}

// isInterruptibleTaskState checks if a task state is interruptible.
func isInterruptibleTaskState(state a2a.TaskState) bool {
	return state == a2a.TaskStateInputRequired || state == a2a.TaskStateAuthRequired
}

// ConsumeWithTimeout processes events with a timeout.
func (r *defaultResultAggregator) ConsumeWithTimeout(ctx context.Context, input <-chan event.Event, timeout time.Duration) (*a2a.Task, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.ConsumeAll(timeoutCtx, input)
}

// ConsumeWithDeadline processes events with a deadline.
func (r *defaultResultAggregator) ConsumeWithDeadline(ctx context.Context, input <-chan event.Event, deadline time.Time) (*a2a.Task, error) {
	deadlineCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	return r.ConsumeAll(deadlineCtx, input)
}
