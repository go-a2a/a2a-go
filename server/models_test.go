// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"reflect"
	"testing"

	"gorm.io/gorm"

	"github.com/go-a2a/a2a"
)

// TestTaskStatusJSON tests JSON serialization and deserialization for TaskStatus
func TestTaskStatusJSON(t *testing.T) {
	t.Run("Value", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    TaskStatusJSON
			expected string
		}{
			{
				name: "valid submitted status",
				input: TaskStatusJSON{
					TaskStatus: a2a.TaskStatus{State: a2a.TaskStateSubmitted},
				},
				expected: `{"state":"submitted"}`,
			},
			{
				name: "valid running status",
				input: TaskStatusJSON{
					TaskStatus: a2a.TaskStatus{State: a2a.TaskStateRunning},
				},
				expected: `{"state":"running"}`,
			},
			{
				name: "valid completed status",
				input: TaskStatusJSON{
					TaskStatus: a2a.TaskStatus{State: a2a.TaskStateCompleted},
				},
				expected: `{"state":"completed"}`,
			},
			{
				name: "valid failed status",
				input: TaskStatusJSON{
					TaskStatus: a2a.TaskStatus{State: a2a.TaskStateFailed},
				},
				expected: `{"state":"failed"}`,
			},
			{
				name: "valid canceled status",
				input: TaskStatusJSON{
					TaskStatus: a2a.TaskStatus{State: a2a.TaskStateCanceled},
				},
				expected: `{"state":"canceled"}`,
			},
			{
				name:     "zero value",
				input:    TaskStatusJSON{},
				expected: "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				value, err := tc.input.Value()
				if err != nil {
					t.Fatalf("Value() returned error: %v", err)
				}

				if tc.expected == "" {
					if value != nil {
						t.Errorf("Expected nil value for zero struct, got %v", value)
					}
				} else {
					if value == nil {
						t.Error("Expected non-nil value, got nil")
					} else if string(value.([]byte)) != tc.expected {
						t.Errorf("Expected %s, got %s", tc.expected, string(value.([]byte)))
					}
				}
			})
		}
	})

	t.Run("Scan", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    any
			expected TaskStatusJSON
			hasError bool
		}{
			{
				name:     "nil input",
				input:    nil,
				expected: TaskStatusJSON{},
			},
			{
				name:  "valid byte slice",
				input: []byte(`{"state":"submitted"}`),
				expected: TaskStatusJSON{
					TaskStatus: a2a.TaskStatus{State: a2a.TaskStateSubmitted},
				},
			},
			{
				name:  "valid string",
				input: `{"state":"running"}`,
				expected: TaskStatusJSON{
					TaskStatus: a2a.TaskStatus{State: a2a.TaskStateRunning},
				},
			},
			{
				name:     "invalid JSON",
				input:    []byte(`{"state":"invalid"`),
				hasError: true,
			},
			{
				name:     "invalid type",
				input:    123,
				hasError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var result TaskStatusJSON
				err := result.Scan(tc.input)

				if tc.hasError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
				} else {
					if err != nil {
						t.Fatalf("Scan() returned error: %v", err)
					}
					if !reflect.DeepEqual(result, tc.expected) {
						t.Errorf("Expected %+v, got %+v", tc.expected, result)
					}
				}
			})
		}
	})
}

