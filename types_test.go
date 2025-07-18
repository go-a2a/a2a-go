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
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-json-experiment/json"
	jsonv1 "github.com/go-json-experiment/json/v1"

	"github.com/go-a2a/a2a-go/internal/jsonrpc2"
)

func TestWireMessage(t *testing.T) {
	tests := map[string]struct {
		msg     jsonrpc2.Message
		encoded []byte
	}{
		"computing fix edits": {
			msg: newResponse(3, nil, jsonrpc2.NewError(0, "computing fix edits")),
			encoded: []byte(`{
		"jsonrpc":"2.0",
		"id":3,
		"error":{
			"code":0,
			"message":"computing fix edits"
		}
	}`),
		},
		"InternalError": {
			msg: newResponse(3, nil, ErrTaskNotFound),
			encoded: []byte(`{
		"jsonrpc":"2.0",
		"id":3,
		"error":{
			"code":-32001,
			"message":"JSON RPC task not found"
		}
	}`),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			b, err := jsonrpc2.EncodeMessage(tt.msg)
			if err != nil {
				t.Fatal(err)
			}
			checkJSON(t, b, tt.encoded)
			msg, err := jsonrpc2.DecodeMessage(tt.encoded)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(msg, tt.msg) {
				t.Errorf("decoded message does not match\nGot:\n%+#v\nWant:\n%+#v", msg, tt.msg)
			}
		})
	}
}

func newID(id any) jsonrpc2.ID {
	switch v := id.(type) {
	case nil:
		return jsonrpc2.ID{}
	case string:
		return jsonrpc2.StringID(v)
	case int:
		return jsonrpc2.Int64ID(int64(v))
	case int64:
		return jsonrpc2.Int64ID(v)
	default:
		panic("invalid ID type")
	}
}

func newResponse(id any, result any, rerr error) jsonrpc2.Message {
	msg, err := jsonrpc2.NewResponse(newID(id), result, rerr)
	if err != nil {
		panic(err)
	}
	return msg
}

func checkJSON(t *testing.T, got, want []byte) {
	// compare the compact form, to allow for formatting differences
	g := &bytes.Buffer{}
	if err := jsonv1.Compact(g, []byte(got)); err != nil {
		t.Fatal(err)
	}
	w := &bytes.Buffer{}
	if err := jsonv1.Compact(w, []byte(want)); err != nil {
		t.Fatal(err)
	}
	if g.String() != w.String() {
		t.Errorf("encoded message does not match\nGot:\n%s\nWant:\n%s", g, w)
	}
}

// TestEnums tests all enum constants
func TestEnums(t *testing.T) {
	t.Parallel()

	t.Run("In", func(t *testing.T) {
		t.Parallel()

		tests := map[string]struct {
			actual   In
			expected string
		}{
			"InCookie": {InCookie, "cookie"},
			"InHeader": {InHeader, "header"},
			"InQuery":  {InQuery, "query"},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				if string(tt.actual) != tt.expected {
					t.Errorf("constant %s = %q, want %q", name, string(tt.actual), tt.expected)
				}
			})
		}
	})

	t.Run("Role", func(t *testing.T) {
		t.Parallel()

		tests := map[string]struct {
			actual   Role
			expected string
		}{
			"RoleAgent": {RoleAgent, "agent"},
			"RoleUser":  {RoleUser, "user"},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				if string(tt.actual) != tt.expected {
					t.Errorf("constant %s = %q, want %q", name, string(tt.actual), tt.expected)
				}
			})
		}
	})

	t.Run("PartKind", func(t *testing.T) {
		t.Parallel()

		tests := map[string]struct {
			actual   PartKind
			expected string
		}{
			"TextPartKind": {TextPartKind, "text"},
			"FilePartKind": {FilePartKind, "file"},
			"DataPartKind": {DataPartKind, "data"},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				if string(tt.actual) != tt.expected {
					t.Errorf("constant %s = %q, want %q", name, string(tt.actual), tt.expected)
				}
			})
		}
	})

	t.Run("SecuritySchemeType", func(t *testing.T) {
		t.Parallel()

		tests := map[string]struct {
			actual   SecuritySchemeType
			expected string
		}{
			"APIKeySecuritySchemeType":        {APIKeySecuritySchemeType, "apiKey"},
			"HTTPSecuritySchemeType":          {HTTPSecuritySchemeType, "http"},
			"OpenIDConnectSecuritySchemeType": {OpenIDConnectSecuritySchemeType, "openIdConnect"},
			"OAuth2SecuritySchemeType":        {OAuth2SecuritySchemeType, "oauth2"},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				if string(tt.actual) != tt.expected {
					t.Errorf("constant %s = %q, want %q", name, string(tt.actual), tt.expected)
				}
			})
		}
	})

	t.Run("TaskState", func(t *testing.T) {
		t.Parallel()

		tests := map[string]struct {
			actual   TaskState
			expected string
		}{
			"TaskStateSubmitted":     {TaskStateSubmitted, "submitted"},
			"TaskStateWorking":       {TaskStateWorking, "working"},
			"TaskStateInputRequired": {TaskStateInputRequired, "input-required"},
			"TaskStateCompleted":     {TaskStateCompleted, "completed"},
			"TaskStateCanceled":      {TaskStateCanceled, "canceled"},
			"TaskStateFailed":        {TaskStateFailed, "failed"},
			"TaskStateRejected":      {TaskStateRejected, "rejected"},
			"TaskStateAuthRequired":  {TaskStateAuthRequired, "auth-required"},
			"TaskStateUnknown":       {TaskStateUnknown, "unknown"},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				if string(tt.actual) != tt.expected {
					t.Errorf("constant %s = %q, want %q", name, string(tt.actual), tt.expected)
				}
			})
		}
	})
}

