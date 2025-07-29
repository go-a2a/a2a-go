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

package agent_execution

import (
	"fmt"
	"strings"

	"github.com/go-a2a/a2a-go"
)

// NewTextArtifact creates a new artifact containing text content.
func NewTextArtifact(name, text string) *a2a.Artifact {
	return &a2a.Artifact{
		ArtifactID: generateArtifactID(),
		Name:       name,
		Parts:      []a2a.Part{a2a.NewTextPart(text)},
	}
}

// NewFileArtifact creates a new artifact containing file content.
func NewFileArtifact(name string, file a2a.File) *a2a.Artifact {
	return &a2a.Artifact{
		ArtifactID: generateArtifactID(),
		Name:       name,
		Parts: []a2a.Part{&a2a.FilePart{
			Kind: a2a.FilePartKind,
			File: file,
		}},
	}
}

// NewDataArtifact creates a new artifact containing structured data.
func NewDataArtifact(name string, data map[string]any) *a2a.Artifact {
	return &a2a.Artifact{
		ArtifactID: generateArtifactID(),
		Name:       name,
		Parts:      []a2a.Part{a2a.NewDataPart(data)},
	}
}

// NewAgentMessage creates a new message from the agent.
func NewAgentMessage(text string) *a2a.Message {
	return &a2a.Message{
		Role:      a2a.RoleAgent,
		Parts:     []a2a.Part{a2a.NewTextPart(text)},
		MessageID: generateMessageID(),
		Kind:      a2a.MessageEventKind,
	}
}

// ExtractTextFromMessage extracts all text content from a message.
func ExtractTextFromMessage(message *a2a.Message) string {
	var texts []string
	for _, part := range message.Parts {
		if textPart, ok := part.(*a2a.TextPart); ok {
			texts = append(texts, textPart.Text)
		}
	}
	return strings.Join(texts, "\n")
}

// ExtractFilesFromMessage extracts all file parts from a message.
func ExtractFilesFromMessage(message *a2a.Message) []a2a.File {
	var files []a2a.File
	for _, part := range message.Parts {
		if filePart, ok := part.(*a2a.FilePart); ok {
			files = append(files, filePart.File)
		}
	}
	return files
}

// ExtractDataFromMessage extracts all data parts from a message.
func ExtractDataFromMessage(message *a2a.Message) []map[string]any {
	var data []map[string]any
	for _, part := range message.Parts {
		if dataPart, ok := part.(*a2a.DataPart); ok {
			data = append(data, dataPart.Data)
		}
	}
	return data
}

// CreateErrorArtifact creates an artifact containing error information.
func CreateErrorArtifact(err error) *a2a.Artifact {
	return &a2a.Artifact{
		ArtifactID:  generateArtifactID(),
		Name:        "error",
		Description: "Error information",
		Parts: []a2a.Part{
			a2a.NewDataPart(map[string]any{
				"error":   err.Error(),
				"type":    "error",
				"details": fmt.Sprintf("%+v", err),
			}),
		},
	}
}

// IsInputRequired checks if a task is in the input-required state.
func IsInputRequired(task *a2a.Task) bool {
	return task.Status.State == a2a.TaskStateInputRequired
}

// IsTerminalState checks if a task is in a terminal state.
func IsTerminalState(state a2a.TaskState) bool {
	switch state {
	case a2a.TaskStateCompleted,
		a2a.TaskStateFailed,
		a2a.TaskStateCanceled,
		a2a.TaskStateRejected,
		a2a.TaskStateUnknown:
		return true
	default:
		return false
	}
}

// generateArtifactID generates a unique artifact ID.
// In a production system, this would use a proper UUID generator.
func generateArtifactID() string {
	return fmt.Sprintf("artifact_%d", generateID())
}

// generateMessageID generates a unique message ID.
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", generateID())
}

// Simple ID generator for demonstration.
// In production, use a proper UUID library.
var idCounter int64

func generateID() int64 {
	idCounter++
	return idCounter
}