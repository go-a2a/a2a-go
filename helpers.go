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
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
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
	// Generate context ID if not present
	if messageSendParams.Message.ContextID == "" {
		messageSendParams.Message.ContextID = uuid.NewString()
	}

	// Create new task with submitted status
	task := &Task{
		ID:        uuid.NewString(),
		ContextID: messageSendParams.Message.ContextID,
		Status: TaskStatus{
			State:     TaskStateSubmitted,
			Timestamp: time.Now().Format(time.RFC3339),
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
func AppendArtifactToTask(ctx context.Context, task *Task, event *TaskArtifactUpdateEvent) error {
	logger := slog.Default()

	// Initialize artifacts slice if it doesn't exist
	if task.Artifacts == nil {
		task.Artifacts = []*Artifact{}
	}

	newArtifactData := event.Artifact
	artifactID := newArtifactData.ArtifactID
	appendParts := event.Append

	var existingArtifact *Artifact
	existingArtifactIndex := -1

	// Find existing artifact by its ID
	for i, artifact := range task.Artifacts {
		if artifact.ArtifactID == artifactID {
			existingArtifact = artifact
			existingArtifactIndex = i
			break
		}
	}

	switch {
	case !appendParts:
		switch existingArtifactIndex {
		case -1:
			// Append the new artifact since no artifact with this ID exists yet
			logger.InfoContext(ctx, "adding new artifact with id for task", slog.String("artifact_id", artifactID), slog.String("task_id", task.ID))
			task.Artifacts = append(task.Artifacts, newArtifactData)
		default:
			// Replace the existing artifact entirely with the new data
			logger.Info("Replacing artifact at id for task", slog.String("artifact_id", artifactID), slog.String("task_id", task.ID))
			task.Artifacts[existingArtifactIndex] = newArtifactData
		}
	case existingArtifact != nil:
		// Append new parts to the existing artifact's part list
		logger.Info("Appending parts to artifact id for task", slog.String("artifact_id", artifactID), slog.String("task_id", task.ID))
		existingArtifact.Parts = append(existingArtifact.Parts, newArtifactData.Parts...)
	default:
		// We received a chunk to append, but we don't have an existing artifact.
		// We will ignore this chunk
		logger.Info("Received append=true for nonexistent artifact id in task. Ignoring chunk", slog.String("artifact_id", artifactID), slog.String("task_id", task.ID))
	}

	return nil
}
