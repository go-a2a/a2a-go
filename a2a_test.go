package a2a_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/client"
	"github.com/go-a2a/a2a/server"
)

// TestNewTextPart tests the NewTextPart function
func TestNewTextPart(t *testing.T) {
	text := "Hello, world!"
	part := a2a.NewTextPart(text)

	if part.Type != a2a.PartTypeText {
		t.Errorf("Expected part type %s, got %s", a2a.PartTypeText, part.Type)
	}

	if part.Text == nil {
		t.Fatalf("Expected non-nil text, got nil")
	}

	if *part.Text != text {
		t.Errorf("Expected text %q, got %q", text, *part.Text)
	}
}

// TestNewDataPart tests the NewDataPart function
func TestNewDataPart(t *testing.T) {
	data := map[string]string{"key": "value"}
	part := a2a.NewDataPart(data)

	if part.Type != a2a.PartTypeData {
		t.Errorf("Expected part type %s, got %s", a2a.PartTypeData, part.Type)
	}

	if part.Data == nil {
		t.Fatalf("Expected non-nil data, got nil")
	}

	// Compare the data
	jsonData, err := json.Marshal(part.Data)
	if err != nil {
		t.Fatalf("Failed to marshal part data: %v", err)
	}

	expectedJSON, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal expected data: %v", err)
	}

	if string(jsonData) != string(expectedJSON) {
		t.Errorf("Expected data %s, got %s", string(expectedJSON), string(jsonData))
	}
}

// TestNewFilePart tests the NewFilePart function
func TestNewFilePart(t *testing.T) {
	fileBytes := []byte("file content")
	mimeType := "text/plain"
	fileName := "test.txt"

	part := a2a.NewFilePart(fileBytes, mimeType, fileName)

	if part.Type != a2a.PartTypeFile {
		t.Errorf("Expected part type %s, got %s", a2a.PartTypeFile, part.Type)
	}

	if part.FileBytes == nil {
		t.Fatalf("Expected non-nil file bytes, got nil")
	}

	if string(*part.FileBytes) != string(fileBytes) {
		t.Errorf("Expected file bytes %q, got %q", string(fileBytes), string(*part.FileBytes))
	}

	if part.MimeType == nil {
		t.Fatalf("Expected non-nil mime type, got nil")
	}

	if *part.MimeType != mimeType {
		t.Errorf("Expected mime type %q, got %q", mimeType, *part.MimeType)
	}

	if part.Metadata == nil {
		t.Fatalf("Expected non-nil metadata, got nil")
	}

	if part.Metadata.FileName == nil {
		t.Fatalf("Expected non-nil file name, got nil")
	}

	if *part.Metadata.FileName != fileName {
		t.Errorf("Expected file name %q, got %q", fileName, *part.Metadata.FileName)
	}
}

// TestNewFileUriPart tests the NewFileUriPart function
func TestNewFileUriPart(t *testing.T) {
	fileURI := "https://example.com/test.txt"
	mimeType := "text/plain"
	fileName := "test.txt"

	part := a2a.NewFileUriPart(fileURI, mimeType, fileName)

	if part.Type != a2a.PartTypeFile {
		t.Errorf("Expected part type %s, got %s", a2a.PartTypeFile, part.Type)
	}

	if part.FileURI == nil {
		t.Fatalf("Expected non-nil file URI, got nil")
	}

	if *part.FileURI != fileURI {
		t.Errorf("Expected file URI %q, got %q", fileURI, *part.FileURI)
	}

	if part.MimeType == nil {
		t.Fatalf("Expected non-nil mime type, got nil")
	}

	if *part.MimeType != mimeType {
		t.Errorf("Expected mime type %q, got %q", mimeType, *part.MimeType)
	}

	if part.Metadata == nil {
		t.Fatalf("Expected non-nil metadata, got nil")
	}

	if part.Metadata.FileName == nil {
		t.Fatalf("Expected non-nil file name, got nil")
	}

	if *part.Metadata.FileName != fileName {
		t.Errorf("Expected file name %q, got %q", fileName, *part.Metadata.FileName)
	}
}

