// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package a2a provides task utilities for the Agent-to-Agent (A2A) protocol.
// This file contains utility functions for creating and managing A2A Task objects,
// converted from Python to idiomatic Go code with proper error handling and validation.
package a2a

import (
	"fmt"
)

// TaskMessageInput represents the input interface for creating tasks.
// It supports both Message and A2AMessage types by providing access to
// the necessary fields for task creation.
type TaskMessageInput interface {
	GetTaskID() *string
	GetContextID() *string
	AsMessage() *Message
}

// MessageTaskInput wraps a Message to implement TaskMessageInput.
type MessageTaskInput struct {
	Message   *Message
	TaskID    *string
	ContextID *string
}

// GetTaskID returns the task ID.
func (m MessageTaskInput) GetTaskID() *string {
	return m.TaskID
}

// GetContextID returns the context ID.
func (m MessageTaskInput) GetContextID() *string {
	return m.ContextID
}

// AsMessage returns the wrapped message.
func (m MessageTaskInput) AsMessage() *Message {
	return m.Message
}

// A2AMessageTaskInput wraps an A2AMessage to implement TaskMessageInput.
type A2AMessageTaskInput struct {
	A2AMessage *A2AMessage
}

// GetTaskID returns the task ID from the A2AMessage.
func (a A2AMessageTaskInput) GetTaskID() *string {
	return a.A2AMessage.TaskID
}

// GetContextID returns the context ID from the A2AMessage.
func (a A2AMessageTaskInput) GetContextID() *string {
	return a.A2AMessage.ContextID
}

// AsMessage converts the A2AMessage to a Message.
func (a A2AMessageTaskInput) AsMessage() *Message {
	if a.A2AMessage == nil {
		return nil
	}

	// Extract text content from A2AMessage parts
	content := GetMessageTextWithNewlines(a.A2AMessage)

	// Create basic Message from A2AMessage
	message := &Message{
		Content: content,
	}

	// Set ContextID if available
	if a.A2AMessage.ContextID != nil {
		message.ContextID = *a.A2AMessage.ContextID
	}

	return message
}

// NewTask creates a new Task object from a message input.
//
// This function mirrors the Python new_task function behavior, accepting either
// a Message or A2AMessage through the TaskMessageInput interface.
//
// If taskId or contextId are not provided in the request message, it generates
// new UUIDs for them. The created Task is initialized with a TaskStatus of
// "submitted" and includes the input request message in its history.
//
// Args:
//
//	request: The message input that implements TaskMessageInput interface.
//
// Returns:
//
//	A new Task object with "submitted" status and an error if the operation fails.
func NewTask(request TaskMessageInput) (*Task, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	message := request.AsMessage()
	if message == nil {
		return nil, fmt.Errorf("request message cannot be nil")
	}

	// Validate the message
	if err := message.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request message: %w", err)
	}

	// Get or generate task ID
	var taskID string
	if reqTaskID := request.GetTaskID(); reqTaskID != nil && *reqTaskID != "" {
		taskID = *reqTaskID
	} else {
		id, err := generateUUID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate task ID: %w", err)
		}
		taskID = id
	}

	// Get or generate context ID
	var contextID string
	if reqContextID := request.GetContextID(); reqContextID != nil && *reqContextID != "" {
		contextID = *reqContextID
	} else {
		id, err := generateUUID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate context ID: %w", err)
		}
		contextID = id
	}

	// Create task with submitted status
	task := &Task{
		ID:        taskID,
		ContextID: contextID,
		Status: TaskStatus{
			State: TaskStateSubmitted,
		},
		History: []*Message{message},
	}

	return task, nil
}

// NewTaskFromMessage creates a new Task object from a Message.
//
// This is a convenience function that wraps the Message in a MessageTaskInput
// and calls NewTask.
//
// Args:
//
//	message: The Message object.
//	taskID: Optional task ID (if nil, a new UUID will be generated).
//	contextID: Optional context ID (if nil, a new UUID will be generated).
//
// Returns:
//
//	A new Task object with "submitted" status and an error if the operation fails.
func NewTaskFromMessage(message *Message, taskID *string, contextID *string) (*Task, error) {
	if message == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	input := MessageTaskInput{
		Message:   message,
		TaskID:    taskID,
		ContextID: contextID,
	}

	return NewTask(input)
}

// NewTaskFromA2AMessage creates a new Task object from an A2AMessage.
//
// This is a convenience function that wraps the A2AMessage in an A2AMessageTaskInput
// and calls NewTask.
//
// Args:
//
//	a2aMessage: The A2AMessage object.
//
// Returns:
//
//	A new Task object with "submitted" status and an error if the operation fails.
func NewTaskFromA2AMessage(a2aMessage *A2AMessage) (*Task, error) {
	if a2aMessage == nil {
		return nil, fmt.Errorf("a2aMessage cannot be nil")
	}

	input := A2AMessageTaskInput{
		A2AMessage: a2aMessage,
	}

	return NewTask(input)
}

// CompletedTask creates a Task object in the "completed" state.
//
// This function mirrors the Python completed_task function behavior.
// It creates a Task with the "completed" status, the provided task_id,
// context_id, and artifacts. An optional history list of Message objects
// can also be provided.
//
// This function is used to construct the final representation of a task
// once an agent finishes its work.
//
// Args:
//
//	taskID: The task ID.
//	contextID: The context ID.
//	artifacts: List of Artifact objects produced by the task.
//	history: Optional list of Message objects representing the task history.
//
// Returns:
//
//	A new Task object with "completed" status and an error if the operation fails.
func CompletedTask(taskID, contextID string, artifacts []*Artifact, history []*Message) (*Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}
	if contextID == "" {
		return nil, fmt.Errorf("context ID cannot be empty")
	}
	if artifacts == nil {
		return nil, fmt.Errorf("artifacts cannot be nil")
	}

	// Validate artifacts
	for i, artifact := range artifacts {
		if artifact == nil {
			return nil, fmt.Errorf("artifact at index %d cannot be nil", i)
		}
		if err := artifact.Validate(); err != nil {
			return nil, fmt.Errorf("artifact at index %d is invalid: %w", i, err)
		}
	}

	// Use empty history if not provided
	if history == nil {
		history = []*Message{}
	}

	// Validate history messages
	for i, message := range history {
		if message == nil {
			return nil, fmt.Errorf("history message at index %d cannot be nil", i)
		}
		if err := message.Validate(); err != nil {
			return nil, fmt.Errorf("history message at index %d is invalid: %w", i, err)
		}
	}

	// Create task with completed status
	task := &Task{
		ID:        taskID,
		ContextID: contextID,
		Status: TaskStatus{
			State: TaskStateCompleted,
		},
		History:   history,
		Artifacts: artifacts,
	}

	return task, nil
}

// CompletedTaskWithNoHistory creates a Task object in the "completed" state with no history.
//
// This is a convenience function that calls CompletedTask with an empty history.
//
// Args:
//
//	taskID: The task ID.
//	contextID: The context ID.
//	artifacts: List of Artifact objects produced by the task.
//
// Returns:
//
//	A new Task object with "completed" status and an error if the operation fails.
func CompletedTaskWithNoHistory(taskID, contextID string, artifacts []*Artifact) (*Task, error) {
	return CompletedTask(taskID, contextID, artifacts, nil)
}
