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
	"fmt"
	"testing"

	a2a_v1 "github.com/go-a2a/a2a-grpc/v1"
	gocmp "github.com/google/go-cmp/cmp"
)

func TestMessageToProtoRoundTrip(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		msg *Message
	}{
		"standard message": {
			msg: &Message{
				MessageID: "msg1",
				ContextID: "ctx1",
				TaskID:    "task1",
				Role:      RoleUser,
				Kind:      MessageEventKind,
				Parts: []Part{
					NewTextPart("Hello"),
				},
				Metadata: map[string]any{"key": "value"},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoMsg, err := ToProtoMessage(tt.msg)
			if err != nil {
				t.Fatalf("ToProtoMessage() error = %v", err)
			}

			convertedMsg, err := FromProtoMessage(protoMsg)
			if err != nil {
				t.Fatalf("FromProtoMessage() error = %v", err)
			}

			if diff := gocmp.Diff(tt.msg, convertedMsg); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestNilMessageToProto(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		testType string // "ToProto" or "FromProto"
		wantNil  bool
	}{
		"ToProtoMessage with nil": {
			testType: "ToProto",
			wantNil:  true,
		},
		"FromProtoMessage with nil": {
			testType: "FromProto",
			wantNil:  true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			switch tt.testType {
			case "ToProto":
				protoMsg, err := ToProtoMessage(nil)
				if err != nil {
					t.Fatalf("ToProtoMessage(nil) error = %v", err)
				}
				if protoMsg != nil {
					t.Errorf("ToProtoMessage(nil) = %+v, want nil", protoMsg)
				}
			case "FromProto":
				msg, err := FromProtoMessage(nil)
				if err != nil {
					t.Fatalf("FromProtoMessage(nil) error = %v", err)
				}
				if msg != nil {
					t.Errorf("FromProtoMessage(nil) = %+v, want nil", msg)
				}
			}
		})
	}
}

func TestPartToProtoRoundTrip(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		part Part
	}{
		"TextPart": {
			part: NewTextPart("some text"),
		},
		"FilePart with URI": {
			part: NewFilePart(&FileWithURI{URI: "file:///tmp/test", MIMEType: "text/plain"}),
		},
		"FilePart with Bytes": {
			part: NewFilePart(&FileWithBytes{Bytes: "file content", MIMEType: "text/plain"}),
		},
		"DataPart": {
			part: NewDataPart(map[string]any{"field": "value"}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoPart, err := ToProtoPart(tt.part)
			if err != nil {
				t.Fatalf("ToProtoPart() error = %v", err)
			}

			convertedPart, err := FromProtoPart(protoPart)
			if err != nil {
				t.Fatalf("FromProtoPart() error = %v", err)
			}

			if diff := gocmp.Diff(tt.part, convertedPart); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestTaskToProtoRoundTrip(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		task *Task
	}{
		"complete task": {
			task: &Task{
				ID:        "task123",
				ContextID: "ctx456",
				Status:    TaskStatus{State: TaskStateWorking},
				History: []*Message{
					{
						MessageID: "msg1",
						Role:      RoleUser,
						Parts:     []Part{NewTextPart("User query")},
						Kind:      MessageEventKind,
					},
					{
						MessageID: "msg2",
						Role:      RoleAgent,
						Parts:     []Part{NewTextPart("Agent response")},
						Kind:      MessageEventKind,
					},
				},
				Artifacts: []*Artifact{
					{ArtifactID: "art1", Name: "artifact1", Parts: []Part{NewTextPart("artifact content")}},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoTask, err := ToProtoTask(tt.task)
			if err != nil {
				t.Fatalf("ToProtoTask() error = %v", err)
			}

			convertedTask, err := FromProtoTask(protoTask)
			if err != nil {
				t.Fatalf("FromProtoTask() error = %v", err)
			}

			if diff := gocmp.Diff(tt.task, convertedTask); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestTaskStateToProtoRoundTrip(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		goState    TaskState
		protoState a2a_v1.TaskState
	}{
		"TaskStateSubmitted": {
			goState:    TaskStateSubmitted,
			protoState: a2a_v1.TaskState_TASK_STATE_SUBMITTED,
		},
		"TaskStateWorking": {
			goState:    TaskStateWorking,
			protoState: a2a_v1.TaskState_TASK_STATE_WORKING,
		},
		"TaskStateCompleted": {
			goState:    TaskStateCompleted,
			protoState: a2a_v1.TaskState_TASK_STATE_COMPLETED,
		},
		"TaskStateFailed": {
			goState:    TaskStateFailed,
			protoState: a2a_v1.TaskState_TASK_STATE_FAILED,
		},
		"TaskStateCanceled": {
			goState:    TaskStateCanceled,
			protoState: a2a_v1.TaskState_TASK_STATE_CANCELLED,
		},
		"TaskStateRejected": {
			goState:    TaskStateRejected,
			protoState: a2a_v1.TaskState_TASK_STATE_REJECTED,
		},
		"TaskStateInputRequired": {
			goState:    TaskStateInputRequired,
			protoState: a2a_v1.TaskState_TASK_STATE_INPUT_REQUIRED,
		},
		"TaskStateAuthRequired": {
			goState:    TaskStateAuthRequired,
			protoState: a2a_v1.TaskState_TASK_STATE_AUTH_REQUIRED,
		},
		"TaskStateUnknown": {
			goState:    TaskStateUnknown,
			protoState: a2a_v1.TaskState_TASK_STATE_UNSPECIFIED,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pState := ToProtoTaskState(tt.goState)
			if pState != tt.protoState {
				t.Errorf("ToProtoTaskState() = %v, want %v", pState, tt.protoState)
			}
			gState := FromProtoTaskState(tt.protoState)
			// Note: FromProtoTaskState maps UNSPECIFIED to TaskStateUnknown
			if tt.protoState == a2a_v1.TaskState_TASK_STATE_UNSPECIFIED {
				if gState != TaskStateUnknown {
					t.Errorf("FromProtoTaskState() for UNSPECIFIED = %v, want %v", gState, TaskStateUnknown)
				}
			} else if gState != tt.goState {
				t.Errorf("FromProtoTaskState() = %v, want %v", gState, tt.goState)
			}
		})
	}
}

func TestAgentCardToProtoRoundTrip(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		card *AgentCard
	}{
		"standard agent card": {
			card: &AgentCard{
				Name:    "TestAgent",
				Version: "1.0.0",
				Skills: []*AgentSkill{
					{ID: "skill1", Name: "Test Skill"},
				},
				SecuritySchemes: map[string]SecurityScheme{
					"apiKey": &APIKeySecurityScheme{Name: "X-API-KEY", In: InHeader, Type: APIKeySecuritySchemeType},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoCard, err := ToProtoAgentCard(tt.card)
			if err != nil {
				t.Fatalf("ToProtoAgentCard() error = %v", err)
			}

			convertedCard, err := FromProtoAgentCard(protoCard)
			if err != nil {
				t.Fatalf("FromProtoAgentCard() error = %v", err)
			}

			if diff := gocmp.Diff(tt.card, convertedCard); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestSecuritySchemeToProtoRoundTrip(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		scheme SecurityScheme
	}{
		"APIKey": {
			scheme: &APIKeySecurityScheme{Description: "api key", Name: "X-API-KEY", In: InHeader, Type: APIKeySecuritySchemeType},
		},
		"HTTPAuth": {
			scheme: &HTTPAuthSecurityScheme{Description: "http auth", Scheme: "bearer", BearerFormat: "JWT", Type: HTTPSecuritySchemeType},
		},
		"OAuth2 AuthorizationCode": {
			scheme: &OAuth2SecurityScheme{
				Description: "oauth2",
				Flows: &OAuthFlows{
					AuthorizationCode: &AuthorizationCodeOAuthFlow{
						AuthorizationURL: "https://example.com/auth",
						TokenURL:         "https://example.com/token",
						Scopes:           map[string]string{"read": "read data"},
					},
				},
				Type: OAuth2SecuritySchemeType,
			},
		},
		"OpenIDConnect": {
			scheme: &OpenIDConnectSecurityScheme{Description: "oidc", OpenIDConnectURL: "https://example.com/.well-known/openid-configuration", Type: OpenIDConnectSecuritySchemeType},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoScheme, err := ToProtoSecurityScheme(tt.scheme)
			if err != nil {
				t.Fatalf("ToProtoSecurityScheme() error = %v", err)
			}

			convertedScheme, err := FromProtoSecurityScheme(protoScheme)
			if err != nil {
				t.Fatalf("FromProtoSecurityScheme() error = %v", err)
			}

			if diff := gocmp.Diff(tt.scheme, convertedScheme); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestFromProtoTaskIDParams(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		testFunc func(t *testing.T)
	}{
		"CancelTaskRequest": {
			testFunc: func(t *testing.T) {
				req := &a2a_v1.CancelTaskRequest{Name: "tasks/task123"}
				params, err := FromProtoTaskIDParams(req)
				if err != nil {
					t.Fatalf("FromProtoTaskIDParams() error = %v", err)
				}
				if params.ID != "task123" {
					t.Errorf("Got ID %q, want %q", params.ID, "task123")
				}
			},
		},
		"TaskSubscriptionRequest": {
			testFunc: func(t *testing.T) {
				req := &a2a_v1.TaskSubscriptionRequest{Name: "tasks/task456"}
				params, err := FromProtoTaskIDParams(req)
				if err != nil {
					t.Fatalf("FromProtoTaskIDParams() error = %v", err)
				}
				if params.ID != "task456" {
					t.Errorf("Got ID %q, want %q", params.ID, "task456")
				}
			},
		},
		"GetTaskPushNotificationConfigRequest": {
			testFunc: func(t *testing.T) {
				req := &a2a_v1.GetTaskPushNotificationConfigRequest{Name: "tasks/task789/pushNotificationConfigs/task789"}
				params, err := FromProtoTaskIDParams(req)
				if err != nil {
					t.Fatalf("FromProtoTaskIDParams() error = %v", err)
				}
				if params.ID != "task789" {
					t.Errorf("Got ID %q, want %q", params.ID, "task789")
				}
			},
		},
		"InvalidName": {
			testFunc: func(t *testing.T) {
				req := &a2a_v1.CancelTaskRequest{Name: "invalid/name"}
				_, err := FromProtoTaskIDParams(req)
				if err == nil {
					t.Error("Expected an error but got nil")
				}
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestMetadataToProtoRoundTrip(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		metadata map[string]any
	}{
		"complex metadata": {
			metadata: map[string]any{
				"string": "value",
				"number": 123.45,
				"bool":   true,
				"list":   []any{"a", float64(1)},
				"nested": map[string]any{"key": "nested_value"},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoStruct, err := ToProtoMetadata(tt.metadata)
			if err != nil {
				t.Fatalf("ToProtoMetadata() error = %v", err)
			}

			// Check if it's a valid structpb.Struct
			if _, ok := protoStruct.Fields["string"]; !ok {
				t.Error("ToProtoMetadata() did not convert correctly")
			}

			convertedMetadata, err := FromProtoMetadata(protoStruct)
			if err != nil {
				t.Fatalf("FromProtoMetadata() error = %v", err)
			}

			if diff := gocmp.Diff(tt.metadata, convertedMetadata); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestTaskStatusUpdateEventToProtoRoundTrip(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		event *TaskStatusUpdateEvent
	}{
		"completed status event": {
			event: &TaskStatusUpdateEvent{
				TaskID:    "task-1",
				ContextID: "ctx-1",
				Status:    TaskStatus{State: TaskStateCompleted},
				Kind:      StatusUpdateEventKind,
				Metadata:  map[string]any{"final_status": "success"},
				Final:     true,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoEvent, err := ToProtoTaskStatusUpdateEvent(tt.event)
			if err != nil {
				t.Fatalf("ToProtoTaskStatusUpdateEvent() error = %v", err)
			}

			convertedEvent, err := FromProtoTaskStatusUpdateEvent(protoEvent)
			if err != nil {
				t.Fatalf("FromProtoTaskStatusUpdateEvent() error = %v", err)
			}

			if diff := gocmp.Diff(tt.event, convertedEvent); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestTaskArtifactUpdateEventToProtoRoundTrip(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		originalEvent *TaskArtifactUpdateEvent
		expectedEvent *TaskArtifactUpdateEvent
	}{
		"artifact update event": {
			originalEvent: &TaskArtifactUpdateEvent{
				TaskID:    "task-2",
				ContextID: "ctx-2",
				Artifact: &Artifact{
					ArtifactID: "art-2",
					Name:       "result.txt",
					Parts:      []Part{NewTextPart("final result")},
				},
				Kind:      ArtifactUpdateEventKind,
				Append:    false,
				LastChunk: true,
			},
			// FromProto conversion adds Kind and empty Metadata, so we adjust the expected struct
			expectedEvent: &TaskArtifactUpdateEvent{
				TaskID:    "task-2",
				ContextID: "ctx-2",
				Artifact: &Artifact{
					ArtifactID: "art-2",
					Name:       "result.txt",
					Parts:      []Part{NewTextPart("final result")},
				},
				Append:    false,
				LastChunk: true,
				Kind:      ArtifactUpdateEventKind,
				Metadata:  map[string]any{},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoEvent, err := ToProtoTaskArtifactUpdateEvent(tt.originalEvent)
			if err != nil {
				t.Fatalf("ToProtoTaskArtifactUpdateEvent() error = %v", err)
			}

			convertedEvent, err := FromProtoTaskArtifactUpdateEvent(protoEvent)
			if err != nil {
				t.Fatalf("FromProtoTaskArtifactUpdateEvent() error = %v", err)
			}

			if diff := gocmp.Diff(tt.expectedEvent, convertedEvent); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestPushNotificationConfigToProtoRoundTrip(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config *PushNotificationConfig
	}{
		"push notification config": {
			config: &PushNotificationConfig{
				ID:    "pnc-1",
				URL:   "https://example.com/notify",
				Token: "secret-token",
				Authentication: &PushNotificationAuthenticationInfo{
					Schemes:     []string{"bearer"},
					Credentials: "{\"bearer\": \"token\"}",
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoConfig, err := ToProtoPushNotificationConfig(tt.config)
			if err != nil {
				t.Fatalf("ToProtoPushNotificationConfig() error = %v", err)
			}

			convertedConfig, err := FromProtoPushNotificationConfig(protoConfig)
			if err != nil {
				t.Fatalf("FromProtoPushNotificationConfig() error = %v", err)
			}

			if diff := gocmp.Diff(tt.config, convertedConfig); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}

func createComplexMetadata() map[string]any {
	return map[string]any{
		"string":    "test_value",
		"int":       123,
		"float":     45.67,
		"bool":      true,
		"null":      nil,
		"array":     []any{"a", "b", 123, true},
		"nested":    map[string]any{"key": "value", "number": 456},
		"unicode":   "测试🔥",
		"empty_str": "",
		"zero":      0,
		"false":     false,
	}
}

func createLargeMetadata() map[string]any {
	metadata := make(map[string]any)
	for i := range 1000 {
		metadata[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}
	return metadata
}

// Additional comprehensive tests for missing functions

func TestToProtoMessageSendConfiguration(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config *MessageSendConfiguration
		want   bool // true if should succeed
	}{
		"Nil config": {
			config: nil,
			want:   true,
		},
		"Empty config": {
			config: &MessageSendConfiguration{},
			want:   true,
		},
		"Full config": {
			config: &MessageSendConfiguration{
				AcceptedOutputModes: []string{"text", "json"},
				HistoryLength:       10,
				Blocking:            true,
				PushNotificationConfig: &PushNotificationConfig{
					ID:    "test-id",
					URL:   "https://example.com/notify",
					Token: "test-token",
				},
			},
			want: true,
		},
		"Negative history length": {
			config: &MessageSendConfiguration{
				HistoryLength: -1,
			},
			want: true, // Should still work, will be cast to int32
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoConfig, err := ToProtoMessageSendConfiguration(tt.config)
			if tt.want {
				if err != nil {
					t.Errorf("ToProtoMessageSendConfiguration() error = %v", err)
				}
				if tt.config == nil {
					if protoConfig == nil {
						t.Error("Expected non-nil proto config for nil input")
					}
				} else {
					if protoConfig.HistoryLength != int32(tt.config.HistoryLength) {
						t.Errorf("HistoryLength: got %d, want %d", protoConfig.HistoryLength, tt.config.HistoryLength)
					}
					if protoConfig.Blocking != tt.config.Blocking {
						t.Errorf("Blocking: got %t, want %t", protoConfig.Blocking, tt.config.Blocking)
					}
				}
			} else {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			}
		})
	}
}

func TestToProtoUpdateEvent(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		event SendStreamingMessageResponse
		want  bool
	}{
		"Nil event": {
			event: nil,
			want:  true,
		},
		"TaskStatusUpdateEvent": {
			event: &TaskStatusUpdateEvent{
				TaskID:    "task-1",
				ContextID: "ctx-1",
				Kind:      StatusUpdateEventKind,
				Status:    TaskStatus{State: TaskStateCompleted},
			},
			want: true,
		},
		"TaskArtifactUpdateEvent": {
			event: &TaskArtifactUpdateEvent{
				TaskID:    "task-2",
				ContextID: "ctx-2",
				Kind:      ArtifactUpdateEventKind,
				Artifact: &Artifact{
					ArtifactID: "art-1",
					Name:       "test.txt",
					Parts:      []Part{NewTextPart("content")},
				},
			},
			want: true,
		},
		"Message": {
			event: &Message{
				MessageID: "msg-1",
				Role:      RoleUser,
				Parts:     []Part{NewTextPart("Hello")},
				Kind:      MessageEventKind,
			},
			want: true,
		},
		"Task": {
			event: &Task{
				ID:     "task-3",
				Status: TaskStatus{State: TaskStateWorking},
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoEvent, err := ToProtoUpdateEvent(tt.event)
			if tt.want {
				if err != nil {
					t.Errorf("ToProtoUpdateEvent() error = %v", err)
				}
				if tt.event == nil {
					if protoEvent != nil {
						t.Error("Expected nil proto event for nil input")
					}
				} else {
					if protoEvent == nil {
						t.Error("Expected non-nil proto event")
					}
				}
			} else {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			}
		})
	}
}

func TestToProtoTaskOrMessage(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		event MessageOrTask
		want  bool
	}{
		"Nil event": {
			event: nil,
			want:  true,
		},
		"Message": {
			event: &Message{
				MessageID: "msg-1",
				Role:      RoleUser,
				Parts:     []Part{NewTextPart("Hello")},
				Kind:      MessageEventKind,
			},
			want: true,
		},
		"Task": {
			event: &Task{
				ID:     "task-1",
				Status: TaskStatus{State: TaskStateCompleted},
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			protoResp, err := ToProtoTaskOrMessage(tt.event)
			if tt.want {
				if err != nil {
					t.Errorf("ToProtoTaskOrMessage() error = %v", err)
				}
				if tt.event == nil {
					if protoResp != nil {
						t.Error("Expected nil proto response for nil input")
					}
				} else {
					if protoResp == nil {
						t.Error("Expected non-nil proto response")
					}
				}
			} else {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			}
		})
	}
}

func TestToProtoStreamResponse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		event SendStreamingMessageResponse
		want  bool
	}{
		"Nil event": {
			event: nil,
			want:  true,
		},
		"Message": {
			event: &Message{
				MessageID: "msg-1",
				Role:      RoleUser,
				Parts:     []Part{NewTextPart("Hello")},
				Kind:      MessageEventKind,
			},
			want: true,
		},
		"Task": {
			event: &Task{
				ID:     "task-1",
				Status: TaskStatus{State: TaskStateCompleted},
			},
			want: true,
		},
		"TaskStatusUpdateEvent": {
			event: &TaskStatusUpdateEvent{
				TaskID:    "task-1",
				ContextID: "ctx-1",
				Kind:      StatusUpdateEventKind,
				Status:    TaskStatus{State: TaskStateCompleted},
			},
			want: true,
		},
		"TaskArtifactUpdateEvent": {
			event: &TaskArtifactUpdateEvent{
				TaskID:    "task-2",
				ContextID: "ctx-2",
				Kind:      ArtifactUpdateEventKind,
				Artifact: &Artifact{
					ArtifactID: "art-1",
					Name:       "test.txt",
					Parts:      []Part{NewTextPart("content")},
				},
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoResp, err := ToProtoStreamResponse(tt.event)
			if tt.want {
				if err != nil {
					t.Errorf("ToProtoStreamResponse() error = %v", err)
				}
				if tt.event == nil {
					if protoResp != nil {
						t.Error("Expected nil proto response for nil input")
					}
				} else {
					if protoResp == nil {
						t.Error("Expected non-nil proto response")
					}
				}
			} else {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			}
		})
	}
}

func TestToProtoTaskPushNotificationConfig(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config *TaskPushNotificationConfig
		want   bool
	}{
		"Nil config": {
			config: nil,
			want:   true,
		},
		"Valid config": {
			config: &TaskPushNotificationConfig{
				TaskID: "task-123",
				PushNotificationConfig: &PushNotificationConfig{
					ID:    "notify-1",
					URL:   "https://example.com/webhook",
					Token: "secret-token",
				},
			},
			want: true,
		},
		"Config without notification": {
			config: &TaskPushNotificationConfig{
				TaskID: "task-456",
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoConfig, err := ToProtoTaskPushNotificationConfig(tt.config)
			if tt.want {
				if err != nil {
					t.Errorf("ToProtoTaskPushNotificationConfig() error = %v", err)
				}
				if tt.config == nil {
					if protoConfig != nil {
						t.Error("Expected nil proto config for nil input")
					}
				} else {
					if protoConfig == nil {
						t.Error("Expected non-nil proto config")
					}
					expectedName := fmt.Sprintf("tasks/%s/pushNotificationConfigs/%s", tt.config.TaskID, tt.config.TaskID)
					if protoConfig.Name != expectedName {
						t.Errorf("Name: got %q, want %q", protoConfig.Name, expectedName)
					}
				}
			} else {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			}
		})
	}
}

func TestFromProtoMessageSendParams(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		request *a2a_v1.SendMessageRequest
		want    bool
	}{
		"Nil request": {
			request: nil,
			want:    true,
		},
		"Valid request": {
			request: &a2a_v1.SendMessageRequest{
				Request: &a2a_v1.Message{
					MessageId: "msg-1",
					Role:      a2a_v1.Role_ROLE_USER,
					Content: []*a2a_v1.Part{
						{
							Part: &a2a_v1.Part_Text{
								Text: "Hello",
							},
						},
					},
				},
				Configuration: &a2a_v1.SendMessageConfiguration{
					AcceptedOutputModes: []string{"text"},
					HistoryLength:       5,
					Blocking:            true,
				},
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			params, err := FromProtoMessageSendParams(tt.request)
			if tt.want {
				if err != nil {
					t.Errorf("FromProtoMessageSendParams() error = %v", err)
				}
				if tt.request == nil {
					if params != nil {
						t.Error("Expected nil params for nil input")
					}
				} else {
					if params == nil {
						t.Error("Expected non-nil params")
					}
				}
			} else {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			}
		})
	}
}

func TestFromProtoTaskQueryParams(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		request *a2a_v1.GetTaskRequest
		want    bool
	}{
		"Nil request": {
			request: nil,
			want:    true,
		},
		"Valid request": {
			request: &a2a_v1.GetTaskRequest{
				Name:          "tasks/task-123",
				HistoryLength: 10,
			},
			want: true,
		},
		"Invalid name format": {
			request: &a2a_v1.GetTaskRequest{
				Name: "invalid-name",
			},
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			params, err := FromProtoTaskQueryParams(tt.request)
			if tt.want {
				if err != nil {
					t.Errorf("FromProtoTaskQueryParams() error = %v", err)
				}
				if tt.request == nil {
					if params != nil {
						t.Error("Expected nil params for nil input")
					}
				} else {
					if params == nil {
						t.Error("Expected non-nil params")
					}
					if params.HistoryLength != int(tt.request.HistoryLength) {
						t.Errorf("HistoryLength: got %d, want %d", params.HistoryLength, tt.request.HistoryLength)
					}
				}
			} else {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			}
		})
	}
}

// Error handling tests

func TestToProtoMetadata_InvalidStructure(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		metadata map[string]any
		wantErr  bool
	}{
		"Nil metadata": {
			metadata: nil,
			wantErr:  false,
		},
		"Valid metadata": {
			metadata: map[string]any{
				"string": "value",
				"number": 123,
			},
			wantErr: false,
		},
		"Complex valid metadata": {
			metadata: createComplexMetadata(),
			wantErr:  false,
		},
		"Metadata with function (should fail)": {
			metadata: map[string]any{
				"func": func() {},
			},
			wantErr: true,
		},
		"Metadata with channel (should fail)": {
			metadata: map[string]any{
				"chan": make(chan int),
			},
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			protoStruct, err := ToProtoMetadata(tt.metadata)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ToProtoMetadata() error = %v", err)
				}
				if tt.metadata == nil {
					if protoStruct != nil {
						t.Error("Expected nil proto struct for nil input")
					}
				}
			}
		})
	}
}

func TestInvalidEnumValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		testFunc func(t *testing.T)
	}{
		"Invalid TaskState": {
			testFunc: func(t *testing.T) {
				invalidState := TaskState("invalid_state")
				protoState := ToProtoTaskState(invalidState)
				if protoState != a2a_v1.TaskState_TASK_STATE_UNSPECIFIED {
					t.Errorf("Expected TASK_STATE_UNSPECIFIED for invalid state, got %v", protoState)
				}
			},
		},
		"Invalid Role": {
			testFunc: func(t *testing.T) {
				invalidRole := Role("invalid_role")
				protoRole := ToProtoRole(invalidRole)
				if protoRole != a2a_v1.Role_ROLE_UNSPECIFIED {
					t.Errorf("Expected ROLE_UNSPECIFIED for invalid role, got %v", protoRole)
				}
			},
		},
		"Invalid Proto TaskState": {
			testFunc: func(t *testing.T) {
				invalidProtoState := a2a_v1.TaskState(999)
				state := FromProtoTaskState(invalidProtoState)
				if state != TaskStateUnknown {
					t.Errorf("Expected TaskStateUnknown for invalid proto state, got %v", state)
				}
			},
		},
		"Invalid Proto Role": {
			testFunc: func(t *testing.T) {
				invalidProtoRole := a2a_v1.Role(999)
				role := FromProtoRole(invalidProtoRole)
				if role != RoleAgent {
					t.Errorf("Expected RoleAgent for invalid proto role, got %v", role)
				}
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestNilHandling_ComprehensiveTests(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		testFunc func(t *testing.T)
	}{
		"ToProtoFile with nil": {
			testFunc: func(t *testing.T) {
				protoFile, err := ToProtoFile(nil)
				if err != nil {
					t.Errorf("ToProtoFile(nil) error = %v", err)
				}
				if protoFile != nil {
					t.Error("Expected nil proto file for nil input")
				}
			},
		},
		"ToProtoData with nil": {
			testFunc: func(t *testing.T) {
				// ToProtoData with nil input creates an empty DataPart
				dataPart, err := ToProtoData(nil)
				if err != nil {
					t.Errorf("ToProtoData(nil) error = %v", err)
				}
				if dataPart == nil {
					t.Error("Expected non-nil DataPart for nil input")
				}
				if dataPart.Data == nil {
					t.Error("Expected non-nil Data struct")
				}
			},
		},
		"ToProtoAuthenticationInfo with nil": {
			testFunc: func(t *testing.T) {
				protoAuth, err := ToProtoAuthenticationInfo(nil)
				if err != nil {
					t.Errorf("ToProtoAuthenticationInfo(nil) error = %v", err)
				}
				if protoAuth != nil {
					t.Error("Expected nil proto auth for nil input")
				}
			},
		},
		"FromProtoFile with nil": {
			testFunc: func(t *testing.T) {
				file, err := FromProtoFile(nil)
				if err != nil {
					t.Errorf("FromProtoFile(nil) error = %v", err)
				}
				if file != nil {
					t.Error("Expected nil file for nil input")
				}
			},
		},
		"FromProtoData with nil": {
			testFunc: func(t *testing.T) {
				_, err := FromProtoData(nil)
				if err == nil {
					t.Error("Expected error for nil proto data")
				}
			},
		},
		"FromProtoAuthenticationInfo with nil": {
			testFunc: func(t *testing.T) {
				auth, err := FromProtoAuthenticationInfo(nil)
				if err != nil {
					t.Errorf("FromProtoAuthenticationInfo(nil) error = %v", err)
				}
				if auth != nil {
					t.Error("Expected nil auth for nil input")
				}
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestEmptyCollections(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		testFunc func(t *testing.T)
	}{
		"Empty vs nil slices": {
			testFunc: func(t *testing.T) {
				// Test with empty slice
				msgWithEmptyParts := &Message{
					MessageID: "msg-1",
					Role:      RoleUser,
					Parts:     []Part{}, // Empty slice
					Kind:      MessageEventKind,
				}
				protoMsg, err := ToProtoMessage(msgWithEmptyParts)
				if err != nil {
					t.Errorf("ToProtoMessage with empty parts error = %v", err)
				}
				if len(protoMsg.Content) != 0 {
					t.Errorf("Expected empty content, got %d items", len(protoMsg.Content))
				}

				// Test with nil slice
				msgWithNilParts := &Message{
					MessageID: "msg-2",
					Role:      RoleUser,
					Parts:     nil, // Nil slice
					Kind:      MessageEventKind,
				}
				protoMsg2, err := ToProtoMessage(msgWithNilParts)
				if err != nil {
					t.Errorf("ToProtoMessage with nil parts error = %v", err)
				}
				// ToProtoMessage creates an empty slice for nil Parts, not nil
				if len(protoMsg2.Content) != 0 {
					t.Errorf("Expected empty content, got %d items", len(protoMsg2.Content))
				}
			},
		},
		"Empty vs nil maps": {
			testFunc: func(t *testing.T) {
				// Test with empty map
				msgWithEmptyMetadata := &Message{
					MessageID: "msg-3",
					Role:      RoleUser,
					Parts:     []Part{NewTextPart("test")},
					Kind:      MessageEventKind,
					Metadata:  map[string]any{}, // Empty map
				}
				protoMsg, err := ToProtoMessage(msgWithEmptyMetadata)
				if err != nil {
					t.Errorf("ToProtoMessage with empty metadata error = %v", err)
				}
				if protoMsg.Metadata == nil || len(protoMsg.Metadata.Fields) != 0 {
					t.Error("Expected empty metadata struct")
				}

				// Test with nil map
				msgWithNilMetadata := &Message{
					MessageID: "msg-4",
					Role:      RoleUser,
					Parts:     []Part{NewTextPart("test")},
					Kind:      MessageEventKind,
					Metadata:  nil, // Nil map
				}
				protoMsg2, err := ToProtoMessage(msgWithNilMetadata)
				if err != nil {
					t.Errorf("ToProtoMessage with nil metadata error = %v", err)
				}
				if protoMsg2.Metadata != nil {
					t.Error("Expected nil metadata")
				}
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestLargeDataStructures(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		testFunc func(t *testing.T)
	}{
		"Large metadata conversion": {
			testFunc: func(t *testing.T) {
				largeMetadata := createLargeMetadata()
				protoStruct, err := ToProtoMetadata(largeMetadata)
				if err != nil {
					t.Errorf("ToProtoMetadata with large data error = %v", err)
				}
				if len(protoStruct.Fields) != len(largeMetadata) {
					t.Errorf("Expected %d fields, got %d", len(largeMetadata), len(protoStruct.Fields))
				}

				// Test round-trip
				convertedMetadata, err := FromProtoMetadata(protoStruct)
				if err != nil {
					t.Errorf("FromProtoMetadata with large data error = %v", err)
				}
				if len(convertedMetadata) != len(largeMetadata) {
					t.Errorf("Round-trip failed: expected %d fields, got %d", len(largeMetadata), len(convertedMetadata))
				}
			},
		},
		"Large message history": {
			testFunc: func(t *testing.T) {
				history := make([]*Message, 1000)
				for i := range 1000 {
					history[i] = &Message{
						MessageID: fmt.Sprintf("msg-%d", i),
						Role:      RoleUser,
						Parts:     []Part{NewTextPart(fmt.Sprintf("Message %d", i))},
						Kind:      MessageEventKind,
					}
				}

				task := &Task{
					ID:      "large-task",
					History: history,
				}

				protoTask, err := ToProtoTask(task)
				if err != nil {
					t.Errorf("ToProtoTask with large history error = %v", err)
				}
				if len(protoTask.History) != len(history) {
					t.Errorf("Expected %d history items, got %d", len(history), len(protoTask.History))
				}

				// Test round-trip
				convertedTask, err := FromProtoTask(protoTask)
				if err != nil {
					t.Errorf("FromProtoTask with large history error = %v", err)
				}
				if len(convertedTask.History) != len(history) {
					t.Errorf("Round-trip failed: expected %d history items, got %d", len(history), len(convertedTask.History))
				}
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestUnicodeHandling(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		text string
	}{
		"ASCII": {
			text: "Hello World",
		},
		"Latin": {
			text: "Héllo Wörld",
		},
		"Chinese": {
			text: "你好世界",
		},
		"Japanese": {
			text: "こんにちは世界",
		},
		"Korean": {
			text: "안녕하세요 세계",
		},
		"Arabic": {
			text: "مرحبا بالعالم",
		},
		"Hebrew": {
			text: "שלום עולם",
		},
		"Russian": {
			text: "Привет мир",
		},
		"Emoji": {
			text: "Hello 🌍🔥💯",
		},
		"Mixed": {
			text: "Hello 世界 🌍 test",
		},
		"Control chars": {
			text: "Hello\n\t\r\x00World",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			part := NewTextPart(tt.text)
			protoPart, err := ToProtoPart(part)
			if err != nil {
				t.Errorf("ToProtoPart error = %v", err)
			}

			convertedPart, err := FromProtoPart(protoPart)
			if err != nil {
				t.Errorf("FromProtoPart error = %v", err)
			}

			if diff := gocmp.Diff(part, convertedPart); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestBoundaryConditions(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		testFunc func(t *testing.T)
	}{
		"Integer overflow": {
			testFunc: func(t *testing.T) {
				config := &MessageSendConfiguration{
					HistoryLength: int(^uint(0) >> 1), // Max int
				}
				protoConfig, err := ToProtoMessageSendConfiguration(config)
				if err != nil {
					t.Errorf("ToProtoMessageSendConfiguration error = %v", err)
				}
				// Should be truncated to int32 max
				if protoConfig.HistoryLength != int32(config.HistoryLength) {
					t.Errorf("Expected truncated value, got %d", protoConfig.HistoryLength)
				}
			},
		},
		"Empty strings": {
			testFunc: func(t *testing.T) {
				msg := &Message{
					MessageID: "",
					ContextID: "",
					TaskID:    "",
					Role:      RoleUser,
					Parts:     []Part{NewTextPart("")},
					Kind:      MessageEventKind,
				}
				protoMsg, err := ToProtoMessage(msg)
				if err != nil {
					t.Errorf("ToProtoMessage with empty strings error = %v", err)
				}
				convertedMsg, err := FromProtoMessage(protoMsg)
				if err != nil {
					t.Errorf("FromProtoMessage with empty strings error = %v", err)
				}
				if diff := gocmp.Diff(msg, convertedMsg); diff != "" {
					t.Fatalf("(-want +got):\n%s", diff)
				}
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

// Performance benchmarks

func BenchmarkToProtoMessage(b *testing.B) {
	msg := &Message{
		MessageID: "benchmark-msg",
		ContextID: "benchmark-ctx",
		TaskID:    "benchmark-task",
		Role:      RoleUser,
		Parts:     []Part{NewTextPart("Benchmark message content")},
		Kind:      MessageEventKind,
		Metadata:  createComplexMetadata(),
	}

	for b.Loop() {
		_, err := ToProtoMessage(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFromProtoMessage(b *testing.B) {
	msg := &Message{
		MessageID: "benchmark-msg",
		ContextID: "benchmark-ctx",
		TaskID:    "benchmark-task",
		Role:      RoleUser,
		Parts:     []Part{NewTextPart("Benchmark message content")},
		Kind:      MessageEventKind,
		Metadata:  createComplexMetadata(),
	}

	protoMsg, err := ToProtoMessage(msg)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		_, err := FromProtoMessage(protoMsg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLargeMetadataConversion(b *testing.B) {
	metadata := createLargeMetadata()

	for b.Loop() {
		protoStruct, err := ToProtoMetadata(metadata)
		if err != nil {
			b.Fatal(err)
		}
		_, err = FromProtoMetadata(protoStruct)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkToProtoTask(b *testing.B) {
	task := &Task{
		ID:        "benchmark-task",
		ContextID: "benchmark-ctx",
		Status:    TaskStatus{State: TaskStateWorking},
		History: []*Message{
			{MessageID: "msg1", Role: RoleUser, Parts: []Part{NewTextPart("Hello")}},
			{MessageID: "msg2", Role: RoleAgent, Parts: []Part{NewTextPart("Hi there")}},
		},
		Artifacts: []*Artifact{
			{ArtifactID: "art1", Name: "test.txt", Parts: []Part{NewTextPart("content")}},
		},
	}

	for b.Loop() {
		_, err := ToProtoTask(task)
		if err != nil {
			b.Fatal(err)
		}
	}
}
