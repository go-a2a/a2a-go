// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package task

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/server"
)

// DatabaseTaskStore is a database implementation of TaskStore using GORM.
// It uses the existing TaskModel from the server package for persistence.
type DatabaseTaskStore struct {
	db          *gorm.DB
	tableName   string
	createTable bool
}

var _ TaskStore = (*DatabaseTaskStore)(nil)

// DatabaseTaskStoreConfig holds configuration for DatabaseTaskStore.
type DatabaseTaskStoreConfig struct {
	DB          *gorm.DB
	TableName   string // Optional, defaults to "tasks"
	CreateTable bool   // Whether to create the table if it doesn't exist
}

// NewDatabaseTaskStore creates a new DatabaseTaskStore.
func NewDatabaseTaskStore(config DatabaseTaskStoreConfig) (*DatabaseTaskStore, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection cannot be nil")
	}

	tableName := config.TableName
	if tableName == "" {
		tableName = "tasks"
	}

	return &DatabaseTaskStore{
		db:          config.DB,
		tableName:   tableName,
		createTable: config.CreateTable,
	}, nil
}

// Save persists a task to the database.
func (s *DatabaseTaskStore) Save(ctx context.Context, task *a2a.Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	if err := task.Validate(); err != nil {
		return NewTaskValidationError(task.ID, err)
	}

	// Convert task to database model
	model, err := server.NewTaskModelFromTask(task)
	if err != nil {
		return NewTaskStoreError("save", task.ID, fmt.Errorf("failed to convert task to model: %w", err))
	}

	// Use dynamic table name if different from default
	db := s.db.WithContext(ctx)
	if s.tableName != "tasks" {
		db = db.Table(s.tableName)
	}

	// Use GORM's Save method which handles both create and update
	if err := db.Save(model).Error; err != nil {
		return NewTaskStoreError("save", task.ID, err)
	}

	return nil
}

// Get retrieves a task by its ID from the database.
func (s *DatabaseTaskStore) Get(ctx context.Context, taskID string) (*a2a.Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	var model server.TaskModel
	db := s.db.WithContext(ctx)
	if s.tableName != "tasks" {
		db = db.Table(s.tableName)
	}

	if err := db.Where("id = ?", taskID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, a2a.TaskNotFoundError{TaskID: taskID}
		}
		return nil, NewTaskStoreError("get", taskID, err)
	}

	// Convert database model back to task
	task, err := model.ToTask()
	if err != nil {
		return nil, NewTaskStoreError("get", taskID, fmt.Errorf("failed to convert model to task: %w", err))
	}

	return task, nil
}

// Delete removes a task from the database.
func (s *DatabaseTaskStore) Delete(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	db := s.db.WithContext(ctx)
	if s.tableName != "tasks" {
		db = db.Table(s.tableName)
	}

	result := db.Where("id = ?", taskID).Delete(&server.TaskModel{})
	if result.Error != nil {
		return NewTaskStoreError("delete", taskID, result.Error)
	}

	if result.RowsAffected == 0 {
		return a2a.TaskNotFoundError{TaskID: taskID}
	}

	return nil
}

// List retrieves tasks with optional filtering.
func (s *DatabaseTaskStore) List(ctx context.Context, contextID string, limit, offset int) ([]*a2a.Task, error) {
	var models []server.TaskModel
	db := s.db.WithContext(ctx)
	if s.tableName != "tasks" {
		db = db.Table(s.tableName)
	}

	// Apply context filter if provided
	if contextID != "" {
		db = db.Where("context_id = ?", contextID)
	}

	// Apply limit and offset
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}

	// Order by creation time (assuming ID is ordered)
	db = db.Order("id")

	if err := db.Find(&models).Error; err != nil {
		return nil, NewTaskStoreError("list", "", err)
	}

	// Convert models to tasks
	tasks := make([]*a2a.Task, len(models))
	for i, model := range models {
		task, err := model.ToTask()
		if err != nil {
			return nil, NewTaskStoreError("list", model.ID, fmt.Errorf("failed to convert model to task: %w", err))
		}
		tasks[i] = task
	}

	return tasks, nil
}

