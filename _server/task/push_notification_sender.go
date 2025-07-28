// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package task

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-json-experiment/json"

	a2a "github.com/go-a2a/a2a-go"
)

// PushNotificationSender defines the interface for sending push notifications.
// This interface abstracts the notification mechanism to allow different
// implementations (HTTP, websocket, etc.) while maintaining a consistent
// API for push notification operations.
type PushNotificationSender interface {
	// SendNotification sends a push notification for a task event.
	// The notification is sent to the configured endpoint with the task information.
	SendNotification(ctx context.Context, taskID string, event any, config *a2a.PushNotificationConfig) error

	// SendTaskStatusNotification sends a notification when a task status changes.
	SendTaskStatusNotification(ctx context.Context, taskID string, status a2a.TaskStatus, config *a2a.PushNotificationConfig) error

	// SendTaskArtifactNotification sends a notification when a task artifact is updated.
	SendTaskArtifactNotification(ctx context.Context, taskID string, artifact *a2a.Artifact, config *a2a.PushNotificationConfig) error

	// SendTaskCompletedNotification sends a notification when a task is completed.
	SendTaskCompletedNotification(ctx context.Context, task *a2a.Task, config *a2a.PushNotificationConfig) error

	// SendTaskNotification sends a notification for a task using internal configuration store.
	// This method mirrors the Python implementation's behavior by fetching configurations
	// internally and sending notifications to all configured endpoints concurrently.
	SendTaskNotification(ctx context.Context, task *a2a.Task) error

	// Close shuts down the notification sender and releases resources.
	Close() error
}

// HTTPPushNotificationSender implements PushNotificationSender using HTTP requests.
type HTTPPushNotificationSender struct {
	client      *http.Client
	timeout     time.Duration
	configStore PushNotificationConfigStore
	logger      *slog.Logger
	mu          sync.RWMutex
}

var _ PushNotificationSender = (*HTTPPushNotificationSender)(nil)

// HTTPPushNotificationSenderConfig holds configuration for HTTPPushNotificationSender.
type HTTPPushNotificationSenderConfig struct {
	Client      *http.Client
	Timeout     time.Duration
	ConfigStore PushNotificationConfigStore
	Logger      *slog.Logger
}

// NewHTTPPushNotificationSender creates a new HTTP-based push notification sender.
func NewHTTPPushNotificationSender(config HTTPPushNotificationSenderConfig) *HTTPPushNotificationSender {
	client := config.Client
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &HTTPPushNotificationSender{
		client:      client,
		timeout:     timeout,
		configStore: config.ConfigStore,
		logger:      logger,
	}
}

// SendNotification sends a push notification for a task event.
func (s *HTTPPushNotificationSender) SendNotification(ctx context.Context, taskID string, event any, config *a2a.PushNotificationConfig) error {
	if config == nil {
		return fmt.Errorf("push notification config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid push notification config: %w", err)
	}

	// Create notification payload
	payload := map[string]any{
		"task_id":   taskID,
		"event":     event,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	return s.sendHTTPNotification(ctx, config, payload)
}

// SendTaskStatusNotification sends a notification when a task status changes.
func (s *HTTPPushNotificationSender) SendTaskStatusNotification(ctx context.Context, taskID string, status a2a.TaskStatus, config *a2a.PushNotificationConfig) error {
	if err := status.Validate(); err != nil {
		return fmt.Errorf("invalid task status: %w", err)
	}

	event := map[string]any{
		"type":   "task_status_update",
		"status": status,
	}

	return s.SendNotification(ctx, taskID, event, config)
}

// SendTaskArtifactNotification sends a notification when a task artifact is updated.
func (s *HTTPPushNotificationSender) SendTaskArtifactNotification(ctx context.Context, taskID string, artifact *a2a.Artifact, config *a2a.PushNotificationConfig) error {
	if artifact == nil {
		return fmt.Errorf("artifact cannot be nil")
	}

	if err := artifact.Validate(); err != nil {
		return fmt.Errorf("invalid artifact: %w", err)
	}

	event := map[string]any{
		"type":     "task_artifact_update",
		"artifact": artifact,
	}

	return s.SendNotification(ctx, taskID, event, config)
}

// SendTaskCompletedNotification sends a notification when a task is completed.
func (s *HTTPPushNotificationSender) SendTaskCompletedNotification(ctx context.Context, task *a2a.Task, config *a2a.PushNotificationConfig) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	if err := task.Validate(); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	event := map[string]any{
		"type": "task_completed",
		"task": task,
	}

	return s.SendNotification(ctx, task.ID, event, config)
}

