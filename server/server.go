// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package server provides a server implementation for the Google A2A protocol.
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
	"time"

	"github.com/go-a2a/a2a"
)

// Common errors that may occur during server operations
var (
	ErrTaskNotFound   = errors.New("task not found")
	ErrInvalidRequest = errors.New("invalid request")
	ErrMethodNotFound = errors.New("method not found")
	ErrInternalError  = errors.New("internal server error")
	ErrInvalidParams  = errors.New("invalid parameters")
	ErrNotImplemented = errors.New("method not implemented")
	ErrInvalidTaskID  = errors.New("invalid task ID")
	ErrInvalidState   = errors.New("invalid task state")
	ErrUnauthorized   = errors.New("unauthorized")
)

// JSON-RPC error codes
const (
	ErrorCodeParse          = -32700
	ErrorCodeInvalidRequest = -32600
	ErrorCodeMethodNotFound = -32601
	ErrorCodeInvalidParams  = -32602
	ErrorCodeInternal       = -32603
	ErrorCodeServerError    = -32000
)

// TaskStore is an interface for storing and retrieving tasks
type TaskStore interface {
	// GetTask retrieves a task by ID
	GetTask(ctx context.Context, id string) (*a2a.Task, error)

	// CreateTask creates a new task
	CreateTask(ctx context.Context, task *a2a.Task) error

	// UpdateTask updates an existing task
	UpdateTask(ctx context.Context, task *a2a.Task) error

	// DeleteTask deletes a task by ID
	DeleteTask(ctx context.Context, id string) error

	// ListTasks lists all tasks, optionally filtered by session ID
	ListTasks(ctx context.Context, sessionID string) ([]*a2a.Task, error)
}

// TaskHandler is the interface for handling tasks
type TaskHandler interface {
	// HandleTask processes a task and returns the updated task
	HandleTask(ctx context.Context, task *a2a.Task) (*a2a.Task, error)

	// HandleTaskUpdate processes an update to a task and returns the updated task
	HandleTaskUpdate(ctx context.Context, task *a2a.Task, message a2a.Message) (*a2a.Task, error)

	// HandleTaskCancel processes a task cancellation request
	HandleTaskCancel(ctx context.Context, task *a2a.Task, reason string) error
}

// A2AServer is a server for the A2A protocol
type A2AServer struct {
	// AgentCard contains metadata about the server
	AgentCard a2a.AgentCard

	// TaskStore is the store for tasks
	TaskStore TaskStore

	// TaskHandler is the handler for tasks
	TaskHandler TaskHandler

	// Path is the path to expose the server on
	Path string

	// Subscriptions is a map of task ID to subscribers
	Subscriptions map[string]map[chan<- *a2a.Task]struct{}

	// SubscriptionsMutex is a mutex for the subscriptions map
	SubscriptionsMutex sync.RWMutex

	// Middleware contains HTTP middleware functions
	Middleware []func(http.Handler) http.Handler
}

// ServerOptions contains options for creating a new A2A server
type ServerOptions struct {
	AgentCard   a2a.AgentCard
	TaskStore   TaskStore
	TaskHandler TaskHandler
	Path        string
	Middleware  []func(http.Handler) http.Handler
}

// NewServer creates a new A2A server
func NewServer(opts ServerOptions) *A2AServer {
	// Set default path
	if opts.Path == "" {
		opts.Path = "/"
	}

	// Set default task store
	if opts.TaskStore == nil {
		opts.TaskStore = NewInMemoryTaskStore()
	}

	// Set default agent card
	if opts.AgentCard.Name == "" {
		opts.AgentCard = a2a.AgentCard{
			AgentType:   "A2A",
			Name:        "Go A2A Server",
			Description: "A Go implementation of the A2A protocol",
			Version:     a2a.Version,
			Capabilities: a2a.AgentCapabilities{
				Streaming:          true,
				PushNotifications:  false,
				MultiTurn:          true,
				MultiTask:          true,
				DefaultOutputMode:  "text",
				AcceptedInputModes: []string{"text"},
			},
		}
	}

	return &A2AServer{
		AgentCard:     opts.AgentCard,
		TaskStore:     opts.TaskStore,
		TaskHandler:   opts.TaskHandler,
		Path:          opts.Path,
		Subscriptions: make(map[string]map[chan<- *a2a.Task]struct{}),
		Middleware:    opts.Middleware,
	}
}

