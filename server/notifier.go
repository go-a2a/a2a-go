// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bytedance/sonic"

	"github.com/go-a2a/a2a"
)

// Notifier handles push notifications for tasks
type Notifier struct {
	configs map[string]*a2a.PushNotificationConfig
	mu      sync.RWMutex
	client  *http.Client
}

// NewNotifier creates a new push notification manager
func NewNotifier() *Notifier {
	return &Notifier{
		configs: make(map[string]*a2a.PushNotificationConfig),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetPushNotification sets or updates a push notification config for a task
func (n *Notifier) SetPushNotification(taskID string, config *a2a.PushNotificationConfig) error {
	if config.URL == "" {
		return fmt.Errorf("push notification URL is required")
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	// Store a copy of the config
	configCopy := config
	n.configs[taskID] = configCopy

	return nil
}

// GetPushNotification retrieves a push notification config for a task
func (n *Notifier) GetPushNotification(taskID string) *a2a.PushNotificationConfig {
	n.mu.RLock()
	defer n.mu.RUnlock()

	config, exists := n.configs[taskID]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	configCopy := *config
	return &configCopy
}

// SendStatusUpdate sends a task status update via push notification
func (n *Notifier) SendStatusUpdate(taskID string, status a2a.TaskStatus, isFinal bool) error {
	config := n.GetPushNotification(taskID)
	if config == nil {
		return fmt.Errorf("no push notification config found for task")
	}

	event := a2a.TaskStatusUpdateEvent{
		ID:     taskID,
		Status: status,
		Final:  isFinal,
	}

	return n.sendEvent(config, event)
}

// SendArtifactUpdate sends a task artifact update via push notification
func (n *Notifier) SendArtifactUpdate(taskID string, artifact a2a.Artifact) error {
	config := n.GetPushNotification(taskID)
	if config == nil {
		return fmt.Errorf("no push notification config found for task")
	}

	event := a2a.TaskArtifactUpdateEvent{
		ID:       taskID,
		Artifact: artifact,
	}

	return n.sendEvent(config, event)
}

// sendEvent sends an event to the push notification endpoint
func (n *Notifier) sendEvent(config *a2a.PushNotificationConfig, event any) error {
	data, err := sonic.ConfigDefault.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, config.URL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication if provided
	if config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+config.Token)
	}

	// Add custom authentication if provided
	if config.Authentication != nil {
		// This is a simplified implementation
		// In a real-world scenario, you would handle different auth schemes
		if contains(config.Authentication.Schemes, "bearer") && config.Authentication.Credentials != "" {
			req.Header.Set("Authorization", "Bearer "+config.Authentication.Credentials)
		}
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("notification failed with status code: %d", resp.StatusCode)
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
