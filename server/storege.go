package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/go-a2a/a2a"
)

// FileTaskStore is a file-based implementation of TaskStore
type FileTaskStore struct {
	// Directory is the directory where tasks are stored
	Directory string

	// IndexFile is the file for storing the index of tasks
	IndexFile string

	// SessionIndexFile is the file for storing the session index
	SessionIndexFile string

	// TaskExtension is the file extension for task files
	TaskExtension string

	// mutex is a mutex for synchronizing access to the store
	mutex sync.RWMutex

	// taskIndex is an in-memory index of tasks
	taskIndex map[string]string

	// sessionIndex is an in-memory index of sessions to tasks
	sessionIndex map[string][]string
}

// NewFileTaskStore creates a new file-based task store
func NewFileTaskStore(directory string) (*FileTaskStore, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	store := &FileTaskStore{
		Directory:        directory,
		IndexFile:        filepath.Join(directory, "tasks.index"),
		SessionIndexFile: filepath.Join(directory, "sessions.index"),
		TaskExtension:    ".task.json",
		taskIndex:        make(map[string]string),
		sessionIndex:     make(map[string][]string),
	}

	// Load index
	if err := store.loadIndexes(); err != nil {
		return nil, fmt.Errorf("failed to load indexes: %w", err)
	}

	return store, nil
}

// loadIndexes loads task and session indexes from disk
func (s *FileTaskStore) loadIndexes() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Initialize with empty indexes
	s.taskIndex = make(map[string]string)
	s.sessionIndex = make(map[string][]string)

	// Load task index if it exists
	if _, err := os.Stat(s.IndexFile); err == nil {
		data, err := os.ReadFile(s.IndexFile)
		if err != nil {
			return fmt.Errorf("failed to read index file: %w", err)
		}

		if err := json.Unmarshal(data, &s.taskIndex); err != nil {
			return fmt.Errorf("failed to parse index file: %w", err)
		}
	}

	// Load session index if it exists
	if _, err := os.Stat(s.SessionIndexFile); err == nil {
		data, err := os.ReadFile(s.SessionIndexFile)
		if err != nil {
			return fmt.Errorf("failed to read session index file: %w", err)
		}

		if err := json.Unmarshal(data, &s.sessionIndex); err != nil {
			return fmt.Errorf("failed to parse session index file: %w", err)
		}
	}

	return nil
}

// saveIndexes saves task and session indexes to disk
func (s *FileTaskStore) saveIndexes() error {
	// Save task index
	taskIndexData, err := json.Marshal(s.taskIndex)
	if err != nil {
		return fmt.Errorf("failed to marshal task index: %w", err)
	}

	if err := os.WriteFile(s.IndexFile, taskIndexData, 0o644); err != nil {
		return fmt.Errorf("failed to write task index: %w", err)
	}

	// Save session index
	sessionIndexData, err := json.Marshal(s.sessionIndex)
	if err != nil {
		return fmt.Errorf("failed to marshal session index: %w", err)
	}

	if err := os.WriteFile(s.SessionIndexFile, sessionIndexData, 0o644); err != nil {
		return fmt.Errorf("failed to write session index: %w", err)
	}

	return nil
}

// getTaskFilePath returns the file path for a task
func (s *FileTaskStore) getTaskFilePath(id string) string {
	filename := fmt.Sprintf("%s%s", id, s.TaskExtension)
	return filepath.Join(s.Directory, filename)
}

// GetTask retrieves a task by ID
func (s *FileTaskStore) GetTask(ctx context.Context, id string) (*a2a.Task, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if task exists in index
	_, ok := s.taskIndex[id]
	if !ok {
		return nil, ErrTaskNotFound
	}

	// Read task file
	filePath := s.getTaskFilePath(id)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to read task file: %w", err)
	}

	// Parse task
	var task a2a.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to parse task file: %w", err)
	}

	return &task, nil
}

// CreateTask creates a new task
func (s *FileTaskStore) CreateTask(ctx context.Context, task *a2a.Task) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if task already exists
	if _, ok := s.taskIndex[task.ID]; ok {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	}

	// Serialize task
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Write task file
	filePath := s.getTaskFilePath(task.ID)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	// Update indexes
	s.taskIndex[task.ID] = filePath
	if task.SessionID != "" {
		s.sessionIndex[task.SessionID] = append(s.sessionIndex[task.SessionID], task.ID)
	}

	// Save indexes
	return s.saveIndexes()
}

// UpdateTask updates an existing task
func (s *FileTaskStore) UpdateTask(ctx context.Context, task *a2a.Task) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if task exists
	if _, ok := s.taskIndex[task.ID]; !ok {
		return ErrTaskNotFound
	}

	// Serialize task
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Write task file
	filePath := s.getTaskFilePath(task.ID)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	return nil
}

// DeleteTask deletes a task by ID
func (s *FileTaskStore) DeleteTask(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if task exists
	filePath, ok := s.taskIndex[id]
	if !ok {
		return ErrTaskNotFound
	}

	// Get task to find session ID
	var task *a2a.Task
	taskData, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read task file: %w", err)
	}

	if taskData != nil {
		if err := json.Unmarshal(taskData, &task); err == nil && task.SessionID != "" {
			// Remove from session index
			taskIDs := s.sessionIndex[task.SessionID]
			for i, taskID := range taskIDs {
				if taskID == id {
					s.sessionIndex[task.SessionID] = append(taskIDs[:i], taskIDs[i+1:]...)
					if len(s.sessionIndex[task.SessionID]) == 0 {
						delete(s.sessionIndex, task.SessionID)
					}
					break
				}
			}
		}
	}

	// Delete task file
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete task file: %w", err)
	}

	// Update indexes
	delete(s.taskIndex, id)

	// Save indexes
	return s.saveIndexes()
}