// TestArtifactSliceJSON tests JSON serialization and deserialization for []Artifact
func TestArtifactSliceJSON(t *testing.T) {
	// Create a test artifact with valid parts
	testArtifact, err := a2a.NewTextArtifact("test-artifact", "test content", "test description")
	if err != nil {
		t.Fatalf("Failed to create test artifact: %v", err)
	}

	t.Run("Value", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    ArtifactSliceJSON
			expected string
		}{
			{
				name: "valid artifacts",
				input: ArtifactSliceJSON{
					Artifacts: []a2a.Artifact{*testArtifact},
				},
				expected: `[{"artifactId":"` + testArtifact.ArtifactID + `","name":"test-artifact","description":"test description","parts":[{"kind":"text","text":"test content"}]}]`,
			},
			{
				name: "empty slice",
				input: ArtifactSliceJSON{
					Artifacts: []a2a.Artifact{},
				},
				expected: `[]`,
			},
			{
				name:     "nil slice",
				input:    ArtifactSliceJSON{},
				expected: "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				value, err := tc.input.Value()
				if err != nil {
					t.Fatalf("Value() returned error: %v", err)
				}

				if tc.expected == "" {
					if value != nil {
						t.Errorf("Expected nil value for nil slice, got %v", value)
					}
				} else {
					if value == nil {
						t.Error("Expected non-nil value, got nil")
					} else if string(value.([]byte)) != tc.expected {
						t.Errorf("Expected %s, got %s", tc.expected, string(value.([]byte)))
					}
				}
			})
		}
	})

	t.Run("Scan", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    any
			expected ArtifactSliceJSON
			hasError bool
		}{
			{
				name:     "nil input",
				input:    nil,
				expected: ArtifactSliceJSON{},
			},
			{
				name:  "valid byte slice",
				input: []byte(`[{"artifactId":"` + testArtifact.ArtifactID + `","name":"test-artifact","description":"test description","parts":[{"kind":"text","text":"test content"}]}]`),
				expected: ArtifactSliceJSON{
					Artifacts: []a2a.Artifact{*testArtifact},
				},
			},
			{
				name:  "valid string",
				input: `[]`,
				expected: ArtifactSliceJSON{
					Artifacts: []a2a.Artifact{},
				},
			},
			{
				name:     "invalid JSON",
				input:    []byte(`[{"artifactId":"test-id"`),
				hasError: true,
			},
			{
				name:     "invalid type",
				input:    123,
				hasError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var result ArtifactSliceJSON
				err := result.Scan(tc.input)

				if tc.hasError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
				} else {
					if err != nil {
						t.Fatalf("Scan() returned error: %v", err)
					}
					if !reflect.DeepEqual(result, tc.expected) {
						t.Errorf("Expected %+v, got %+v", tc.expected, result)
					}
				}
			})
		}
	})
}

// TestMessageSliceJSON tests JSON serialization and deserialization for []Message
func TestMessageSliceJSON(t *testing.T) {
	testMessage := a2a.Message{
		ContextID: "test-context",
		Content:   "test content",
		Metadata:  map[string]any{"key": "value"},
	}

	t.Run("Value", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    MessageSliceJSON
			expected string
		}{
			{
				name: "valid messages",
				input: MessageSliceJSON{
					Messages: []a2a.Message{testMessage},
				},
				expected: `[{"contextId":"test-context","content":"test content","metadata":{"key":"value"}}]`,
			},
			{
				name: "empty slice",
				input: MessageSliceJSON{
					Messages: []a2a.Message{},
				},
				expected: `[]`,
			},
			{
				name:     "nil slice",
				input:    MessageSliceJSON{},
				expected: "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				value, err := tc.input.Value()
				if err != nil {
					t.Fatalf("Value() returned error: %v", err)
				}

				if tc.expected == "" {
					if value != nil {
						t.Errorf("Expected nil value for nil slice, got %v", value)
					}
				} else {
					if value == nil {
						t.Error("Expected non-nil value, got nil")
					} else if string(value.([]byte)) != tc.expected {
						t.Errorf("Expected %s, got %s", tc.expected, string(value.([]byte)))
					}
				}
			})
		}
	})

	t.Run("Scan", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    any
			expected MessageSliceJSON
			hasError bool
		}{
			{
				name:     "nil input",
				input:    nil,
				expected: MessageSliceJSON{},
			},
			{
				name:  "valid byte slice",
				input: []byte(`[{"contextId":"test-context","content":"test content","metadata":{"key":"value"}}]`),
				expected: MessageSliceJSON{
					Messages: []a2a.Message{testMessage},
				},
			},
			{
				name:  "valid string",
				input: `[]`,
				expected: MessageSliceJSON{
					Messages: []a2a.Message{},
				},
			},
			{
				name:     "invalid JSON",
				input:    []byte(`[{"contextId":"test-context"`),
				hasError: true,
			},
			{
				name:     "invalid type",
				input:    123,
				hasError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var result MessageSliceJSON
				err := result.Scan(tc.input)

				if tc.hasError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
				} else {
					if err != nil {
						t.Fatalf("Scan() returned error: %v", err)
					}
					if !reflect.DeepEqual(result, tc.expected) {
						t.Errorf("Expected %+v, got %+v", tc.expected, result)
					}
				}
			})
		}
	})
}

