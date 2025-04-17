package a2a_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/go-a2a/a2a"
)

func TestMessageMarshalUnmarshal(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name    string
		message a2a.Message
	}{
		{
			name: "Text message",
			message: a2a.Message{
				Role: a2a.RoleUser,
				Parts: []a2a.Part{
					&a2a.TextPart{Text: "Hello, world!"},
				},
			},
		},
		{
			name: "File message",
			message: a2a.Message{
				Role: a2a.RoleAgent,
				Parts: []a2a.Part{
					&a2a.FilePart{
						MimeType: "image/png",
						FileName: "image.png",
						Data:     "base64data",
					},
				},
			},
		},
		{
			name: "Data message",
			message: a2a.Message{
				Role: a2a.RoleUser,
				Parts: []a2a.Part{
					&a2a.DataPart{
						Data: json.RawMessage(`{"key":"value"}`),
					},
				},
			},
		},
		{
			name: "Mixed parts message",
			message: a2a.Message{
				Role: a2a.RoleAgent,
				Parts: []a2a.Part{
					&a2a.TextPart{Text: "Here's a file:"},
					&a2a.FilePart{
						MimeType: "application/pdf",
						FileName: "document.pdf",
						URI:      "https://example.com/document.pdf",
					},
					&a2a.DataPart{
						Data: json.RawMessage(`{"status":"success"}`),
					},
				},
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Marshal the message
			data, err := json.Marshal(tc.message)
			if err != nil {
				t.Fatalf("Failed to marshal message: %v", err)
			}

			// Unmarshal the message
			var unmarshalled a2a.Message
			if err := json.Unmarshal(data, &unmarshalled); err != nil {
				t.Fatalf("Failed to unmarshal message: %v", err)
			}

			// Compare the original and unmarshalled messages
			// Using go-cmp for better error messages
			if diff := cmp.Diff(tc.message, unmarshalled,
				cmpopts.IgnoreUnexported(a2a.Message{})); diff != "" {
				t.Errorf("Message mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestArtifactMarshalUnmarshal(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		artifact a2a.Artifact
	}{
		{
			name: "Text artifact",
			artifact: a2a.Artifact{
				ID:    "artifact-1",
				Name:  "Simple Text",
				Index: 0,
				Parts: []a2a.Part{
					&a2a.TextPart{Text: "This is a text artifact."},
				},
			},
		},
		{
			name: "File artifact",
			artifact: a2a.Artifact{
				ID:    "artifact-2",
				Name:  "Image File",
				Index: 1,
				Parts: []a2a.Part{
					&a2a.FilePart{
						MimeType: "image/jpeg",
						FileName: "photo.jpg",
						Data:     "base64encodeddata",
					},
				},
			},
		},
		{
			name: "Data artifact",
			artifact: a2a.Artifact{
				ID:    "artifact-3",
				Name:  "JSON Data",
				Index: 2,
				Parts: []a2a.Part{
					&a2a.DataPart{
						Data: json.RawMessage(`{"results":[1,2,3]}`),
					},
				},
			},
		},
		{
			name: "Complex artifact",
			artifact: a2a.Artifact{
				ID:    "artifact-4",
				Name:  "Mixed Content",
				Index: 3,
				Parts: []a2a.Part{
					&a2a.TextPart{Text: "Analysis results:"},
					&a2a.DataPart{
						Data: json.RawMessage(`{"summary":{"count":42,"status":"complete"}}`),
					},
					&a2a.FilePart{
						MimeType: "application/pdf",
						FileName: "report.pdf",
						URI:      "https://example.com/reports/42.pdf",
					},
				},
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Marshal the artifact
			data, err := json.Marshal(tc.artifact)
			if err != nil {
				t.Fatalf("Failed to marshal artifact: %v", err)
			}

			// Unmarshal the artifact
			var unmarshalled a2a.Artifact
			if err := json.Unmarshal(data, &unmarshalled); err != nil {
				t.Fatalf("Failed to unmarshal artifact: %v", err)
			}

			// Compare the original and unmarshalled artifacts
			if diff := cmp.Diff(tc.artifact, unmarshalled,
				cmpopts.IgnoreUnexported(a2a.Artifact{})); diff != "" {
				t.Errorf("Artifact mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTaskStateMachine(t *testing.T) {
	// Define test case for task state transitions
	t.Run("Valid task state transitions", func(t *testing.T) {
		// Create a task in submitted state
		task := &a2a.Task{
			ID: "task-1",
			Status: a2a.TaskStatus{
				State:     a2a.TaskStateSubmitted,
				Timestamp: time.Now(),
			},
			History: []a2a.Message{
				{
					Role: a2a.RoleUser,
					Parts: []a2a.Part{
						&a2a.TextPart{Text: "Hello, agent!"},
					},
				},
			},
		}

		// Check initial state
		if task.Status.State != a2a.TaskStateSubmitted {
			t.Errorf("Expected initial state to be %s, got %s", a2a.TaskStateSubmitted, task.Status.State)
		}

		// Transition to working state
		task.Status.State = a2a.TaskStateWorking
		task.Status.Timestamp = time.Now()

		// Check working state
		if task.Status.State != a2a.TaskStateWorking {
			t.Errorf("Expected state to be %s, got %s", a2a.TaskStateWorking, task.Status.State)
		}

		// Transition to input-required state
		task.Status.State = a2a.TaskStateInputRequired
		task.Status.Timestamp = time.Now()
		task.Status.Message = "Please provide more information."

		// Check input-required state
		if task.Status.State != a2a.TaskStateInputRequired {
			t.Errorf("Expected state to be %s, got %s", a2a.TaskStateInputRequired, task.Status.State)
		}

		// Add a message from the agent
		task.History = append(task.History, a2a.Message{
			Role: a2a.RoleAgent,
			Parts: []a2a.Part{
				&a2a.TextPart{Text: "Could you please provide more details?"},
			},
		})

		// Add a response from the user
		task.History = append(task.History, a2a.Message{
			Role: a2a.RoleUser,
			Parts: []a2a.Part{
				&a2a.TextPart{Text: "I need help with Go programming."},
			},
		})

		// Transition back to working state
		task.Status.State = a2a.TaskStateWorking
		task.Status.Timestamp = time.Now()
		task.Status.Message = ""

		// Transition to completed state
		task.Status.State = a2a.TaskStateCompleted
		task.Status.Timestamp = time.Now()

		// Add an artifact
		task.Artifacts = []a2a.Artifact{
			{
				ID:    "artifact-1",
				Index: 0,
				Parts: []a2a.Part{
					&a2a.TextPart{Text: "Here's some help with Go programming..."},
				},
			},
		}

		// Check completed state
		if task.Status.State != a2a.TaskStateCompleted {
			t.Errorf("Expected state to be %s, got %s", a2a.TaskStateCompleted, task.Status.State)
		}

		// Add a final message from the agent
		task.History = append(task.History, a2a.Message{
			Role: a2a.RoleAgent,
			Parts: []a2a.Part{
				&a2a.TextPart{Text: "I've provided some information about Go programming. Let me know if you need anything else!"},
			},
		})

		// Marshal and unmarshal the task
		data, err := json.Marshal(task)
		if err != nil {
			t.Fatalf("Failed to marshal task: %v", err)
		}

		var unmarshalled a2a.Task
		if err := json.Unmarshal(data, &unmarshalled); err != nil {
			t.Fatalf("Failed to unmarshal task: %v", err)
		}

		// Compare the original and unmarshalled tasks
		if diff := cmp.Diff(task, &unmarshalled,
			cmpopts.IgnoreUnexported(a2a.Task{}, a2a.Message{}, a2a.Artifact{}),
			cmpopts.IgnoreFields(a2a.TaskStatus{}, "Timestamp")); diff != "" {
			t.Errorf("Task mismatch (-want +got):\n%s", diff)
		}
	})

	// Test invalid state transitions (this is more of a documentation test since
	// the types package doesn't enforce state transitions)
	t.Run("Document valid state transitions", func(t *testing.T) {
		// This is a table documenting valid state transitions
		validTransitions := map[a2a.TaskState][]a2a.TaskState{
			a2a.TaskStateSubmitted: {
				a2a.TaskStateWorking,
				a2a.TaskStateFailed,
				a2a.TaskStateCanceled,
			},
			a2a.TaskStateWorking: {
				a2a.TaskStateInputRequired,
				a2a.TaskStateCompleted,
				a2a.TaskStateFailed,
				a2a.TaskStateCanceled,
			},
			a2a.TaskStateInputRequired: {
				a2a.TaskStateWorking,
				a2a.TaskStateFailed,
				a2a.TaskStateCanceled,
			},
			a2a.TaskStateCompleted: {}, // Terminal state
			a2a.TaskStateFailed:    {}, // Terminal state
			a2a.TaskStateCanceled:  {}, // Terminal state
		}

		// Verify that all states have defined transitions
		allStates := []a2a.TaskState{
			a2a.TaskStateSubmitted,
			a2a.TaskStateWorking,
			a2a.TaskStateInputRequired,
			a2a.TaskStateCompleted,
			a2a.TaskStateFailed,
			a2a.TaskStateCanceled,
		}

		for _, state := range allStates {
			if _, ok := validTransitions[state]; !ok {
				t.Errorf("State %s is missing from valid transitions table", state)
			}
		}

		// This test doesn't actually validate anything at runtime,
		// it's more documentation for developers about valid state transitions
	})
}

func TestJSONRPCTypes(t *testing.T) {
	// Test JSON-RPC request/response marshalling
	t.Run("JSON-RPC request", func(t *testing.T) {
		// Create a JSON-RPC request
		req := a2a.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      "request-1",
			Method:  "tasks/send",
			Params:  json.RawMessage(`{"id":"task-1","message":{"role":"user","parts":[{"type":"text","text":"Hello!"}]}}`),
		}

		// Marshal the request
		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("Failed to marshal JSON-RPC request: %v", err)
		}

		// Unmarshal the request
		var unmarshalled a2a.JSONRPCRequest
		if err := json.Unmarshal(data, &unmarshalled); err != nil {
			t.Fatalf("Failed to unmarshal JSON-RPC request: %v", err)
		}

		// Compare the original and unmarshalled requests
		if diff := cmp.Diff(req, unmarshalled,
			cmpopts.IgnoreUnexported(a2a.JSONRPCRequest{})); diff != "" {
			t.Errorf("JSON-RPC request mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("JSON-RPC response", func(t *testing.T) {
		// Create a JSON-RPC success response
		successResp := a2a.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      "request-1",
			Result: map[string]any{
				"id":     string("task-1"),
				"status": map[string]any{"state": string("completed"), "timestamp": string("2023-01-01T12:00:00Z")},
			},
			// Result:  json.RawMessage(`{"id":"task-1","status":{"state":"completed","timestamp":"2023-01-01T12:00:00Z"}}`),
		}

		// Marshal the success response
		successData, err := json.Marshal(successResp)
		if err != nil {
			t.Fatalf("Failed to marshal JSON-RPC success response: %v", err)
		}

		// Unmarshal the success response
		var unmarshalledSuccess a2a.JSONRPCResponse
		if err := json.Unmarshal(successData, &unmarshalledSuccess); err != nil {
			t.Fatalf("Failed to unmarshal JSON-RPC success response: %v", err)
		}

		// Compare the original and unmarshalled success responses
		if diff := cmp.Diff(successResp, unmarshalledSuccess,
			cmpopts.IgnoreUnexported(a2a.JSONRPCResponse{})); diff != "" {
			t.Errorf("JSON-RPC success response mismatch (-want +got):\n%s", diff)
		}

		// Create a JSON-RPC error response
		errorResp := a2a.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      "request-1",
			Error: &a2a.JSONRPCError{
				Code:    -32600,
				Message: "Invalid Request",
				Data:    "Additional error information",
			},
		}

		// Marshal the error response
		errorData, err := json.Marshal(errorResp)
		if err != nil {
			t.Fatalf("Failed to marshal JSON-RPC error response: %v", err)
		}

		// Unmarshal the error response
		var unmarshalledError a2a.JSONRPCResponse
		if err := json.Unmarshal(errorData, &unmarshalledError); err != nil {
			t.Fatalf("Failed to unmarshal JSON-RPC error response: %v", err)
		}

		// Compare the original and unmarshalled error responses
		if diff := cmp.Diff(errorResp, unmarshalledError,
			cmpopts.IgnoreUnexported(a2a.JSONRPCResponse{}, a2a.JSONRPCError{})); diff != "" {
			t.Errorf("JSON-RPC error response mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestAgentCardMarshalUnmarshal(t *testing.T) {
	// Create an agent card
	card := a2a.AgentCard{
		Name:                 "TestAgent",
		Description:          "A test agent for unit tests",
		Version:              "1.0.0",
		Contact:              "test@example.com",
		SupportedOutputModes: []string{"text", "file", "data"},
		MetadataEndpoint:     "https://example.com/agent/metadata",
		Capabilities: a2a.AgentCapabilities{
			Streaming:         true,
			PushNotifications: true,
		},
		Authentication: &a2a.AgentAuthentication{
			Schemes: []string{"jwt+kid123"},
		},
	}

	// Marshal the card
	data, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("Failed to marshal agent card: %v", err)
	}

	// Unmarshal the card
	var unmarshalled a2a.AgentCard
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal agent card: %v", err)
	}

	// Compare the original and unmarshalled cards
	if diff := cmp.Diff(card, unmarshalled,
		cmpopts.IgnoreUnexported(a2a.AgentCard{}, a2a.AgentCapabilities{}, a2a.AgentAuthentication{})); diff != "" {
		t.Errorf("Agent card mismatch (-want +got):\n%s", diff)
	}
}
