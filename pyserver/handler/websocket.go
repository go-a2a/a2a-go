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
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/gorilla/websocket"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/internal/jsonrpc2"
)

// WebSocketHandler handles A2A requests over WebSocket transport.
type WebSocketHandler struct {
	router   Router
	upgrader websocket.Upgrader
	sessions sync.Map // map[string]*wsSession
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler(router Router) *WebSocketHandler {
	return &WebSocketHandler{
		router: router,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// TODO: Implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// ServeHTTP implements http.Handler for WebSocket endpoints.
func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create session
	session := &wsSession{
		conn:      conn,
		handler:   h,
		send:      make(chan []byte, 256),
		sessionID: r.Header.Get("X-Session-ID"),
		ctx:       r.Context(),
		done:      make(chan struct{}),
	}

	// Store session
	if session.sessionID != "" {
		h.sessions.Store(session.sessionID, session)
		defer h.sessions.Delete(session.sessionID)
	}

	// Start session handlers
	go session.writePump()
	go session.readPump()

	// Wait for session to complete
	<-session.done
}

// wsSession represents a WebSocket session.
type wsSession struct {
	conn      *websocket.Conn
	handler   *WebSocketHandler
	send      chan []byte
	sessionID string
	ctx       context.Context
	done      chan struct{}
	closeOnce sync.Once
}

// readPump handles incoming messages from the WebSocket connection.
func (s *wsSession) readPump() {
	defer s.close()

	// Configure connection
	s.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	s.conn.SetPongHandler(func(string) error {
		s.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Reset read deadline
		s.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		// Process message
		s.handleMessage(message)
	}
}

// writePump handles outgoing messages to the WebSocket connection.
func (s *wsSession) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		s.close()
	}()

	for {
		select {
		case message, ok := <-s.send:
			s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := s.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes an incoming WebSocket message.
func (s *wsSession) handleMessage(message []byte) {
	// Try to parse as JSON-RPC request
	var req jsonrpc2.Request
	if err := json.Unmarshal(message, &req); err != nil {
		s.sendError(nil, a2a.ErrParse.Code, "parse error", nil)
		return
	}

	// Create request context
	reqCtx := &RequestContext{
		Method:    a2a.Method(req.Method),
		RequestID: fmt.Sprintf("%v", req.ID),
		Session: &Session{
			ID:       s.sessionID,
			Metadata: make(map[string]any),
		},
	}

	// Parse params if present
	if req.Params != nil {
		params, err := parseJSONRPCParams(reqCtx.Method, req.Params)
		if err != nil {
			s.sendError(req.ID, a2a.ErrInvalidParams, "invalid params", nil)
			return
		}
		reqCtx.Params = params
	}

	// Check if this is a streaming method
	isStreaming := false
	switch reqCtx.Method {
	case a2a.MethodMessageStream, a2a.MethodTasksResubscribe:
		isStreaming = true
	}

	if isStreaming {
		s.handleStreamingRequest(req, reqCtx)
	} else {
		s.handleStandardRequest(req, reqCtx)
	}
}

// handleStandardRequest handles non-streaming requests.
func (s *wsSession) handleStandardRequest(req jsonrpc2.Request, reqCtx *RequestContext) {
	// Route to handler
	handler, err := s.handler.router.Route(reqCtx.Method)
	if err != nil {
		s.sendError(req.ID, jsonrpc2.CodeMethodNotFound, "method not found", nil)
		return
	}

	// Execute handler
	result, err := handler.Handle(s.ctx, reqCtx.Method, reqCtx.Params)
	if err != nil {
		// Convert to JSON-RPC error
		var jsonrpcErr *a2a.Error
		if e, ok := err.(*a2a.Error); ok {
			jsonrpcErr = e
		} else {
			jsonrpcErr = &a2a.Error{
				Code:    jsonrpc2.CodeInternalError,
				Message: err.Error(),
			}
		}
		s.sendError(req.ID, jsonrpcErr.Code, jsonrpcErr.Message, jsonrpcErr.Data)
		return
	}

	// Send response
	s.sendResult(req.ID, result)
}

// handleStreamingRequest handles streaming requests.
func (s *wsSession) handleStreamingRequest(req jsonrpc2.Request, reqCtx *RequestContext) {
	// Route to streaming handler
	handler, err := s.handler.router.RouteStreaming(reqCtx.Method)
	if err != nil {
		s.sendError(req.ID, jsonrpc2.CodeMethodNotFound, "streaming method not found", nil)
		return
	}

	// Start streaming
	eventChan, err := handler.HandleStream(s.ctx, reqCtx.Method, reqCtx.Params)
	if err != nil {
		s.sendError(req.ID, jsonrpc2.CodeInternalError, err.Error(), nil)
		return
	}

	// Send initial acknowledgment
	s.sendResult(req.ID, map[string]any{
		"status": "streaming",
		"taskId": reqCtx.Params.(*a2a.MessageSendParams).Message.TaskID,
	})

	// Stream events
	go s.streamEvents(req.ID, eventChan)
}

// streamEvents sends streaming events over the WebSocket.
func (s *wsSession) streamEvents(requestID any, eventChan <-chan a2a.SendStreamingMessageResponse) {
	for event := range eventChan {
		// Create a notification for each event
		notification := map[string]any{
			"jsonrpc": "2.0",
			"method":  "stream.event",
			"params": map[string]any{
				"requestId": requestID,
				"event":     event,
			},
		}

		data, err := json.Marshal(notification)
		if err != nil {
			log.Printf("Failed to marshal streaming event: %v", err)
			continue
		}

		select {
		case s.send <- data:
		case <-s.done:
			return
		}
	}

	// Send stream complete notification
	s.sendNotification("stream.complete", map[string]any{
		"requestId": requestID,
	})
}

// sendResult sends a successful JSON-RPC response.
func (s *wsSession) sendResult(id jsonrpc2.ID, result any) {
	resp := &jsonrpc2.Response{
		ID:     id,
		Result: result,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	select {
	case s.send <- data:
	case <-s.done:
	}
}

// sendError sends a JSON-RPC error response.
func (s *wsSession) sendError(id any, code int64, message string, data any) {
	resp := &jsonrpc2.Response{
		ID: id,
		Error: &a2a.Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	respData, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal error response: %v", err)
		return
	}

	select {
	case s.send <- respData:
	case <-s.done:
	}
}

// sendNotification sends a JSON-RPC notification.
func (s *wsSession) sendNotification(method string, params any) {
	notification := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}

	data, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal notification: %v", err)
		return
	}

	select {
	case s.send <- data:
	case <-s.done:
	}
}

// close cleanly closes the WebSocket session.
func (s *wsSession) close() {
	s.closeOnce.Do(func() {
		close(s.done)
		close(s.send)
		s.conn.Close()
	})
}

