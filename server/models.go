// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package server provides database models for the A2A protocol.
// This file contains SQLAlchemy-equivalent database models converted from Python
// to idiomatic Go code with GORM ORM support and proper JSON handling.
package server

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"github.com/go-a2a/a2a"
)

// TaskStatusJSON provides JSON serialization for TaskStatus in database columns.
// This is equivalent to Python's PydanticType for TaskStatus.
type TaskStatusJSON struct {
	a2a.TaskStatus
}

// Value implements the driver.Valuer interface for database storage.
func (ts TaskStatusJSON) Value() (driver.Value, error) {
	if ts.TaskStatus == (a2a.TaskStatus{}) {
		return nil, nil
	}
	return json.Marshal(ts.TaskStatus)
}

// Scan implements the sql.Scanner interface for database retrieval.
func (ts *TaskStatusJSON) Scan(value any) error {
	if value == nil {
		*ts = TaskStatusJSON{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into TaskStatusJSON", value)
	}

	var status a2a.TaskStatus
	if err := json.Unmarshal(bytes, &status); err != nil {
		return fmt.Errorf("cannot unmarshal TaskStatusJSON: %w", err)
	}

	ts.TaskStatus = status
	return nil
}

// ArtifactSliceJSON provides JSON serialization for []Artifact in database columns.
// This is equivalent to Python's PydanticListType for Artifact.
type ArtifactSliceJSON struct {
	Artifacts []a2a.Artifact
}

// Value implements the driver.Valuer interface for database storage.
func (as ArtifactSliceJSON) Value() (driver.Value, error) {
	if as.Artifacts == nil {
		return nil, nil
	}
	return json.Marshal(as.Artifacts)
}

// Scan implements the sql.Scanner interface for database retrieval.
func (as *ArtifactSliceJSON) Scan(value any) error {
	if value == nil {
		*as = ArtifactSliceJSON{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into ArtifactSliceJSON", value)
	}

	var artifacts []a2a.Artifact
	if err := json.Unmarshal(bytes, &artifacts); err != nil {
		return fmt.Errorf("cannot unmarshal ArtifactSliceJSON: %w", err)
	}

	as.Artifacts = artifacts
	return nil
}

// MessageSliceJSON provides JSON serialization for []Message in database columns.
// This is equivalent to Python's PydanticListType for Message.
type MessageSliceJSON struct {
	Messages []a2a.Message
}

// Value implements the driver.Valuer interface for database storage.
func (ms MessageSliceJSON) Value() (driver.Value, error) {
	if ms.Messages == nil {
		return nil, nil
	}
	return json.Marshal(ms.Messages)
}

// Scan implements the sql.Scanner interface for database retrieval.
func (ms *MessageSliceJSON) Scan(value any) error {
	if value == nil {
		*ms = MessageSliceJSON{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into MessageSliceJSON", value)
	}

	var messages []a2a.Message
	if err := json.Unmarshal(bytes, &messages); err != nil {
		return fmt.Errorf("cannot unmarshal MessageSliceJSON: %w", err)
	}

	ms.Messages = messages
	return nil
}

// TaskMixin provides standard task columns with proper type handling.
// This is equivalent to Python's TaskMixin class.
type TaskMixin struct {
	ID        string            `gorm:"primaryKey;size:36" json:"id"`
	ContextID string            `gorm:"size:36;not null" json:"contextId"`
	Kind      string            `gorm:"size:16;default:task;not null" json:"kind"`
	Status    TaskStatusJSON    `gorm:"type:json" json:"status"`
	Artifacts ArtifactSliceJSON `gorm:"type:json" json:"artifacts"`
	History   MessageSliceJSON  `gorm:"type:json" json:"history"`
	Metadata  map[string]any    `gorm:"type:json" json:"metadata"`
}

// Validate ensures the TaskMixin is in a valid state.
func (tm *TaskMixin) Validate() error {
	if tm.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if tm.ContextID == "" {
		return fmt.Errorf("task context ID cannot be empty")
	}
	if tm.Kind == "" {
		return fmt.Errorf("task kind cannot be empty")
	}

	// Validate status
	if err := tm.Status.TaskStatus.Validate(); err != nil {
		return fmt.Errorf("task status is invalid: %w", err)
	}

	// Validate artifacts
	for i, artifact := range tm.Artifacts.Artifacts {
		if err := artifact.Validate(); err != nil {
			return fmt.Errorf("artifact at index %d is invalid: %w", i, err)
		}
	}

	// Validate history messages
	for i, message := range tm.History.Messages {
		if err := message.Validate(); err != nil {
			return fmt.Errorf("history message at index %d is invalid: %w", i, err)
		}
	}

	return nil
}

// String returns a string representation of the TaskMixin for debugging.
func (tm *TaskMixin) String() string {
	return fmt.Sprintf("TaskMixin{ID: %s, ContextID: %s, Status: %s}",
		tm.ID, tm.ContextID, tm.Status.TaskStatus.State)
}

// TaskModel represents the default task model with standard table name.
// This is equivalent to Python's TaskModel class.
type TaskModel struct {
	TaskMixin
}

var _ TaskModelInterface = (*TaskModel)(nil)

// TableName returns the table name for the TaskModel.
func (TaskModel) TableName() string {
	return "tasks"
}

// NewTaskModel creates a new TaskModel instance.
func NewTaskModel() *TaskModel {
	return &TaskModel{
		TaskMixin: TaskMixin{
			Kind:      "task",
			Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
			Artifacts: ArtifactSliceJSON{Artifacts: []a2a.Artifact{}},
			History:   MessageSliceJSON{Messages: []a2a.Message{}},
			Metadata:  make(map[string]any),
		},
	}
}

// TaskModelInterface defines the interface for task models.
// This allows for different implementations with different table names.
type TaskModelInterface interface {
	TableName() string
	Validate() error
	String() string
}

// DynamicTaskModel represents a task model with a configurable table name.
// This is used by the CreateTaskModel factory function.
type DynamicTaskModel struct {
	TaskMixin
	tableName string
}

var _ TaskModelInterface = (*DynamicTaskModel)(nil)

// TableName returns the configured table name for the DynamicTaskModel.
func (dtm *DynamicTaskModel) TableName() string {
	return dtm.tableName
}

// CreateTaskModel creates a TaskModel struct with a configurable table name.
// This is equivalent to Python's create_task_model function.
//
// Args:
//
//	tableName: Name of the database table. Defaults to 'tasks'.
//
// Returns:
//
//	A TaskModel instance with the specified table name.
//
// Example:
//
//	// Create a task model with default table name
//	taskModel := CreateTaskModel("tasks")
//
//	// Create a task model with custom table name
//	customTaskModel := CreateTaskModel("my_tasks")
func CreateTaskModel(tableName string) *DynamicTaskModel {
	if tableName == "" {
		tableName = "tasks"
	}

	return &DynamicTaskModel{
		TaskMixin: TaskMixin{
			Kind:      "task",
			Status:    TaskStatusJSON{a2a.TaskStatus{State: a2a.TaskStateSubmitted}},
			Artifacts: ArtifactSliceJSON{Artifacts: []a2a.Artifact{}},
			History:   MessageSliceJSON{Messages: []a2a.Message{}},
			Metadata:  make(map[string]any),
		},
		tableName: tableName,
	}
}

// NewTaskModelFromTask creates a TaskModel from an existing A2A Task.
// This is a convenience function for converting A2A tasks to database models.
func NewTaskModelFromTask(task *a2a.Task) (*TaskModel, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	if err := task.Validate(); err != nil {
		return nil, fmt.Errorf("task is invalid: %w", err)
	}

	model := &TaskModel{
		TaskMixin: TaskMixin{
			ID:        task.ID,
			ContextID: task.ContextID,
			Kind:      "task",
			Status:    TaskStatusJSON{task.Status},
			Artifacts: ArtifactSliceJSON{Artifacts: make([]a2a.Artifact, len(task.Artifacts))},
			History:   MessageSliceJSON{Messages: make([]a2a.Message, len(task.History))},
			Metadata:  make(map[string]any),
		},
	}

	// Copy artifacts
	for i, artifact := range task.Artifacts {
		if artifact != nil {
			model.Artifacts.Artifacts[i] = *artifact
		}
	}

	// Copy history messages
	for i, message := range task.History {
		if message != nil {
			model.History.Messages[i] = *message
		}
	}

	return model, nil
}

// ToTask converts a TaskModel to an A2A Task.
// This is useful for converting database models back to A2A protocol types.
func (tm *TaskModel) ToTask() (*a2a.Task, error) {
	if err := tm.Validate(); err != nil {
		return nil, fmt.Errorf("task model is invalid: %w", err)
	}

	task := &a2a.Task{
		ID:        tm.ID,
		ContextID: tm.ContextID,
		Status:    tm.Status.TaskStatus,
		History:   make([]*a2a.Message, len(tm.History.Messages)),
		Artifacts: make([]*a2a.Artifact, len(tm.Artifacts.Artifacts)),
	}

	// Copy history messages
	for i, message := range tm.History.Messages {
		task.History[i] = &message
	}

	// Copy artifacts
	for i, artifact := range tm.Artifacts.Artifacts {
		task.Artifacts[i] = &artifact
	}

	return task, nil
}

// BeforeCreate is a GORM hook called before creating a record.
func (tm *TaskModel) BeforeCreate(tx *gorm.DB) error {
	return tm.Validate()
}

// BeforeUpdate is a GORM hook called before updating a record.
func (tm *TaskModel) BeforeUpdate(tx *gorm.DB) error {
	return tm.Validate()
}

// BeforeCreate is a GORM hook called before creating a record.
func (dtm *DynamicTaskModel) BeforeCreate(tx *gorm.DB) error {
	return dtm.Validate()
}

// BeforeUpdate is a GORM hook called before updating a record.
func (dtm *DynamicTaskModel) BeforeUpdate(tx *gorm.DB) error {
	return dtm.Validate()
}
