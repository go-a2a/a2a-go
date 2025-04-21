// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package a2a provides an open protocol enabling communication and interoperability between opaque agentic applications for Go.
package a2a

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
)

// Version is the current version of the A2A protocol.
const Version = "0.1.0"

// TaskState represents the lifecycle state of a task.
type TaskState string

// Task state constants.
const (
	// TaskStateSubmitted indicates the task has been submitted to the agent.
	TaskStateSubmitted TaskState = "submitted"
	// TaskStateWorking indicates the agent is actively working on the task.
	TaskStateWorking TaskState = "working"
	// TaskStateInputRequired indicates the agent needs additional input from the client.
	TaskStateInputRequired TaskState = "input-required"
	// TaskStateCompleted indicates the task has been successfully completed.
	TaskStateCompleted TaskState = "completed"
	// TaskStateFailed indicates the task has failed.
	TaskStateFailed TaskState = "failed"
	// TaskStateCanceled indicates the task has been canceled.
	TaskStateCanceled TaskState = "canceled"
)

// Role represents the possible roles in a message.
type Role string

const (
	// RoleUser indicates the message is from the client.
	RoleUser Role = "user"
	// RoleAgent indicates the message is from the agent.
	RoleAgent Role = "agent"
)

// PartType represents the possible types of message parts.
type PartType string

const (
	// PartTypeText indicates the part contains text.
	PartTypeText PartType = "text"
	// PartTypeFile indicates the part contains file data.
	PartTypeFile PartType = "file"
	// PartTypeData indicates the part contains structured data.
	PartTypeData PartType = "data"
)

// Part is the interface for all part types.
type Part interface {
	PartType() PartType
}

