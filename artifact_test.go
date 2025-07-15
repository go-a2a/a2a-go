// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/google/go-cmp/cmp"
)

func TestTextPart(t *testing.T) {
	tests := map[string]struct {
		part    TextPart
		wantErr bool
	}{
		"valid text part": {
			part: TextPart{
				Kind: "text",
				Text: "Hello, World!",
			},
			wantErr: false,
		},
		"valid text part with metadata": {
			part: TextPart{
				Kind: "text",
				Text: "Hello, World!",
				Metadata: map[string]any{
					"author": "test",
				},
			},
			wantErr: false,
		},
		"invalid kind": {
			part: TextPart{
				Kind: "invalid",
				Text: "Hello, World!",
			},
			wantErr: true,
		},
		"empty text": {
			part: TextPart{
				Kind: "text",
				Text: "",
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.part.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TextPart.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if got := tt.part.GetKind(); got != "text" {
					t.Errorf("TextPart.GetKind() = %v, want text", got)
				}
				if got := tt.part.GetMetadata(); !cmp.Equal(got, tt.part.Metadata) {
					t.Errorf("TextPart.GetMetadata() = %v, want %v", got, tt.part.Metadata)
				}
			}
		})
	}
}

func TestDataPart(t *testing.T) {
	tests := map[string]struct {
		part    DataPart
		wantErr bool
	}{
		"valid data part": {
			part: DataPart{
				Kind: "data",
				Data: map[string]any{
					"key": "value",
				},
			},
			wantErr: false,
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
			wantErr: false,
		},
		"invalid kind": {
			part: DataPart{
				Kind: "invalid",
				Data: map[string]any{
					"key": "value",
				},
			},
			wantErr: true,
		},
		"nil data": {
			part: DataPart{
				Kind: "data",
				Data: nil,
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.part.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DataPart.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if got := tt.part.GetKind(); got != "data" {
					t.Errorf("DataPart.GetKind() = %v, want data", got)
				}
				if got := tt.part.GetMetadata(); !cmp.Equal(got, tt.part.Metadata) {
					t.Errorf("DataPart.GetMetadata() = %v, want %v", got, tt.part.Metadata)
				}
			}
		})
	}
}

func TestArtifactFilePart(t *testing.T) {
	file := &FileWithURI{
		FileBase: FileBase{
			Name:        "test.txt",
			ContentType: "text/plain",
		},
		URI: "file://test.txt",
	}

	tests := map[string]struct {
		part    ArtifactFilePart
		wantErr bool
	}{
		"valid file part": {
			part: ArtifactFilePart{
				Kind: "file",
				File: file,
			},
			wantErr: false,
		},
		"valid file part with metadata": {
			part: ArtifactFilePart{
				Kind: "file",
				File: file,
				Metadata: map[string]any{
					"author": "test",
				},
			},
			wantErr: false,
		},
		"invalid kind": {
			part: ArtifactFilePart{
				Kind: "invalid",
				File: file,
			},
			wantErr: true,
		},
		"nil file": {
			part: ArtifactFilePart{
				Kind: "file",
				File: nil,
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.part.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ArtifactFilePart.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if got := tt.part.GetKind(); got != "file" {
					t.Errorf("ArtifactFilePart.GetKind() = %v, want file", got)
				}
				if got := tt.part.GetMetadata(); !cmp.Equal(got, tt.part.Metadata) {
					t.Errorf("ArtifactFilePart.GetMetadata() = %v, want %v", got, tt.part.Metadata)
				}
			}
		})
	}
}

func TestPartWrapper(t *testing.T) {
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
		part    Part
		wantErr bool
	}{
		"valid text part": {
			part:    textPart,
			wantErr: false,
		},
		"valid data part": {
			part:    dataPart,
			wantErr: false,
		},
		"nil part": {
			part:    nil,
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			wrapper := NewPartWrapper(tt.part)
			err := wrapper.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PartWrapper.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if got := wrapper.GetPart(); got != tt.part {
					t.Errorf("PartWrapper.GetPart() = %v, want %v", got, tt.part)
				}
			}
		})
	}
}