// Count returns the total number of tasks in the database.
func (s *DatabaseTaskStore) Count(ctx context.Context, contextID string) (int64, error) {
	var count int64
	db := s.db.WithContext(ctx)
	if s.tableName != "tasks" {
		db = db.Table(s.tableName)
	}

	query := db.Model(&server.TaskModel{})

	// Apply context filter if provided
	if contextID != "" {
		query = query.Where("context_id = ?", contextID)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, NewTaskStoreError("count", "", err)
	}

	return count, nil
}

// Initialize prepares the database for use.
func (s *DatabaseTaskStore) Initialize(ctx context.Context) error {
	if !s.createTable {
		return nil
	}

	// Create the table if it doesn't exist
	if s.tableName == "tasks" {
		// Use default TaskModel
		if err := s.db.WithContext(ctx).AutoMigrate(&server.TaskModel{}); err != nil {
			return NewTaskStoreError("initialize", "", err)
		}
	} else {
		// Use dynamic table name
		dynamicModel := server.CreateTaskModel(s.tableName)
		if err := s.db.WithContext(ctx).AutoMigrate(dynamicModel); err != nil {
			return NewTaskStoreError("initialize", "", err)
		}
	}

	return nil
}

// Close cleanly shuts down the database store.
func (s *DatabaseTaskStore) Close(ctx context.Context) error {
	// GORM doesn't require explicit closing in most cases
	// The underlying database connection is managed by GORM
	return nil
}

// GetByContextID retrieves all tasks for a specific context.
func (s *DatabaseTaskStore) GetByContextID(ctx context.Context, contextID string) ([]*a2a.Task, error) {
	if contextID == "" {
		return nil, fmt.Errorf("context ID cannot be empty")
	}

	var models []server.TaskModel
	db := s.db.WithContext(ctx)
	if s.tableName != "tasks" {
		db = db.Table(s.tableName)
	}

	if err := db.Where("context_id = ?", contextID).Find(&models).Error; err != nil {
		return nil, NewTaskStoreError("get_by_context", contextID, err)
	}

	// Convert models to tasks
	tasks := make([]*a2a.Task, len(models))
	for i, model := range models {
		task, err := model.ToTask()
		if err != nil {
			return nil, NewTaskStoreError("get_by_context", model.ID, fmt.Errorf("failed to convert model to task: %w", err))
		}
		tasks[i] = task
	}

	return tasks, nil
}

// GetByStatus retrieves tasks with a specific status.
func (s *DatabaseTaskStore) GetByStatus(ctx context.Context, status a2a.TaskState) ([]*a2a.Task, error) {
	var models []server.TaskModel
	db := s.db.WithContext(ctx)
	if s.tableName != "tasks" {
		db = db.Table(s.tableName)
	}

	// Query by status using JSON path
	if err := db.Where("JSON_EXTRACT(status, '$.state') = ?", string(status)).Find(&models).Error; err != nil {
		return nil, NewTaskStoreError("get_by_status", "", err)
	}

	// Convert models to tasks
	tasks := make([]*a2a.Task, len(models))
	for i, model := range models {
		task, err := model.ToTask()
		if err != nil {
			return nil, NewTaskStoreError("get_by_status", model.ID, fmt.Errorf("failed to convert model to task: %w", err))
		}
		tasks[i] = task
	}

	return tasks, nil
}

// Transaction executes a function within a database transaction.
func (s *DatabaseTaskStore) Transaction(ctx context.Context, fn func(TaskStore) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txStore := &DatabaseTaskStore{
			db:          tx,
			tableName:   s.tableName,
			createTable: s.createTable,
		}
		return fn(txStore)
	})
}
