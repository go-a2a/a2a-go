// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package task

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-json-experiment/json"

	a2a "github.com/go-a2a/a2a-go"
)

func TestHTTPPushNotificationSender_SendNotification(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and headers
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected application/json content type, got %s", r.Header.Get("Content-Type"))
		}

		// Parse request body
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Verify payload
		if payload["task_id"] != "test-task" {
			t.Errorf("Expected task_id to be 'test-task', got %v", payload["task_id"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create sender
	sender := NewHTTPPushNotificationSender(HTTPPushNotificationSenderConfig{
		Timeout: 5 * time.Second,
	})

	// Create config
	config := &a2a.PushNotificationConfig{
		ID:  "test-config",
		URL: server.URL,
	}

	// Send notification
	ctx := context.Background()
	event := map[string]any{"type": "test_event"}

	err := sender.SendNotification(ctx, "test-task", event, config)
	if err != nil {
		t.Errorf("SendNotification failed: %v", err)
	}
}

func TestHTTPPushNotificationSender_SendTaskStatusNotification(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Verify event structure
		event, ok := payload["event"].(map[string]any)
		if !ok {
			t.Errorf("Expected event to be a map, got %T", payload["event"])
		}

		if event["type"] != "task_status_update" {
			t.Errorf("Expected event type to be 'task_status_update', got %v", event["type"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create sender
	sender := NewHTTPPushNotificationSender(HTTPPushNotificationSenderConfig{})

	// Create config
	config := &a2a.PushNotificationConfig{
		ID:  "test-config",
		URL: server.URL,
	}

	// Send notification
	ctx := context.Background()
	status := a2a.TaskStatus{State: a2a.TaskStateRunning}

	err := sender.SendTaskStatusNotification(ctx, "test-task", status, config)
	if err != nil {
		t.Errorf("SendTaskStatusNotification failed: %v", err)
	}
}

func TestHTTPPushNotificationSender_Authentication(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authentication headers
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got %s", r.Header.Get("Authorization"))
		}

		if r.Header.Get("X-API-Key") != "test-api-key" {
			t.Errorf("Expected X-API-Key header 'test-api-key', got %s", r.Header.Get("X-API-Key"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create sender
	sender := NewHTTPPushNotificationSender(HTTPPushNotificationSenderConfig{})

	// Create config with authentication
	config := &a2a.PushNotificationConfig{
		ID:    "test-config",
		URL:   server.URL,
		Token: "test-token",
		Authentication: &a2a.AuthenticationInfo{
			Schemes:     []string{"api_key"},
			Credentials: "test-api-key",
		},
	}

	// Send notification
	ctx := context.Background()
	event := map[string]any{"type": "test_event"}

	err := sender.SendNotification(ctx, "test-task", event, config)
	if err != nil {
		t.Errorf("SendNotification failed: %v", err)
	}
}

func TestNoOpPushNotificationSender(t *testing.T) {
	sender := NewNoOpPushNotificationSender()

	ctx := context.Background()
	config := &a2a.PushNotificationConfig{
		ID:  "test-config",
		URL: "http://example.com",
	}

	// All methods should succeed and do nothing
	err := sender.SendNotification(ctx, "test-task", map[string]any{}, config)
	if err != nil {
		t.Errorf("NoOp SendNotification should not return error, got: %v", err)
	}

	status := a2a.TaskStatus{State: a2a.TaskStateRunning}
	err = sender.SendTaskStatusNotification(ctx, "test-task", status, config)
	if err != nil {
		t.Errorf("NoOp SendTaskStatusNotification should not return error, got: %v", err)
	}

	artifact := &a2a.Artifact{Parts: []*a2a.PartWrapper{}}
	err = sender.SendTaskArtifactNotification(ctx, "test-task", artifact, config)
	if err != nil {
		t.Errorf("NoOp SendTaskArtifactNotification should not return error, got: %v", err)
	}

	task := &a2a.Task{
		ID:        "test-task",
		ContextID: "test-context",
		Status:    a2a.TaskStatus{State: a2a.TaskStateCompleted},
	}
	err = sender.SendTaskCompletedNotification(ctx, task, config)
	if err != nil {
		t.Errorf("NoOp SendTaskCompletedNotification should not return error, got: %v", err)
	}

	err = sender.Close()
	if err != nil {
		t.Errorf("NoOp Close should not return error, got: %v", err)
	}
}

func TestInMemoryPushNotificationConfigStore(t *testing.T) {
	store := NewInMemoryPushNotificationConfigStore()
	ctx := context.Background()

	// Test initial state
	count := store.GetConfigCount()
	if count != 0 {
		t.Errorf("Expected initial count to be 0, got %d", count)
	}

	// Test saving config
	config := &a2a.TaskPushNotificationConfig{
		Name: "test-config",
		PushNotificationConfig: &a2a.PushNotificationConfig{
			ID:  "test-id",
			URL: "http://example.com",
		},
	}

	err := store.SaveConfig(ctx, "test-task", config)
	if err != nil {
		t.Errorf("SaveConfig failed: %v", err)
	}

	// Test config count
	count = store.GetConfigCount()
	if count != 1 {
		t.Errorf("Expected count to be 1, got %d", count)
	}

	// Test retrieving config
	retrieved, err := store.GetConfig(ctx, "test-task")
	if err != nil {
		t.Errorf("GetConfig failed: %v", err)
	}

	if retrieved.Name != config.Name {
		t.Errorf("Expected name to be %s, got %s", config.Name, retrieved.Name)
	}

	// Test config exists
	exists, err := store.ExistsConfig(ctx, "test-task")
	if err != nil {
		t.Errorf("ExistsConfig failed: %v", err)
	}
	if !exists {
		t.Errorf("Expected config to exist")
	}

	// Test non-existent config
	exists, err = store.ExistsConfig(ctx, "non-existent")
	if err != nil {
		t.Errorf("ExistsConfig failed: %v", err)
	}
	if exists {
		t.Errorf("Expected config to not exist")
	}

	// Test listing configs
	configs, err := store.ListConfigs(ctx)
	if err != nil {
		t.Errorf("ListConfigs failed: %v", err)
	}
	if len(configs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(configs))
	}

	// Test deleting config
	err = store.DeleteConfig(ctx, "test-task")
	if err != nil {
		t.Errorf("DeleteConfig failed: %v", err)
	}

	// Test config no longer exists
	exists, err = store.ExistsConfig(ctx, "test-task")
	if err != nil {
		t.Errorf("ExistsConfig failed: %v", err)
	}
	if exists {
		t.Errorf("Expected config to not exist after deletion")
	}

	// Test getting non-existent config
	_, err = store.GetConfig(ctx, "test-task")
	if err == nil {
		t.Errorf("Expected error when getting non-existent config")
	}

	// Test deleting non-existent config
	err = store.DeleteConfig(ctx, "non-existent")
	if err == nil {
		t.Errorf("Expected error when deleting non-existent config")
	}
}

func TestInMemoryPushNotificationConfigStore_GetConfigByName(t *testing.T) {
	store := NewInMemoryPushNotificationConfigStore()
	ctx := context.Background()

	// Save a config
	config := &a2a.TaskPushNotificationConfig{
		Name: "unique-config-name",
		PushNotificationConfig: &a2a.PushNotificationConfig{
			ID:  "test-id",
			URL: "http://example.com",
		},
	}

	err := store.SaveConfig(ctx, "test-task", config)
	if err != nil {
		t.Errorf("SaveConfig failed: %v", err)
	}

	// Get by name
	retrieved, taskID, err := store.GetConfigByName(ctx, "unique-config-name")
	if err != nil {
		t.Errorf("GetConfigByName failed: %v", err)
	}

	if retrieved == nil {
		t.Errorf("Expected config to be found")
	}

	if taskID != "test-task" {
		t.Errorf("Expected task ID to be 'test-task', got %s", taskID)
	}

	// Get non-existent config by name
	retrieved, taskID, err = store.GetConfigByName(ctx, "non-existent-name")
	if err != nil {
		t.Errorf("GetConfigByName failed: %v", err)
	}

	if retrieved != nil {
		t.Errorf("Expected config to be nil")
	}

	if taskID != "" {
		t.Errorf("Expected task ID to be empty, got %s", taskID)
	}
}

func TestInMemoryPushNotificationConfigStore_UpdateConfig(t *testing.T) {
	store := NewInMemoryPushNotificationConfigStore()
	ctx := context.Background()

	// Save initial config
	config := &a2a.TaskPushNotificationConfig{
		Name: "initial-config",
		PushNotificationConfig: &a2a.PushNotificationConfig{
			ID:  "test-id",
			URL: "http://example.com",
		},
	}

	err := store.SaveConfig(ctx, "test-task", config)
	if err != nil {
		t.Errorf("SaveConfig failed: %v", err)
	}

	// Update config
	updatedConfig := &a2a.TaskPushNotificationConfig{
		Name: "updated-config",
		PushNotificationConfig: &a2a.PushNotificationConfig{
			ID:  "test-id",
			URL: "http://updated.com",
		},
	}

	err = store.UpdateConfig(ctx, "test-task", updatedConfig)
	if err != nil {
		t.Errorf("UpdateConfig failed: %v", err)
	}

	// Verify update
	retrieved, err := store.GetConfig(ctx, "test-task")
	if err != nil {
		t.Errorf("GetConfig failed: %v", err)
	}

	if retrieved.Name != "updated-config" {
		t.Errorf("Expected name to be 'updated-config', got %s", retrieved.Name)
	}

	if retrieved.PushNotificationConfig.URL != "http://updated.com" {
		t.Errorf("Expected URL to be 'http://updated.com', got %s", retrieved.PushNotificationConfig.URL)
	}

	// Test updating non-existent config
	err = store.UpdateConfig(ctx, "non-existent", updatedConfig)
	if err == nil {
		t.Errorf("Expected error when updating non-existent config")
	}
}
