// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"encoding/json"
	"time"

	"github.com/go-a2a/a2a/internal/jsonrpc2"
)

// AgentAuthentication defines authentication requirements for an agent
type AgentAuthentication struct {
	// Schemes is a list of authentication schemes supported by the agent
	Schemes []string `json:"schemes"`
	// Credentials is an optional credential string
	Credentials string `json:"credentials,omitzero"`
}

// AgentCapabilities defines capabilities supported by an agent
type AgentCapabilities struct {
	// Streaming indicates if the agent supports tasks/sendSubscribe
	Streaming bool `json:"streaming,omitzero"`
	// PushNotifications indicates if the agent supports push notifications
	PushNotifications bool `json:"pushNotifications,omitzero"`
	// StateTransitionHistory indicates if the agent supports providing state transition history
	StateTransitionHistory bool `json:"stateTransitionHistory,omitzero"`
}

// AgentProvider defines information about the agent provider
type AgentProvider struct {
	// Organization is the name of the organization providing the agent
	Organization string `json:"organization"`
	// URL is the provider's URL
	URL string `json:"url,omitzero"`
}

// AgentSkill defines a skill provided by an agent
type AgentSkill struct {
	// ID is the unique identifier for the skill
	ID string `json:"id"`
	// Name is the human-readable name of the skill
	Name string `json:"name"`
	// Description is an optional description of the skill
	Description string `json:"description,omitzero"`
	// Tags is an optional list of keywords for the skill
	Tags []string `json:"tags,omitzero"`
	// Examples is an optional list of usage examples
	Examples []string `json:"examples,omitzero"`
	// InputModes is an optional list of input modes specific to this skill
	InputModes []string `json:"inputModes,omitzero"`
	// OutputModes is an optional list of output modes specific to this skill
	OutputModes []string `json:"outputModes,omitzero"`
}

// AgentCard contains metadata about an agent
type AgentCard struct {
	// Name is the human-readable name of the agent
	Name string `json:"name"`
	// Description is an optional description of the agent
	Description string `json:"description,omitzero"`
	// URL is the base URL endpoint for the agent's A2A service
	URL string `json:"url"`
	// Provider is optional information about the organization providing the agent
	Provider *AgentProvider `json:"provider,omitzero"`
	// Version is the agent/API version
	Version string `json:"version"`
	// DocumentationURL is an optional link to documentation
	DocumentationURL string `json:"documentationUrl,omitzero"`
	// Capabilities defines the features supported by the agent
	Capabilities AgentCapabilities `json:"capabilities"`
	// Authentication defines authentication requirements for the agent
	Authentication *AgentAuthentication `json:"authentication,omitzero"`
	// DefaultInputModes is a list of input types supported by default
	DefaultInputModes []string `json:"defaultInputModes"`
	// DefaultOutputModes is a list of output types supported by default
	DefaultOutputModes []string `json:"defaultOutputModes"`
	// Skills is a list of specific capabilities provided by the agent
	Skills []AgentSkill `json:"skills"`
}

// FileContent represents file data
type FileContent struct {
	// Name is the optional filename
	Name string `json:"name,omitzero"`
	// MimeType is the optional MIME type
	MimeType string `json:"mimeType,omitzero"`
	// Bytes is the base64 encoded file content (mutually exclusive with URI)
	Bytes string `json:"bytes,omitzero"`
	// URI is a URI pointing to the file content (mutually exclusive with Bytes)
	URI string `json:"uri,omitzero"`
}

// TextPart represents text content within a Message or Artifact
type TextPart struct {
	// Type is the type of part (always "text")
	Type PartType `json:"type"`
	// Text is the textual content
	Text string `json:"text"`
	// Metadata is optional metadata for this part
	Metadata json.RawMessage `json:"metadata,omitzero"`
}

// FilePart represents file content within a Message or Artifact
type FilePart struct {
	// Type is the type of part (always "file")
	Type PartType `json:"type"`
	// File contains the file details
	File FileContent `json:"file"`
	// Metadata is optional metadata for this part
	Metadata json.RawMessage `json:"metadata,omitzero"`
}

