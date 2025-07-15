// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"fmt"
	"strings"
	"testing"
)

func TestMessageTaskInput(t *testing.T) {
	message := &Message{
		Content: "Hello, world!",
	}
	contextID := "ctx-123"
	taskID := "task-456"

	input := MessageTaskInput{
		Message:   message,
		TaskID:    &taskID,
		ContextID: &contextID,
	}

	// Test GetTaskID
	if input.GetTaskID() == nil || *input.GetTaskID() != taskID {
		t.Errorf("Expected task ID %s, got %v", taskID, input.GetTaskID())
	}

	// Test GetContextID
	if input.GetContextID() == nil || *input.GetContextID() != contextID {
		t.Errorf("Expected context ID %s, got %v", contextID, input.GetContextID())
	}

	// Test AsMessage
	if input.AsMessage() != message {
		t.Error("Expected AsMessage to return the original message")
	}
}

func TestA2AMessageTaskInput(t *testing.T) {
	contextID := "ctx-123"
	taskID := "task-456"
	a2aMessage := &A2AMessage{
		Role:      RoleAgent,
		Parts:     []*MessagePart{},
		MessageID: "msg-789",
		TaskID:    &taskID,
		ContextID: &contextID,
	}

	input := A2AMessageTaskInput{
		A2AMessage: a2aMessage,
	}

	// Test GetTaskID
	if input.GetTaskID() == nil || *input.GetTaskID() != taskID {
		t.Errorf("Expected task ID %s, got %v", taskID, input.GetTaskID())
	}

	// Test GetContextID
	if input.GetContextID() == nil || *input.GetContextID() != contextID {
		t.Errorf("Expected context ID %s, got %v", contextID, input.GetContextID())
	}

	// Test AsMessage
	convertedMessage := input.AsMessage()
	if convertedMessage == nil {
		t.Error("Expected AsMessage to return a message")
	}
	if convertedMessage.ContextID != contextID {
		t.Errorf("Expected context ID %s, got %s", contextID, convertedMessage.ContextID)
	}
}

func TestA2AMessageTaskInput_AsMessage_NilA2AMessage(t *testing.T) {
	input := A2AMessageTaskInput{
		A2AMessage: nil,
	}

	message := input.AsMessage()
	if message != nil {
		t.Error("Expected AsMessage to return nil for nil A2AMessage")
	}
}

func TestA2AMessageTaskInput_AsMessage_WithTextParts(t *testing.T) {
	contextID := "ctx-123"
	a2aMessage := &A2AMessage{
		Role: RoleAgent,
		Parts: []*MessagePart{
			{
				Root: &TextPart{
					Kind: "text",
					Text: "Hello",
				},
			},
			{
				Root: &TextPart{
					Kind: "text",
					Text: "World",
				},
			},
		},
		MessageID: "msg-789",
		ContextID: &contextID,
	}

	input := A2AMessageTaskInput{
		A2AMessage: a2aMessage,
	}

	message := input.AsMessage()
	if message == nil {
		t.Error("Expected AsMessage to return a message")
	}

	expectedContent := "Hello\nWorld"
	if message.Content != expectedContent {
		t.Errorf("Expected content %s, got %s", expectedContent, message.Content)
	}
}

func TestNewTask(t *testing.T) {
	tests := []struct {
		name      string
		request   TaskMessageInput
		wantError bool
	}{
		{
			name: "valid message task input",
			request: MessageTaskInput{
				Message: &Message{
					Content: "Hello, world!",
				},
				TaskID:    nil,
				ContextID: nil,
			},
			wantError: false,
		},
		{
			name: "valid message task input with IDs",
			request: MessageTaskInput{
				Message: &Message{
					Content: "Hello, world!",
				},
				TaskID:    stringPtr("task-123"),
				ContextID: stringPtr("ctx-456"),
			},
			wantError: false,
		},
		{
			name:      "nil request",
			request:   nil,
			wantError: true,
		},
		{
			name: "invalid message",
			request: MessageTaskInput{
				Message: &Message{
					Content: "",
				},
				TaskID:    nil,
				ContextID: nil,
			},
			wantError: true,
		},
		{
			name: "nil message",
			request: MessageTaskInput{
				Message:   nil,
				TaskID:    nil,
				ContextID: nil,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewTask(tt.request)
			if tt.wantError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
			if tt.wantError {
				return
			}

			// Validate task properties
			if task.ID == "" {
				t.Error("Expected non-empty task ID")
			}
			if task.ContextID == "" {
				t.Error("Expected non-empty context ID")
			}
			if task.Status.State != TaskStateSubmitted {
				t.Errorf("Expected task state %s, got %s", TaskStateSubmitted, task.Status.State)
			}
			if len(task.History) != 1 {
				t.Errorf("Expected 1 history item, got %d", len(task.History))
			}

			// Check if provided IDs are used
			if tt.request.GetTaskID() != nil && *tt.request.GetTaskID() != "" {
				if task.ID != *tt.request.GetTaskID() {
					t.Errorf("Expected task ID %s, got %s", *tt.request.GetTaskID(), task.ID)
				}
			}
			if tt.request.GetContextID() != nil && *tt.request.GetContextID() != "" {
				if task.ContextID != *tt.request.GetContextID() {
					t.Errorf("Expected context ID %s, got %s", *tt.request.GetContextID(), task.ContextID)
				}
			}
		})
	}
}