// TestConstructors tests all constructor functions
func TestConstructors(t *testing.T) {
	t.Parallel()

	t.Run("NewTextPart", func(t *testing.T) {
		t.Parallel()

		text := "test text"
		part := NewTextPart(text)

		if part.GetKind() != TextPartKind {
			t.Errorf("GetKind() = %q, want %q", part.GetKind(), TextPartKind)
		}
		if part.Text != text {
			t.Errorf("Text = %q, want %q", part.Text, text)
		}
	})

	t.Run("NewFilePart", func(t *testing.T) {
		t.Parallel()

		file := &FileWithBytes{
			Bytes:    "dGVzdA==",
			MIMEType: "text/plain",
			Name:     "test.txt",
		}
		part := NewFilePart(file)

		if part.GetKind() != FilePartKind {
			t.Errorf("GetKind() = %q, want %q", part.GetKind(), FilePartKind)
		}
		if part.File != file {
			t.Errorf("File = %v, want %v", part.File, file)
		}
	})
}

// TestGetterMethods tests all getter methods
func TestGetterMethods(t *testing.T) {
	t.Parallel()

	t.Run("APIKeySecurityScheme", func(t *testing.T) {
		t.Parallel()

		scheme := APIKeySecurityScheme{
			Type:        APIKeySecuritySchemeType,
			Description: "API Key Authentication",
			In:          InHeader,
			Name:        "Authorization",
		}

		if scheme.GetType() != APIKeySecuritySchemeType {
			t.Errorf("GetType() = %q, want %q", scheme.GetType(), APIKeySecuritySchemeType)
		}
		if scheme.GetDescription() != "API Key Authentication" {
			t.Errorf("GetDescription() = %q, want %q", scheme.GetDescription(), "API Key Authentication")
		}
	})

	t.Run("HTTPAuthSecurityScheme", func(t *testing.T) {
		t.Parallel()

		scheme := HTTPAuthSecurityScheme{
			Type:        HTTPSecuritySchemeType,
			Description: "HTTP Authentication",
			Scheme:      "bearer",
		}

		if scheme.GetType() != HTTPSecuritySchemeType {
			t.Errorf("GetType() = %q, want %q", scheme.GetType(), HTTPSecuritySchemeType)
		}
		if scheme.GetDescription() != "HTTP Authentication" {
			t.Errorf("GetDescription() = %q, want %q", scheme.GetDescription(), "HTTP Authentication")
		}
	})

	t.Run("OpenIDConnectSecurityScheme", func(t *testing.T) {
		t.Parallel()

		scheme := OpenIDConnectSecurityScheme{
			Type:             OpenIDConnectSecuritySchemeType,
			Description:      "OpenID Connect",
			OpenIDConnectURL: "https://example.com/.well-known/openid-connect",
		}

		if scheme.GetType() != OpenIDConnectSecuritySchemeType {
			t.Errorf("GetType() = %q, want %q", scheme.GetType(), OpenIDConnectSecuritySchemeType)
		}
		if scheme.GetDescription() != "OpenID Connect" {
			t.Errorf("GetDescription() = %q, want %q", scheme.GetDescription(), "OpenID Connect")
		}
	})

	t.Run("OAuth2SecurityScheme", func(t *testing.T) {
		t.Parallel()

		scheme := OAuth2SecurityScheme{
			Type:        OAuth2SecuritySchemeType,
			Description: "OAuth 2.0",
			Flows: &OAuthFlows{
				AuthorizationCode: &AuthorizationCodeOAuthFlow{
					AuthorizationURL: "https://example.com/oauth/authorize",
					TokenURL:         "https://example.com/oauth/token",
					Scopes:           map[string]string{"read": "Read access"},
				},
			},
		}

		if scheme.GetType() != OAuth2SecuritySchemeType {
			t.Errorf("GetType() = %q, want %q", scheme.GetType(), OAuth2SecuritySchemeType)
		}
		if scheme.GetDescription() != "OAuth 2.0" {
			t.Errorf("GetDescription() = %q, want %q", scheme.GetDescription(), "OAuth 2.0")
		}
	})

	t.Run("FileWithBytes", func(t *testing.T) {
		t.Parallel()

		file := FileWithBytes{
			Bytes:    "dGVzdA==",
			MIMEType: "text/plain",
			Name:     "test.txt",
		}

		if file.GetMIMEType() != "text/plain" {
			t.Errorf("GetMIMEType() = %q, want %q", file.GetMIMEType(), "text/plain")
		}
		if file.GetName() != "test.txt" {
			t.Errorf("GetName() = %q, want %q", file.GetName(), "test.txt")
		}
	})

	t.Run("FileWithURI", func(t *testing.T) {
		t.Parallel()

		file := FileWithURI{
			URI:      "https://example.com/file.txt",
			MIMEType: "text/plain",
			Name:     "file.txt",
		}

		if file.GetMIMEType() != "text/plain" {
			t.Errorf("GetMIMEType() = %q, want %q", file.GetMIMEType(), "text/plain")
		}
		if file.GetName() != "file.txt" {
			t.Errorf("GetName() = %q, want %q", file.GetName(), "file.txt")
		}
	})

	t.Run("Message", func(t *testing.T) {
		t.Parallel()

		msg := Message{
			TaskID:    "test-message-id",
			Kind:      "message",
			ContextID: "test-context-id",
		}

		if msg.GetTaskID() != "test-message-id" {
			t.Errorf("GetID() = %q, want %q", msg.GetTaskID(), "test-message-id")
		}
		if msg.GetKind() != "message" {
			t.Errorf("GetKind() = %q, want %q", msg.GetKind(), "message")
		}
	})

	t.Run("Task", func(t *testing.T) {
		t.Parallel()

		task := Task{
			ID:        "test-task-id",
			Kind:      "task",
			ContextID: "test-context-id",
		}

		if task.GetTaskID() != "test-task-id" {
			t.Errorf("GetID() = %q, want %q", task.GetTaskID(), "test-task-id")
		}
		if task.GetKind() != "task" {
			t.Errorf("GetKind() = %q, want %q", task.GetKind(), "task")
		}
	})
}

