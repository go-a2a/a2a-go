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

// Version is the current version of the A2A protocol.
const Version = "0.1.0"

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

// CancelTaskResponse represents a response to a CancelTaskRequest.
type CancelTaskResponse struct {
	Task      Task   `json:"task"`
	RequestID string `json:"requestId"`
}

// TaskStatusUpdateEvent represents an event sent when a task's status changes.
type TaskStatusUpdateEvent struct {
	Task      Task   `json:"task"`
	RequestID string `json:"requestId"`
}

// Client represents an A2A client.
type Client interface {
	SendTask(ctx context.Context, req SendTaskRequest) (*SendTaskResponse, error)
	GetTask(ctx context.Context, req GetTaskRequest) (*GetTaskResponse, error)
	CancelTask(ctx context.Context, req CancelTaskRequest) (*CancelTaskResponse, error)
}

// Server represents an A2A server.
type Server interface {
	HandleSendTask(ctx context.Context, req SendTaskRequest) (*SendTaskResponse, error)
	HandleGetTask(ctx context.Context, req GetTaskRequest) (*GetTaskResponse, error)
	HandleCancelTask(ctx context.Context, req CancelTaskRequest) (*CancelTaskResponse, error)
}

// TaskBuilder helps build a Task with a fluent API.
type TaskBuilder struct {
	task Task
}

// NewTaskBuilder creates a new TaskBuilder.
func NewTaskBuilder() *TaskBuilder {
	return &TaskBuilder{
		task: Task{
			ID:        uuid.New().String(),
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
		artifact.ID = uuid.New().String()
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
			ID:        uuid.New().String(),
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
