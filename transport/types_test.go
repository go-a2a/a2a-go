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

package transport

import (
	"testing"

	a2a "github.com/go-a2a/a2a-go"
)

func TestMessageOrTaskResult(t *testing.T) {
	t.Parallel()

	t.Run("unmarshal valid message", func(t *testing.T) {
		t.Parallel()

		jsonData := []byte(`{
			"kind": "message",
			"messageId": "test-msg-123",
			"role": "user",
			"parts": [
				{
					"kind": "text",
					"text": "Hello, world!"
				}
			]
		}`)

		var result messageOrTaskResult
		err := result.UnmarshalJSON(jsonData)
		if err != nil {
			t.Fatalf("UnmarshalJSON() error: %v", err)
		}

		if result.result == nil {
			t.Fatal("UnmarshalJSON() result is nil")
		}

		msg, ok := result.result.(*a2a.Message)
		if !ok {
			t.Fatalf("Result type %T, want *a2a.Message", result.result)
		}

		if msg.MessageID != "test-msg-123" {
			t.Errorf("MessageID = %q, want %q", msg.MessageID, "test-msg-123")
		}

		if msg.Role != a2a.RoleUser {
			t.Errorf("Role = %q, want %q", msg.Role, a2a.RoleUser)
		}

		if len(msg.Parts) != 1 {
			t.Fatalf("Parts length = %d, want 1", len(msg.Parts))
		}

		textPart, ok := msg.Parts[0].(*a2a.TextPart)
		if !ok {
			t.Fatalf("Part type %T, want *a2a.TextPart", msg.Parts[0])
		}

		if textPart.Text != "Hello, world!" {
			t.Errorf("Text = %q, want %q", textPart.Text, "Hello, world!")
		}
	})

	t.Run("unmarshal valid task", func(t *testing.T) {
		t.Parallel()

		jsonData := []byte(`{
			"kind": "task",
			"id": "task-123",
			"contextId": "ctx-456",
			"status": {
				"state": "working"
			}
		}`)

		var result messageOrTaskResult
		err := result.UnmarshalJSON(jsonData)
		if err != nil {
			t.Fatalf("UnmarshalJSON() error: %v", err)
		}

		if result.result == nil {
			t.Fatal("UnmarshalJSON() result is nil")
		}

		task, ok := result.result.(*a2a.Task)
		if !ok {
			t.Fatalf("Result type %T, want *a2a.Task", result.result)
		}

		if task.ID != "task-123" {
			t.Errorf("ID = %q, want %q", task.ID, "task-123")
		}

		if task.ContextID != "ctx-456" {
			t.Errorf("ContextID = %q, want %q", task.ContextID, "ctx-456")
		}

		if task.Status.State != a2a.TaskStateWorking {
			t.Errorf("Status.State = %q, want %q", task.Status.State, a2a.TaskStateWorking)
		}
	})

	t.Run("GetTaskID from message", func(t *testing.T) {
		t.Parallel()

		result := &messageOrTaskResult{
			result: &a2a.Message{
				TaskID: "msg-task-123",
			},
		}

		taskID := result.GetTaskID()
		if taskID != "msg-task-123" {
			t.Errorf("GetTaskID() = %q, want %q", taskID, "msg-task-123")
		}
	})

	t.Run("GetTaskID from task", func(t *testing.T) {
		t.Parallel()

		result := &messageOrTaskResult{
			result: &a2a.Task{
				ID: "task-456",
			},
		}

		taskID := result.GetTaskID()
		if taskID != "task-456" {
			t.Errorf("GetTaskID() = %q, want %q", taskID, "task-456")
		}
	})

	t.Run("GetKind from message", func(t *testing.T) {
		t.Parallel()

		result := &messageOrTaskResult{
			result: &a2a.Message{
				Kind: a2a.MessageEventKind,
			},
		}

		kind := result.GetKind()
		if kind != a2a.MessageEventKind {
			t.Errorf("GetKind() = %q, want %q", kind, a2a.MessageEventKind)
		}
	})

	t.Run("GetKind from task", func(t *testing.T) {
		t.Parallel()

		result := &messageOrTaskResult{
			result: &a2a.Task{
				Kind: a2a.TaskEventKind,
			},
		}

		kind := result.GetKind()
		if kind != a2a.TaskEventKind {
			t.Errorf("GetKind() = %q, want %q", kind, a2a.TaskEventKind)
		}
	})

	t.Run("nil result", func(t *testing.T) {
		t.Parallel()

		result := &messageOrTaskResult{result: nil}

		taskID := result.GetTaskID()
		if taskID != "" {
			t.Errorf("GetTaskID() = %q, want empty string", taskID)
		}

		kind := result.GetKind()
		if kind != a2a.EventKind("") {
			t.Errorf("GetKind() = %q, want empty string", kind)
		}
	})

	t.Run("unmarshal invalid JSON", func(t *testing.T) {
		t.Parallel()

		jsonData := []byte(`{invalid json}`)

		var result messageOrTaskResult
		err := result.UnmarshalJSON(jsonData)
		if err == nil {
			t.Fatal("Expected error for invalid JSON")
		}
	})

	t.Run("unmarshal unknown kind", func(t *testing.T) {
		t.Parallel()

		jsonData := []byte(`{
			"kind": "unknown-kind",
			"id": "test-123"
		}`)

		var result messageOrTaskResult
		err := result.UnmarshalJSON(jsonData)
		if err == nil {
			t.Fatal("Expected error for unknown kind")
		}
	})
}

