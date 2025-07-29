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
	"context"
	"fmt"

	"github.com/go-a2a/a2a-go"
)

// ContextBuilder defines the interface for building RequestContext from various sources.
type ContextBuilder interface {
	// Build creates a RequestContext from the provided parameters.
	Build(ctx context.Context, params *ContextBuildParams) (*RequestContext, error)
}

// ContextBuildParams contains parameters for building a RequestContext.
type ContextBuildParams struct {
	// TaskID is the unique identifier for the task.
	TaskID string

	// ContextID is the server-generated id for contextual alignment.
	ContextID string

	// Message contains the incoming message.
	Message *a2a.Message

	// Task contains existing task information if available.
	Task *a2a.Task

	// AgentCard contains the agent's configuration.
	AgentCard *a2a.AgentCard

	// AuthInfo contains authentication information.
	AuthInfo *AuthInfo

	// Metadata contains additional metadata.
	Metadata map[string]any
}

// SimpleContextBuilder provides a basic implementation of ContextBuilder.
type SimpleContextBuilder struct {
	// defaultAgentCard is used when no agent card is provided in params.
	defaultAgentCard *a2a.AgentCard

	// requireAuth specifies whether authentication is required.
	requireAuth bool
}

// NewSimpleContextBuilder creates a new SimpleContextBuilder.
func NewSimpleContextBuilder() *SimpleContextBuilder {
	return &SimpleContextBuilder{}
}

// WithDefaultAgentCard sets the default agent card.
func (b *SimpleContextBuilder) WithDefaultAgentCard(card *a2a.AgentCard) *SimpleContextBuilder {
	b.defaultAgentCard = card
	return b
}

// WithRequireAuth sets whether authentication is required.
func (b *SimpleContextBuilder) WithRequireAuth(require bool) *SimpleContextBuilder {
	b.requireAuth = require
	return b
}

// Build creates a RequestContext from the provided parameters.
func (b *SimpleContextBuilder) Build(ctx context.Context, params *ContextBuildParams) (*RequestContext, error) {
	if params == nil {
		return nil, fmt.Errorf("context build params cannot be nil")
	}

	// Validate required fields
	if params.TaskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	if params.Message == nil {
		return nil, fmt.Errorf("message is required")
	}

	// Check authentication if required
	if b.requireAuth && (params.AuthInfo == nil || params.AuthInfo.UserID == "") {
		return nil, fmt.Errorf("authentication is required")
	}

	// Create the request context
	reqCtx := NewRequestContext(params.TaskID, params.ContextID, params.Message)

	// Add task if provided
	if params.Task != nil {
		reqCtx.WithTask(params.Task)
	}

	// Add agent card
	if params.AgentCard != nil {
		reqCtx.WithAgentCard(params.AgentCard)
	} else if b.defaultAgentCard != nil {
		reqCtx.WithAgentCard(b.defaultAgentCard)
	}

	// Add authentication info
	if params.AuthInfo != nil {
		reqCtx.WithAuthInfo(params.AuthInfo)
	}

	// Add metadata
	if params.Metadata != nil {
		reqCtx.WithMetadata(params.Metadata)
	}

	return reqCtx, nil
}

// EnhancedContextBuilder provides advanced context building with additional features.
type EnhancedContextBuilder struct {
	*SimpleContextBuilder

	// messageValidators are used to validate incoming messages.
	messageValidators []MessageValidator

	// contextEnrichers add additional information to the context.
	contextEnrichers []ContextEnricher
}

// MessageValidator validates messages during context building.
type MessageValidator interface {
	Validate(message *a2a.Message) error
}

// ContextEnricher enriches the context with additional information.
type ContextEnricher interface {
	Enrich(ctx context.Context, reqCtx *RequestContext) error
}

// NewEnhancedContextBuilder creates a new EnhancedContextBuilder.
func NewEnhancedContextBuilder() *EnhancedContextBuilder {
	return &EnhancedContextBuilder{
		SimpleContextBuilder: NewSimpleContextBuilder(),
		messageValidators:    make([]MessageValidator, 0),
		contextEnrichers:     make([]ContextEnricher, 0),
	}
}

// WithMessageValidator adds a message validator.
func (b *EnhancedContextBuilder) WithMessageValidator(validator MessageValidator) *EnhancedContextBuilder {
	b.messageValidators = append(b.messageValidators, validator)
	return b
}

// WithContextEnricher adds a context enricher.
func (b *EnhancedContextBuilder) WithContextEnricher(enricher ContextEnricher) *EnhancedContextBuilder {
	b.contextEnrichers = append(b.contextEnrichers, enricher)
	return b
}

// Build creates a RequestContext with validation and enrichment.
func (b *EnhancedContextBuilder) Build(ctx context.Context, params *ContextBuildParams) (*RequestContext, error) {
	// Validate the message
	if params.Message != nil {
		for _, validator := range b.messageValidators {
			if err := validator.Validate(params.Message); err != nil {
				return nil, fmt.Errorf("message validation failed: %w", err)
			}
		}
	}

	// Build the base context
	reqCtx, err := b.SimpleContextBuilder.Build(ctx, params)
	if err != nil {
		return nil, err
	}

	// Enrich the context
	for _, enricher := range b.contextEnrichers {
		if err := enricher.Enrich(ctx, reqCtx); err != nil {
			return nil, fmt.Errorf("context enrichment failed: %w", err)
		}
	}

	return reqCtx, nil
}

