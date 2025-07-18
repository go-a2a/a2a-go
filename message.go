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

	"github.com/google/uuid"
)

// NewAgentTextMessage creates a new agent message containing a single TextPart.
//
// This function mirrors the Python new_agent_text_message function behavior.
//
// Args:
//
//	text: The text content of the message.
//	contextID: The context ID for the message (optional).
//	taskID: The task ID for the message (optional).
//
// Returns:
//
//	A new A2AMessage object with role 'agent' and an error if the operation fails.
func NewAgentTextMessage(text, contextID, taskID string) *Message {
	// Create TextPart with "text" kind
	textPart := &TextPart{
		Kind: TextPartKind,
		Text: text,
	}

	// Create A2AMessage
	message := &Message{
		Role:      RoleAgent,
		Parts:     []Part{textPart},
		MessageID: uuid.NewString(),
		Kind:      MessageEventKind,
		TaskID:    taskID,
		ContextID: contextID,
	}

	return message
}

// NewAgentPartsMessage creates a new agent message containing a list of MessageParts.
//
// This function mirrors the Python new_agent_parts_message function behavior.
//
// Args:
//
//	parts: The list of MessagePart objects for the message content.
//	contextID: The context ID for the message (optional).
//	taskID: The task ID for the message (optional).
//
// Returns:
//
//	A new A2AMessage object with role 'agent' and an error if the operation fails.
func NewAgentPartsMessage(parts []Part, contextID, taskID string) *Message {
	return &Message{
		Role:      RoleAgent,
		Parts:     parts,
		MessageID: uuid.NewString(),
		Kind:      MessageEventKind,
		TaskID:    taskID,
		ContextID: contextID,
	}
}

// GetTextParts extracts text content from all TextPart objects in a list of MessageParts.
//
// This function mirrors the Python get_text_parts function behavior.
//
// Args:
//
//	parts: A list of [Part] objects.
//
// Returns:
//
//	A list of strings containing the text content from any [TextPart] objects found.
func GetTextParts(parts []Part) []string {
	textParts := make([]string, 0, len(parts))

	for _, part := range parts {
		if part == nil {
			continue
		}
		// Check if the root is a TextPart
		if textPart, ok := part.(*TextPart); ok {
			textParts = append(textParts, textPart.Text)
		}
	}

	return textParts
}

// GetFileParts extracts file data from all [FilePart] objects in a list of Parts.
//
// This function mirrors the Python get_data_parts function behavior.
//
// Args:
//
//	parts: A list of [Part] objects.
//
// Returns:
//
//	A list of [FileWithBytes] or [FileWithURI] objects containing the file data from any [FilePart] objects found.
func GetFileParts(parts []Part) []File {
	fileContentParts := make([]File, 0, len(parts))

	for _, part := range parts {
		if part == nil {
			continue
		}
		// Check if the part is a FilePart
		if filePart, ok := part.(*FilePart); ok && filePart.File != nil {
			fileContentParts = append(fileContentParts, filePart.File)
		}
	}

	return fileContentParts
}

// GetMessageText extracts and joins all text content from a [Message] parts.
//
// This function mirrors the Python get_message_text function behavior.
//
// Args:
//
//	message: The [Message] object.
//	delimiter: The string to use when joining text from multiple [TextPart].
//
// Returns:
//
//	A single string containing all text content, or an empty string if no text parts are found.
func GetMessageText(message *Message, delimiter string) string {
	if message == nil {
		return ""
	}

	textParts := GetTextParts(message.Parts)

	return strings.Join(textParts, delimiter)
}
