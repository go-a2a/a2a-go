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
	"strings"
	"testing"
)

func TestCreateTaskObj(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		messageSendParams  *MessageSendParams
		wantErr            bool
		wantErrContains    string
		validateContextID  bool
		validateTaskID     bool
		validateStatus     bool
		validateHistoryLen int
	}{
		"valid input with all fields": {
			messageSendParams: &MessageSendParams{
				Message: &Message{
					ContextID: "test-context-id",
					Parts: []Part{
						&TextPart{
							Kind: TextPartKind,
							Text: "test message content",
						},
					},
					Kind: MessageEventKind,
					Metadata: map[string]any{
						"key": "value",
					},
				},
			},
			wantErr:            false,
			validateContextID:  true,
			validateTaskID:     true,
			validateStatus:     true,
			validateHistoryLen: 1,
		},
		"valid input with empty context ID (should generate one)": {
			messageSendParams: &MessageSendParams{
				Message: &Message{
					ContextID: "",
					Kind:      MessageEventKind,
					Parts: []Part{
						&TextPart{
							Kind: TextPartKind,
							Text: "test message content",
						},
					},
				},
			},
			wantErr:            false,
			validateContextID:  true,
			validateTaskID:     true,
			validateStatus:     true,
			validateHistoryLen: 1,
		},
		"valid input with minimal message": {
			messageSendParams: &MessageSendParams{
				Message: &Message{
					Parts: []Part{
						&TextPart{
							Kind: TextPartKind,
							Text: "minimal message",
						},
					},
					Kind: MessageEventKind,
				},
			},
			wantErr:            false,
			validateContextID:  true,
			validateTaskID:     true,
			validateStatus:     true,
			validateHistoryLen: 1,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			task, err := CreateTaskObj(tt.messageSendParams)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateTaskObj() expected error but got none")
					return
				}
				if tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("CreateTaskObj() error = %v, want error containing %v", err, tt.wantErrContains)
				}
				return
			}

			if err != nil {
				t.Errorf("CreateTaskObj() unexpected error = %v", err)
				return
			}

			if task == nil {
				t.Errorf("CreateTaskObj() returned nil task")
				return
			}

			if tt.validateTaskID {
				if task.ID == "" {
					t.Errorf("CreateTaskObj() task ID is empty")
				}
				// Basic UUID format validation (36 characters with hyphens)
				if len(task.ID) != 36 || strings.Count(task.ID, "-") != 4 {
					t.Errorf("CreateTaskObj() task ID format invalid: %s", task.ID)
				}
			}

			if tt.validateContextID {
				if task.ContextID == "" {
					t.Errorf("CreateTaskObj() context ID is empty")
				}
				// If original message had context ID, it should be preserved
				if tt.messageSendParams.Message.ContextID != "" {
					if task.ContextID != tt.messageSendParams.Message.ContextID {
						t.Errorf("CreateTaskObj() context ID = %v, want %v", task.ContextID, tt.messageSendParams.Message.ContextID)
					}
				} else {
					// If original message had no context ID, one should be generated
					if len(task.ContextID) != 36 || strings.Count(task.ContextID, "-") != 4 {
						t.Errorf("CreateTaskObj() context ID format invalid: %s", task.ContextID)
					}
				}
			}

			if tt.validateStatus {
				if task.Status.State != TaskStateSubmitted {
					t.Errorf("CreateTaskObj() task status = %v, want %v", task.Status.State, TaskStateSubmitted)
				}
			}

			if tt.validateHistoryLen >= 0 {
				if len(task.History) != tt.validateHistoryLen {
					t.Errorf("CreateTaskObj() history length = %d, want %d", len(task.History), tt.validateHistoryLen)
				}
				if tt.validateHistoryLen > 0 {
					if task.History[0] != tt.messageSendParams.Message {
						t.Errorf("CreateTaskObj() history[0] is not the original message")
					}
				}
			}
		})
	}
}