// TestUnmarshalFunctions tests all unmarshal functions
func TestUnmarshalFunctions(t *testing.T) {
	t.Parallel()

	t.Run("UnmarshalA2ARequest", func(t *testing.T) {
		t.Parallel()

		tests := map[string]struct {
			jsonData    string
			expectError bool
			expectType  string
		}{
			"SendMessageRequest": {
				jsonData:    `{"id": "1", "jsonrpc": "2.0", "method": "message/send", "params": {"message": {"messageId": "test", "kind": "message", "role": "user", "parts": []}}}`,
				expectError: false,
				expectType:  "*a2a.MessageSendParams",
			},
			"SendStreamingMessageRequest": {
				jsonData:    `{"id": "1", "jsonrpc": "2.0", "method": "message/stream", "params": {"message": {"messageId": "test", "kind": "message", "role": "user", "parts": []}}}`,
				expectError: false,
				expectType:  "*a2a.MessageSendParams",
			},
			"GetTaskRequest": {
				jsonData:    `{"id": "1", "jsonrpc": "2.0", "method": "tasks/get", "params": {"id": "task-id"}}`,
				expectError: false,
				expectType:  "*a2a.TaskQueryParams",
			},
			"ListTaskRequest": {
				jsonData:    `{"id": "1", "jsonrpc": "2.0", "method": "tasks/list"}`,
				expectError: false,
				expectType:  "*a2a.EmptyParams",
			},
			"CancelTaskRequest": {
				jsonData:    `{"id": "1", "jsonrpc": "2.0", "method": "tasks/cancel", "params": {"id": "task-id"}}`,
				expectError: false,
				expectType:  "*a2a.TaskQueryParams",
			},
			"UnknownMethod": {
				jsonData:    `{"id": "1", "jsonrpc": "2.0", "method": "unknown/method", "params": {}}`,
				expectError: true,
				expectType:  "",
			},
			"InvalidJSON": {
				jsonData:    `{"id": "1", "method": "invalid}`,
				expectError: true,
				expectType:  "",
			},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				req, err := UnmarshalParams([]byte(tt.jsonData))
				if tt.expectError {
					if err == nil {
						t.Errorf("expected error but got none")
					}
					return
				}

				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if req == nil {
					t.Errorf("expected request but got nil")
					return
				}

				actualType := fmt.Sprintf("%T", req)
				if actualType != tt.expectType {
					t.Errorf("expected type %s but got %s", tt.expectType, actualType)
				}
			})
		}
	})

	t.Run("UnmarshalSecurityScheme", func(t *testing.T) {
		t.Parallel()

		tests := map[string]struct {
			jsonData    string
			expectError bool
			expectType  string
		}{
			"APIKeySecurityScheme": {
				jsonData:    `{"type": "apiKey", "name": "Authorization", "in": "header"}`,
				expectError: false,
				expectType:  "*a2a.APIKeySecurityScheme",
			},
			"HTTPAuthSecurityScheme": {
				jsonData:    `{"type": "http", "scheme": "bearer"}`,
				expectError: false,
				expectType:  "*a2a.HTTPAuthSecurityScheme",
			},
			"OAuth2SecurityScheme": {
				jsonData:    `{"type": "oauth2", "flows": {"authorizationCode": {"authorizationUrl": "https://example.com/oauth/authorize", "tokenUrl": "https://example.com/oauth/token", "scopes": {}}}}`,
				expectError: false,
				expectType:  "*a2a.OAuth2SecurityScheme",
			},
			"UnknownType": {
				jsonData:    `{"type": "unknown"}`,
				expectError: true,
				expectType:  "",
			},
			"InvalidJSON": {
				jsonData:    `{"type": "apiKey", "name": }`,
				expectError: true,
				expectType:  "",
			},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				scheme, err := UnmarshalSecurityScheme([]byte(tt.jsonData))

				if tt.expectError {
					if err == nil {
						t.Errorf("expected error but got none")
					}
					return
				}

				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if scheme == nil {
					t.Errorf("expected scheme but got nil")
					return
				}

				actualType := fmt.Sprintf("%T", scheme)
				if actualType != tt.expectType {
					t.Errorf("expected type %s but got %s", tt.expectType, actualType)
				}
			})
		}
	})

	t.Run("UnmarshalPart", func(t *testing.T) {
		t.Parallel()

		tests := map[string]struct {
			jsonData    string
			expectError bool
			expectType  string
		}{
			"TextPart": {
				jsonData:    `{"kind": "text", "text": "Hello world"}`,
				expectError: false,
				expectType:  "*a2a.TextPart",
			},
			"DataPart": {
				jsonData:    `{"kind": "data", "data": {"key": "value"}}`,
				expectError: false,
				expectType:  "*a2a.DataPart",
			},
			"FilePart": {
				jsonData:    `{"kind": "file", "file": {"bytes": "dGVzdA==", "mimeType": "text/plain"}}`,
				expectError: false,
				expectType:  "*a2a.FilePart",
			},
			"UnknownKind": {
				jsonData:    `{"kind": "unknown"}`,
				expectError: true,
				expectType:  "",
			},
			"InvalidJSON": {
				jsonData:    `{"kind": "text", "text": }`,
				expectError: true,
				expectType:  "",
			},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				part, err := UnmarshalPart([]byte(tt.jsonData))

				if tt.expectError {
					if err == nil {
						t.Errorf("expected error but got none")
					}
					return
				}

				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if part == nil {
					t.Errorf("expected part but got nil")
					return
				}

				actualType := fmt.Sprintf("%T", part)
				if actualType != tt.expectType {
					t.Errorf("expected type %s but got %s", tt.expectType, actualType)
				}
			})
		}
	})

	t.Run("UnmarshalFileContent", func(t *testing.T) {
		t.Parallel()

		tests := map[string]struct {
			jsonData    string
			expectError bool
			expectType  string
		}{
			"FileWithBytes": {
				jsonData:    `{"bytes": "dGVzdA==", "mimeType": "text/plain", "name": "test.txt"}`,
				expectError: false,
				expectType:  "*a2a.FileWithBytes",
			},
			"FileWithURI": {
				jsonData:    `{"uri": "https://example.com/file.txt", "mimeType": "text/plain", "name": "file.txt"}`,
				expectError: false,
				expectType:  "*a2a.FileWithURI",
			},
			"NoContent": {
				jsonData:    `{"mimeType": "text/plain", "name": "file.txt"}`,
				expectError: true,
				expectType:  "",
			},
			"InvalidJSON": {
				jsonData:    `{"bytes": "dGVzdA==", "mimeType": }`,
				expectError: true,
				expectType:  "",
			},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				file, err := UnmarshalFileContent([]byte(tt.jsonData))

				if tt.expectError {
					if err == nil {
						t.Errorf("expected error but got none")
					}
					return
				}

				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if file == nil {
					t.Errorf("expected file but got nil")
					return
				}

				actualType := fmt.Sprintf("%T", file)
				if actualType != tt.expectType {
					t.Errorf("expected type %s but got %s", tt.expectType, actualType)
				}
			})
		}
	})
}