// TestTaskMixin tests the TaskMixin validation and string methods
func TestTaskMixin(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    TaskMixin
			hasError bool
		}{
			{
				name: "valid task mixin",
				input: TaskMixin{
					ID:        "test-id",
					ContextID: "test-context",
					Kind:      "task",
					Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
					Artifacts: ArtifactSliceJSON{Artifacts: []a2a.Artifact{}},
					History:   MessageSliceJSON{Messages: []a2a.Message{}},
					Metadata:  map[string]any{},
				},
				hasError: false,
			},
			{
				name: "empty ID",
				input: TaskMixin{
					ID:        "",
					ContextID: "test-context",
					Kind:      "task",
					Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
				},
				hasError: true,
			},
			{
				name: "empty context ID",
				input: TaskMixin{
					ID:        "test-id",
					ContextID: "",
					Kind:      "task",
					Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
				},
				hasError: true,
			},
			{
				name: "empty kind",
				input: TaskMixin{
					ID:        "test-id",
					ContextID: "test-context",
					Kind:      "",
					Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
				},
				hasError: true,
			},
			{
				name: "invalid status",
				input: TaskMixin{
					ID:        "test-id",
					ContextID: "test-context",
					Kind:      "task",
					Status:    TaskStatusJSON{a2a.TaskStatus{State: "invalid"}},
				},
				hasError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := tc.input.Validate()
				if tc.hasError {
					if err == nil {
						t.Error("Expected error, got nil")
					}
				} else {
					if err != nil {
						t.Errorf("Validate() returned error: %v", err)
					}
				}
			})
		}
	})

	t.Run("String", func(t *testing.T) {
		tm := TaskMixin{
			ID:        "test-id",
			ContextID: "test-context",
			Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
		}

		result := tm.String()
		expected := "TaskMixin{ID: test-id, ContextID: test-context, Status: submitted}"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}

// TestTaskModel tests the TaskModel implementation
func TestTaskModel(t *testing.T) {
	t.Run("TableName", func(t *testing.T) {
		tm := TaskModel{}
		if tm.TableName() != "tasks" {
			t.Errorf("Expected table name 'tasks', got '%s'", tm.TableName())
		}
	})

	t.Run("NewTaskModel", func(t *testing.T) {
		tm := NewTaskModel()
		if tm == nil {
			t.Fatal("NewTaskModel() returned nil")
		}
		if tm.Kind != "task" {
			t.Errorf("Expected kind 'task', got '%s'", tm.Kind)
		}
		if tm.Status.TaskStatus.State != a2a.TaskStateSubmitted {
			t.Errorf("Expected status 'submitted', got '%s'", tm.Status.TaskStatus.State)
		}
		if tm.Artifacts.Artifacts == nil {
			t.Error("Expected initialized artifacts slice")
		}
		if tm.History.Messages == nil {
			t.Error("Expected initialized history slice")
		}
		if tm.Metadata == nil {
			t.Error("Expected initialized metadata map")
		}
	})
}

// TestCreateTaskModel tests the CreateTaskModel factory function
func TestCreateTaskModel(t *testing.T) {
	testCases := []struct {
		name          string
		tableName     string
		expectedTable string
	}{
		{
			name:          "custom table name",
			tableName:     "custom_tasks",
			expectedTable: "custom_tasks",
		},
		{
			name:          "empty table name defaults to tasks",
			tableName:     "",
			expectedTable: "tasks",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dtm := CreateTaskModel(tc.tableName)
			if dtm == nil {
				t.Fatal("CreateTaskModel() returned nil")
			}
			if dtm.TableName() != tc.expectedTable {
				t.Errorf("Expected table name '%s', got '%s'", tc.expectedTable, dtm.TableName())
			}
			if dtm.Kind != "task" {
				t.Errorf("Expected kind 'task', got '%s'", dtm.Kind)
			}
			if dtm.Status.TaskStatus.State != a2a.TaskStateSubmitted {
				t.Errorf("Expected status 'submitted', got '%s'", dtm.Status.TaskStatus.State)
			}
		})
	}
}

// TestNewTaskModelFromTask tests conversion from a2a.Task to TaskModel
func TestNewTaskModelFromTask(t *testing.T) {
	t.Run("valid task", func(t *testing.T) {
		task := &a2a.Task{
			ID:        "test-id",
			ContextID: "test-context",
			Status:    a2a.TaskStatus{State: a2a.TaskStateRunning},
			History:   []*a2a.Message{},
			Artifacts: []*a2a.Artifact{},
		}

		tm, err := NewTaskModelFromTask(task)
		if err != nil {
			t.Fatalf("NewTaskModelFromTask() returned error: %v", err)
		}
		if tm == nil {
			t.Fatal("NewTaskModelFromTask() returned nil")
		}
		if tm.ID != task.ID {
			t.Errorf("Expected ID '%s', got '%s'", task.ID, tm.ID)
		}
		if tm.ContextID != task.ContextID {
			t.Errorf("Expected ContextID '%s', got '%s'", task.ContextID, tm.ContextID)
		}
		if tm.Status.TaskStatus.State != task.Status.State {
			t.Errorf("Expected status '%s', got '%s'", task.Status.State, tm.Status.TaskStatus.State)
		}
	})

	t.Run("nil task", func(t *testing.T) {
		_, err := NewTaskModelFromTask(nil)
		if err == nil {
			t.Error("Expected error for nil task, got nil")
		}
	})

	t.Run("invalid task", func(t *testing.T) {
		task := &a2a.Task{
			ID:        "",
			ContextID: "test-context",
			Status:    a2a.TaskStatus{State: a2a.TaskStateRunning},
		}

		_, err := NewTaskModelFromTask(task)
		if err == nil {
			t.Error("Expected error for invalid task, got nil")
		}
	})
}

