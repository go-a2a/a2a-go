// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package server implements the A2A server functionality.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-a2a/a2a"
)

// TaskStore is an interface for storing and retrieving tasks.
type TaskStore interface {
	// GetTask returns a task by ID.
	GetTask(ctx context.Context, id string) (*a2a.Task, error)
	// CreateTask creates a new task.
	CreateTask(ctx context.Context, task *a2a.Task) error
	// UpdateTask updates an existing task.
	UpdateTask(ctx context.Context, task *a2a.Task) error
	// DeleteTask deletes a task.
	DeleteTask(ctx context.Context, id string) error
	// ListTasks returns all tasks.
	ListTasks(ctx context.Context) ([]*a2a.Task, error)
}

// InMemoryTaskStore is an in-memory implementation of TaskStore.
type InMemoryTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*a2a.Task
}

var _ TaskStore = (*InMemoryTaskStore)(nil)

// NewInMemoryTaskStore creates a new in-memory task store.
func NewInMemoryTaskStore() *InMemoryTaskStore {
	return &InMemoryTaskStore{
		tasks: make(map[string]*a2a.Task),
	}
}

// GetTask returns a task by ID.
func (s *InMemoryTaskStore) GetTask(ctx context.Context, id string) (*a2a.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	return task, nil
}

// CreateTask creates a new task.
func (s *InMemoryTaskStore) CreateTask(ctx context.Context, task *a2a.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[task.ID]; ok {
		return fmt.Errorf("task already exists: %s", task.ID)
	}

	s.tasks[task.ID] = task
	return nil
}

// UpdateTask updates an existing task.
func (s *InMemoryTaskStore) UpdateTask(ctx context.Context, task *a2a.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[task.ID]; !ok {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	s.tasks[task.ID] = task
	return nil
}

// DeleteTask deletes a task.
func (s *InMemoryTaskStore) DeleteTask(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return fmt.Errorf("task not found: %s", id)
	}

	delete(s.tasks, id)
	return nil
}

// ListTasks returns all tasks.
func (s *InMemoryTaskStore) ListTasks(ctx context.Context) ([]*a2a.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*a2a.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// A2AHandler is an interface for handling A2A method calls.
type A2AHandler interface {
	// GetAgentCard returns the agent's card.
	GetAgentCard(ctx context.Context) (*a2a.AgentCard, error)
	// TaskSend handles a tasks/send request.
	TaskSend(ctx context.Context, params *a2a.TaskSendParams) (*a2a.Task, error)
	// TaskGet handles a tasks/get request.
	TaskGet(ctx context.Context, params *a2a.TaskGetParams) (*a2a.Task, error)
	// TaskCancel handles a tasks/cancel request.
	TaskCancel(ctx context.Context, params *a2a.TaskCancelParams) (*a2a.Task, error)
	// SetPushNotification handles a tasks/pushNotification/set request.
	SetPushNotification(ctx context.Context, params *a2a.SetPushNotificationParams) (any, error)
}

// TaskEventEmitter is an interface for emitting task events.
type TaskEventEmitter interface {
	// EmitStatusUpdate emits a task status update event.
	EmitStatusUpdate(ctx context.Context, event *a2a.TaskStatusUpdateEvent) error
	// EmitArtifactUpdate emits a task artifact update event.
	EmitArtifactUpdate(ctx context.Context, event *a2a.TaskArtifactUpdateEvent) error
	// EmitTaskCompletion emits a task completion event.
	EmitTaskCompletion(ctx context.Context, task *a2a.Task) error
}

// A2AServer is a server for the A2A protocol.
type A2AServer struct {
	// HTTPServer is the underlying HTTP server.
	HTTPServer *http.Server

	closed atomic.Bool

	// Router is the HTTP router.
	// Router *http.ServeMux
	// Handler is the A2A method handler.
	Handler A2AHandler
	// TaskStore is the task store.
	TaskStore TaskStore
	// EventEmitter is the task event emitter.
	EventEmitter TaskEventEmitter
	// AgentCard is the agent's card.
	AgentCard *a2a.AgentCard
	// SubscriptionManager manages task subscriptions.
	SubscriptionManager *SubscriptionManager
}

// NewA2AServer creates a new A2A server.
func NewA2AServer(addr string, handler A2AHandler, taskStore TaskStore, eventEmitter TaskEventEmitter) *A2AServer {
	// Set up HTTP routes
	router := http.NewServeMux()

	server := &A2AServer{
		HTTPServer: &http.Server{
			Addr: addr,
			// Handler: router,
		},
		// Router:              router,
		Handler:             handler,
		TaskStore:           taskStore,
		EventEmitter:        eventEmitter,
		SubscriptionManager: NewSubscriptionManager(),
	}

	router.HandleFunc("POST /", server.handleJSONRPC)
	router.HandleFunc("POST /tasks/sendSubscribe", server.handleTaskSendSubscribe)

	server.HTTPServer.Handler = router

	return server
}

// Start starts the HTTP server.
func (s *A2AServer) Start() error {
	return s.HTTPServer.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *A2AServer) Shutdown(ctx context.Context) error {
	return s.HTTPServer.Shutdown(ctx)
}

// handleJSONRPC handles JSON-RPC requests.
func (s *A2AServer) handleJSONRPC(w http.ResponseWriter, r *http.Request) {
	var req a2a.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSONRPCError(w, http.StatusBadRequest, -32700, "Parse error", err)
		return
	}

	if req.JSONRPC != "2.0" {
		s.writeJSONRPCError(w, http.StatusBadRequest, -32600, "Invalid Request", errors.New("invalid jsonrpc version"))
		return
	}

	result, err := s.dispatchMethod(r.Context(), &req)
	if err != nil {
		s.writeJSONRPCError(w, http.StatusInternalServerError, -32603, "Internal error", err)
		return
	}

	s.writeJSONRPCResponse(w, http.StatusOK, req.ID, result)
}

