package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/server"
)

// TestInMemoryTaskStore tests the in-memory task store.
func TestInMemoryTaskStore(t *testing.T) {
	// Create a task store
	store := server.NewInMemoryTaskStore()
	ctx := context.Background()

	// Create a test task
	task := &a2a.Task{
		ID: uuid.New().String(),
		Status: a2a.TaskStatus{
			State:     a2a.TaskStateSubmitted,
			Timestamp: time.Now(),
		},
		History: []a2a.Message{
			{
				Role: a2a.RoleUser,
				Parts: []a2a.Part{
					a2a.TextPart{Text: "Hello, agent!"},
				},
			},
		},
	}

	// Test creating a task
	t.Run("CreateTask", func(t *testing.T) {
		if err := store.CreateTask(ctx, task); err != nil {
			t.Fatalf("CreateTask failed: %v", err)
		}
	})

	// Test getting a task
	t.Run("GetTask", func(t *testing.T) {
		retrievedTask, err := store.GetTask(ctx, task.ID)
		if err != nil {
			t.Fatalf("GetTask failed: %v", err)
		}

		opts := cmp.Options{
			cmpopts.IgnoreUnexported(a2a.Task{}, a2a.TaskStatus{}, a2a.Message{}),
			cmpopts.IgnoreFields(a2a.TaskStatus{}, "Timestamp"),
		}
		if diff := cmp.Diff(task, retrievedTask, opts); diff != "" {
			t.Errorf("Task mismatch (-want +got):\n%s", diff)
		}
	})

	// Test updating a task
	t.Run("UpdateTask", func(t *testing.T) {
		// Update the task
		task.Status.State = a2a.TaskStateCompleted
		task.Status.Timestamp = time.Now()
		task.Artifacts = []a2a.Artifact{
			{
				Parts: []a2a.Part{
					a2a.TextPart{Text: "Hello, user!"},
				},
				Index: 0,
			},
		}

		if err := store.UpdateTask(ctx, task); err != nil {
			t.Fatalf("UpdateTask failed: %v", err)
		}

		// Verify the update
		updatedTask, err := store.GetTask(ctx, task.ID)
		if err != nil {
			t.Fatalf("GetTask after update failed: %v", err)
		}

		opts := cmp.Options{
			cmpopts.IgnoreUnexported(a2a.Task{}, a2a.TaskStatus{}, a2a.Message{}, a2a.Artifact{}),
			cmpopts.IgnoreFields(a2a.TaskStatus{}, "Timestamp"),
		}
		if diff := cmp.Diff(task, updatedTask, opts); diff != "" {
			t.Errorf("Updated task mismatch (-want +got):\n%s", diff)
		}
	})

	// Test listing tasks
	t.Run("ListTasks", func(t *testing.T) {
		// Create another task
		task2 := &a2a.Task{
			ID: uuid.New().String(),
			Status: a2a.TaskStatus{
				State:     a2a.TaskStateSubmitted,
				Timestamp: time.Now(),
			},
		}
		if err := store.CreateTask(ctx, task2); err != nil {
			t.Fatalf("CreateTask for task2 failed: %v", err)
		}

		// List tasks
		tasks, err := store.ListTasks(ctx)
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}

		// Verify the number of tasks
		if len(tasks) != 2 {
			t.Errorf("Expected 2 tasks, got %d", len(tasks))
		}

		// Verify task IDs
		taskIDs := make(map[string]bool)
		for _, t := range tasks {
			taskIDs[t.ID] = true
		}
		if !taskIDs[task.ID] {
			t.Errorf("Task 1 not found in list")
		}
		if !taskIDs[task2.ID] {
			t.Errorf("Task 2 not found in list")
		}
	})

	// Test deleting a task
	t.Run("DeleteTask", func(t *testing.T) {
		if err := store.DeleteTask(ctx, task.ID); err != nil {
			t.Fatalf("DeleteTask failed: %v", err)
		}

		// Verify the task is gone
		_, err := store.GetTask(ctx, task.ID)
		if err == nil {
			t.Errorf("Expected error getting deleted task, got nil")
		}
	})
}

