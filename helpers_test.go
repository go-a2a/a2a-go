// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"strings"
	"testing"
)

func TestCreateTaskObj(t *testing.T) {
	tests := []struct {
		name               string
		messageSendParams  *MessageSendParams
		wantErr            bool
		wantErrContains    string
		validateContextID  bool
		validateTaskID     bool
		validateStatus     bool
		validateHistoryLen int
	}{
		{
			name: "valid input with all fields",
			messageSendParams: &MessageSendParams{
				Message: &Message{
					ContextID: "test-context-id",
					Content:   "test message content",
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
		{
			name: "valid input with empty context ID (should generate one)",
			messageSendParams: &MessageSendParams{
				Message: &Message{
					ContextID: "",
					Content:   "test message content",
				},
			},
			wantErr:            false,
			validateContextID:  true,
			validateTaskID:     true,
			validateStatus:     true,
			validateHistoryLen: 1,
		},
		{
			name: "valid input with minimal message",
			messageSendParams: &MessageSendParams{
				Message: &Message{
					Content: "minimal message",
				},
			},
			wantErr:            false,
			validateContextID:  true,
			validateTaskID:     true,
			validateStatus:     true,
			validateHistoryLen: 1,
		},
		{
			name:              "nil message send params",
			messageSendParams: nil,
			wantErr:           true,
			wantErrContains:   "message send params cannot be nil",
		},
		{
			name: "nil message in params",
			messageSendParams: &MessageSendParams{
				Message: nil,
			},
			wantErr:         true,
			wantErrContains: "message send params message cannot be nil",
		},
		{
			name: "empty message content",
			messageSendParams: &MessageSendParams{
				Message: &Message{
					ContextID: "test-context",
					Content:   "",
				},
			},
			wantErr:         true,
			wantErrContains: "message content cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	// Helper function to create test artifacts
	createTestArtifact := func(id, name, text string) *Artifact {
		artifact, err := NewTextArtifact(name, text, "test description")
		if err != nil {
			t.Fatalf("Failed to create test artifact: %v", err)
		}
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

	tests := []struct {
		name                string
		task                *Task
		event               *TaskArtifactUpdateEvent
		wantErr             bool
		wantErrContains     string
		validateArtifacts   bool
		expectedArtifactLen int
		expectedArtifactID  string
		expectedPartsLen    int
	}{
		{
			name: "add new artifact to empty task",
			task: createTestTask(),
			event: &TaskArtifactUpdateEvent{
				Artifact: createTestArtifact("new-artifact", "test-artifact", "test content"),
				Append:   false,
			},
			wantErr:             false,
			validateArtifacts:   true,
			expectedArtifactLen: 1,
			expectedArtifactID:  "new-artifact",
			expectedPartsLen:    1,
		},
		{
			name: "replace existing artifact (append=false)",
			task: func() *Task {
				task := createTestTask()
				task.Artifacts = []*Artifact{
					createTestArtifact("existing-artifact", "old-artifact", "old content"),
				}
				return task
			}(),
			event: &TaskArtifactUpdateEvent{
				Artifact: createTestArtifact("existing-artifact", "new-artifact", "new content"),
				Append:   false,
			},
			wantErr:             false,
			validateArtifacts:   true,
			expectedArtifactLen: 1,
			expectedArtifactID:  "existing-artifact",
			expectedPartsLen:    1,
		},
		{
			name: "append parts to existing artifact (append=true)",
			task: func() *Task {
				task := createTestTask()
				task.Artifacts = []*Artifact{
					createTestArtifact("existing-artifact", "base-artifact", "base content"),
				}
				return task
			}(),
			event: &TaskArtifactUpdateEvent{
				Artifact: createTestArtifact("existing-artifact", "append-artifact", "append content"),
				Append:   true,
			},
			wantErr:             false,
			validateArtifacts:   true,
			expectedArtifactLen: 1,
			expectedArtifactID:  "existing-artifact",
			expectedPartsLen:    2, // Original + appended
		},
		{
			name: "append to non-existent artifact (should be ignored)",
			task: createTestTask(),
			event: &TaskArtifactUpdateEvent{
				Artifact: createTestArtifact("non-existent", "test-artifact", "test content"),
				Append:   true,
			},
			wantErr:             false,
			validateArtifacts:   true,
			expectedArtifactLen: 0, // Should remain empty
		},
		{
			name: "nil task",
			task: nil,
			event: &TaskArtifactUpdateEvent{
				Artifact: createTestArtifact("test-artifact", "test-artifact", "test content"),
				Append:   false,
			},
			wantErr:         true,
			wantErrContains: "task cannot be nil",
		},
		{
			name:            "nil event",
			task:            createTestTask(),
			event:           nil,
			wantErr:         true,
			wantErrContains: "event cannot be nil",
		},
		{
			name: "invalid event (nil artifact)",
			task: createTestTask(),
			event: &TaskArtifactUpdateEvent{
				Artifact: nil,
				Append:   false,
			},
			wantErr:         true,
			wantErrContains: "task artifact update event artifact cannot be nil",
		},
		{
			name: "invalid event (invalid artifact)",
			task: createTestTask(),
			event: &TaskArtifactUpdateEvent{
				Artifact: &Artifact{
					ArtifactID: "", // Invalid: empty artifact ID
					Parts:      []*PartWrapper{},
				},
				Append: false,
			},
			wantErr:         true,
			wantErrContains: "artifact ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AppendArtifactToTask(tt.task, tt.event)

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

	artifact, err := NewTextArtifact("test-artifact", "test content", "test description")
	if err != nil {
		t.Fatalf("Failed to create test artifact: %v", err)
	}

	event := &TaskArtifactUpdateEvent{
		Artifact: artifact,
		Append:   false,
	}

	err = AppendArtifactToTask(task, event)
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
			Content:   "test message content",
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

	artifact, err := NewTextArtifact("test-artifact", "test content", "test description")
	if err != nil {
		b.Fatalf("Failed to create test artifact: %v", err)
	}

	event := &TaskArtifactUpdateEvent{
		Artifact: artifact,
		Append:   false,
	}

	for b.Loop() {
		// Reset task artifacts for each iteration
		task.Artifacts = []*Artifact{}
		err := AppendArtifactToTask(task, event)
		if err != nil {
			b.Fatalf("AppendArtifactToTask() error = %v", err)
		}
	}
}
