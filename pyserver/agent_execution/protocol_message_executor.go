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
	"time"

	"github.com/go-a2a/a2a-go"
)

// ProtocolMessageExecutor handles the execution of A2A protocol messages.
// It manages the lifecycle of task execution, including validation,
// event processing, and result formatting.
type ProtocolMessageExecutor struct {
	executor AgentExecutor
	options  *ExecutorOptions
}

// ExecutorOptions contains configuration options for the ProtocolMessageExecutor.
type ExecutorOptions struct {
	// MaxEventQueueSize is the maximum number of events that can be queued.
	MaxEventQueueSize int

	// ValidateMessages enables message validation before processing.
	ValidateMessages bool

	// AllowConcurrentExecution allows multiple tasks to execute concurrently.
	AllowConcurrentExecution bool

	// DefaultTimeout is the default timeout for task execution.
	DefaultTimeout int
}

// DefaultExecutorOptions returns the default executor options.
func DefaultExecutorOptions() *ExecutorOptions {
	return &ExecutorOptions{
		MaxEventQueueSize:        1000,
		ValidateMessages:         true,
		AllowConcurrentExecution: false,
		DefaultTimeout:           300, // 5 minutes
	}
}

// NewProtocolMessageExecutor creates a new ProtocolMessageExecutor with the given agent executor.
func NewProtocolMessageExecutor(executor AgentExecutor, options *ExecutorOptions) *ProtocolMessageExecutor {
	if options == nil {
		options = DefaultExecutorOptions()
	}
	return &ProtocolMessageExecutor{
		executor: executor,
		options:  options,
	}
}

// ExecuteMessage processes an incoming message and executes the associated task.
func (p *ProtocolMessageExecutor) ExecuteMessage(ctx context.Context, params *a2a.MessageSendParams) (*a2a.Task, error) {
	// Validate the message if enabled
	if p.options.ValidateMessages {
		if err := p.validateMessage(params.Message); err != nil {
			return nil, fmt.Errorf("message validation failed: %w", err)
		}
	}

	// Create request context from the message
	taskID := params.Message.TaskID
	if taskID == "" {
		taskID = params.Message.MessageID
	}
	reqCtx := NewRequestContext(
		taskID,
		params.Message.ContextID,
		params.Message,
	).WithMetadata(params.Metadata)

	// Create event queue for collecting execution results
	eventQueue := NewChannelEventQueue(p.options.MaxEventQueueSize)
	defer eventQueue.Close()

	// Start the executor in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- p.executor.Execute(ctx, reqCtx, eventQueue)
	}()

	// Collect events and build the task response
	task := &a2a.Task{
		ID:        reqCtx.TaskID,
		ContextID: reqCtx.ContextID,
		Kind:      a2a.TaskEventKind,
		Status: a2a.TaskStatus{
			State: a2a.TaskStateSubmitted,
		},
		Artifacts: make([]*a2a.Artifact, 0),
		Metadata:  params.Metadata,
	}

	// Process events until execution completes
	executionComplete := false
	for !executionComplete {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errChan:
			executionComplete = true
			if err != nil {
				task.Status.State = a2a.TaskStateFailed
				task.Status.Message = &a2a.Message{
					Role:  a2a.RoleAgent,
					Parts: []a2a.Part{a2a.NewTextPart(err.Error())},
				}
			}
			// Continue processing remaining events
		default:
			// Try to get an event with a short timeout
			eventCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
			event, err := eventQueue.Get(eventCtx)
			cancel()

			if err == nil {
				if err := p.processEvent(event, task); err != nil {
					return nil, fmt.Errorf("failed to process event: %w", err)
				}

				// Check if we've reached a terminal state
				if p.isTerminalState(task.Status.State) && executionComplete {
					return task, nil
				}
			} else if executionComplete {
				// No more events and execution is complete
				return task, nil
			}
		}
	}

	return task, nil
}