// ListTasks lists all tasks, optionally filtered by session ID
func (s *FileTaskStore) ListTasks(ctx context.Context, sessionID string) ([]*a2a.Task, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var taskIDs []string
	if sessionID != "" {
		// Get tasks for session
		taskIDs = append(taskIDs, s.sessionIndex[sessionID]...)
	} else {
		// Get all tasks
		for id := range s.taskIndex {
			taskIDs = append(taskIDs, id)
		}
	}

	// Sort task IDs by creation time
	sort.Slice(taskIDs, func(i, j int) bool {
		taskA, err := s.loadTaskWithoutLock(taskIDs[i])
		if err != nil {
			return false
		}
		taskB, err := s.loadTaskWithoutLock(taskIDs[j])
		if err != nil {
			return true
		}
		return taskA.CreatedAt.After(taskB.CreatedAt) // Sort by newest first
	})

	// Load tasks
	var tasks []*a2a.Task
	for _, id := range taskIDs {
		task, err := s.loadTaskWithoutLock(id)
		if err != nil {
			// Skip tasks that can't be loaded
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// loadTaskWithoutLock loads a task from disk without locking
func (s *FileTaskStore) loadTaskWithoutLock(id string) (*a2a.Task, error) {
	// Check if task exists in index
	_, ok := s.taskIndex[id]
	if !ok {
		return nil, ErrTaskNotFound
	}

	// Read task file
	filePath := s.getTaskFilePath(id)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to read task file: %w", err)
	}

	// Parse task
	var task a2a.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to parse task file: %w", err)
	}

	return &task, nil
}

// Cleanup removes completed, failed, and canceled tasks older than the specified duration
func (s *FileTaskStore) Cleanup(ctx context.Context, maxAge time.Duration) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	var removedCount int

	// Find tasks to remove
	var toRemove []string
	for id := range s.taskIndex {
		task, err := s.loadTaskWithoutLock(id)
		if err != nil {
			continue
		}

		// Check if task is completed, failed, or canceled and older than cutoff
		if (task.Status.State == a2a.TaskCompleted ||
			task.Status.State == a2a.TaskFailed ||
			task.Status.State == a2a.TaskCanceled) &&
			task.UpdatedAt.Before(cutoff) {
			toRemove = append(toRemove, id)
		}
	}

	// Remove tasks
	for _, id := range toRemove {
		// Get task to find session ID
		task, err := s.loadTaskWithoutLock(id)
		if err != nil {
			continue
		}

		// Remove task file
		filePath := s.getTaskFilePath(id)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			continue
		}

		// Update indexes
		delete(s.taskIndex, id)

		// Remove from session index
		if task.SessionID != "" {
			taskIDs := s.sessionIndex[task.SessionID]
			for i, taskID := range taskIDs {
				if taskID == id {
					s.sessionIndex[task.SessionID] = append(taskIDs[:i], taskIDs[i+1:]...)
					if len(s.sessionIndex[task.SessionID]) == 0 {
						delete(s.sessionIndex, task.SessionID)
					}
					break
				}
			}
		}

		removedCount++
	}

	// Save indexes
	if removedCount > 0 {
		if err := s.saveIndexes(); err != nil {
			return removedCount, fmt.Errorf("failed to save indexes: %w", err)
		}
	}

	return removedCount, nil
}

// GetTasksByStatus returns tasks with the specified status
func (s *FileTaskStore) GetTasksByStatus(ctx context.Context, states []a2a.TaskState, limit int) ([]*a2a.Task, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var matchingTasks []*a2a.Task
	count := 0

	// Create a map for faster lookups
	stateMap := make(map[a2a.TaskState]bool)
	for _, state := range states {
		stateMap[state] = true
	}

	// Check all tasks
	for id := range s.taskIndex {
		if limit > 0 && count >= limit {
			break
		}

		task, err := s.loadTaskWithoutLock(id)
		if err != nil {
			continue
		}

		if stateMap[task.Status.State] {
			matchingTasks = append(matchingTasks, task)
			count++
		}
	}

	// Sort tasks by update time (newest first)
	sort.Slice(matchingTasks, func(i, j int) bool {
		return matchingTasks[i].UpdatedAt.After(matchingTasks[j].UpdatedAt)
	})

	return matchingTasks, nil
}

// GetTaskCountBySession returns the number of tasks for each session
func (s *FileTaskStore) GetTaskCountBySession(ctx context.Context) (map[string]int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make(map[string]int)
	for sessionID, taskIDs := range s.sessionIndex {
		result[sessionID] = len(taskIDs)
	}

	return result, nil
}

// RunPeriodicCleanup runs the cleanup task periodically
func (s *FileTaskStore) RunPeriodicCleanup(ctx context.Context, interval, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := s.Cleanup(ctx, maxAge)
			if err != nil {
				// Log error but continue
				fmt.Printf("Error during periodic cleanup: %v\n", err)
			} else if count > 0 {
				fmt.Printf("Periodic cleanup removed %d tasks\n", count)
			}
		}
	}
}
