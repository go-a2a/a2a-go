// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"strings"
	"testing"
)

func TestRole_Constants(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		expected string
	}{
		{"RoleAgent", RoleAgent, "agent"},
		{"RoleUser", RoleUser, "user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.role) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.role))
			}
		})
	}
}

func TestMessagePart_Validate(t *testing.T) {
	tests := []struct {
		name      string
		part      MessagePart
		wantError bool
	}{
		{
			name: "valid text part",
			part: MessagePart{
				Root: &TextPart{
					Kind: "text",
					Text: "Hello, world!",
				},
			},
			wantError: false,
		},
		{
			name: "nil root",
			part: MessagePart{
				Root: nil,
			},
			wantError: true,
		},
		{
			name: "invalid text part",
			part: MessagePart{
				Root: &TextPart{
					Kind: "text",
					Text: "",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.part.Validate()
			if tt.wantError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
		})
	}
}

func TestA2AMessage_Validate(t *testing.T) {
	validPart := &MessagePart{
		Root: &TextPart{
			Kind: "text",
			Text: "Hello, world!",
		},
	}

	tests := []struct {
		name      string
		message   A2AMessage
		wantError bool
	}{
		{
			name: "valid message",
			message: A2AMessage{
				Role:      RoleAgent,
				Parts:     []*MessagePart{validPart},
				MessageID: "msg-123",
			},
			wantError: false,
		},
		{
			name: "invalid role",
			message: A2AMessage{
				Role:      Role("invalid"),
				Parts:     []*MessagePart{validPart},
				MessageID: "msg-123",
			},
			wantError: true,
		},
		{
			name: "empty message ID",
			message: A2AMessage{
				Role:      RoleAgent,
				Parts:     []*MessagePart{validPart},
				MessageID: "",
			},
			wantError: true,
		},
		{
			name: "empty parts",
			message: A2AMessage{
				Role:      RoleAgent,
				Parts:     []*MessagePart{},
				MessageID: "msg-123",
			},
			wantError: true,
		},
		{
			name: "nil part in parts",
			message: A2AMessage{
				Role:      RoleAgent,
				Parts:     []*MessagePart{nil},
				MessageID: "msg-123",
			},
			wantError: true,
		},
		{
			name: "invalid part in parts",
			message: A2AMessage{
				Role: RoleAgent,
				Parts: []*MessagePart{
					{
						Root: &TextPart{
							Kind: "text",
							Text: "",
						},
					},
				},
				MessageID: "msg-123",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.Validate()
			if tt.wantError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
		})
	}
}

func TestNewAgentTextMessage(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		contextID   *string
		taskID      *string
		wantError   bool
		expectRole  Role
		expectParts int
	}{
		{
			name:        "valid text message",
			text:        "Hello, world!",
			contextID:   nil,
			taskID:      nil,
			wantError:   false,
			expectRole:  RoleAgent,
			expectParts: 1,
		},
		{
			name:        "empty text",
			text:        "",
			contextID:   nil,
			taskID:      nil,
			wantError:   true,
			expectRole:  "",
			expectParts: 0,
		},
		{
			name:        "with context and task IDs",
			text:        "Hello, world!",
			contextID:   stringPtr("ctx-123"),
			taskID:      stringPtr("task-456"),
			wantError:   false,
			expectRole:  RoleAgent,
			expectParts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := NewAgentTextMessage(tt.text, tt.contextID, tt.taskID)
			if tt.wantError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
			if tt.wantError {
				return
			}

			if msg.Role != tt.expectRole {
				t.Errorf("Expected role %s, got %s", tt.expectRole, msg.Role)
			}
			if len(msg.Parts) != tt.expectParts {
				t.Errorf("Expected %d parts, got %d", tt.expectParts, len(msg.Parts))
			}
			if msg.MessageID == "" {
				t.Error("Expected non-empty MessageID")
			}
			if tt.contextID != nil && msg.ContextID != tt.contextID {
				t.Errorf("Expected ContextID %v, got %v", tt.contextID, msg.ContextID)
			}
			if tt.taskID != nil && msg.TaskID != tt.taskID {
				t.Errorf("Expected TaskID %v, got %v", tt.taskID, msg.TaskID)
			}

			// Validate the text part
			if len(msg.Parts) > 0 {
				textPart, ok := msg.Parts[0].Root.(*TextPart)
				if !ok {
					t.Error("Expected TextPart in first part")
				} else if textPart.Text != tt.text {
					t.Errorf("Expected text %s, got %s", tt.text, textPart.Text)
				}
			}
		})
	}
}