// ServeHTTP implements the http.Handler interface
func (s *A2AServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeRPCError(w, nil, ErrorCodeParse, "Failed to read request body", err)
		return
	}

	// Parse request
	var request a2a.JsonRpcRequest
	if err := json.Unmarshal(body, &request); err != nil {
		writeRPCError(w, nil, ErrorCodeParse, "Failed to parse request", err)
		return
	}

	// Create context
	ctx := r.Context()

	// Handle request
	response := s.handleRequest(ctx, request)

	// Handle streaming responses
	if isStreamingMethod(request.Method) {
		// Set headers for streaming
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		// Create subscriber channel
		taskCh := make(chan *a2a.Task)

		// Extract task ID from request
		taskID := extractTaskIDFromRequest(request)
		if taskID == "" {
			writeRPCError(w, request.ID, ErrorCodeInvalidParams, "Invalid task ID", nil)
			return
		}

		// Add subscriber
		s.addSubscriber(taskID, taskCh)
		defer s.removeSubscriber(taskID, taskCh)

		// Flush initial response if not nil
		if response != nil && response.Result != nil {
			if err := writeJSONToStream(w, response); err != nil {
				return
			}
			w.(http.Flusher).Flush()
		}

		// Stream updates
		flusher, ok := w.(http.Flusher)
		if !ok {
			writeRPCError(w, request.ID, ErrorCodeInternal, "Streaming not supported", nil)
			return
		}

		// Handle streaming method
		if request.Method == "tasks/sendSubscribe" {
			// Process the task and send initial update
			if task, ok := response.Result.(*a2a.Task); ok {
				s.publishTaskUpdate(task)
			}
		}

		// Create a channel to detect when the client disconnects
		disconnectCh := make(chan struct{})
		go func() {
			<-ctx.Done()
			close(disconnectCh)
		}()

		// Stream updates
		for {
			select {
			case task, ok := <-taskCh:
				if !ok {
					return
				}

				streamResp := a2a.JsonRpcResponse{
					JsonRpc: "2.0",
					ID:      request.ID,
					Result:  task,
				}

				if err := writeJSONToStream(w, streamResp); err != nil {
					return
				}
				flusher.Flush()

				// If task is completed, failed, or canceled, stop streaming
				if task.Status.State == a2a.TaskCompleted ||
					task.Status.State == a2a.TaskFailed ||
					task.Status.State == a2a.TaskCanceled {
					return
				}
			case <-disconnectCh:
				return
			}
		}
	} else {
		// Write response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			return
		}
	}
}

// Handler returns an http.Handler for the server
func (s *A2AServer) Handler() http.Handler {
	// Apply middleware
	var handler http.Handler = s
	for i := len(s.Middleware) - 1; i >= 0; i-- {
		handler = s.Middleware[i](handler)
	}
	return handler
}

// Start starts the server on the specified address
func (s *A2AServer) Start(addr string) error {
	mux := http.NewServeMux()
	mux.Handle(s.Path, s.Handler())
	return http.ListenAndServe(addr, mux)
}

// StartTLS starts the server with TLS on the specified address
func (s *A2AServer) StartTLS(addr, certFile, keyFile string) error {
	mux := http.NewServeMux()
	mux.Handle(s.Path, s.Handler())
	return http.ListenAndServeTLS(addr, certFile, keyFile, mux)
}

// addSubscriber adds a subscriber for task updates
func (s *A2AServer) addSubscriber(taskID string, ch chan<- *a2a.Task) {
	s.SubscriptionsMutex.Lock()
	defer s.SubscriptionsMutex.Unlock()

	if _, ok := s.Subscriptions[taskID]; !ok {
		s.Subscriptions[taskID] = make(map[chan<- *a2a.Task]struct{})
	}
	s.Subscriptions[taskID][ch] = struct{}{}
}

