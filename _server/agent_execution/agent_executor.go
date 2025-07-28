// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package agent_execution

import (
	"context"
	"errors"
	"time"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/server/event"
)

// AgentExecutor defines the interface for executing agent logic within the A2A server.
// This is equivalent to Python's AgentExecutor abstract base class.
//
// Implementations of AgentExecutor contain the core business logic of an agent,
// executing tasks based on requests and publishing updates to an event queue.
type AgentExecutor interface {
	// Execute runs the agent's main logic.
	// It reads information from the RequestContext and publishes Task or Message events,
	// or TaskStatusUpdateEvent/TaskArtifactUpdateEvent to the EventQueue.
	//
	// This is equivalent to Python's execute(self, context: RequestContext, event_queue: EventQueue) method.
	//
	// Parameters:
	//   ctx: Go context for cancellation and timeouts
	//   requestContext: Contains request information and state
	//   eventQueue: Queue for publishing events during execution
	//
	// Returns:
	//   error: Any error that occurred during execution
	Execute(ctx context.Context, requestContext *RequestContext, eventQueue *event.EventQueue) error

	// Cancel requests the agent to cancel an ongoing task.
	// The agent should attempt to stop the task and publish a TaskStatusUpdateEvent
	// with TaskState.canceled to the EventQueue.
	//
	// This is equivalent to Python's cancel(self, context: RequestContext, event_queue: EventQueue) method.
	//
	// Parameters:
	//   ctx: Go context for cancellation and timeouts
	//   requestContext: Contains request information and state
	//   eventQueue: Queue for publishing cancellation events
	//
	// Returns:
	//   error: Any error that occurred during cancellation
	Cancel(ctx context.Context, requestContext *RequestContext, eventQueue *event.EventQueue) error
}

// BaseAgentExecutor provides a basic implementation of AgentExecutor
// that can be embedded in other implementations.
type BaseAgentExecutor struct{}

var _ AgentExecutor = (*BaseAgentExecutor)(nil)

// Execute provides a default implementation that returns an error.
// Implementations should override this method with their specific logic.
func (bae *BaseAgentExecutor) Execute(ctx context.Context, requestContext *RequestContext, eventQueue *event.EventQueue) error {
	return errors.New("Execute method must be implemented by concrete AgentExecutor")
}

// Cancel provides a default implementation that returns an error.
// Implementations should override this method with their specific logic.
func (bae *BaseAgentExecutor) Cancel(ctx context.Context, requestContext *RequestContext, eventQueue *event.EventQueue) error {
	return errors.New("Cancel method must be implemented by concrete AgentExecutor")
}

// DummyAgentExecutor provides a simple test implementation of AgentExecutor.
// This is equivalent to Python's DummyAgentExecutor used for testing.
type DummyAgentExecutor struct{}

var _ AgentExecutor = (*DummyAgentExecutor)(nil)

// NewDummyAgentExecutor creates a new DummyAgentExecutor instance.
func NewDummyAgentExecutor() *DummyAgentExecutor {
	return &DummyAgentExecutor{}
}

// Execute implements a dummy execution that publishes a status update.
func (dae *DummyAgentExecutor) Execute(ctx context.Context, requestContext *RequestContext, eventQueue *event.EventQueue) error {
	if requestContext == nil {
		return errors.New("request context cannot be nil")
	}
	if eventQueue == nil {
		return errors.New("event queue cannot be nil")
	}

	// Create a task status update event
	statusUpdateEvent := &event.TaskStatusUpdateEvent{
		TaskID:    requestContext.TaskID(),
		ContextID: requestContext.ContextID(),
		Status:    a2a.TaskStatus{State: a2a.TaskStateRunning},
		Final:     false,
	}

	// Publish the status update
	if err := eventQueue.EnqueueEvent(ctx, statusUpdateEvent); err != nil {
		return err
	}

	// Simulate some work
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
		// Continue
	}

	// Mark as completed
	completedEvent := &event.TaskStatusUpdateEvent{
		TaskID:    requestContext.TaskID(),
		ContextID: requestContext.ContextID(),
		Status:    a2a.TaskStatus{State: a2a.TaskStateCompleted},
		Final:     true,
	}

	return eventQueue.EnqueueEvent(ctx, completedEvent)
}

// Cancel implements a dummy cancellation that publishes a canceled status.
func (dae *DummyAgentExecutor) Cancel(ctx context.Context, requestContext *RequestContext, eventQueue *event.EventQueue) error {
	if requestContext == nil {
		return errors.New("request context cannot be nil")
	}
	if eventQueue == nil {
		return errors.New("event queue cannot be nil")
	}

	// Create a canceled task status update event
	canceledEvent := &event.TaskStatusUpdateEvent{
		TaskID:    requestContext.TaskID(),
		ContextID: requestContext.ContextID(),
		Status:    a2a.TaskStatus{State: a2a.TaskStateCanceled},
		Final:     true,
	}

	return eventQueue.EnqueueEvent(ctx, canceledEvent)
}
