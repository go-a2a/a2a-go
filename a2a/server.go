// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	json "github.com/bytedance/sonic"
)

// TaskManager defines the interface for managing tasks in an A2A server.
type TaskManager interface {
	// CreateTask creates a new task.
	CreateTask(params TaskSendParams) (*Task, error)
	// GetTask retrieves a task by ID.
	GetTask(id string, historyLength *int) (*Task, error)
	// CancelTask cancels a task by ID.
	CancelTask(id string) (*Task, error)
	// SetTaskPushNotification sets push notification configuration for a task.
	SetTaskPushNotification(config TaskPushNotificationConfig) (*TaskPushNotificationConfig, error)
	// GetTaskPushNotification gets push notification configuration for a task.
	GetTaskPushNotification(id string) (*TaskPushNotificationConfig, error)
}

// InMemoryTaskManager implements TaskManager using in-memory storage.
type InMemoryTaskManager struct {
	tasks                   map[string]*Task
	pushNotificationConfigs map[string]PushNotificationConfig
	mu                      sync.RWMutex
}

// NewInMemoryTaskManager creates a new InMemoryTaskManager.
func NewInMemoryTaskManager() *InMemoryTaskManager {
	return &InMemoryTaskManager{
		tasks:                   make(map[string]*Task),
		pushNotificationConfigs: make(map[string]PushNotificationConfig),
	}
}

// CreateTask creates a new task or updates an existing one.
func (m *InMemoryTaskManager) CreateTask(params TaskSendParams) (*Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if the task already exists
	var task *Task
	if existingTask, ok := m.tasks[params.ID]; ok {
		task = existingTask
	} else {
		// Create a new task
		task = &Task{
			ID: params.ID,
			Status: TaskStatus{
				State:     TaskStateSubmitted,
				Timestamp: time.Now(),
			},
		}
	}

	// Update the task with the new message
	task.SessionID = params.SessionID
	task.Status.Message = &params.Message
	task.Status.State = TaskStateWorking
	task.Status.Timestamp = time.Now()

	// Store the push notification config if provided
	if params.PushNotification != nil {
		m.pushNotificationConfigs[params.ID] = *params.PushNotification
	}

	// Store the task
	m.tasks[params.ID] = task

	return task, nil
}

// GetTask retrieves a task by ID.
func (m *InMemoryTaskManager) GetTask(id string, historyLength *int) (*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get the task
	task, ok := m.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	return task, nil
}

// CancelTask cancels a task by ID.
func (m *InMemoryTaskManager) CancelTask(id string) (*Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the task
	task, ok := m.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	// Check if the task is in a final state
	if task.Status.State == TaskStateCompleted || task.Status.State == TaskStateCanceled || task.Status.State == TaskStateFailed {
		return nil, fmt.Errorf("task is in a final state and cannot be canceled")
	}

	// Update the task status
	task.Status.State = TaskStateCanceled
	task.Status.Timestamp = time.Now()

	return task, nil
}

// SetTaskPushNotification sets push notification configuration for a task.
func (m *InMemoryTaskManager) SetTaskPushNotification(config TaskPushNotificationConfig) (*TaskPushNotificationConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the task
	_, ok := m.tasks[config.ID]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", config.ID)
	}

	// Store the push notification config
	m.pushNotificationConfigs[config.ID] = config.PushNotificationConfig

	return &config, nil
}

// GetTaskPushNotification gets push notification configuration for a task.
func (m *InMemoryTaskManager) GetTaskPushNotification(id string) (*TaskPushNotificationConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get the task
	_, ok := m.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	// Get the push notification config
	config, ok := m.pushNotificationConfigs[id]
	if !ok {
		return nil, fmt.Errorf("push notification configuration not found for task: %s", id)
	}

	return &TaskPushNotificationConfig{
		ID:                     id,
		PushNotificationConfig: config,
	}, nil
}

// Server represents an A2A server.
type Server struct {
	AgentCard   AgentCard
	TaskManager TaskManager
}

