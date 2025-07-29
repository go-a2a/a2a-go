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

package agent_execution

import (
	"testing"
	"time"

	"github.com/go-a2a/a2a-go"
	"github.com/google/go-cmp/cmp"
)

func TestNewRequestContext(t *testing.T) {
	taskID := "task-123"
	contextID := "ctx-456"
	message := &a2a.Message{
		Role:      a2a.RoleUser,
		MessageID: "msg-789",
		Parts:     []a2a.Part{a2a.NewTextPart("Hello")},
	}

	before := time.Now()
	reqCtx := NewRequestContext(taskID, contextID, message)
	after := time.Now()

	if reqCtx.TaskID != taskID {
		t.Errorf("TaskID = %q, want %q", reqCtx.TaskID, taskID)
	}

	if reqCtx.ContextID != contextID {
		t.Errorf("ContextID = %q, want %q", reqCtx.ContextID, contextID)
	}

	if reqCtx.Message != message {
		t.Errorf("Message = %v, want %v", reqCtx.Message, message)
	}

	if reqCtx.CreatedAt.Before(before) || reqCtx.CreatedAt.After(after) {
		t.Errorf("CreatedAt = %v, want between %v and %v", reqCtx.CreatedAt, before, after)
	}

	if reqCtx.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

func TestRequestContext_WithTask(t *testing.T) {
	reqCtx := NewRequestContext("task-123", "ctx-456", &a2a.Message{})
	task := &a2a.Task{
		ID:        "task-123",
		ContextID: "ctx-456",
		Status:    a2a.TaskStatus{State: a2a.TaskStateWorking},
	}

	result := reqCtx.WithTask(task)

	if result != reqCtx {
		t.Error("WithTask should return the same context for chaining")
	}

	if reqCtx.Task != task {
		t.Errorf("Task = %v, want %v", reqCtx.Task, task)
	}
}

func TestRequestContext_WithMetadata(t *testing.T) {
	tests := map[string]struct {
		initial  map[string]any
		add      map[string]any
		expected map[string]any
	}{
		"success: add to empty metadata": {
			initial: nil,
			add: map[string]any{
				"key1": "value1",
				"key2": 42,
			},
			expected: map[string]any{
				"key1": "value1",
				"key2": 42,
			},
		},
		"success: merge metadata": {
			initial: map[string]any{
				"existing": "value",
			},
			add: map[string]any{
				"new": "data",
			},
			expected: map[string]any{
				"existing": "value",
				"new":      "data",
			},
		},
		"success: overwrite existing key": {
			initial: map[string]any{
				"key": "old",
			},
			add: map[string]any{
				"key": "new",
			},
			expected: map[string]any{
				"key": "new",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			reqCtx := NewRequestContext("task-123", "ctx-456", &a2a.Message{})
			if tc.initial != nil {
				reqCtx.Metadata = tc.initial
			}

			result := reqCtx.WithMetadata(tc.add)

			if result != reqCtx {
				t.Error("WithMetadata should return the same context for chaining")
			}

			if diff := cmp.Diff(tc.expected, reqCtx.Metadata); diff != "" {
				t.Errorf("Metadata mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRequestContext_WithAgentCard(t *testing.T) {
	reqCtx := NewRequestContext("task-123", "ctx-456", &a2a.Message{})
	agentCard := &a2a.AgentCard{
		Name:        "TestAgent",
		Description: "A test agent",
		Version:     "1.0.0",
	}

	result := reqCtx.WithAgentCard(agentCard)

	if result != reqCtx {
		t.Error("WithAgentCard should return the same context for chaining")
	}

	if reqCtx.AgentCard != agentCard {
		t.Errorf("AgentCard = %v, want %v", reqCtx.AgentCard, agentCard)
	}
}

func TestRequestContext_WithAuthInfo(t *testing.T) {
	reqCtx := NewRequestContext("task-123", "ctx-456", &a2a.Message{})
	authInfo := &AuthInfo{
		UserID: "user-123",
		Claims: map[string]any{
			"email": "user@example.com",
		},
		Scopes: []string{"read", "write"},
	}

	result := reqCtx.WithAuthInfo(authInfo)

	if result != reqCtx {
		t.Error("WithAuthInfo should return the same context for chaining")
	}

	if reqCtx.AuthInfo != authInfo {
		t.Errorf("AuthInfo = %v, want %v", reqCtx.AuthInfo, authInfo)
	}
}

func TestRequestContext_IsAuthenticated(t *testing.T) {
	tests := map[string]struct {
		authInfo *AuthInfo
		want     bool
	}{
		"authenticated with user ID": {
			authInfo: &AuthInfo{
				UserID: "user-123",
			},
			want: true,
		},
		"not authenticated - nil auth info": {
			authInfo: nil,
			want:     false,
		},
		"not authenticated - empty user ID": {
			authInfo: &AuthInfo{
				UserID: "",
			},
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			reqCtx := NewRequestContext("task-123", "ctx-456", &a2a.Message{})
			if tc.authInfo != nil {
				reqCtx.WithAuthInfo(tc.authInfo)
			}

			got := reqCtx.IsAuthenticated()
			if got != tc.want {
				t.Errorf("IsAuthenticated() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRequestContext_HasScope(t *testing.T) {
	tests := map[string]struct {
		authInfo *AuthInfo
		scope    string
		want     bool
	}{
		"has scope": {
			authInfo: &AuthInfo{
				UserID: "user-123",
				Scopes: []string{"read", "write", "admin"},
			},
			scope: "write",
			want:  true,
		},
		"does not have scope": {
			authInfo: &AuthInfo{
				UserID: "user-123",
				Scopes: []string{"read"},
			},
			scope: "write",
			want:  false,
		},
		"nil auth info": {
			authInfo: nil,
			scope:    "read",
			want:     false,
		},
		"empty scopes": {
			authInfo: &AuthInfo{
				UserID: "user-123",
				Scopes: []string{},
			},
			scope: "read",
			want:  false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			reqCtx := NewRequestContext("task-123", "ctx-456", &a2a.Message{})
			if tc.authInfo != nil {
				reqCtx.WithAuthInfo(tc.authInfo)
			}

			got := reqCtx.HasScope(tc.scope)
			if got != tc.want {
				t.Errorf("HasScope(%q) = %v, want %v", tc.scope, got, tc.want)
			}
		})
	}
}

func TestRequestContext_Chaining(t *testing.T) {
	message := &a2a.Message{Role: a2a.RoleUser}
	task := &a2a.Task{ID: "task-123"}
	agentCard := &a2a.AgentCard{Name: "TestAgent"}
	authInfo := &AuthInfo{UserID: "user-123"}
	metadata := map[string]any{"key": "value"}

	reqCtx := NewRequestContext("task-123", "ctx-456", message).
		WithTask(task).
		WithAgentCard(agentCard).
		WithAuthInfo(authInfo).
		WithMetadata(metadata)

	if reqCtx.Task != task {
		t.Error("Task not set correctly")
	}
	if reqCtx.AgentCard != agentCard {
		t.Error("AgentCard not set correctly")
	}
	if reqCtx.AuthInfo != authInfo {
		t.Error("AuthInfo not set correctly")
	}
	if diff := cmp.Diff(metadata, reqCtx.Metadata); diff != "" {
		t.Errorf("Metadata mismatch (-want +got):\n%s", diff)
	}
}