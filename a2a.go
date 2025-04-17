// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package a2a defines the core data structures for the A2A protocol.
package a2a

import (
	"encoding/json"
	"fmt"
	"time"
)

// TaskState represents the possible states of a task.
type TaskState string

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

// TaskStatus represents the status of a task.
type TaskStatus struct {
	// State is the current state of the task.
	State TaskState `json:"state"`
	// Timestamp is the time at which the status was updated.
	Timestamp time.Time `json:"timestamp"`
	// Message is an optional message providing additional details about the status.
	Message string `json:"message,omitempty"`
}

// AgentCapabilities represents the capabilities of an agent.
type AgentCapabilities struct {
	// Streaming indicates whether the agent supports streaming responses.
	Streaming bool `json:"streaming"`
	// PushNotifications indicates whether the agent supports push notifications.
	PushNotifications bool `json:"pushNotifications"`
}

// AgentAuthentication represents the authentication schemes and credentials for an agent.
type AgentAuthentication struct {
	// Schemes lists the authentication schemes supported by the agent.
	Schemes []string `json:"schemes"`
	// Credentials contains the credentials for authentication if provided.
	Credentials string `json:"credentials,omitempty"`
}

// AgentCard represents the metadata and capabilities of an agent.
type AgentCard struct {
	// Name is the name of the agent.
	Name string `json:"name"`
	// Description is a brief description of the agent's purpose.
	Description string `json:"description"`
	// Version is the version of the agent.
	Version string `json:"version"`
	// Contact contains contact information for the agent's maintainer.
	Contact string `json:"contact"`
	// SupportedOutputModes lists the output modes supported by the agent.
	SupportedOutputModes []string `json:"supportedOutputModes"`
	// MetadataEndpoint is the URL for retrieving additional agent metadata.
	MetadataEndpoint string `json:"metadataEndpoint,omitempty"`
	// Capabilities describes the agent's capabilities.
	Capabilities AgentCapabilities `json:"capabilities"`
	// Authentication describes the agent's authentication requirements.
	Authentication *AgentAuthentication `json:"authentication,omitempty"`
}

// Part is the interface for all part types.
type Part interface {
	Type() PartType
}

// TextPart represents a text message part.
type TextPart struct {
	// Text is the content of the text part.
	Text string `json:"text"`
}

// Type returns the type of the part.
func (p *TextPart) Type() PartType {
	return PartTypeText
}

// FilePart represents a file message part.
type FilePart struct {
	// MimeType is the MIME type of the file.
	MimeType string `json:"mimeType"`
	// FileName is the name of the file.
	FileName string `json:"fileName,omitempty"`
	// Data contains the file data as a base64-encoded string.
	Data string `json:"data,omitempty"`
	// URI is a URI pointing to the file data.
	URI string `json:"uri,omitempty"`
}

// Type returns the type of the part.
func (p *FilePart) Type() PartType {
	return PartTypeFile
}

// DataPart represents a structured data message part.
type DataPart struct {
	// Data contains the structured data as a JSON object.
	Data json.RawMessage `json:"data"`
}

// Type returns the type of the part.
func (p *DataPart) Type() PartType {
	return PartTypeData
}