// SendTaskNotification sends a notification for a task using internal configuration store.
// This method mirrors the Python implementation's behavior by fetching configurations
// internally and sending notifications to all configured endpoints concurrently.
func (s *HTTPPushNotificationSender) SendTaskNotification(ctx context.Context, task *a2a.Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	if err := task.Validate(); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	// Check if config store is available
	if s.configStore == nil {
		s.logger.Debug("No configuration store available, skipping notification", "task_id", task.ID)
		return nil
	}

	// Fetch push notification configurations using GetInfo to match Python behavior
	configs, err := s.configStore.GetInfo(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch push notification configs: %w", err)
	}

	// If no configs exist, return early like Python implementation
	if len(configs) == 0 {
		s.logger.Debug("No push notification configurations found", "task_id", task.ID)
		return nil
	}

	return s.sendConcurrentNotifications(ctx, task, configs)
}

// sendConcurrentNotifications sends notifications to multiple endpoints concurrently.
// This mirrors the Python implementation's asyncio.gather() behavior.
func (s *HTTPPushNotificationSender) sendConcurrentNotifications(ctx context.Context, task *a2a.Task, configs []*a2a.PushNotificationConfig) error {
	type result struct {
		config  *a2a.PushNotificationConfig
		success bool
	}

	results := make(chan result, len(configs))
	var wg sync.WaitGroup

	// Send notifications concurrently like Python's asyncio.gather()
	for _, config := range configs {
		wg.Add(1)
		go func(cfg *a2a.PushNotificationConfig) {
			defer wg.Done()
			success := s.dispatchTaskNotification(ctx, task, cfg)
			results <- result{config: cfg, success: success}
		}(config)
	}

	// Wait for all notifications to complete
	wg.Wait()
	close(results)

	// Collect results and handle partial failures like Python implementation
	var successCount int
	var failureCount int

	for res := range results {
		if res.success {
			successCount++
		} else {
			failureCount++
		}
	}

	// Log warning for partial failures like Python implementation:
	// "Some push notifications failed to send for task_id={task.id}"
	if failureCount > 0 && successCount > 0 {
		s.logger.Warn("Some push notifications failed to send",
			"task_id", task.ID,
			"success_count", successCount,
			"failure_count", failureCount)
	}

	// Python implementation doesn't return error, just logs warning
	// So we mirror that behavior - no error returned even if some fail
	return nil
}

// dispatchTaskNotification sends a notification to a single endpoint.
// This mirrors the Python implementation's _dispatch_notification method.
// Returns true on success, false on failure.
func (s *HTTPPushNotificationSender) dispatchTaskNotification(ctx context.Context, task *a2a.Task, config *a2a.PushNotificationConfig) bool {
	if config == nil {
		s.logger.Error("Push notification config cannot be nil",
			"task_id", task.ID)
		return false
	}

	if err := config.Validate(); err != nil {
		s.logger.Error("Invalid push notification config",
			"task_id", task.ID,
			"url", config.URL,
			"error", err)
		return false
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Serialize the complete task object like Python implementation:
	// task.model_dump(mode='json', exclude_none=True)
	jsonData, err := json.Marshal(task)
	if err != nil {
		s.logger.Error("Failed to marshal task",
			"task_id", task.ID,
			"url", config.URL,
			"error", err)
		return false
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(timeoutCtx, "POST", config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		s.logger.Error("Failed to create HTTP request",
			"task_id", task.ID,
			"url", config.URL,
			"error", err)
		return false
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "a2a-go-push-notification-sender")

	// Add Python-style token authentication: {'X-A2A-Notification-Token': push_info.token}
	if config.Token != "" {
		req.Header.Set("X-A2A-Notification-Token", config.Token)
	}

	// Add authentication if provided
	if config.Authentication != nil {
		if err := s.addAuthentication(req, config.Authentication); err != nil {
			s.logger.Error("Failed to add authentication",
				"task_id", task.ID,
				"url", config.URL,
				"error", err)
			return false
		}
	}

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("Error sending push-notification",
			"task_id", task.ID,
			"url", config.URL,
			"error", err)
		return false
	}
	defer resp.Body.Close()

	// Check response status (response.raise_for_status() equivalent)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("Error sending push-notification",
			"task_id", task.ID,
			"url", config.URL,
			"error", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)))
		return false
	}

	// Success - log like Python implementation
	s.logger.Info("Push-notification sent",
		"task_id", task.ID,
		"url", config.URL)
	return true
}