// dispatchMethod dispatches a JSON-RPC method call to the appropriate handler.
func (s *A2AServer) dispatchMethod(ctx context.Context, req *a2a.JSONRPCRequest) (any, error) {
	switch req.Method {
	case "agent/getCard":
		return s.Handler.GetAgentCard(ctx)

	case "tasks/send":
		var params a2a.TaskSendParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		return s.Handler.TaskSend(ctx, &params)

	case "tasks/get":
		var params a2a.TaskGetParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		return s.Handler.TaskGet(ctx, &params)

	case "tasks/cancel":
		var params a2a.TaskCancelParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		return s.Handler.TaskCancel(ctx, &params)

	case "tasks/pushNotification/set":
		var params a2a.SetPushNotificationParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		return s.Handler.SetPushNotification(ctx, &params)

	default:
		return nil, fmt.Errorf("method not found: %s", req.Method)
	}
}

// writeJSONRPCResponse writes a JSON-RPC response.
func (s *A2AServer) writeJSONRPCResponse(w http.ResponseWriter, status int, id any, result any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := a2a.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to write JSON-RPC response: %v", err)
	}
}

// writeJSONRPCError writes a JSON-RPC error response.
func (s *A2AServer) writeJSONRPCError(w http.ResponseWriter, status int, code int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := a2a.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      nil, // Set to null for parse errors
		Error: &a2a.JSONRPCError{
			Code:    code,
			Message: message,
			Data:    err.Error(),
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to write JSON-RPC error response: %v", err)
	}
}

// handleTaskSendSubscribe handles a tasks/sendSubscribe request.
func (s *A2AServer) handleTaskSendSubscribe(w http.ResponseWriter, r *http.Request) {
	var params a2a.TaskSendParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Create subscription
	// subscription, err := s.SubscriptionManager.CreateSubscription(params.ID)
	// if err != nil {
	// 	http.Error(w, fmt.Sprintf("Failed to create subscription: %v", err), http.StatusInternalServerError)
	// 	return
	// }
	// defer s.SubscriptionManager.DeleteSubscription(params.ID)

	// TODO(zchee): use subscription
	// _ = subscription

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	// Create a channel for the connection closing
	done := make(chan struct{})

	// Handle client disconnect
	go func() {
		<-r.Context().Done()
		s.closed.Store(true)
		close(done)
	}()

	// Submit the task
	task, err := s.Handler.TaskSend(r.Context(), &params)
	if err != nil {
		s.writeSSEEvent(w, "error", &a2a.JSONRPCEvent{
			JSONRPC: "2.0",
			Method:  "error",
			Params:  fmt.Sprintf("Failed to send task: %v", err),
		})
		return
	}

	// Send task completion event
	s.writeSSEEvent(w, "message", &a2a.JSONRPCEvent{
		JSONRPC: "2.0",
		Method:  "task/complete",
		Params:  task,
	})

	// Flush and signal we're done
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// writeSSEEvent writes a server-sent event.
func (s *A2AServer) writeSSEEvent(w io.Writer, event string, data any) error {
	// Marshal the data
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Write the event
	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(dataJSON))
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	// Flush if possible
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

// SubscriptionManager manages task subscriptions.
type SubscriptionManager struct {
	mu            sync.RWMutex
	subscriptions map[string]*Subscription
}

// NewSubscriptionManager creates a new subscription manager.
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		subscriptions: make(map[string]*Subscription),
	}
}

