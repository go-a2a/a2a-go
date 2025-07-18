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
	"github.com/google/uuid"
)

// NewArtifact creates a new Artifact object from a list of parts, name and description.
func NewArtifact(parts []Part, name, description string) *Artifact {
	return &Artifact{
		ArtifactID:  uuid.NewString(),
		Parts:       parts,
		Name:        name,
		Description: description,
	}
}

// NewTextArtifact creates a new Artifact object containing only a single [TextPart].
func NewTextArtifact(name, text, description string) *Artifact {
	textPart := NewTextPart(text)

	return NewArtifact([]Part{textPart}, name, description)
}

// NewDataArtifact creates a new Artifact object containing only a single DataPart.
func NewDataArtifact(name string, data map[string]any, description string) *Artifact {
	dataPart := NewDataPart(data)

	return NewArtifact([]Part{dataPart}, name, description)
}
