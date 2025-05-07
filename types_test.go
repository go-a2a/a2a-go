// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/go-cmp/cmp"
)

func TestPartMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		part     Part
		wantJSON string
	}{
		{
			name: "TextPart",
			part: Part{
				Type: PartTypeText,
				Text: "Hello World",
			},
			wantJSON: `{"type":"text","text":"Hello World"}`,
		},
		{
			name: "FilePart",
			part: Part{
				Type: PartTypeFile,
				File: &FileContent{
					Name:     "test.txt",
					MimeType: "text/plain",
					Bytes:    "SGVsbG8gV29ybGQ=",
				},
			},
			wantJSON: `{"type":"file","file":{"name":"test.txt","mimeType":"text/plain","bytes":"SGVsbG8gV29ybGQ="}}`,
		},
		{
			name: "DataPart",
			part: Part{
				Type: PartTypeData,
				Data: json.RawMessage(`{"key":"value"}`),
			},
			wantJSON: `{"type":"data","data":{"key":"value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			got, err := sonic.ConfigFastest.Marshal(tt.part)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if diff := cmp.Diff(string(got), tt.wantJSON); diff != "" {
				t.Errorf("Marshal() mismatch (-got +want):\n%s", diff)
			}

			// Test unmarshaling
			var gotPart Part
			if err := sonic.ConfigFastest.Unmarshal([]byte(tt.wantJSON), &gotPart); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Use go-cmp to compare the structs
			if diff := cmp.Diff(gotPart, tt.part, cmp.AllowUnexported()); diff != "" {
				t.Errorf("Unmarshal() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestMessageMarshaling(t *testing.T) {
	message := Message{
		Role: MessageRoleUser,
		Parts: []Part{
			{
				Type: PartTypeText,
				Text: "Hello",
			},
			{
				Type: PartTypeData,
				Data: json.RawMessage(`{"form":{"name":"John"}}`),
			},
		},
	}

	// Test marshaling
	data, err := sonic.ConfigFastest.Marshal(message)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Test unmarshaling
	var gotMessage Message
	if err := sonic.ConfigFastest.Unmarshal(data, &gotMessage); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Use go-cmp to compare
	if diff := cmp.Diff(gotMessage, message); diff != "" {
		t.Errorf("Unmarshal() mismatch (-got +want):\n%s", diff)
	}
}

func TestTaskMarshaling(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	task := Task{
		ID:        "task-123",
		SessionID: "session-456",
		Status: TaskStatus{
			State:     TaskStateWorking,
			Timestamp: now,
			Message: &Message{
				Role: MessageRoleAgent,
				Parts: []Part{
					{
						Type: PartTypeText,
						Text: "Working on it...",
					},
				},
			},
		},
		Artifacts: []Artifact{
			{
				Name:  "result",
				Index: 0,
				Parts: []Part{
					{
						Type: PartTypeText,
						Text: "Partial result",
					},
				},
			},
		},
	}

	// Test marshaling
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Test unmarshaling
	var gotTask Task
	if err := json.Unmarshal(data, &gotTask); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Use go-cmp to compare
	if diff := cmp.Diff(gotTask, task); diff != "" {
		t.Errorf("Unmarshal() mismatch (-got +want):\n%s", diff)
	}
}
