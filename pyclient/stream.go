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

package pyclient

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-json-experiment/json"
)

// Stream represents a Server-Sent Events (SSE) stream from an A2A agent.
type Stream struct {
	client   *Client
	taskID   string
	resp     *http.Response
	reader   *bufio.Reader
	events   chan StreamEvent
	closed   bool
	closeMu  sync.Mutex
	closeErr error
	ctx      context.Context
	cancel   context.CancelFunc
}

// newStream creates a new SSE stream.
func newStream(client *Client, taskID string, resp *http.Response) *Stream {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Stream{
		client: client,
		taskID: taskID,
		resp:   resp,
		reader: bufio.NewReader(resp.Body),
		events: make(chan StreamEvent, client.opts.streamBufferSize),
		ctx:    ctx,
		cancel: cancel,
	}

	// Start reading events
	go s.readLoop()

	return s
}

// TaskID returns the task ID associated with this stream.
func (s *Stream) TaskID() string {
	return s.taskID
}

// Events returns a channel of stream events.
func (s *Stream) Events() <-chan StreamEvent {
	return s.events
}

// Close closes the stream and releases resources.
func (s *Stream) Close() error {
	s.closeMu.Lock()
	defer s.closeMu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	s.cancel()

	if s.resp != nil && s.resp.Body != nil {
		err := s.resp.Body.Close()
		if err != nil && s.closeErr == nil {
			s.closeErr = err
		}
	}

	close(s.events)
	return s.closeErr
}

// readLoop reads SSE events from the response body.
func (s *Stream) readLoop() {
	defer s.Close()

	var event sseEvent

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		// Set read deadline
		if s.client.opts.streamReadTimeout > 0 {
			deadline := time.Now().Add(s.client.opts.streamReadTimeout)
			if conn, ok := s.resp.Body.(interface{ SetReadDeadline(time.Time) error }); ok {
				conn.SetReadDeadline(deadline)
			}
		}

		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				s.sendError(fmt.Errorf("read error: %w", err))
			}
			return
		}

		line = strings.TrimSpace(line)

		// Handle SSE format
		if line == "" {
			// Empty line signals end of event
			if event.data != "" {
				s.processEvent(&event)
				event = sseEvent{} // Reset
			}
			continue
		}

		if strings.HasPrefix(line, "event:") {
			event.eventType = strings.TrimSpace(line[6:])
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(line[5:])
			if event.data != "" {
				event.data += "\n" + data
			} else {
				event.data = data
			}
		} else if strings.HasPrefix(line, "id:") {
			event.id = strings.TrimSpace(line[3:])
		} else if strings.HasPrefix(line, "retry:") {
			// Ignore retry field for now
		} else if strings.HasPrefix(line, ":") {
			// Comment, ignore
		}
	}
}

// processEvent processes a complete SSE event.
func (s *Stream) processEvent(event *sseEvent) {
	// Parse the JSON data based on event type
	var streamEvent StreamEvent

	switch event.eventType {
	case string(a2a.MessageEventKind):
		var msg a2a.Message
		if err := json.Unmarshal([]byte(event.data), &msg); err != nil {
			s.sendError(fmt.Errorf("parse message event: %w", err))
			return
		}
		streamEvent = &MessageEvent{
			Message:    &msg,
			receivedAt: time.Now(),
		}

	case string(a2a.ArtifactUpdateEventKind):
		var update a2a.TaskArtifactUpdateEvent
		if err := json.Unmarshal([]byte(event.data), &update); err != nil {
			s.sendError(fmt.Errorf("parse artifact update event: %w", err))
			return
		}
		streamEvent = &ArtifactUpdateEvent{
			TaskArtifactUpdateEvent: &update,
			receivedAt:              time.Now(),
		}

	case string(a2a.StatusUpdateEventKind):
		var update a2a.TaskStatusUpdateEvent
		if err := json.Unmarshal([]byte(event.data), &update); err != nil {
			s.sendError(fmt.Errorf("parse status update event: %w", err))
			return
		}
		streamEvent = &StatusUpdateEvent{
			TaskStatusUpdateEvent: &update,
			receivedAt:            time.Now(),
		}

	case string(a2a.TaskEventKind):
		var task a2a.Task
		if err := json.Unmarshal([]byte(event.data), &task); err != nil {
			s.sendError(fmt.Errorf("parse task event: %w", err))
			return
		}
		streamEvent = &TaskEvent{
			Task:       &task,
			receivedAt: time.Now(),
		}

	case "error":
		// Server-sent error event
		s.sendError(fmt.Errorf("server error: %s", event.data))
		return

	default:
		// Unknown event type, log but don't fail
		if s.client.opts.enableDebugLogging {
			// In production, this would use a proper logger
			fmt.Printf("Unknown SSE event type: %s\n", event.eventType)
		}
		return
	}

	// Send event to channel
	select {
	case s.events <- streamEvent:
	case <-s.ctx.Done():
		return
	}
}

// sendError sends an error event to the events channel.
func (s *Stream) sendError(err error) {
	select {
	case s.events <- &ErrorEvent{
		Err:        err,
		receivedAt: time.Now(),
	}:
	case <-s.ctx.Done():
	}
}

// sseEvent represents a raw SSE event.
type sseEvent struct {
	id        string
	eventType string
	data      string
}

// StreamIterator provides a Python-like iterator interface for streams.
type StreamIterator struct {
	stream *Stream
	ctx    context.Context
}

// NewStreamIterator creates a new stream iterator.
func NewStreamIterator(stream *Stream) *StreamIterator {
	return &StreamIterator{
		stream: stream,
		ctx:    context.Background(),
	}
}

// WithContext sets the context for the iterator.
func (it *StreamIterator) WithContext(ctx context.Context) *StreamIterator {
	it.ctx = ctx
	return it
}

// Next returns the next event or an error.
func (it *StreamIterator) Next() (StreamEvent, error) {
	select {
	case event, ok := <-it.stream.Events():
		if !ok {
			return nil, ErrStreamClosed
		}
		// Check if it's an error event
		if errEvent, ok := event.(*ErrorEvent); ok {
			return nil, errEvent.Err
		}
		return event, nil
	case <-it.ctx.Done():
		return nil, it.ctx.Err()
	}
}

// ForEach applies a function to each event in the stream.
func (it *StreamIterator) ForEach(fn func(StreamEvent) error) error {
	for {
		event, err := it.Next()
		if err != nil {
			if err == ErrStreamClosed {
				return nil
			}
			return err
		}

		if err := fn(event); err != nil {
			return err
		}
	}
}
