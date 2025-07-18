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

	"github.com/google/uuid"
)

func TestNewAgentTextMessage(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		text      string
		contextID string
		taskID    string
		want      func(*Message) bool
	}{
		"with all parameters": {
			text:      "Hello, world!",
			contextID: "ctx-123",
			taskID:    "task-456",
			want: func(msg *Message) bool {
				if msg.Kind != MessageEventKind {
					t.Errorf("expected Kind %s, got %s", MessageEventKind, msg.Kind)
					return false
				}
				if msg.Role != RoleAgent {
					t.Errorf("expected role %s, got %s", RoleAgent, msg.Role)
					return false
				}
				if msg.ContextID != "ctx-123" {
					t.Errorf("expected contextID ctx-123, got %s", msg.ContextID)
					return false
				}
				if msg.TaskID != "task-456" {
					t.Errorf("expected taskID task-456, got %s", msg.TaskID)
					return false
				}
				if len(msg.Parts) != 1 {
					t.Errorf("expected 1 part, got %d", len(msg.Parts))
					return false
				}
				textPart, ok := msg.Parts[0].(*TextPart)
				if !ok {
					t.Error("expected first part to be TextPart")
					return false
				}
				if textPart.Kind != TextPartKind {
					t.Errorf("expected kind %s, got %s", TextPartKind, textPart.Kind)
					return false
				}
				if textPart.Text != "Hello, world!" {
					t.Errorf("expected text 'Hello, world!', got %s", textPart.Text)
					return false
				}
				return true
			},
		},
		"with empty text": {
			text:      "",
			contextID: "ctx-789",
			taskID:    "task-012",
			want: func(msg *Message) bool {
				if msg.Kind != MessageEventKind {
					t.Errorf("expected Kind %s, got %s", MessageEventKind, msg.Kind)
					return false
				}
				if msg.Role != RoleAgent {
					t.Errorf("expected role %s, got %s", RoleAgent, msg.Role)
					return false
				}
				if len(msg.Parts) != 1 {
					t.Errorf("expected 1 part, got %d", len(msg.Parts))
					return false
				}
				textPart, ok := msg.Parts[0].(*TextPart)
				if !ok {
					t.Error("expected first part to be TextPart")
					return false
				}
				if textPart.Text != "" {
					t.Errorf("expected empty text, got %s", textPart.Text)
					return false
				}
				return true
			},
		},
		"with empty contextID and taskID": {
			text:      "Test message",
			contextID: "",
			taskID:    "",
			want: func(msg *Message) bool {
				if msg.Kind != MessageEventKind {
					t.Errorf("expected Kind %s, got %s", MessageEventKind, msg.Kind)
					return false
				}
				if msg.Role != RoleAgent {
					t.Errorf("expected role %s, got %s", RoleAgent, msg.Role)
					return false
				}
				if msg.ContextID != "" {
					t.Errorf("expected empty contextID, got %s", msg.ContextID)
					return false
				}
				if msg.TaskID != "" {
					t.Errorf("expected empty taskID, got %s", msg.TaskID)
					return false
				}
				return true
			},
		},
		"with multiline text": {
			text:      "Line 1\nLine 2\nLine 3",
			contextID: "ctx-multi",
			taskID:    "task-multi",
			want: func(msg *Message) bool {
				textPart, ok := msg.Parts[0].(*TextPart)
				if !ok {
					t.Error("expected first part to be TextPart")
					return false
				}
				if textPart.Text != "Line 1\nLine 2\nLine 3" {
					t.Errorf("expected multiline text, got %s", textPart.Text)
					return false
				}
				return true
			},
		},
		"with special characters": {
			text:      "Hello\t\"world\"\n!@#$%^&*()",
			contextID: "ctx-special",
			taskID:    "task-special",
			want: func(msg *Message) bool {
				textPart, ok := msg.Parts[0].(*TextPart)
				if !ok {
					t.Error("expected first part to be TextPart")
					return false
				}
				if textPart.Text != "Hello\t\"world\"\n!@#$%^&*()" {
					t.Errorf("expected special characters text, got %s", textPart.Text)
					return false
				}
				return true
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := NewAgentTextMessage(tt.text, tt.contextID, tt.taskID)

			// Check if MessageID is a valid UUID
			if _, err := uuid.Parse(got.MessageID); err != nil {
				t.Errorf("expected valid UUID for MessageID, got %s", got.MessageID)
			}

			if !tt.want(got) {
				t.Errorf("NewAgentTextMessage() validation failed")
			}
		})
	}
}