// TestNewUserMessage tests the NewUserMessage function
func TestNewUserMessage(t *testing.T) {
	part := a2a.NewTextPart("Hello, world!")
	message := a2a.NewUserMessage(part)

	if message.Role != a2a.RoleUser {
		t.Errorf("Expected role %s, got %s", a2a.RoleUser, message.Role)
	}

	if len(message.Parts) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(message.Parts))
	}

	if message.Parts[0].Type != a2a.PartTypeText {
		t.Errorf("Expected part type %s, got %s", a2a.PartTypeText, message.Parts[0].Type)
	}

	if message.Parts[0].Text == nil {
		t.Fatalf("Expected non-nil text, got nil")
	}

	if *message.Parts[0].Text != "Hello, world!" {
		t.Errorf("Expected text %q, got %q", "Hello, world!", *message.Parts[0].Text)
	}
}

// TestNewAgentMessage tests the NewAgentMessage function
func TestNewAgentMessage(t *testing.T) {
	part := a2a.NewTextPart("Hello, world!")
	message := a2a.NewAgentMessage(part)

	if message.Role != a2a.RoleAgent {
		t.Errorf("Expected role %s, got %s", a2a.RoleAgent, message.Role)
	}

	if len(message.Parts) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(message.Parts))
	}

	if message.Parts[0].Type != a2a.PartTypeText {
		t.Errorf("Expected part type %s, got %s", a2a.PartTypeText, message.Parts[0].Type)
	}

	if message.Parts[0].Text == nil {
		t.Fatalf("Expected non-nil text, got nil")
	}

	if *message.Parts[0].Text != "Hello, world!" {
		t.Errorf("Expected text %q, got %q", "Hello, world!", *message.Parts[0].Text)
	}
}

// TestNewArtifact tests the NewArtifact function
func TestNewArtifact(t *testing.T) {
	part := a2a.NewTextPart("Hello, world!")
	title := "Test Artifact"
	index := 0
	artifact := a2a.NewArtifact(index, title, part)

	if artifact.Index != index {
		t.Errorf("Expected index %d, got %d", index, artifact.Index)
	}

	if artifact.Title != title {
		t.Errorf("Expected title %q, got %q", title, artifact.Title)
	}

	if len(artifact.Parts) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(artifact.Parts))
	}

	if artifact.Parts[0].Type != a2a.PartTypeText {
		t.Errorf("Expected part type %s, got %s", a2a.PartTypeText, artifact.Parts[0].Type)
	}

	if artifact.Parts[0].Text == nil {
		t.Fatalf("Expected non-nil text, got nil")
	}

	if *artifact.Parts[0].Text != "Hello, world!" {
		t.Errorf("Expected text %q, got %q", "Hello, world!", *artifact.Parts[0].Text)
	}
}