// mockA2AHandler is a mock implementation of A2AHandler for testing.
type mockA2AHandler struct {
	// GetAgentCardFn is the function to call for GetAgentCard.
	GetAgentCardFn func(ctx context.Context) (*a2a.AgentCard, error)
	// TaskSendFn is the function to call for TaskSend.
	TaskSendFn func(ctx context.Context, params *a2a.TaskSendParams) (*a2a.Task, error)
	// TaskGetFn is the function to call for TaskGet.
	TaskGetFn func(ctx context.Context, params *a2a.TaskGetParams) (*a2a.Task, error)
	// TaskCancelFn is the function to call for TaskCancel.
	TaskCancelFn func(ctx context.Context, params *a2a.TaskCancelParams) (*a2a.Task, error)
	// SetPushNotificationFn is the function to call for SetPushNotification.
	SetPushNotificationFn func(ctx context.Context, params *a2a.SetPushNotificationParams) (any, error)
}

var _ server.A2AHandler = (*mockA2AHandler)(nil)

func (h *mockA2AHandler) GetAgentCard(ctx context.Context) (*a2a.AgentCard, error) {
	return h.GetAgentCardFn(ctx)
}

func (h *mockA2AHandler) TaskSend(ctx context.Context, params *a2a.TaskSendParams) (*a2a.Task, error) {
	return h.TaskSendFn(ctx, params)
}

func (h *mockA2AHandler) TaskGet(ctx context.Context, params *a2a.TaskGetParams) (*a2a.Task, error) {
	return h.TaskGetFn(ctx, params)
}

func (h *mockA2AHandler) TaskCancel(ctx context.Context, params *a2a.TaskCancelParams) (*a2a.Task, error) {
	return h.TaskCancelFn(ctx, params)
}

func (h *mockA2AHandler) SetPushNotification(ctx context.Context, params *a2a.SetPushNotificationParams) (any, error) {
	return h.SetPushNotificationFn(ctx, params)
}

// mockTaskStore is a mock implementation of TaskStore for testing.
type mockTaskStore struct {
	// GetTaskFn is the function to call for GetTask.
	GetTaskFn func(ctx context.Context, id string) (*a2a.Task, error)
	// CreateTaskFn is the function to call for CreateTask.
	CreateTaskFn func(ctx context.Context, task *a2a.Task) error
	// UpdateTaskFn is the function to call for UpdateTask.
	UpdateTaskFn func(ctx context.Context, task *a2a.Task) error
	// DeleteTaskFn is the function to call for DeleteTask.
	DeleteTaskFn func(ctx context.Context, id string) error
	// ListTasksFn is the function to call for ListTasks.
	ListTasksFn func(ctx context.Context) ([]*a2a.Task, error)
}

var _ server.TaskStore = (*mockTaskStore)(nil)

func (s *mockTaskStore) GetTask(ctx context.Context, id string) (*a2a.Task, error) {
	return s.GetTaskFn(ctx, id)
}

func (s *mockTaskStore) CreateTask(ctx context.Context, task *a2a.Task) error {
	return s.CreateTaskFn(ctx, task)
}

func (s *mockTaskStore) UpdateTask(ctx context.Context, task *a2a.Task) error {
	return s.UpdateTaskFn(ctx, task)
}

func (s *mockTaskStore) DeleteTask(ctx context.Context, id string) error {
	return s.DeleteTaskFn(ctx, id)
}

func (s *mockTaskStore) ListTasks(ctx context.Context) ([]*a2a.Task, error) {
	return s.ListTasksFn(ctx)
}