// TestJSONMarshaling tests JSON marshaling/unmarshaling for all types
func TestJSONMarshaling(t *testing.T) {
	t.Parallel()

	t.Run("TextPart", func(t *testing.T) {
		t.Parallel()

		original := &TextPart{
			Kind:     TextPartKind,
			Text:     "Hello world",
			Metadata: map[string]any{"key": "value"},
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
			return
		}

		var unmarshaled TextPart
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Unmarshal() error = %v", err)
			return
		}

		if !reflect.DeepEqual(original, &unmarshaled) {
			t.Errorf("Marshal/Unmarshal mismatch: original = %+v, unmarshaled = %+v", original, &unmarshaled)
		}
	})

	t.Run("DataPart", func(t *testing.T) {
		t.Parallel()

		original := &DataPart{
			Kind:     DataPartKind,
			Data:     map[string]any{"key": "value", "number": 42.0},
			Metadata: map[string]any{"meta": "data"},
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
			return
		}

		var unmarshaled DataPart
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Unmarshal() error = %v", err)
			return
		}

		// Check individual fields instead of reflect.DeepEqual for better debugging
		if original.Kind != unmarshaled.Kind {
			t.Errorf("Kind mismatch: original = %v, unmarshaled = %v", original.Kind, unmarshaled.Kind)
		}

		if original.Data["key"] != unmarshaled.Data["key"] {
			t.Errorf("Data[key] mismatch: original = %v, unmarshaled = %v", original.Data["key"], unmarshaled.Data["key"])
		}

		if original.Data["number"] != unmarshaled.Data["number"] {
			t.Errorf("Data[number] mismatch: original = %v, unmarshaled = %v", original.Data["number"], unmarshaled.Data["number"])
		}

		if original.Metadata["meta"] != unmarshaled.Metadata["meta"] {
			t.Errorf("Metadata[meta] mismatch: original = %v, unmarshaled = %v", original.Metadata["meta"], unmarshaled.Metadata["meta"])
		}
	})

	t.Run("FileWithBytes", func(t *testing.T) {
		t.Parallel()

		original := &FileWithBytes{
			Bytes:    "dGVzdA==",
			MIMEType: "text/plain",
			Name:     "test.txt",
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
			return
		}

		var unmarshaled FileWithBytes
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Unmarshal() error = %v", err)
			return
		}

		if !reflect.DeepEqual(original, &unmarshaled) {
			t.Errorf("Marshal/Unmarshal mismatch: original = %+v, unmarshaled = %+v", original, &unmarshaled)
		}
	})

	t.Run("FileWithURI", func(t *testing.T) {
		t.Parallel()

		original := &FileWithURI{
			URI:      "https://example.com/file.txt",
			MIMEType: "text/plain",
			Name:     "file.txt",
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
			return
		}

		var unmarshaled FileWithURI
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Unmarshal() error = %v", err)
			return
		}

		if !reflect.DeepEqual(original, &unmarshaled) {
			t.Errorf("Marshal/Unmarshal mismatch: original = %+v, unmarshaled = %+v", original, &unmarshaled)
		}
	})

	t.Run("APIKeySecurityScheme", func(t *testing.T) {
		t.Parallel()

		original := &APIKeySecurityScheme{
			Type:        APIKeySecuritySchemeType,
			Description: "API Key Authentication",
			In:          InHeader,
			Name:        "Authorization",
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
			return
		}

		var unmarshaled APIKeySecurityScheme
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Unmarshal() error = %v", err)
			return
		}

		if !reflect.DeepEqual(original, &unmarshaled) {
			t.Errorf("Marshal/Unmarshal mismatch: original = %+v, unmarshaled = %+v", original, &unmarshaled)
		}
	})

	t.Run("HTTPAuthSecurityScheme", func(t *testing.T) {
		t.Parallel()

		original := &HTTPAuthSecurityScheme{
			Type:         HTTPSecuritySchemeType,
			Description:  "HTTP Authentication",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
			return
		}

		var unmarshaled HTTPAuthSecurityScheme
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Unmarshal() error = %v", err)
			return
		}

		if !reflect.DeepEqual(original, &unmarshaled) {
			t.Errorf("Marshal/Unmarshal mismatch: original = %+v, unmarshaled = %+v", original, &unmarshaled)
		}
	})

	t.Run("OAuth2SecurityScheme", func(t *testing.T) {
		t.Parallel()

		original := &OAuth2SecurityScheme{
			Type:        OAuth2SecuritySchemeType,
			Description: "OAuth 2.0",
			Flows: &OAuthFlows{
				AuthorizationCode: &AuthorizationCodeOAuthFlow{
					AuthorizationURL: "https://example.com/oauth/authorize",
					TokenURL:         "https://example.com/oauth/token",
					Scopes:           map[string]string{"read": "Read access", "write": "Write access"},
				},
				ClientCredentials: &ClientCredentialsOAuthFlow{
					TokenURL: "https://example.com/oauth/token",
					Scopes:   map[string]string{"client": "Client access"},
				},
			},
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
			return
		}

		var unmarshaled OAuth2SecurityScheme
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Unmarshal() error = %v", err)
			return
		}

		if !reflect.DeepEqual(original, &unmarshaled) {
			t.Errorf("Marshal/Unmarshal mismatch: original = %+v, unmarshaled = %+v", original, &unmarshaled)
		}
	})

	t.Run("TaskStatus", func(t *testing.T) {
		t.Parallel()

		original := TaskStatus{
			State:     TaskStateWorking,
			Timestamp: "2023-10-27T10:00:00Z",
			Message: &Message{
				MessageID: "status-msg-id",
				Kind:      "message",
				Role:      RoleAgent,
				Parts:     []Part{},
			},
		}

		data, err := json.Marshal(&original)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
			return
		}

		var unmarshaled TaskStatus
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Unmarshal() error = %v", err)
			return
		}

		if !reflect.DeepEqual(&original, &unmarshaled) {
			t.Errorf("Marshal/Unmarshal mismatch: original = %+v, unmarshaled = %+v", &original, &unmarshaled)
		}
	})

	t.Run("PushNotificationConfig", func(t *testing.T) {
		t.Parallel()

		original := &PushNotificationConfig{
			ID:    "push-config-id",
			Token: "push-token",
			URL:   "https://example.com/push",
			Authentication: &PushNotificationAuthenticationInfo{
				Schemes:     []string{"Bearer"},
				Credentials: "auth-token",
			},
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
			return
		}

		var unmarshaled PushNotificationConfig
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Unmarshal() error = %v", err)
			return
		}

		if !reflect.DeepEqual(original, &unmarshaled) {
			t.Errorf("Marshal/Unmarshal mismatch: original = %+v, unmarshaled = %+v", original, &unmarshaled)
		}
	})

	t.Run("AgentSkill", func(t *testing.T) {
		t.Parallel()

		original := &AgentSkill{
			ID:          "skill-id",
			Name:        "Test Skill",
			Description: "A test skill",
			Examples:    []string{"example 1", "example 2"},
			Tags:        []string{"tag1", "tag2"},
			InputModes:  []string{"text", "file"},
			OutputModes: []string{"text", "data"},
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
			return
		}

		var unmarshaled AgentSkill
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Unmarshal() error = %v", err)
			return
		}

		if !reflect.DeepEqual(original, &unmarshaled) {
			t.Errorf("Marshal/Unmarshal mismatch: original = %+v, unmarshaled = %+v", original, &unmarshaled)
		}
	})

	t.Run("AgentCapabilities", func(t *testing.T) {
		t.Parallel()

		original := &AgentCapabilities{
			PushNotifications:      true,
			Streaming:              true,
			StateTransitionHistory: false,
			Extensions: []*AgentExtension{
				{
					URI:         "https://example.com/extension",
					Description: "Test extension",
					Required:    true,
					Params:      map[string]any{"param": "value"},
				},
			},
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
			return
		}

		var unmarshaled AgentCapabilities
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Unmarshal() error = %v", err)
			return
		}

		if !reflect.DeepEqual(original, &unmarshaled) {
			t.Errorf("Marshal/Unmarshal mismatch: original = %+v, unmarshaled = %+v", original, &unmarshaled)
		}
	})
}