// ExecuteStreamingMessage processes a message and returns a stream of events.
func (p *ProtocolMessageExecutor) ExecuteStreamingMessage(ctx context.Context, params *a2a.MessageSendParams) (<-chan a2a.SendStreamingMessageResponse, error) {
	// Validate the message if enabled
	if p.options.ValidateMessages {
		if err := p.validateMessage(params.Message); err != nil {
			return nil, fmt.Errorf("message validation failed: %w", err)
		}
	}

	// Create request context
	taskID := params.Message.TaskID
	if taskID == "" {
		taskID = params.Message.MessageID
	}
	reqCtx := NewRequestContext(
		taskID,
		params.Message.ContextID,
		params.Message,
	).WithMetadata(params.Metadata)

	// Create event queue
	eventQueue := NewChannelEventQueue(p.options.MaxEventQueueSize)

	// Create response channel
	responseChan := make(chan a2a.SendStreamingMessageResponse, 100)

	// Start executor and event processor
	go func() {
		defer close(responseChan)
		defer eventQueue.Close()

		// Start the executor
		errChan := make(chan error, 1)
		go func() {
			errChan <- p.executor.Execute(ctx, reqCtx, eventQueue)
		}()

		// Process events and send responses
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-errChan:
				if err != nil {
					// Send error as final status update
					responseChan <- &a2a.TaskStatusUpdateEvent{
						TaskID:    reqCtx.TaskID,
						ContextID: reqCtx.ContextID,
						Status: a2a.TaskStatus{
							State: a2a.TaskStateFailed,
							Message: &a2a.Message{
								Role:  a2a.RoleAgent,
								Parts: []a2a.Part{a2a.NewTextPart(err.Error())},
							},
						},
						Final: true,
						Kind:  a2a.StatusUpdateEventKind,
					}
				}
				return
			default:
				event, err := eventQueue.Get(ctx)
				if err != nil {
					continue
				}

				// Convert internal events to protocol events
				if response := p.eventToStreamingResponse(event); response != nil {
					responseChan <- response
				}
			}
		}
	}()

	return responseChan, nil
}

// CancelTask attempts to cancel a running task.
func (p *ProtocolMessageExecutor) CancelTask(ctx context.Context, taskID string) (*a2a.Task, error) {
	// Create a minimal request context for cancellation
	reqCtx := &RequestContext{
		TaskID: taskID,
	}

	// Create event queue for cancellation response
	eventQueue := NewChannelEventQueue(10)
	defer eventQueue.Close()

	// Execute cancellation
	if err := p.executor.Cancel(ctx, reqCtx, eventQueue); err != nil {
		return nil, fmt.Errorf("failed to cancel task: %w", err)
	}

	// Build cancellation response
	task := &a2a.Task{
		ID:   taskID,
		Kind: a2a.TaskEventKind,
		Status: a2a.TaskStatus{
			State: a2a.TaskStateCanceled,
		},
	}

	return task, nil
}

// validateMessage validates an incoming message.
func (p *ProtocolMessageExecutor) validateMessage(message *a2a.Message) error {
	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}

	if message.Role == "" {
		return fmt.Errorf("message role is required")
	}

	if len(message.Parts) == 0 {
		return fmt.Errorf("message must have at least one part")
	}

	return nil
}

// processEvent processes an event and updates the task state.
func (p *ProtocolMessageExecutor) processEvent(event Event, task *a2a.Task) error {
	switch e := event.(type) {
	case *TaskStatusUpdateEvent:
		task.Status.State = e.Status.State
		if e.Status.Message != "" {
			task.Status.Message = &a2a.Message{
				Role:  a2a.RoleAgent,
				Parts: []a2a.Part{a2a.NewTextPart(e.Status.Message)},
			}
		}

	case *TaskArtifactUpdateEvent:
		if e.Append {
			// Find existing artifact and append
			for _, artifact := range task.Artifacts {
				if artifact.ArtifactID == e.Artifact.ArtifactID {
					artifact.Parts = append(artifact.Parts, e.Artifact.Parts...)
					return nil
				}
			}
		}
		// Add new artifact
		task.Artifacts = append(task.Artifacts, e.Artifact)

	case *MessageEvent:
		// Store message in task history
		if task.History == nil {
			task.History = make([]*a2a.Message, 0)
		}
		task.History = append(task.History, e.Message)
	}

	return nil
}

// eventToStreamingResponse converts internal events to streaming response types.
func (p *ProtocolMessageExecutor) eventToStreamingResponse(event Event) a2a.SendStreamingMessageResponse {
	switch e := event.(type) {
	case *TaskStatusUpdateEvent:
		return &a2a.TaskStatusUpdateEvent{
			TaskID:    e.TaskID,
			ContextID: e.ContextID,
			Status: a2a.TaskStatus{
				State: e.Status.State,
				Message: &a2a.Message{
					Role:  a2a.RoleAgent,
					Parts: []a2a.Part{a2a.NewTextPart(e.Status.Message)},
				},
			},
			Final: e.Final,
			Kind:  a2a.StatusUpdateEventKind,
		}

	case *TaskArtifactUpdateEvent:
		return &a2a.TaskArtifactUpdateEvent{
			TaskID:    e.TaskID,
			ContextID: e.ContextID,
			Artifact:  e.Artifact,
			Append:    e.Append,
			Kind:      a2a.ArtifactUpdateEventKind,
		}

	case *MessageEvent:
		return e.Message

	default:
		return nil
	}
}

// isTerminalState checks if a task state is terminal.
func (p *ProtocolMessageExecutor) isTerminalState(state a2a.TaskState) bool {
	return state == a2a.TaskStateCompleted ||
		state == a2a.TaskStateFailed ||
		state == a2a.TaskStateCanceled ||
		state == a2a.TaskStateRejected ||
		state == a2a.TaskStateUnknown
}

