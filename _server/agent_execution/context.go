// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package agent_execution

import (
	"errors"
	"fmt"
	"sync"
	"time"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/server"
)

// RequestContext holds information about the current request being processed by the server.
// This is equivalent to Python's RequestContext class.
//
// This includes the incoming message, task and context identifiers, and related tasks.
// RequestContext is thread-safe and can be safely used across multiple goroutines.
type RequestContext struct {
	params       *a2a.MessageSendParams
	taskID       string
	contextID    string
	currentTask  *a2a.Task
	relatedTasks []*a2a.Task
	callContext  *server.ServerCallContext
	mu           sync.RWMutex
}

// NewRequestContext creates a new RequestContext with the provided parameters.
func NewRequestContext(params *a2a.MessageSendParams, taskID, contextID string, currentTask *a2a.Task, callContext *server.ServerCallContext) *RequestContext {
	return &RequestContext{
		params:       params,
		taskID:       taskID,
		contextID:    contextID,
		currentTask:  currentTask,
		relatedTasks: make([]*a2a.Task, 0),
		callContext:  callContext,
	}
}

// Params returns the incoming MessageSendParams request payload.
// This is equivalent to Python's _params attribute.
func (rc *RequestContext) Params() *a2a.MessageSendParams {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.params
}

// TaskID returns the ID of the task.
// This is equivalent to Python's _task_id attribute.
func (rc *RequestContext) TaskID() string {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.taskID
}

// ContextID returns the ID of the conversation context.
// This is equivalent to Python's _context_id attribute.
func (rc *RequestContext) ContextID() string {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.contextID
}

// CurrentTask returns the existing Task object being processed.
// This is equivalent to Python's _current_task attribute.
func (rc *RequestContext) CurrentTask() *a2a.Task {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.currentTask
}

// RelatedTasks returns a copy of the list of other tasks related to the current request.
// This is equivalent to Python's _related_tasks attribute.
func (rc *RequestContext) RelatedTasks() []*a2a.Task {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	// Return a copy to prevent external modification
	tasks := make([]*a2a.Task, len(rc.relatedTasks))
	copy(tasks, rc.relatedTasks)
	return tasks
}

// CallContext returns the server call context associated with this request.
// This is equivalent to Python's _call_context attribute.
func (rc *RequestContext) CallContext() *server.ServerCallContext {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.callContext
}

// GetUserInput extracts text content from the user's message parts.
// This is equivalent to Python's get_user_input method.
//
// Parameters:
//
//	delimiter: The delimiter to join message parts. Defaults to newline.
//
// Returns:
//
//	The text content from the user's message parts.
func (rc *RequestContext) GetUserInput(delimiter string) string {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	if rc.params == nil || rc.params.Message == nil {
		return ""
	}

	if delimiter == "" {
		delimiter = "\n"
	}

	// In the current implementation, Message.Content is a string
	// In a more complex implementation, this might need to extract
	// text from different message parts
	return rc.params.Message.Content
}

// AttachRelatedTask attaches a related task to the context.
// This is useful for scenarios like tool execution.
// This is equivalent to Python's attach_related_task method.
//
// Parameters:
//
//	task: The task to attach as a related task.
//
// Returns:
//
//	error: Any error that occurred during attachment.
func (rc *RequestContext) AttachRelatedTask(task *a2a.Task) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}

	if err := task.Validate(); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.relatedTasks = append(rc.relatedTasks, task)
	return nil
}

// checkOrGenerateTaskID ensures that a task ID is present, generating one if necessary.
// This is equivalent to Python's _check_or_generate_task_id method.
func (rc *RequestContext) checkOrGenerateTaskID() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.taskID == "" {
		// Generate a new task ID
		// In a real implementation, this would use a proper UUID generator
		rc.taskID = generateID()
	}
}

// checkOrGenerateContextID ensures that a context ID is present, generating one if necessary.
// This is equivalent to Python's _check_or_generate_context_id method.
func (rc *RequestContext) checkOrGenerateContextID() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.contextID == "" {
		// Generate a new context ID
		// In a real implementation, this would use a proper UUID generator
		rc.contextID = generateID()
	}
}

// Validate ensures the RequestContext is in a valid state.
func (rc *RequestContext) Validate() error {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	if rc.params == nil {
		return errors.New("request context params cannot be nil")
	}

	if err := rc.params.Validate(); err != nil {
		return fmt.Errorf("request context params are invalid: %w", err)
	}

	if rc.taskID == "" {
		return errors.New("request context task ID cannot be empty")
	}

	if rc.contextID == "" {
		return errors.New("request context context ID cannot be empty")
	}

	if rc.currentTask != nil {
		if err := rc.currentTask.Validate(); err != nil {
			return fmt.Errorf("request context current task is invalid: %w", err)
		}
	}

	for i, task := range rc.relatedTasks {
		if task == nil {
			return fmt.Errorf("request context related task at index %d cannot be nil", i)
		}
		if err := task.Validate(); err != nil {
			return fmt.Errorf("request context related task at index %d is invalid: %w", i, err)
		}
	}

	if rc.callContext == nil {
		return errors.New("request context call context cannot be nil")
	}

	return rc.callContext.Validate()
}

// String returns a string representation of the RequestContext for debugging.
func (rc *RequestContext) String() string {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	return fmt.Sprintf("RequestContext{taskID: %s, contextID: %s, relatedTasks: %d}",
		rc.taskID, rc.contextID, len(rc.relatedTasks))
}

// generateID generates a simple ID for demonstration purposes.
// In a real implementation, this would use a proper UUID generator.
func generateID() string {
	// This is a simplified implementation
	// In practice, you'd use github.com/google/uuid or similar
	return fmt.Sprintf("id-%d", time.Now().UnixNano())
}