// TextPart represents a text message part.
type TextPart struct {
	// Text is the textual content.
	Type PartType `json:"type"`

	// Text is the textual content.
	Text string `json:"text"`

	// Metadata contains optional part-specific metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}

var _ Part = (*TextPart)(nil)

// Type implements [Part].
func (*TextPart) PartType() PartType {
	return PartTypeText
}

// FilePart represents a file message part.
type FilePart struct {
	// Text is the textual content.
	Type PartType `json:"type"`

	// File contains the file content details.
	File FileContent `json:"file"`

	// Metadata contains optional part-specific metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}

var _ Part = (*FilePart)(nil)

// Type implements [Part].
func (*FilePart) PartType() PartType {
	return PartTypeFile
}

// FileContent represents the content of a file, either as base64 encoded bytes or a URI.
type FileContent struct {
	Name     string `json:"name,omitzero"`
	MIMEType string `json:"mimeType,omitzero"`
	Bytes    string `json:"bytes,omitzero"`
	URI      string `json:"uri,omitzero"`
}

// CheckContent checks the content of the file to ensure it has either Bytes or URI set, but not both.
func (fc FileContent) CheckContent() error {
	if fc.Bytes == "" && fc.URI == "" {
		return errors.New("either 'Bytes' or 'URI' fields must be present in the file data")
	}
	if fc.Bytes != "" && fc.URI != "" {
		return errors.New("only one of 'Bytes' or 'URI' fields can be present in the file data")
	}
	return nil
}

// DataPart represents a structured data message part.
type DataPart struct {
	// Text is the textual content.
	Type PartType `json:"type"`

	// Data contains the structured JSON data.
	Data map[string]any `json:"data"`

	// Metadata contains optional part-specific metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}

var _ Part = (*DataPart)(nil)

// Type implements [Part].
func (*DataPart) PartType() PartType {
	return PartTypeData
}

// Message represents a communication unit between user and agent.
type Message struct {
	// Role is the sender role ("user" or "agent").
	Role Role `json:"role"`

	// Parts contains the content parts (text, file, data).
	Parts []Part `json:"parts"`

	// Metadata contains optional message-specific metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// UnmarshalJSON implements [json.Unmarshaler].
func (r *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	tmp := &struct {
		*Alias
		Parts []json.RawMessage `json:"parts"`
	}{
		Alias: (*Alias)(r),
	}
	if err := sonic.ConfigFastest.Unmarshal(data, tmp); err != nil {
		return fmt.Errorf("Message: unmarshal data: %w", err)
	}

	r.Parts = make([]Part, len(tmp.Parts))
	for i, part := range tmp.Parts {
		// Try to unmarshal content as TextContent first
		var textPart TextPart
		if err := sonic.ConfigFastest.Unmarshal(part, &textPart); err == nil {
			r.Parts[i] = &textPart
			continue
		}

		// Try to unmarshal content as ImageContent
		var filePart FilePart
		if err := sonic.ConfigFastest.Unmarshal(part, &filePart); err == nil {
			r.Parts[i] = &filePart
			continue
		}

		// Try to unmarshal content as AudioContent
		var dataPart DataPart
		if err := sonic.ConfigFastest.Unmarshal(part, &dataPart); err == nil {
			r.Parts[i] = &dataPart
			continue
		}

		return fmt.Errorf("unknown part type at index: %d", i)
	}

	return nil
}

// TaskStatus represents the current status of a task.
type TaskStatus struct {
	// State is the current lifecycle state.
	State TaskState `json:"state"`
	// Message is optionally associated with this status.
	Message *Message `json:"message,omitempty"`
	// Timestamp is the ISO 8601 timestamp of the status update.
	Timestamp time.Time `json:"timestamp"`
}

// Artifact represents output generated by a task.
type Artifact struct {
	// Name is an optional name for the artifact.
	Name string `json:"name,omitzero"`
	// Description is an optional description of the artifact.
	Description string `json:"description,omitzero"`
	// Parts contains the content parts of the artifact.
	Parts []Part `json:"parts"`
	// Index is the order index, useful for streaming/updates.
	Index int `json:"index,omitempty"`
	// Append indicates if content should append to an artifact at the same index (for streaming).
	Append bool `json:"append,omitzero"`
	// LastChunk indicates the final chunk for this artifact (for streaming).
	LastChunk bool `json:"lastChunk,omitzero"`
	// Metadata contains optional additional artifact metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Task represents a unit of work processed by an agent.
type Task struct {
	// ID is the unique task identifier.
	ID string `json:"id"`
	// SessionID optionally groups related tasks.
	SessionID string `json:"sessionId,omitzero"`
	// Status contains the current state and associated message.
	Status TaskStatus `json:"status"`
	// Artifacts contains outputs generated by the task.
	Artifacts []Artifact `json:"artifacts,omitempty"`
	// History is the list of messages exchanged during the task.
	History []Message `json:"history,omitempty"`
	// Metadata contains additional task metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// TaskEvent represents an event related to a task.
type TaskEvent interface {
	// TaskID returns the task ID that this event is for.
	TaskID() string
}

// TaskStatusUpdateEvent signals a change in task status.
type TaskStatusUpdateEvent struct {
	// ID is the task identifier.
	ID string `json:"id"`
	// Status is the new status object.
	Status TaskStatus `json:"status"`
	// Final indicates if this is the terminal update for the task.
	Final bool `json:"final,omitempty"`
	// Metadata contains optional event metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// TaskID implements [TaskEvent].
func (e *TaskStatusUpdateEvent) TaskID() string {
	return e.ID
}

// TaskArtifactUpdateEvent signals a new or updated artifact.
type TaskArtifactUpdateEvent struct {
	// ID is the task identifier.
	ID string `json:"id"`
	// Artifact contains the artifact data.
	Artifact Artifact `json:"artifact"`
	// Metadata contains optional event metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// GetTaskID implements [TaskEvent].
func (e *TaskArtifactUpdateEvent) TaskID() string {
	return e.ID
}

// AuthenticationInfo represents authentication information.
type AuthenticationInfo struct {
	// Schemes is the list of authentication schemes.
	Schemes []string `json:"schemes"`
	// Credentials is an optional credential string.
	Credentials string `json:"credentials,omitzero"`
}

// PushNotificationConfig represents configuration for push notifications.
type PushNotificationConfig struct {
	// URL is the endpoint URL for the agent to POST notifications to.
	URL string `json:"url"`
	// Token is an optional token for the agent to include.
	Token string `json:"token,omitzero"`
	// Authentication contains auth details the agent needs to call the URL.
	Authentication *AuthenticationInfo `json:"authentication,omitempty"`
}

// TaskIDParams represents parameters for methods that require a task ID.
type TaskIDParams struct {
	// ID is the unique task identifier.
	ID string `json:"id"`
	// Metadata contains optional additional metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// TaskQueryParams represents parameters for querying a task.
type TaskQueryParams struct {
	TaskIDParams

	// HistoryLength optionally limits the number of historical messages to include.
	HistoryLength int `json:"historyLength,omitzero"`
}

// TaskSendParams represents parameters for sending a task.
type TaskSendParams struct {
	// ID is the unique task identifier.
	ID string `json:"id"`
	// SessionID optionally groups related tasks.
	SessionID uuid.UUID `json:"sessionId,omitzero"`
	// Message contains the content to send.
	Message Message `json:"message"`
	// AcceptedOutputModes lists the output modes the client can handle.
	AcceptedOutputModes []string `json:"acceptedOutputModes,omitempty"`
	// PushNotification optionally configures push notifications for this task.
	PushNotification *PushNotificationConfig `json:"pushNotification,omitempty"`
	// HistoryLength optionally limits the number of historical messages to include.
	HistoryLength int `json:"historyLength,omitzero"`
	// Metadata contains optional additional metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// TaskPushNotificationConfig associates a PushNotificationConfig with a task ID.
type TaskPushNotificationConfig struct {
	// ID is the unique task identifier.
	ID string `json:"id"`
	// PushNotificationConfig contains the push notification configuration.
	PushNotificationConfig PushNotificationConfig `json:"pushNotificationConfig"`
}

// AgentProvider represents information about the provider of an agent.
type AgentProvider struct {
	// Organization is the name of the organization providing the agent.
	Organization string `json:"organization"`
	// URL is an optional URL for the organization.
	URL string `json:"url,omitzero"`
}

// AgentCapabilities represents the capabilities supported by an agent.
type AgentCapabilities struct {
	// Streaming indicates if the agent supports task streaming.
	Streaming bool `json:"streaming,omitzero"`
	// PushNotifications indicates if the agent supports push notifications.
	PushNotifications bool `json:"pushNotifications,omitzero"`
	// StateTransitionHistory indicates if the agent supports state transition history.
	StateTransitionHistory bool `json:"stateTransitionHistory,omitzero"`
}

// AgentAuthentication represents the authentication schemes and credentials required by an agent.
type AgentAuthentication struct {
	// Schemes is the list of authentication schemes supported by the agent.
	Schemes []string `json:"schemes"`
	// Credentials is an optional credential string.
	Credentials string `json:"credentials,omitzero"`
}

// AgentSkill represents a specific capability provided by an agent.
type AgentSkill struct {
	// ID is the unique identifier for the skill.
	ID string `json:"id"`
	// Name is the human-readable name of the skill.
	Name string `json:"name"`
	// Description is an optional description of the skill.
	Description string `json:"description,omitzero"`
	// Tags is an optional list of keywords associated with the skill.
	Tags []string `json:"tags,omitempty"`
	// Examples is an optional list of usage examples for the skill.
	Examples []string `json:"examples,omitempty"`
	// InputModes optionally overrides the default input modes for this skill.
	InputModes []string `json:"inputModes,omitempty"`
	// OutputModes optionally overrides the default output modes for this skill.
	OutputModes []string `json:"outputModes,omitempty"`
}

// AgentCard provides metadata about an agent.
type AgentCard struct {
	// Name is the human-readable name of the agent.
	Name string `json:"name"`
	// Description is an optional description of the agent.
	Description string `json:"description,omitzero"`
	// URL is the base URL endpoint for the agent's A2A service.
	URL string `json:"url"`
	// Provider contains optional details about the organization providing the agent.
	Provider *AgentProvider `json:"provider,omitempty"`
	// Version is the agent/API version.
	Version string `json:"version"`
	// DocumentationURL is an optional link to documentation.
	DocumentationURL string `json:"documentationUrl,omitzero"`
	// Capabilities defines the features supported by this agent.
	Capabilities AgentCapabilities `json:"capabilities"`
	// Authentication defines the authentication schemes and credentials needed.
	Authentication *AgentAuthentication `json:"authentication,omitempty"`
	// DefaultInputModes is the list of default supported input types.
	DefaultInputModes []string `json:"defaultInputModes,omitempty"`
	// DefaultOutputModes is the list of default supported output types.
	DefaultOutputModes []string `json:"defaultOutputModes,omitempty"`
	// Skills is the list of specific capabilities provided by the agent.
	Skills []AgentSkill `json:"skills"`
}