// TestInMemoryTaskStore tests the in-memory task store
func TestInMemoryTaskStore(t *testing.T) {
	store := server.NewInMemoryTaskStore()
	ctx := context.Background()

	// Create a task
	task := &a2a.Task{
		ID:        "test-task",
		SessionID: "test-session",
		Status: a2a.TaskStatus{
			State:     a2a.TaskSubmitted,
			Timestamp: time.Now(),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test CreateTask
	err := store.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Test GetTask
	retrievedTask, err := store.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if diff := cmp.Diff(task, retrievedTask, cmpopts.IgnoreFields(a2a.Task{}, "CreatedAt", "UpdatedAt", "Status")); diff != "" {
		t.Errorf("Task mismatch (-want +got):\n%s", diff)
	}

	// Test UpdateTask
	task.Status.State = a2a.TaskCompleted
	err = store.UpdateTask(ctx, task)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	retrievedTask, err = store.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}

	if retrievedTask.Status.State != a2a.TaskCompleted {
		t.Errorf("Expected task state %s, got %s", a2a.TaskCompleted, retrievedTask.Status.State)
	}

	// Test ListTasks
	tasks, err := store.ListTasks(ctx, "")
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	// Test ListTasks with session ID
	tasks, err = store.ListTasks(ctx, task.SessionID)
	if err != nil {
		t.Fatalf("Failed to list tasks by session: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task for session, got %d", len(tasks))
	}

	// Test DeleteTask
	err = store.DeleteTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	_, err = store.GetTask(ctx, task.ID)
	if err == nil {
		t.Error("Expected error when getting deleted task, got nil")
	}
}

// TestFileTaskStore tests the file-based task store
func TestFileTaskStore(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "a2a-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a file store
	store, err := server.NewFileTaskStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}

	ctx := context.Background()

	// Create a task
	task := &a2a.Task{
		ID:        "test-task",
		SessionID: "test-session",
		Status: a2a.TaskStatus{
			State:     a2a.TaskSubmitted,
			Timestamp: time.Now(),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test CreateTask
	err = store.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Verify that the task file exists
	taskFilePath := filepath.Join(tempDir, task.ID+".task.json")
	if _, err := os.Stat(taskFilePath); os.IsNotExist(err) {
		t.Errorf("Task file %s does not exist", taskFilePath)
	}

	// Test GetTask
	retrievedTask, err := store.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if diff := cmp.Diff(task, retrievedTask, cmpopts.IgnoreFields(a2a.Task{}, "CreatedAt", "UpdatedAt", "Status")); diff != "" {
		t.Errorf("Task mismatch (-want +got):\n%s", diff)
	}

	// Test UpdateTask
	task.Status.State = a2a.TaskCompleted
	err = store.UpdateTask(ctx, task)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	retrievedTask, err = store.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}

	if retrievedTask.Status.State != a2a.TaskCompleted {
		t.Errorf("Expected task state %s, got %s", a2a.TaskCompleted, retrievedTask.Status.State)
	}

	// Test ListTasks
	tasks, err := store.ListTasks(ctx, "")
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	// Test ListTasks with session ID
	tasks, err = store.ListTasks(ctx, task.SessionID)
	if err != nil {
		t.Fatalf("Failed to list tasks by session: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task for session, got %d", len(tasks))
	}

	// Test DeleteTask
	err = store.DeleteTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	_, err = store.GetTask(ctx, task.ID)
	if err == nil {
		t.Error("Expected error when getting deleted task, got nil")
	}

	// Verify that the task file has been deleted
	if _, err := os.Stat(taskFilePath); !os.IsNotExist(err) {
		t.Errorf("Task file %s still exists after deletion", taskFilePath)
	}
}

// TestA2AClient tests the A2A client
func TestA2AClient(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the request
		var request a2a.JsonRpcRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Handle the request based on the method
		var result any
		switch request.Method {
		case "agent/card":
			result = a2a.AgentCard{
				AgentType:   "TestAgent",
				Name:        "Test Agent",
				Description: "A test agent",
				Version:     "1.0.0",
				Capabilities: a2a.AgentCapabilities{
					Streaming: true,
					MultiTurn: true,
				},
			}
		case "tasks/get":
			// Parse the parameters
			var params a2a.TasksGetRequest
			paramsJSON, _ := json.Marshal(request.Params)
			json.Unmarshal(paramsJSON, &params)

			// Return a task
			result = a2a.Task{
				ID:        params.ID,
				SessionID: "test-session",
				Status: a2a.TaskStatus{
					State:     a2a.TaskCompleted,
					Timestamp: time.Now(),
				},
				Artifacts: []a2a.Artifact{
					{
						Parts: []a2a.Part{
							{
								Type: a2a.PartTypeText,
								Text: func() *string { s := "Hello, world!"; return &s }(),
							},
						},
						Index: 0,
					},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		case "tasks/send":
			// Parse the parameters
			var params a2a.TasksSendRequest
			paramsJSON, _ := json.Marshal(request.Params)
			json.Unmarshal(paramsJSON, &params)

			// Create a response message
			responseText := "Hello, agent!"
			if len(params.Message.Parts) > 0 && params.Message.Parts[0].Type == a2a.PartTypeText && params.Message.Parts[0].Text != nil {
				responseText = "Echo: " + *params.Message.Parts[0].Text
			}

			// Return a task
			result = a2a.Task{
				ID:        params.ID,
				SessionID: params.SessionID,
				Status: a2a.TaskStatus{
					State:     a2a.TaskCompleted,
					Timestamp: time.Now(),
				},
				Artifacts: []a2a.Artifact{
					{
						Parts: []a2a.Part{
							{
								Type: a2a.PartTypeText,
								Text: &responseText,
							},
						},
						Index: 0,
					},
				},
				History: []a2a.Message{
					params.Message,
					{
						Role: a2a.RoleAgent,
						Parts: []a2a.Part{
							{
								Type: a2a.PartTypeText,
								Text: &responseText,
							},
						},
					},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		default:
			http.Error(w, "Method not found", http.StatusNotFound)
			return
		}

		// Create the response
		response := a2a.JsonRpcResponse{
			JsonRpc: "2.0",
			ID:      request.ID,
			Result:  result,
		}

		// Write the response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	// Create a client
	a2aClient, err := client.NewClient(ts.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test FetchAgentCard
	card, err := a2aClient.FetchAgentCard(context.Background())
	if err != nil {
		t.Fatalf("Failed to fetch agent card: %v", err)
	}

	if card.Name != "Test Agent" {
		t.Errorf("Expected agent name %q, got %q", "Test Agent", card.Name)
	}

	// Test GetTask
	task, err := a2aClient.GetTask(context.Background(), "test-task")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if task.ID != "test-task" {
		t.Errorf("Expected task ID %q, got %q", "test-task", task.ID)
	}

	// Test SendTask
	message := a2a.NewUserMessage(a2a.NewTextPart("Hello, A2A!"))
	taskOpts := client.TaskOptions{
		ID:                  "test-task",
		SessionID:           "test-session",
		AcceptedOutputModes: []string{"text"},
		Message:             message,
	}

	task, err = a2aClient.SendTask(context.Background(), taskOpts)
	if err != nil {
		t.Fatalf("Failed to send task: %v", err)
	}

	if task.ID != "test-task" {
		t.Errorf("Expected task ID %q, got %q", "test-task", task.ID)
	}

	if len(task.Artifacts) != 1 {
		t.Fatalf("Expected 1 artifact, got %d", len(task.Artifacts))
	}

	if len(task.Artifacts[0].Parts) != 1 {
		t.Fatalf("Expected 1 part in artifact, got %d", len(task.Artifacts[0].Parts))
	}

	if task.Artifacts[0].Parts[0].Type != a2a.PartTypeText {
		t.Errorf("Expected part type %s, got %s", a2a.PartTypeText, task.Artifacts[0].Parts[0].Type)
	}

	if task.Artifacts[0].Parts[0].Text == nil {
		t.Fatalf("Expected non-nil text in artifact part, got nil")
	}

	expectedText := "Echo: Hello, A2A!"
	if *task.Artifacts[0].Parts[0].Text != expectedText {
		t.Errorf("Expected text %q in artifact part, got %q", expectedText, *task.Artifacts[0].Parts[0].Text)
	}
}

// TestTaskWorkflow tests a complete task workflow
func TestTaskWorkflow(t *testing.T) {
	// Create a simple agent that echoes messages
	handler := &echoAgent{}

	// Create agent capabilities
	capabilities := a2a.AgentCapabilities{
		Streaming:         true,
		MultiTurn:         true,
		MultiTask:         true,
		DefaultOutputMode: "text",
	}

	// Create server options
	opts := server.ServerOptions{
		AgentCard: a2a.AgentCard{
			AgentType:    "EchoAgent",
			Name:         "Echo Agent",
			Description:  "An agent that echoes messages",
			Version:      "1.0.0",
			Capabilities: capabilities,
		},
		TaskStore:   server.NewInMemoryTaskStore(),
		TaskHandler: handler,
		Path:        "/",
	}

	// Create A2A server
	a2aServer := server.NewServer(opts)

	// Create test server
	ts := httptest.NewServer(a2aServer)
	defer ts.Close()

	// Create A2A client
	a2aClient, err := client.NewClient(ts.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Fetch agent card
	card, err := a2aClient.FetchAgentCard(context.Background())
	if err != nil {
		t.Fatalf("Failed to fetch agent card: %v", err)
	}

	if card.Name != "Echo Agent" {
		t.Errorf("Expected agent name %q, got %q", "Echo Agent", card.Name)
	}

	// Send a task
	message := a2a.NewUserMessage(a2a.NewTextPart("Hello, Echo Agent!"))
	taskOpts := client.TaskOptions{
		ID:                  "test-task",
		SessionID:           "test-session",
		AcceptedOutputModes: []string{"text"},
		Message:             message,
	}

	task, err := a2aClient.SendTask(context.Background(), taskOpts)
	if err != nil {
		t.Fatalf("Failed to send task: %v", err)
	}

	if task.ID != "test-task" {
		t.Errorf("Expected task ID %q, got %q", "test-task", task.ID)
	}

	if task.Status.State != a2a.TaskCompleted {
		t.Errorf("Expected task state %s, got %s", a2a.TaskCompleted, task.Status.State)
	}

	if len(task.Artifacts) != 1 {
		t.Fatalf("Expected 1 artifact, got %d", len(task.Artifacts))
	}

	if len(task.Artifacts[0].Parts) != 1 {
		t.Fatalf("Expected 1 part in artifact, got %d", len(task.Artifacts[0].Parts))
	}

	if task.Artifacts[0].Parts[0].Type != a2a.PartTypeText {
		t.Errorf("Expected part type %s, got %s", a2a.PartTypeText, task.Artifacts[0].Parts[0].Type)
	}

	if task.Artifacts[0].Parts[0].Text == nil {
		t.Fatalf("Expected non-nil text in artifact part, got nil")
	}

	expectedText := "Echo: Hello, Echo Agent!"
	if *task.Artifacts[0].Parts[0].Text != expectedText {
		t.Errorf("Expected text %q in artifact part, got %q", expectedText, *task.Artifacts[0].Parts[0].Text)
	}

	// Test multi-turn conversation
	handler.SetInputRequired(true)

	message = a2a.NewUserMessage(a2a.NewTextPart("Hello again!"))
	taskOpts = client.TaskOptions{
		ID:                  "test-task-2",
		SessionID:           "test-session",
		AcceptedOutputModes: []string{"text"},
		Message:             message,
	}

	task, err = a2aClient.SendTask(context.Background(), taskOpts)
	if err != nil {
		t.Fatalf("Failed to send task: %v", err)
	}

	if task.Status.State != a2a.TaskInputRequired {
		t.Errorf("Expected task state %s, got %s", a2a.TaskInputRequired, task.Status.State)
	}

	// Send a follow-up message
	handler.SetInputRequired(false)

	message = a2a.NewUserMessage(a2a.NewTextPart("Here's more information."))
	taskOpts = client.TaskOptions{
		ID:                  "test-task-2",
		SessionID:           "test-session",
		AcceptedOutputModes: []string{"text"},
		Message:             message,
	}

	task, err = a2aClient.SendTask(context.Background(), taskOpts)
	if err != nil {
		t.Fatalf("Failed to send follow-up message: %v", err)
	}

	if task.Status.State != a2a.TaskCompleted {
		t.Errorf("Expected task state %s, got %s", a2a.TaskCompleted, task.Status.State)
	}

	if len(task.History) < 3 {
		t.Fatalf("Expected at least 3 messages in history, got %d", len(task.History))
	}

	// Test data parts
	data := map[string]string{"key": "value"}
	message = a2a.NewUserMessage(a2a.NewDataPart(data))
	taskOpts = client.TaskOptions{
		ID:                  "test-task-3",
		SessionID:           "test-session",
		AcceptedOutputModes: []string{"text"},
		Message:             message,
	}

	task, err = a2aClient.SendTask(context.Background(), taskOpts)
	if err != nil {
		t.Fatalf("Failed to send task with data: %v", err)
	}

	if task.Status.State != a2a.TaskCompleted {
		t.Errorf("Expected task state %s, got %s", a2a.TaskCompleted, task.Status.State)
	}

	if len(task.Artifacts) != 1 {
		t.Fatalf("Expected 1 artifact, got %d", len(task.Artifacts))
	}

	if len(task.Artifacts[0].Parts) != 1 {
		t.Fatalf("Expected 1 part in artifact, got %d", len(task.Artifacts[0].Parts))
	}

	if task.Artifacts[0].Parts[0].Type != a2a.PartTypeText {
		t.Errorf("Expected part type %s, got %s", a2a.PartTypeText, task.Artifacts[0].Parts[0].Type)
	}

	if task.Artifacts[0].Parts[0].Text == nil {
		t.Fatalf("Expected non-nil text in artifact part, got nil")
	}

	expectedText = "Echo DataPart: {\"key\":\"value\"}"
	if *task.Artifacts[0].Parts[0].Text != expectedText {
		t.Errorf("Expected text %q in artifact part, got %q", expectedText, *task.Artifacts[0].Parts[0].Text)
	}
}

// echoAgent is a simple agent that echoes messages
type echoAgent struct {
	inputRequired bool
}

// SetInputRequired sets whether the agent should request more input
func (a *echoAgent) SetInputRequired(inputRequired bool) {
	a.inputRequired = inputRequired
}

// HandleTask processes a new task
func (a *echoAgent) HandleTask(ctx context.Context, task *a2a.Task) (*a2a.Task, error) {
	// Extract message from the task
	if task.Message == nil || len(task.Message.Parts) == 0 {
		task.Status.State = a2a.TaskFailed
		task.Status.Timestamp = time.Now()
		task.Status.Error = &a2a.TaskError{
			Code:    "invalid_message",
			Message: "No message provided",
		}
		return task, nil
	}

	// Process the message
	var response string
	for _, part := range task.Message.Parts {
		switch part.Type {
		case a2a.PartTypeText:
			if part.Text != nil {
				response = "Echo: " + *part.Text
			}
		case a2a.PartTypeData:
			if part.Data != nil {
				data, _ := json.Marshal(part.Data)
				response = "Echo DataPart: " + string(data)
			}
		case a2a.PartTypeFile:
			if part.FileURI != nil {
				response = "Echo FileURI: " + *part.FileURI
			} else if part.FileBytes != nil {
				response = "Echo FileBytes: " + string(*part.FileBytes)
			}
		}
	}

	// Add message to history
	task.History = append(task.History, *task.Message)

	// Check if we should request more input
	if a.inputRequired {
		task.Status.State = a2a.TaskInputRequired
		task.Status.Timestamp = time.Now()

		// Add response message
		responseMsg := a2a.NewAgentMessage(a2a.NewTextPart("Please provide more information."))
		task.History = append(task.History, responseMsg)

		return task, nil
	}

	// Create response artifact
	artifact := a2a.NewArtifact(0, "Response", a2a.NewTextPart(response))
	task.Artifacts = append(task.Artifacts, artifact)

	// Add response message to history
	responseMsg := a2a.NewAgentMessage(a2a.NewTextPart(response))
	task.History = append(task.History, responseMsg)

	// Update task status
	task.Status.State = a2a.TaskCompleted
	task.Status.Timestamp = time.Now()
	task.UpdatedAt = time.Now()

	return task, nil
}

// HandleTaskUpdate processes an update to an existing task
func (a *echoAgent) HandleTaskUpdate(ctx context.Context, task *a2a.Task, message a2a.Message) (*a2a.Task, error) {
	// Extract text from the message
	var text string
	for _, part := range message.Parts {
		if part.Type == a2a.PartTypeText && part.Text != nil {
			text = *part.Text
			break
		}
	}

	// Create response
	response := "Echo update: " + text

	// Create response artifact
	artifact := a2a.NewArtifact(0, "Response", a2a.NewTextPart(response))
	task.Artifacts = append(task.Artifacts, artifact)

	// Add response message to history
	responseMsg := a2a.NewAgentMessage(a2a.NewTextPart(response))
	task.History = append(task.History, responseMsg)

	// Update task status
	task.Status.State = a2a.TaskCompleted
	task.Status.Timestamp = time.Now()
	task.UpdatedAt = time.Now()

	return task, nil
}

// HandleTaskCancel processes a task cancellation
func (a *echoAgent) HandleTaskCancel(ctx context.Context, task *a2a.Task, reason string) error {
	return nil
}