// TestTypesEdgeCases tests edge cases and error conditions
func TestTypesEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("Empty and nil values", func(t *testing.T) {
		t.Parallel()

		// Test with empty strings
		part := NewTextPart("")
		if part.GetKind() != TextPartKind {
			t.Errorf("NewTextPart(\"\") should still have correct kind")
		}
		if part.Text != "" {
			t.Errorf("NewTextPart(\"\") should preserve empty text")
		}

		// Test with nil data
		dataPart := NewDataPart(nil)
		if dataPart.GetKind() != DataPartKind {
			t.Errorf("NewDataPart(nil) should still have correct kind")
		}
		if dataPart.Data != nil {
			t.Errorf("NewDataPart(nil) should preserve nil data")
		}

		// Test with empty map
		emptyData := make(map[string]any)
		dataPart2 := NewDataPart(emptyData)
		if dataPart2.GetKind() != DataPartKind {
			t.Errorf("NewDataPart(empty map) should still have correct kind")
		}
		if len(dataPart2.Data) != 0 {
			t.Errorf("NewDataPart(empty map) should preserve empty map")
		}
	})

	t.Run("String enum values consistency", func(t *testing.T) {
		t.Parallel()

		// Test that string enum values are consistent
		if string(TextPartKind) != "text" {
			t.Errorf("TextPartKind string value inconsistent")
		}
		if string(FilePartKind) != "file" {
			t.Errorf("FilePartKind string value inconsistent")
		}
		if string(DataPartKind) != "data" {
			t.Errorf("DataPartKind string value inconsistent")
		}

		if string(RoleAgent) != "agent" {
			t.Errorf("RoleAgent string value inconsistent")
		}
		if string(RoleUser) != "user" {
			t.Errorf("RoleUser string value inconsistent")
		}

		if string(InCookie) != "cookie" {
			t.Errorf("InCookie string value inconsistent")
		}
		if string(InHeader) != "header" {
			t.Errorf("InHeader string value inconsistent")
		}
		if string(InQuery) != "query" {
			t.Errorf("InQuery string value inconsistent")
		}
	})
}

