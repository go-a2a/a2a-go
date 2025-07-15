// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package task

import (
	"context"

	"github.com/go-a2a/a2a"
)

// TaskStore defines the interface for task persistence operations.
// This interface abstracts the storage mechanism to allow different
// implementations (database, in-memory, etc.) while maintaining a
// consistent API for task management operations.
type TaskStore interface {
	// Save persists a task to the storage backend.
	// If the task already exists, it will be updated.
	Save(ctx context.Context, task *a2a.Task) error

	// Get retrieves a task by its ID from the storage backend.
	// Returns TaskNotFoundError if the task doesn't exist.
	Get(ctx context.Context, taskID string) (*a2a.Task, error)

	// Delete removes a task from the storage backend.
	// Returns TaskNotFoundError if the task doesn't exist.
	Delete(ctx context.Context, taskID string) error

	// List retrieves tasks with optional filtering.
	// The contextID parameter can be used to filter tasks by context.
	// If contextID is empty, all tasks are returned.
	List(ctx context.Context, contextID string, limit, offset int) ([]*a2a.Task, error)

	// Count returns the total number of tasks in the storage backend.
	// The contextID parameter can be used to count tasks by context.
	// If contextID is empty, all tasks are counted.
	Count(ctx context.Context, contextID string) (int64, error)

	// Initialize prepares the storage backend for use.
	// This may involve creating tables, indexes, or other setup operations.
	Initialize(ctx context.Context) error

	// Close cleanly shuts down the storage backend.
	// This should be called when the store is no longer needed.
	Close(ctx context.Context) error
}
