// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package agent_execution

import (
	"context"
	"testing"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/server"
)

func TestSimpleRequestContextBuilder_Build(t *testing.T) {
	builder := NewSimpleRequestContextBuilder()

	params := &a2a.MessageSendParams{
		Message: &a2a.Message{
			Content: "test message",
		},
	}
	callContext := server.NewServerCallContext(nil)
	currentTask := &a2a.Task{
		ID:        "existing-task",
		ContextID: "existing-context",
		Status:    a2a.TaskStatus{State: a2a.TaskStateRunning},
	}

	ctx := context.Background()

	requestContext, err := builder.Build(ctx, params, "test-task", "test-context", currentTask, callContext)
	if err != nil {
		t.Fatalf("Failed to build request context: %v", err)
	}

	if requestContext.TaskID() != "test-task" {
		t.Errorf("Expected task ID 'test-task', got '%s'", requestContext.TaskID())
	}

	if requestContext.ContextID() != "test-context" {
		t.Errorf("Expected context ID 'test-context', got '%s'", requestContext.ContextID())
	}

	if requestContext.CurrentTask() != currentTask {
		t.Error("Expected current task to be set correctly")
	}

	if requestContext.Params() != params {
		t.Error("Expected params to be set correctly")
	}

	if requestContext.CallContext() != callContext {
		t.Error("Expected call context to be set correctly")
	}
}

func TestSimpleRequestContextBuilder_BuildWithEmptyIDs(t *testing.T) {
	builder := NewSimpleRequestContextBuilder()

	params := &a2a.MessageSendParams{
		Message: &a2a.Message{
			Content: "test message",
		},
	}
	callContext := server.NewServerCallContext(nil)

	ctx := context.Background()

	// Build with empty task ID and context ID
	requestContext, err := builder.Build(ctx, params, "", "", nil, callContext)
	if err != nil {
		t.Fatalf("Failed to build request context: %v", err)
	}

	// IDs should be generated
	if requestContext.TaskID() == "" {
		t.Error("Expected task ID to be generated")
	}

	if requestContext.ContextID() == "" {
		t.Error("Expected context ID to be generated")
	}
}

func TestSimpleRequestContextBuilder_BuildWithRelatedTasks(t *testing.T) {
	builder := NewSimpleRequestContextBuilderWithRelatedTasks(true)

	params := &a2a.MessageSendParams{
		Message: &a2a.Message{
			Content: "test message",
		},
	}
	callContext := server.NewServerCallContext(nil)

	ctx := context.Background()

	requestContext, err := builder.Build(ctx, params, "test-task", "test-context", nil, callContext)
	if err != nil {
		t.Fatalf("Failed to build request context: %v", err)
	}

	// Check that populate related tasks option is set
	if !builder.PopulateRelatedTasks() {
		t.Error("Expected populate related tasks to be true")
	}

	// In current implementation, no related tasks are actually populated
	// but the functionality is there for future implementation
	relatedTasks := requestContext.RelatedTasks()
	if len(relatedTasks) != 0 {
		t.Errorf("Expected 0 related tasks (placeholder implementation), got %d", len(relatedTasks))
	}
}

func TestSimpleRequestContextBuilder_ErrorHandling(t *testing.T) {
	builder := NewSimpleRequestContextBuilder()

	ctx := context.Background()
	callContext := server.NewServerCallContext(nil)

	// Test with nil context
	_, err := builder.Build(nil, &a2a.MessageSendParams{}, "test-task", "test-context", nil, callContext)
	if err == nil {
		t.Error("Expected error with nil context")
	}

	// Test with nil params
	_, err = builder.Build(ctx, nil, "test-task", "test-context", nil, callContext)
	if err == nil {
		t.Error("Expected error with nil params")
	}

	// Test with invalid params
	invalidParams := &a2a.MessageSendParams{
		Message: &a2a.Message{
			Content: "", // Empty content is invalid
		},
	}
	_, err = builder.Build(ctx, invalidParams, "test-task", "test-context", nil, callContext)
	if err == nil {
		t.Error("Expected error with invalid params")
	}

	// Test with nil call context
	validParams := &a2a.MessageSendParams{
		Message: &a2a.Message{
			Content: "test message",
		},
	}
	_, err = builder.Build(ctx, validParams, "test-task", "test-context", nil, nil)
	if err == nil {
		t.Error("Expected error with nil call context")
	}
}

func TestSimpleRequestContextBuilder_Configuration(t *testing.T) {
	builder := NewSimpleRequestContextBuilder()

	// Test default configuration
	if builder.PopulateRelatedTasks() {
		t.Error("Expected populate related tasks to be false by default")
	}

	// Test setting configuration
	builder.SetPopulateRelatedTasks(true)
	if !builder.PopulateRelatedTasks() {
		t.Error("Expected populate related tasks to be true after setting")
	}

	// Test validation
	if err := builder.Validate(); err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}

	// Test string representation
	str := builder.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
}