func TestNewAgentPartsMessage(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		parts     []Part
		contextID string
		taskID    string
		want      func(*Message) bool
	}{
		"with multiple parts": {
			parts: []Part{
				&TextPart{Kind: TextPartKind, Text: "Hello"},
				&TextPart{Kind: TextPartKind, Text: "World"},
				&DataPart{Kind: DataPartKind, Data: map[string]any{"key": "value"}},
			},
			contextID: "ctx-multi",
			taskID:    "task-multi",
			want: func(msg *Message) bool {
				if msg.Kind != MessageEventKind {
					t.Errorf("expected Kind %s, got %s", MessageEventKind, msg.Kind)
					return false
				}
				if msg.Role != RoleAgent {
					t.Errorf("expected role %s, got %s", RoleAgent, msg.Role)
					return false
				}
				if len(msg.Parts) != 3 {
					t.Errorf("expected 3 parts, got %d", len(msg.Parts))
					return false
				}
				if msg.ContextID != "ctx-multi" {
					t.Errorf("expected contextID ctx-multi, got %s", msg.ContextID)
					return false
				}
				if msg.TaskID != "task-multi" {
					t.Errorf("expected taskID task-multi, got %s", msg.TaskID)
					return false
				}
				return true
			},
		},
		"with empty parts slice": {
			parts:     []Part{},
			contextID: "ctx-empty",
			taskID:    "task-empty",
			want: func(msg *Message) bool {
				if msg.Kind != MessageEventKind {
					t.Errorf("expected Kind %s, got %s", MessageEventKind, msg.Kind)
					return false
				}
				if msg.Role != RoleAgent {
					t.Errorf("expected role %s, got %s", RoleAgent, msg.Role)
					return false
				}
				if len(msg.Parts) != 0 {
					t.Errorf("expected 0 parts, got %d", len(msg.Parts))
					return false
				}
				return true
			},
		},
		"with nil parts slice": {
			parts:     nil,
			contextID: "ctx-nil",
			taskID:    "task-nil",
			want: func(msg *Message) bool {
				if msg.Kind != MessageEventKind {
					t.Errorf("expected Kind %s, got %s", MessageEventKind, msg.Kind)
					return false
				}
				if msg.Role != RoleAgent {
					t.Errorf("expected role %s, got %s", RoleAgent, msg.Role)
					return false
				}
				if msg.Parts != nil {
					t.Error("expected nil parts slice")
					return false
				}
				return true
			},
		},
		"with FilePart": {
			parts: []Part{
				&FilePart{
					Kind: FilePartKind,
					File: &FileWithBytes{
						Bytes:    "base64encodedcontent",
						MIMEType: "text/plain",
						Name:     "test.txt",
					},
				},
				&FilePart{
					Kind: FilePartKind,
					File: &FileWithURI{
						URI:      "https://example.com/file.pdf",
						MIMEType: "application/pdf",
						Name:     "document.pdf",
					},
				},
			},
			contextID: "ctx-file",
			taskID:    "task-file",
			want: func(msg *Message) bool {
				if msg.Kind != MessageEventKind {
					t.Errorf("expected Kind %s, got %s", MessageEventKind, msg.Kind)
					return false
				}
				if len(msg.Parts) != 2 {
					t.Errorf("expected 2 parts, got %d", len(msg.Parts))
					return false
				}
				filePart1, ok := msg.Parts[0].(*FilePart)
				if !ok {
					t.Error("expected first part to be FilePart")
					return false
				}
				if filePart1.Kind != FilePartKind {
					t.Errorf("expected kind %s, got %s", FilePartKind, filePart1.Kind)
					return false
				}
				return true
			},
		},
		"with metadata in parts": {
			parts: []Part{
				&TextPart{
					Kind:     TextPartKind,
					Text:     "Text with metadata",
					Metadata: map[string]any{"author": "test", "version": 1},
				},
			},
			contextID: "ctx-meta",
			taskID:    "task-meta",
			want: func(msg *Message) bool {
				textPart, ok := msg.Parts[0].(*TextPart)
				if !ok {
					t.Error("expected first part to be TextPart")
					return false
				}
				if textPart.Metadata["author"] != "test" {
					t.Errorf("expected metadata author='test', got %v", textPart.Metadata["author"])
					return false
				}
				if textPart.Metadata["version"] != 1 {
					t.Errorf("expected metadata version=1, got %v", textPart.Metadata["version"])
					return false
				}
				return true
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := NewAgentPartsMessage(tt.parts, tt.contextID, tt.taskID)

			// Check if MessageID is a valid UUID
			if _, err := uuid.Parse(got.MessageID); err != nil {
				t.Errorf("expected valid UUID for MessageID, got %s", got.MessageID)
			}

			if !tt.want(got) {
				t.Errorf("NewAgentPartsMessage() validation failed")
			}
		})
	}
}

