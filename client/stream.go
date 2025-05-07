// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/go-a2a/a2a"
)

// StreamConn represents a connection to a streaming API.
type StreamConn struct {
	reader   *bufio.Reader
	closer   io.Closer
	mu       sync.Mutex
	closed   bool
	lastErr  error
	doneChan chan struct{}
}

// NewStreamConn creates a new StreamConn from an io.ReadCloser.
func NewStreamConn(rc io.ReadCloser) *StreamConn {
	return &StreamConn{
		reader:   bufio.NewReader(rc),
		closer:   rc,
		doneChan: make(chan struct{}),
	}
}

// Close closes the stream connection.
func (s *StreamConn) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	close(s.doneChan)
	return s.closer.Close()
}

// Done returns a channel that's closed when the stream is closed.
func (s *StreamConn) Done() <-chan struct{} {
	return s.doneChan
}

// Err returns the last error that occurred while reading from the stream.
func (s *StreamConn) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastErr
}

// setError sets the last error encountered.
func (s *StreamConn) setError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastErr = err
	if err != nil && !s.closed {
		s.closed = true
		close(s.doneChan)
	}
}

// ReadEvent reads a single SSE event from the stream.
func (s *StreamConn) ReadEvent(ctx context.Context) (EventType, []byte, error) {
	s.mu.Lock()
	if s.closed {
		err := s.lastErr
		if err == nil {
			err = errors.New("stream closed")
		}
		s.mu.Unlock()
		return "", nil, err
	}
	s.mu.Unlock()

	var eventType EventType
	var data []byte

	for {
		select {
		case <-ctx.Done():
			return "", nil, ctx.Err()
		default:
		}

		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.setError(err)
				return "", nil, err
			}
			s.setError(err)
			return "", nil, fmt.Errorf("reading line: %w", err)
		}

		line = strings.TrimRight(line, "\n")
		if line == "" {
			// Empty line indicates end of event
			if eventType != "" && len(data) > 0 {
				return eventType, data, nil
			}
			continue
		}

		if strings.HasPrefix(line, "event:") {
			eventType = EventType(strings.TrimSpace(line[6:]))
		} else if strings.HasPrefix(line, "data:") {
			data = append(data, []byte(strings.TrimSpace(line[5:]))...)
		}
	}
}

// ReadTaskStatusUpdateEvent reads a TaskStatusUpdateEvent from the stream.
func (s *StreamConn) ReadTaskStatusUpdateEvent(ctx context.Context) (*a2a.TaskStatusUpdateEvent, bool, error) {
	eventType, data, err := s.ReadEvent(ctx)
	if err != nil {
		return nil, false, err
	}

	if eventType != EventTypeMessage {
		return nil, false, fmt.Errorf("unexpected event type: %s", eventType)
	}

	var resp struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, false, fmt.Errorf("unmarshaling event data: %w", err)
	}

	var event a2a.TaskStatusUpdateEvent
	if err := json.Unmarshal(resp.Result, &event); err != nil {
		// Try to unmarshal as TaskArtifactUpdateEvent instead
		var artifactEvent a2a.TaskArtifactUpdateEvent
		if err := json.Unmarshal(resp.Result, &artifactEvent); err != nil {
			return nil, false, fmt.Errorf("unmarshaling event result: %w", err)
		}
		// It's an artifact event, not a status update event
		return nil, false, nil
	}

	return &event, true, nil
}

// ReadTaskArtifactUpdateEvent reads a TaskArtifactUpdateEvent from the stream.
func (s *StreamConn) ReadTaskArtifactUpdateEvent(ctx context.Context) (*a2a.TaskArtifactUpdateEvent, bool, error) {
	eventType, data, err := s.ReadEvent(ctx)
	if err != nil {
		return nil, false, err
	}

	if eventType != EventTypeMessage {
		return nil, false, fmt.Errorf("unexpected event type: %s", eventType)
	}

	var resp struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, false, fmt.Errorf("unmarshaling event data: %w", err)
	}

	var event a2a.TaskArtifactUpdateEvent
	if err := json.Unmarshal(resp.Result, &event); err != nil {
		// Try to unmarshal as TaskStatusUpdateEvent instead
		var statusEvent a2a.TaskStatusUpdateEvent
		if err := json.Unmarshal(resp.Result, &statusEvent); err != nil {
			return nil, false, fmt.Errorf("unmarshaling event result: %w", err)
		}
		// It's a status event, not an artifact event
		return nil, false, nil
	}

	return &event, true, nil
}

// EventType represents the type of a server-sent event.
type EventType string

// Event types
const (
	EventTypeMessage EventType = "message"
	EventTypeError   EventType = "error"
	EventTypeOpen    EventType = "open"
)