// TestTaskModelToTask tests conversion from TaskModel to a2a.Task
func TestTaskModelToTask(t *testing.T) {
	t.Run("valid task model", func(t *testing.T) {
		tm := &TaskModel{
			TaskMixin: TaskMixin{
				ID:        "test-id",
				ContextID: "test-context",
				Kind:      "task",
				Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateRunning}},
				Artifacts: ArtifactSliceJSON{Artifacts: []a2a.Artifact{}},
				History:   MessageSliceJSON{Messages: []a2a.Message{}},
				Metadata:  map[string]any{},
			},
		}

		task, err := tm.ToTask()
		if err != nil {
			t.Fatalf("ToTask() returned error: %v", err)
		}
		if task == nil {
			t.Fatal("ToTask() returned nil")
		}
		if task.ID != tm.ID {
			t.Errorf("Expected ID '%s', got '%s'", tm.ID, task.ID)
		}
		if task.ContextID != tm.ContextID {
			t.Errorf("Expected ContextID '%s', got '%s'", tm.ContextID, task.ContextID)
		}
		if task.Status.State != tm.Status.TaskStatus.State {
			t.Errorf("Expected status '%s', got '%s'", tm.Status.TaskStatus.State, task.Status.State)
		}
	})

	t.Run("invalid task model", func(t *testing.T) {
		tm := &TaskModel{
			TaskMixin: TaskMixin{
				ID:        "",
				ContextID: "test-context",
				Kind:      "task",
				Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateRunning}},
			},
		}

		_, err := tm.ToTask()
		if err == nil {
			t.Error("Expected error for invalid task model, got nil")
		}
	})
}

// TestDynamicTaskModel tests the DynamicTaskModel implementation
func TestDynamicTaskModel(t *testing.T) {
	t.Run("TableName", func(t *testing.T) {
		dtm := &DynamicTaskModel{tableName: "custom_tasks"}
		if dtm.TableName() != "custom_tasks" {
			t.Errorf("Expected table name 'custom_tasks', got '%s'", dtm.TableName())
		}
	})
}