func TestPartWrapperJSON(t *testing.T) {
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
			wrapper := NewPartWrapper(tt.part)

			// Test marshaling
			data, err := json.Marshal(wrapper)
			if err != nil {
				t.Errorf("json.Marshal() error = %v", err)
				return
			}

			// Test unmarshaling
			var newWrapper PartWrapper
			if err := json.Unmarshal(data, &newWrapper); err != nil {
				t.Errorf("json.Unmarshal() error = %v", err)
				return
			}

			// Compare the parts
			if newWrapper.GetPart().GetKind() != tt.part.GetKind() {
				t.Errorf("Unmarshaled part kind = %v, want %v", newWrapper.GetPart().GetKind(), tt.part.GetKind())
			}
		})
	}
}

func TestNewArtifact(t *testing.T) {
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

	invalidPart := &TextPart{
		Kind: "invalid",
		Text: "Hello, World!",
	}

	tests := map[string]struct {
		parts        []Part
		artifactName string
		description  string
		wantErr      bool
	}{
		"valid artifact with text part": {
			parts:        []Part{textPart},
			artifactName: "test artifact",
			description:  "test description",
			wantErr:      false,
		},
		"valid artifact with multiple parts": {
			parts:        []Part{textPart, dataPart},
			artifactName: "test artifact",
			description:  "test description",
			wantErr:      false,
		},
		"valid artifact with empty description": {
			parts:        []Part{textPart},
			artifactName: "test artifact",
			description:  "",
			wantErr:      false,
		},
		"empty parts": {
			parts:        []Part{},
			artifactName: "test artifact",
			description:  "test description",
			wantErr:      true,
		},
		"nil parts": {
			parts:        nil,
			artifactName: "test artifact",
			description:  "test description",
			wantErr:      true,
		},
		"invalid part": {
			parts:        []Part{invalidPart},
			artifactName: "test artifact",
			description:  "test description",
			wantErr:      true,
		},
		"nil part in slice": {
			parts:        []Part{textPart, nil},
			artifactName: "test artifact",
			description:  "test description",
			wantErr:      true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			artifact, err := NewArtifact(tt.parts, tt.artifactName, tt.description)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewArtifact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
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

				// Validate the artifact
				if err := artifact.Validate(); err != nil {
					t.Errorf("NewArtifact() created invalid artifact: %v", err)
				}
			}
		})
	}
}

func TestNewTextArtifact(t *testing.T) {
	tests := map[string]struct {
		artifactName string
		text         string
		description  string
		wantErr      bool
	}{
		"valid text artifact": {
			artifactName: "test text artifact",
			text:         "Hello, World!",
			description:  "test description",
			wantErr:      false,
		},
		"valid text artifact with empty description": {
			artifactName: "test text artifact",
			text:         "Hello, World!",
			description:  "",
			wantErr:      false,
		},
		"empty text": {
			artifactName: "test text artifact",
			text:         "",
			description:  "test description",
			wantErr:      true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			artifact, err := NewTextArtifact(tt.artifactName, tt.text, tt.description)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTextArtifact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
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

				part := artifact.Parts[0].GetPart()
				if part.GetKind() != "text" {
					t.Errorf("NewTextArtifact() Part kind = %v, want text", part.GetKind())
				}

				if textPart, ok := part.(*TextPart); ok {
					if textPart.Text != tt.text {
						t.Errorf("NewTextArtifact() Part text = %v, want %v", textPart.Text, tt.text)
					}
				} else {
					t.Error("NewTextArtifact() Part is not a TextPart")
				}
			}
		})
	}
}