// NewServer creates a new A2A server with the given agent card and task manager.
func NewServer(agentCard AgentCard, taskManager TaskManager) *Server {
	return &Server{
		AgentCard:   agentCard,
		TaskManager: taskManager,
	}
}

// ServeHTTP implements http.Handler for the A2A server.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle well-known agent.json
	if r.URL.Path == "/.well-known/agent.json" {
		s.handleAgentCard(w, r)
		return
	}

	// Handle A2A API requests
	if r.Method == http.MethodPost {
		s.handleAPIRequest(w, r)
		return
	}

	// Method not allowed
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleAgentCard handles requests for the agent card.
func (s *Server) handleAgentCard(w http.ResponseWriter, r *http.Request) {
	// Only support GET for the agent card
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Marshal the agent card to JSON
	jsonData, err := json.ConfigFastest.Marshal(s.AgentCard)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the content type and write the response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// handleAPIRequest handles A2A API requests.
func (s *Server) handleAPIRequest(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, NewJSONParseError())
		return
	}

	// Parse the request
	var request JSONRPCRequest
	if err := json.ConfigFastest.Unmarshal(body, &request); err != nil {
		s.writeError(w, NewJSONParseError())
		return
	}

	// Handle the request based on the method
	switch request.Method {
	case MethodTasksSend:
		s.handleTasksSend(w, request)
	case MethodTasksGet:
		s.handleTasksGet(w, request)
	case MethodTasksCancel:
		s.handleTasksCancel(w, request)
	case MethodTasksPushNotificationSet:
		s.handleTasksPushNotificationSet(w, request)
	case MethodTasksPushNotificationGet:
		s.handleTasksPushNotificationGet(w, request)
	default:
		s.writeError(w, NewMethodNotFoundError())
	}
}

// handleTasksSend handles tasks/send requests.
func (s *Server) handleTasksSend(w http.ResponseWriter, request JSONRPCRequest) {
	// Parse the request parameters
	var params TaskSendParams
	paramsJson, err := json.ConfigFastest.Marshal(request.Params)
	if err != nil {
		s.writeError(w, NewInvalidParamsError())
		return
	}

	if err := json.ConfigFastest.Unmarshal(paramsJson, &params); err != nil {
		s.writeError(w, NewInvalidParamsError())
		return
	}

	// Create the task
	task, err := s.TaskManager.CreateTask(params)
	if err != nil {
		s.writeError(w, NewInternalError())
		return
	}

	// Create the response
	response := SendTaskResponse{
		JSONRPCMessage: JSONRPCMessage{
			JSONRPC: "2.0",
			ID:      request.ID,
		},
		Result: task,
	}

	// Write the response
	s.writeResponse(w, response)
}

// handleTasksGet handles tasks/get requests.
func (s *Server) handleTasksGet(w http.ResponseWriter, request JSONRPCRequest) {
	// Parse the request parameters
	var params TaskQueryParams
	paramsJson, err := json.ConfigFastest.Marshal(request.Params)
	if err != nil {
		s.writeError(w, NewInvalidParamsError())
		return
	}

	if err := json.ConfigFastest.Unmarshal(paramsJson, &params); err != nil {
		s.writeError(w, NewInvalidParamsError())
		return
	}

	// Get the task
	task, err := s.TaskManager.GetTask(params.ID, params.HistoryLength)
	if err != nil {
		s.writeError(w, NewTaskNotFoundError())
		return
	}

	// Create the response
	response := GetTaskResponse{
		JSONRPCMessage: JSONRPCMessage{
			JSONRPC: "2.0",
			ID:      request.ID,
		},
		Result: task,
	}

	// Write the response
	s.writeResponse(w, response)
}