func TestNewAgentPartsMessage(t *testing.T) {
	validPart := &MessagePart{
		Root: &TextPart{
			Kind: "text",
			Text: "Hello, world!",
		},
	}

	tests := []struct {
		name        string
		parts       []*MessagePart
		contextID   *string
		taskID      *string
		wantError   bool
		expectRole  Role
		expectParts int
	}{
		{
			name:        "valid parts message",
			parts:       []*MessagePart{validPart},
			contextID:   nil,
			taskID:      nil,
			wantError:   false,
			expectRole:  RoleAgent,
			expectParts: 1,
		},
		{
			name:        "empty parts",
			parts:       []*MessagePart{},
			contextID:   nil,
			taskID:      nil,
			wantError:   true,
			expectRole:  "",
			expectParts: 0,
		},
		{
			name:        "nil parts",
			parts:       nil,
			contextID:   nil,
			taskID:      nil,
			wantError:   true,
			expectRole:  "",
			expectParts: 0,
		},
		{
			name:        "nil part in parts",
			parts:       []*MessagePart{nil},
			contextID:   nil,
			taskID:      nil,
			wantError:   true,
			expectRole:  "",
			expectParts: 0,
		},
		{
			name: "invalid part in parts",
			parts: []*MessagePart{
				{
					Root: &TextPart{
						Kind: "text",
						Text: "",
					},
				},
			},
			contextID:   nil,
			taskID:      nil,
			wantError:   true,
			expectRole:  "",
			expectParts: 0,
		},
		{
			name:        "multiple valid parts",
			parts:       []*MessagePart{validPart, validPart},
			contextID:   stringPtr("ctx-123"),
			taskID:      stringPtr("task-456"),
			wantError:   false,
			expectRole:  RoleAgent,
			expectParts: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := NewAgentPartsMessage(tt.parts, tt.contextID, tt.taskID)
			if tt.wantError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
			if tt.wantError {
				return
			}

			if msg.Role != tt.expectRole {
				t.Errorf("Expected role %s, got %s", tt.expectRole, msg.Role)
			}
			if len(msg.Parts) != tt.expectParts {
				t.Errorf("Expected %d parts, got %d", tt.expectParts, len(msg.Parts))
			}
			if msg.MessageID == "" {
				t.Error("Expected non-empty MessageID")
			}
			if tt.contextID != nil && msg.ContextID != tt.contextID {
				t.Errorf("Expected ContextID %v, got %v", tt.contextID, msg.ContextID)
			}
			if tt.taskID != nil && msg.TaskID != tt.taskID {
				t.Errorf("Expected TaskID %v, got %v", tt.taskID, msg.TaskID)
			}
		})
	}
}

func TestGetTextParts(t *testing.T) {
	tests := []struct {
		name     string
		parts    []*MessagePart
		expected []string
	}{
		{
			name:     "empty parts",
			parts:    []*MessagePart{},
			expected: []string{},
		},
		{
			name:     "nil parts",
			parts:    nil,
			expected: []string{},
		},
		{
			name: "single text part",
			parts: []*MessagePart{
				{
					Root: &TextPart{
						Kind: "text",
						Text: "Hello, world!",
					},
				},
			},
			expected: []string{"Hello, world!"},
		},
		{
			name: "multiple text parts",
			parts: []*MessagePart{
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
			expected: []string{"Hello", "World"},
		},
		{
			name: "mixed parts with non-text",
			parts: []*MessagePart{
				{
					Root: &TextPart{
						Kind: "text",
						Text: "Hello",
					},
				},
				{
					Root: &DataPart{
						Kind: "data",
						Data: map[string]any{"key": "value"},
					},
				},
			},
			expected: []string{"Hello"},
		},
		{
			name: "nil part",
			parts: []*MessagePart{
				{
					Root: &TextPart{
						Kind: "text",
						Text: "Hello",
					},
				},
				nil,
			},
			expected: []string{"Hello"},
		},
		{
			name: "part with nil root",
			parts: []*MessagePart{
				{
					Root: &TextPart{
						Kind: "text",
						Text: "Hello",
					},
				},
				{
					Root: nil,
				},
			},
			expected: []string{"Hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTextParts(tt.parts)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d text parts, got %d", len(tt.expected), len(result))
			}
			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected text part %d to be %s, got %s", i, expected, result[i])
				}
			}
		})
	}
}