// Close shuts down the notification sender and releases resources.
func (s *HTTPPushNotificationSender) Close() error {
	// HTTP client doesn't need explicit closing
	return nil
}

// sendHTTPNotification sends an HTTP POST request with the notification payload.
func (s *HTTPPushNotificationSender) sendHTTPNotification(ctx context.Context, config *a2a.PushNotificationConfig, payload map[string]any) error {
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal notification payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(timeoutCtx, "POST", config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "a2a-go-push-notification-sender")

	// Add authentication if provided
	if config.Authentication != nil {
		if err := s.addAuthentication(req, config.Authentication); err != nil {
			return fmt.Errorf("failed to add authentication: %w", err)
		}
	}

	// Add token if provided
	if config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+config.Token)
	}

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// addAuthentication adds authentication headers based on the authentication info.
func (s *HTTPPushNotificationSender) addAuthentication(req *http.Request, auth *a2a.AuthenticationInfo) error {
	if auth == nil {
		return nil
	}

	// Handle different authentication schemes
	for _, scheme := range auth.Schemes {
		switch scheme {
		case "basic":
			if auth.Credentials != "" {
				req.Header.Set("Authorization", "Basic "+auth.Credentials)
			}
		case "bearer":
			if auth.Credentials != "" {
				req.Header.Set("Authorization", "Bearer "+auth.Credentials)
			}
		case "api_key":
			if auth.Credentials != "" {
				req.Header.Set("X-API-Key", auth.Credentials)
			}
		default:
			return fmt.Errorf("unsupported authentication scheme: %s", scheme)
		}
	}

	return nil
}

// NoOpPushNotificationSender is a no-op implementation that doesn't send notifications.
// This is useful for testing or when push notifications are disabled.
type NoOpPushNotificationSender struct{}

var _ PushNotificationSender = (*NoOpPushNotificationSender)(nil)

// NewNoOpPushNotificationSender creates a new no-op push notification sender.
func NewNoOpPushNotificationSender() *NoOpPushNotificationSender {
	return &NoOpPushNotificationSender{}
}

// SendNotification does nothing in the no-op implementation.
func (s *NoOpPushNotificationSender) SendNotification(ctx context.Context, taskID string, event any, config *a2a.PushNotificationConfig) error {
	return nil
}

// SendTaskStatusNotification does nothing in the no-op implementation.
func (s *NoOpPushNotificationSender) SendTaskStatusNotification(ctx context.Context, taskID string, status a2a.TaskStatus, config *a2a.PushNotificationConfig) error {
	return nil
}

// SendTaskArtifactNotification does nothing in the no-op implementation.
func (s *NoOpPushNotificationSender) SendTaskArtifactNotification(ctx context.Context, taskID string, artifact *a2a.Artifact, config *a2a.PushNotificationConfig) error {
	return nil
}

// SendTaskCompletedNotification does nothing in the no-op implementation.
func (s *NoOpPushNotificationSender) SendTaskCompletedNotification(ctx context.Context, task *a2a.Task, config *a2a.PushNotificationConfig) error {
	return nil
}

// SendTaskNotification does nothing in the no-op implementation.
func (s *NoOpPushNotificationSender) SendTaskNotification(ctx context.Context, task *a2a.Task) error {
	return nil
}

// Close does nothing in the no-op implementation.
func (s *NoOpPushNotificationSender) Close() error {
	return nil
}
