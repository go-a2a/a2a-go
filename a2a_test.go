// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a_test

import (
	"slices"
	"testing"
	"time"

	gocmp "github.com/google/go-cmp/cmp"
	gocmpopts "github.com/google/go-cmp/cmp/cmpopts"

	"github.com/go-a2a/a2a"
)

func TestRole(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		role a2a.MessageRole
		want string
	}{
		"user": {
			role: a2a.MessageRoleUser,
			want: "user",
		},
		"agent": {
			role: a2a.MessageRoleAgent,
			want: "agent",
		},
		"empty": {
			role: a2a.MessageRole(""),
			want: "",
		},
		"arbitrary": {
			role: a2a.MessageRole("custom"),
			want: "custom",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := string(tt.role); got != tt.want {
				t.Errorf("Role = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPart_Type(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		part a2a.Part
		want a2a.PartType
	}{
		"text": {
			part: &a2a.TextPart{},
			want: a2a.PartTypeText,
		},
		"file": {
			part: &a2a.FilePart{},
			want: a2a.PartTypeFile,
		},
		"data": {
			part: &a2a.DataPart{},
			want: a2a.PartTypeData,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tt.part.PartType(); got != tt.want {
				t.Errorf("Part.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestFileContent_CheckContent(t *testing.T) {
// 	t.Parallel()
//
// 	tests := map[string]struct {
// 		fc      a2a.File
// 		wantErr bool
// 	}{
// 		"valid_bytes": {
// 			fc: a2a.FileContent{
// 				Name:     "test.txt",
// 				MIMEType: "text/plain",
// 				Bytes:    "ZGF0YQ==",
// 			},
// 			wantErr: false,
// 		},
// 		"valid_uri": {
// 			fc: a2a.FileContent{
// 				Name:     "test.txt",
// 				MIMEType: "text/plain",
// 				URI:      "https://example.com/file.txt",
// 			},
// 			wantErr: false,
// 		},
// 		"missing_both": {
// 			fc: a2a.FileContent{
// 				Name:     "test.txt",
// 				MIMEType: "text/plain",
// 			},
// 			wantErr: true,
// 		},
// 		"both_present": {
// 			fc: a2a.FileContent{
// 				Name:     "test.txt",
// 				MIMEType: "text/plain",
// 				Bytes:    "ZGF0YQ==",
// 				URI:      "https://example.com/file.txt",
// 			},
// 			wantErr: true,
// 		},
// 	}
//
// 	for name, tt := range tests {
// 		t.Run(name, func(t *testing.T) {
// 			t.Parallel()
//
// 			err := tt.fc.CheckContent()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("FileContent.CheckContent() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func TestTaskStatusUpdateEvent_TaskID(t *testing.T) {
	t.Parallel()

	event := &a2a.TaskStatusUpdateEvent{ID: "test-id"}
	if got := event.TaskID(); got != "test-id" {
		t.Errorf("TaskStatusUpdateEvent.TaskID() = %v, want %v", got, "test-id")
	}
}

func TestTaskArtifactUpdateEvent_TaskID(t *testing.T) {
	t.Parallel()

	event := &a2a.TaskArtifactUpdateEvent{ID: "test-id"}
	if got := event.TaskID(); got != "test-id" {
		t.Errorf("TaskArtifactUpdateEvent.TaskID() = %v, want %v", got, "test-id")
	}
}

func TestTaskStatus(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	status := a2a.TaskStatus{
		State:     a2a.TaskStateSubmitted,
		Timestamp: now,
	}

	if status.State != a2a.TaskStateSubmitted {
		t.Errorf("TaskStatus.State = %v, want %v", status.State, a2a.TaskStateSubmitted)
	}

	// Compare time with slight precision loss
	if !status.Timestamp.Equal(now) {
		t.Errorf("TaskStatus.Timestamp = %v, want %v", status.Timestamp, now)
	}
}

func TestTask(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	task := &a2a.Task{
		ID:        "test-id",
		ContextID: "session-id",
		Status: a2a.TaskStatus{
			State:     a2a.TaskStateSubmitted,
			Timestamp: now,
		},
		Artifacts: []a2a.Artifact{
			{
				Name:        "artifact",
				Description: "test artifact",
				Parts: []a2a.Part{
					&a2a.TextPart{Text: "artifact content"},
				},
			},
		},
		History: []a2a.Message{
			{
				Role: "user",
				Parts: []a2a.Part{
					&a2a.TextPart{Text: "hello"},
				},
			},
		},
		Metadata: map[string]any{"key": "value"},
	}

	if task.ID != "test-id" {
		t.Errorf("Task.ID = %v, want %v", task.ID, "test-id")
	}
	if task.ContextID != "session-id" {
		t.Errorf("Task.SessionID = %v, want %v", task.ContextID, "session-id")
	}
	if task.Status.State != a2a.TaskStateSubmitted {
		t.Errorf("Task.Status.State = %v, want %v", task.Status.State, a2a.TaskStateSubmitted)
	}
	if len(task.Artifacts) != 1 {
		t.Errorf("len(Task.Artifacts) = %v, want %v", len(task.Artifacts), 1)
	}
	if len(task.History) != 1 {
		t.Errorf("len(Task.History) = %v, want %v", len(task.History), 1)
	}

	// Check metadata
	expectedMetadata := map[string]any{"key": "value"}
	if diff := gocmp.Diff(expectedMetadata, task.Metadata); diff != "" {
		t.Errorf("Task.Metadata mismatch (-want +got):\n%s", diff)
	}
}

func TestMessage(t *testing.T) {
	t.Parallel()

	message := a2a.Message{
		Role: "user",
		Parts: []a2a.Part{
			&a2a.TextPart{
				Text: "hello",
			},
			&a2a.FilePart{
				File: &a2a.FileWithBytes{
					Name:     "test.txt",
					MIMEType: "text/plain",
					Bytes:    "ZGF0YQ==",
				},
			},
		},
		Metadata: map[string]any{"key": "value"},
	}

	if message.Role != "user" {
		t.Errorf("Message.Role = %v, want %v", message.Role, "user")
	}
	if len(message.Parts) != 2 {
		t.Errorf("len(Message.Parts) = %v, want %v", len(message.Parts), 2)
	}

	// Check text part
	textPart, ok := message.Parts[0].(*a2a.TextPart)
	if !ok {
		t.Fatalf("message.Parts[0] is not *TextPart")
	}
	if textPart.Text != "hello" {
		t.Errorf("TextPart.Text = %v, want %v", textPart.Text, "hello")
	}

	// Check file part
	filePart, ok := message.Parts[1].(*a2a.FilePart)
	if !ok {
		t.Fatalf("message.Parts[1] is not *FilePart")
	}
	if filePart.File.Name != "test.txt" {
		t.Errorf("FilePart.File.Name = %v, want %v", filePart.File.Name, "test.txt")
	}

	// Check metadata
	expectedMetadata := map[string]any{"key": "value"}
	if diff := gocmp.Diff(expectedMetadata, message.Metadata, gocmpopts.EquateEmpty()); diff != "" {
		t.Errorf("Message.Metadata mismatch (-want +got):\n%s", diff)
	}
}

func TestArtifact(t *testing.T) {
	t.Parallel()

	artifact := a2a.Artifact{
		Name:        "artifact",
		Description: "test artifact",
		Parts: []a2a.Part{
			&a2a.TextPart{Text: "artifact content"},
		},
		// Index:     1,
		// Append:    true,
		// LastChunk: true,
		Metadata: map[string]any{"key": "value"},
	}

	if artifact.Name != "artifact" {
		t.Errorf("Artifact.Name = %v, want %v", artifact.Name, "artifact")
	}
	if artifact.Description != "test artifact" {
		t.Errorf("Artifact.Description = %v, want %v", artifact.Description, "test artifact")
	}
	if len(artifact.Parts) != 1 {
		t.Errorf("len(Artifact.Parts) = %v, want %v", len(artifact.Parts), 1)
	}
	// if artifact.Index != 1 {
	// 	t.Errorf("Artifact.Index = %v, want %v", artifact.Index, 1)
	// }
	// if !artifact.Append {
	// 	t.Errorf("Artifact.Append = %v, want %v", artifact.Append, true)
	// }
	// if !artifact.LastChunk {
	// 	t.Errorf("Artifact.LastChunk = %v, want %v", artifact.LastChunk, true)
	// }

	// Check metadata
	expectedMetadata := map[string]any{"key": "value"}
	if diff := gocmp.Diff(expectedMetadata, artifact.Metadata, gocmpopts.EquateEmpty()); diff != "" {
		t.Errorf("Artifact.Metadata mismatch (-want +got):\n%s", diff)
	}
}

func TestAuthenticationInfo(t *testing.T) {
	t.Parallel()

	auth := a2a.PushNotificationAuthenticationInfo{
		Schemes:     []string{"basic", "oauth2"},
		Credentials: "token123",
	}

	if len(auth.Schemes) != 2 {
		t.Errorf("len(AuthenticationInfo.Schemes) = %v, want %v", len(auth.Schemes), 2)
	}
	if auth.Schemes[0] != "basic" {
		t.Errorf("AuthenticationInfo.Schemes[0] = %v, want %v", auth.Schemes[0], "basic")
	}
	if auth.Schemes[1] != "oauth2" {
		t.Errorf("AuthenticationInfo.Schemes[1] = %v, want %v", auth.Schemes[1], "oauth2")
	}
	if auth.Credentials != "token123" {
		t.Errorf("AuthenticationInfo.Credentials = %v, want %v", auth.Credentials, "token123")
	}
}

func TestPushNotificationConfig(t *testing.T) {
	t.Parallel()

	config := a2a.PushNotificationConfig{
		URL:   "https://example.com/push",
		Token: "token123",
		Authentication: &a2a.PushNotificationAuthenticationInfo{
			Schemes:     []string{"basic"},
			Credentials: "token123",
		},
	}

	if config.URL != "https://example.com/push" {
		t.Errorf("PushNotificationConfig.URL = %v, want %v", config.URL, "https://example.com/push")
	}
	if config.Token != "token123" {
		t.Errorf("PushNotificationConfig.Token = %v, want %v", config.Token, "token123")
	}
	if config.Authentication == nil {
		t.Fatal("PushNotificationConfig.Authentication is nil")
	}
	if len(config.Authentication.Schemes) != 1 {
		t.Errorf("len(PushNotificationConfig.Authentication.Schemes) = %v, want %v", len(config.Authentication.Schemes), 1)
	}
}

func TestTaskIDParams(t *testing.T) {
	t.Parallel()

	params := a2a.TaskIDParams{
		ID:       "test-id",
		Metadata: map[string]any{"key": "value"},
	}

	if params.ID != "test-id" {
		t.Errorf("TaskIDParams.ID = %v, want %v", params.ID, "test-id")
	}

	// Check metadata
	wantMetadata := map[string]any{"key": "value"}
	if diff := gocmp.Diff(wantMetadata, params.Metadata, gocmpopts.EquateEmpty()); diff != "" {
		t.Errorf("TaskIDParams.Metadata mismatch (-want +got):\n%s", diff)
	}
}

func TestTaskQueryParams(t *testing.T) {
	t.Parallel()

	params := a2a.TaskQueryParams{
		ID:            "test-id",
		HistoryLength: 10,
		Metadata:      map[string]any{"key": "value"},
	}

	if params.ID != "test-id" {
		t.Errorf("TaskQueryParams.ID = %v, want %v", params.ID, "test-id")
	}
	if params.HistoryLength != 10 {
		t.Errorf("TaskQueryParams.HistoryLength = %v, want %v", params.HistoryLength, 10)
	}

	// Check metadata
	wantMetadata := map[string]any{"key": "value"}
	if diff := gocmp.Diff(wantMetadata, params.Metadata, gocmpopts.EquateEmpty()); diff != "" {
		t.Errorf("TaskQueryParams.Metadata mismatch (-want +got):\n%s", diff)
	}
}

func TestAgentCapabilities(t *testing.T) {
	t.Parallel()

	capabilities := a2a.AgentCapabilities{
		Streaming:              true,
		PushNotifications:      true,
		StateTransitionHistory: true,
	}

	if !capabilities.Streaming {
		t.Errorf("AgentCapabilities.Streaming = %v, want %v", capabilities.Streaming, true)
	}
	if !capabilities.PushNotifications {
		t.Errorf("AgentCapabilities.PushNotifications = %v, want %v", capabilities.PushNotifications, true)
	}
	if !capabilities.StateTransitionHistory {
		t.Errorf("AgentCapabilities.StateTransitionHistory = %v, want %v", capabilities.StateTransitionHistory, true)
	}
}

func TestAgentCard(t *testing.T) {
	t.Parallel()

	card := a2a.AgentCard{
		Name:        "Test Agent",
		Description: "A test agent",
		URL:         "https://example.com/agent",
		Provider: &a2a.AgentProvider{
			Organization: "Test Org",
			URL:          "https://example.com",
		},
		Version:          "1.0.0",
		DocumentationURL: "https://example.com/docs",
		Capabilities: a2a.AgentCapabilities{
			Streaming:              true,
			PushNotifications:      true,
			StateTransitionHistory: true,
		},
		SecuritySchemes: &a2a.PushNotificationAuthenticationInfo{
			Schemes:     []string{"basic"},
			Credentials: "token123",
		},
		DefaultInputModes:  []string{"text", "file"},
		DefaultOutputModes: []string{"text", "file"},
		Skills: []a2a.AgentSkill{
			{
				ID:          "skill1",
				Name:        "Test Skill",
				Description: "A test skill",
				Tags:        []string{"test", "example"},
				Examples:    []string{"Example 1", "Example 2"},
				InputModes:  []string{"text"},
				OutputModes: []string{"text"},
			},
		},
	}

	if got, want := card.Name, "Test Agent"; got != want {
		t.Errorf("AgentCard.Name = %v, want %v", got, want)
	}
	if got, want := card.Description, "A test agent"; got != want {
		t.Errorf("AgentCard.Description = %v, want %v", got, want)
	}
	if got, want := card.URL, "https://example.com/agent"; got != want {
		t.Errorf("AgentCard.URL = %v, want %v", got, want)
	}
	if card.Provider == nil {
		t.Fatal("AgentCard.Provider is nil")
	}
	if got, want := card.Provider.Organization, "Test Org"; got != want {
		t.Errorf("AgentCard.Provider.Organization = %v, want %v", got, want)
	}
	if got, want := card.Version, "1.0.0"; got != want {
		t.Errorf("AgentCard.Version = %v, want %v", got, want)
	}
	if got, want := card.DocumentationURL, "https://example.com/docs"; got != want {
		t.Errorf("AgentCard.DocumentationURL = %v, want %v", got, want)
	}
	wantCapabilities := a2a.AgentCapabilities{Streaming: true, PushNotifications: true, StateTransitionHistory: true}
	if diff := gocmp.Diff(wantCapabilities, card.Capabilities); diff != "" {
		t.Errorf("card.Capabilities: (-want +got):\n%s", diff)
	}
	wantAuthentication := &a2a.PushNotificationAuthenticationInfo{
		Schemes:     []string{"basic"},
		Credentials: "token123",
	}
	if diff := gocmp.Diff(wantAuthentication, card.SecuritySchemes); diff != "" {
		t.Errorf("card.Authentication: (-want +got):\n%s", diff)
	}
	wantDefaultInputModes := []string{"text", "file"}
	if !slices.Equal(wantDefaultInputModes, card.DefaultInputModes) {
		t.Errorf("AgentCard.DefaultInputModes = %v, want %v", card.DefaultInputModes, wantDefaultInputModes)
	}
	wantDefaultOutputModes := []string{"text", "file"}
	if !slices.Equal(wantDefaultOutputModes, card.DefaultOutputModes) {
		t.Errorf("AgentCard.DefaultOutputModes = %v, want %v", card.DefaultOutputModes, wantDefaultInputModes)
	}
	if len(card.Skills) != 1 {
		t.Errorf("len(AgentCard.Skills) = %v, want %v", len(card.Skills), 1)
	}
	if got, want := card.Skills[0].ID, "skill1"; got != want {
		t.Errorf("AgentCard.Skills[0].ID = %v, want %v", got, want)
	}
}
