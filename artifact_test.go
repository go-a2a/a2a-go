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
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/google/go-cmp/cmp"
)

func TestTextPart(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		part TextPart
	}{
		"valid text part": {
			part: TextPart{
				Kind: "text",
				Text: "Hello, World!",
			},
		},
		"valid text part with metadata": {
			part: TextPart{
				Kind: "text",
				Text: "Hello, World!",
				Metadata: map[string]any{
					"author": "test",
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tt.part.GetKind(); got != TextPartKind {
				t.Errorf("TextPart.GetKind() = %v, want %v", got, TextPartKind)
			}
		})
	}
}

func TestDataPart(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		part DataPart
	}{
		"valid data part": {
			part: DataPart{
				Kind: "data",
				Data: map[string]any{
					"key": "value",
				},
			},
		},
		"valid data part with metadata": {
			part: DataPart{
				Kind: "data",
				Data: map[string]any{
					"key": "value",
				},
				Metadata: map[string]any{
					"author": "test",
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tt.part.GetKind(); got != DataPartKind {
				t.Errorf("DataPart.GetKind() = %v, want %v", got, DataPartKind)
			}
		})
	}
}

func TestFilePart(t *testing.T) {
	t.Parallel()

	file := &FileWithURI{
		Name:     "test.txt",
		MIMEType: "text/plain",
		URI:      "file://test.txt",
	}

	tests := map[string]struct {
		part FilePart
	}{
		"valid file part": {
			part: FilePart{
				Kind: "file",
				File: file,
			},
		},
		"valid file part with metadata": {
			part: FilePart{
				Kind: "file",
				File: file,
				Metadata: map[string]any{
					"author": "test",
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tt.part.GetKind(); got != FilePartKind {
				t.Errorf("FilePart.GetKind() = %v, want %v", got, FilePartKind)
			}
		})
	}
}

func TestPartJSON(t *testing.T) {
	t.Parallel()

	textPart := &TextPart{
		Kind: "text",
		Text: "Hello, World!",
	}

	dataPart := &DataPart{
		Kind: "data",
		Data: map[string]any{
			"key": "value",
		},
	}

	tests := map[string]struct {
		part Part
	}{
		"text part": {
			part: textPart,
		},
		"data part": {
			part: dataPart,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Test marshaling
			data, err := json.Marshal(tt.part)
			if err != nil {
				t.Errorf("json.Marshal() error = %v", err)
				return
			}

			// Test unmarshaling
			newPart, err := UnmarshalPart(data)
			if err != nil {
				t.Errorf("UnmarshalPart() error = %v", err)
				return
			}

			// Compare the parts
			if newPart.GetKind() != tt.part.GetKind() {
				t.Errorf("Unmarshaled part kind = %v, want %v", newPart.GetKind(), tt.part.GetKind())
			}
		})
	}
}

func TestNewArtifact(t *testing.T) {
	t.Parallel()

	textPart := &TextPart{
		Kind: "text",
		Text: "Hello, World!",
	}

	dataPart := &DataPart{
		Kind: "data",
		Data: map[string]any{
			"key": "value",
		},
	}

	tests := map[string]struct {
		parts        []Part
		artifactName string
		description  string
	}{
		"valid artifact with text part": {
			parts:        []Part{textPart},
			artifactName: "test artifact",
			description:  "test description",
		},
		"valid artifact with multiple parts": {
			parts:        []Part{textPart, dataPart},
			artifactName: "test artifact",
			description:  "test description",
		},
		"valid artifact with empty description": {
			parts:        []Part{textPart},
			artifactName: "test artifact",
			description:  "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			artifact := NewArtifact(tt.parts, tt.artifactName, tt.description)
			if artifact == nil {
				t.Fatal("NewArtifact() returned nil")
			}

			if artifact.ArtifactID == "" {
				t.Error("NewArtifact() ArtifactID is empty")
			}
			if artifact.Name != tt.artifactName {
				t.Errorf("NewArtifact() Name = %v, want %v", artifact.Name, tt.artifactName)
			}
			if artifact.Description != tt.description {
				t.Errorf("NewArtifact() Description = %v, want %v", artifact.Description, tt.description)
			}
			if len(artifact.Parts) != len(tt.parts) {
				t.Errorf("NewArtifact() Parts length = %v, want %v", len(artifact.Parts), len(tt.parts))
			}
		})
	}
}

