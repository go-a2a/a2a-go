// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package a2a provides an open protocol enabling communication and interoperability between opaque agentic applications for Go.
package a2a

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func init() {
	uuid.EnableRandPool()
}

// Version is the current version of the A2A protocol.
const Version = "0.1.0"

// Request is an interface implemented by all request types in the A2A protocol.
type Request interface {
	request()

	// Validate validates the request and returns an error if it is invalid.
	Validate() error
}

// TaskState represents the state of a Task.
type TaskState string

const (
	// TaskStateSubmitted indicates the task has been submitted.
	TaskStateSubmitted TaskState = "submitted"

	// TaskStateWorking indicates the task is being worked on.
	TaskStateWorking TaskState = "working"

	// TaskStateCompleted indicates the task has been completed.
	TaskStateCompleted TaskState = "completed"

	// TaskStateCanceled indicates the task has been canceled.
	TaskStateCanceled TaskState = "canceled"

	// TaskStateFailed indicates the task has failed.
	TaskStateFailed TaskState = "failed"
)

// AgentCard represents metadata about an agent, including its capabilities.
type AgentCard struct {
	Name         string       `json:"name"`
	URL          string       `json:"url"`
	Version      string       `json:"version"`
	Description  string       `json:"description,omitempty"`
	Vendor       string       `json:"vendor,omitempty"`
	Capabilities []Capability `json:"capabilities,omitempty"`
}

// Capability represents a specific capability that an agent has.
type Capability struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Models      []string `json:"models,omitempty"`
}

// Part represents a part of a message or artifact. It can be text, a file, or data.
type Part struct {
	Type        string         `json:"type"`
	Text        string         `json:"text,omitempty"`
	DataMIME    string         `json:"dataMime,omitempty"`
	Data        string         `json:"data,omitempty"`
	FileName    string         `json:"fileName,omitempty"`
	FileContent string         `json:"fileContent,omitempty"`
	Index       map[string]any `json:"index,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// Message represents a message in a task, which can be from a user or agent.
type Message struct {
	Role      string    `json:"role"`
	Parts     []Part    `json:"parts"`
	CreatedAt time.Time `json:"createdAt"`
}

// Artifact represents an output generated during a task.
type Artifact struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Parts      []Part         `json:"parts"`
	Index      map[string]any `json:"index,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	CreatedAt  time.Time      `json:"createdAt"`
	ModifiedAt time.Time      `json:"modifiedAt,omitzero"`
}

