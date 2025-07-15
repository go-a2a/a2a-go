// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package task

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-a2a/a2a"
)

// PushNotificationConfigStore defines the interface for storing and retrieving
// push notification configurations. This interface abstracts the storage mechanism
// to allow different implementations (database, in-memory, etc.) while maintaining
// a consistent API for push notification configuration management.
type PushNotificationConfigStore interface {
	// GetConfig retrieves a push notification configuration by task ID.
	// Returns TaskNotFoundError if the configuration doesn't exist.
	GetConfig(ctx context.Context, taskID string) (*a2a.TaskPushNotificationConfig, error)

	// GetInfo retrieves push notification configurations for a task ID.
	// Returns empty slice if no configurations exist for the task.
	// This method mirrors the Python implementation's get_info behavior.
	GetInfo(ctx context.Context, taskID string) ([]*a2a.PushNotificationConfig, error)

	// SaveConfig saves a push notification configuration for a task.
	// If the configuration already exists, it will be updated.
	SaveConfig(ctx context.Context, taskID string, config *a2a.TaskPushNotificationConfig) error

	// DeleteConfig removes a push notification configuration for a task.
	// Returns TaskNotFoundError if the configuration doesn't exist.
	DeleteConfig(ctx context.Context, taskID string) error

	// ListConfigs retrieves all push notification configurations.
	// Returns a map of task ID to configuration.
	ListConfigs(ctx context.Context) (map[string]*a2a.TaskPushNotificationConfig, error)

	// ExistsConfig checks if a push notification configuration exists for a task.
	ExistsConfig(ctx context.Context, taskID string) (bool, error)

	// Initialize prepares the storage for use.
	// This may involve creating tables, indexes, or other setup operations.
	Initialize(ctx context.Context) error

	// Close cleanly shuts down the storage.
	// This should be called when the store is no longer needed.
	Close(ctx context.Context) error
}

// InMemoryPushNotificationConfigStore is an in-memory implementation of PushNotificationConfigStore.
// Configuration data is lost when the server process stops.
// All operations are thread-safe using sync.RWMutex.
type InMemoryPushNotificationConfigStore struct {
	mu      sync.RWMutex
	configs map[string]*a2a.TaskPushNotificationConfig
}

var _ PushNotificationConfigStore = (*InMemoryPushNotificationConfigStore)(nil)

// NewInMemoryPushNotificationConfigStore creates a new in-memory push notification config store.
func NewInMemoryPushNotificationConfigStore() *InMemoryPushNotificationConfigStore {
	return &InMemoryPushNotificationConfigStore{
		configs: make(map[string]*a2a.TaskPushNotificationConfig),
	}
}

// GetConfig retrieves a push notification configuration by task ID.
func (s *InMemoryPushNotificationConfigStore) GetConfig(ctx context.Context, taskID string) (*a2a.TaskPushNotificationConfig, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[taskID]
	if !exists {
		return nil, a2a.TaskNotFoundError{TaskID: taskID}
	}

	// Return a deep copy to avoid race conditions
	return s.copyConfig(config), nil
}

// GetInfo retrieves push notification configurations for a task ID.
// Returns empty slice if no configurations exist for the task.
// This method mirrors the Python implementation's get_info behavior.
func (s *InMemoryPushNotificationConfigStore) GetInfo(ctx context.Context, taskID string) ([]*a2a.PushNotificationConfig, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[taskID]
	if !exists {
		// Return empty slice if no configuration exists, matching Python behavior
		return []*a2a.PushNotificationConfig{}, nil
	}

	// Return the PushNotificationConfig from the TaskPushNotificationConfig
	// Currently we store one config per task, but this allows for future extension
	return []*a2a.PushNotificationConfig{s.copyPushNotificationConfig(config.PushNotificationConfig)}, nil
}