// mockTaskEventEmitter is a mock implementation of TaskEventEmitter for testing.
type mockTaskEventEmitter struct {
	// EmitStatusUpdateFn is the function to call for EmitStatusUpdate.
	EmitStatusUpdateFn func(ctx context.Context, event *a2a.TaskStatusUpdateEvent) error
	// EmitArtifactUpdateFn is the function to call for EmitArtifactUpdate.
	EmitArtifactUpdateFn func(ctx context.Context, event *a2a.TaskArtifactUpdateEvent) error
	// EmitTaskCompletionFn is the function to call for EmitTaskCompletion.
	EmitTaskCompletionFn func(ctx context.Context, task *a2a.Task) error
}

var _ server.TaskEventEmitter = (*mockTaskEventEmitter)(nil)

func (e *mockTaskEventEmitter) EmitStatusUpdate(ctx context.Context, event *a2a.TaskStatusUpdateEvent) error {
	return e.EmitStatusUpdateFn(ctx, event)
}

func (e *mockTaskEventEmitter) EmitArtifactUpdate(ctx context.Context, event *a2a.TaskArtifactUpdateEvent) error {
	return e.EmitArtifactUpdateFn(ctx, event)
}

func (e *mockTaskEventEmitter) EmitTaskCompletion(ctx context.Context, task *a2a.Task) error {
	return e.EmitTaskCompletionFn(ctx, task)
}

