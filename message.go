// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package a2a provides message utilities for the Agent-to-Agent (A2A) protocol.
// This file contains utility functions for creating and handling A2A Message objects,
// converted from Python to idiomatic Go code with proper error handling and validation.
package a2a

import (
	"fmt"
	"strings"
)

// Role represents the role of a message sender in the A2A protocol.
type Role string

// Role constants for message senders.
const (
	RoleAgent Role = "agent"
	RoleUser  Role = "user"
)

// MessagePart represents a part of a message's content.
// This matches the Python Part structure with a root field.
type MessagePart struct {
	Root Part `json:"root"`
}

// Validate ensures the MessagePart is valid.
func (mp MessagePart) Validate() error {
	if mp.Root == nil {
		return fmt.Errorf("message part root cannot be nil")
	}
	return mp.Root.Validate()
}

// A2AMessage represents a message in the A2A protocol.
// This matches the Python Message structure with role, parts, messageId, taskId, and contextId.
type A2AMessage struct {
	Role      Role           `json:"role"`
	Parts     []*MessagePart `json:"parts"`
	MessageID string         `json:"messageId"`
	TaskID    *string        `json:"taskId,omitzero"`
	ContextID *string        `json:"contextId,omitzero"`
}

// Validate ensures the A2AMessage is valid.
func (m A2AMessage) Validate() error {
	if m.Role != RoleAgent && m.Role != RoleUser {
		return fmt.Errorf("invalid message role: %s", m.Role)
	}
	if m.MessageID == "" {
		return fmt.Errorf("message ID cannot be empty")
	}
	if len(m.Parts) == 0 {
		return fmt.Errorf("message must contain at least one part")
	}
	for i, part := range m.Parts {
		if part == nil {
			return fmt.Errorf("message part at index %d cannot be nil", i)
		}
		if err := part.Validate(); err != nil {
			return fmt.Errorf("message part at index %d is invalid: %w", i, err)
		}
	}
	return nil
}

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
func NewAgentTextMessage(text string, contextID *string, taskID *string) (*A2AMessage, error) {
	if text == "" {
		return nil, fmt.Errorf("text content cannot be empty")
	}

	// Generate unique message ID
	messageID, err := generateUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate message ID: %w", err)
	}

	// Create TextPart with "text" kind
	textPart := &TextPart{
		Kind: "text",
		Text: text,
	}

	// Create MessagePart with TextPart as root
	messagePart := &MessagePart{
		Root: textPart,
	}

	// Create A2AMessage
	message := &A2AMessage{
		Role:      RoleAgent,
		Parts:     []*MessagePart{messagePart},
		MessageID: messageID,
		TaskID:    taskID,
		ContextID: contextID,
	}

	return message, nil
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
func NewAgentPartsMessage(parts []*MessagePart, contextID *string, taskID *string) (*A2AMessage, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("parts list cannot be empty")
	}

	// Validate all parts
	for i, part := range parts {
		if part == nil {
			return nil, fmt.Errorf("part at index %d cannot be nil", i)
		}
		if err := part.Validate(); err != nil {
			return nil, fmt.Errorf("part at index %d is invalid: %w", i, err)
		}
	}

	// Generate unique message ID
	messageID, err := generateUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate message ID: %w", err)
	}

	// Create A2AMessage
	message := &A2AMessage{
		Role:      RoleAgent,
		Parts:     parts,
		MessageID: messageID,
		TaskID:    taskID,
		ContextID: contextID,
	}

	return message, nil
}

// GetTextParts extracts text content from all TextPart objects in a list of MessageParts.
//
// This function mirrors the Python get_text_parts function behavior.
//
// Args:
//
//	parts: A list of MessagePart objects.
//
// Returns:
//
//	A list of strings containing the text content from any TextPart objects found.
func GetTextParts(parts []*MessagePart) []string {
	var textParts []string

	for _, part := range parts {
		if part == nil || part.Root == nil {
			continue
		}

		// Check if the root is a TextPart
		if textPart, ok := part.Root.(*TextPart); ok {
			textParts = append(textParts, textPart.Text)
		}
	}

	return textParts
}

// GetMessageText extracts and joins all text content from an A2AMessage's parts.
//
// This function mirrors the Python get_message_text function behavior.
//
// Args:
//
//	message: The A2AMessage object.
//	delimiter: The string to use when joining text from multiple TextParts.
//
// Returns:
//
//	A single string containing all text content, or an empty string if no text parts are found.
func GetMessageText(message *A2AMessage, delimiter string) string {
	if message == nil {
		return ""
	}

	textParts := GetTextParts(message.Parts)
	return strings.Join(textParts, delimiter)
}

// GetMessageTextWithNewlines extracts and joins all text content from an A2AMessage's parts using newline delimiter.
//
// This is a convenience function that uses "\n" as the default delimiter.
//
// Args:
//
//	message: The A2AMessage object.
//
// Returns:
//
//	A single string containing all text content joined with newlines.
func GetMessageTextWithNewlines(message *A2AMessage) string {
	return GetMessageText(message, "\n")
}
