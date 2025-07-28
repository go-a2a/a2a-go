// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"context"
	"testing"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/server"
)

// mockAgentExecutor is a mock implementation of AgentExecutor for testing.
type mockAgentExecutor struct {
	executeFunc       func(ctx context.Context, messages []*a2a.Message) (*ExecutionResult, error)
	executeStreamFunc func(ctx context.Context, messages []*a2a.Message) (<-chan *a2a.Message, error)
	cancelFunc        func(ctx context.Context, taskID string) error
}

func (m *mockAgentExecutor) Execute(ctx context.Context, messages []*a2a.Message) (*ExecutionResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, messages)
	}
	return &ExecutionResult{
		Messages: []*a2a.Message{
			{
				Content: "Mock response",
			},
		},
		TaskID: "test-task-id",
	}, nil
}

func (m *mockAgentExecutor) ExecuteStream(ctx context.Context, messages []*a2a.Message) (<-chan *a2a.Message, error) {
	if m.executeStreamFunc != nil {
		return m.executeStreamFunc(ctx, messages)
	}
	ch := make(chan *a2a.Message, 1)
	ch <- &a2a.Message{
		Content: "Mock streaming response",
	}
	close(ch)
	return ch, nil
}

func (m *mockAgentExecutor) Cancel(ctx context.Context, taskID string) error {
	if m.cancelFunc != nil {
		return m.cancelFunc(ctx, taskID)
	}
	return nil
}

// mockTaskStore is a mock implementation of TaskStore for testing.
type mockTaskStore struct {
	tasks      map[string]*a2a.Task
	getFunc    func(ctx context.Context, taskID string) (*a2a.Task, error)
	createFunc func(ctx context.Context, task *a2a.Task) error
	updateFunc func(ctx context.Context, task *a2a.Task) error
	deleteFunc func(ctx context.Context, taskID string) error
}

func (m *mockTaskStore) Get(ctx context.Context, taskID string) (*a2a.Task, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, taskID)
	}
	if task, exists := m.tasks[taskID]; exists {
		return task, nil
	}
	return nil, NewTaskNotFoundError(taskID)
}

func (m *mockTaskStore) Create(ctx context.Context, task *a2a.Task) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, task)
	}
	if m.tasks == nil {
		m.tasks = make(map[string]*a2a.Task)
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskStore) Update(ctx context.Context, task *a2a.Task) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, task)
	}
	if m.tasks == nil {
		m.tasks = make(map[string]*a2a.Task)
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskStore) Delete(ctx context.Context, taskID string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, taskID)
	}
	if m.tasks != nil {
		delete(m.tasks, taskID)
	}
	return nil
}

// TestDefaultRequestHandler_OnGetTask tests the OnGetTask method.
func TestDefaultRequestHandler_OnGetTask(t *testing.T) {
	// Create test task
	testTask := &a2a.Task{
		ID:        "test-task-id",
		ContextID: "test-context-id",
		Status:    a2a.TaskStatus{State: a2a.TaskStateCompleted},
		History:   []*a2a.Message{},
		Artifacts: []*a2a.Artifact{},
	}

	// Create mocks
	mockExecutor := &mockAgentExecutor{}
	mockStore := &mockTaskStore{
		tasks: map[string]*a2a.Task{
			"test-task-id": testTask,
		},
	}

	// Create handler
	handler := NewDefaultRequestHandler(mockExecutor, mockStore)

	// Create context and request
	ctx := context.Background()
	callCtx := server.NewServerCallContext(nil)
	request := &GetTaskRequest{TaskID: "test-task-id"}

	// Test successful get
	response, err := handler.OnGetTask(ctx, callCtx, request)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response == nil {
		t.Error("Expected response, got nil")
	}
	if response.Task.ID != "test-task-id" {
		t.Errorf("Expected task ID 'test-task-id', got '%s'", response.Task.ID)
	}

	// Test task not found
	request.TaskID = "non-existent-task"
	response, err = handler.OnGetTask(ctx, callCtx, request)
	if err == nil {
		t.Error("Expected error for non-existent task")
	}
	if _, ok := err.(*TaskNotFoundError); !ok {
		t.Errorf("Expected TaskNotFoundError, got %T", err)
	}

	// Test validation error
	request.TaskID = ""
	response, err = handler.OnGetTask(ctx, callCtx, request)
	if err == nil {
		t.Error("Expected validation error for empty task ID")
	}
	if _, ok := err.(*InvalidRequestError); !ok {
		t.Errorf("Expected InvalidRequestError, got %T", err)
	}
}

// TestDefaultRequestHandler_OnMessageSend tests the OnMessageSend method.
func TestDefaultRequestHandler_OnMessageSend(t *testing.T) {
	// Create mocks
	mockExecutor := &mockAgentExecutor{}
	mockStore := &mockTaskStore{}

	// Create handler
	handler := NewDefaultRequestHandler(mockExecutor, mockStore)

	// Create context and request
	ctx := context.Background()
	callCtx := server.NewServerCallContext(nil)
	request := &MessageSendRequest{
		Messages: []*a2a.Message{
			{
				Content: "Test message",
			},
		},
	}

	// Test successful message send
	response, err := handler.OnMessageSend(ctx, callCtx, request)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response == nil {
		t.Error("Expected response, got nil")
	}
	if len(response.Messages) == 0 {
		t.Error("Expected messages in response")
	}
	if response.TaskID == "" {
		t.Error("Expected task ID in response")
	}

	// Test validation error
	request.Messages = nil
	response, err = handler.OnMessageSend(ctx, callCtx, request)
	if err == nil {
		t.Error("Expected validation error for empty messages")
	}
	if _, ok := err.(*InvalidRequestError); !ok {
		t.Errorf("Expected InvalidRequestError, got %T", err)
	}
}

