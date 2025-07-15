// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-a2a/a2a"
	"github.com/google/uuid"
)

// CreateTextMessageObject creates a text message object for A2A communication.
func CreateTextMessageObject(content string, contextID ...string) *a2a.Message {
	message := &a2a.Message{
		Content: content,
	}

	if len(contextID) > 0 && contextID[0] != "" {
		message.ContextID = contextID[0]
	}

	return message
}

// CreateTextMessageObjectWithMetadata creates a text message object with metadata.
func CreateTextMessageObjectWithMetadata(content string, metadata map[string]any, contextID ...string) *a2a.Message {
	message := CreateTextMessageObject(content, contextID...)
	message.Metadata = metadata
	return message
}

// GenerateRequestID generates a unique request ID.
func GenerateRequestID() string {
	return uuid.New().String()
}

// CreateSendMessageConfiguration creates a send message configuration.
func CreateSendMessageConfiguration(opts ...SendMessageConfigOption) *a2a.SendMessageConfiguration {
	config := &a2a.SendMessageConfiguration{}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// SendMessageConfigOption configures a send message configuration.
type SendMessageConfigOption func(*a2a.SendMessageConfiguration)

// WithAcceptedOutputModes sets the accepted output modes.
func WithAcceptedOutputModes(modes []string) SendMessageConfigOption {
	return func(config *a2a.SendMessageConfiguration) {
		config.AcceptedOutputModes = modes
	}
}

// WithPushNotification sets the push notification configuration.
func WithPushNotification(pushNotification *a2a.PushNotificationConfig) SendMessageConfigOption {
	return func(config *a2a.SendMessageConfiguration) {
		config.PushNotification = pushNotification
	}
}

// WithHistoryLength sets the history length.
func WithHistoryLength(length int32) SendMessageConfigOption {
	return func(config *a2a.SendMessageConfiguration) {
		config.HistoryLength = length
	}
}

// WithBlocking sets the blocking mode.
func WithBlocking(blocking bool) SendMessageConfigOption {
	return func(config *a2a.SendMessageConfiguration) {
		config.Blocking = blocking
	}
}

// CreatePushNotificationConfig creates a push notification configuration.
func CreatePushNotificationConfig(id, url, token string) *a2a.PushNotificationConfig {
	return &a2a.PushNotificationConfig{
		ID:    id,
		URL:   url,
		Token: token,
	}
}

// CreatePushNotificationConfigWithAuth creates a push notification configuration with authentication.
func CreatePushNotificationConfigWithAuth(id, url, token string, auth *a2a.AuthenticationInfo) *a2a.PushNotificationConfig {
	config := CreatePushNotificationConfig(id, url, token)
	config.Authentication = auth
	return config
}

// CreateAuthenticationInfo creates authentication information.
func CreateAuthenticationInfo(schemes []string, credentials string) *a2a.AuthenticationInfo {
	return &a2a.AuthenticationInfo{
		Schemes:     schemes,
		Credentials: credentials,
	}
}

// IsRetryableError checks if an error is retryable.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a specific client error type
	if clientErr, ok := err.(interface{ IsRetryable() bool }); ok {
		return clientErr.IsRetryable()
	}

	// Check if it's an HTTP error
	if httpErr, ok := err.(interface{ Code() int }); ok {
		code := httpErr.Code()
		return code >= 500 || code == 408 || code == 429
	}

	return false
}

// IsTemporaryError checks if an error is temporary.
func IsTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it implements the Temporary interface
	if tempErr, ok := err.(interface{ Temporary() bool }); ok {
		return tempErr.Temporary()
	}

	return IsRetryableError(err)
}

// GetHTTPStatusCode extracts the HTTP status code from an error.
func GetHTTPStatusCode(err error) int {
	if err == nil {
		return 0
	}

	if httpErr, ok := err.(interface{ Code() int }); ok {
		return httpErr.Code()
	}

	return 0
}

// ContextWithTimeout creates a context with timeout, preferring the request timeout if set.
func ContextWithTimeout(ctx context.Context, requestTimeout, defaultTimeout time.Duration) (context.Context, context.CancelFunc) {
	timeout := defaultTimeout
	if requestTimeout > 0 {
		timeout = requestTimeout
	}

	return context.WithTimeout(ctx, timeout)
}

// SetDefaultHeaders sets default headers for HTTP requests.
func SetDefaultHeaders(req *http.Request, userAgent string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
}

// SetSSEHeaders sets headers for Server-Sent Events requests.
func SetSSEHeaders(req *http.Request, userAgent string) {
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
}

// MarshalJSON marshals an object to JSON with proper error handling.
func MarshalJSON(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return data, nil
}

// UnmarshalJSON unmarshals JSON data to an object with proper error handling.
func UnmarshalJSON(data []byte, v any) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// ValidateURL validates if a URL is properly formatted.
func ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	return nil
}

// ValidateRequestID validates if a request ID is properly formatted.
func ValidateRequestID(id string) error {
	if id == "" {
		return fmt.Errorf("request ID cannot be empty")
	}

	// Check if it's a valid UUID
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("request ID must be a valid UUID: %w", err)
	}

	return nil
}

// SanitizeContextID sanitizes a context ID by removing invalid characters.
func SanitizeContextID(contextID string) string {
	// Remove any characters that are not alphanumeric, dash, or underscore
	sanitized := ""
	for _, char := range contextID {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_' {
			sanitized += string(char)
		}
	}
	return sanitized
}

// GetErrorMessage extracts a human-readable error message from an error.
func GetErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	// Check if it's a client error with a message method
	if clientErr, ok := err.(interface{ Message() string }); ok {
		return clientErr.Message()
	}

	return err.Error()
}

// CreateDefaultSendMessageConfiguration creates a default send message configuration.
func CreateDefaultSendMessageConfiguration() *a2a.SendMessageConfiguration {
	return &a2a.SendMessageConfiguration{
		AcceptedOutputModes: []string{"text", "json"},
		HistoryLength:       10,
		Blocking:            false,
	}
}

// CreateTaskStatusUpdateEvent creates a task status update event.
func CreateTaskStatusUpdateEvent(taskID string, status a2a.TaskStatus, final bool) *a2a.TaskStatusUpdateEvent {
	return &a2a.TaskStatusUpdateEvent{
		TaskID: taskID,
		Status: status,
		Final:  final,
	}
}

// CreateTaskArtifactUpdateEvent creates a task artifact update event.
func CreateTaskArtifactUpdateEvent(artifact *a2a.Artifact, append bool) *a2a.TaskArtifactUpdateEvent {
	return &a2a.TaskArtifactUpdateEvent{
		Artifact: artifact,
		Append:   append,
	}
}
