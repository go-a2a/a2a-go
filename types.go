// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package a2a provides Go types for the Agent-to-Agent (A2A) protocol.
// This file contains type definitions that mirror the Python A2A types,
// converted to idiomatic Go code with proper JSON serialization support.
package a2a

import (
	"fmt"

	"github.com/go-json-experiment/json"
)

// Location represents the location of an API key in HTTP requests.
type Location string

// Valid locations for API keys.
const (
	LocationCookie Location = "cookie"
	LocationHeader Location = "header"
	LocationQuery  Location = "query"
)

// Error codes for A2A protocol.
const (
	ErrorCodeTaskNotFound      = -32001
	ErrorCodeTaskNotCancelable = -32002
	ErrorCodeJSONParse         = -32700
	ErrorCodeInvalidRequest    = -32600
	ErrorCodeMethodNotFound    = -32601
	ErrorCodeInvalidParams     = -32602
	ErrorCodeInternalError     = -32603
)

// A2AError represents an error in the A2A protocol.
type A2AError interface {
	error
	Code() int
	Message() string
}

// A2A represents the base A2A specification.
type A2A interface {
	Validate() error
}

// APIKeySecurityScheme defines an API Key security scheme.
type APIKeySecurityScheme struct {
	Description string   `json:"description"`
	In          Location `json:"in"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
}

// Validate ensures the APIKeySecurityScheme is valid.
func (a APIKeySecurityScheme) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("API key security scheme name cannot be empty")
	}
	if a.Type == "" {
		return fmt.Errorf("API key security scheme type cannot be empty")
	}
	if a.In != LocationCookie && a.In != LocationHeader && a.In != LocationQuery {
		return fmt.Errorf("invalid location for API key: %s", a.In)
	}
	return nil
}

// HTTPAuthSecurityScheme defines an HTTP Authentication security scheme.
type HTTPAuthSecurityScheme struct {
	Description string `json:"description"`
	Scheme      string `json:"scheme"`
	Type        string `json:"type"`
}

// Validate ensures the HTTPAuthSecurityScheme is valid.
func (h HTTPAuthSecurityScheme) Validate() error {
	if h.Scheme == "" {
		return fmt.Errorf("HTTP auth security scheme scheme cannot be empty")
	}
	if h.Type == "" {
		return fmt.Errorf("HTTP auth security scheme type cannot be empty")
	}
	return nil
}

// AgentProvider represents the service provider of an agent.
type AgentProvider struct {
	Organization string `json:"organization"`
	URL          string `json:"url"`
}

// Validate ensures the AgentProvider is valid.
func (a AgentProvider) Validate() error {
	if a.Organization == "" {
		return fmt.Errorf("agent provider organization cannot be empty")
	}
	if a.URL == "" {
		return fmt.Errorf("agent provider URL cannot be empty")
	}
	return nil
}

// AgentSkill describes a unit of capability an agent can perform.
type AgentSkill struct {
	Description string   `json:"description"`
	Examples    []string `json:"examples,omitzero"`
	ID          string   `json:"id"`
	InputMode   string   `json:"input_mode"`
	OutputMode  string   `json:"output_mode"`
	Name        string   `json:"name"`
	Tags        []string `json:"tags,omitzero"`
}

// Validate ensures the AgentSkill is valid.
func (a AgentSkill) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("agent skill ID cannot be empty")
	}
	if a.Name == "" {
		return fmt.Errorf("agent skill name cannot be empty")
	}
	if a.InputMode == "" {
		return fmt.Errorf("agent skill input mode cannot be empty")
	}
	if a.OutputMode == "" {
		return fmt.Errorf("agent skill output mode cannot be empty")
	}
	return nil
}

// File represents a file interface that can be either URI-based or byte-based.
type File interface {
	GetURI() string
	GetBytes() []byte
}

// FileBase provides base functionality for files.
type FileBase struct {
	ContentType string `json:"content_type,omitzero"`
	Name        string `json:"name,omitzero"`
}

// FileWithURI represents a file referenced by URI.
type FileWithURI struct {
	FileBase
	URI string `json:"uri"`
}

// GetURI returns the URI of the file.
func (f FileWithURI) GetURI() string {
	return f.URI
}

// GetBytes returns nil for URI-based files.
func (f FileWithURI) GetBytes() []byte {
	return nil
}

// FileWithBytes represents a file with embedded byte data.
type FileWithBytes struct {
	FileBase
	Bytes []byte `json:"bytes"`
}

// GetURI returns empty string for byte-based files.
func (f FileWithBytes) GetURI() string {
	return ""
}

// GetBytes returns the file bytes.
func (f FileWithBytes) GetBytes() []byte {
	return f.Bytes
}

// FilePart represents a file segment within message parts.
type FilePart struct {
	File File `json:"file"`
}

// A2ARequest represents the union of all supported A2A JSON-RPC request types.
type A2ARequest interface {
	GetMethod() string
	GetParams() any
	GetID() string
}

// GetTaskParams represents parameters for getting a task.
type GetTaskParams struct {
	TaskID string `json:"task_id"`
}

// GetTaskRequest represents a JSON-RPC request for getting a task.
type GetTaskRequest struct {
	Method string        `json:"method"`
	Params GetTaskParams `json:"params"`
	ID     string        `json:"id"`
}

// GetMethod returns the JSON-RPC method name.
func (r GetTaskRequest) GetMethod() string {
	return r.Method
}

// GetParams returns the request parameters.
func (r GetTaskRequest) GetParams() any {
	return r.Params
}

// GetID returns the request ID.
func (r GetTaskRequest) GetID() string {
	return r.ID
}

// CancelTaskParams represents parameters for canceling a task.
type CancelTaskParams struct {
	TaskID string `json:"task_id"`
}

// CancelTaskRequest represents a JSON-RPC request for canceling a task.
type CancelTaskRequest struct {
	Method string           `json:"method"`
	Params CancelTaskParams `json:"params"`
	ID     string           `json:"id"`
}

// GetMethod returns the JSON-RPC method name.
func (r CancelTaskRequest) GetMethod() string {
	return r.Method
}

// GetParams returns the request parameters.
func (r CancelTaskRequest) GetParams() any {
	return r.Params
}

// GetID returns the request ID.
func (r CancelTaskRequest) GetID() string {
	return r.ID
}

// GetTaskPushNotificationConfigParams represents parameters for getting task push notification config.
type GetTaskPushNotificationConfigParams struct {
	TaskID string `json:"task_id"`
}

// GetTaskPushNotificationConfigRequest represents a JSON-RPC request for getting task push notification config.
type GetTaskPushNotificationConfigRequest struct {
	Method string                              `json:"method"`
	Params GetTaskPushNotificationConfigParams `json:"params"`
	ID     string                              `json:"id"`
}

// GetMethod returns the JSON-RPC method name.
func (r GetTaskPushNotificationConfigRequest) GetMethod() string {
	return r.Method
}

// GetParams returns the request parameters.
func (r GetTaskPushNotificationConfigRequest) GetParams() any {
	return r.Params
}

// GetID returns the request ID.
func (r GetTaskPushNotificationConfigRequest) GetID() string {
	return r.ID
}

// DeleteTaskPushNotificationConfigParams represents parameters for deleting task push notification config.
type DeleteTaskPushNotificationConfigParams struct {
	TaskID string `json:"task_id"`
}

// DeleteTaskPushNotificationConfigRequest represents a JSON-RPC request for deleting task push notification config.
type DeleteTaskPushNotificationConfigRequest struct {
	Method string                                 `json:"method"`
	Params DeleteTaskPushNotificationConfigParams `json:"params"`
	ID     string                                 `json:"id"`
}

// GetMethod returns the JSON-RPC method name.
func (r DeleteTaskPushNotificationConfigRequest) GetMethod() string {
	return r.Method
}

// GetParams returns the request parameters.
func (r DeleteTaskPushNotificationConfigRequest) GetParams() any {
	return r.Params
}

// GetID returns the request ID.
func (r DeleteTaskPushNotificationConfigRequest) GetID() string {
	return r.ID
}

// SendMessageParams represents parameters for sending a message.
type SendMessageParams struct {
	Message string `json:"message"`
	TaskID  string `json:"task_id,omitzero"`
}

// SendMessageRequest represents a JSON-RPC request for sending a message.
type SendMessageRequest struct {
	Method string            `json:"method"`
	Params SendMessageParams `json:"params"`
	ID     string            `json:"id"`
}

// GetMethod returns the JSON-RPC method name.
func (r SendMessageRequest) GetMethod() string {
	return r.Method
}

// GetParams returns the request parameters.
func (r SendMessageRequest) GetParams() any {
	return r.Params
}

// GetID returns the request ID.
func (r SendMessageRequest) GetID() string {
	return r.ID
}

// TaskNotFoundError represents an error when a task is not found.
type TaskNotFoundError struct {
	TaskID string
}

// Error returns the error message.
func (e TaskNotFoundError) Error() string {
	return fmt.Sprintf("task not found: %s", e.TaskID)
}

// Code returns the error code.
func (e TaskNotFoundError) Code() int {
	return ErrorCodeTaskNotFound
}

// Message returns the error message.
func (e TaskNotFoundError) Message() string {
	return "A2A specific error indicating the requested task ID was not found"
}

// TaskNotCancelableError represents an error when a task cannot be canceled.
type TaskNotCancelableError struct {
	TaskID string
}

// Error returns the error message.
func (e TaskNotCancelableError) Error() string {
	return fmt.Sprintf("task cannot be canceled: %s", e.TaskID)
}

// Code returns the error code.
func (e TaskNotCancelableError) Code() int {
	return ErrorCodeTaskNotCancelable
}

// Message returns the error message.
func (e TaskNotCancelableError) Message() string {
	return "Task cannot be canceled"
}

// JSONParseError represents a JSON parsing error.
type JSONParseError struct {
	Msg string
}

// Error returns the error message.
func (e JSONParseError) Error() string {
	return fmt.Sprintf("JSON parse error: %s", e.Msg)
}

// Code returns the error code.
func (e JSONParseError) Code() int {
	return ErrorCodeJSONParse
}

// Message returns the error message.
func (e JSONParseError) Message() string {
	return "Parse error"
}

// InvalidRequestError represents an invalid request error.
type InvalidRequestError struct {
	Msg string
}

// Error returns the error message.
func (e InvalidRequestError) Error() string {
	return fmt.Sprintf("invalid request: %s", e.Msg)
}

// Code returns the error code.
func (e InvalidRequestError) Code() int {
	return ErrorCodeInvalidRequest
}

// Message returns the error message.
func (e InvalidRequestError) Message() string {
	return "Invalid Request"
}

// MethodNotFoundError represents a method not found error.
type MethodNotFoundError struct {
	Method string
}

// Error returns the error message.
func (e MethodNotFoundError) Error() string {
	return fmt.Sprintf("method not found: %s", e.Method)
}

// Code returns the error code.
func (e MethodNotFoundError) Code() int {
	return ErrorCodeMethodNotFound
}

// Message returns the error message.
func (e MethodNotFoundError) Message() string {
	return "Method not found"
}

// InvalidParamsError represents an invalid parameters error.
type InvalidParamsError struct {
	Msg string
}

// Error returns the error message.
func (e InvalidParamsError) Error() string {
	return fmt.Sprintf("invalid params: %s", e.Msg)
}

// Code returns the error code.
func (e InvalidParamsError) Code() int {
	return ErrorCodeInvalidParams
}

// Message returns the error message.
func (e InvalidParamsError) Message() string {
	return "Invalid params"
}

// InternalError represents an internal error.
type InternalError struct {
	Msg string
}

// Error returns the error message.
func (e InternalError) Error() string {
	return fmt.Sprintf("internal error: %s", e.Msg)
}

// Code returns the error code.
func (e InternalError) Code() int {
	return ErrorCodeInternalError
}

// Message returns the error message.
func (e InternalError) Message() string {
	return "Internal error"
}

// WireRequest provides a way to handle the union of different request types.
type WireRequest struct {
	SendMessage                      *SendMessageRequest                      `json:"send_message,omitzero"`
	GetTask                          *GetTaskRequest                          `json:"get_task,omitzero"`
	CancelTask                       *CancelTaskRequest                       `json:"cancel_task,omitzero"`
	GetTaskPushNotificationConfig    *GetTaskPushNotificationConfigRequest    `json:"get_task_push_notification_config,omitzero"`
	DeleteTaskPushNotificationConfig *DeleteTaskPushNotificationConfigRequest `json:"delete_task_push_notification_config,omitzero"`
}

// ToRequest converts the union to an A2ARequest interface.
func (u WireRequest) ToRequest() A2ARequest {
	switch {
	case u.SendMessage != nil:
		return u.SendMessage
	case u.GetTask != nil:
		return u.GetTask
	case u.CancelTask != nil:
		return u.CancelTask
	case u.GetTaskPushNotificationConfig != nil:
		return u.GetTaskPushNotificationConfig
	case u.DeleteTaskPushNotificationConfig != nil:
		return u.DeleteTaskPushNotificationConfig
	default:
		return nil
	}
}

// MarshalJSON implements custom JSON marshaling for the union type.
func (u WireRequest) MarshalJSON() ([]byte, error) {
	if req := u.ToRequest(); req != nil {
		return json.Marshal(req)
	}
	return nil, fmt.Errorf("no request type set in union")
}

// UnmarshalJSON implements custom JSON unmarshaling for the union type.
func (u *WireRequest) UnmarshalJSON(data []byte) error {
	var method struct {
		Method string `json:"method"`
	}

	if err := json.Unmarshal(data, &method); err != nil {
		return err
	}

	switch method.Method {
	case "send_message":
		var req SendMessageRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return err
		}
		u.SendMessage = &req

	case "get_task":
		var req GetTaskRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return err
		}
		u.GetTask = &req

	case "cancel_task":
		var req CancelTaskRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return err
		}
		u.CancelTask = &req

	case "get_task_push_notification_config":
		var req GetTaskPushNotificationConfigRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return err
		}
		u.GetTaskPushNotificationConfig = &req

	case "delete_task_push_notification_config":
		var req DeleteTaskPushNotificationConfigRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return err
		}
		u.DeleteTaskPushNotificationConfig = &req

	default:
		return fmt.Errorf("unknown method: %s", method.Method)
	}

	return nil
}