// removeSubscriber removes a subscriber for task updates
func (s *A2AServer) removeSubscriber(taskID string, ch chan<- *a2a.Task) {
	s.SubscriptionsMutex.Lock()
	defer s.SubscriptionsMutex.Unlock()

	if subscribers, ok := s.Subscriptions[taskID]; ok {
		delete(subscribers, ch)
		if len(subscribers) == 0 {
			delete(s.Subscriptions, taskID)
		}
	}
}

// publishTaskUpdate sends a task update to all subscribers
func (s *A2AServer) publishTaskUpdate(task *a2a.Task) {
	s.SubscriptionsMutex.RLock()
	defer s.SubscriptionsMutex.RUnlock()

	subscribers, ok := s.Subscriptions[task.ID]
	if !ok {
		return
	}

	// Send update to all subscribers
	for ch := range subscribers {
		select {
		case ch <- task:
			// Update sent successfully
		default:
			// Subscriber not ready, skip
		}
	}
}

// handleRequest handles a JSON-RPC request
func (s *A2AServer) handleRequest(ctx context.Context, request a2a.JsonRpcRequest) *a2a.JsonRpcResponse {
	// Create response
	response := &a2a.JsonRpcResponse{
		JsonRpc: "2.0",
		ID:      request.ID,
	}

	// Handle request based on method
	var err error
	switch request.Method {
	case "agent/card":
		response.Result = s.handleAgentCard(ctx, request.Params)
	case "tasks/get":
		response.Result, err = s.handleTasksGet(ctx, request.Params)
	case "tasks/send":
		response.Result, err = s.handleTasksSend(ctx, request.Params)
	case "tasks/cancel":
		response.Result, err = s.handleTasksCancel(ctx, request.Params)
	case "tasks/subscribe":
		response.Result, err = s.handleTasksSubscribe(ctx, request.Params)
	case "tasks/sendSubscribe":
		response.Result, err = s.handleTasksSendSubscribe(ctx, request.Params)
	default:
		err = fmt.Errorf("%w: %s", ErrMethodNotFound, request.Method)
	}

	// Handle error
	if err != nil {
		code := ErrorCodeServerError
		message := err.Error()

		// Map error to appropriate code
		switch {
		case errors.Is(err, ErrMethodNotFound):
			code = ErrorCodeMethodNotFound
		case errors.Is(err, ErrInvalidParams):
			code = ErrorCodeInvalidParams
		case errors.Is(err, ErrInvalidRequest):
			code = ErrorCodeInvalidRequest
		case errors.Is(err, ErrInternalError):
			code = ErrorCodeInternal
		}

		response.Error = &a2a.JsonRpcError{
			Code:    code,
			Message: message,
		}
		response.Result = nil
	}

	return response
}

// handleAgentCard handles the agent/card method
func (s *A2AServer) handleAgentCard(ctx context.Context, params any) any {
	return s.AgentCard
}

// handleTasksGet handles the tasks/get method
func (s *A2AServer) handleTasksGet(ctx context.Context, params any) (any, error) {
	// Parse parameters
	var reqParams a2a.TasksGetRequest
	if err := parseParams(params, &reqParams); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidParams, err)
	}

	// Validate ID
	if reqParams.ID == "" {
		return nil, fmt.Errorf("%w: missing task ID", ErrInvalidParams)
	}

	// Get task
	task, err := s.TaskStore.GetTask(ctx, reqParams.ID)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrTaskNotFound, reqParams.ID)
		}
		return nil, fmt.Errorf("%w: %v", ErrInternalError, err)
	}

	return task, nil
}

