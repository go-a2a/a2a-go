// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package task

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-a2a/a2a"
)

// InMemoryTaskStore is an in-memory implementation of TaskStore.
// Task data is lost when the server process stops.
// All operations are thread-safe using sync.RWMutex.
type InMemoryTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*a2a.Task
}

var _ TaskStore = (*InMemoryTaskStore)(nil)

// NewInMemoryTaskStore creates a new InMemoryTaskStore.
func NewInMemoryTaskStore() *InMemoryTaskStore {
	return &InMemoryTaskStore{
		tasks: make(map[string]*a2a.Task),
	}
}

// Save persists a task to the in-memory storage.
func (s *InMemoryTaskStore) Save(ctx context.Context, task *a2a.Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	if err := task.Validate(); err != nil {
		return NewTaskValidationError(task.ID, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a deep copy to avoid race conditions
	taskCopy := s.copyTask(task)
	s.tasks[task.ID] = taskCopy

	return nil
}

// Get retrieves a task by its ID from the in-memory storage.
func (s *InMemoryTaskStore) Get(ctx context.Context, taskID string) (*a2a.Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, a2a.TaskNotFoundError{TaskID: taskID}
	}

	// Return a deep copy to avoid race conditions
	return s.copyTask(task), nil
}

// Delete removes a task from the in-memory storage.
func (s *InMemoryTaskStore) Delete(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[taskID]; !exists {
		return a2a.TaskNotFoundError{TaskID: taskID}
	}

	delete(s.tasks, taskID)
	return nil
}

// List retrieves tasks with optional filtering.
func (s *InMemoryTaskStore) List(ctx context.Context, contextID string, limit, offset int) ([]*a2a.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tasks []*a2a.Task
	count := 0

	for _, task := range s.tasks {
		// Filter by context ID if provided
		if contextID != "" && task.ContextID != contextID {
			continue
		}

		// Apply offset
		if count < offset {
			count++
			continue
		}

		// Apply limit
		if limit > 0 && len(tasks) >= limit {
			break
		}

		tasks = append(tasks, s.copyTask(task))
		count++
	}

	return tasks, nil
}

// Count returns the total number of tasks in the in-memory storage.
func (s *InMemoryTaskStore) Count(ctx context.Context, contextID string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if contextID == "" {
		return int64(len(s.tasks)), nil
	}

	count := int64(0)
	for _, task := range s.tasks {
		if task.ContextID == contextID {
			count++
		}
	}

	return count, nil
}

// Initialize prepares the in-memory storage for use.
func (s *InMemoryTaskStore) Initialize(ctx context.Context) error {
	// No initialization needed for in-memory storage
	return nil
}

// Close cleanly shuts down the in-memory storage.
func (s *InMemoryTaskStore) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear all tasks
	s.tasks = make(map[string]*a2a.Task)
	return nil
}

// copyTask creates a deep copy of a task to avoid race conditions.
func (s *InMemoryTaskStore) copyTask(task *a2a.Task) *a2a.Task {
	if task == nil {
		return nil
	}

	copy := &a2a.Task{
		ID:        task.ID,
		ContextID: task.ContextID,
		Status:    task.Status,
		History:   make([]*a2a.Message, len(task.History)),
		Artifacts: make([]*a2a.Artifact, len(task.Artifacts)),
	}

	// Copy history messages
	for i, message := range task.History {
		if message != nil {
			copy.History[i] = &a2a.Message{
				ContextID: message.ContextID,
				Content:   message.Content,
				Metadata:  s.copyMetadata(message.Metadata),
			}
		}
	}

	// Copy artifacts
	for i, artifact := range task.Artifacts {
		if artifact != nil {
			copy.Artifacts[i] = &a2a.Artifact{
				ArtifactID:  artifact.ArtifactID,
				Name:        artifact.Name,
				Description: artifact.Description,
				Parts:       s.copyArtifactParts(artifact.Parts),
				Extensions:  s.copyExtensions(artifact.Extensions),
				Metadata:    s.copyMetadata(artifact.Metadata),
			}
		}
	}

	return copy
}

// copyMetadata creates a deep copy of metadata map.
func (s *InMemoryTaskStore) copyMetadata(metadata map[string]any) map[string]any {
	if metadata == nil {
		return nil
	}

	copy := make(map[string]any, len(metadata))
	for k, v := range metadata {
		copy[k] = v
	}
	return copy
}

// copyArtifactParts creates a deep copy of artifact parts.
func (s *InMemoryTaskStore) copyArtifactParts(parts []*a2a.PartWrapper) []*a2a.PartWrapper {
	if parts == nil {
		return nil
	}

	copy := make([]*a2a.PartWrapper, len(parts))
	for i, part := range parts {
		if part != nil {
			// For now, we'll just copy the reference since PartWrapper is immutable
			// In a production environment, you might want to implement deep copying
			copy[i] = part
		}
	}
	return copy
}

// copyExtensions creates a deep copy of extensions slice.
func (s *InMemoryTaskStore) copyExtensions(extensions []string) []string {
	if extensions == nil {
		return nil
	}

	copy := make([]string, len(extensions))
	for i, ext := range extensions {
		copy[i] = ext
	}
	return copy
}

// Clear removes all tasks from the in-memory storage.
// This is useful for testing purposes.
func (s *InMemoryTaskStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks = make(map[string]*a2a.Task)
}

// Size returns the current number of tasks in the in-memory storage.
// This is useful for testing and monitoring purposes.
func (s *InMemoryTaskStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.tasks)
}
