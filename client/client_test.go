package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/client"
)

// TestA2AClient_GetAgentCard tests the GetAgentCard method.
func TestA2AClient_GetAgentCard(t *testing.T) {
	// Create a test agent card
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

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Decode the request
		var req a2a.JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Verify request fields
		if req.JSONRPC != "2.0" {
			t.Errorf("Expected JSONRPC version 2.0, got %s", req.JSONRPC)
		}
		if req.Method != "agent/getCard" {
			t.Errorf("Expected method agent/getCard, got %s", req.Method)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := a2a.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  agentCard,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create a client
	a2aClient := client.NewA2AClient(server.URL)

	// Call GetAgentCard
	card, err := a2aClient.GetAgentCard(context.Background())
	if err != nil {
		t.Fatalf("GetAgentCard failed: %v", err)
	}

	// Verify the result
	if diff := cmp.Diff(agentCard, card,
		cmpopts.IgnoreUnexported(a2a.AgentCard{}, a2a.AgentCapabilities{})); diff != "" {
		t.Errorf("Agent card mismatch (-want +got):\n%s", diff)
	}
}

// TestA2AClient_TaskSend tests the TaskSend method.
func TestA2AClient_TaskSend(t *testing.T) {
	// Create test data
	taskID := uuid.New().String()
	sessionID := uuid.New().String()
	message := a2a.Message{
		Role: a2a.RoleUser,
		Parts: []a2a.Part{
			a2a.TextPart{Text: "Hello, agent!"},
		},
	}
	params := &a2a.TaskSendParams{
		ID:                  taskID,
		Message:             message,
		SessionID:           sessionID,
		AcceptedOutputModes: []string{"text"},
	}

	// Create expected response
	expectedTask := &a2a.Task{
		ID: taskID,
		Status: a2a.TaskStatus{
			State:     a2a.TaskStateCompleted,
			Timestamp: time.Now(),
		},
		History: []a2a.Message{message},
		Artifacts: []a2a.Artifact{
			{
				Parts: []a2a.Part{
					a2a.TextPart{Text: "Hello, user!"},
				},
				Index: 0,
			},
		},
		SessionID: sessionID,
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Decode the request
		var req a2a.JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Verify request fields
		if req.Method != "tasks/send" {
			t.Errorf("Expected method tasks/send, got %s", req.Method)
		}

		// Decode params
		var reqParams a2a.TaskSendParams
		if err := json.Unmarshal(req.Params, &reqParams); err != nil {
			t.Errorf("Failed to decode params: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Verify params
		if reqParams.ID != taskID {
			t.Errorf("Expected task ID %s, got %s", taskID, reqParams.ID)
		}
		if reqParams.SessionID != sessionID {
			t.Errorf("Expected session ID %s, got %s", sessionID, reqParams.SessionID)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := a2a.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  expectedTask,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create a client
	a2aClient := client.NewA2AClient(server.URL)

	// Call TaskSend
	task, err := a2aClient.TaskSend(context.Background(), params)
	if err != nil {
		t.Fatalf("TaskSend failed: %v", err)
	}

	// Verify the result
	opts := cmp.Options{
		cmpopts.IgnoreUnexported(a2a.Task{}, a2a.TaskStatus{}, a2a.Message{}, a2a.Artifact{}),
		cmpopts.IgnoreFields(a2a.TaskStatus{}, "Timestamp"),
	}
	if diff := cmp.Diff(expectedTask, task, opts); diff != "" {
		t.Errorf("Task mismatch (-want +got):\n%s", diff)
	}
}

// TestA2AClient_TaskGet tests the TaskGet method.
func TestA2AClient_TaskGet(t *testing.T) {
	// Create test data
	taskID := uuid.New().String()
	sessionID := uuid.New().String()

	// Create expected response
	expectedTask := &a2a.Task{
		ID: taskID,
		Status: a2a.TaskStatus{
			State:     a2a.TaskStateCompleted,
			Timestamp: time.Now(),
		},
		History: []a2a.Message{
			{
				Role: a2a.RoleUser,
				Parts: []a2a.Part{
					a2a.TextPart{Text: "Hello, agent!"},
				},
			},
			{
				Role: a2a.RoleAgent,
				Parts: []a2a.Part{
					a2a.TextPart{Text: "Hello, user!"},
				},
			},
		},
		Artifacts: []a2a.Artifact{
			{
				Parts: []a2a.Part{
					a2a.TextPart{Text: "Hello, user!"},
				},
				Index: 0,
			},
		},
		SessionID: sessionID,
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Decode the request
		var req a2a.JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Verify request fields
		if req.Method != "tasks/get" {
			t.Errorf("Expected method tasks/get, got %s", req.Method)
		}

		// Decode params
		var reqParams a2a.TaskGetParams
		if err := json.Unmarshal(req.Params, &reqParams); err != nil {
			t.Errorf("Failed to decode params: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Verify params
		if reqParams.ID != taskID {
			t.Errorf("Expected task ID %s, got %s", taskID, reqParams.ID)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := a2a.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  expectedTask,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create a client
	a2aClient := client.NewA2AClient(server.URL)

	// Call TaskGet
	task, err := a2aClient.TaskGet(context.Background(), taskID)
	if err != nil {
		t.Fatalf("TaskGet failed: %v", err)
	}

	// Verify the result
	opts := cmp.Options{
		cmpopts.IgnoreUnexported(a2a.Task{}, a2a.TaskStatus{}, a2a.Message{}, a2a.Artifact{}),
		cmpopts.IgnoreFields(a2a.TaskStatus{}, "Timestamp"),
	}
	if diff := cmp.Diff(expectedTask, task, opts); diff != "" {
		t.Errorf("Task mismatch (-want +got):\n%s", diff)
	}
}

// TestA2AClient_TaskCancel tests the TaskCancel method.
func TestA2AClient_TaskCancel(t *testing.T) {
	// Create test data
	taskID := uuid.New().String()
	sessionID := uuid.New().String()

	// Create expected response
	expectedTask := &a2a.Task{
		ID: taskID,
		Status: a2a.TaskStatus{
			State:     a2a.TaskStateCanceled,
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
		SessionID: sessionID,
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Decode the request
		var req a2a.JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Verify request fields
		if req.Method != "tasks/cancel" {
			t.Errorf("Expected method tasks/cancel, got %s", req.Method)
		}

		// Decode params
		var reqParams a2a.TaskCancelParams
		if err := json.Unmarshal(req.Params, &reqParams); err != nil {
			t.Errorf("Failed to decode params: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Verify params
		if reqParams.ID != taskID {
			t.Errorf("Expected task ID %s, got %s", taskID, reqParams.ID)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := a2a.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  expectedTask,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create a client
	a2aClient := client.NewA2AClient(server.URL)

	// Call TaskCancel
	task, err := a2aClient.TaskCancel(context.Background(), taskID)
	if err != nil {
		t.Fatalf("TaskCancel failed: %v", err)
	}

	// Verify the result
	opts := cmp.Options{
		cmpopts.IgnoreUnexported(a2a.Task{}, a2a.TaskStatus{}, a2a.Message{}),
		cmpopts.IgnoreFields(a2a.TaskStatus{}, "Timestamp"),
	}
	if diff := cmp.Diff(expectedTask, task, opts); diff != "" {
		t.Errorf("Task mismatch (-want +got):\n%s", diff)
	}
}

// TestA2ACardResolver tests the A2ACardResolver.
func TestA2ACardResolver(t *testing.T) {
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

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Decode the request
		var req a2a.JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Verify request fields
		if req.Method != "agent/getCard" {
			t.Errorf("Expected method agent/getCard, got %s", req.Method)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := a2a.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  agentCard,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create a resolver
	resolver := client.NewA2ACardResolver()

	// Call ResolveCard
	card, err := resolver.ResolveCard(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("ResolveCard failed: %v", err)
	}

	// Verify the result
	if diff := cmp.Diff(agentCard, card,
		cmpopts.IgnoreUnexported(a2a.AgentCard{}, a2a.AgentCapabilities{})); diff != "" {
		t.Errorf("Agent card mismatch (-want +got):\n%s", diff)
	}

	// Test cache
	t.Run("Cache hit", func(t *testing.T) {
		// Call ResolveCard again
		cachedCard, err := resolver.ResolveCard(context.Background(), server.URL)
		if err != nil {
			t.Fatalf("ResolveCard (cached) failed: %v", err)
		}

		// Verify the result
		if diff := cmp.Diff(agentCard, cachedCard,
			cmpopts.IgnoreUnexported(a2a.AgentCard{}, a2a.AgentCapabilities{})); diff != "" {
			t.Errorf("Cached agent card mismatch (-want +got):\n%s", diff)
		}
	})

	// Test clearing the cache
	t.Run("Clear cache", func(t *testing.T) {
		// Clear the cache
		resolver.ClearCache()

		// Call ResolveCard again
		cachedCard, err := resolver.ResolveCard(context.Background(), server.URL)
		if err != nil {
			t.Fatalf("ResolveCard (after clear) failed: %v", err)
		}

		// Verify the result
		if diff := cmp.Diff(agentCard, cachedCard,
			cmpopts.IgnoreUnexported(a2a.AgentCard{}, a2a.AgentCapabilities{})); diff != "" {
			t.Errorf("Agent card (after clear) mismatch (-want +got):\n%s", diff)
		}
	})
}
