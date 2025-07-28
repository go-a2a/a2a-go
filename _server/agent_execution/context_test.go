// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package agent_execution

import (
	"testing"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/server"
)

func TestNewRequestContext(t *testing.T) {
	params := &a2a.MessageSendParams{
		Message: &a2a.Message{
			Content: "test message",
		},
	}
	callContext := server.NewServerCallContext(nil)
	task := &a2a.Task{
		ID:        "test-task",
		ContextID: "test-context",
		Status:    a2a.TaskStatus{State: a2a.TaskStateSubmitted},
	}

	rc := NewRequestContext(params, "test-task", "test-context", task, callContext)

	if rc.TaskID() != "test-task" {
		t.Errorf("Expected task ID 'test-task', got '%s'", rc.TaskID())
	}

	if rc.ContextID() != "test-context" {
		t.Errorf("Expected context ID 'test-context', got '%s'", rc.ContextID())
	}

	if rc.CurrentTask() != task {
		t.Error("Expected current task to be set correctly")
	}

	if rc.Params() != params {
		t.Error("Expected params to be set correctly")
	}

	if rc.CallContext() != callContext {
		t.Error("Expected call context to be set correctly")
	}
}

func TestRequestContext_GetUserInput(t *testing.T) {
	params := &a2a.MessageSendParams{
		Message: &a2a.Message{
			Content: "test message content",
		},
	}
	callContext := server.NewServerCallContext(nil)
	rc := NewRequestContext(params, "test-task", "test-context", nil, callContext)

	userInput := rc.GetUserInput("")
	if userInput != "test message content" {
		t.Errorf("Expected 'test message content', got '%s'", userInput)
	}

	// Test with nil message
	rc2 := NewRequestContext(nil, "test-task", "test-context", nil, callContext)
	userInput2 := rc2.GetUserInput("")
	if userInput2 != "" {
		t.Errorf("Expected empty string with nil params, got '%s'", userInput2)
	}
}

func TestRequestContext_AttachRelatedTask(t *testing.T) {
	params := &a2a.MessageSendParams{
		Message: &a2a.Message{
			Content: "test message",
		},
	}
	callContext := server.NewServerCallContext(nil)
	rc := NewRequestContext(params, "test-task", "test-context", nil, callContext)

	// Test attaching a valid task
	relatedTask := &a2a.Task{
		ID:        "related-task",
		ContextID: "test-context",
		Status:    a2a.TaskStatus{State: a2a.TaskStateSubmitted},
	}

	err := rc.AttachRelatedTask(relatedTask)
	if err != nil {
		t.Fatalf("Failed to attach related task: %v", err)
	}

	relatedTasks := rc.RelatedTasks()
	if len(relatedTasks) != 1 {
		t.Errorf("Expected 1 related task, got %d", len(relatedTasks))
	}

	if relatedTasks[0] != relatedTask {
		t.Error("Expected related task to be attached correctly")
	}

	// Test attaching nil task
	err = rc.AttachRelatedTask(nil)
	if err == nil {
		t.Error("Expected error when attaching nil task")
	}
}

func TestRequestContext_IDGeneration(t *testing.T) {
	params := &a2a.MessageSendParams{
		Message: &a2a.Message{
			Content: "test message",
		},
	}
	callContext := server.NewServerCallContext(nil)

	// Test with empty task ID
	rc := NewRequestContext(params, "", "test-context", nil, callContext)
	rc.checkOrGenerateTaskID()

	if rc.TaskID() == "" {
		t.Error("Expected task ID to be generated")
	}

	// Test with empty context ID
	rc2 := NewRequestContext(params, "test-task", "", nil, callContext)
	rc2.checkOrGenerateContextID()

	if rc2.ContextID() == "" {
		t.Error("Expected context ID to be generated")
	}
}

func TestRequestContext_ThreadSafety(t *testing.T) {
	params := &a2a.MessageSendParams{
		Message: &a2a.Message{
			Content: "test message",
		},
	}
	callContext := server.NewServerCallContext(nil)
	rc := NewRequestContext(params, "test-task", "test-context", nil, callContext)

	// Test concurrent access
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			relatedTask := &a2a.Task{
				ID:        "related-task",
				ContextID: "test-context",
				Status:    a2a.TaskStatus{State: a2a.TaskStateSubmitted},
			}
			rc.AttachRelatedTask(relatedTask)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			rc.RelatedTasks()
		}
		done <- true
	}()

	<-done
	<-done

	// Verify final state
	relatedTasks := rc.RelatedTasks()
	if len(relatedTasks) != 100 {
		t.Errorf("Expected 100 related tasks, got %d", len(relatedTasks))
	}
}