// CreateSubscription creates a new subscription for a task.
func (m *SubscriptionManager) CreateSubscription(taskID string) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.subscriptions[taskID]; ok {
		return nil, fmt.Errorf("subscription already exists for task: %s", taskID)
	}

	subscription := &Subscription{
		TaskID:          taskID,
		StatusChannel:   make(chan *a2a.TaskStatusUpdateEvent, 10),
		ArtifactChannel: make(chan *a2a.TaskArtifactUpdateEvent, 10),
	}

	m.subscriptions[taskID] = subscription
	return subscription, nil
}

// GetSubscription returns a subscription for a task.
func (m *SubscriptionManager) GetSubscription(taskID string) (*Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	subscription, ok := m.subscriptions[taskID]
	if !ok {
		return nil, fmt.Errorf("subscription not found for task: %s", taskID)
	}

	return subscription, nil
}

// DeleteSubscription deletes a subscription for a task.
func (m *SubscriptionManager) DeleteSubscription(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if subscription, ok := m.subscriptions[taskID]; ok {
		close(subscription.StatusChannel)
		close(subscription.ArtifactChannel)
		delete(m.subscriptions, taskID)
	}
}

// Subscription represents a client subscription to task events.
type Subscription struct {
	// TaskID is the ID of the task.
	TaskID string
	// StatusChannel is a channel for status updates.
	StatusChannel chan *a2a.TaskStatusUpdateEvent
	// ArtifactChannel is a channel for artifact updates.
	ArtifactChannel chan *a2a.TaskArtifactUpdateEvent
}

// DefaultA2AHandler is a default implementation of A2AHandler.
type DefaultA2AHandler struct {
	// AgentCard is the agent's card.
	AgentCard *a2a.AgentCard
	// TaskStore is the task store.
	TaskStore TaskStore
	// EventEmitter is the task event emitter.
	EventEmitter TaskEventEmitter
	// TaskCallback is a function that processes a task.
	TaskCallback func(ctx context.Context, task *a2a.Task) error
	// PushNotificationConfig is the push notification configuration.
	PushNotificationConfig *a2a.PushNotificationConfig
}

var _ A2AHandler = (*DefaultA2AHandler)(nil)

// NewDefaultA2AHandler creates a new default A2A handler.
func NewDefaultA2AHandler(agentCard *a2a.AgentCard, taskStore TaskStore, eventEmitter TaskEventEmitter, taskCallback func(ctx context.Context, task *a2a.Task) error) *DefaultA2AHandler {
	return &DefaultA2AHandler{
		AgentCard:    agentCard,
		TaskStore:    taskStore,
		EventEmitter: eventEmitter,
		TaskCallback: taskCallback,
	}
}

// GetAgentCard returns the agent's card.
func (h *DefaultA2AHandler) GetAgentCard(ctx context.Context) (*a2a.AgentCard, error) {
	return h.AgentCard, nil
}