// TestA2AServer_HandleJSONRPC tests the A2A server's JSON-RPC request handling.
func TestA2AServer_HandleJSONRPC(t *testing.T) {
	// Create test data
	agentCard := &a2a.AgentCard{
		Name:                 "TestAgent",
		Description:          "A test agent for unit tests",
		Version:              "1.0.0",
		SupportedOutputModes: []string{"text"},
		Capabilities: a2a.AgentCapabilities{
			Streaming:         true,
			PushNotifications: false,
		},
	}

	taskID := uuid.New().String()
	task := &a2a.Task{
		ID: taskID,
		Status: a2a.TaskStatus{
			State:     a2a.TaskStateCompleted,
			Timestamp: time.Now(),
		},
		Artifacts: []a2a.Artifact{
			{
				Parts: []a2a.Part{
					a2a.TextPart{Text: "Hello, user!"},
				},
				Index: 0,
			},
		},
	}

	// Create mock components
	mockHandler := &mockA2AHandler{
		GetAgentCardFn: func(ctx context.Context) (*a2a.AgentCard, error) {
			return agentCard, nil
		},
		TaskSendFn: func(ctx context.Context, params *a2a.TaskSendParams) (*a2a.Task, error) {
			return task, nil
		},
		TaskGetFn: func(ctx context.Context, params *a2a.TaskGetParams) (*a2a.Task, error) {
			return task, nil
		},
		TaskCancelFn: func(ctx context.Context, params *a2a.TaskCancelParams) (*a2a.Task, error) {
			task.Status.State = a2a.TaskStateCanceled
			return task, nil
		},
		SetPushNotificationFn: func(ctx context.Context, params *a2a.SetPushNotificationParams) (any, error) {
			return struct{}{}, nil
		},
	}

	mockStore := &mockTaskStore{
		GetTaskFn: func(ctx context.Context, id string) (*a2a.Task, error) {
			return task, nil
		},
		CreateTaskFn: func(ctx context.Context, t *a2a.Task) error {
			return nil
		},
		UpdateTaskFn: func(ctx context.Context, t *a2a.Task) error {
			return nil
		},
		DeleteTaskFn: func(ctx context.Context, id string) error {
			return nil
		},
		ListTasksFn: func(ctx context.Context) ([]*a2a.Task, error) {
			return []*a2a.Task{task}, nil
		},
	}

	mockEmitter := &mockTaskEventEmitter{
		EmitStatusUpdateFn: func(ctx context.Context, event *a2a.TaskStatusUpdateEvent) error {
			return nil
		},
		EmitArtifactUpdateFn: func(ctx context.Context, event *a2a.TaskArtifactUpdateEvent) error {
			return nil
		},
		EmitTaskCompletionFn: func(ctx context.Context, t *a2a.Task) error {
			return nil
		},
	}

	// Create the server
	a2aServer := server.NewA2AServer(":8080", mockHandler, mockStore, mockEmitter)

	// Test agent/getCard
	t.Run("agent/getCard", func(t *testing.T) {
		// Create a request
		req := a2a.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      "1",
			Method:  "agent/getCard",
			Params:  json.RawMessage("{}"),
		}
		reqJSON, _ := json.Marshal(req)

		// Create an HTTP request
		httpReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(reqJSON)))
		httpReq.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		// Handle the request
		a2aServer.HTTPServer.Handler.ServeHTTP(recorder, httpReq)

		// Check the response
		if recorder.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
		}

		var resp a2a.JSONRPCResponse
		if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp.Error != nil {
			t.Errorf("Expected no error, got %v", resp.Error)
		}

		var resultCard a2a.AgentCard
		resultJSON, _ := json.Marshal(resp.Result)
		if err := json.Unmarshal(resultJSON, &resultCard); err != nil {
			t.Fatalf("Failed to decode result: %v", err)
		}

		opts := cmp.Options{
			cmpopts.IgnoreUnexported(a2a.AgentCard{}, a2a.AgentCapabilities{}),
		}
		if diff := cmp.Diff(*agentCard, resultCard, opts); diff != "" {
			t.Errorf("Agent card mismatch (-want +got):\n%s", diff)
		}
	})

	// Test tasks/send
	t.Run("tasks/send", func(t *testing.T) {
		// Create a request
		params := a2a.TaskSendParams{
			ID: taskID,
			Message: a2a.Message{
				Role: a2a.RoleUser,
				Parts: []a2a.Part{
					a2a.TextPart{Text: "Hello, agent!"},
				},
			},
			AcceptedOutputModes: []string{"text"},
		}
		paramsJSON, _ := json.Marshal(params)

		req := a2a.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      "2",
			Method:  "tasks/send",
			Params:  paramsJSON,
		}
		reqJSON, _ := json.Marshal(req)

		// Create an HTTP request
		httpReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(reqJSON)))
		httpReq.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		// Handle the request
		a2aServer.HTTPServer.Handler.ServeHTTP(recorder, httpReq)

		// Check the response
		if recorder.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
		}

		var resp a2a.JSONRPCResponse
		if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp.Error != nil {
			t.Errorf("Expected no error, got %v", resp.Error)
		}

		var resultTask *a2a.Task
		resultJSON, _ := json.Marshal(resp.Result)
		if err := json.Unmarshal(resultJSON, resultTask); err != nil {
			t.Fatalf("Failed to decode result: %v", err)
		}

		opts := cmp.Options{
			cmpopts.IgnoreUnexported(a2a.Task{}, a2a.TaskStatus{}, a2a.Artifact{}),
			cmpopts.IgnoreFields(a2a.TaskStatus{}, "Timestamp"),
		}
		if diff := cmp.Diff(task, resultTask, opts); diff != "" {
			t.Errorf("Task mismatch (-want +got):\n%s", diff)
		}
	})

	// Test tasks/get
	t.Run("tasks/get", func(t *testing.T) {
		// Create a request
		params := a2a.TaskGetParams{
			ID: taskID,
		}
		paramsJSON, _ := json.Marshal(params)

		req := a2a.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      "3",
			Method:  "tasks/get",
			Params:  paramsJSON,
		}
		reqJSON, _ := json.Marshal(req)

		// Create an HTTP request
		httpReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(reqJSON)))
		httpReq.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		// Handle the request
		a2aServer.HTTPServer.Handler.ServeHTTP(recorder, httpReq)

		// Check the response
		if recorder.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
		}

		var resp a2a.JSONRPCResponse
		if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp.Error != nil {
			t.Errorf("Expected no error, got %v", resp.Error)
		}

		var resultTask a2a.Task
		resultJSON, _ := json.Marshal(resp.Result)
		if err := json.Unmarshal(resultJSON, &resultTask); err != nil {
			t.Fatalf("Failed to decode result: %v", err)
		}

		opts := cmp.Options{
			cmpopts.IgnoreUnexported(a2a.Task{}, a2a.TaskStatus{}, a2a.Artifact{}),
			cmpopts.IgnoreFields(a2a.TaskStatus{}, "Timestamp"),
		}
		if diff := cmp.Diff(*task, resultTask, opts); diff != "" {
			t.Errorf("Task mismatch (-want +got):\n%s", diff)
		}
	})

	// Test tasks/cancel
	t.Run("tasks/cancel", func(t *testing.T) {
		// Create a request
		params := a2a.TaskCancelParams{
			ID: taskID,
		}
		paramsJSON, _ := json.Marshal(params)

		req := a2a.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      "4",
			Method:  "tasks/cancel",
			Params:  paramsJSON,
		}
		reqJSON, _ := json.Marshal(req)

		// Create an HTTP request
		httpReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(reqJSON)))
		httpReq.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		// Handle the request
		a2aServer.HTTPServer.Handler.ServeHTTP(recorder, httpReq)

		// Check the response
		if recorder.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
		}

		var resp a2a.JSONRPCResponse
		if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp.Error != nil {
			t.Errorf("Expected no error, got %v", resp.Error)
		}

		var resultTask a2a.Task
		resultJSON, _ := json.Marshal(resp.Result)
		if err := json.Unmarshal(resultJSON, &resultTask); err != nil {
			t.Fatalf("Failed to decode result: %v", err)
		}

		if resultTask.Status.State != a2a.TaskStateCanceled {
			t.Errorf("Expected task state %s, got %s", a2a.TaskStateCanceled, resultTask.Status.State)
		}
	})

	// Test invalid method
	t.Run("invalid method", func(t *testing.T) {
		// Create a request
		req := a2a.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      "5",
			Method:  "invalid/method",
			Params:  json.RawMessage("{}"),
		}
		reqJSON, _ := json.Marshal(req)

		// Create an HTTP request
		httpReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(reqJSON)))
		httpReq.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		// Handle the request
		a2aServer.HTTPServer.Handler.ServeHTTP(recorder, httpReq)

		// Check the response
		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, recorder.Code)
		}

		var resp a2a.JSONRPCResponse
		if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp.Error == nil {
			t.Errorf("Expected error, got nil")
		}
		if resp.Error.Code != -32603 {
			t.Errorf("Expected error code %d, got %d", -32603, resp.Error.Code)
		}
	})

	// Test invalid JSON
	t.Run("invalid JSON", func(t *testing.T) {
		// Create an HTTP request with invalid JSON
		httpReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{invalid json}"))
		httpReq.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		// Handle the request
		a2aServer.HTTPServer.Handler.ServeHTTP(recorder, httpReq)

		// Check the response
		if recorder.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, recorder.Code)
		}

		var resp a2a.JSONRPCResponse
		if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp.Error == nil {
			t.Errorf("Expected error, got nil")
		}
		if resp.Error.Code != -32700 {
			t.Errorf("Expected error code %d, got %d", -32700, resp.Error.Code)
		}
	})
}

