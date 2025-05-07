// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/go-a2a/a2a"
)

// PushHandler is a function that handles incoming push notifications.
type PushHandler func(ctx context.Context, taskID string, event any) error

// PushServer handles incoming push notifications from A2A agents.
type PushServer struct {
	handlers map[string]PushHandler
	mu       sync.RWMutex
	server   *http.Server
}

// NewPushServer creates a new push notification server.
func NewPushServer(addr string) *PushServer {
	ps := &PushServer{
		handlers: make(map[string]PushHandler),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", ps.handleNotification)

	ps.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return ps
}

// RegisterHandler registers a handler for a specific task ID.
func (ps *PushServer) RegisterHandler(taskID string, handler PushHandler) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.handlers[taskID] = handler
}

// UnregisterHandler removes a handler for a specific task ID.
func (ps *PushServer) UnregisterHandler(taskID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.handlers, taskID)
}

// Start starts the push notification server.
func (ps *PushServer) Start() error {
	return ps.server.ListenAndServe()
}

// StartTLS starts the push notification server with TLS.
func (ps *PushServer) StartTLS(certFile, keyFile string) error {
	return ps.server.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown gracefully shuts down the push notification server.
func (ps *PushServer) Shutdown(ctx context.Context) error {
	return ps.server.Shutdown(ctx)
}

// handleNotification handles incoming push notifications.
func (ps *PushServer) handleNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from path
	taskID := strings.TrimPrefix(r.URL.Path, "/")
	if taskID == "" {
		http.Error(w, "Task ID not specified in URL path", http.StatusBadRequest)
		return
	}

	// Parse JSON request body
	var notification any
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing JSON request body: %v", err), http.StatusBadRequest)
		return
	}

	// Get handler for this task ID
	ps.mu.RLock()
	handler, ok := ps.handlers[taskID]
	ps.mu.RUnlock()

	if !ok {
		http.Error(w, fmt.Sprintf("No handler registered for task ID %s", taskID), http.StatusNotFound)
		return
	}

	// Call handler
	ctx := r.Context()
	if err := handler(ctx, taskID, notification); err != nil {
		http.Error(w, fmt.Sprintf("Error handling notification: %v", err), http.StatusInternalServerError)
		return
	}

	// Acknowledge receipt
	w.WriteHeader(http.StatusOK)
}

// CreatePushNotificationConfig creates a push notification configuration.
func CreatePushNotificationConfig(url, token string, auth *a2a.AuthenticationInfo) a2a.PushNotificationConfig {
	return a2a.PushNotificationConfig{
		URL:            url,
		Token:          token,
		Authentication: auth,
	}
}

// CreateTaskPushNotificationConfig creates a task-specific push notification configuration.
func CreateTaskPushNotificationConfig(taskID string, config a2a.PushNotificationConfig) a2a.TaskPushNotificationConfig {
	return a2a.TaskPushNotificationConfig{
		ID:                     taskID,
		PushNotificationConfig: config,
	}
}