func TestGetTextParts(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		parts []Part
		want  []string
	}{
		"only TextParts": {
			parts: []Part{
				&TextPart{Kind: TextPartKind, Text: "First"},
				&TextPart{Kind: TextPartKind, Text: "Second"},
				&TextPart{Kind: TextPartKind, Text: "Third"},
			},
			want: []string{"First", "Second", "Third"},
		},
		"mixed parts": {
			parts: []Part{
				&TextPart{Kind: TextPartKind, Text: "Text1"},
				&DataPart{Kind: DataPartKind, Data: map[string]any{"key": "value"}},
				&TextPart{Kind: TextPartKind, Text: "Text2"},
				&FilePart{
					Kind: FilePartKind,
					File: &FileWithBytes{Bytes: "content"},
				},
				&TextPart{Kind: TextPartKind, Text: "Text3"},
			},
			want: []string{"Text1", "Text2", "Text3"},
		},
		"no TextParts": {
			parts: []Part{
				&DataPart{Kind: DataPartKind, Data: map[string]any{"key": "value"}},
				&FilePart{
					Kind: FilePartKind,
					File: &FileWithURI{URI: "https://example.com"},
				},
			},
			want: []string{},
		},
		"empty slice": {
			parts: []Part{},
			want:  []string{},
		},
		"nil slice": {
			parts: nil,
			want:  []string{},
		},
		"with nil parts in slice": {
			parts: []Part{
				&TextPart{Kind: TextPartKind, Text: "First"},
				nil,
				&TextPart{Kind: TextPartKind, Text: "Second"},
				nil,
			},
			want: []string{"First", "Second"},
		},
		"TextParts with empty text": {
			parts: []Part{
				&TextPart{Kind: TextPartKind, Text: ""},
				&TextPart{Kind: TextPartKind, Text: "Non-empty"},
				&TextPart{Kind: TextPartKind, Text: ""},
			},
			want: []string{"", "Non-empty", ""},
		},
		"TextParts with special characters": {
			parts: []Part{
				&TextPart{Kind: TextPartKind, Text: "Line1\nLine2"},
				&TextPart{Kind: TextPartKind, Text: "Tab\tSeparated"},
				&TextPart{Kind: TextPartKind, Text: "\"Quoted\""},
			},
			want: []string{"Line1\nLine2", "Tab\tSeparated", "\"Quoted\""},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := GetTextParts(tt.parts)
			if len(got) != len(tt.want) {
				t.Errorf("GetTextParts() returned %d elements, want %d", len(got), len(tt.want))
				return
			}
			for i, text := range got {
				if text != tt.want[i] {
					t.Errorf("GetTextParts()[%d] = %q, want %q", i, text, tt.want[i])
				}
			}
		})
	}
}

