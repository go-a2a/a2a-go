// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-a2a/a2a"
)

// TaskStream handles a streaming task interaction.
type TaskStream struct {
	client    *Client
	taskID    string
	sessionID string
	stream    *StreamConn
	closed    bool
}

// SendTaskStream sends a task and returns a stream of updates.
func (c *Client) SendTaskStream(ctx context.Context, params a2a.TaskSendParams) (*TaskStream, error) {
	// Create a streaming request
	req := a2a.NewSendTaskStreamingRequest(c.nextID(), params)

	// Send the request and get a stream connection
	stream, err := c.transport.Stream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("starting stream: %w", err)
	}

	return &TaskStream{
		client:    c,
		taskID:    params.ID,
		sessionID: params.SessionID,
		stream:    stream,
	}, nil
}

// Resubscribe resubscribes to task updates for an existing task.
func (c *Client) Resubscribe(ctx context.Context, params a2a.TaskQueryParams) (*TaskStream, error) {
	// Create a resubscription request
	req := a2a.NewTaskResubscriptionRequest(c.nextID(), params)

	// Send the request and get a stream connection
	stream, err := c.transport.Stream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("resubscribing: %w", err)
	}

	return &TaskStream{
		client: c,
		taskID: params.ID,
		stream: stream,
	}, nil
}

// Close closes the task stream.
func (ts *TaskStream) Close() error {
	if ts.closed {
		return nil
	}
	ts.closed = true
	return ts.stream.Close()
}

// TaskID returns the ID of the task being streamed.
func (ts *TaskStream) TaskID() string {
	return ts.taskID
}

// SessionID returns the session ID of the task being streamed.
func (ts *TaskStream) SessionID() string {
	return ts.sessionID
}

// NextUpdate waits for the next update from the stream.
// Returns either a status update or an artifact update, and a boolean indicating
// whether this is the final update.
func (ts *TaskStream) NextUpdate(ctx context.Context) (any, bool, error) {
	if ts.closed {
		return nil, false, errors.New("stream closed")
	}

	eventType, data, err := ts.stream.ReadEvent(ctx)
	if err != nil {
		return nil, false, err
	}

	if eventType != EventTypeMessage {
		return nil, false, fmt.Errorf("unexpected event type: %s", eventType)
	}

	// Parse the JSON-RPC response
	var resp struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, false, fmt.Errorf("unmarshaling event data: %w", err)
	}

	// Check if it's a status or artifact update
	var rawResult map[string]json.RawMessage
	if err := json.Unmarshal(resp.Result, &rawResult); err != nil {
		return nil, false, fmt.Errorf("unmarshaling result: %w", err)
	}

	if _, ok := rawResult["status"]; ok {
		// It's a status update
		var event a2a.TaskStatusUpdateEvent
		if err := json.Unmarshal(resp.Result, &event); err != nil {
			return nil, false, fmt.Errorf("unmarshaling status event: %w", err)
		}
		return event, event.Final, nil
	} else if _, ok := rawResult["artifact"]; ok {
		// It's an artifact update
		var event a2a.TaskArtifactUpdateEvent
		if err := json.Unmarshal(resp.Result, &event); err != nil {
			return nil, false, fmt.Errorf("unmarshaling artifact event: %w", err)
		}
		return event, event.Final, nil
	}

	return nil, false, fmt.Errorf("unknown event type in result: %v", rawResult)
}

// WaitForCompletion waits for the task to complete, cancel, or fail.
// Returns the final task state, or an error if the context is canceled or the stream is closed unexpectedly.
func (ts *TaskStream) WaitForCompletion(ctx context.Context) (a2a.TaskState, error) {
	for {
		update, final, err := ts.NextUpdate(ctx)
		if err != nil {
			return a2a.TaskStateUnknown, err
		}

		// Check if it's a status update and if the task is done
		if statusUpdate, ok := update.(a2a.TaskStatusUpdateEvent); ok {
			state := statusUpdate.Status.State
			if state == a2a.TaskStateCompleted || state == a2a.TaskStateCanceled || state == a2a.TaskStateFailed {
				return state, nil
			}
			if final {
				return state, nil
			}
		}

		// If it's the final update and we haven't returned yet, something's wrong
		if final {
			return a2a.TaskStateUnknown, errors.New("stream ended without terminal task state")
		}
	}
}

// WaitForCompletionWithTimeout is like WaitForCompletion but with a timeout.
func (ts *TaskStream) WaitForCompletionWithTimeout(timeout time.Duration) (a2a.TaskState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return ts.WaitForCompletion(ctx)
}

// CreateTextMessage creates a simple text message from the user to the agent.
func CreateTextMessage(text string) a2a.Message {
	return a2a.Message{
		Role: a2a.MessageRoleUser,
		Parts: []a2a.Part{
			{
				Type: a2a.PartTypeText,
				Text: text,
			},
		},
	}
}

// CreateFileMessage creates a message with a file attachment.
func CreateFileMessage(text string, fileContent a2a.FileContent) a2a.Message {
	parts := []a2a.Part{
		{
			Type: a2a.PartTypeText,
			Text: text,
		},
		{
			Type: a2a.PartTypeFile,
			File: &fileContent,
		},
	}

	return a2a.Message{
		Role:  a2a.MessageRoleUser,
		Parts: parts,
	}
}

// CreateDataMessage creates a message with structured data.
func CreateDataMessage(text string, data any) (a2a.Message, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return a2a.Message{}, fmt.Errorf("marshaling data: %w", err)
	}

	parts := []a2a.Part{
		{
			Type: a2a.PartTypeText,
			Text: text,
		},
		{
			Type: a2a.PartTypeData,
			Data: dataBytes,
		},
	}

	return a2a.Message{
		Role:  a2a.MessageRoleUser,
		Parts: parts,
	}, nil
}
