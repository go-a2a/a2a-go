// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"errors"
	"sync"
	"time"

	"github.com/go-a2a/a2a"
)

// TaskManager defines the interface for managing A2A tasks
type TaskManager interface {
	// ProcessTask processes a task synchronously and returns the result
	ProcessTask(task *a2a.Task) (*a2a.Task, error)

	// StartTask begins processing a task asynchronously with streaming updates
	StartTask(task *a2a.Task, stream *Stream) error

	// GetTask retrieves a task by ID
	GetTask(id string) (*a2a.Task, error)

	// CancelTask cancels a task
	CancelTask(id string) (*a2a.Task, error)

	// GetTaskHistory retrieves history for a task with specified length limit
	GetTaskHistory(id string, limit int) ([]a2a.Message, error)

	// ResumeTaskUpdates resumes sending updates for an existing task to a new stream
	ResumeTaskUpdates(task *a2a.Task, stream *Stream) error
}

// InMemoryTaskManager implements TaskManager with in-memory storage
type InMemoryTaskManager struct {
	tasks    map[string]*a2a.Task
	history  map[string][]a2a.Message
	mu       sync.RWMutex
	handlers map[string]func(*a2a.Task, *Stream) error
}

// NewInMemoryTaskManager creates a new in-memory task manager
func NewInMemoryTaskManager(taskHandlers map[string]func(*a2a.Task, *Stream) error) *InMemoryTaskManager {
	return &InMemoryTaskManager{
		tasks:    make(map[string]*a2a.Task),
		history:  make(map[string][]a2a.Message),
		handlers: taskHandlers,
	}
}

// ProcessTask processes a task synchronously and returns the result
func (m *InMemoryTaskManager) ProcessTask(task *a2a.Task) (*a2a.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store the task
	m.tasks[task.ID] = task

	// Store message in history if it exists
	if task.Status.Message != nil {
		m.addToHistory(task.ID, *task.Status.Message)
	}

	// Update the task status to working
	task.Status.State = a2a.TaskStateWorking
	task.Status.Timestamp = time.Now()

	// Find an appropriate handler based on the message content
	// This is a simplified approach - in a real implementation, you would
	// have more sophisticated logic to route tasks
	handler := m.getDefaultHandler()
	if handler == nil {
		task.Status.State = a2a.TaskStateFailed
		task.Status.Timestamp = time.Now()
		task.Status.Message = &a2a.Message{
			Role: a2a.MessageRoleAgent,
			Parts: []a2a.Part{
				{
					Type: a2a.PartTypeText,
					Text: "No handler available for this task",
				},
			},
		}
		return task, nil
	}

	// Process the task (mock simple response)
	// In a real implementation, this would call a model or service
	task.Status.State = a2a.TaskStateCompleted
	task.Status.Timestamp = time.Now()
	task.Status.Message = &a2a.Message{
		Role: a2a.MessageRoleAgent,
		Parts: []a2a.Part{
			{
				Type: a2a.PartTypeText,
				Text: "Task processed successfully",
			},
		},
	}

	// Store the response in history
	if task.Status.Message != nil {
		m.addToHistory(task.ID, *task.Status.Message)
	}

	// Generate a simple artifact
	artifact := a2a.Artifact{
		Name:  "Response",
		Parts: task.Status.Message.Parts,
		Index: 0,
	}
	task.Artifacts = append(task.Artifacts, artifact)

	return task, nil
}

// StartTask begins processing a task asynchronously with streaming updates
func (m *InMemoryTaskManager) StartTask(task *a2a.Task, stream *Stream) error {
	m.mu.Lock()

	// Store the task
	m.tasks[task.ID] = task

	// Store message in history if it exists
	if task.Status.Message != nil {
		m.addToHistory(task.ID, *task.Status.Message)
	}

	// Update task status to working
	task.Status.State = a2a.TaskStateWorking
	task.Status.Timestamp = time.Now()

	// Send initial status update
	initialStatus := a2a.TaskStatusUpdateEvent{
		ID:     task.ID,
		Status: task.Status,
		Final:  false,
	}

	// Signal status update via stream
	if err := stream.SendStatusUpdate(initialStatus); err != nil {
		m.mu.Unlock()
		return err
	}

	// Find an appropriate handler
	handler := m.getDefaultHandler()
	if handler == nil {
		// No handler found
		task.Status.State = a2a.TaskStateFailed
		task.Status.Timestamp = time.Now()
		task.Status.Message = &a2a.Message{
			Role: a2a.MessageRoleAgent,
			Parts: []a2a.Part{
				{
					Type: a2a.PartTypeText,
					Text: "No handler available for this task",
				},
			},
		}

		// Send final status
		finalStatus := a2a.TaskStatusUpdateEvent{
			ID:     task.ID,
			Status: task.Status,
			Final:  true,
		}
		stream.SendStatusUpdate(finalStatus)
		m.mu.Unlock()
		return nil
	}

	m.mu.Unlock()

	// Execute handler in a goroutine
	go handler(task, stream)

	return nil
}