// DataPart represents structured data within a Message or Artifact
type DataPart struct {
	// Type is the type of part (always "data")
	Type PartType `json:"type"`
	// Data is the structured JSON data
	Data json.RawMessage `json:"data"`
	// Metadata is optional metadata for this part
	Metadata json.RawMessage `json:"metadata,omitzero"`
}

// Part is a union type that represents a piece of content
type Part struct {
	// Type determines which part type this is
	Type PartType `json:"type"`
	// Used for text parts
	Text string `json:"text,omitzero"`
	// Used for file parts
	File *FileContent `json:"file,omitzero"`
	// Used for data parts
	Data json.RawMessage `json:"data,omitzero"`
	// Metadata for any part type
	Metadata json.RawMessage `json:"metadata,omitzero"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (p *Part) UnmarshalJSON(data []byte) error {
	// First unmarshal just the type field
	var typeContainer struct {
		Type PartType `json:"type"`
	}
	if err := json.Unmarshal(data, &typeContainer); err != nil {
		return err
	}

	p.Type = typeContainer.Type

	// Then unmarshal the complete structure based on the type
	switch p.Type {
	case PartTypeText:
		var textPart TextPart
		if err := json.Unmarshal(data, &textPart); err != nil {
			return err
		}
		p.Text = textPart.Text
		p.Metadata = textPart.Metadata
	case PartTypeFile:
		var filePart FilePart
		if err := json.Unmarshal(data, &filePart); err != nil {
			return err
		}
		p.File = &filePart.File
		p.Metadata = filePart.Metadata
	case PartTypeData:
		var dataPart DataPart
		if err := json.Unmarshal(data, &dataPart); err != nil {
			return err
		}
		p.Data = dataPart.Data
		p.Metadata = dataPart.Metadata
	}

	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (p Part) MarshalJSON() ([]byte, error) {
	switch p.Type {
	case PartTypeText:
		return json.Marshal(TextPart{
			Type:     p.Type,
			Text:     p.Text,
			Metadata: p.Metadata,
		})
	case PartTypeFile:
		return json.Marshal(FilePart{
			Type:     p.Type,
			File:     *p.File,
			Metadata: p.Metadata,
		})
	case PartTypeData:
		return json.Marshal(DataPart{
			Type:     p.Type,
			Data:     p.Data,
			Metadata: p.Metadata,
		})
	default:
		// If the type is unrecognized, marshal the raw struct
		type RawPart Part
		return json.Marshal(RawPart(p))
	}
}

// Message represents a communication unit between user and agent
type Message struct {
	// Role is the sender role (user or agent)
	Role MessageRole `json:"role"`
	// Parts is the content parts (text, file, data)
	Parts []Part `json:"parts"`
	// Metadata is optional message-specific metadata
	Metadata json.RawMessage `json:"metadata,omitzero"`
}

// Artifact represents output generated by a task
type Artifact struct {
	// Name is the optional artifact name
	Name string `json:"name,omitzero"`
	// Description is the optional artifact description
	Description string `json:"description,omitzero"`
	// Parts is the content parts
	Parts []Part `json:"parts"`
	// Index is the order index (useful for streaming/updates)
	Index int `json:"index"`
	// Append indicates if content should append to artifact at the same index (for streaming)
	Append bool `json:"append,omitzero"`
	// LastChunk indicates the final chunk for this artifact (for streaming)
	LastChunk bool `json:"lastChunk,omitzero"`
	// Metadata is optional artifact metadata
	Metadata json.RawMessage `json:"metadata,omitzero"`
}

// TaskStatus represents the current state and associated message of a task
type TaskStatus struct {
	// State is the current lifecycle state
	State TaskState `json:"state"`
	// Message is the message associated with this status
	Message *Message `json:"message,omitzero"`
	// Timestamp is the ISO 8601 timestamp of the status update
	Timestamp time.Time `json:"timestamp"`
}

// Task represents a unit of work processed by an agent
type Task struct {
	// ID is the unique task identifier
	ID string `json:"id"`
	// SessionID groups related tasks
	SessionID string `json:"sessionId,omitzero"`
	// Status is the current state and associated message
	Status TaskStatus `json:"status"`
	// Artifacts is the outputs generated by the task
	Artifacts []Artifact `json:"artifacts,omitzero"`
	// History is the optional history of messages exchanged for this task
	History []Message `json:"history,omitzero"`
	// Metadata is optional additional task metadata
	Metadata json.RawMessage `json:"metadata,omitzero"`
}

// AuthenticationInfo defines authentication details
type AuthenticationInfo struct {
	// Schemes is a list of authentication schemes
	Schemes []string `json:"schemes"`
	// Credentials is an optional credential string
	Credentials string `json:"credentials,omitzero"`
}

// PushNotificationConfig defines configuration for push notifications
type PushNotificationConfig struct {
	// URL is the endpoint URL for the agent to POST notifications to
	URL string `json:"url"`
	// Token is an optional token for the agent to include
	Token string `json:"token,omitzero"`
	// Authentication is optional auth details the agent needs to call the URL
	Authentication *AuthenticationInfo `json:"authentication,omitzero"`
}

// TaskPushNotificationConfig associates a PushNotificationConfig with a task ID
type TaskPushNotificationConfig struct {
	// ID is the task ID
	ID string `json:"id"`
	// PushNotificationConfig is the push notification configuration
	PushNotificationConfig PushNotificationConfig `json:"pushNotificationConfig"`
}

// TaskIdParams contains parameters for operations that require just a task ID
type TaskIdParams struct {
	// ID is the task ID
	ID string `json:"id"`
	// Metadata is optional additional metadata
	Metadata json.RawMessage `json:"metadata,omitzero"`
}

// TaskQueryParams contains parameters for querying a task
type TaskQueryParams struct {
	// ID is the task ID
	ID string `json:"id"`
	// HistoryLength is the optional number of history messages to include
	HistoryLength int `json:"historyLength,omitzero"`
	// Metadata is optional additional metadata
	Metadata json.RawMessage `json:"metadata,omitzero"`
}

// TaskSendParams contains parameters for sending a task
type TaskSendParams struct {
	// ID is the task ID
	ID string `json:"id"`
	// SessionID is the optional session ID
	SessionID string `json:"sessionId,omitzero"`
	// Message is the message to send
	Message Message `json:"message"`
	// PushNotification is the optional push notification configuration
	PushNotification *PushNotificationConfig `json:"pushNotification,omitzero"`
	// HistoryLength is the optional number of history messages to include
	HistoryLength int `json:"historyLength,omitzero"`
	// Metadata is optional additional metadata
	Metadata json.RawMessage `json:"metadata,omitzero"`
}

// TaskStatusUpdateEvent signals a change in task status
type TaskStatusUpdateEvent struct {
	// ID is the task ID
	ID string `json:"id"`
	// Status is the new status object
	Status TaskStatus `json:"status"`
	// Final indicates if this is the terminal update for the task
	Final bool `json:"final"`
	// Metadata is optional event metadata
	Metadata json.RawMessage `json:"metadata,omitzero"`
}

// TaskArtifactUpdateEvent signals a new or updated artifact
type TaskArtifactUpdateEvent struct {
	// ID is the task ID
	ID string `json:"id"`
	// Artifact is the artifact data
	Artifact Artifact `json:"artifact"`
	// Final indicates if this is the final event (usually false for artifacts)
	Final bool `json:"final,omitzero"`
	// Metadata is optional event metadata
	Metadata json.RawMessage `json:"metadata,omitzero"`
}

// Request types

// SendTaskRequest represents a request to send a task
type SendTaskRequest struct {
	jsonrpc2.Request
	// Params contains the task send parameters
	Params TaskSendParams `json:"params"`
}

// NewSendTaskRequest creates a new SendTaskRequest
func NewSendTaskRequest(id jsonrpc2.ID, params TaskSendParams) *SendTaskRequest {
	return &SendTaskRequest{
		Request: jsonrpc2.Request{
			ID:     id,
			Method: MethodTasksSend,
		},
		Params: params,
	}
}

// GetTaskRequest represents a request to get a task
type GetTaskRequest struct {
	jsonrpc2.Request
	// Params contains the task query parameters
	Params TaskQueryParams `json:"params"`
}

// NewGetTaskRequest creates a new GetTaskRequest
func NewGetTaskRequest(id jsonrpc2.ID, params TaskQueryParams) *GetTaskRequest {
	return &GetTaskRequest{
		Request: jsonrpc2.Request{
			ID:     id,
			Method: MethodTasksGet,
		},
		Params: params,
	}
}

// CancelTaskRequest represents a request to cancel a task
type CancelTaskRequest struct {
	jsonrpc2.Request
	// Params contains the task ID parameters
	Params TaskIdParams `json:"params"`
}

// NewCancelTaskRequest creates a new CancelTaskRequest
func NewCancelTaskRequest(id jsonrpc2.ID, params TaskIdParams) *CancelTaskRequest {
	return &CancelTaskRequest{
		Request: jsonrpc2.Request{
			ID:     id,
			Method: MethodTasksCancel,
		},
		Params: params,
	}
}

// SendTaskStreamingRequest represents a request to send a task and subscribe to updates
type SendTaskStreamingRequest struct {
	jsonrpc2.Request
	// Params contains the task send parameters
	Params TaskSendParams `json:"params"`
}

// NewSendTaskStreamingRequest creates a new SendTaskStreamingRequest
func NewSendTaskStreamingRequest(id jsonrpc2.ID, params TaskSendParams) *SendTaskStreamingRequest {
	return &SendTaskStreamingRequest{
		Request: jsonrpc2.Request{
			ID:     id,
			Method: MethodTasksSendSubscribe,
		},
		Params: params,
	}
}

// SetTaskPushNotificationRequest represents a request to set a push notification
type SetTaskPushNotificationRequest struct {
	jsonrpc2.Request
	// Params contains the task push notification parameters
	Params TaskPushNotificationConfig `json:"params"`
}

// NewSetTaskPushNotificationRequest creates a new SetTaskPushNotificationRequest
func NewSetTaskPushNotificationRequest(id jsonrpc2.ID, params TaskPushNotificationConfig) *SetTaskPushNotificationRequest {
	return &SetTaskPushNotificationRequest{
		Request: jsonrpc2.Request{
			ID:     id,
			Method: MethodTasksPushNotificationSet,
		},
		Params: params,
	}
}

// GetTaskPushNotificationRequest represents a request to get a push notification
type GetTaskPushNotificationRequest struct {
	jsonrpc2.Request
	// Params contains the task ID parameters
	Params TaskIdParams `json:"params"`
}

// NewGetTaskPushNotificationRequest creates a new GetTaskPushNotificationRequest
func NewGetTaskPushNotificationRequest(id jsonrpc2.ID, params TaskIdParams) *GetTaskPushNotificationRequest {
	return &GetTaskPushNotificationRequest{
		Request: jsonrpc2.Request{
			ID:     id,
			Method: MethodTasksPushNotificationGet,
		},
		Params: params,
	}
}

// TaskResubscriptionRequest represents a request to resubscribe to task updates
type TaskResubscriptionRequest struct {
	jsonrpc2.Request
	// Params contains the task query parameters
	Params TaskQueryParams `json:"params"`
}

// NewTaskResubscriptionRequest creates a new TaskResubscriptionRequest
func NewTaskResubscriptionRequest(id jsonrpc2.ID, params TaskQueryParams) *TaskResubscriptionRequest {
	return &TaskResubscriptionRequest{
		Request: jsonrpc2.Request{
			ID:     id,
			Method: MethodTasksResubscribe,
		},
		Params: params,
	}
}

// Response types

// SendTaskResponse represents a response to a send task request
type SendTaskResponse struct {
	jsonrpc2.Response
	// Result contains the task
	Result *Task `json:"result,omitzero"`
}

// NewSendTaskResponse creates a new SendTaskResponse
func NewSendTaskResponse(id jsonrpc2.ID, result *Task) *SendTaskResponse {
	return &SendTaskResponse{
		Response: jsonrpc2.Response{
			ID: id,
		},
		Result: result,
	}
}

// GetTaskResponse represents a response to a get task request
type GetTaskResponse struct {
	jsonrpc2.Response
	// Result contains the task
	Result *Task `json:"result,omitzero"`
}

// NewGetTaskResponse creates a new GetTaskResponse
func NewGetTaskResponse(id jsonrpc2.ID, result *Task) *GetTaskResponse {
	return &GetTaskResponse{
		Response: jsonrpc2.Response{
			ID: id,
		},
		Result: result,
	}
}

// CancelTaskResponse represents a response to a cancel task request
type CancelTaskResponse struct {
	jsonrpc2.Response
	// Result contains the task
	Result *Task `json:"result,omitzero"`
}

// NewCancelTaskResponse creates a new CancelTaskResponse
func NewCancelTaskResponse(id jsonrpc2.ID, result *Task) *CancelTaskResponse {
	return &CancelTaskResponse{
		Response: jsonrpc2.Response{
			ID: id,
		},
		Result: result,
	}
}

// SendTaskStreamingResponse represents a streaming response to a send task request
type SendTaskStreamingResponse struct {
	jsonrpc2.Response
	// Result contains either a TaskStatusUpdateEvent or a TaskArtifactUpdateEvent
	Result json.RawMessage `json:"result,omitzero"`
}

// NewTaskStatusUpdateEventResponse creates a new SendTaskStreamingResponse with a TaskStatusUpdateEvent
func NewTaskStatusUpdateEventResponse(id jsonrpc2.ID, event TaskStatusUpdateEvent) (*SendTaskStreamingResponse, error) {
	resultData, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	return &SendTaskStreamingResponse{
		Response: jsonrpc2.Response{
			ID: id,
		},
		Result: resultData,
	}, nil
}

// NewTaskArtifactUpdateEventResponse creates a new SendTaskStreamingResponse with a TaskArtifactUpdateEvent
func NewTaskArtifactUpdateEventResponse(id jsonrpc2.ID, event TaskArtifactUpdateEvent) (*SendTaskStreamingResponse, error) {
	resultData, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	return &SendTaskStreamingResponse{
		Response: jsonrpc2.Response{
			ID: id,
		},
		Result: resultData,
	}, nil
}

// SetTaskPushNotificationResponse represents a response to a set task push notification request
type SetTaskPushNotificationResponse struct {
	jsonrpc2.Response
	// Result contains the task push notification configuration
	Result *TaskPushNotificationConfig `json:"result,omitzero"`
}

// NewSetTaskPushNotificationResponse creates a new SetTaskPushNotificationResponse
func NewSetTaskPushNotificationResponse(id jsonrpc2.ID, result *TaskPushNotificationConfig) *SetTaskPushNotificationResponse {
	return &SetTaskPushNotificationResponse{
		Response: jsonrpc2.Response{
			ID: id,
		},
		Result: result,
	}
}

// GetTaskPushNotificationResponse represents a response to a get task push notification request
type GetTaskPushNotificationResponse struct {
	jsonrpc2.Response
	// Result contains the task push notification configuration
	Result *TaskPushNotificationConfig `json:"result,omitzero"`
}

// NewGetTaskPushNotificationResponse creates a new GetTaskPushNotificationResponse
func NewGetTaskPushNotificationResponse(id jsonrpc2.ID, result *TaskPushNotificationConfig) *GetTaskPushNotificationResponse {
	return &GetTaskPushNotificationResponse{
		Response: jsonrpc2.Response{
			ID: id,
		},
		Result: result,
	}
}