func TestNewDataArtifact(t *testing.T) {
	data := map[string]any{
		"key":    "value",
		"number": 42,
	}

	tests := map[string]struct {
		artifactName string
		data         map[string]any
		description  string
		wantErr      bool
	}{
		"valid data artifact": {
			artifactName: "test data artifact",
			data:         data,
			description:  "test description",
			wantErr:      false,
		},
		"valid data artifact with empty description": {
			artifactName: "test data artifact",
			data:         data,
			description:  "",
			wantErr:      false,
		},
		"nil data": {
			artifactName: "test data artifact",
			data:         nil,
			description:  "test description",
			wantErr:      true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			artifact, err := NewDataArtifact(tt.artifactName, tt.data, tt.description)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDataArtifact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
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

				part := artifact.Parts[0].GetPart()
				if part.GetKind() != "data" {
					t.Errorf("NewDataArtifact() Part kind = %v, want data", part.GetKind())
				}

				if dataPart, ok := part.(*DataPart); ok {
					if !cmp.Equal(dataPart.Data, tt.data) {
						t.Errorf("NewDataArtifact() Part data = %v, want %v", dataPart.Data, tt.data)
					}
				} else {
					t.Error("NewDataArtifact() Part is not a DataPart")
				}
			}
		})
	}
}

func TestNewFileArtifact(t *testing.T) {
	file := &FileWithURI{
		FileBase: FileBase{
			Name:        "test.txt",
			ContentType: "text/plain",
		},
		URI: "file://test.txt",
	}

	tests := map[string]struct {
		artifactName string
		file         File
		description  string
		wantErr      bool
	}{
		"valid file artifact": {
			artifactName: "test file artifact",
			file:         file,
			description:  "test description",
			wantErr:      false,
		},
		"valid file artifact with empty description": {
			artifactName: "test file artifact",
			file:         file,
			description:  "",
			wantErr:      false,
		},
		"nil file": {
			artifactName: "test file artifact",
			file:         nil,
			description:  "test description",
			wantErr:      true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			artifact, err := NewFileArtifact(tt.artifactName, tt.file, tt.description)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFileArtifact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if artifact.ArtifactID == "" {
					t.Error("NewFileArtifact() ArtifactID is empty")
				}
				if artifact.Name != tt.artifactName {
					t.Errorf("NewFileArtifact() Name = %v, want %v", artifact.Name, tt.artifactName)
				}
				if artifact.Description != tt.description {
					t.Errorf("NewFileArtifact() Description = %v, want %v", artifact.Description, tt.description)
				}
				if len(artifact.Parts) != 1 {
					t.Errorf("NewFileArtifact() Parts length = %v, want 1", len(artifact.Parts))
				}

				part := artifact.Parts[0].GetPart()
				if part.GetKind() != "file" {
					t.Errorf("NewFileArtifact() Part kind = %v, want file", part.GetKind())
				}

				if filePart, ok := part.(*ArtifactFilePart); ok {
					if filePart.File != tt.file {
						t.Errorf("NewFileArtifact() Part file = %v, want %v", filePart.File, tt.file)
					}
				} else {
					t.Error("NewFileArtifact() Part is not an ArtifactFilePart")
				}
			}
		})
	}
}

func TestArtifactJSON(t *testing.T) {
	textPart := &TextPart{
		Kind: "text",
		Text: "Hello, World!",
	}

	artifact, err := NewArtifact([]Part{textPart}, "test artifact", "test description")
	if err != nil {
		t.Fatalf("NewArtifact() error = %v", err)
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

func TestGenerateUUID(t *testing.T) {
	uuid1, err := generateUUID()
	if err != nil {
		t.Errorf("generateUUID() error = %v", err)
	}

	uuid2, err := generateUUID()
	if err != nil {
		t.Errorf("generateUUID() error = %v", err)
	}

	if uuid1 == uuid2 {
		t.Error("generateUUID() generated identical UUIDs")
	}

	// Basic format check (36 characters with dashes in right places)
	if len(uuid1) != 36 {
		t.Errorf("generateUUID() UUID length = %v, want 36", len(uuid1))
	}
	if uuid1[8] != '-' || uuid1[13] != '-' || uuid1[18] != '-' || uuid1[23] != '-' {
		t.Errorf("generateUUID() UUID format incorrect: %v", uuid1)
	}
}
