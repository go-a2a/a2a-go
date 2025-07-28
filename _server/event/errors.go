// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"errors"
	"fmt"
)

// Standard error definitions for the event system.
var (
	// ErrQueueClosed is returned when attempting to operate on a closed queue.
	ErrQueueClosed = errors.New("queue is closed")

	// ErrQueueEmpty is returned when attempting to dequeue from an empty queue in non-blocking mode.
	ErrQueueEmpty = errors.New("queue is empty")

	// ErrQueueFull is returned when attempting to enqueue to a full queue.
	ErrQueueFull = errors.New("queue is full")

	// ErrInvalidEvent is returned when an event fails validation.
	ErrInvalidEvent = errors.New("invalid event")

	// ErrConsumerClosed is returned when attempting to consume from a closed consumer.
	ErrConsumerClosed = errors.New("consumer is closed")
)

// NoTaskQueueError represents an error when a task queue is not found.
type NoTaskQueueError struct {
	TaskID string
}

// Error returns the error message.
func (e *NoTaskQueueError) Error() string {
	return fmt.Sprintf("no task queue found for task ID: %s", e.TaskID)
}

// Is implements error matching for NoTaskQueueError.
func (e *NoTaskQueueError) Is(target error) bool {
	_, ok := target.(*NoTaskQueueError)
	return ok
}

// TaskQueueExistsError represents an error when trying to create a task queue that already exists.
type TaskQueueExistsError struct {
	TaskID string
}

// Error returns the error message.
func (e *TaskQueueExistsError) Error() string {
	return fmt.Sprintf("task queue already exists for task ID: %s", e.TaskID)
}

// Is implements error matching for TaskQueueExistsError.
func (e *TaskQueueExistsError) Is(target error) bool {
	_, ok := target.(*TaskQueueExistsError)
	return ok
}

// ServerError represents a server-side error during event processing.
type ServerError struct {
	Message string
	Cause   error
}

// Error returns the error message.
func (e *ServerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("server error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("server error: %s", e.Message)
}

// Unwrap returns the underlying cause of the error.
func (e *ServerError) Unwrap() error {
	return e.Cause
}

// Is implements error matching for ServerError.
func (e *ServerError) Is(target error) bool {
	_, ok := target.(*ServerError)
	return ok
}

// NewServerError creates a new ServerError with the given message.
func NewServerError(message string) *ServerError {
	return &ServerError{Message: message}
}

// NewServerErrorWithCause creates a new ServerError with the given message and cause.
func NewServerErrorWithCause(message string, cause error) *ServerError {
	return &ServerError{Message: message, Cause: cause}
}

// EventProcessingError represents an error during event processing.
type EventProcessingError struct {
	Event   Event
	Message string
	Cause   error
}

// Error returns the error message.
func (e *EventProcessingError) Error() string {
	eventStr := "nil"
	if e.Event != nil {
		eventStr = e.Event.String()
	}

	if e.Cause != nil {
		return fmt.Sprintf("event processing error for %s: %s: %v", eventStr, e.Message, e.Cause)
	}
	return fmt.Sprintf("event processing error for %s: %s", eventStr, e.Message)
}

// Unwrap returns the underlying cause of the error.
func (e *EventProcessingError) Unwrap() error {
	return e.Cause
}

// Is implements error matching for EventProcessingError.
func (e *EventProcessingError) Is(target error) bool {
	_, ok := target.(*EventProcessingError)
	return ok
}

// NewEventProcessingError creates a new EventProcessingError.
func NewEventProcessingError(event Event, message string, cause error) *EventProcessingError {
	return &EventProcessingError{
		Event:   event,
		Message: message,
		Cause:   cause,
	}
}

// QueueOperationError represents an error during queue operations.
type QueueOperationError struct {
	Operation string
	TaskID    string
	Message   string
	Cause     error
}

// Error returns the error message.
func (e *QueueOperationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("queue operation error (%s) for task %s: %s: %v",
			e.Operation, e.TaskID, e.Message, e.Cause)
	}
	return fmt.Sprintf("queue operation error (%s) for task %s: %s",
		e.Operation, e.TaskID, e.Message)
}

// Unwrap returns the underlying cause of the error.
func (e *QueueOperationError) Unwrap() error {
	return e.Cause
}

// Is implements error matching for QueueOperationError.
func (e *QueueOperationError) Is(target error) bool {
	_, ok := target.(*QueueOperationError)
	return ok
}

// NewQueueOperationError creates a new QueueOperationError.
func NewQueueOperationError(operation, taskID, message string, cause error) *QueueOperationError {
	return &QueueOperationError{
		Operation: operation,
		TaskID:    taskID,
		Message:   message,
		Cause:     cause,
	}
}

// IsQueueClosedError checks if an error is related to a closed queue.
func IsQueueClosedError(err error) bool {
	return errors.Is(err, ErrQueueClosed)
}

// IsQueueEmptyError checks if an error is related to an empty queue.
func IsQueueEmptyError(err error) bool {
	return errors.Is(err, ErrQueueEmpty)
}

// IsQueueFullError checks if an error is related to a full queue.
func IsQueueFullError(err error) bool {
	return errors.Is(err, ErrQueueFull)
}

// IsNoTaskQueueError checks if an error is a NoTaskQueueError.
func IsNoTaskQueueError(err error) bool {
	var noTaskQueueErr *NoTaskQueueError
	return errors.As(err, &noTaskQueueErr)
}

// IsTaskQueueExistsError checks if an error is a TaskQueueExistsError.
func IsTaskQueueExistsError(err error) bool {
	var taskQueueExistsErr *TaskQueueExistsError
	return errors.As(err, &taskQueueExistsErr)
}

// IsServerError checks if an error is a ServerError.
func IsServerError(err error) bool {
	var serverErr *ServerError
	return errors.As(err, &serverErr)
}

// IsEventProcessingError checks if an error is an EventProcessingError.
func IsEventProcessingError(err error) bool {
	var eventProcessingErr *EventProcessingError
	return errors.As(err, &eventProcessingErr)
}

// IsQueueOperationError checks if an error is a QueueOperationError.
func IsQueueOperationError(err error) bool {
	var queueOperationErr *QueueOperationError
	return errors.As(err, &queueOperationErr)
}
