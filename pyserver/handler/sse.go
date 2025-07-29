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

package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-json-experiment/json"

	a2a "github.com/go-a2a/a2a-go"
)

// SSEHandler handles Server-Sent Events for streaming A2A responses.
type SSEHandler struct {
	router Router
}

// NewSSEHandler creates a new SSE handler.
func NewSSEHandler(router Router) *SSEHandler {
	return &SSEHandler{
		router: router,
	}
}

// ServeHTTP implements http.Handler for SSE endpoints.
func (h *SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable Nginx buffering

	// Ensure writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	// Extract method from URL
	method := extractMethod(r.URL.Path)
	if method == "" {
		h.writeSSEError(w, flusher, "invalid endpoint")
		return
	}

	// Create request context
	reqCtx := &RequestContext{
		Method:    a2a.Method(method),
		Headers:   r.Header,
		RequestID: r.Header.Get("X-Request-ID"),
	}

	// Add session if available
	session := extractSession(r)
	if session != nil {
		reqCtx.Session = session
	}

	// Parse request body for initial params
	params, err := h.parseStreamingParams(r, reqCtx.Method)
	if err != nil {
		h.writeSSEError(w, flusher, fmt.Sprintf("invalid parameters: %v", err))
		return
	}
	reqCtx.Params = params

	// Route to streaming handler
	handler, err := h.router.RouteStreaming(reqCtx.Method)
	if err != nil {
		h.writeSSEError(w, flusher, "method not found")
		return
	}

	// Start streaming
	eventChan, err := handler.HandleStream(ctx, reqCtx.Method, params)
	if err != nil {
		h.writeSSEError(w, flusher, err.Error())
		return
	}

	// Send initial connection event
	h.writeSSEEvent(w, flusher, "connected", map[string]string{
		"status": "connected",
		"time":   time.Now().Format(time.RFC3339),
	})

	// Stream events
	h.streamEvents(ctx, w, flusher, eventChan)
}

// parseStreamingParams parses parameters for streaming requests.
func (h *SSEHandler) parseStreamingParams(r *http.Request, method a2a.Method) (a2a.Params, error) {
	switch method {
	case a2a.MethodMessageStream:
		// For streaming, params might come from query or body
		var params a2a.MessageSendParams
		if r.Method == http.MethodPost {
			if err := json.UnmarshalRead(r.Body, &params); err != nil {
				return nil, err
			}
		}
		return &params, nil
	case a2a.MethodTasksResubscribe:
		// Task ID from query param
		taskID := r.URL.Query().Get("id")
		if taskID == "" {
			return nil, fmt.Errorf("task ID required")
		}
		return &a2a.TaskIDParams{ID: taskID}, nil
	default:
		return nil, fmt.Errorf("unsupported streaming method: %s", method)
	}
}

// streamEvents streams events from the channel to the client.
func (h *SSEHandler) streamEvents(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, eventChan <-chan a2a.SendStreamingMessageResponse) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Context cancelled
			h.writeSSEEvent(w, flusher, "close", map[string]string{
				"reason": "context cancelled",
			})
			return

		case <-ticker.C:
			// Send heartbeat
			h.writeSSEComment(w, flusher, "heartbeat")

		case event, ok := <-eventChan:
			if !ok {
				// Channel closed, stream is done
				h.writeSSEEvent(w, flusher, "close", map[string]string{
					"reason": "stream completed",
				})
				return
			}

			// Send event based on type
			eventType := string(event.GetEventKind())
			if err := h.writeSSEData(w, flusher, eventType, event); err != nil {
				// Client disconnected
				return
			}
		}
	}
}

// writeSSEEvent writes a simple SSE event.
func (h *SSEHandler) writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, event string, data any) {
	fmt.Fprintf(w, "event: %s\n", event)
	if data != nil {
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n", jsonData)
	}
	fmt.Fprintln(w)
	flusher.Flush()
}

// writeSSEData writes SSE data with proper formatting.
func (h *SSEHandler) writeSSEData(w http.ResponseWriter, flusher http.Flusher, eventType string, data any) error {
	// Write event type
	if _, err := fmt.Fprintf(w, "event: %s\n", eventType); err != nil {
		return err
	}

	// Marshal and write data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", jsonData); err != nil {
		return err
	}

	flusher.Flush()
	return nil
}

// writeSSEComment writes an SSE comment (for heartbeats).
func (h *SSEHandler) writeSSEComment(w http.ResponseWriter, flusher http.Flusher, comment string) {
	fmt.Fprintf(w, ": %s\n\n", comment)
	flusher.Flush()
}

// writeSSEError writes an error event and closes the stream.
func (h *SSEHandler) writeSSEError(w http.ResponseWriter, flusher http.Flusher, errMsg string) {
	h.writeSSEEvent(w, flusher, "error", map[string]string{
		"error": errMsg,
	})
	h.writeSSEEvent(w, flusher, "close", map[string]string{
		"reason": "error",
	})
}

// SSEClient provides a client for testing SSE endpoints.
type SSEClient struct {
	URL    string
	Events chan SSEEvent
	Errors chan error
	Done   chan struct{}
}

// SSEEvent represents a parsed SSE event.
type SSEEvent struct {
	Type string
	Data string
	ID   string
}

// Connect establishes an SSE connection.
func (c *SSEClient) Connect(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Start reading events in background
	go c.readEvents(resp)

	return nil
}

// readEvents reads SSE events from the response.
func (c *SSEClient) readEvents(resp *http.Response) {
	defer resp.Body.Close()
	defer close(c.Done)

	// Simple SSE parser (production code would use a proper parser)
	// This is a simplified implementation for the example
	// TODO: Implement proper SSE parsing
}