func TestGetFileParts(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		parts []Part
		want  int // number of expected file parts
		check func([]File) bool
	}{
		"only FileParts with FileWithBytes": {
			parts: []Part{
				&FilePart{
					Kind: FilePartKind,
					File: &FileWithBytes{
						Bytes:    "content1",
						MIMEType: "text/plain",
						Name:     "file1.txt",
					},
				},
				&FilePart{
					Kind: FilePartKind,
					File: &FileWithBytes{
						Bytes:    "content2",
						MIMEType: "image/png",
						Name:     "image.png",
					},
				},
			},
			want: 2,
			check: func(files []File) bool {
				if files[0].GetMIMEType() != "text/plain" {
					t.Errorf("expected MIMEType text/plain, got %s", files[0].GetMIMEType())
					return false
				}
				if files[0].GetName() != "file1.txt" {
					t.Errorf("expected name file1.txt, got %s", files[0].GetName())
					return false
				}
				return true
			},
		},
		"only FileParts with FileWithURI": {
			parts: []Part{
				&FilePart{
					Kind: FilePartKind,
					File: &FileWithURI{
						URI:      "https://example.com/doc.pdf",
						MIMEType: "application/pdf",
						Name:     "document.pdf",
					},
				},
			},
			want: 1,
			check: func(files []File) bool {
				fileWithURI, ok := files[0].(*FileWithURI)
				if !ok {
					t.Error("expected FileWithURI")
					return false
				}
				if fileWithURI.URI != "https://example.com/doc.pdf" {
					t.Errorf("expected URI https://example.com/doc.pdf, got %s", fileWithURI.URI)
					return false
				}
				return true
			},
		},
		"mixed FileParts (FileWithBytes and FileWithURI)": {
			parts: []Part{
				&FilePart{
					Kind: FilePartKind,
					File: &FileWithBytes{
						Bytes: "bytes-content",
						Name:  "bytes.bin",
					},
				},
				&FilePart{
					Kind: FilePartKind,
					File: &FileWithURI{
						URI:  "https://example.com/remote.txt",
						Name: "remote.txt",
					},
				},
			},
			want: 2,
			check: func(files []File) bool {
				_, isByytes := files[0].(*FileWithBytes)
				_, isURI := files[1].(*FileWithURI)
				return isByytes && isURI
			},
		},
		"mixed parts with non-File parts": {
			parts: []Part{
				&TextPart{Kind: TextPartKind, Text: "Text"},
				&FilePart{
					Kind: FilePartKind,
					File: &FileWithBytes{Bytes: "content"},
				},
				&DataPart{Kind: DataPartKind, Data: map[string]any{"key": "value"}},
				&FilePart{
					Kind: FilePartKind,
					File: &FileWithURI{URI: "https://example.com"},
				},
			},
			want: 2,
			check: func(files []File) bool {
				return len(files) == 2
			},
		},
		"no FileParts": {
			parts: []Part{
				&TextPart{Kind: TextPartKind, Text: "Text"},
				&DataPart{Kind: DataPartKind, Data: map[string]any{"key": "value"}},
			},
			want: 0,
			check: func(files []File) bool {
				return len(files) == 0
			},
		},
		"empty slice": {
			parts: []Part{},
			want:  0,
			check: func(files []File) bool {
				return len(files) == 0
			},
		},
		"nil slice": {
			parts: nil,
			want:  0,
			check: func(files []File) bool {
				return len(files) == 0
			},
		},
		"with nil parts in slice": {
			parts: []Part{
				&FilePart{
					Kind: FilePartKind,
					File: &FileWithBytes{Bytes: "content1"},
				},
				nil,
				&FilePart{
					Kind: FilePartKind,
					File: &FileWithURI{URI: "https://example.com"},
				},
				nil,
			},
			want: 2,
			check: func(files []File) bool {
				return len(files) == 2
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := GetFileParts(tt.parts)
			if len(got) != tt.want {
				t.Errorf("GetFileParts() returned %d elements, want %d", len(got), tt.want)
				return
			}
			if tt.check != nil && !tt.check(got) {
				t.Errorf("GetFileParts() validation failed")
			}
		})
	}
}