// PartWrapper is a wrapper for serializing and deserializing parts.
type PartWrapper struct {
	Type     PartType        `json:"type"`
	Text     string          `json:"text,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
	MimeType string          `json:"mimeType,omitempty"`
	FileName string          `json:"fileName,omitempty"`
	URI      string          `json:"uri,omitempty"`
}

// MarshalJSON implements the json.Marshaler interface.
func (p PartWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type     PartType        `json:"type"`
		Text     string          `json:"text,omitempty"`
		Data     json.RawMessage `json:"data,omitempty"`
		MimeType string          `json:"mimeType,omitempty"`
		FileName string          `json:"fileName,omitempty"`
		URI      string          `json:"uri,omitempty"`
	}{
		Type:     p.Type,
		Text:     p.Text,
		Data:     p.Data,
		MimeType: p.MimeType,
		FileName: p.FileName,
		URI:      p.URI,
	})
}

// UnmarshalPart converts a PartWrapper to a Part.
func UnmarshalPart(pw *PartWrapper) (Part, error) {
	switch pw.Type {
	case PartTypeText:
		return &TextPart{Text: pw.Text}, nil
	case PartTypeFile:
		return &FilePart{
			MimeType: pw.MimeType,
			FileName: pw.FileName,
			Data:     pw.Text,
			URI:      pw.URI,
		}, nil
	case PartTypeData:
		return &DataPart{Data: pw.Data}, nil
	default:
		return nil, fmt.Errorf("unknown part type: %#v", pw)
	}
}

// MarshalPart converts a Part to a PartWrapper.
func MarshalPart(p Part) (*PartWrapper, error) {
	pw := &PartWrapper{Type: p.Type()}

	switch v := p.(type) {
	case *TextPart:
		pw.Type = PartTypeText
		pw.Text = v.Text
	case *FilePart:
		pw.Type = PartTypeFile
		pw.MimeType = v.MimeType
		pw.FileName = v.FileName
		pw.Text = v.Data
		pw.URI = v.URI
	case *DataPart:
		pw.Type = PartTypeData
		pw.Data = v.Data
	default:
		return nil, fmt.Errorf("unknown part type: %T", p)
	}

	return pw, nil
}

// Message represents a message in a conversation.
type Message struct {
	// Role is the role of the message sender (user or agent).
	Role Role `json:"role"`
	// Parts is the list of parts that make up the message.
	Parts []Part `json:"parts"`
}

// MessageWrapper is a wrapper for serializing and deserializing messages.
type MessageWrapper struct {
	Role  Role           `json:"role"`
	Parts []*PartWrapper `json:"parts"`
}

// MarshalJSON implements the json.Marshaler interface.
func (m *Message) MarshalJSON() ([]byte, error) {
	parts := make([]*PartWrapper, len(m.Parts))
	for i, p := range m.Parts {
		pw, err := MarshalPart(p)
		if err != nil {
			return nil, err
		}
		parts[i] = pw
	}

	return json.Marshal(&MessageWrapper{
		Role:  m.Role,
		Parts: parts,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (m *Message) UnmarshalJSON(data []byte) error {
	var mw MessageWrapper
	if err := json.Unmarshal(data, &mw); err != nil {
		return err
	}

	m.Role = mw.Role
	m.Parts = make([]Part, len(mw.Parts))
	for i, pw := range mw.Parts {
		p, err := UnmarshalPart(pw)
		if err != nil {
			return err
		}
		m.Parts[i] = p
	}

	return nil
}

// Artifact represents an output generated by an agent during a task.
type Artifact struct {
	// ID is the unique identifier for the artifact.
	ID string `json:"id,omitempty"`
	// Parts is the list of parts that make up the artifact.
	Parts []Part `json:"parts"`
	// Name is an optional name for the artifact.
	Name string `json:"name,omitempty"`
	// Index is the position of the artifact in the list of artifacts.
	Index int `json:"index"`
}

// ArtifactWrapper is a wrapper for serializing and deserializing artifacts.
type ArtifactWrapper struct {
	ID    string         `json:"id,omitempty"`
	Parts []*PartWrapper `json:"parts"`
	Name  string         `json:"name,omitempty"`
	Index int            `json:"index"`
}

// MarshalJSON implements the json.Marshaler interface.
func (a *Artifact) MarshalJSON() ([]byte, error) {
	parts := make([]*PartWrapper, len(a.Parts))
	for i, p := range a.Parts {
		pw, err := MarshalPart(p)
		if err != nil {
			return nil, err
		}
		parts[i] = pw
	}

	return json.Marshal(&ArtifactWrapper{
		ID:    a.ID,
		Parts: parts,
		Name:  a.Name,
		Index: a.Index,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (a *Artifact) UnmarshalJSON(data []byte) error {
	var aw ArtifactWrapper
	if err := json.Unmarshal(data, &aw); err != nil {
		return err
	}

	a.ID = aw.ID
	a.Parts = make([]Part, len(aw.Parts))
	for i, pw := range aw.Parts {
		p, err := UnmarshalPart(pw)
		if err != nil {
			return err
		}
		a.Parts[i] = p
	}
	a.Name = aw.Name
	a.Index = aw.Index

	return nil
}

// Task represents an A2A task.
type Task struct {
	// ID is the unique identifier for the task.
	ID string `json:"id"`
	// Status contains the current status of the task.
	Status TaskStatus `json:"status"`
	// Artifacts is the list of artifacts produced by the task.
	Artifacts []Artifact `json:"artifacts,omitempty"`
	// History is the list of messages exchanged during the task.
	History []Message `json:"history,omitempty"`
	// SessionID is an optional identifier for grouping related tasks.
	SessionID string `json:"sessionId,omitempty"`
}

// TaskSendParams represents the parameters for a tasks/send request.
type TaskSendParams struct {
	// ID is the unique identifier for the task.
	ID string `json:"id"`
	// Message is the initial message for the task.
	Message Message `json:"message"`
	// SessionID is an optional identifier for grouping related tasks.
	SessionID string `json:"sessionId,omitempty"`
	// AcceptedOutputModes lists the output modes the client can handle.
	AcceptedOutputModes []string `json:"acceptedOutputModes"`
}

// TaskGetParams represents the parameters for a tasks/get request.
type TaskGetParams struct {
	// ID is the unique identifier for the task.
	ID string `json:"id"`
}

// TaskCancelParams represents the parameters for a tasks/cancel request.
type TaskCancelParams struct {
	// ID is the unique identifier for the task.
	ID string `json:"id"`
}

// PushNotificationConfig represents the configuration for push notifications.
type PushNotificationConfig struct {
	// URL is the webhook URL for push notifications.
	URL string `json:"url"`
	// Authentication contains the authentication configuration for push notifications.
	Authentication *AgentAuthentication `json:"authentication,omitempty"`
}

// SetPushNotificationParams represents the parameters for a tasks/pushNotification/set request.
type SetPushNotificationParams struct {
	// Config is the push notification configuration.
	Config PushNotificationConfig `json:"config"`
}

// TaskStatusUpdateEvent represents a task status update event.
type TaskStatusUpdateEvent struct {
	// ID is the unique identifier for the task.
	ID string `json:"id"`
	// Status contains the updated status of the task.
	Status TaskStatus `json:"status"`
}

// TaskArtifactUpdateEvent represents a task artifact update event.
type TaskArtifactUpdateEvent struct {
	// ID is the unique identifier for the task.
	ID string `json:"id"`
	// Artifact is the new or updated artifact.
	Artifact Artifact `json:"artifact"`
}

// JSONRPCRequest represents a JSON-RPC request.
type JSONRPCRequest struct {
	// JSONRPC is the JSON-RPC version (must be "2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the request identifier.
	ID any `json:"id"`
	// Method is the name of the method to invoke.
	Method string `json:"method"`
	// Params contains the parameters for the method.
	Params json.RawMessage `json:"params"`
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	// Code is the error code.
	Code int `json:"code"`
	// Message is a short description of the error.
	Message string `json:"message"`
	// Data contains additional error information.
	Data any `json:"data,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC response.
type JSONRPCResponse struct {
	// JSONRPC is the JSON-RPC version (must be "2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the request identifier.
	ID any `json:"id"`
	// Result contains the result of the method call if successful.
	Result any `json:"result,omitempty"`
	// Error contains the error information if the method call failed.
	Error *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCEvent represents a JSON-RPC event (used for server-sent events).
type JSONRPCEvent struct {
	// JSONRPC is the JSON-RPC version (must be "2.0").
	JSONRPC string `json:"jsonrpc"`
	// Method is the name of the event.
	Method string `json:"method"`
	// Params contains the parameters for the event.
	Params any `json:"params"`
}

// SSEEvent represents a server-sent event.
type SSEEvent struct {
	// Event is the event type.
	Event string `json:"event"`
	// Data is the event data.
	Data string `json:"data"`
	// ID is an optional event identifier.
	ID string `json:"id,omitempty"`
	// Retry is an optional reconnection time in milliseconds.
	Retry int `json:"retry,omitempty"`
}