// GetTask retrieves a task by ID
func (m *InMemoryTaskManager) GetTask(id string) (*a2a.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[id]
	if !exists {
		return nil, errors.New("task not found")
	}

	return task, nil
}

// CancelTask cancels a task
func (m *InMemoryTaskManager) CancelTask(id string) (*a2a.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[id]
	if !exists {
		return nil, errors.New("task not found")
	}

	// Update task status to canceled
	task.Status.State = a2a.TaskStateCanceled
	task.Status.Timestamp = time.Now()
	task.Status.Message = &a2a.Message{
		Role: a2a.MessageRoleAgent,
		Parts: []a2a.Part{
			{
				Type: a2a.PartTypeText,
				Text: "Task was canceled",
			},
		},
	}

	// Store the cancellation message in history
	if task.Status.Message != nil {
		m.addToHistory(task.ID, *task.Status.Message)
	}

	return task, nil
}

// GetTaskHistory retrieves history for a task with specified length limit
func (m *InMemoryTaskManager) GetTaskHistory(id string, limit int) ([]a2a.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, exists := m.history[id]
	if !exists {
		return nil, errors.New("history not found for task")
	}

	// If limit is 0 or negative, return all history
	if limit <= 0 || limit >= len(history) {
		return history, nil
	}

	// Return the most recent entries up to the limit
	start := len(history) - limit
	return history[start:], nil
}

// ResumeTaskUpdates resumes sending updates for an existing task to a new stream
func (m *InMemoryTaskManager) ResumeTaskUpdates(task *a2a.Task, stream *Stream) error {
	// Send current status
	statusEvent := a2a.TaskStatusUpdateEvent{
		ID:     task.ID,
		Status: task.Status,
		Final:  isTerminalState(task.Status.State),
	}

	if err := stream.SendStatusUpdate(statusEvent); err != nil {
		return err
	}

	// Send existing artifacts
	for _, artifact := range task.Artifacts {
		artifactEvent := a2a.TaskArtifactUpdateEvent{
			ID:       task.ID,
			Artifact: artifact,
		}

		if err := stream.SendArtifactUpdate(artifactEvent); err != nil {
			return err
		}
	}

	// If task is complete, send final status update
	if isTerminalState(task.Status.State) {
		finalEvent := a2a.TaskStatusUpdateEvent{
			ID:     task.ID,
			Status: task.Status,
			Final:  true,
		}

		if err := stream.SendStatusUpdate(finalEvent); err != nil {
			return err
		}
	}

	return nil
}

// addToHistory adds a message to the task history
func (m *InMemoryTaskManager) addToHistory(taskID string, message a2a.Message) {
	if _, exists := m.history[taskID]; !exists {
		m.history[taskID] = []a2a.Message{}
	}
	m.history[taskID] = append(m.history[taskID], message)
}

// getDefaultHandler returns a default task handler for demo purposes
func (m *InMemoryTaskManager) getDefaultHandler() func(*a2a.Task, *Stream) error {
	// In a real implementation, this would select a handler based on the task content
	return func(task *a2a.Task, stream *Stream) error {
		// Simulate processing time
		time.Sleep(1 * time.Second)

		// Send an artifact
		artifact := a2a.Artifact{
			Name:  "Response",
			Index: 0,
			Parts: []a2a.Part{
				{
					Type: a2a.PartTypeText,
					Text: "Processing task...",
				},
			},
		}

		artifactEvent := a2a.TaskArtifactUpdateEvent{
			ID:       task.ID,
			Artifact: artifact,
		}

		if err := stream.SendArtifactUpdate(artifactEvent); err != nil {
			return err
		}

		// Simulate more processing
		time.Sleep(1 * time.Second)

		// Update task status to completed
		m.mu.Lock()
		task.Status.State = a2a.TaskStateCompleted
		task.Status.Timestamp = time.Now()
		task.Status.Message = &a2a.Message{
			Role: a2a.MessageRoleAgent,
			Parts: []a2a.Part{
				{
					Type: a2a.PartTypeText,
					Text: "Task processed successfully",
				},
			},
		}

		// Store response in history
		if task.Status.Message != nil {
			m.addToHistory(task.ID, *task.Status.Message)
		}

		// Add final artifact
		finalArtifact := a2a.Artifact{
			Name:  "Final Response",
			Index: 1,
			Parts: task.Status.Message.Parts,
		}
		task.Artifacts = append(task.Artifacts, finalArtifact)
		m.mu.Unlock()

		// Send final artifact
		finalArtifactEvent := a2a.TaskArtifactUpdateEvent{
			ID:       task.ID,
			Artifact: finalArtifact,
		}

		if err := stream.SendArtifactUpdate(finalArtifactEvent); err != nil {
			return err
		}

		// Send final status update
		finalStatus := a2a.TaskStatusUpdateEvent{
			ID:     task.ID,
			Status: task.Status,
			Final:  true,
		}

		return stream.SendStatusUpdate(finalStatus)
	}
}

// isTerminalState checks if a task state is terminal
func isTerminalState(state a2a.TaskState) bool {
	return state == a2a.TaskStateCompleted ||
		state == a2a.TaskStateCanceled ||
		state == a2a.TaskStateFailed
}