func TestGetMessageText(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		message   *Message
		delimiter string
		want      string
	}{
		"multiple TextParts with space delimiter": {
			message: &Message{
				Parts: []Part{
					&TextPart{Kind: TextPartKind, Text: "Hello"},
					&TextPart{Kind: TextPartKind, Text: "World"},
					&TextPart{Kind: TextPartKind, Text: "!"},
				},
			},
			delimiter: " ",
			want:      "Hello World !",
		},
		"multiple TextParts with newline delimiter": {
			message: &Message{
				Parts: []Part{
					&TextPart{Kind: TextPartKind, Text: "Line 1"},
					&TextPart{Kind: TextPartKind, Text: "Line 2"},
					&TextPart{Kind: TextPartKind, Text: "Line 3"},
				},
			},
			delimiter: "\n",
			want:      "Line 1\nLine 2\nLine 3",
		},
		"single TextPart": {
			message: &Message{
				Parts: []Part{
					&TextPart{Kind: TextPartKind, Text: "Single text part"},
				},
			},
			delimiter: " ",
			want:      "Single text part",
		},
		"mixed parts with TextParts": {
			message: &Message{
				Parts: []Part{
					&TextPart{Kind: TextPartKind, Text: "Text1"},
					&DataPart{Kind: DataPartKind, Data: map[string]any{"key": "value"}},
					&TextPart{Kind: TextPartKind, Text: "Text2"},
					&FilePart{
						Kind: FilePartKind,
						File: &FileWithBytes{Bytes: "content"},
					},
					&TextPart{Kind: TextPartKind, Text: "Text3"},
				},
			},
			delimiter: "-",
			want:      "Text1-Text2-Text3",
		},
		"no TextParts": {
			message: &Message{
				Parts: []Part{
					&DataPart{Kind: DataPartKind, Data: map[string]any{"key": "value"}},
					&FilePart{
						Kind: FilePartKind,
						File: &FileWithURI{URI: "https://example.com"},
					},
				},
			},
			delimiter: " ",
			want:      "",
		},
		"nil message": {
			message:   nil,
			delimiter: " ",
			want:      "",
		},
		"empty parts": {
			message: &Message{
				Parts: []Part{},
			},
			delimiter: " ",
			want:      "",
		},
		"empty delimiter": {
			message: &Message{
				Parts: []Part{
					&TextPart{Kind: TextPartKind, Text: "A"},
					&TextPart{Kind: TextPartKind, Text: "B"},
					&TextPart{Kind: TextPartKind, Text: "C"},
				},
			},
			delimiter: "",
			want:      "ABC",
		},
		"custom delimiter with special characters": {
			message: &Message{
				Parts: []Part{
					&TextPart{Kind: TextPartKind, Text: "Part1"},
					&TextPart{Kind: TextPartKind, Text: "Part2"},
					&TextPart{Kind: TextPartKind, Text: "Part3"},
				},
			},
			delimiter: " | ",
			want:      "Part1 | Part2 | Part3",
		},
		"TextParts with empty text": {
			message: &Message{
				Parts: []Part{
					&TextPart{Kind: TextPartKind, Text: "Start"},
					&TextPart{Kind: TextPartKind, Text: ""},
					&TextPart{Kind: TextPartKind, Text: "End"},
				},
			},
			delimiter: "-",
			want:      "Start--End",
		},
		"message with metadata and other properties": {
			message: &Message{
				Role:      RoleAgent,
				MessageID: "msg-123",
				ContextID: "ctx-123",
				TaskID:    "task-123",
				Parts: []Part{
					&TextPart{Kind: TextPartKind, Text: "Text with"},
					&TextPart{Kind: TextPartKind, Text: "metadata"},
				},
			},
			delimiter: " ",
			want:      "Text with metadata",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := GetMessageText(tt.message, tt.delimiter)
			if got != tt.want {
				t.Errorf("GetMessageText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMessageConstructorsIntegration(t *testing.T) {
	t.Parallel()

	// Test that NewAgentTextMessage and GetMessageText work together
	t.Run("NewAgentTextMessage with GetMessageText", func(t *testing.T) {
		t.Parallel()

		text := "This is a test message"
		msg := NewAgentTextMessage(text, "ctx-test", "task-test")

		result := GetMessageText(msg, " ")
		if result != text {
			t.Errorf("GetMessageText() = %q, want %q", result, text)
		}
	})

	// Test that NewAgentPartsMessage with multiple TextParts and GetMessageText work together
	t.Run("NewAgentPartsMessage with GetMessageText", func(t *testing.T) {
		t.Parallel()

		parts := []Part{
			&TextPart{Kind: TextPartKind, Text: "First"},
			&TextPart{Kind: TextPartKind, Text: "Second"},
			&TextPart{Kind: TextPartKind, Text: "Third"},
		}
		msg := NewAgentPartsMessage(parts, "ctx-test", "task-test")

		result := GetMessageText(msg, " ")
		expected := "First Second Third"
		if result != expected {
			t.Errorf("GetMessageText() = %q, want %q", result, expected)
		}
	})

	// Test GetTextParts with message created by NewAgentTextMessage
	t.Run("GetTextParts with NewAgentTextMessage", func(t *testing.T) {
		t.Parallel()

		text := "Test text"
		msg := NewAgentTextMessage(text, "", "")

		textParts := GetTextParts(msg.Parts)
		if len(textParts) != 1 {
			t.Errorf("expected 1 text part, got %d", len(textParts))
		}
		if textParts[0] != text {
			t.Errorf("expected text %q, got %q", text, textParts[0])
		}
	})

	// Test GetFileParts with message containing file parts
	t.Run("GetFileParts with NewAgentPartsMessage", func(t *testing.T) {
		t.Parallel()

		parts := []Part{
			&TextPart{Kind: TextPartKind, Text: "Text"},
			&FilePart{
				Kind: FilePartKind,
				File: &FileWithBytes{
					Bytes:    "content",
					MIMEType: "text/plain",
					Name:     "test.txt",
				},
			},
			&FilePart{
				Kind: FilePartKind,
				File: &FileWithURI{
					URI:      "https://example.com/file.pdf",
					MIMEType: "application/pdf",
					Name:     "document.pdf",
				},
			},
		}
		msg := NewAgentPartsMessage(parts, "ctx-file", "task-file")

		fileParts := GetFileParts(msg.Parts)
		if len(fileParts) != 2 {
			t.Errorf("expected 2 file parts, got %d", len(fileParts))
		}

		// Check first file part
		file1, ok := fileParts[0].(*FileWithBytes)
		if !ok {
			t.Error("expected first file to be FileWithBytes")
		}
		if file1.Name != "test.txt" {
			t.Errorf("expected file name test.txt, got %s", file1.Name)
		}

		// Check second file part
		file2, ok := fileParts[1].(*FileWithURI)
		if !ok {
			t.Error("expected second file to be FileWithURI")
		}
		if file2.URI != "https://example.com/file.pdf" {
			t.Errorf("expected URI https://example.com/file.pdf, got %s", file2.URI)
		}
	})
}

func TestEdgeCases(t *testing.T) {
	t.Parallel()

	// Test with very long text
	t.Run("very long text", func(t *testing.T) {
		t.Parallel()

		longText := strings.Repeat("A", 10000)
		msg := NewAgentTextMessage(longText, "", "")

		result := GetMessageText(msg, "")
		if result != longText {
			t.Error("failed to handle very long text")
		}
	})

	// Test with Unicode characters
	t.Run("unicode characters", func(t *testing.T) {
		t.Parallel()

		unicodeText := "Hello 世界 🌍 مرحبا"
		msg := NewAgentTextMessage(unicodeText, "", "")

		result := GetMessageText(msg, " ")
		if result != unicodeText {
			t.Errorf("failed to handle unicode text: got %q, want %q", result, unicodeText)
		}
	})

	// Test all nil parts in GetTextParts
	t.Run("all nil parts in GetTextParts", func(t *testing.T) {
		t.Parallel()

		parts := []Part{nil, nil, nil}
		result := GetTextParts(parts)

		if len(result) != 0 {
			t.Errorf("expected empty result for all nil parts, got %d elements", len(result))
		}
	})

	// Test all nil parts in GetFileParts
	t.Run("all nil parts in GetFileParts", func(t *testing.T) {
		t.Parallel()

		parts := []Part{nil, nil, nil}
		result := GetFileParts(parts)

		if len(result) != 0 {
			t.Errorf("expected empty result for all nil parts, got %d elements", len(result))
		}
	})
}