// TestMethodStrings tests method string values
func TestMethodStrings(t *testing.T) {
	t.Parallel()

	expectedMethods := map[Method]string{
		MethodMessageSend:                       "message/send",
		MethodMessageStream:                     "message/stream",
		MethodTasksGet:                          "tasks/get",
		MethodTasksList:                         "tasks/list",
		MethodTasksCancel:                       "tasks/cancel",
		MethodTasksPushNotificationConfigSet:    "tasks/pushNotificationConfig/set",
		MethodTasksPushNotificationConfigGet:    "tasks/pushNotificationConfig/get",
		MethodTasksPushNotificationConfigList:   "tasks/pushNotificationConfig/list",
		MethodTasksPushNotificationConfigDelete: "tasks/pushNotificationConfig/delete",
		MethodTasksResubscribe:                  "tasks/resubscribe",
		MethodAgentAuthenticatedExtendedCard:    "agent/authenticatedExtendedCard",
	}
	for method, expected := range expectedMethods {
		if string(method) != expected {
			t.Errorf("Method %s = %q, want %q", expected, string(method), expected)
		}
	}
}

// TestTaskStates tests task state values
func TestTaskStates(t *testing.T) {
	t.Parallel()

	expectedStates := map[TaskState]string{
		TaskStateSubmitted:     "submitted",
		TaskStateWorking:       "working",
		TaskStateInputRequired: "input-required",
		TaskStateCompleted:     "completed",
		TaskStateCanceled:      "canceled",
		TaskStateFailed:        "failed",
		TaskStateRejected:      "rejected",
		TaskStateAuthRequired:  "auth-required",
		TaskStateUnknown:       "unknown",
	}
	for state, expected := range expectedStates {
		if string(state) != expected {
			t.Errorf("TaskState %s = %q, want %q", expected, string(state), expected)
		}
	}
}

// TestSecuritySchemeTypes tests security scheme type values
func TestSecuritySchemeTypes(t *testing.T) {
	t.Parallel()

	expectedTypes := map[SecuritySchemeType]string{
		APIKeySecuritySchemeType:        "apiKey",
		HTTPSecuritySchemeType:          "http",
		OpenIDConnectSecuritySchemeType: "openIdConnect",
		OAuth2SecuritySchemeType:        "oauth2",
	}
	for schemeType, expected := range expectedTypes {
		if string(schemeType) != expected {
			t.Errorf("SecuritySchemeType %s = %q, want %q", expected, string(schemeType), expected)
		}
	}
}