// handleTasksSend handles the tasks/send method
func (s *A2AServer) handleTasksSend(ctx context.Context, params any) (any, error) {
	// Parse parameters
	var reqParams a2a.TasksSendRequest
	if err := parseParams(params, &reqParams); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidParams, err)
	}

	// Validate ID
	if reqParams.ID == "" {
		return nil, fmt.Errorf("%w: missing task ID", ErrInvalidParams)
	}

	// Validate message
	if len(reqParams.Message.Parts) == 0 {
		return nil, fmt.Errorf("%w: empty message parts", ErrInvalidParams)
	}

	// Check if task exists
	existingTask, err := s.TaskStore.GetTask(ctx, reqParams.ID)
	if err != nil && !errors.Is(err, ErrTaskNotFound) {
		return nil, fmt.Errorf("%w: %v", ErrInternalError, err)
	}

	if existingTask == nil {
		// Create new task
		now := time.Now()
		task := &a2a.Task{
			ID:                  reqParams.ID,
			SessionID:           reqParams.SessionID,
			Status:              a2a.TaskStatus{State: a2a.TaskSubmitted, Timestamp: now},
			Artifacts:           []a2a.Artifact{},
			History:             []a2a.Message{},
			AcceptedOutputModes: reqParams.AcceptedOutputModes,
			Metadata:            reqParams.Metadata,
			ParentTask:          reqParams.ParentTask,
			PrevTasks:           reqParams.PrevTasks,
			CreatedAt:           now,
			UpdatedAt:           now,
			Message:             &reqParams.Message,
		}

		// Save task
		if err := s.TaskStore.CreateTask(ctx, task); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInternalError, err)
		}

		// Process task
		if s.TaskHandler != nil {
			processedTask, err := s.TaskHandler.HandleTask(ctx, task)
			if err != nil {
				// Update task status to failed
				task.Status.State = a2a.TaskFailed
				task.Status.Timestamp = time.Now()
				task.Status.Error = &a2a.TaskError{
					Code:    "task_handler_error",
					Message: err.Error(),
				}
				s.TaskStore.UpdateTask(ctx, task)
				return task, nil
			}
			task = processedTask
		}

		// Publish task update
		s.publishTaskUpdate(task)

		return task, nil
	}

	// Task exists, process update
	if s.TaskHandler != nil {
		updatedTask, err := s.TaskHandler.HandleTaskUpdate(ctx, existingTask, reqParams.Message)
		if err != nil {
			// Update task status to failed
			existingTask.Status.State = a2a.TaskFailed
			existingTask.Status.Timestamp = time.Now()
			existingTask.Status.Error = &a2a.TaskError{
				Code:    "task_handler_error",
				Message: err.Error(),
			}
			s.TaskStore.UpdateTask(ctx, existingTask)
			return existingTask, nil
		}
		existingTask = updatedTask
	}

	// Add message to history
	existingTask.History = append(existingTask.History, reqParams.Message)
	existingTask.UpdatedAt = time.Now()

	// Update task
	if err := s.TaskStore.UpdateTask(ctx, existingTask); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalError, err)
	}

	// Publish task update
	s.publishTaskUpdate(existingTask)

	return existingTask, nil
}

// handleTasksCancel handles the tasks/cancel method
func (s *A2AServer) handleTasksCancel(ctx context.Context, params any) (any, error) {
	// Parse parameters
	var reqParams a2a.TasksCancelRequest
	if err := parseParams(params, &reqParams); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidParams, err)
	}

	// Validate ID
	if reqParams.ID == "" {
		return nil, fmt.Errorf("%w: missing task ID", ErrInvalidParams)
	}

	// Get task
	task, err := s.TaskStore.GetTask(ctx, reqParams.ID)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrTaskNotFound, reqParams.ID)
		}
		return nil, fmt.Errorf("%w: %v", ErrInternalError, err)
	}

	// Update task status
	task.Status.State = a2a.TaskCanceled
	task.Status.Timestamp = time.Now()
	if reqParams.Reason != "" {
		reason := reqParams.Reason
		task.Status.Reason = &reason
	}

	// Handle cancel
	if s.TaskHandler != nil {
		if err := s.TaskHandler.HandleTaskCancel(ctx, task, reqParams.Reason); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInternalError, err)
		}
	}

	// Update task
	if err := s.TaskStore.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalError, err)
	}

	// Publish task update
	s.publishTaskUpdate(task)

	return task, nil
}

