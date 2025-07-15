// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package task

import (
	"fmt"

	"github.com/go-a2a/a2a"
)

// TaskNotUpdatableError represents an error when attempting to update a task in a terminal state.
type TaskNotUpdatableError struct {
	TaskID string
	State  a2a.TaskState
}

// Error returns the error message.
func (e TaskNotUpdatableError) Error() string {
	return fmt.Sprintf("task %s in state %s cannot be updated", e.TaskID, e.State)
}

// TaskStoreError represents an error from the task store.
type TaskStoreError struct {
	Operation string
	TaskID    string
	Err       error
}

// Error returns the error message.
func (e TaskStoreError) Error() string {
	return fmt.Sprintf("task store %s operation failed for task %s: %v", e.Operation, e.TaskID, e.Err)
}

// Unwrap returns the underlying error.
func (e TaskStoreError) Unwrap() error {
	return e.Err
}

// TaskValidationError represents an error when task validation fails.
type TaskValidationError struct {
	TaskID string
	Err    error
}

// Error returns the error message.
func (e TaskValidationError) Error() string {
	return fmt.Sprintf("task %s validation failed: %v", e.TaskID, e.Err)
}

// Unwrap returns the underlying error.
func (e TaskValidationError) Unwrap() error {
	return e.Err
}

// TaskManagerError represents an error from the task manager.
type TaskManagerError struct {
	Operation string
	TaskID    string
	Err       error
}

// Error returns the error message.
func (e TaskManagerError) Error() string {
	return fmt.Sprintf("task manager %s operation failed for task %s: %v", e.Operation, e.TaskID, e.Err)
}

// Unwrap returns the underlying error.
func (e TaskManagerError) Unwrap() error {
	return e.Err
}

// TaskUpdaterError represents an error from the task updater.
type TaskUpdaterError struct {
	Operation string
	TaskID    string
	Err       error
}

// Error returns the error message.
func (e TaskUpdaterError) Error() string {
	return fmt.Sprintf("task updater %s operation failed for task %s: %v", e.Operation, e.TaskID, e.Err)
}

// Unwrap returns the underlying error.
func (e TaskUpdaterError) Unwrap() error {
	return e.Err
}

// ResultAggregatorError represents an error from the result aggregator.
type ResultAggregatorError struct {
	Operation string
	TaskID    string
	Err       error
}

// Error returns the error message.
func (e ResultAggregatorError) Error() string {
	return fmt.Sprintf("result aggregator %s operation failed for task %s: %v", e.Operation, e.TaskID, e.Err)
}

// Unwrap returns the underlying error.
func (e ResultAggregatorError) Unwrap() error {
	return e.Err
}

// NewTaskNotUpdatableError creates a new TaskNotUpdatableError.
func NewTaskNotUpdatableError(taskID string, state a2a.TaskState) TaskNotUpdatableError {
	return TaskNotUpdatableError{
		TaskID: taskID,
		State:  state,
	}
}

// NewTaskStoreError creates a new TaskStoreError.
func NewTaskStoreError(operation, taskID string, err error) TaskStoreError {
	return TaskStoreError{
		Operation: operation,
		TaskID:    taskID,
		Err:       err,
	}
}

// NewTaskValidationError creates a new TaskValidationError.
func NewTaskValidationError(taskID string, err error) TaskValidationError {
	return TaskValidationError{
		TaskID: taskID,
		Err:    err,
	}
}

// NewTaskManagerError creates a new TaskManagerError.
func NewTaskManagerError(operation, taskID string, err error) TaskManagerError {
	return TaskManagerError{
		Operation: operation,
		TaskID:    taskID,
		Err:       err,
	}
}

// NewTaskUpdaterError creates a new TaskUpdaterError.
func NewTaskUpdaterError(operation, taskID string, err error) TaskUpdaterError {
	return TaskUpdaterError{
		Operation: operation,
		TaskID:    taskID,
		Err:       err,
	}
}

// NewResultAggregatorError creates a new ResultAggregatorError.
func NewResultAggregatorError(operation, taskID string, err error) ResultAggregatorError {
	return ResultAggregatorError{
		Operation: operation,
		TaskID:    taskID,
		Err:       err,
	}
}