// handleTasksCancel handles tasks/cancel requests.
func (s *Server) handleTasksCancel(w http.ResponseWriter, request JSONRPCRequest) {
	// Parse the request parameters
	var params TaskIdParams
	paramsJson, err := json.ConfigFastest.Marshal(request.Params)
	if err != nil {
		s.writeError(w, NewInvalidParamsError())
		return
	}

	if err := json.ConfigFastest.Unmarshal(paramsJson, &params); err != nil {
		s.writeError(w, NewInvalidParamsError())
		return
	}

	// Cancel the task
	task, err := s.TaskManager.CancelTask(params.ID)
	if err != nil {
		if err.Error() == "task not found: "+params.ID {
			s.writeError(w, NewTaskNotFoundError())
		} else {
			s.writeError(w, NewTaskNotCancelableError())
		}
		return
	}

	// Create the response
	response := CancelTaskResponse{
		JSONRPCMessage: JSONRPCMessage{
			JSONRPC: "2.0",
			ID:      request.ID,
		},
		Result: task,
	}

	// Write the response
	s.writeResponse(w, response)
}

// handleTasksPushNotificationSet handles tasks/pushNotification/set requests.
func (s *Server) handleTasksPushNotificationSet(w http.ResponseWriter, request JSONRPCRequest) {
	// Check if push notifications are supported
	if !s.AgentCard.Capabilities.PushNotifications {
		s.writeError(w, NewPushNotificationNotSupportedError())
		return
	}

	// Parse the request parameters
	var params TaskPushNotificationConfig
	paramsJson, err := json.ConfigFastest.Marshal(request.Params)
	if err != nil {
		s.writeError(w, NewInvalidParamsError())
		return
	}

	if err := json.ConfigFastest.Unmarshal(paramsJson, &params); err != nil {
		s.writeError(w, NewInvalidParamsError())
		return
	}

	// Set the push notification config
	config, err := s.TaskManager.SetTaskPushNotification(params)
	if err != nil {
		if err.Error() == "task not found: "+params.ID {
			s.writeError(w, NewTaskNotFoundError())
		} else {
			s.writeError(w, NewInternalError())
		}
		return
	}

	// Create the response
	response := SetTaskPushNotificationResponse{
		JSONRPCMessage: JSONRPCMessage{
			JSONRPC: "2.0",
			ID:      request.ID,
		},
		Result: config,
	}

	// Write the response
	s.writeResponse(w, response)
}

// handleTasksPushNotificationGet handles tasks/pushNotification/get requests.
func (s *Server) handleTasksPushNotificationGet(w http.ResponseWriter, request JSONRPCRequest) {
	// Check if push notifications are supported
	if !s.AgentCard.Capabilities.PushNotifications {
		s.writeError(w, NewPushNotificationNotSupportedError())
		return
	}

	// Parse the request parameters
	var params TaskIdParams
	paramsJson, err := json.ConfigFastest.Marshal(request.Params)
	if err != nil {
		s.writeError(w, NewInvalidParamsError())
		return
	}

	if err := json.ConfigFastest.Unmarshal(paramsJson, &params); err != nil {
		s.writeError(w, NewInvalidParamsError())
		return
	}

	// Get the push notification config
	config, err := s.TaskManager.GetTaskPushNotification(params.ID)
	if err != nil {
		if err.Error() == "task not found: "+params.ID {
			s.writeError(w, NewTaskNotFoundError())
		} else {
			s.writeError(w, NewInternalError())
		}
		return
	}

	// Create the response
	response := GetTaskPushNotificationResponse{
		JSONRPCMessage: JSONRPCMessage{
			JSONRPC: "2.0",
			ID:      request.ID,
		},
		Result: config,
	}

	// Write the response
	s.writeResponse(w, response)
}

// writeError writes a JSON-RPC error response.
func (s *Server) writeError(w http.ResponseWriter, err *JSONRPCError) {
	response := JSONRPCResponse{
		JSONRPCMessage: JSONRPCMessage{
			JSONRPC: "2.0",
			ID:      nil,
		},
		Error: err,
	}

	// Marshal the response to JSON
	jsonData, merr := json.ConfigFastest.Marshal(response)
	if merr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the content type and write the response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// writeResponse writes a JSON-RPC response.
func (s *Server) writeResponse(w http.ResponseWriter, response any) {
	// Marshal the response to JSON
	jsonData, err := json.ConfigFastest.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the content type and write the response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