// handleTasksSubscribe handles the tasks/subscribe method
func (s *A2AServer) handleTasksSubscribe(ctx context.Context, params any) (any, error) {
	// Parse parameters
	var reqParams a2a.TasksSubscribeRequest
	if err := parseParams(params, &reqParams); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidParams, err)
	}

	// Validate ID
	if reqParams.ID == "" {
		return nil, fmt.Errorf("%w: missing task ID", ErrInvalidParams)
	}

	// Get task
	task, err := s.TaskStore.GetTask(ctx, reqParams.ID)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrTaskNotFound, reqParams.ID)
		}
		return nil, fmt.Errorf("%w: %v", ErrInternalError, err)
	}

	return task, nil
}

// handleTasksSendSubscribe handles the tasks/sendSubscribe method
func (s *A2AServer) handleTasksSendSubscribe(ctx context.Context, params any) (any, error) {
	// Parse parameters
	var reqParams a2a.TasksSendSubscribeRequest
	if err := parseParams(params, &reqParams); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidParams, err)
	}

	// Validate ID
	if reqParams.ID == "" {
		return nil, fmt.Errorf("%w: missing task ID", ErrInvalidParams)
	}

	// Validate message
	if len(reqParams.Message.Parts) == 0 {
		return nil, fmt.Errorf("%w: empty message parts", ErrInvalidParams)
	}

	// Check if task exists
	existingTask, err := s.TaskStore.GetTask(ctx, reqParams.ID)
	if err != nil && !errors.Is(err, ErrTaskNotFound) {
		return nil, fmt.Errorf("%w: %v", ErrInternalError, err)
	}

	if existingTask == nil {
		// Create new task
		now := time.Now()
		task := &a2a.Task{
			ID:                  reqParams.ID,
			SessionID:           reqParams.SessionID,
			Status:              a2a.TaskStatus{State: a2a.TaskSubmitted, Timestamp: now},
			Artifacts:           []a2a.Artifact{},
			History:             []a2a.Message{},
			AcceptedOutputModes: reqParams.AcceptedOutputModes,
			Metadata:            reqParams.Metadata,
			ParentTask:          reqParams.ParentTask,
			PrevTasks:           reqParams.PrevTasks,
			CreatedAt:           now,
			UpdatedAt:           now,
			Message:             &reqParams.Message,
		}

		// Save task
		if err := s.TaskStore.CreateTask(ctx, task); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInternalError, err)
		}

		// Process task asynchronously
		go func() {
			if s.TaskHandler != nil {
				// Update status to working
				task.Status.State = a2a.TaskWorking
				task.Status.Timestamp = time.Now()
				s.TaskStore.UpdateTask(context.Background(), task)
				s.publishTaskUpdate(task)

				// Process task
				processedTask, err := s.TaskHandler.HandleTask(context.Background(), task)
				if err != nil {
					// Update task status to failed
					task.Status.State = a2a.TaskFailed
					task.Status.Timestamp = time.Now()
					task.Status.Error = &a2a.TaskError{
						Code:    "task_handler_error",
						Message: err.Error(),
					}
					s.TaskStore.UpdateTask(context.Background(), task)
					s.publishTaskUpdate(task)
					return
				}

				// Update task
				s.TaskStore.UpdateTask(context.Background(), processedTask)
				s.publishTaskUpdate(processedTask)
			}
		}()

		return task, nil
	}

	// Task exists, process update asynchronously
	go func() {
		if s.TaskHandler != nil {
			// Update status to working
			existingTask.Status.State = a2a.TaskWorking
			existingTask.Status.Timestamp = time.Now()
			existingTask.Message = &reqParams.Message
			s.TaskStore.UpdateTask(context.Background(), existingTask)
			s.publishTaskUpdate(existingTask)

			// Process update
			updatedTask, err := s.TaskHandler.HandleTaskUpdate(context.Background(), existingTask, reqParams.Message)
			if err != nil {
				// Update task status to failed
				existingTask.Status.State = a2a.TaskFailed
				existingTask.Status.Timestamp = time.Now()
				existingTask.Status.Error = &a2a.TaskError{
					Code:    "task_handler_error",
					Message: err.Error(),
				}
				s.TaskStore.UpdateTask(context.Background(), existingTask)
				s.publishTaskUpdate(existingTask)
				return
			}

			// Add message to history
			updatedTask.History = append(updatedTask.History, reqParams.Message)
			updatedTask.UpdatedAt = time.Now()

			// Update task
			s.TaskStore.UpdateTask(context.Background(), updatedTask)
			s.publishTaskUpdate(updatedTask)
		}
	}()

	return existingTask, nil
}