// TestSubscriptionManager tests the subscription manager.
func TestSubscriptionManager(t *testing.T) {
	// Create a subscription manager
	manager := server.NewSubscriptionManager()

	// Test creating a subscription
	t.Run("CreateSubscription", func(t *testing.T) {
		taskID := uuid.New().String()

		subscription, err := manager.CreateSubscription(taskID)
		if err != nil {
			t.Fatalf("CreateSubscription failed: %v", err)
		}

		if subscription.TaskID != taskID {
			t.Errorf("Expected task ID %s, got %s", taskID, subscription.TaskID)
		}
		if subscription.StatusChannel == nil {
			t.Errorf("Expected non-nil status channel")
		}
		if subscription.ArtifactChannel == nil {
			t.Errorf("Expected non-nil artifact channel")
		}
	})

	// Test getting a subscription
	t.Run("GetSubscription", func(t *testing.T) {
		taskID := uuid.New().String()

		// Create a subscription
		subscription, err := manager.CreateSubscription(taskID)
		if err != nil {
			t.Fatalf("CreateSubscription failed: %v", err)
		}

		// Get the subscription
		retrievedSubscription, err := manager.GetSubscription(taskID)
		if err != nil {
			t.Fatalf("GetSubscription failed: %v", err)
		}

		if retrievedSubscription != subscription {
			t.Errorf("Expected subscription %v, got %v", subscription, retrievedSubscription)
		}
	})

	// Test deleting a subscription
	t.Run("DeleteSubscription", func(t *testing.T) {
		taskID := uuid.New().String()

		// Create a subscription
		_, err := manager.CreateSubscription(taskID)
		if err != nil {
			t.Fatalf("CreateSubscription failed: %v", err)
		}

		// Delete the subscription
		manager.DeleteSubscription(taskID)

		// Try to get the deleted subscription
		_, err = manager.GetSubscription(taskID)
		if err == nil {
			t.Errorf("Expected error getting deleted subscription, got nil")
		}
	})
}

