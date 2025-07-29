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
	"maps"
	"slices"
	"time"

	"github.com/go-a2a/a2a-go"
)

// RequestContext provides context about an incoming request to an agent executor.
// It contains all the necessary information from the request that the executor
// needs to process the task.
type RequestContext struct {
	// TaskID is the unique identifier for the task being executed.
	TaskID string

	// ContextID is the server-generated id for contextual alignment across interactions.
	ContextID string

	// Message contains the incoming message from the user or previous interaction.
	Message *a2a.Message

	// Task contains the current task state if this is a continuation of an existing task.
	Task *a2a.Task

	// Metadata contains any additional metadata from the request.
	Metadata map[string]any

	// CreatedAt represents when this request context was created.
	CreatedAt time.Time

	// AgentCard contains the agent's configuration and capabilities.
	AgentCard *a2a.AgentCard

	// AuthInfo contains authentication information if present.
	AuthInfo *AuthInfo
}

// AuthInfo contains authentication information for a request.
type AuthInfo struct {
	// UserID is the authenticated user's identifier.
	UserID string

	// Claims contains any additional authentication claims.
	Claims map[string]any

	// Scopes contains the authorized scopes for this request.
	Scopes []string
}

// NewRequestContext creates a new RequestContext with the provided parameters.
func NewRequestContext(taskID, contextID string, message *a2a.Message) *RequestContext {
	return &RequestContext{
		TaskID:    taskID,
		ContextID: contextID,
		Message:   message,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]any),
	}
}

// WithTask adds task information to the request context.
func (rc *RequestContext) WithTask(task *a2a.Task) *RequestContext {
	rc.Task = task
	return rc
}

// WithMetadata adds metadata to the request context.
func (rc *RequestContext) WithMetadata(metadata map[string]any) *RequestContext {
	if rc.Metadata == nil {
		rc.Metadata = make(map[string]any)
	}
	maps.Copy(rc.Metadata, metadata)
	return rc
}

// WithAgentCard adds agent card information to the request context.
func (rc *RequestContext) WithAgentCard(card *a2a.AgentCard) *RequestContext {
	rc.AgentCard = card
	return rc
}

// WithAuthInfo adds authentication information to the request context.
func (rc *RequestContext) WithAuthInfo(authInfo *AuthInfo) *RequestContext {
	rc.AuthInfo = authInfo
	return rc
}

// IsAuthenticated returns true if the request has authentication information.
func (rc *RequestContext) IsAuthenticated() bool {
	return rc.AuthInfo != nil && rc.AuthInfo.UserID != ""
}

// HasScope checks if the request has a specific authorization scope.
func (rc *RequestContext) HasScope(scope string) bool {
	if rc.AuthInfo == nil {
		return false
	}
	return slices.Contains(rc.AuthInfo.Scopes, scope)
}
