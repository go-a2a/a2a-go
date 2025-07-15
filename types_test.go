// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/google/go-cmp/cmp"
)

func TestAgentSkill_Validate(t *testing.T) {
	tests := map[string]struct {
		skill   AgentSkill
		wantErr bool
	}{
		"valid skill": {
			skill: AgentSkill{
				ID:          "test-skill",
				Name:        "Test Skill",
				Description: "A test skill",
				InputModes:  []string{"text"},
				OutputModes: []string{"text"},
			},
			wantErr: false,
		},
		"missing ID": {
			skill: AgentSkill{
				Name:        "Test Skill",
				Description: "A test skill",
				InputModes:  []string{"text"},
				OutputModes: []string{"text"},
			},
			wantErr: true,
		},
		"missing name": {
			skill: AgentSkill{
				ID:          "test-skill",
				Description: "A test skill",
				InputModes:  []string{"text"},
				OutputModes: []string{"text"},
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.skill.Validate()
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf("AgentSkill.Validate() gotErr = %v, wantErr = %v, err = %v", gotErr, tt.wantErr, err)
			}
		})
	}
}

func TestAPIKeySecurityScheme_Validate(t *testing.T) {
	tests := map[string]struct {
		scheme  APIKeySecurityScheme
		wantErr bool
	}{
		"valid scheme": {
			scheme: APIKeySecurityScheme{
				Name:        "api-key",
				Type:        "apiKey",
				In:          LocationHeader,
				Description: "API key in header",
			},
			wantErr: false,
		},
		"invalid location": {
			scheme: APIKeySecurityScheme{
				Name: "api-key",
				Type: "apiKey",
				In:   "invalid",
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.scheme.Validate()
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf("APIKeySecurityScheme.Validate() gotErr = %v, wantErr = %v, err = %v", gotErr, tt.wantErr, err)
			}
		})
	}
}

func TestTaskNotFoundError(t *testing.T) {
	err := TaskNotFoundError{TaskID: "test-task"}

	got := err.Error()
	want := "task not found: test-task"
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TaskNotFoundError.Error() mismatch (-want +got):\n%s", diff)
	}

	gotCode := err.Code()
	wantCode := ErrorCodeTaskNotFound
	if diff := cmp.Diff(wantCode, gotCode); diff != "" {
		t.Errorf("TaskNotFoundError.Code() mismatch (-want +got):\n%s", diff)
	}
}

func TestRequestUnion_JSON(t *testing.T) {
	// Test marshaling
	req := SendMessageRequest{
		Method: "send_message",
		Params: SendMessageParams{
			Message: "Hello, World!",
			TaskID:  "test-task",
		},
		ID: "1",
	}

	union := WireRequest{
		SendMessage: &req,
	}

	data, err := json.Marshal(union)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Test unmarshaling
	var unmarshaled WireRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.SendMessage == nil {
		t.Fatal("Expected SendMessage to be non-nil after unmarshaling")
	}

	got := unmarshaled.SendMessage.Method
	want := "send_message"
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("SendMessage.Method mismatch (-want +got):\n%s", diff)
	}
}

func TestFileWithURI(t *testing.T) {
	file := FileWithURI{
		FileBase: FileBase{
			Name:        "test.txt",
			ContentType: "text/plain",
		},
		URI: "https://example.com/test.txt",
	}

	got := file.GetURI()
	want := "https://example.com/test.txt"
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("FileWithURI.GetURI() mismatch (-want +got):\n%s", diff)
	}

	if gotBytes := file.GetBytes(); gotBytes != nil {
		t.Errorf("FileWithURI.GetBytes() = %v, want nil", gotBytes)
	}
}

func TestFileWithBytes(t *testing.T) {
	data := []byte("test content")
	file := FileWithBytes{
		FileBase: FileBase{
			Name:        "test.txt",
			ContentType: "text/plain",
		},
		Bytes: data,
	}

	got := file.GetURI()
	want := ""
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("FileWithBytes.GetURI() mismatch (-want +got):\n%s", diff)
	}

	gotBytes := string(file.GetBytes())
	wantBytes := "test content"
	if diff := cmp.Diff(wantBytes, gotBytes); diff != "" {
		t.Errorf("FileWithBytes.GetBytes() mismatch (-want +got):\n%s", diff)
	}
}