// TestDefaultA2AHandler tests the default A2A handler.
func TestDefaultA2AHandler(t *testing.T) {
	// Create test data
	agentCard := &a2a.AgentCard{
		Name:                 "TestAgent",
		Description:          "A test agent for unit tests",
		Version:              "1.0.0",
		SupportedOutputModes: []string{"text"},
		Capabilities: a2a.AgentCapabilities{
			Streaming:         true,
			PushNotifications: false,
		},
	}

	// Create a task store
	taskStore := server.NewInMemoryTaskStore()

	// Create a subscription manager
	subscriptionManager := server.NewSubscriptionManager()

	// Create a task event emitter
	eventEmitter := server.NewDefaultTaskEventEmitter(subscriptionManager)

	// Create a task callback
	taskCallbackCalled := false
	taskCallback := func(ctx context.Context, task *a2a.Task) error {
		taskCallbackCalled = true
		task.Status.State = a2a.TaskStateCompleted
		task.Status.Timestamp = time.Now()
		task.Artifacts = []a2a.Artifact{
			{
				Parts: []a2a.Part{
					a2a.TextPart{Text: "Hello, user!"},
				},
				Index: 0,
			},
		}
		return nil
	}

	// Create the handler
	handler := server.NewDefaultA2AHandler(agentCard, taskStore, eventEmitter, taskCallback)

	// Test GetAgentCard
	t.Run("GetAgentCard", func(t *testing.T) {
		card, err := handler.GetAgentCard(context.Background())
		if err != nil {
			t.Fatalf("GetAgentCard failed: %v", err)
		}

		if diff := cmp.Diff(agentCard, card,
			cmpopts.IgnoreUnexported(a2a.AgentCard{}, a2a.AgentCapabilities{})); diff != "" {
			t.Errorf("Agent card mismatch (-want +got):\n%s", diff)
		}
	})

	// Test TaskSend
	t.Run("TaskSend", func(t *testing.T) {
		// Create task parameters
		taskID := uuid.New().String()
		params := &a2a.TaskSendParams{
			ID: taskID,
			Message: a2a.Message{
				Role: a2a.RoleUser,
				Parts: []a2a.Part{
					a2a.TextPart{Text: "Hello, agent!"},
				},
			},
			AcceptedOutputModes: []string{"text"},
		}

		// Reset the flag
		taskCallbackCalled = false

		// Call TaskSend
		task, err := handler.TaskSend(context.Background(), params)
		if err != nil {
			t.Fatalf("TaskSend failed: %v", err)
		}

		// Verify the task
		if task.ID != taskID {
			t.Errorf("Expected task ID %s, got %s", taskID, task.ID)
		}
		if task.Status.State != a2a.TaskStateCompleted {
			t.Errorf("Expected task state %s, got %s", a2a.TaskStateCompleted, task.Status.State)
		}
		if len(task.Artifacts) != 1 {
			t.Errorf("Expected 1 artifact, got %d", len(task.Artifacts))
		}
		if !taskCallbackCalled {
			t.Errorf("Task callback was not called")
		}
	})

	// Test TaskGet
	t.Run("TaskGet", func(t *testing.T) {
		// Create a task
		taskID := uuid.New().String()
		task := &a2a.Task{
			ID: taskID,
			Status: a2a.TaskStatus{
				State:     a2a.TaskStateCompleted,
				Timestamp: time.Now(),
			},
			Artifacts: []a2a.Artifact{
				{
					Parts: []a2a.Part{
						a2a.TextPart{Text: "Hello, user!"},
					},
					Index: 0,
				},
			},
		}
		if err := taskStore.CreateTask(context.Background(), task); err != nil {
			t.Fatalf("Failed to create test task: %v", err)
		}

		// Call TaskGet
		params := &a2a.TaskGetParams{ID: taskID}
		retrievedTask, err := handler.TaskGet(context.Background(), params)
		if err != nil {
			t.Fatalf("TaskGet failed: %v", err)
		}

		// Verify the task
		opts := cmp.Options{
			cmpopts.IgnoreUnexported(a2a.Task{}, a2a.TaskStatus{}, a2a.Artifact{}),
			cmpopts.IgnoreFields(a2a.TaskStatus{}, "Timestamp"),
		}
		if diff := cmp.Diff(task, retrievedTask, opts); diff != "" {
			t.Errorf("Task mismatch (-want +got):\n%s", diff)
		}
	})

	// Test TaskCancel
	t.Run("TaskCancel", func(t *testing.T) {
		// Create a task
		taskID := uuid.New().String()
		task := &a2a.Task{
			ID: taskID,
			Status: a2a.TaskStatus{
				State:     a2a.TaskStateWorking,
				Timestamp: time.Now(),
			},
		}
		if err := taskStore.CreateTask(context.Background(), task); err != nil {
			t.Fatalf("Failed to create test task: %v", err)
		}

		// Call TaskCancel
		params := &a2a.TaskCancelParams{ID: taskID}
		canceledTask, err := handler.TaskCancel(context.Background(), params)
		if err != nil {
			t.Fatalf("TaskCancel failed: %v", err)
		}

		// Verify the task
		if canceledTask.Status.State != a2a.TaskStateCanceled {
			t.Errorf("Expected task state %s, got %s", a2a.TaskStateCanceled, canceledTask.Status.State)
		}
	})

	// Test SetPushNotification
	t.Run("SetPushNotification", func(t *testing.T) {
		// Create push notification config
		config := &a2a.PushNotificationConfig{
			URL: "https://example.com/webhook",
			Authentication: &a2a.AgentAuthentication{
				Schemes: []string{"jwt+kid123"},
			},
		}
		params := &a2a.SetPushNotificationParams{Config: *config}

		// Call SetPushNotification
		_, err := handler.SetPushNotification(context.Background(), params)
		if err != nil {
			t.Fatalf("SetPushNotification failed: %v", err)
		}

		// Verify the config was set
		if handler.PushNotificationConfig == nil {
			t.Fatalf("Push notification config was not set")
		}
		if handler.PushNotificationConfig.URL != config.URL {
			t.Errorf("Expected URL %s, got %s", config.URL, handler.PushNotificationConfig.URL)
		}
	})
}
