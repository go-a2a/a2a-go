// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package a2a provides artifact utilities for the Agent-to-Agent (A2A) protocol.
// This file contains artifact creation utilities that mirror the Python A2A artifact utils,
// converted to idiomatic Go code with proper JSON serialization support.
package a2a

import (
	"crypto/rand"
	"fmt"

	"github.com/go-json-experiment/json"
)

// Part represents a part of an artifact's content.
// It can be a text part, data part, or file part.
type Part interface {
	GetKind() string
	GetMetadata() map[string]any
	Validate() error
}

// TextPart represents a plain text segment within an artifact.
type TextPart struct {
	Kind     string         `json:"kind"`
	Text     string         `json:"text"`
	Metadata map[string]any `json:"metadata,omitzero"`
}

// GetKind returns the part kind.
func (tp TextPart) GetKind() string {
	return tp.Kind
}

// GetMetadata returns the part metadata.
func (tp TextPart) GetMetadata() map[string]any {
	return tp.Metadata
}

// Validate ensures the TextPart is valid.
func (tp TextPart) Validate() error {
	if tp.Kind != "text" {
		return fmt.Errorf("text part kind must be 'text', got '%s'", tp.Kind)
	}
	if tp.Text == "" {
		return fmt.Errorf("text part text cannot be empty")
	}
	return nil
}

// DataPart represents a structured data segment within an artifact.
type DataPart struct {
	Kind     string         `json:"kind"`
	Data     map[string]any `json:"data"`
	Metadata map[string]any `json:"metadata,omitzero"`
}

// GetKind returns the part kind.
func (dp DataPart) GetKind() string {
	return dp.Kind
}

// GetMetadata returns the part metadata.
func (dp DataPart) GetMetadata() map[string]any {
	return dp.Metadata
}

// Validate ensures the DataPart is valid.
func (dp DataPart) Validate() error {
	if dp.Kind != "data" {
		return fmt.Errorf("data part kind must be 'data', got '%s'", dp.Kind)
	}
	if dp.Data == nil {
		return fmt.Errorf("data part data cannot be nil")
	}
	return nil
}

// ArtifactFilePart represents a file segment within an artifact.
type ArtifactFilePart struct {
	Kind     string         `json:"kind"`
	File     File           `json:"file"`
	Metadata map[string]any `json:"metadata,omitzero"`
}

// GetKind returns the part kind.
func (afp ArtifactFilePart) GetKind() string {
	return afp.Kind
}

// GetMetadata returns the part metadata.
func (afp ArtifactFilePart) GetMetadata() map[string]any {
	return afp.Metadata
}

// Validate ensures the ArtifactFilePart is valid.
func (afp ArtifactFilePart) Validate() error {
	if afp.Kind != "file" {
		return fmt.Errorf("file part kind must be 'file', got '%s'", afp.Kind)
	}
	if afp.File == nil {
		return fmt.Errorf("file part file cannot be nil")
	}
	return nil
}

// PartWrapper wraps a Part interface to enable JSON serialization.
type PartWrapper struct {
	part Part
}

// NewPartWrapper creates a new PartWrapper.
func NewPartWrapper(part Part) *PartWrapper {
	return &PartWrapper{part: part}
}

// GetPart returns the wrapped part.
func (pw *PartWrapper) GetPart() Part {
	return pw.part
}

// MarshalJSON implements custom JSON marshaling for PartWrapper.
func (pw PartWrapper) MarshalJSON() ([]byte, error) {
	if pw.part == nil {
		return nil, fmt.Errorf("cannot marshal nil part")
	}
	return json.Marshal(pw.part)
}