// TestGORMHooks tests the GORM hooks for validation
func TestGORMHooks(t *testing.T) {
	t.Run("TaskModel BeforeCreate", func(t *testing.T) {
		validModel := &TaskModel{
			TaskMixin: TaskMixin{
				ID:        "test-id",
				ContextID: "test-context",
				Kind:      "task",
				Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
				Artifacts: ArtifactSliceJSON{Artifacts: []a2a.Artifact{}},
				History:   MessageSliceJSON{Messages: []a2a.Message{}},
				Metadata:  map[string]any{},
			},
		}

		// Mock GORM transaction
		var tx *gorm.DB
		err := validModel.BeforeCreate(tx)
		if err != nil {
			t.Errorf("BeforeCreate() returned error: %v", err)
		}

		invalidModel := &TaskModel{
			TaskMixin: TaskMixin{
				ID:        "",
				ContextID: "test-context",
				Kind:      "task",
				Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
			},
		}

		err = invalidModel.BeforeCreate(tx)
		if err == nil {
			t.Error("Expected error for invalid model, got nil")
		}
	})

	t.Run("TaskModel BeforeUpdate", func(t *testing.T) {
		validModel := &TaskModel{
			TaskMixin: TaskMixin{
				ID:        "test-id",
				ContextID: "test-context",
				Kind:      "task",
				Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
				Artifacts: ArtifactSliceJSON{Artifacts: []a2a.Artifact{}},
				History:   MessageSliceJSON{Messages: []a2a.Message{}},
				Metadata:  map[string]any{},
			},
		}

		// Mock GORM transaction
		var tx *gorm.DB
		err := validModel.BeforeUpdate(tx)
		if err != nil {
			t.Errorf("BeforeUpdate() returned error: %v", err)
		}

		invalidModel := &TaskModel{
			TaskMixin: TaskMixin{
				ID:        "",
				ContextID: "test-context",
				Kind:      "task",
				Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
			},
		}

		err = invalidModel.BeforeUpdate(tx)
		if err == nil {
			t.Error("Expected error for invalid model, got nil")
		}
	})

	t.Run("DynamicTaskModel BeforeCreate", func(t *testing.T) {
		validModel := &DynamicTaskModel{
			TaskMixin: TaskMixin{
				ID:        "test-id",
				ContextID: "test-context",
				Kind:      "task",
				Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
				Artifacts: ArtifactSliceJSON{Artifacts: []a2a.Artifact{}},
				History:   MessageSliceJSON{Messages: []a2a.Message{}},
				Metadata:  map[string]any{},
			},
			tableName: "custom_tasks",
		}

		// Mock GORM transaction
		var tx *gorm.DB
		err := validModel.BeforeCreate(tx)
		if err != nil {
			t.Errorf("BeforeCreate() returned error: %v", err)
		}

		invalidModel := &DynamicTaskModel{
			TaskMixin: TaskMixin{
				ID:        "",
				ContextID: "test-context",
				Kind:      "task",
				Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
			},
			tableName: "custom_tasks",
		}

		err = invalidModel.BeforeCreate(tx)
		if err == nil {
			t.Error("Expected error for invalid model, got nil")
		}
	})

	t.Run("DynamicTaskModel BeforeUpdate", func(t *testing.T) {
		validModel := &DynamicTaskModel{
			TaskMixin: TaskMixin{
				ID:        "test-id",
				ContextID: "test-context",
				Kind:      "task",
				Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
				Artifacts: ArtifactSliceJSON{Artifacts: []a2a.Artifact{}},
				History:   MessageSliceJSON{Messages: []a2a.Message{}},
				Metadata:  map[string]any{},
			},
			tableName: "custom_tasks",
		}

		// Mock GORM transaction
		var tx *gorm.DB
		err := validModel.BeforeUpdate(tx)
		if err != nil {
			t.Errorf("BeforeUpdate() returned error: %v", err)
		}

		invalidModel := &DynamicTaskModel{
			TaskMixin: TaskMixin{
				ID:        "",
				ContextID: "test-context",
				Kind:      "task",
				Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
			},
			tableName: "custom_tasks",
		}

		err = invalidModel.BeforeUpdate(tx)
		if err == nil {
			t.Error("Expected error for invalid model, got nil")
		}
	})
}

// TestEdgeCases tests edge cases and error scenarios
func TestEdgeCases(t *testing.T) {
	t.Run("TaskMixin with invalid artifacts", func(t *testing.T) {
		invalidArtifact := a2a.Artifact{
			ArtifactID:  "",
			Name:        "test",
			Description: "test",
			Parts:       []*a2a.PartWrapper{},
		}

		tm := TaskMixin{
			ID:        "test-id",
			ContextID: "test-context",
			Kind:      "task",
			Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
			Artifacts: ArtifactSliceJSON{Artifacts: []a2a.Artifact{invalidArtifact}},
			History:   MessageSliceJSON{Messages: []a2a.Message{}},
			Metadata:  map[string]any{},
		}

		err := tm.Validate()
		if err == nil {
			t.Error("Expected error for invalid artifacts, got nil")
		}
	})

	t.Run("TaskMixin with invalid history messages", func(t *testing.T) {
		invalidMessage := a2a.Message{
			ContextID: "test-context",
			Content:   "",
			Metadata:  map[string]any{},
		}

		tm := TaskMixin{
			ID:        "test-id",
			ContextID: "test-context",
			Kind:      "task",
			Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
			Artifacts: ArtifactSliceJSON{Artifacts: []a2a.Artifact{}},
			History:   MessageSliceJSON{Messages: []a2a.Message{invalidMessage}},
			Metadata:  map[string]any{},
		}

		err := tm.Validate()
		if err == nil {
			t.Error("Expected error for invalid history messages, got nil")
		}
	})

	t.Run("NewTaskModelFromTask with nil artifacts and messages", func(t *testing.T) {
		// Create task with valid empty slices (nil pointers in slices will cause validation errors)
		task := &a2a.Task{
			ID:        "test-id",
			ContextID: "test-context",
			Status:    a2a.TaskStatus{State: a2a.TaskStateRunning},
			History:   []*a2a.Message{},
			Artifacts: []*a2a.Artifact{},
		}

		tm, err := NewTaskModelFromTask(task)
		if err != nil {
			t.Fatalf("NewTaskModelFromTask() returned error: %v", err)
		}

		if len(tm.History.Messages) != 0 {
			t.Error("Expected empty history for empty input")
		}
		if len(tm.Artifacts.Artifacts) != 0 {
			t.Error("Expected empty artifacts for empty input")
		}
	})
}