// Task represents a unit of work in the A2A protocol.
type Task struct {
	ID         string         `json:"id"`
	State      TaskState      `json:"state"`
	AgentCard  AgentCard      `json:"agentCard"`
	Client     string         `json:"client,omitempty"`
	Messages   []Message      `json:"messages"`
	Artifacts  []Artifact     `json:"artifacts,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	CreatedAt  time.Time      `json:"createdAt"`
	ModifiedAt time.Time      `json:"modifiedAt,omitzero"`
}

// SendTaskRequest represents a request to send a task to an A2A server.
type SendTaskRequest struct {
	Task Task `json:"task"`
}

var _ Request = (*SendTaskRequest)(nil)

func (SendTaskRequest) request() {}

// Validate validates the SendTaskRequest.
func (r SendTaskRequest) Validate() error {
	if r.Task.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if len(r.Task.Messages) == 0 {
		return fmt.Errorf("task must have at least one message")
	}
	return nil
}

// SendTaskResponse represents a response to a SendTaskRequest.
type SendTaskResponse struct {
	Task      Task   `json:"task"`
	RequestID string `json:"requestId"`
}

// GetTaskRequest represents a request to get a task from an A2A server.
type GetTaskRequest struct {
	TaskID    string `json:"taskId"`
	RequestID string `json:"requestId,omitempty"`
}

var _ Request = (*GetTaskRequest)(nil)

func (GetTaskRequest) request() {}

// Validate validates the GetTaskRequest.
func (r GetTaskRequest) Validate() error {
	if r.TaskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	return nil
}

// GetTaskResponse represents a response to a GetTaskRequest.
type GetTaskResponse struct {
	Task      Task   `json:"task"`
	RequestID string `json:"requestId"`
}

// CancelTaskRequest represents a request to cancel a task on an A2A server.
type CancelTaskRequest struct {
	TaskID    string `json:"taskId"`
	RequestID string `json:"requestId,omitempty"`
}

var _ Request = (*CancelTaskRequest)(nil)

func (CancelTaskRequest) request() {}

// Validate validates the CancelTaskRequest.
func (r CancelTaskRequest) Validate() error {
	if r.TaskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	return nil
}

// CancelTaskResponse represents a response to a CancelTaskRequest.
type CancelTaskResponse struct {
	Task      Task   `json:"task"`
	RequestID string `json:"requestId"`
}

// TaskPushNotificationConfig represents the configuration for push notifications for a task.
type TaskPushNotificationConfig struct {
	TaskID                 string                 `json:"taskId"`
	PushNotificationConfig PushNotificationConfig `json:"pushNotificationConfig"`
}

// PushNotificationConfig represents the configuration for push notifications.
type PushNotificationConfig struct {
	URL            string              `json:"url"`
	Token          string              `json:"token,omitempty"`
	Authentication *AuthenticationInfo `json:"authentication,omitempty"`
}

// AuthenticationInfo represents authentication information for push notifications.
type AuthenticationInfo struct {
	Schemes     []string       `json:"schemes"`
	Credentials string         `json:"credentials,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// SetTaskPushNotificationRequest represents a request to set push notification configuration for a task.
type SetTaskPushNotificationRequest struct {
	TaskID                 string                 `json:"taskId"`
	PushNotificationConfig PushNotificationConfig `json:"pushNotificationConfig"`
}

var _ Request = (*SetTaskPushNotificationRequest)(nil)

func (SetTaskPushNotificationRequest) request() {}

// Validate validates the SetTaskPushNotificationRequest.
func (r SetTaskPushNotificationRequest) Validate() error {
	if r.TaskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if r.PushNotificationConfig.URL == "" {
		return fmt.Errorf("push notification URL cannot be empty")
	}
	return nil
}

// GetTaskPushNotificationRequest represents a request to get push notification configuration for a task.
type GetTaskPushNotificationRequest struct {
	TaskID    string         `json:"taskId"`
	RequestID string         `json:"requestId,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

var _ Request = (*GetTaskPushNotificationRequest)(nil)

func (GetTaskPushNotificationRequest) request() {}

// Validate validates the GetTaskPushNotificationRequest.
func (r GetTaskPushNotificationRequest) Validate() error {
	if r.TaskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	return nil
}

// TaskResubscriptionRequest represents a request to resubscribe to a task's updates.
type TaskResubscriptionRequest struct {
	TaskID        string         `json:"taskId"`
	HistoryLength int            `json:"historyLength,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

var _ Request = (*TaskResubscriptionRequest)(nil)

func (TaskResubscriptionRequest) request() {}

// Validate validates the TaskResubscriptionRequest.
func (r TaskResubscriptionRequest) Validate() error {
	if r.TaskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if r.HistoryLength < 0 {
		return fmt.Errorf("history length cannot be negative")
	}
	return nil
}

// SetTaskPushNotificationResponse represents a response to a SetTaskPushNotificationRequest.
type SetTaskPushNotificationResponse struct {
	Config    TaskPushNotificationConfig `json:"config"`
	RequestID string                     `json:"requestId"`
}

// GetTaskPushNotificationResponse represents a response to a GetTaskPushNotificationRequest.
type GetTaskPushNotificationResponse struct {
	Config    TaskPushNotificationConfig `json:"config"`
	RequestID string                     `json:"requestId"`
}

// TaskStatusUpdateEvent represents an event sent when a task's status changes.
type TaskStatusUpdateEvent struct {
	Task      Task   `json:"task"`
	RequestID string `json:"requestId"`
}

// GetTaskID returns the task ID for a TaskStatusUpdateEvent.
func (e TaskStatusUpdateEvent) GetTaskID() string {
	return e.Task.ID
}

// Client represents an A2A client.
type Client interface {
	SendTask(ctx context.Context, req Request) (*SendTaskResponse, error)
	GetTask(ctx context.Context, req Request) (*GetTaskResponse, error)
	CancelTask(ctx context.Context, req Request) (*CancelTaskResponse, error)
	SetTaskPushNotification(ctx context.Context, req Request) (*SetTaskPushNotificationResponse, error)
	GetTaskPushNotification(ctx context.Context, req Request) (*GetTaskPushNotificationResponse, error)
	ResubscribeTask(ctx context.Context, req Request) (*TaskStatusUpdateEvent, error)
}

// Server represents an A2A server.
type Server interface {
	HandleSendTask(ctx context.Context, req Request) (*SendTaskResponse, error)
	HandleGetTask(ctx context.Context, req Request) (*GetTaskResponse, error)
	HandleCancelTask(ctx context.Context, req Request) (*CancelTaskResponse, error)
	HandleSetTaskPushNotification(ctx context.Context, req Request) (*SetTaskPushNotificationResponse, error)
	HandleGetTaskPushNotification(ctx context.Context, req Request) (*GetTaskPushNotificationResponse, error)
	HandleTaskResubscription(ctx context.Context, req Request) (*TaskStatusUpdateEvent, error)
}

// TaskBuilder helps build a Task with a fluent API.
type TaskBuilder struct {
	task Task
}

// NewTaskBuilder creates a new TaskBuilder.
func NewTaskBuilder() *TaskBuilder {
	return &TaskBuilder{
		task: Task{
			ID:        uuid.NewString(),
			State:     TaskStateSubmitted,
			Messages:  make([]Message, 0),
			Artifacts: make([]Artifact, 0),
			Metadata:  make(map[string]any),
			CreatedAt: time.Now().UTC(),
		},
	}
}

// WithAgentCard sets the AgentCard for the Task.
func (b *TaskBuilder) WithAgentCard(agentCard AgentCard) *TaskBuilder {
	b.task.AgentCard = agentCard

	return b
}

// WithClient sets the client for the Task.
func (b *TaskBuilder) WithClient(client string) *TaskBuilder {
	b.task.Client = client

	return b
}

// AddMessage adds a Message to the Task.
func (b *TaskBuilder) AddMessage(message Message) *TaskBuilder {
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now().UTC()
	}
	b.task.Messages = append(b.task.Messages, message)
	b.task.ModifiedAt = time.Now().UTC()

	return b
}

// AddArtifact adds an Artifact to the Task.
func (b *TaskBuilder) AddArtifact(artifact Artifact) *TaskBuilder {
	if artifact.ID == "" {
		artifact.ID = uuid.NewString()
	}
	if artifact.CreatedAt.IsZero() {
		artifact.CreatedAt = time.Now().UTC()
	}
	b.task.Artifacts = append(b.task.Artifacts, artifact)
	b.task.ModifiedAt = time.Now().UTC()

	return b
}

// WithMetadata sets metadata for the Task.
func (b *TaskBuilder) WithMetadata(key string, value any) *TaskBuilder {
	b.task.Metadata[key] = value
	b.task.ModifiedAt = time.Now().UTC()

	return b
}

// WithState sets the state for the Task.
func (b *TaskBuilder) WithState(state TaskState) *TaskBuilder {
	b.task.State = state
	b.task.ModifiedAt = time.Now().UTC()

	return b
}

// Build builds the Task.
func (b *TaskBuilder) Build() Task {
	return b.task
}

// MessageBuilder helps build a Message with a fluent API.
type MessageBuilder struct {
	message Message
}

// NewMessageBuilder creates a new MessageBuilder.
func NewMessageBuilder(role string) *MessageBuilder {
	return &MessageBuilder{
		message: Message{
			Role:      role,
			Parts:     make([]Part, 0),
			CreatedAt: time.Now().UTC(),
		},
	}
}

// AddTextPart adds a text part to the Message.
func (b *MessageBuilder) AddTextPart(text string) *MessageBuilder {
	b.message.Parts = append(b.message.Parts, Part{
		Type: "text",
		Text: text,
	})

	return b
}

// AddFilePart adds a file part to the Message.
func (b *MessageBuilder) AddFilePart(fileName, fileContent string) *MessageBuilder {
	b.message.Parts = append(b.message.Parts, Part{
		Type:        "file",
		FileName:    fileName,
		FileContent: fileContent,
	})

	return b
}

// AddDataPart adds a data part to the Message.
func (b *MessageBuilder) AddDataPart(dataMIME, data string) *MessageBuilder {
	b.message.Parts = append(b.message.Parts, Part{
		Type:     "data",
		DataMIME: dataMIME,
		Data:     data,
	})

	return b
}

// Build builds the Message.
func (b *MessageBuilder) Build() Message {
	return b.message
}

// ArtifactBuilder helps build an Artifact with a fluent API.
type ArtifactBuilder struct {
	artifact Artifact
}

// NewArtifactBuilder creates a new ArtifactBuilder.
func NewArtifactBuilder(artifactType string) *ArtifactBuilder {
	now := time.Now().UTC()
	return &ArtifactBuilder{
		artifact: Artifact{
			ID:        uuid.NewString(),
			Type:      artifactType,
			Parts:     make([]Part, 0),
			Index:     make(map[string]any),
			Metadata:  make(map[string]any),
			CreatedAt: now,
		},
	}
}

// AddTextPart adds a text part to the Artifact.
func (b *ArtifactBuilder) AddTextPart(text string) *ArtifactBuilder {
	b.artifact.Parts = append(b.artifact.Parts, Part{
		Type: "text",
		Text: text,
	})
	b.artifact.ModifiedAt = time.Now().UTC()
	return b
}

// AddFilePart adds a file part to the Artifact.
func (b *ArtifactBuilder) AddFilePart(fileName, fileContent string) *ArtifactBuilder {
	b.artifact.Parts = append(b.artifact.Parts, Part{
		Type:        "file",
		FileName:    fileName,
		FileContent: fileContent,
	})
	b.artifact.ModifiedAt = time.Now().UTC()

	return b
}

// AddDataPart adds a data part to the Artifact.
func (b *ArtifactBuilder) AddDataPart(dataMIME, data string) *ArtifactBuilder {
	b.artifact.Parts = append(b.artifact.Parts, Part{
		Type:     "data",
		DataMIME: dataMIME,
		Data:     data,
	})
	b.artifact.ModifiedAt = time.Now().UTC()

	return b
}

// WithIndex adds an index to the Artifact.
func (b *ArtifactBuilder) WithIndex(key string, value any) *ArtifactBuilder {
	b.artifact.Index[key] = value
	b.artifact.ModifiedAt = time.Now().UTC()

	return b
}

// WithMetadata adds metadata to the Artifact.
func (b *ArtifactBuilder) WithMetadata(key string, value any) *ArtifactBuilder {
	b.artifact.Metadata[key] = value
	b.artifact.ModifiedAt = time.Now().UTC()

	return b
}

// Build builds the Artifact.
func (b *ArtifactBuilder) Build() Artifact {
	return b.artifact
}

// TaskOptions configures a task to be executed.
type TaskOptions struct {
	Timeout time.Duration
}

// RunTask executes a task on an A2A server.
func RunTask(ctx context.Context, client Client, task Task, opts TaskOptions) (*Task, error) {
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	req := SendTaskRequest{
		Task: task,
	}

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid task request: %w", err)
	}

	resp, err := client.SendTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("send task: %w", err)
	}

	task = resp.Task

	for task.State != TaskStateCompleted && task.State != TaskStateCanceled && task.State != TaskStateFailed {
		getReq := GetTaskRequest{
			TaskID:    task.ID,
			RequestID: resp.RequestID,
		}

		if err := getReq.Validate(); err != nil {
			return nil, fmt.Errorf("invalid get task request: %w", err)
		}

		getResp, err := client.GetTask(ctx, getReq)
		if err != nil {
			return nil, fmt.Errorf("get task: %w", err)
		}
		task = getResp.Task

		// Simple backoff
		time.Sleep(1 * time.Second)
	}

	return &task, nil
}
