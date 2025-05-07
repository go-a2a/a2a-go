// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/bytedance/sonic"

	"github.com/go-a2a/a2a"
)

// Stream represents a Server-Sent Events (SSE) connection
type Stream struct {
	w           http.ResponseWriter
	flusher     http.Flusher
	taskID      string
	mu          sync.Mutex
	isConnected bool
}

// StreamRegistry manages active SSE streams
type StreamRegistry struct {
	streams map[string]*Stream
	mu      sync.RWMutex
}

// NewStreamRegistry creates a new StreamRegistry
func NewStreamRegistry() *StreamRegistry {
	return &StreamRegistry{
		streams: make(map[string]*Stream),
	}
}

// CreateStream initializes a new SSE stream for a task
func (r *StreamRegistry) CreateStream(taskID string, w http.ResponseWriter, req *http.Request) *Stream {
	// Check if the client supports SSE
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Close existing stream if any
	if existingStream, exists := r.streams[taskID]; exists {
		existingStream.isConnected = false
		delete(r.streams, taskID)
	}

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // For Nginx proxy

	// Create and register new stream
	stream := &Stream{
		w:           w,
		flusher:     flusher,
		taskID:      taskID,
		isConnected: true,
	}
	r.streams[taskID] = stream

	return stream
}

// GetStream retrieves a stream by task ID
func (r *StreamRegistry) GetStream(taskID string) *Stream {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.streams[taskID]
}

// CloseStream removes a stream from the registry
func (r *StreamRegistry) CloseStream(taskID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if stream, exists := r.streams[taskID]; exists {
		stream.isConnected = false
		delete(r.streams, taskID)
	}
}

// SendStatusUpdate sends a task status update through the stream
func (s *Stream) SendStatusUpdate(event a2a.TaskStatusUpdateEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isConnected {
		return fmt.Errorf("stream is closed")
	}

	data, err := sonic.ConfigDefault.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	fmt.Fprintf(s.w, "event: status\ndata: %s\n\n", data)
	s.flusher.Flush()

	return nil
}

// SendArtifactUpdate sends a task artifact update through the stream
func (s *Stream) SendArtifactUpdate(event a2a.TaskArtifactUpdateEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isConnected {
		return fmt.Errorf("stream is closed")
	}

	data, err := sonic.ConfigDefault.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	fmt.Fprintf(s.w, "event: artifact\ndata: %s\n\n", data)
	s.flusher.Flush()

	return nil
}