// TestComplexNestedStructures tests complex nested structures and their marshaling
func TestComplexNestedStructures(t *testing.T) {
	t.Parallel()

	t.Run("AgentCard with complex nested data", func(t *testing.T) {
		t.Parallel()

		agentCard := &AgentCard{
			Name:        "Complex Agent",
			Description: "A complex agent with many features",
			URL:         "https://example.com/agent",
			Version:     "1.0.0",
			Capabilities: &AgentCapabilities{
				PushNotifications:      true,
				Streaming:              true,
				StateTransitionHistory: true,
				Extensions: []*AgentExtension{
					{
						URI:         "https://example.com/ext1",
						Description: "Extension 1",
						Required:    true,
						Params: map[string]any{
							"param1": "value1",
							"param2": 42.0,
							"param3": true,
						},
					},
					{
						URI:         "https://example.com/ext2",
						Description: "Extension 2",
						Required:    false,
						Params: map[string]any{
							"nested": map[string]any{
								"deep": "value",
							},
						},
					},
				},
			},
			Skills: []*AgentSkill{
				{
					ID:          "skill1",
					Name:        "Skill 1",
					Description: "First skill",
					Examples:    []string{"example 1", "example 2"},
					Tags:        []string{"tag1", "tag2"},
					InputModes:  []string{"text", "file"},
					OutputModes: []string{"text", "data"},
				},
				{
					ID:          "skill2",
					Name:        "Skill 2",
					Description: "Second skill",
					Examples:    []string{"example 3"},
					Tags:        []string{"tag3"},
					InputModes:  []string{"text"},
					OutputModes: []string{"text"},
				},
			},
			SecuritySchemes: map[string]SecurityScheme{
				"apiKey": &APIKeySecurityScheme{
					Type:        APIKeySecuritySchemeType,
					Description: "API Key auth",
					In:          InHeader,
					Name:        "Authorization",
				},
				"oauth2": &OAuth2SecurityScheme{
					Type:        OAuth2SecuritySchemeType,
					Description: "OAuth 2.0",
					Flows: &OAuthFlows{
						AuthorizationCode: &AuthorizationCodeOAuthFlow{
							AuthorizationURL: "https://example.com/oauth/authorize",
							TokenURL:         "https://example.com/oauth/token",
							Scopes: map[string]string{
								"read":  "Read access",
								"write": "Write access",
							},
						},
					},
				},
			},
			DefaultInputModes:  []string{"text", "file", "data"},
			DefaultOutputModes: []string{"text", "data"},
		}

		// Test marshaling
		data, err := json.Marshal(agentCard)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
			return
		}

		// Test unmarshaling
		var unmarshaled AgentCard
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Unmarshal() error = %v", err)
			return
		}

		// Verify critical fields
		if unmarshaled.Name != agentCard.Name {
			t.Errorf("Name mismatch: got %v, want %v", unmarshaled.Name, agentCard.Name)
		}
		if len(unmarshaled.Skills) != len(agentCard.Skills) {
			t.Errorf("Skills length mismatch: got %d, want %d", len(unmarshaled.Skills), len(agentCard.Skills))
		}
		if len(unmarshaled.SecuritySchemes) != len(agentCard.SecuritySchemes) {
			t.Errorf("SecuritySchemes length mismatch: got %d, want %d", len(unmarshaled.SecuritySchemes), len(agentCard.SecuritySchemes))
		}
		if len(unmarshaled.Capabilities.Extensions) != len(agentCard.Capabilities.Extensions) {
			t.Errorf("Extensions length mismatch: got %d, want %d", len(unmarshaled.Capabilities.Extensions), len(agentCard.Capabilities.Extensions))
		}
	})

	t.Run("Task with complex history and artifacts", func(t *testing.T) {
		t.Parallel()

		task := &Task{
			ID:        "complex-task-id",
			Kind:      "task",
			ContextID: "complex-context-id",
			Status: TaskStatus{
				State:     TaskStateCompleted,
				Timestamp: "2023-10-27T10:00:00Z",
				Message: &Message{
					MessageID: "status-msg-id",
					Kind:      "message",
					Role:      RoleAgent,
					Parts: []Part{
						NewTextPart("Task completed successfully"),
					},
				},
			},
			History: []*Message{
				{
					MessageID: "msg1",
					Kind:      "message",
					Role:      RoleUser,
					Parts: []Part{
						NewTextPart("Initial request"),
						NewDataPart(map[string]any{
							"priority": "high",
							"category": "urgent",
						}),
					},
				},
				{
					MessageID: "msg2",
					Kind:      "message",
					Role:      RoleAgent,
					Parts: []Part{
						NewTextPart("Processing request"),
						NewFilePart(&FileWithURI{
							URI:      "https://example.com/processing.log",
							MIMEType: "text/plain",
							Name:     "processing.log",
						}),
					},
				},
			},
			Artifacts: []*Artifact{
				{
					ArtifactID:  "artifact1",
					Name:        "Result Document",
					Description: "The final result",
					Parts: []Part{
						NewTextPart("Final result content"),
						NewDataPart(map[string]any{
							"resultType": "success",
							"score":      95.5,
						}),
					},
				},
			},
		}

		// Test marshaling
		data, err := json.Marshal(task)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
			return
		}

		// Test unmarshaling
		var unmarshaled Task
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Unmarshal() error = %v", err)
			return
		}

		// Verify critical fields
		if unmarshaled.ID != task.ID {
			t.Errorf("ID mismatch: got %v, want %v", unmarshaled.ID, task.ID)
		}
		if len(unmarshaled.History) != len(task.History) {
			t.Errorf("History length mismatch: got %d, want %d", len(unmarshaled.History), len(task.History))
		}
		if len(unmarshaled.Artifacts) != len(task.Artifacts) {
			t.Errorf("Artifacts length mismatch: got %d, want %d", len(unmarshaled.Artifacts), len(task.Artifacts))
		}
		if unmarshaled.Status.State != task.Status.State {
			t.Errorf("Status.State mismatch: got %v, want %v", unmarshaled.Status.State, task.Status.State)
		}
	})
}