func TestNewTaskFromMessage(t *testing.T) {
	tests := []struct {
		name      string
		message   *Message
		taskID    *string
		contextID *string
		wantError bool
	}{
		{
			name: "valid message",
			message: &Message{
				Content: "Hello, world!",
			},
			taskID:    nil,
			contextID: nil,
			wantError: false,
		},
		{
			name: "valid message with IDs",
			message: &Message{
				Content: "Hello, world!",
			},
			taskID:    stringPtr("task-123"),
			contextID: stringPtr("ctx-456"),
			wantError: false,
		},
		{
			name:      "nil message",
			message:   nil,
			taskID:    nil,
			contextID: nil,
			wantError: true,
		},
		{
			name: "invalid message",
			message: &Message{
				Content: "",
			},
			taskID:    nil,
			contextID: nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewTaskFromMessage(tt.message, tt.taskID, tt.contextID)
			if tt.wantError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
			if tt.wantError {
				return
			}

			// Validate task properties
			if task.ID == "" {
				t.Error("Expected non-empty task ID")
			}
			if task.ContextID == "" {
				t.Error("Expected non-empty context ID")
			}
			if task.Status.State != TaskStateSubmitted {
				t.Errorf("Expected task state %s, got %s", TaskStateSubmitted, task.Status.State)
			}
			if len(task.History) != 1 {
				t.Errorf("Expected 1 history item, got %d", len(task.History))
			}

			// Check if provided IDs are used
			if tt.taskID != nil && *tt.taskID != "" {
				if task.ID != *tt.taskID {
					t.Errorf("Expected task ID %s, got %s", *tt.taskID, task.ID)
				}
			}
			if tt.contextID != nil && *tt.contextID != "" {
				if task.ContextID != *tt.contextID {
					t.Errorf("Expected context ID %s, got %s", *tt.contextID, task.ContextID)
				}
			}
		})
	}
}

func TestNewTaskFromA2AMessage(t *testing.T) {
	tests := []struct {
		name       string
		a2aMessage *A2AMessage
		wantError  bool
	}{
		{
			name: "valid A2A message",
			a2aMessage: &A2AMessage{
				Role: RoleAgent,
				Parts: []*MessagePart{
					{
						Root: &TextPart{
							Kind: "text",
							Text: "Hello, world!",
						},
					},
				},
				MessageID: "msg-123",
			},
			wantError: false,
		},
		{
			name: "valid A2A message with IDs",
			a2aMessage: &A2AMessage{
				Role: RoleAgent,
				Parts: []*MessagePart{
					{
						Root: &TextPart{
							Kind: "text",
							Text: "Hello, world!",
						},
					},
				},
				MessageID: "msg-123",
				TaskID:    stringPtr("task-456"),
				ContextID: stringPtr("ctx-789"),
			},
			wantError: false,
		},
		{
			name:       "nil A2A message",
			a2aMessage: nil,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewTaskFromA2AMessage(tt.a2aMessage)
			if tt.wantError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
			if tt.wantError {
				return
			}

			// Validate task properties
			if task.ID == "" {
				t.Error("Expected non-empty task ID")
			}
			if task.ContextID == "" {
				t.Error("Expected non-empty context ID")
			}
			if task.Status.State != TaskStateSubmitted {
				t.Errorf("Expected task state %s, got %s", TaskStateSubmitted, task.Status.State)
			}
			if len(task.History) != 1 {
				t.Errorf("Expected 1 history item, got %d", len(task.History))
			}

			// Check if provided IDs are used
			if tt.a2aMessage.TaskID != nil && *tt.a2aMessage.TaskID != "" {
				if task.ID != *tt.a2aMessage.TaskID {
					t.Errorf("Expected task ID %s, got %s", *tt.a2aMessage.TaskID, task.ID)
				}
			}
			if tt.a2aMessage.ContextID != nil && *tt.a2aMessage.ContextID != "" {
				if task.ContextID != *tt.a2aMessage.ContextID {
					t.Errorf("Expected context ID %s, got %s", *tt.a2aMessage.ContextID, task.ContextID)
				}
			}
		})
	}
}

