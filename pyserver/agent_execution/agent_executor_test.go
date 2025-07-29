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
	"errors"
	"testing"
	"time"

	"github.com/go-a2a/a2a-go"
)

// MockAgentExecutor is a mock implementation of AgentExecutor for testing.
type MockAgentExecutor struct {
	BaseAgentExecutor
	ExecuteFunc func(ctx context.Context, reqCtx *RequestContext, eventQueue EventQueue) error
	CancelFunc  func(ctx context.Context, reqCtx *RequestContext, eventQueue EventQueue) error
}

func (m *MockAgentExecutor) Execute(ctx context.Context, reqCtx *RequestContext, eventQueue EventQueue) error {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, reqCtx, eventQueue)
	}
	return nil
}

func (m *MockAgentExecutor) Cancel(ctx context.Context, reqCtx *RequestContext, eventQueue EventQueue) error {
	if m.CancelFunc != nil {
		return m.CancelFunc(ctx, reqCtx, eventQueue)
	}
	return m.BaseAgentExecutor.Cancel(ctx, reqCtx, eventQueue)
}

func TestBaseAgentExecutor_Cancel(t *testing.T) {
	tests := map[string]struct {
		reqCtx   *RequestContext
		wantErr  bool
		checkFn  func(t *testing.T, event Event)
	}{
		"success: cancels task with default message": {
			reqCtx: &RequestContext{
				TaskID:    "task-123",
				ContextID: "ctx-456",
			},
			wantErr: false,
			checkFn: func(t *testing.T, event Event) {
				statusEvent, ok := event.(*TaskStatusUpdateEvent)
				if !ok {
					t.Fatalf("expected TaskStatusUpdateEvent, got %T", event)
				}
				if statusEvent.TaskID != "task-123" {
					t.Errorf("TaskID = %q, want %q", statusEvent.TaskID, "task-123")
				}
				if statusEvent.Status.State != a2a.TaskStateCanceled {
					t.Errorf("Status.State = %v, want %v", statusEvent.Status.State, a2a.TaskStateCanceled)
				}
				if statusEvent.Status.Message != "Task canceled by user request" {
					t.Errorf("Status.Message = %q, want %q", statusEvent.Status.Message, "Task canceled by user request")
				}
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			executor := &BaseAgentExecutor{}
			eventQueue := NewChannelEventQueue(10)
			defer eventQueue.Close()

			err := executor.Cancel(ctx, tc.reqCtx, eventQueue)
			if (err != nil) != tc.wantErr {
				t.Errorf("Cancel() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.checkFn != nil && !tc.wantErr {
				event, err := eventQueue.Get(ctx)
				if err != nil {
					t.Fatalf("Failed to get event from queue: %v", err)
				}
				tc.checkFn(t, event)
			}
		})
	}
}

func TestMockAgentExecutor_Execute(t *testing.T) {
	tests := map[string]struct {
		setupMock func(*MockAgentExecutor)
		reqCtx    *RequestContext
		wantErr   bool
		checkFn   func(t *testing.T, events []Event)
	}{
		"success: simple execution": {
			setupMock: func(m *MockAgentExecutor) {
				m.ExecuteFunc = func(ctx context.Context, reqCtx *RequestContext, eventQueue EventQueue) error {
					// Send working status
					eventQueue.Put(ctx, CreateWorkingStatusEvent(reqCtx.TaskID, reqCtx.ContextID))
					
					// Send result
					artifact := NewTextArtifact("result", "Hello, World!")
					eventQueue.Put(ctx, NewTaskArtifactUpdateEvent(reqCtx.TaskID, reqCtx.ContextID, artifact))
					
					// Send completion
					eventQueue.Put(ctx, CreateCompletedStatusEvent(reqCtx.TaskID, reqCtx.ContextID))
					
					return nil
				}
			},
			reqCtx: &RequestContext{
				TaskID:    "task-123",
				ContextID: "ctx-456",
				Message:   &a2a.Message{Role: a2a.RoleUser},
			},
			wantErr: false,
			checkFn: func(t *testing.T, events []Event) {
				if len(events) != 3 {
					t.Fatalf("expected 3 events, got %d", len(events))
				}

				// Check working status
				if status, ok := events[0].(*TaskStatusUpdateEvent); ok {
					if status.Status.State != a2a.TaskStateWorking {
						t.Errorf("First event state = %v, want %v", status.Status.State, a2a.TaskStateWorking)
					}
				} else {
					t.Errorf("First event type = %T, want *TaskStatusUpdateEvent", events[0])
				}

				// Check artifact
				if artifact, ok := events[1].(*TaskArtifactUpdateEvent); ok {
					if artifact.Artifact.Name != "result" {
						t.Errorf("Artifact name = %q, want %q", artifact.Artifact.Name, "result")
					}
				} else {
					t.Errorf("Second event type = %T, want *TaskArtifactUpdateEvent", events[1])
				}

				// Check completion
				if status, ok := events[2].(*TaskStatusUpdateEvent); ok {
					if status.Status.State != a2a.TaskStateCompleted {
						t.Errorf("Final event state = %v, want %v", status.Status.State, a2a.TaskStateCompleted)
					}
					if !status.Final {
						t.Error("Final event should have Final = true")
					}
				} else {
					t.Errorf("Final event type = %T, want *TaskStatusUpdateEvent", events[2])
				}
			},
		},
		"error: execution fails": {
			setupMock: func(m *MockAgentExecutor) {
				m.ExecuteFunc = func(ctx context.Context, reqCtx *RequestContext, eventQueue EventQueue) error {
					return errors.New("execution failed")
				}
			},
			reqCtx: &RequestContext{
				TaskID:    "task-123",
				ContextID: "ctx-456",
			},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			executor := &MockAgentExecutor{}
			tc.setupMock(executor)

			eventQueue := NewChannelEventQueue(10)
			defer eventQueue.Close()

			// Run execution in goroutine
			errChan := make(chan error, 1)
			go func() {
				errChan <- executor.Execute(ctx, tc.reqCtx, eventQueue)
			}()

			// Collect events
			var events []Event
			timeout := time.After(1 * time.Second)
			done := false
			
			for !done {
				select {
				case err := <-errChan:
					if (err != nil) != tc.wantErr {
						t.Errorf("Execute() error = %v, wantErr %v", err, tc.wantErr)
					}
					// Continue collecting remaining events
					for {
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
						event, err := eventQueue.Get(ctx)
						cancel()
						if err != nil {
							done = true
							break
						}
						events = append(events, event)
					}
				case <-timeout:
					t.Fatal("Test timed out")
				default:
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
					event, err := eventQueue.Get(ctx)
					cancel()
					if err == nil {
						events = append(events, event)
					}
				}
			}

			if tc.checkFn != nil && !tc.wantErr {
				tc.checkFn(t, events)
			}
		})
	}
}

func TestAgentExecutorInterface(t *testing.T) {
	// Ensure BaseAgentExecutor implements AgentExecutor
	var _ AgentExecutor = (*BaseAgentExecutor)(nil)
	
	// Ensure MockAgentExecutor implements AgentExecutor
	var _ AgentExecutor = (*MockAgentExecutor)(nil)
}