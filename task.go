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

package a2a

import (
	"cmp"
	"errors"
	"time"

	"github.com/google/uuid"
)

var now = time.Now

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
func NewTask(message *Message) (*Task, error) {
	if message == nil {
		return nil, errors.New("message cannot be nil")
	}

	if message.Role == "" {
		return nil, errors.New("message role cannot be empty")
	}

	if len(message.Parts) == 0 {
		return nil, errors.New("message parts cannot be empty")
	}

	// Get or generate task ID
	id := cmp.Or(message.TaskID, uuid.NewString())

	// Get or generate context ID
	contextID := cmp.Or(message.ContextID, uuid.NewString())

	// Create task with submitted status
	task := &Task{
		Status: TaskStatus{
			State:     TaskStateSubmitted,
			Timestamp: now().Format(time.RFC3339),
		},
		ID:        id,
		ContextID: contextID,
		History:   []*Message{message},
		Kind:      TaskEventKind,
	}

	return task, nil
}

// CompletedTask creates a Task object in the "completed" state.
//
// Useful for constructing a final Task representation when the agent
// finishes and produces artifacts.
//
// This function mirrors the Python completed_task function behavior.
//
// It creates a Task with the "completed" status, the provided taskID,
// contextID, and artifacts. An optional history list of Message objects
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
	// Use empty history if not provided
	if history == nil {
		history = []*Message{}
	}

	// Create task with completed status
	task := &Task{
		Status: TaskStatus{
			State:     TaskStateCompleted,
			Timestamp: now().Format(time.RFC3339),
		},
		ID:        taskID,
		ContextID: contextID,
		Artifacts: artifacts,
		History:   history,
		Kind:      TaskEventKind,
	}

	return task, nil
}