func TestCompletedTask(t *testing.T) {
	validArtifact := &Artifact{
		ArtifactID: "art-123",
		Parts: []*PartWrapper{
			NewPartWrapper(&TextPart{
				Kind: "text",
				Text: "Hello, world!",
			}),
		},
	}

	validMessage := &Message{
		Content: "Hello, world!",
	}

	tests := []struct {
		name      string
		taskID    string
		contextID string
		artifacts []*Artifact
		history   []*Message
		wantError bool
	}{
		{
			name:      "valid completed task",
			taskID:    "task-123",
			contextID: "ctx-456",
			artifacts: []*Artifact{validArtifact},
			history:   []*Message{validMessage},
			wantError: false,
		},
		{
			name:      "valid completed task with no history",
			taskID:    "task-123",
			contextID: "ctx-456",
			artifacts: []*Artifact{validArtifact},
			history:   nil,
			wantError: false,
		},
		{
			name:      "empty task ID",
			taskID:    "",
			contextID: "ctx-456",
			artifacts: []*Artifact{validArtifact},
			history:   []*Message{validMessage},
			wantError: true,
		},
		{
			name:      "empty context ID",
			taskID:    "task-123",
			contextID: "",
			artifacts: []*Artifact{validArtifact},
			history:   []*Message{validMessage},
			wantError: true,
		},
		{
			name:      "nil artifacts",
			taskID:    "task-123",
			contextID: "ctx-456",
			artifacts: nil,
			history:   []*Message{validMessage},
			wantError: true,
		},
		{
			name:      "nil artifact in artifacts",
			taskID:    "task-123",
			contextID: "ctx-456",
			artifacts: []*Artifact{nil},
			history:   []*Message{validMessage},
			wantError: true,
		},
		{
			name:      "invalid artifact",
			taskID:    "task-123",
			contextID: "ctx-456",
			artifacts: []*Artifact{
				{
					ArtifactID: "",
					Parts:      []*PartWrapper{},
				},
			},
			history:   []*Message{validMessage},
			wantError: true,
		},
		{
			name:      "nil message in history",
			taskID:    "task-123",
			contextID: "ctx-456",
			artifacts: []*Artifact{validArtifact},
			history:   []*Message{nil},
			wantError: true,
		},
		{
			name:      "invalid message in history",
			taskID:    "task-123",
			contextID: "ctx-456",
			artifacts: []*Artifact{validArtifact},
			history: []*Message{
				{
					Content: "",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := CompletedTask(tt.taskID, tt.contextID, tt.artifacts, tt.history)
			if tt.wantError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
			if tt.wantError {
				return
			}

			// Validate task properties
			if task.ID != tt.taskID {
				t.Errorf("Expected task ID %s, got %s", tt.taskID, task.ID)
			}
			if task.ContextID != tt.contextID {
				t.Errorf("Expected context ID %s, got %s", tt.contextID, task.ContextID)
			}
			if task.Status.State != TaskStateCompleted {
				t.Errorf("Expected task state %s, got %s", TaskStateCompleted, task.Status.State)
			}
			if len(task.Artifacts) != len(tt.artifacts) {
				t.Errorf("Expected %d artifacts, got %d", len(tt.artifacts), len(task.Artifacts))
			}
			expectedHistoryLen := len(tt.history)
			if tt.history == nil {
				expectedHistoryLen = 0
			}
			if len(task.History) != expectedHistoryLen {
				t.Errorf("Expected %d history items, got %d", expectedHistoryLen, len(task.History))
			}
		})
	}
}

func TestCompletedTaskWithNoHistory(t *testing.T) {
	validArtifact := &Artifact{
		ArtifactID: "art-123",
		Parts: []*PartWrapper{
			NewPartWrapper(&TextPart{
				Kind: "text",
				Text: "Hello, world!",
			}),
		},
	}

	tests := []struct {
		name      string
		taskID    string
		contextID string
		artifacts []*Artifact
		wantError bool
	}{
		{
			name:      "valid completed task with no history",
			taskID:    "task-123",
			contextID: "ctx-456",
			artifacts: []*Artifact{validArtifact},
			wantError: false,
		},
		{
			name:      "empty task ID",
			taskID:    "",
			contextID: "ctx-456",
			artifacts: []*Artifact{validArtifact},
			wantError: true,
		},
		{
			name:      "nil artifacts",
			taskID:    "task-123",
			contextID: "ctx-456",
			artifacts: nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := CompletedTaskWithNoHistory(tt.taskID, tt.contextID, tt.artifacts)
			if tt.wantError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
			if tt.wantError {
				return
			}

			// Validate task properties
			if task.ID != tt.taskID {
				t.Errorf("Expected task ID %s, got %s", tt.taskID, task.ID)
			}
			if task.ContextID != tt.contextID {
				t.Errorf("Expected context ID %s, got %s", tt.contextID, task.ContextID)
			}
			if task.Status.State != TaskStateCompleted {
				t.Errorf("Expected task state %s, got %s", TaskStateCompleted, task.Status.State)
			}
			if len(task.Artifacts) != len(tt.artifacts) {
				t.Errorf("Expected %d artifacts, got %d", len(tt.artifacts), len(task.Artifacts))
			}
			if len(task.History) != 0 {
				t.Errorf("Expected 0 history items, got %d", len(task.History))
			}
		})
	}
}