func TestAppendArtifactToTask(t *testing.T) {
	t.Parallel()

	// Helper function to create test artifacts
	createTestArtifact := func(id, name, text string) *Artifact {
		artifact := NewTextArtifact(name, text, "test description")
		artifact.ArtifactID = id // Override the generated ID for testing
		return artifact
	}

	// Helper function to create test task
	createTestTask := func() *Task {
		return &Task{
			ID:        "test-task-id",
			ContextID: "test-context-id",
			Status: TaskStatus{
				State: TaskStateSubmitted,
			},
			History:   []*Message{},
			Artifacts: []*Artifact{},
		}
	}

	tests := map[string]struct {
		task                *Task
		event               *TaskArtifactUpdateEvent
		wantErr             bool
		wantErrContains     string
		validateArtifacts   bool
		expectedArtifactLen int
		expectedArtifactID  string
		expectedPartsLen    int
	}{
		"add new artifact to empty task": {
			task: createTestTask(),
			event: &TaskArtifactUpdateEvent{
				Artifact: createTestArtifact("new-artifact", "test-artifact", "test content"),
				Kind:     ArtifactUpdateEventKind,
				Append:   false,
			},
			wantErr:             false,
			validateArtifacts:   true,
			expectedArtifactLen: 1,
			expectedArtifactID:  "new-artifact",
			expectedPartsLen:    1,
		},
		"replace existing artifact (append=false)": {
			task: func() *Task {
				task := createTestTask()
				task.Artifacts = []*Artifact{
					createTestArtifact("existing-artifact", "old-artifact", "old content"),
				}
				return task
			}(),
			event: &TaskArtifactUpdateEvent{
				Artifact: createTestArtifact("existing-artifact", "new-artifact", "new content"),
				Kind:     ArtifactUpdateEventKind,
				Append:   false,
			},
			wantErr:             false,
			validateArtifacts:   true,
			expectedArtifactLen: 1,
			expectedArtifactID:  "existing-artifact",
			expectedPartsLen:    1,
		},
		"append parts to existing artifact (append=true)": {
			task: func() *Task {
				task := createTestTask()
				task.Artifacts = []*Artifact{
					createTestArtifact("existing-artifact", "base-artifact", "base content"),
				}
				return task
			}(),
			event: &TaskArtifactUpdateEvent{
				Artifact: createTestArtifact("existing-artifact", "append-artifact", "append content"),
				Kind:     ArtifactUpdateEventKind,
				Append:   true,
			},
			wantErr:             false,
			validateArtifacts:   true,
			expectedArtifactLen: 1,
			expectedArtifactID:  "existing-artifact",
			expectedPartsLen:    2, // Original + appended
		},
		"append to non-existent artifact (should be ignored)": {
			task: createTestTask(),
			event: &TaskArtifactUpdateEvent{
				Artifact: createTestArtifact("non-existent", "test-artifact", "test content"),
				Kind:     ArtifactUpdateEventKind,
				Append:   true,
			},
			wantErr:             false,
			validateArtifacts:   true,
			expectedArtifactLen: 0, // Should remain empty
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := AppendArtifactToTask(t.Context(), tt.task, tt.event)

			if tt.wantErr {
				if err == nil {
					t.Errorf("AppendArtifactToTask() expected error but got none")
					return
				}
				if tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("AppendArtifactToTask() error = %v, want error containing %v", err, tt.wantErrContains)
				}
				return
			}

			if err != nil {
				t.Errorf("AppendArtifactToTask() unexpected error = %v", err)
				return
			}

			if tt.validateArtifacts {
				if len(tt.task.Artifacts) != tt.expectedArtifactLen {
					t.Errorf("AppendArtifactToTask() artifacts length = %d, want %d", len(tt.task.Artifacts), tt.expectedArtifactLen)
				}

				if tt.expectedArtifactLen > 0 {
					foundArtifact := false
					for _, artifact := range tt.task.Artifacts {
						if artifact.ArtifactID == tt.expectedArtifactID {
							foundArtifact = true
							if len(artifact.Parts) != tt.expectedPartsLen {
								t.Errorf("AppendArtifactToTask() artifact parts length = %d, want %d", len(artifact.Parts), tt.expectedPartsLen)
							}
							break
						}
					}
					if !foundArtifact {
						t.Errorf("AppendArtifactToTask() expected artifact with ID %s not found", tt.expectedArtifactID)
					}
				}
			}
		})
	}
}

func TestAppendArtifactToTask_InitializesArtifactsSlice(t *testing.T) {
	t.Parallel()

	// Test that artifacts slice is initialized if nil
	task := &Task{
		ID:        "test-task-id",
		ContextID: "test-context-id",
		Status: TaskStatus{
			State: TaskStateSubmitted,
		},
		History:   []*Message{},
		Artifacts: nil, // Explicitly set to nil
	}

	artifact := NewTextArtifact("test-artifact", "test content", "test description")

	event := &TaskArtifactUpdateEvent{
		Artifact: artifact,
		Kind:     ArtifactUpdateEventKind,
		Append:   false,
	}

	err := AppendArtifactToTask(t.Context(), task, event)
	if err != nil {
		t.Errorf("AppendArtifactToTask() unexpected error = %v", err)
	}

	if task.Artifacts == nil {
		t.Errorf("AppendArtifactToTask() artifacts slice was not initialized")
	}

	if len(task.Artifacts) != 1 {
		t.Errorf("AppendArtifactToTask() artifacts length = %d, want 1", len(task.Artifacts))
	}
}

// Benchmark tests for performance
func BenchmarkCreateTaskObj(b *testing.B) {
	messageSendParams := &MessageSendParams{
		Message: &Message{
			ContextID: "test-context-id",
			Parts: []Part{
				&TextPart{
					Kind: TextPartKind,
					Text: "test message content",
				},
			},
			Kind: MessageEventKind,
		},
	}

	for b.Loop() {
		_, err := CreateTaskObj(messageSendParams)
		if err != nil {
			b.Fatalf("CreateTaskObj() error = %v", err)
		}
	}
}

func BenchmarkAppendArtifactToTask(b *testing.B) {
	task := &Task{
		ID:        "test-task-id",
		ContextID: "test-context-id",
		Status: TaskStatus{
			State: TaskStateSubmitted,
		},
		History:   []*Message{},
		Artifacts: []*Artifact{},
	}

	artifact := NewTextArtifact("test-artifact", "test content", "test description")

	event := &TaskArtifactUpdateEvent{
		Artifact: artifact,
		Kind:     ArtifactUpdateEventKind,
		Append:   false,
	}

	for b.Loop() {
		// Reset task artifacts for each iteration
		task.Artifacts = []*Artifact{}
		err := AppendArtifactToTask(b.Context(), task, event)
		if err != nil {
			b.Fatalf("AppendArtifactToTask() error = %v", err)
		}
	}
}
