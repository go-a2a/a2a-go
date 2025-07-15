// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "google.golang.org/a2a/v1"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestToProtoMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    *Message
		expected *v1.Message
	}{
		{
			name:     "nil message",
			input:    nil,
			expected: nil,
		},
		{
			name: "basic message",
			input: &Message{
				ContextID: "test-context",
				Content:   "Hello, world!",
			},
			expected: &v1.Message{
				ContextId: "test-context",
				Content: []*v1.Part{{
					Part: &v1.Part_Text{
						Text: "Hello, world!",
					},
				}},
			},
		},
		{
			name: "message with metadata",
			input: &Message{
				ContextID: "test-context",
				Content:   "Hello, world!",
				Metadata: map[string]any{
					"key1": "value1",
					"key2": 42,
				},
			},
			expected: &v1.Message{
				ContextId: "test-context",
				Content: []*v1.Part{{
					Part: &v1.Part_Text{
						Text: "Hello, world!",
					},
				}},
				Metadata: func() *structpb.Struct {
					s, _ := structpb.NewStruct(map[string]any{
						"key1": "value1",
						"key2": 42,
					})
					return s
				}(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToProtoMessage(tt.input)
			if err != nil {
				t.Errorf("ToProtoMessage() error = %v", err)
				return
			}

			if diff := cmp.Diff(tt.expected, got, protocmp.Transform()); diff != "" {
				t.Errorf("ToProtoMessage() mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}

func TestFromProtoMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    *v1.Message
		expected *Message
	}{
		{
			name:     "nil message",
			input:    nil,
			expected: nil,
		},
		{
			name: "basic message",
			input: &v1.Message{
				ContextId: "test-context",
				Content: []*v1.Part{{
					Part: &v1.Part_Text{
						Text: "Hello, world!",
					},
				}},
			},
			expected: &Message{
				ContextID: "test-context",
				Content:   "Hello, world!",
			},
		},
		{
			name: "message with metadata",
			input: &v1.Message{
				ContextId: "test-context",
				Content: []*v1.Part{{
					Part: &v1.Part_Text{
						Text: "Hello, world!",
					},
				}},
				Metadata: func() *structpb.Struct {
					s, _ := structpb.NewStruct(map[string]any{
						"key1": "value1",
						"key2": 42,
					})
					return s
				}(),
			},
			expected: &Message{
				ContextID: "test-context",
				Content:   "Hello, world!",
				Metadata: map[string]any{
					"key1": "value1",
					"key2": float64(42), // Protobuf converts integers to float64
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromProtoMessage(tt.input)
			if err != nil {
				t.Errorf("FromProtoMessage() error = %v", err)
				return
			}

			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("FromProtoMessage() mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}

func TestToProtoTaskState(t *testing.T) {
	tests := []struct {
		name     string
		input    TaskState
		expected v1.TaskState
	}{
		{
			name:     "submitted state",
			input:    TaskStateSubmitted,
			expected: v1.TaskState_TASK_STATE_SUBMITTED,
		},
		{
			name:     "running state",
			input:    TaskStateRunning,
			expected: v1.TaskState_TASK_STATE_WORKING,
		},
		{
			name:     "completed state",
			input:    TaskStateCompleted,
			expected: v1.TaskState_TASK_STATE_COMPLETED,
		},
		{
			name:     "failed state",
			input:    TaskStateFailed,
			expected: v1.TaskState_TASK_STATE_FAILED,
		},
		{
			name:     "canceled state",
			input:    TaskStateCanceled,
			expected: v1.TaskState_TASK_STATE_CANCELLED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToProtoTaskState(tt.input)
			if got != tt.expected {
				t.Errorf("ToProtoTaskState() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestFromProtoTaskState(t *testing.T) {
	tests := []struct {
		name     string
		input    v1.TaskState
		expected TaskState
	}{
		{
			name:     "submitted state",
			input:    v1.TaskState_TASK_STATE_SUBMITTED,
			expected: TaskStateSubmitted,
		},
		{
			name:     "working state",
			input:    v1.TaskState_TASK_STATE_WORKING,
			expected: TaskStateRunning,
		},
		{
			name:     "completed state",
			input:    v1.TaskState_TASK_STATE_COMPLETED,
			expected: TaskStateCompleted,
		},
		{
			name:     "failed state",
			input:    v1.TaskState_TASK_STATE_FAILED,
			expected: TaskStateFailed,
		},
		{
			name:     "cancelled state",
			input:    v1.TaskState_TASK_STATE_CANCELLED,
			expected: TaskStateCanceled,
		},
		{
			name:     "unspecified state",
			input:    v1.TaskState_TASK_STATE_UNSPECIFIED,
			expected: TaskStateSubmitted, // Default fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromProtoTaskState(tt.input)
			if got != tt.expected {
				t.Errorf("FromProtoTaskState() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestToProtoTaskStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    *TaskStatus
		expected *v1.TaskStatus
	}{
		{
			name:     "nil status",
			input:    nil,
			expected: nil,
		},
		{
			name: "submitted status",
			input: &TaskStatus{
				State: TaskStateSubmitted,
			},
			expected: &v1.TaskStatus{
				State: v1.TaskState_TASK_STATE_SUBMITTED,
			},
		},
		{
			name: "completed status",
			input: &TaskStatus{
				State: TaskStateCompleted,
			},
			expected: &v1.TaskStatus{
				State: v1.TaskState_TASK_STATE_COMPLETED,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToProtoTaskStatus(tt.input)
			if err != nil {
				t.Errorf("ToProtoTaskStatus() error = %v", err)
				return
			}

			if diff := cmp.Diff(tt.expected, got, protocmp.Transform()); diff != "" {
				t.Errorf("ToProtoTaskStatus() mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	original := &Message{
		ContextID: "test-context-123",
		Content:   "This is a test message",
		Metadata: map[string]any{
			"timestamp": float64(1234567890), // Use float64 to match protobuf behavior
			"user":      "test-user",
			"priority":  "high",
		},
	}

	// Convert to proto
	proto, err := ToProtoMessage(original)
	if err != nil {
		t.Fatalf("ToProtoMessage() error = %v", err)
	}

	// Convert back to Go
	converted, err := FromProtoMessage(proto)
	if err != nil {
		t.Fatalf("FromProtoMessage() error = %v", err)
	}

	// Check that the round-trip conversion preserves the data
	if converted.ContextID != original.ContextID {
		t.Errorf("ContextID mismatch: got %v, expected %v", converted.ContextID, original.ContextID)
	}
	if converted.Content != original.Content {
		t.Errorf("Content mismatch: got %v, expected %v", converted.Content, original.Content)
	}

	// Check metadata (note: floating-point numbers may have precision differences)
	if diff := cmp.Diff(original.Metadata, converted.Metadata); diff != "" {
		t.Errorf("Metadata mismatch (-expected +got):\n%s", diff)
	}
}