func TestNewTextArtifact(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		artifactName string
		text         string
		description  string
	}{
		"valid text artifact": {
			artifactName: "test text artifact",
			text:         "Hello, World!",
			description:  "test description",
		},
		"valid text artifact with empty description": {
			artifactName: "test text artifact",
			text:         "Hello, World!",
			description:  "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			artifact := NewTextArtifact(tt.artifactName, tt.text, tt.description)
			if artifact == nil {
				t.Fatal("NewTextArtifact() returned nil")
			}

			if artifact.ArtifactID == "" {
				t.Error("NewTextArtifact() ArtifactID is empty")
			}
			if artifact.Name != tt.artifactName {
				t.Errorf("NewTextArtifact() Name = %v, want %v", artifact.Name, tt.artifactName)
			}
			if artifact.Description != tt.description {
				t.Errorf("NewTextArtifact() Description = %v, want %v", artifact.Description, tt.description)
			}
			if len(artifact.Parts) != 1 {
				t.Errorf("NewTextArtifact() Parts length = %v, want 1", len(artifact.Parts))
			}

			part := artifact.Parts[0]
			if part.GetKind() != TextPartKind {
				t.Errorf("NewTextArtifact() Part kind = %v, want %v", part.GetKind(), TextPartKind)
			}

			if textPart, ok := part.(*TextPart); ok {
				if textPart.Text != tt.text {
					t.Errorf("NewTextArtifact() Part text = %v, want %v", textPart.Text, tt.text)
				}
			} else {
				t.Error("NewTextArtifact() Part is not a TextPart")
			}
		})
	}
}

func TestNewDataArtifact(t *testing.T) {
	t.Parallel()

	data := map[string]any{
		"key":    "value",
		"number": 42,
	}

	tests := map[string]struct {
		artifactName string
		data         map[string]any
		description  string
	}{
		"valid data artifact": {
			artifactName: "test data artifact",
			data:         data,
			description:  "test description",
		},
		"valid data artifact with empty description": {
			artifactName: "test data artifact",
			data:         data,
			description:  "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			artifact := NewDataArtifact(tt.artifactName, tt.data, tt.description)
			if artifact == nil {
				t.Fatal("NewDataArtifact() returned nil")
			}

			if artifact.ArtifactID == "" {
				t.Error("NewDataArtifact() ArtifactID is empty")
			}
			if artifact.Name != tt.artifactName {
				t.Errorf("NewDataArtifact() Name = %v, want %v", artifact.Name, tt.artifactName)
			}
			if artifact.Description != tt.description {
				t.Errorf("NewDataArtifact() Description = %v, want %v", artifact.Description, tt.description)
			}
			if len(artifact.Parts) != 1 {
				t.Errorf("NewDataArtifact() Parts length = %v, want 1", len(artifact.Parts))
			}

			part := artifact.Parts[0]
			if part.GetKind() != DataPartKind {
				t.Errorf("NewDataArtifact() Part kind = %v, want %v", part.GetKind(), DataPartKind)
			}

			if dataPart, ok := part.(*DataPart); ok {
				if !cmp.Equal(dataPart.Data, tt.data) {
					t.Errorf("NewDataArtifact() Part data = %v, want %v", dataPart.Data, tt.data)
				}
			} else {
				t.Error("NewDataArtifact() Part is not a DataPart")
			}
		})
	}
}

func TestArtifactJSON(t *testing.T) {
	t.Parallel()

	textPart := &TextPart{
		Kind: "text",
		Text: "Hello, World!",
	}

	artifact := NewArtifact([]Part{textPart}, "test artifact", "test description")
	if artifact == nil {
		t.Fatal("NewArtifact() returned nil")
	}

	// Test marshaling
	data, err := json.Marshal(artifact)
	if err != nil {
		t.Errorf("json.Marshal() error = %v", err)
		return
	}

	// Test unmarshaling
	var newArtifact Artifact
	if err := json.Unmarshal(data, &newArtifact); err != nil {
		t.Errorf("json.Unmarshal() error = %v", err)
		return
	}

	// Compare the artifacts
	if newArtifact.ArtifactID != artifact.ArtifactID {
		t.Errorf("Unmarshaled artifact ID = %v, want %v", newArtifact.ArtifactID, artifact.ArtifactID)
	}
	if newArtifact.Name != artifact.Name {
		t.Errorf("Unmarshaled artifact name = %v, want %v", newArtifact.Name, artifact.Name)
	}
	if newArtifact.Description != artifact.Description {
		t.Errorf("Unmarshaled artifact description = %v, want %v", newArtifact.Description, artifact.Description)
	}
	if len(newArtifact.Parts) != len(artifact.Parts) {
		t.Errorf("Unmarshaled artifact parts length = %v, want %v", len(newArtifact.Parts), len(artifact.Parts))
	}
}
