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

package client

import (
	"context"
	"fmt"
	"io"

	"github.com/go-json-experiment/json"

	"github.com/go-a2a/a2a-go/client/internal/sse"
)

// StreamingClient provides streaming capabilities for the A2A client.
type StreamingClient struct {
	decoder *sse.Decoder
	closer  io.Closer
}

// NewStreamingClient creates a new streaming client.
func NewStreamingClient(reader io.Reader, closer io.Closer) *StreamingClient {
	return &StreamingClient{
		decoder: sse.NewDecoder(reader),
		closer:  closer,
	}
}

// Stream returns a channel of streaming responses.
func (sc *StreamingClient) Stream(ctx context.Context) (<-chan *SendStreamingMessageResponse, <-chan error) {
	responses := make(chan *SendStreamingMessageResponse, 10)
	errors := make(chan error, 1)

	go func() {
		defer close(responses)
		defer close(errors)
		defer sc.closer.Close()

		for {
			select {
			case <-ctx.Done():
				errors <- ctx.Err()
				return
			default:
				event, err := sc.decoder.Decode()
				if err != nil {
					if err == io.EOF {
						return
					}
					errors <- fmt.Errorf("failed to decode SSE event: %w", err)
					return
				}

				if event.Data == "" {
					continue
				}

				var response SendStreamingMessageResponse
				if err := json.Unmarshal([]byte(event.Data), &response); err != nil {
					errors <- fmt.Errorf("failed to unmarshal streaming response: %w", err)
					return
				}

				select {
				case responses <- &response:
				case <-ctx.Done():
					errors <- ctx.Err()
					return
				}
			}
		}
	}()

	return responses, errors
}

// StreamEvents returns a channel of raw SSE events.
func (sc *StreamingClient) StreamEvents(ctx context.Context) (<-chan *sse.Event, <-chan error) {
	events := make(chan *sse.Event, 10)
	errors := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errors)
		defer sc.closer.Close()

		for {
			select {
			case <-ctx.Done():
				errors <- ctx.Err()
				return
			default:
				event, err := sc.decoder.Decode()
				if err != nil {
					if err == io.EOF {
						return
					}
					errors <- fmt.Errorf("failed to decode SSE event: %w", err)
					return
				}

				select {
				case events <- event:
				case <-ctx.Done():
					errors <- ctx.Err()
					return
				}
			}
		}
	}()

	return events, errors
}

// Close closes the streaming client.
func (sc *StreamingClient) Close() error {
	return sc.closer.Close()
}

// StreamingResponse represents a union type for streaming responses.
type StreamingResponse struct {
	Task    *Task           `json:"task,omitempty"`
	Message *Message        `json:"message,omitempty"`
	Event   *StreamingEvent `json:"event,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// Task represents a task (placeholder, should import from parent package).
type Task struct {
	ID        string         `json:"id"`
	ContextID string         `json:"context_id"`
	Status    string         `json:"status"`
	History   []any          `json:"history,omitempty"`
	Artifacts []any          `json:"artifacts,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Message represents a message (placeholder, should import from parent package).
type Message struct {
	ContextID string         `json:"context_id,omitempty"`
	Content   string         `json:"content"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// StreamingEventHandler handles different types of streaming events.
type StreamingEventHandler interface {
	HandleTaskStatusUpdate(event *TaskStatusUpdateEvent) error
	HandleTaskArtifactUpdate(event *TaskArtifactUpdateEvent) error
	HandleMessage(message *Message) error
	HandleTask(task *Task) error
	HandleError(err *RPCError) error
}

// DefaultStreamingEventHandler provides a default implementation.
type DefaultStreamingEventHandler struct{}

var _ StreamingEventHandler = (*DefaultStreamingEventHandler)(nil)

// HandleTaskStatusUpdate handles task status update events.
func (h *DefaultStreamingEventHandler) HandleTaskStatusUpdate(event *TaskStatusUpdateEvent) error {
	// Default implementation does nothing
	return nil
}

// HandleTaskArtifactUpdate handles task artifact update events.
func (h *DefaultStreamingEventHandler) HandleTaskArtifactUpdate(event *TaskArtifactUpdateEvent) error {
	// Default implementation does nothing
	return nil
}

// HandleMessage handles message events.
func (h *DefaultStreamingEventHandler) HandleMessage(message *Message) error {
	// Default implementation does nothing
	return nil
}

// HandleTask handles task events.
func (h *DefaultStreamingEventHandler) HandleTask(task *Task) error {
	// Default implementation does nothing
	return nil
}

// HandleError handles error events.
func (h *DefaultStreamingEventHandler) HandleError(err *RPCError) error {
	// Default implementation does nothing
	return nil
}

// StreamWithHandler processes streaming responses using a handler.
func (sc *StreamingClient) StreamWithHandler(ctx context.Context, handler StreamingEventHandler) error {
	responses, errors := sc.Stream(ctx)

	for {
		select {
		case response, ok := <-responses:
			if !ok {
				return nil
			}

			if response.Error != nil {
				if err := handler.HandleError(response.Error); err != nil {
					return err
				}
				continue
			}

			// Try to determine the response type and handle accordingly
			if response.Result != nil {
				// This is a simplified type detection
				// In a real implementation, you'd have better type detection
				switch v := response.Result.(type) {
				case map[string]any:
					// Try to detect if it's a task or message
					if _, hasTaskID := v["id"]; hasTaskID {
						if _, hasStatus := v["status"]; hasStatus {
							// Looks like a task
							task := &Task{}
							if data, err := json.Marshal(v); err == nil {
								if err := json.Unmarshal(data, task); err == nil {
									if err := handler.HandleTask(task); err != nil {
										return err
									}
								}
							}
						}
					} else if _, hasContent := v["content"]; hasContent {
						// Looks like a message
						message := &Message{}
						if data, err := json.Marshal(v); err == nil {
							if err := json.Unmarshal(data, message); err == nil {
								if err := handler.HandleMessage(message); err != nil {
									return err
								}
							}
						}
					}
				}
			}

		case err, ok := <-errors:
			if !ok {
				return nil
			}
			return err

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