// UnmarshalJSON implements custom JSON unmarshaling for PartWrapper.
func (pw *PartWrapper) UnmarshalJSON(data []byte) error {
	var kind struct {
		Kind string `json:"kind"`
	}

	if err := json.Unmarshal(data, &kind); err != nil {
		return fmt.Errorf("failed to unmarshal part kind: %w", err)
	}

	switch kind.Kind {
	case "text":
		var tp TextPart
		if err := json.Unmarshal(data, &tp); err != nil {
			return fmt.Errorf("failed to unmarshal text part: %w", err)
		}
		pw.part = &tp
	case "data":
		var dp DataPart
		if err := json.Unmarshal(data, &dp); err != nil {
			return fmt.Errorf("failed to unmarshal data part: %w", err)
		}
		pw.part = &dp
	case "file":
		var afp ArtifactFilePart
		if err := json.Unmarshal(data, &afp); err != nil {
			return fmt.Errorf("failed to unmarshal file part: %w", err)
		}
		pw.part = &afp
	default:
		return fmt.Errorf("unknown part kind: %s", kind.Kind)
	}

	return nil
}

// Validate validates the wrapped part.
func (pw *PartWrapper) Validate() error {
	if pw.part == nil {
		return fmt.Errorf("part wrapper cannot contain nil part")
	}
	return pw.part.Validate()
}

// Artifact represents a generated output from a task, which can contain multiple parts.
type Artifact struct {
	ArtifactID  string         `json:"artifactId"`
	Name        string         `json:"name,omitzero"`
	Description string         `json:"description,omitzero"`
	Parts       []*PartWrapper `json:"parts"`
	Extensions  []string       `json:"extensions,omitzero"`
	Metadata    map[string]any `json:"metadata,omitzero"`
}

// Validate ensures the Artifact is valid.
func (a Artifact) Validate() error {
	if a.ArtifactID == "" {
		return fmt.Errorf("artifact ID cannot be empty")
	}
	if len(a.Parts) == 0 {
		return fmt.Errorf("artifact must contain at least one part")
	}
	for i, part := range a.Parts {
		if part == nil {
			return fmt.Errorf("artifact part at index %d cannot be nil", i)
		}
		if err := part.Validate(); err != nil {
			return fmt.Errorf("artifact part at index %d is invalid: %w", i, err)
		}
	}
	return nil
}

// generateUUID generates a secure UUID v4.
func generateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Set version 4 (random UUID)
	b[6] = (b[6] & 0x0f) | 0x40
	// Set variant 10
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

// NewArtifact creates a new Artifact object from a list of parts, a name, and an optional description.
// It generates a unique artifactId using a secure UUID v4.
func NewArtifact(parts []Part, name string, description string) (*Artifact, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("artifact must contain at least one part")
	}

	artifactID, err := generateUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate artifact ID: %w", err)
	}

	// Wrap all parts in PartWrapper for JSON serialization
	wrappedParts := make([]*PartWrapper, len(parts))
	for i, part := range parts {
		if part == nil {
			return nil, fmt.Errorf("part at index %d cannot be nil", i)
		}
		if err := part.Validate(); err != nil {
			return nil, fmt.Errorf("part at index %d is invalid: %w", i, err)
		}
		wrappedParts[i] = NewPartWrapper(part)
	}

	artifact := &Artifact{
		ArtifactID:  artifactID,
		Name:        name,
		Description: description,
		Parts:       wrappedParts,
	}

	return artifact, nil
}

// NewTextArtifact creates a new Artifact object containing only a single TextPart.
func NewTextArtifact(name, text, description string) (*Artifact, error) {
	if text == "" {
		return nil, fmt.Errorf("text content cannot be empty")
	}

	textPart := &TextPart{
		Kind: "text",
		Text: text,
	}

	return NewArtifact([]Part{textPart}, name, description)
}

// NewDataArtifact creates a new Artifact object containing only a single DataPart.
func NewDataArtifact(name string, data map[string]any, description string) (*Artifact, error) {
	if data == nil {
		return nil, fmt.Errorf("data content cannot be nil")
	}

	dataPart := &DataPart{
		Kind: "data",
		Data: data,
	}

	return NewArtifact([]Part{dataPart}, name, description)
}

// NewFileArtifact creates a new Artifact object containing only a single ArtifactFilePart.
func NewFileArtifact(name string, file File, description string) (*Artifact, error) {
	if file == nil {
		return nil, fmt.Errorf("file content cannot be nil")
	}

	filePart := &ArtifactFilePart{
		Kind: "file",
		File: file,
	}

	return NewArtifact([]Part{filePart}, name, description)
}