// TestStructFieldOmission tests field omission with omitzero tags
func TestStructFieldOmission(t *testing.T) {
	t.Parallel()

	t.Run("AgentExtension with omitzero fields", func(t *testing.T) {
		t.Parallel()

		// Test with all fields set
		fullExt := &AgentExtension{
			URI:         "https://example.com/ext",
			Description: "Test extension",
			Required:    true,
			Params:      map[string]any{"param": "value"},
		}

		data, err := json.Marshal(fullExt)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
		}

		// Should contain all fields
		dataStr := string(data)
		if !contains(dataStr, "description") || !contains(dataStr, "required") || !contains(dataStr, "params") {
			t.Errorf("Full extension should contain all fields: %s", dataStr)
		}

		// Test with minimal fields (only URI required)
		minimalExt := &AgentExtension{
			URI: "https://example.com/ext",
		}

		data, err = json.Marshal(minimalExt)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
		}

		// Should omit zero values
		dataStr = string(data)
		if contains(dataStr, "description") || contains(dataStr, "required") || contains(dataStr, "params") {
			t.Errorf("Minimal extension should omit zero fields: %s", dataStr)
		}
	})

	t.Run("PushNotificationConfig with omitzero fields", func(t *testing.T) {
		t.Parallel()

		// Test with required fields only
		minimalConfig := &PushNotificationConfig{
			URL: "https://example.com/push",
		}

		data, err := json.Marshal(minimalConfig)
		if err != nil {
			t.Errorf("Marshal() error = %v", err)
		}

		// Should omit zero values
		dataStr := string(data)
		if contains(dataStr, "id") || contains(dataStr, "token") || contains(dataStr, "authentication") {
			t.Errorf("Minimal config should omit zero fields: %s", dataStr)
		}
	})
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmarks for performance testing
func BenchmarkNewTextPart(b *testing.B) {
	text := "Hello, World!"
	for b.Loop() {
		_ = NewTextPart(text)
	}
}

func BenchmarkNewDataPart(b *testing.B) {
	data := map[string]any{"key": "value", "number": 42}
	for b.Loop() {
		_ = NewDataPart(data)
	}
}

func BenchmarkNewFilePart(b *testing.B) {
	file := &FileWithBytes{
		Bytes:    "dGVzdA==",
		MIMEType: "text/plain",
		Name:     "test.txt",
	}
	for b.Loop() {
		_ = NewFilePart(file)
	}
}

func BenchmarkUnmarshalPart(b *testing.B) {
	jsonData := []byte(`{"kind": "text", "text": "Hello world"}`)
	for b.Loop() {
		_, _ = UnmarshalPart(jsonData)
	}
}

func BenchmarkUnmarshalFileContent(b *testing.B) {
	jsonData := []byte(`{"uri": "https://example.com/file.txt", "mimeType": "text/plain", "name": "file.txt"}`)
	for b.Loop() {
		_, _ = UnmarshalFileContent(jsonData)
	}
}

func BenchmarkTextPartMarshaling(b *testing.B) {
	part := &TextPart{
		Kind:     TextPartKind,
		Text:     "Hello, World!",
		Metadata: map[string]any{"key": "value"},
	}

	b.Run("Marshal", func(b *testing.B) {
		for b.Loop() {
			_, _ = json.Marshal(part)
		}
	})

	data, _ := json.Marshal(part)
	b.Run("Unmarshal", func(b *testing.B) {
		for b.Loop() {
			var p TextPart
			_ = json.Unmarshal(data, &p)
		}
	})
}

func BenchmarkAPIKeySecuritySchemeMarshaling(b *testing.B) {
	scheme := &APIKeySecurityScheme{
		Type:        APIKeySecuritySchemeType,
		Description: "API Key Authentication",
		In:          InHeader,
		Name:        "Authorization",
	}

	b.Run("Marshal", func(b *testing.B) {
		for b.Loop() {
			_, _ = json.Marshal(scheme)
		}
	})

	data, _ := json.Marshal(scheme)
	b.Run("Unmarshal", func(b *testing.B) {
		for b.Loop() {
			var s APIKeySecurityScheme
			_ = json.Unmarshal(data, &s)
		}
	})
}

func BenchmarkMethodConstants(b *testing.B) {
	methods := []Method{
		MethodMessageSend,
		MethodMessageStream,
		MethodTasksGet,
		MethodTasksList,
		MethodTasksCancel,
		MethodTasksPushNotificationConfigSet,
		MethodTasksPushNotificationConfigGet,
		MethodTasksPushNotificationConfigList,
		MethodTasksPushNotificationConfigDelete,
		MethodTasksResubscribe,
		MethodAgentAuthenticatedExtendedCard,
	}

	for b.Loop() {
		for _, method := range methods {
			_ = string(method)
		}
	}
}

func BenchmarkEnumConstants(b *testing.B) {
	partKinds := []PartKind{TextPartKind, FilePartKind, DataPartKind}
	roles := []Role{RoleAgent, RoleUser}
	ins := []In{InCookie, InHeader, InQuery}
	states := []TaskState{
		TaskStateSubmitted,
		TaskStateWorking,
		TaskStateInputRequired,
		TaskStateCompleted,
		TaskStateCanceled,
		TaskStateFailed,
		TaskStateRejected,
		TaskStateAuthRequired,
		TaskStateUnknown,
	}

	for b.Loop() {
		for _, kind := range partKinds {
			_ = string(kind)
		}
		for _, role := range roles {
			_ = string(role)
		}
		for _, in := range ins {
			_ = string(in)
		}
		for _, state := range states {
			_ = string(state)
		}
	}
}