// Test that generated UUIDs are unique and valid
func TestTaskUUIDGeneration(t *testing.T) {
	// Create multiple tasks to verify UUID uniqueness
	taskIDs := make(map[string]bool)
	contextIDs := make(map[string]bool)

	for i := 0; i < 100; i++ {
		message := &Message{
			Content: "test message",
		}

		task, err := NewTaskFromMessage(message, nil, nil)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		// Check if task ID is unique
		if taskIDs[task.ID] {
			t.Errorf("Duplicate task ID generated: %s", task.ID)
		}
		taskIDs[task.ID] = true

		// Check if context ID is unique
		if contextIDs[task.ContextID] {
			t.Errorf("Duplicate context ID generated: %s", task.ContextID)
		}
		contextIDs[task.ContextID] = true

		// Basic UUID format check (should have 4 dashes)
		if strings.Count(task.ID, "-") != 4 {
			t.Errorf("Invalid task ID UUID format: %s", task.ID)
		}
		if strings.Count(task.ContextID, "-") != 4 {
			t.Errorf("Invalid context ID UUID format: %s", task.ContextID)
		}
	}
}

// Test task input interface behavior
func TestTaskInputInterface(t *testing.T) {
	// Test that both implementations satisfy the interface
	message := &Message{
		Content: "Hello, world!",
	}

	a2aMessage := &A2AMessage{
		Role: RoleAgent,
		Parts: []*MessagePart{
			{
				Root: &TextPart{
					Kind: "text",
					Text: "Hello, world!",
				},
			},
		},
		MessageID: "msg-123",
	}

	// Test interface compliance
	var inputs []TaskMessageInput
	inputs = append(inputs, MessageTaskInput{Message: message})
	inputs = append(inputs, A2AMessageTaskInput{A2AMessage: a2aMessage})

	for i, input := range inputs {
		t.Run(fmt.Sprintf("input_%d", i), func(t *testing.T) {
			task, err := NewTask(input)
			if err != nil {
				t.Errorf("Failed to create task from input %d: %v", i, err)
			}
			if task == nil {
				t.Error("Expected non-nil task")
			}
		})
	}
}

// Benchmark tests
func BenchmarkNewTask(b *testing.B) {
	input := MessageTaskInput{
		Message: &Message{
			Content: "Hello, world!",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewTask(input)
		if err != nil {
			b.Fatalf("Failed to create task: %v", err)
		}
	}
}

func BenchmarkNewTaskFromMessage(b *testing.B) {
	message := &Message{
		Content: "Hello, world!",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewTaskFromMessage(message, nil, nil)
		if err != nil {
			b.Fatalf("Failed to create task: %v", err)
		}
	}
}

func BenchmarkCompletedTask(b *testing.B) {
	artifact := &Artifact{
		ArtifactID: "art-123",
		Parts: []*PartWrapper{
			NewPartWrapper(&TextPart{
				Kind: "text",
				Text: "Hello, world!",
			}),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CompletedTask("task-123", "ctx-456", []*Artifact{artifact}, nil)
		if err != nil {
			b.Fatalf("Failed to create completed task: %v", err)
		}
	}
}
