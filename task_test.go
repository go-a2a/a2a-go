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
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestMain(m *testing.M) {
	saveNow := now
	now = func() time.Time {
		return time.Date(2025, time.July, 24, 0, 0, 0, 0, time.UTC)
	}
	m.Run()
	now = saveNow
}

// Helper function to create a basic valid message for testing
func createTaskTestMessage(role Role, text string) *Message {
	return &Message{
		Kind:      MessageEventKind,
		MessageID: uuid.NewString(),
		Role:      role,
		Parts: []Part{
			NewTextPart(text),
		},
	}
}

// Helper function to create a message with task and context IDs
func createTaskTestMessageWithIDs(role Role, text, taskID, contextID string) *Message {
	msg := createTaskTestMessage(role, text)
	msg.TaskID = taskID
	msg.ContextID = contextID
	return msg
}

// Helper function to validate UUID format
func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func TestNewTask(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		request *Message
		want    *Task
		wantErr bool
		errMsg  string
	}{
		"nil request": {
			request: nil,
			want:    nil,
			wantErr: true,
			errMsg:  "request cannot be nil",
		},
		"empty role": {
			request: &Message{
				Kind:      MessageEventKind,
				MessageID: uuid.NewString(),
				Role:      "",
				Parts: []Part{
					NewTextPart("test message"),
				},
			},
			want:    nil,
			wantErr: true,
			errMsg:  "message role cannot be empty",
		},
		"empty parts": {
			request: &Message{
				Kind:      MessageEventKind,
				MessageID: uuid.NewString(),
				Role:      RoleUser,
				Parts:     []Part{},
			},
			want:    nil,
			wantErr: true,
			errMsg:  "message parts cannot be empty",
		},
		"nil parts": {
			request: &Message{
				Kind:      MessageEventKind,
				MessageID: uuid.NewString(),
				Role:      RoleUser,
				Parts:     nil,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "message parts cannot be empty",
		},
		"valid request without task and context IDs": {
			request: createTaskTestMessage(RoleUser, "Hello, world!"),
			want: &Task{
				Status: TaskStatus{
					State:     TaskStateSubmitted,
					Timestamp: now().Format(time.RFC3339),
				},
				Kind: "",
				// ID and ContextID will be generated
				History: []*Message{createTaskTestMessage(RoleUser, "Hello, world!")},
			},
			wantErr: false,
		},
		"valid request with both task and context IDs": {
			request: createTaskTestMessageWithIDs(RoleUser, "Hello, world!", "task-123", "context-456"),
			want: &Task{
				Status: TaskStatus{
					State:     TaskStateSubmitted,
					Timestamp: now().Format(time.RFC3339),
				},
				Kind:      "",
				ID:        "task-123",
				ContextID: "context-456",
				History:   []*Message{createTaskTestMessageWithIDs(RoleUser, "Hello, world!", "task-123", "context-456")},
			},
			wantErr: false,
		},
		"valid request with only task ID": {
			request: func() *Message {
				msg := createTaskTestMessage(RoleUser, "Hello, world!")
				msg.TaskID = "task-123"
				return msg
			}(),
			want: &Task{
				Status: TaskStatus{
					State:     TaskStateSubmitted,
					Timestamp: now().Format(time.RFC3339),
				},
				Kind: "",
				ID:   "task-123",
				// ContextID will be generated
				History: []*Message{func() *Message {
					msg := createTaskTestMessage(RoleUser, "Hello, world!")
					msg.TaskID = "task-123"
					return msg
				}()},
			},
			wantErr: false,
		},
		"valid request with only context ID": {
			request: func() *Message {
				msg := createTaskTestMessage(RoleUser, "Hello, world!")
				msg.ContextID = "context-456"
				return msg
			}(),
			want: &Task{
				Status: TaskStatus{
					State:     TaskStateSubmitted,
					Timestamp: now().Format(time.RFC3339),
				},
				Kind:      "",
				ContextID: "context-456",
				// ID will be generated
				History: []*Message{func() *Message {
					msg := createTaskTestMessage(RoleUser, "Hello, world!")
					msg.ContextID = "context-456"
					return msg
				}()},
			},
			wantErr: false,
		},
		"valid request with agent role": {
			request: createTaskTestMessage(RoleAgent, "Agent response"),
			want: &Task{
				Status: TaskStatus{
					State:     TaskStateSubmitted,
					Timestamp: now().Format(time.RFC3339),
				},
				Kind: "",
				// ID and ContextID will be generated
				History: []*Message{createTaskTestMessage(RoleAgent, "Agent response")},
			},
			wantErr: false,
		},
		"valid request with multiple parts": {
			request: &Message{
				Kind:      MessageEventKind,
				MessageID: uuid.NewString(),
				Role:      RoleUser,
				Parts: []Part{
					NewTextPart("First part"),
					NewTextPart("Second part"),
				},
			},
			want: &Task{
				Status: TaskStatus{
					State:     TaskStateSubmitted,
					Timestamp: now().Format(time.RFC3339),
				},
				Kind: "",
				// ID and ContextID will be generated
				History: []*Message{
					{
						Kind:      MessageEventKind,
						MessageID: uuid.NewString(),
						Role:      RoleUser,
						Parts: []Part{
							NewTextPart("First part"),
							NewTextPart("Second part"),
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := NewTask(tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewTask() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NewTask() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got == nil {
				t.Error("NewTask() returned nil task")
				return
			}

			if got.Status.State != TaskStateSubmitted {
				t.Errorf("NewTask() task status state = %v, want %v", got.Status.State, TaskStateSubmitted)
			}

			if len(got.History) != 1 {
				t.Errorf("NewTask() task history length = %d, want 1", len(got.History))
			}

			// For tests with expected values, compare structure
			if tt.want != nil {
				// Compare status
				if got.Status.State != tt.want.Status.State {
					t.Errorf("NewTask() task status state = %v, want %v", got.Status.State, tt.want.Status.State)
				}

				// Compare provided IDs
				if tt.want.ID != "" && got.ID != tt.want.ID {
					t.Errorf("NewTask() task ID = %v, want %v", got.ID, tt.want.ID)
				}

				if tt.want.ContextID != "" && got.ContextID != tt.want.ContextID {
					t.Errorf("NewTask() task context ID = %v, want %v", got.ContextID, tt.want.ContextID)
				}

				// Compare history length
				if len(got.History) != len(tt.want.History) {
					t.Errorf("NewTask() task history length = %d, want %d", len(got.History), len(tt.want.History))
				}

				// Compare history content (basic validation)
				if len(got.History) > 0 && len(tt.want.History) > 0 {
					if got.History[0].Role != tt.want.History[0].Role {
						t.Errorf("NewTask() task history[0] role = %v, want %v", got.History[0].Role, tt.want.History[0].Role)
					}
				}
			}
		})
	}
}

func TestCompletedTask(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		taskID    string
		contextID string
		artifacts []*Artifact
		history   []*Message
		want      *Task
		wantErr   bool
	}{
		"basic completed task": {
			taskID:    "task-123",
			contextID: "context-456",
			artifacts: nil,
			history:   nil,
			want: &Task{
				Status: TaskStatus{
					State:     TaskStateCompleted,
					Timestamp: now().Format(time.RFC3339),
				},
				Kind:      "",
				ID:        "task-123",
				ContextID: "context-456",
				Artifacts: nil,
				History:   []*Message{},
			},
			wantErr: false,
		},
		"completed task with empty strings": {
			taskID:    "",
			contextID: "",
			artifacts: nil,
			history:   nil,
			want: &Task{
				Status: TaskStatus{
					State:     TaskStateCompleted,
					Timestamp: now().Format(time.RFC3339),
				},
				Kind:      "",
				ID:        "",
				ContextID: "",
				Artifacts: nil,
				History:   []*Message{},
			},
			wantErr: false,
		},
		"completed task with artifacts": {
			taskID:    "task-123",
			contextID: "context-456",
			artifacts: []*Artifact{
				{
					ArtifactID: "artifact-1",
					Name:       "Test Artifact",
					Parts:      []Part{NewTextPart("Artifact content")},
				},
			},
			history: nil,
			want: &Task{
				Status: TaskStatus{
					State:     TaskStateCompleted,
					Timestamp: now().Format(time.RFC3339),
				},
				Kind:      "",
				ID:        "task-123",
				ContextID: "context-456",
				Artifacts: []*Artifact{
					{
						ArtifactID: "artifact-1",
						Name:       "Test Artifact",
						Parts:      []Part{NewTextPart("Artifact content")},
					},
				},
				History: []*Message{},
			},
			wantErr: false,
		},
		"completed task with history": {
			taskID:    "task-123",
			contextID: "context-456",
			artifacts: nil,
			history: []*Message{
				createTaskTestMessageWithIDs(RoleUser, "Initial message", "task-123", "context-456"),
				createTaskTestMessageWithIDs(RoleAgent, "Agent response", "task-123", "context-456"),
			},
			want: &Task{
				Status: TaskStatus{
					State:     TaskStateCompleted,
					Timestamp: now().Format(time.RFC3339),
				},
				Kind:      "",
				ID:        "task-123",
				ContextID: "context-456",
				Artifacts: nil,
				History: []*Message{
					createTaskTestMessageWithIDs(RoleUser, "Initial message", "task-123", "context-456"),
					createTaskTestMessageWithIDs(RoleAgent, "Agent response", "task-123", "context-456"),
				},
			},
			wantErr: false,
		},
		"completed task with artifacts and history": {
			taskID:    "task-123",
			contextID: "context-456",
			artifacts: []*Artifact{
				{
					ArtifactID: "artifact-1",
					Name:       "Test Artifact",
					Parts:      []Part{NewTextPart("Artifact content")},
				},
			},
			history: []*Message{
				createTaskTestMessageWithIDs(RoleUser, "Initial message", "task-123", "context-456"),
			},
			want: &Task{
				Status: TaskStatus{
					State:     TaskStateCompleted,
					Timestamp: now().Format(time.RFC3339),
				},
				Kind:      "",
				ID:        "task-123",
				ContextID: "context-456",
				Artifacts: []*Artifact{
					{
						ArtifactID: "artifact-1",
						Name:       "Test Artifact",
						Parts:      []Part{NewTextPart("Artifact content")},
					},
				},
				History: []*Message{
					createTaskTestMessageWithIDs(RoleUser, "Initial message", "task-123", "context-456"),
				},
			},
			wantErr: false,
		},
		"completed task with empty artifacts slice": {
			taskID:    "task-123",
			contextID: "context-456",
			artifacts: []*Artifact{},
			history:   nil,
			want: &Task{
				Status: TaskStatus{
					State:     TaskStateCompleted,
					Timestamp: now().Format(time.RFC3339),
				},
				Kind:      "",
				ID:        "task-123",
				ContextID: "context-456",
				Artifacts: []*Artifact{},
				History:   []*Message{},
			},
			wantErr: false,
		},
		"completed task with empty history slice": {
			taskID:    "task-123",
			contextID: "context-456",
			artifacts: nil,
			history:   []*Message{},
			want: &Task{
				Status: TaskStatus{
					State:     TaskStateCompleted,
					Timestamp: now().Format(time.RFC3339),
				},
				Kind:      "",
				ID:        "task-123",
				ContextID: "context-456",
				Artifacts: nil,
				History:   []*Message{},
			},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := CompletedTask(tt.taskID, tt.contextID, tt.artifacts, tt.history)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CompletedTask() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("CompletedTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got == nil {
				t.Error("CompletedTask() returned nil task")
				return
			}

			if got.Status.State != TaskStateCompleted {
				t.Errorf("CompletedTask() task status state = %v, want %v", got.Status.State, TaskStateCompleted)
			}

			if got.ID != tt.taskID {
				t.Errorf("CompletedTask() task ID = %v, want %v", got.ID, tt.taskID)
			}

			if got.ContextID != tt.contextID {
				t.Errorf("CompletedTask() task context ID = %v, want %v", got.ContextID, tt.contextID)
			}

			// Check history is never nil
			if got.History == nil {
				t.Error("CompletedTask() task history is nil, should be empty slice")
			}

			// Compare artifacts
			if len(got.Artifacts) != len(tt.artifacts) {
				t.Errorf("CompletedTask() artifacts length = %d, want %d", len(got.Artifacts), len(tt.artifacts))
			}

			// Compare history
			if len(got.History) != len(tt.history) {
				t.Errorf("CompletedTask() history length = %d, want %d", len(got.History), len(tt.history))
			}

			// Use go-cmp for deep comparison if want is provided
			if tt.want != nil {
				// Compare without the generated UUID fields and other dynamic fields
				opts := cmp.Options{
					cmp.AllowUnexported(TextPart{}),
					// Ignore MessageID fields since they contain generated UUIDs
					cmp.FilterPath(func(p cmp.Path) bool { return p.String() == "History.MessageID" }, cmp.Ignore()),
				}
				if diff := cmp.Diff(tt.want, got, opts...); diff != "" {
					t.Errorf("CompletedTask() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// Test edge cases and specific behaviors
func TestNewTask_UUIDGeneration(t *testing.T) {
	t.Parallel()

	msg1 := createTaskTestMessage(RoleUser, "test")
	msg2 := createTaskTestMessage(RoleUser, "test")

	// Test multiple calls generate different UUIDs
	task1, err := NewTask(msg1)
	if err != nil {
		t.Fatalf("NewTask() error = %v", err)
	}

	task2, err := NewTask(msg2)
	if err != nil {
		t.Fatalf("NewTask() error = %v", err)
	}

	if task1.ID == task2.ID {
		t.Error("NewTask() generated same UUID for different calls")
	}

	if task1.ContextID == task2.ContextID {
		t.Error("NewTask() generated same context UUID for different calls")
	}
}

func TestNewTask_PreservesProvidedIDs(t *testing.T) {
	t.Parallel()

	msg := createTaskTestMessageWithIDs(RoleUser, "test", "provided-task-id", "provided-context-id")

	task, err := NewTask(msg)
	if err != nil {
		t.Fatalf("NewTask() error = %v", err)
	}

	if task.ID != "provided-task-id" {
		t.Errorf("NewTask() task ID = %v, want %v", task.ID, "provided-task-id")
	}

	if task.ContextID != "provided-context-id" {
		t.Errorf("NewTask() context ID = %v, want %v", task.ContextID, "provided-context-id")
	}
}

func TestCompletedTask_NilHistoryBecomesEmpty(t *testing.T) {
	t.Parallel()

	task, err := CompletedTask("task-id", "context-id", nil, nil)
	if err != nil {
		t.Fatalf("CompletedTask() error = %v", err)
	}

	if task.History == nil {
		t.Error("CompletedTask() history is nil, should be empty slice")
	}

	if len(task.History) != 0 {
		t.Errorf("CompletedTask() history length = %d, want 0", len(task.History))
	}
}

func BenchmarkNewTask(b *testing.B) {
	msg := createTaskTestMessage(RoleUser, "benchmark test message")

	for b.Loop() {
		_, err := NewTask(msg)
		if err != nil {
			b.Fatalf("NewTask() error = %v", err)
		}
	}
}

func BenchmarkCompletedTask(b *testing.B) {
	artifacts := []*Artifact{
		{
			ArtifactID: "artifact-1",
			Name:       "Test Artifact",
			Parts:      []Part{NewTextPart("Artifact content")},
		},
	}
	history := []*Message{
		createTaskTestMessageWithIDs(RoleUser, "Initial message", "task-123", "context-456"),
	}

	for b.Loop() {
		_, err := CompletedTask("task-123", "context-456", artifacts, history)
		if err != nil {
			b.Fatalf("CompletedTask() error = %v", err)
		}
	}
}