// parseParams parses JSON-RPC parameters
func parseParams(params any, dest any) error {
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// writeRPCError writes a JSON-RPC error response
func writeRPCError(w http.ResponseWriter, id any, code int, message string, err error) {
	errResp := a2a.JsonRpcResponse{
		JsonRpc: "2.0",
		ID:      id,
		Error: &a2a.JsonRpcError{
			Code:    code,
			Message: message,
		},
	}

	if err != nil {
		errResp.Error.Data = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(errResp)
}

// writeJSONToStream writes a JSON object to a streaming response
func writeJSONToStream(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}
	return nil
}

// isStreamingMethod checks if a method supports streaming
func isStreamingMethod(method string) bool {
	return method == "tasks/subscribe" || method == "tasks/sendSubscribe"
}

// extractTaskIDFromRequest extracts the task ID from a request
func extractTaskIDFromRequest(request a2a.JsonRpcRequest) string {
	switch request.Method {
	case "tasks/subscribe":
		var params a2a.TasksSubscribeRequest
		data, _ := json.Marshal(request.Params)
		_ = json.Unmarshal(data, &params)
		return params.ID
	case "tasks/sendSubscribe":
		var params a2a.TasksSendSubscribeRequest
		data, _ := json.Marshal(request.Params)
		_ = json.Unmarshal(data, &params)
		return params.ID
	default:
		return ""
	}
}

// InMemoryTaskStore is an in-memory implementation of TaskStore
type InMemoryTaskStore struct {
	tasks map[string]*a2a.Task
	mutex sync.RWMutex
}

// NewInMemoryTaskStore creates a new in-memory task store
func NewInMemoryTaskStore() *InMemoryTaskStore {
	return &InMemoryTaskStore{
		tasks: make(map[string]*a2a.Task),
	}
}

// GetTask retrieves a task by ID
func (s *InMemoryTaskStore) GetTask(ctx context.Context, id string) (*a2a.Task, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	task, ok := s.tasks[id]
	if !ok {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

// CreateTask creates a new task
func (s *InMemoryTaskStore) CreateTask(ctx context.Context, task *a2a.Task) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if task already exists
	if _, ok := s.tasks[task.ID]; ok {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	}

	// Store task
	s.tasks[task.ID] = task
	return nil
}

// UpdateTask updates an existing task
func (s *InMemoryTaskStore) UpdateTask(ctx context.Context, task *a2a.Task) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if task exists
	if _, ok := s.tasks[task.ID]; !ok {
		return ErrTaskNotFound
	}

	// Update task
	s.tasks[task.ID] = task
	return nil
}

// DeleteTask deletes a task by ID
func (s *InMemoryTaskStore) DeleteTask(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if task exists
	if _, ok := s.tasks[id]; !ok {
		return ErrTaskNotFound
	}

	// Delete task
	delete(s.tasks, id)
	return nil
}

// ListTasks lists all tasks, optionally filtered by session ID
func (s *InMemoryTaskStore) ListTasks(ctx context.Context, sessionID string) ([]*a2a.Task, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var tasks []*a2a.Task
	for _, task := range s.tasks {
		if sessionID == "" || task.SessionID == sessionID {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}