// SaveConfig saves a push notification configuration for a task.
func (s *InMemoryPushNotificationConfigStore) SaveConfig(ctx context.Context, taskID string, config *a2a.TaskPushNotificationConfig) error {
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	if config == nil {
		return fmt.Errorf("push notification config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid push notification config: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a deep copy to avoid race conditions
	s.configs[taskID] = s.copyConfig(config)

	return nil
}

// DeleteConfig removes a push notification configuration for a task.
func (s *InMemoryPushNotificationConfigStore) DeleteConfig(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.configs[taskID]; !exists {
		return a2a.TaskNotFoundError{TaskID: taskID}
	}

	delete(s.configs, taskID)
	return nil
}

// ListConfigs retrieves all push notification configurations.
func (s *InMemoryPushNotificationConfigStore) ListConfigs(ctx context.Context) (map[string]*a2a.TaskPushNotificationConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*a2a.TaskPushNotificationConfig, len(s.configs))
	for taskID, config := range s.configs {
		result[taskID] = s.copyConfig(config)
	}

	return result, nil
}

// ExistsConfig checks if a push notification configuration exists for a task.
func (s *InMemoryPushNotificationConfigStore) ExistsConfig(ctx context.Context, taskID string) (bool, error) {
	if taskID == "" {
		return false, fmt.Errorf("task ID cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.configs[taskID]
	return exists, nil
}

// Initialize prepares the in-memory storage for use.
func (s *InMemoryPushNotificationConfigStore) Initialize(ctx context.Context) error {
	// No initialization needed for in-memory storage
	return nil
}

// Close cleanly shuts down the in-memory storage.
func (s *InMemoryPushNotificationConfigStore) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear all configurations
	s.configs = make(map[string]*a2a.TaskPushNotificationConfig)
	return nil
}

// copyConfig creates a deep copy of a TaskPushNotificationConfig.
func (s *InMemoryPushNotificationConfigStore) copyConfig(config *a2a.TaskPushNotificationConfig) *a2a.TaskPushNotificationConfig {
	if config == nil {
		return nil
	}

	// Create a copy of the push notification config
	pnConfig := &a2a.PushNotificationConfig{
		ID:    config.PushNotificationConfig.ID,
		URL:   config.PushNotificationConfig.URL,
		Token: config.PushNotificationConfig.Token,
	}

	// Copy authentication if present
	if config.PushNotificationConfig.Authentication != nil {
		pnConfig.Authentication = &a2a.AuthenticationInfo{
			Schemes:     make([]string, len(config.PushNotificationConfig.Authentication.Schemes)),
			Credentials: config.PushNotificationConfig.Authentication.Credentials,
		}
		copy(pnConfig.Authentication.Schemes, config.PushNotificationConfig.Authentication.Schemes)
	}

	return &a2a.TaskPushNotificationConfig{
		Name:                   config.Name,
		PushNotificationConfig: pnConfig,
	}
}

// copyPushNotificationConfig creates a deep copy of a PushNotificationConfig.
func (s *InMemoryPushNotificationConfigStore) copyPushNotificationConfig(config *a2a.PushNotificationConfig) *a2a.PushNotificationConfig {
	if config == nil {
		return nil
	}

	// Create a copy of the push notification config
	pnConfig := &a2a.PushNotificationConfig{
		ID:    config.ID,
		URL:   config.URL,
		Token: config.Token,
	}

	// Copy authentication if present
	if config.Authentication != nil {
		pnConfig.Authentication = &a2a.AuthenticationInfo{
			Schemes:     make([]string, len(config.Authentication.Schemes)),
			Credentials: config.Authentication.Credentials,
		}
		copy(pnConfig.Authentication.Schemes, config.Authentication.Schemes)
	}

	return pnConfig
}

// GetConfigCount returns the number of configurations stored.
// This is useful for testing and monitoring purposes.
func (s *InMemoryPushNotificationConfigStore) GetConfigCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.configs)
}

// Clear removes all configurations from the store.
// This is useful for testing purposes.
func (s *InMemoryPushNotificationConfigStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.configs = make(map[string]*a2a.TaskPushNotificationConfig)
}

// GetAllTaskIDs returns all task IDs that have push notification configurations.
// This is useful for administration and monitoring purposes.
func (s *InMemoryPushNotificationConfigStore) GetAllTaskIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	taskIDs := make([]string, 0, len(s.configs))
	for taskID := range s.configs {
		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs
}

// UpdateConfig updates an existing push notification configuration.
// Returns TaskNotFoundError if the configuration doesn't exist.
func (s *InMemoryPushNotificationConfigStore) UpdateConfig(ctx context.Context, taskID string, config *a2a.TaskPushNotificationConfig) error {
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	if config == nil {
		return fmt.Errorf("push notification config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid push notification config: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.configs[taskID]; !exists {
		return a2a.TaskNotFoundError{TaskID: taskID}
	}

	// Update the configuration
	s.configs[taskID] = s.copyConfig(config)

	return nil
}

// GetConfigByName retrieves a push notification configuration by name.
// Returns nil if no configuration with the given name exists.
func (s *InMemoryPushNotificationConfigStore) GetConfigByName(ctx context.Context, name string) (*a2a.TaskPushNotificationConfig, string, error) {
	if name == "" {
		return nil, "", fmt.Errorf("config name cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for taskID, config := range s.configs {
		if config.Name == name {
			return s.copyConfig(config), taskID, nil
		}
	}

	return nil, "", nil
}
