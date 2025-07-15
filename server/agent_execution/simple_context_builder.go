// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package agent_execution

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/server"
)

// SimpleRequestContextBuilder provides a concrete implementation of RequestContextBuilder.
// This is equivalent to Python's SimpleRequestContextBuilder class.
//
// This is the default implementation used to build RequestContext objects.
// It can be configured to populate related tasks.
type SimpleRequestContextBuilder struct {
	// populateRelatedTasks determines whether to populate related tasks during building.
	populateRelatedTasks bool
}

var _ RequestContextBuilder = (*SimpleRequestContextBuilder)(nil)

// NewSimpleRequestContextBuilder creates a new SimpleRequestContextBuilder instance.
func NewSimpleRequestContextBuilder() *SimpleRequestContextBuilder {
	return &SimpleRequestContextBuilder{
		populateRelatedTasks: false,
	}
}

// NewSimpleRequestContextBuilderWithRelatedTasks creates a new SimpleRequestContextBuilder
// with the option to populate related tasks.
func NewSimpleRequestContextBuilderWithRelatedTasks(populateRelatedTasks bool) *SimpleRequestContextBuilder {
	return &SimpleRequestContextBuilder{
		populateRelatedTasks: populateRelatedTasks,
	}
}

// Build creates a RequestContext from the provided parameters.
// This is the concrete implementation of the RequestContextBuilder interface.
func (srcb *SimpleRequestContextBuilder) Build(ctx context.Context, params *a2a.MessageSendParams, taskID, contextID string, currentTask *a2a.Task, callContext *server.ServerCallContext) (*RequestContext, error) {
	// Validate required parameters
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}
	if params == nil {
		return nil, errors.New("message send params cannot be nil")
	}
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid message send params: %w", err)
	}
	if callContext == nil {
		return nil, errors.New("call context cannot be nil")
	}

	// Create the request context
	requestContext := NewRequestContext(
		params,
		taskID,
		contextID,
		currentTask,
		callContext,
	)

	// Ensure task ID and context ID are present
	requestContext.checkOrGenerateTaskID()
	requestContext.checkOrGenerateContextID()

	// Populate related tasks if configured to do so
	if srcb.populateRelatedTasks {
		if err := srcb.populateRelatedTasksForContext(ctx, requestContext); err != nil {
			return nil, fmt.Errorf("failed to populate related tasks: %w", err)
		}
	}

	// Validate the built context
	if err := requestContext.Validate(); err != nil {
		return nil, fmt.Errorf("built request context is invalid: %w", err)
	}

	return requestContext, nil
}

// populateRelatedTasksForContext populates related tasks for the request context.
// This is a placeholder implementation that would typically query a database
// or task store to find related tasks.
func (srcb *SimpleRequestContextBuilder) populateRelatedTasksForContext(ctx context.Context, requestContext *RequestContext) error {
	// In a real implementation, this would query a database or task store
	// to find tasks related to the current context ID or task ID.
	// For now, this is a no-op placeholder.

	// Example of what this might look like:
	// relatedTasks, err := taskStore.FindRelatedTasks(ctx, requestContext.ContextID())
	// if err != nil {
	//     return fmt.Errorf("failed to find related tasks: %w", err)
	// }
	//
	// for _, task := range relatedTasks {
	//     if err := requestContext.AttachRelatedTask(task); err != nil {
	//         return fmt.Errorf("failed to attach related task: %w", err)
	//     }
	// }

	return nil
}

// SetPopulateRelatedTasks sets whether to populate related tasks during building.
func (srcb *SimpleRequestContextBuilder) SetPopulateRelatedTasks(populate bool) {
	srcb.populateRelatedTasks = populate
}

// PopulateRelatedTasks returns whether related tasks are populated during building.
func (srcb *SimpleRequestContextBuilder) PopulateRelatedTasks() bool {
	return srcb.populateRelatedTasks
}

// Validate ensures the SimpleRequestContextBuilder is in a valid state.
func (srcb *SimpleRequestContextBuilder) Validate() error {
	// SimpleRequestContextBuilder has no invalid states currently
	return nil
}

// String returns a string representation of the SimpleRequestContextBuilder for debugging.
func (srcb *SimpleRequestContextBuilder) String() string {
	return fmt.Sprintf("SimpleRequestContextBuilder{populateRelatedTasks: %t}",
		srcb.populateRelatedTasks)
}