// TaskSend handles a tasks/send request.
func (h *DefaultA2AHandler) TaskSend(ctx context.Context, params *a2a.TaskSendParams) (*a2a.Task, error) {
	// Create a new task
	task := &a2a.Task{
		ID: params.ID,
		Status: a2a.TaskStatus{
			State:     a2a.TaskStateSubmitted,
			Timestamp: time.Now(),
		},
		History:   []a2a.Message{params.Message},
		SessionID: params.SessionID,
	}

	// Store the task
	if err := h.TaskStore.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Update task status to working
	task.Status.State = a2a.TaskStateWorking
	task.Status.Timestamp = time.Now()
	if err := h.TaskStore.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	// Emit status update event
	if h.EventEmitter != nil {
		statusEvent := &a2a.TaskStatusUpdateEvent{
			ID:     task.ID,
			Status: task.Status,
		}
		if err := h.EventEmitter.EmitStatusUpdate(ctx, statusEvent); err != nil {
			log.Printf("Failed to emit status update event: %v", err)
		}
	}

	// Process the task
	if h.TaskCallback != nil {
		if err := h.TaskCallback(ctx, task); err != nil {
			task.Status.State = a2a.TaskStateFailed
			task.Status.Timestamp = time.Now()
			task.Status.Message = err.Error()
			if errUpdate := h.TaskStore.UpdateTask(ctx, task); errUpdate != nil {
				log.Printf("Failed to update task status: %v", errUpdate)
			}
			return nil, fmt.Errorf("failed to process task: %w", err)
		}
	}

	// Get the updated task
	updatedTask, err := h.TaskStore.GetTask(ctx, task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return updatedTask, nil
}

// TaskGet handles a tasks/get request.
func (h *DefaultA2AHandler) TaskGet(ctx context.Context, params *a2a.TaskGetParams) (*a2a.Task, error) {
	task, err := h.TaskStore.GetTask(ctx, params.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

// TaskCancel handles a tasks/cancel request.
func (h *DefaultA2AHandler) TaskCancel(ctx context.Context, params *a2a.TaskCancelParams) (*a2a.Task, error) {
	task, err := h.TaskStore.GetTask(ctx, params.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Update task status to canceled
	task.Status.State = a2a.TaskStateCanceled
	task.Status.Timestamp = time.Now()

	if err := h.TaskStore.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	// Emit status update event
	if h.EventEmitter != nil {
		statusEvent := &a2a.TaskStatusUpdateEvent{
			ID:     task.ID,
			Status: task.Status,
		}
		if err := h.EventEmitter.EmitStatusUpdate(ctx, statusEvent); err != nil {
			log.Printf("Failed to emit status update event: %v", err)
		}
	}

	return task, nil
}

// SetPushNotification handles a tasks/pushNotification/set request.
func (h *DefaultA2AHandler) SetPushNotification(ctx context.Context, params *a2a.SetPushNotificationParams) (any, error) {
	h.PushNotificationConfig = &params.Config
	return struct{}{}, nil
}

// DefaultTaskEventEmitter is a default implementation of TaskEventEmitter.
type DefaultTaskEventEmitter struct {
	// SubscriptionManager is the subscription manager.
	SubscriptionManager *SubscriptionManager
}

var _ TaskEventEmitter = (*DefaultTaskEventEmitter)(nil)

// NewDefaultTaskEventEmitter creates a new default task event emitter.
func NewDefaultTaskEventEmitter(subscriptionManager *SubscriptionManager) *DefaultTaskEventEmitter {
	return &DefaultTaskEventEmitter{
		SubscriptionManager: subscriptionManager,
	}
}

// EmitStatusUpdate emits a task status update event.
func (e *DefaultTaskEventEmitter) EmitStatusUpdate(ctx context.Context, event *a2a.TaskStatusUpdateEvent) error {
	subscription, err := e.SubscriptionManager.GetSubscription(event.ID)
	if err != nil {
		return nil // No subscription, just ignore
	}

	select {
	case subscription.StatusChannel <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Channel is full, discard the event
		log.Printf("Status channel is full for task %s, discarding event", event.ID)
		return nil
	}
}

// EmitArtifactUpdate emits a task artifact update event.
func (e *DefaultTaskEventEmitter) EmitArtifactUpdate(ctx context.Context, event *a2a.TaskArtifactUpdateEvent) error {
	subscription, err := e.SubscriptionManager.GetSubscription(event.ID)
	if err != nil {
		return nil // No subscription, just ignore
	}

	select {
	case subscription.ArtifactChannel <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Channel is full, discard the event
		log.Printf("Artifact channel is full for task %s, discarding event", event.ID)
		return nil
	}
}

// EmitTaskCompletion emits a task completion event.
func (e *DefaultTaskEventEmitter) EmitTaskCompletion(ctx context.Context, task *a2a.Task) error {
	// This would typically use a webhook or other notification mechanism
	// For this example, we'll just log it
	log.Printf("Task completed: %s", task.ID)
	return nil
}