func TestSendStreamingMessageResponse(t *testing.T) {
	t.Parallel()

	t.Run("unmarshal artifact update event", func(t *testing.T) {
		t.Parallel()

		jsonData := []byte(`{
			"kind": "artifact-update",
			"taskId": "task-123",
			"contextId": "ctx-456",
			"artifact": {
				"artifactId": "artifact-789",
				"parts": []
			}
		}`)

		var result sendStreamingMessageResponse
		err := result.UnmarshalJSON(jsonData)
		if err != nil {
			t.Fatalf("UnmarshalJSON() error: %v", err)
		}

		if result.result == nil {
			t.Fatal("UnmarshalJSON() result is nil")
		}

		event, ok := result.result.(*a2a.TaskArtifactUpdateEvent)
		if !ok {
			t.Fatalf("Result type %T, want *a2a.TaskArtifactUpdateEvent", result.result)
		}

		if event.TaskID != "task-123" {
			t.Errorf("TaskID = %q, want %q", event.TaskID, "task-123")
		}

		if event.ContextID != "ctx-456" {
			t.Errorf("ContextID = %q, want %q", event.ContextID, "ctx-456")
		}

		if event.Artifact.ArtifactID != "artifact-789" {
			t.Errorf("ArtifactID = %q, want %q", event.Artifact.ArtifactID, "artifact-789")
		}
	})

	t.Run("unmarshal status update event", func(t *testing.T) {
		t.Parallel()

		jsonData := []byte(`{
			"kind": "status-update",
			"taskId": "task-123",
			"contextId": "ctx-456",
			"status": {
				"state": "completed"
			}
		}`)

		var result sendStreamingMessageResponse
		err := result.UnmarshalJSON(jsonData)
		if err != nil {
			t.Fatalf("UnmarshalJSON() error: %v", err)
		}

		if result.result == nil {
			t.Fatal("UnmarshalJSON() result is nil")
		}

		event, ok := result.result.(*a2a.TaskStatusUpdateEvent)
		if !ok {
			t.Fatalf("Result type %T, want *a2a.TaskStatusUpdateEvent", result.result)
		}

		if event.TaskID != "task-123" {
			t.Errorf("TaskID = %q, want %q", event.TaskID, "task-123")
		}

		if event.Status.State != a2a.TaskStateCompleted {
			t.Errorf("Status.State = %q, want %q", event.Status.State, a2a.TaskStateCompleted)
		}
	})

	t.Run("unmarshal message", func(t *testing.T) {
		t.Parallel()

		jsonData := []byte(`{
			"kind": "message",
			"messageId": "msg-123",
			"role": "agent",
			"parts": [
				{
					"kind": "text",
					"text": "Response message"
				}
			]
		}`)

		var result sendStreamingMessageResponse
		err := result.UnmarshalJSON(jsonData)
		if err != nil {
			t.Fatalf("UnmarshalJSON() error: %v", err)
		}

		if result.result == nil {
			t.Fatal("UnmarshalJSON() result is nil")
		}

		msg, ok := result.result.(*a2a.Message)
		if !ok {
			t.Fatalf("Result type %T, want *a2a.Message", result.result)
		}

		if msg.MessageID != "msg-123" {
			t.Errorf("MessageID = %q, want %q", msg.MessageID, "msg-123")
		}

		if msg.Role != a2a.RoleAgent {
			t.Errorf("Role = %q, want %q", msg.Role, a2a.RoleAgent)
		}
	})

	t.Run("unmarshal task", func(t *testing.T) {
		t.Parallel()

		jsonData := []byte(`{
			"kind": "task",
			"id": "task-789",
			"contextId": "ctx-abc",
			"status": {
				"state": "failed"
			}
		}`)

		var result sendStreamingMessageResponse
		err := result.UnmarshalJSON(jsonData)
		if err != nil {
			t.Fatalf("UnmarshalJSON() error: %v", err)
		}

		if result.result == nil {
			t.Fatal("UnmarshalJSON() result is nil")
		}

		task, ok := result.result.(*a2a.Task)
		if !ok {
			t.Fatalf("Result type %T, want *a2a.Task", result.result)
		}

		if task.ID != "task-789" {
			t.Errorf("ID = %q, want %q", task.ID, "task-789")
		}

		if task.Status.State != a2a.TaskStateFailed {
			t.Errorf("Status.State = %q, want %q", task.Status.State, a2a.TaskStateFailed)
		}
	})

	t.Run("GetTaskID", func(t *testing.T) {
		t.Parallel()

		result := &sendStreamingMessageResponse{
			result: &a2a.Task{
				ID: "streaming-task-123",
			},
		}

		taskID := result.GetTaskID()
		if taskID != "streaming-task-123" {
			t.Errorf("GetTaskID() = %q, want %q", taskID, "streaming-task-123")
		}
	})

	t.Run("GetEventKind", func(t *testing.T) {
		t.Parallel()

		result := &sendStreamingMessageResponse{
			result: &a2a.TaskArtifactUpdateEvent{
				Kind: a2a.ArtifactUpdateEventKind,
			},
		}

		kind := result.GetEventKind()
		if kind != a2a.ArtifactUpdateEventKind {
			t.Errorf("GetEventKind() = %q, want %q", kind, a2a.ArtifactUpdateEventKind)
		}
	})

	t.Run("nil result", func(t *testing.T) {
		t.Parallel()

		result := &sendStreamingMessageResponse{result: nil}

		taskID := result.GetTaskID()
		if taskID != "" {
			t.Errorf("GetTaskID() = %q, want empty string", taskID)
		}

		kind := result.GetEventKind()
		if kind != a2a.EventKind("") {
			t.Errorf("GetEventKind() = %q, want empty string", kind)
		}
	})

	t.Run("unmarshal invalid JSON", func(t *testing.T) {
		t.Parallel()

		jsonData := []byte(`{invalid json}`)

		var result sendStreamingMessageResponse
		err := result.UnmarshalJSON(jsonData)
		if err == nil {
			t.Fatal("Expected error for invalid JSON")
		}
	})

	t.Run("unmarshal unknown kind", func(t *testing.T) {
		t.Parallel()

		jsonData := []byte(`{
			"kind": "unknown-streaming-kind",
			"taskId": "test-123"
		}`)

		var result sendStreamingMessageResponse
		err := result.UnmarshalJSON(jsonData)
		if err == nil {
			t.Fatal("Expected error for unknown kind")
		}
	})
}