func TestGetMessageText(t *testing.T) {
	tests := []struct {
		name      string
		message   *A2AMessage
		delimiter string
		expected  string
	}{
		{
			name:      "nil message",
			message:   nil,
			delimiter: " ",
			expected:  "",
		},
		{
			name: "single text part",
			message: &A2AMessage{
				Parts: []*MessagePart{
					{
						Root: &TextPart{
							Kind: "text",
							Text: "Hello, world!",
						},
					},
				},
			},
			delimiter: " ",
			expected:  "Hello, world!",
		},
		{
			name: "multiple text parts with space delimiter",
			message: &A2AMessage{
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
			},
			delimiter: " ",
			expected:  "Hello World",
		},
		{
			name: "multiple text parts with custom delimiter",
			message: &A2AMessage{
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
			},
			delimiter: ", ",
			expected:  "Hello, World",
		},
		{
			name: "no text parts",
			message: &A2AMessage{
				Parts: []*MessagePart{
					{
						Root: &DataPart{
							Kind: "data",
							Data: map[string]any{"key": "value"},
						},
					},
				},
			},
			delimiter: " ",
			expected:  "",
		},
		{
			name: "empty parts",
			message: &A2AMessage{
				Parts: []*MessagePart{},
			},
			delimiter: " ",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMessageText(tt.message, tt.delimiter)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetMessageTextWithNewlines(t *testing.T) {
	tests := []struct {
		name     string
		message  *A2AMessage
		expected string
	}{
		{
			name:     "nil message",
			message:  nil,
			expected: "",
		},
		{
			name: "single text part",
			message: &A2AMessage{
				Parts: []*MessagePart{
					{
						Root: &TextPart{
							Kind: "text",
							Text: "Hello, world!",
						},
					},
				},
			},
			expected: "Hello, world!",
		},
		{
			name: "multiple text parts",
			message: &A2AMessage{
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
			},
			expected: "Hello\nWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMessageTextWithNewlines(tt.message)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Test that generated UUIDs are unique and valid
func TestGeneratedUUIDs(t *testing.T) {
	// Create multiple messages to verify UUID uniqueness
	messageIDs := make(map[string]bool)
	for i := 0; i < 100; i++ {
		msg, err := NewAgentTextMessage("test", nil, nil)
		if err != nil {
			t.Fatalf("Failed to create message: %v", err)
		}

		// Check if UUID is unique
		if messageIDs[msg.MessageID] {
			t.Errorf("Duplicate UUID generated: %s", msg.MessageID)
		}
		messageIDs[msg.MessageID] = true

		// Basic UUID format check (should have 4 dashes)
		if strings.Count(msg.MessageID, "-") != 4 {
			t.Errorf("Invalid UUID format: %s", msg.MessageID)
		}
	}
}

// Benchmark tests
func BenchmarkNewAgentTextMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := NewAgentTextMessage("Hello, world!", nil, nil)
		if err != nil {
			b.Fatalf("Failed to create message: %v", err)
		}
	}
}

func BenchmarkGetTextParts(b *testing.B) {
	parts := []*MessagePart{
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
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetTextParts(parts)
	}
}

func BenchmarkGetMessageText(b *testing.B) {
	message := &A2AMessage{
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
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetMessageText(message, " ")
	}
}
