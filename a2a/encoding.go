// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"fmt"

	json "github.com/bytedance/sonic"
)

// MarshalJSON implements [json.Marshaler] for the [Part] interface.
func MarshalPart(part Part) ([]byte, error) {
	switch p := part.(type) {
	case TextPart:
		return json.ConfigFastest.Marshal(p)
	case FilePart:
		return json.ConfigFastest.Marshal(p)
	case DataPart:
		return json.ConfigFastest.Marshal(p)
	default:
		return nil, fmt.Errorf("unknown part type: %T", part)
	}
}

// UnmarshalPart unmarshals a JSON part into the appropriate [Part] type.
func UnmarshalPart(data []byte) (Part, error) {
	// First, unmarshal into a map to determine the part type
	var partMap map[string]any
	if err := json.Unmarshal(data, &partMap); err != nil {
		return nil, err
	}

	// Get the part type
	partType, ok := partMap["type"].(string)
	if !ok {
		return nil, fmt.Errorf("part type not found or not a string")
	}

	// Unmarshal into the appropriate type based on the part type
	switch partType {
	case "text":
		var textPart TextPart
		if err := json.ConfigFastest.Unmarshal(data, &textPart); err != nil {
			return nil, err
		}
		return textPart, nil

	case "file":
		var filePart FilePart
		if err := json.ConfigFastest.Unmarshal(data, &filePart); err != nil {
			return nil, err
		}
		return filePart, nil

	case "data":
		var dataPart DataPart
		if err := json.ConfigFastest.Unmarshal(data, &dataPart); err != nil {
			return nil, err
		}
		return dataPart, nil

	default:
		return nil, fmt.Errorf("unknown part type: %s", partType)
	}
}

// UnmarshalStreamingResponse unmarshals a JSON streaming response into the appropriate type.
func UnmarshalStreamingResponse(data []byte) (*SendTaskStreamingResponse, error) {
	// First unmarshal into a generic response to get the result structure
	var resp SendTaskStreamingResponse
	if err := json.ConfigFastest.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	// If there's an error, no need to check the result
	if resp.Error != nil {
		return &resp, nil
	}

	// Re-marshal the result to determine the actual type
	resultBytes, err := json.ConfigFastest.Marshal(resp.Result)
	if err != nil {
		return nil, err
	}

	// Try to unmarshal as TaskStatusUpdateEvent
	var statusEvent TaskStatusUpdateEvent
	if err := json.ConfigFastest.Unmarshal(resultBytes, &statusEvent); err == nil {
		// Check if it has the required fields for a status event
		if statusEvent.ID != "" && statusEvent.Status.State != "" {
			resp.Result = statusEvent
			return &resp, nil
		}
	}

	// Try to unmarshal as TaskArtifactUpdateEvent
	var artifactEvent TaskArtifactUpdateEvent
	if err := json.ConfigFastest.Unmarshal(resultBytes, &artifactEvent); err == nil {
		// Check if it has the required fields for an artifact event
		if artifactEvent.ID != "" && len(artifactEvent.Artifact.Parts) > 0 {
			resp.Result = artifactEvent
			return &resp, nil
		}
	}

	// Could not determine the event type
	return nil, fmt.Errorf("unknown streaming event type")
}
