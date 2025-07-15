// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package a2a provides helper utilities for the Agent-to-Agent (A2A) protocol.
// This file contains general utility functions that mirror the Python A2A helpers,
// converted to idiomatic Go code with proper error handling and logging support.
package a2a

import (
	"fmt"
	"log"
)

// CreateTaskObj creates a new task object from message send params.
//
// Generates UUIDs for task and context IDs if they are not already present in the message.
// This function mirrors the Python create_task_obj function behavior.
//
// Args:
//
//	messageSendParams: The MessageSendParams object containing the initial message.
//
// Returns:
//
//	A new Task object initialized with 'submitted' status and the input message in history.
//	Returns an error if UUID generation fails or if the input is invalid.
func CreateTaskObj(messageSendParams *MessageSendParams) (*Task, error) {
	if messageSendParams == nil {
		return nil, fmt.Errorf("message send params cannot be nil")
	}

	if err := messageSendParams.Validate(); err != nil {
		return nil, fmt.Errorf("invalid message send params: %w", err)
	}

	// Generate context ID if not present
	if messageSendParams.Message.ContextID == "" {
		contextID, err := generateUUID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate context ID: %w", err)
		}
		messageSendParams.Message.ContextID = contextID
	}

	// Generate task ID
	taskID, err := generateUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate task ID: %w", err)
	}

	// Create new task with submitted status
	task := &Task{
		ID:        taskID,
		ContextID: messageSendParams.Message.ContextID,
		Status: TaskStatus{
			State: TaskStateSubmitted,
		},
		History: []*Message{messageSendParams.Message},
	}

	return task, nil
}

// AppendArtifactToTask updates a Task object with new artifact data from an event.
//
// Handles creating the artifacts list if it doesn't exist, adding new artifacts,
// and appending parts to existing artifacts based on the `append` flag in the event.
// This function mirrors the Python append_artifact_to_task function behavior.
//
// Args:
//
//	task: The Task object to modify.
//	event: The TaskArtifactUpdateEvent containing the artifact data.
//
// Returns:
//
//	An error if the input is invalid or if the operation fails.
func AppendArtifactToTask(task *Task, event *TaskArtifactUpdateEvent) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}

	// Initialize artifacts slice if it doesn't exist
	if task.Artifacts == nil {
		task.Artifacts = []*Artifact{}
	}

	newArtifactData := event.Artifact
	artifactID := newArtifactData.ArtifactID
	appendParts := event.Append

	var existingArtifact *Artifact
	var existingArtifactIndex int = -1

	// Find existing artifact by its ID
	for i, artifact := range task.Artifacts {
		if artifact.ArtifactID == artifactID {
			existingArtifact = artifact
			existingArtifactIndex = i
			break
		}
	}

	if !appendParts {
		// This represents the first chunk for this artifact ID.
		if existingArtifactIndex != -1 {
			// Replace the existing artifact entirely with the new data
			log.Printf("Replacing artifact at id %s for task %s", artifactID, task.ID)
			task.Artifacts[existingArtifactIndex] = newArtifactData
		} else {
			// Append the new artifact since no artifact with this ID exists yet
			log.Printf("Adding new artifact with id %s for task %s", artifactID, task.ID)
			task.Artifacts = append(task.Artifacts, newArtifactData)
		}
	} else if existingArtifact != nil {
		// Append new parts to the existing artifact's part list
		log.Printf("Appending parts to artifact id %s for task %s", artifactID, task.ID)
		existingArtifact.Parts = append(existingArtifact.Parts, newArtifactData.Parts...)
	} else {
		// We received a chunk to append, but we don't have an existing artifact.
		// We will ignore this chunk
		log.Printf("Received append=true for nonexistent artifact id %s in task %s. Ignoring chunk.", artifactID, task.ID)
	}

	return nil
}