// TestDefaultRequestHandler_OnCancelTask tests the OnCancelTask method.
func TestDefaultRequestHandler_OnCancelTask(t *testing.T) {
	// Create test task
	testTask := &a2a.Task{
		ID:        "test-task-id",
		ContextID: "test-context-id",
		Status:    a2a.TaskStatus{State: a2a.TaskStateRunning},
		History:   []*a2a.Message{},
		Artifacts: []*a2a.Artifact{},
	}

	// Create mocks
	mockExecutor := &mockAgentExecutor{}
	mockStore := &mockTaskStore{
		tasks: map[string]*a2a.Task{
			"test-task-id": testTask,
		},
	}

	// Create handler
	handler := NewDefaultRequestHandler(mockExecutor, mockStore)

	// Create context and request
	ctx := context.Background()
	callCtx := server.NewServerCallContext(nil)
	request := &CancelTaskRequest{TaskID: "test-task-id"}

	// Test successful cancel
	response, err := handler.OnCancelTask(ctx, callCtx, request)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response == nil {
		t.Error("Expected response, got nil")
	}
	if !response.Success {
		t.Error("Expected successful cancellation")
	}

	// Test task not found
	request.TaskID = "non-existent-task"
	response, err = handler.OnCancelTask(ctx, callCtx, request)
	if err == nil {
		t.Error("Expected error for non-existent task")
	}
	if _, ok := err.(*TaskNotFoundError); !ok {
		t.Errorf("Expected TaskNotFoundError, got %T", err)
	}
}

// TestJSONRPCHandler_HandleRequest tests the JSON-RPC handler.
func TestJSONRPCHandler_HandleRequest(t *testing.T) {
	// Create mocks
	mockExecutor := &mockAgentExecutor{}
	mockStore := &mockTaskStore{}

	// Create handlers
	requestHandler := NewDefaultRequestHandler(mockExecutor, mockStore)
	jsonrpcHandler := NewJSONRPCHandler(requestHandler)

	// Create context and request
	ctx := context.Background()
	request := &JSONRPCRequest{
		ID:      "test-request-id",
		Method:  "message_send",
		JSONRPC: "2.0",
		Params: map[string]any{
			"messages": []any{
				map[string]any{
					"content": "Test message",
				},
			},
		},
	}

	// Test successful request
	response := jsonrpcHandler.HandleRequest(ctx, request)
	if response == nil {
		t.Error("Expected response, got nil")
	}
	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}
	if response.Result == nil {
		t.Error("Expected result, got nil")
	}

	// Test invalid method
	request.Method = "invalid_method"
	response = jsonrpcHandler.HandleRequest(ctx, request)
	if response == nil {
		t.Error("Expected response, got nil")
	}
	if response.Error == nil {
		t.Error("Expected error for invalid method")
	}
	if response.Result != nil {
		t.Error("Expected no result for error response")
	}
}

// TestValidation tests the validation methods.
func TestValidation(t *testing.T) {
	// Test GetTaskRequest validation
	req := &GetTaskRequest{}
	err := req.Validate()
	if err == nil {
		t.Error("Expected validation error for empty task ID")
	}

	req.TaskID = "test-task-id"
	err = req.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid request, got %v", err)
	}

	// Test MessageSendRequest validation
	msgReq := &MessageSendRequest{}
	err = msgReq.Validate()
	if err == nil {
		t.Error("Expected validation error for empty messages")
	}

	msgReq.Messages = []*a2a.Message{
		{
			Content: "Test message",
		},
	}
	err = msgReq.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid request, got %v", err)
	}
}

// TestErrorHandling tests error handling and conversion.
func TestErrorHandling(t *testing.T) {
	// Test TaskNotFoundError
	err := NewTaskNotFoundError("test-task-id")
	if err.TaskID != "test-task-id" {
		t.Errorf("Expected task ID 'test-task-id', got '%s'", err.TaskID)
	}
	if err.GetCode() != a2a.ErrorCodeTaskNotFound {
		t.Errorf("Expected error code %d, got %d", a2a.ErrorCodeTaskNotFound, err.GetCode())
	}

	// Test JSON-RPC error response
	response := BuildErrorResponse("test-id", err)
	if response.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %v", response.ID)
	}
	if response.Error == nil {
		t.Error("Expected error in response")
	}
	if response.Result != nil {
		t.Error("Expected no result for error response")
	}
}

// TestHandlerValidation tests handler validation.
func TestHandlerValidation(t *testing.T) {
	// Test DefaultRequestHandler validation
	mockExecutor := &mockAgentExecutor{}
	mockStore := &mockTaskStore{}

	handler := NewDefaultRequestHandler(mockExecutor, mockStore)
	err := handler.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid handler, got %v", err)
	}

	// Test JSONRPCHandler validation
	jsonrpcHandler := NewJSONRPCHandler(handler)
	err = jsonrpcHandler.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid JSON-RPC handler, got %v", err)
	}

	// Test GRPCHandler validation
	grpcHandler := NewGRPCHandler(handler)
	err = grpcHandler.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid gRPC handler, got %v", err)
	}
}
