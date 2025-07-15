// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"fmt"

	"github.com/go-a2a/a2a"
)

// ServerError represents an error that occurred during request handling.
// This is the base type for all handler-related errors.
type ServerError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implements the error interface.
func (e ServerError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// GetCode returns the error code.
func (e ServerError) GetCode() int {
	return e.Code
}

// GetMessage returns the error message.
func (e ServerError) GetMessage() string {
	return e.Message
}

// NewServerError creates a new ServerError.
func NewServerError(code int, message string, details ...string) *ServerError {
	err := &ServerError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, message string, args ...any) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: fmt.Sprintf(message, args...),
	}
}

// TaskNotFoundError represents an error when a task is not found.
type TaskNotFoundError struct {
	TaskID string
}

// Error implements the error interface.
func (e TaskNotFoundError) Error() string {
	return fmt.Sprintf("task not found: %s", e.TaskID)
}

// GetCode returns the A2A protocol error code.
func (e TaskNotFoundError) GetCode() int {
	return a2a.ErrorCodeTaskNotFound
}

// GetMessage returns the error message.
func (e TaskNotFoundError) GetMessage() string {
	return e.Error()
}

// NewTaskNotFoundError creates a new TaskNotFoundError.
func NewTaskNotFoundError(taskID string) *TaskNotFoundError {
	return &TaskNotFoundError{TaskID: taskID}
}

// TaskNotCancelableError represents an error when a task cannot be canceled.
type TaskNotCancelableError struct {
	TaskID string
	Reason string
}

// Error implements the error interface.
func (e TaskNotCancelableError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("task %s cannot be canceled: %s", e.TaskID, e.Reason)
	}
	return fmt.Sprintf("task %s cannot be canceled", e.TaskID)
}

// GetCode returns the A2A protocol error code.
func (e TaskNotCancelableError) GetCode() int {
	return a2a.ErrorCodeTaskNotCancelable
}

// GetMessage returns the error message.
func (e TaskNotCancelableError) GetMessage() string {
	return e.Error()
}

// NewTaskNotCancelableError creates a new TaskNotCancelableError.
func NewTaskNotCancelableError(taskID string, reason ...string) *TaskNotCancelableError {
	err := &TaskNotCancelableError{TaskID: taskID}
	if len(reason) > 0 {
		err.Reason = reason[0]
	}
	return err
}

// InvalidRequestError represents an error for invalid requests.
type InvalidRequestError struct {
	Details string
}

// Error implements the error interface.
func (e InvalidRequestError) Error() string {
	return fmt.Sprintf("invalid request: %s", e.Details)
}

// GetCode returns the A2A protocol error code.
func (e InvalidRequestError) GetCode() int {
	return a2a.ErrorCodeInvalidRequest
}

// GetMessage returns the error message.
func (e InvalidRequestError) GetMessage() string {
	return e.Error()
}

// NewInvalidRequestError creates a new InvalidRequestError.
func NewInvalidRequestError(details string) *InvalidRequestError {
	return &InvalidRequestError{Details: details}
}

// InternalError represents an internal server error.
type InternalError struct {
	Details string
}

// Error implements the error interface.
func (e InternalError) Error() string {
	return fmt.Sprintf("internal error: %s", e.Details)
}

// GetCode returns the A2A protocol error code.
func (e InternalError) GetCode() int {
	return a2a.ErrorCodeInternalError
}

// GetMessage returns the error message.
func (e InternalError) GetMessage() string {
	return e.Error()
}

// NewInternalError creates a new InternalError.
func NewInternalError(details string) *InternalError {
	return &InternalError{Details: details}
}

// ExecutionError represents an error during agent execution.
type ExecutionError struct {
	TaskID  string
	Details string
}

// Error implements the error interface.
func (e ExecutionError) Error() string {
	return fmt.Sprintf("execution error for task %s: %s", e.TaskID, e.Details)
}

// GetCode returns the A2A protocol error code.
func (e ExecutionError) GetCode() int {
	return a2a.ErrorCodeInternalError
}

// GetMessage returns the error message.
func (e ExecutionError) GetMessage() string {
	return e.Error()
}

// NewExecutionError creates a new ExecutionError.
func NewExecutionError(taskID, details string) *ExecutionError {
	return &ExecutionError{
		TaskID:  taskID,
		Details: details,
	}
}

// StorageError represents an error in task storage operations.
type StorageError struct {
	Operation string
	TaskID    string
	Details   string
}

// Error implements the error interface.
func (e StorageError) Error() string {
	return fmt.Sprintf("storage error during %s for task %s: %s", e.Operation, e.TaskID, e.Details)
}

// GetCode returns the A2A protocol error code.
func (e StorageError) GetCode() int {
	return a2a.ErrorCodeInternalError
}

// GetMessage returns the error message.
func (e StorageError) GetMessage() string {
	return e.Error()
}

// NewStorageError creates a new StorageError.
func NewStorageError(operation, taskID, details string) *StorageError {
	return &StorageError{
		Operation: operation,
		TaskID:    taskID,
		Details:   details,
	}
}
