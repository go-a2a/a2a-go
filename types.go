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
	BearerFormat string `json:"bearer_format,omitzero"`
	Description  string `json:"description"`
	Scheme       string `json:"scheme"`
	Type         string `json:"type"`
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
	InputModes  []string `json:"input_modes,omitzero"`
	Name        string   `json:"name"`
	OutputModes []string `json:"output_modes,omitzero"`
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
	if len(a.InputModes) == 0 {
		return fmt.Errorf("agent skill input mode cannot be empty")
	}
	if len(a.OutputModes) == 0 {
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

// TaskState represents the state of a task in the A2A protocol.
type TaskState string

// Task state constants.
const (
	TaskStateSubmitted TaskState = "submitted"
	TaskStateRunning   TaskState = "running"
	TaskStateCompleted TaskState = "completed"
	TaskStateFailed    TaskState = "failed"
	TaskStateCanceled  TaskState = "canceled"
)

// TaskStatus represents the status of a task, wrapping the task state.
type TaskStatus struct {
	State TaskState `json:"state"`
}

// Validate ensures the TaskStatus is valid.
func (ts TaskStatus) Validate() error {
	switch ts.State {
	case TaskStateSubmitted, TaskStateRunning, TaskStateCompleted, TaskStateFailed, TaskStateCanceled:
		return nil
	default:
		return fmt.Errorf("invalid task state: %s", ts.State)
	}
}

// Message represents a message in the A2A protocol.
type Message struct {
	ContextID string         `json:"contextId,omitzero"`
	Content   string         `json:"content"`
	Metadata  map[string]any `json:"metadata,omitzero"`
}

// Validate ensures the Message is valid.
func (m Message) Validate() error {
	if m.Content == "" {
		return fmt.Errorf("message content cannot be empty")
	}
	return nil
}

// Task represents a task in the A2A protocol.
type Task struct {
	ID        string      `json:"id"`
	ContextID string      `json:"contextId"`
	Status    TaskStatus  `json:"status"`
	History   []*Message  `json:"history,omitzero"`
	Artifacts []*Artifact `json:"artifacts,omitzero"`
}

// Validate ensures the Task is valid.
func (t Task) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if t.ContextID == "" {
		return fmt.Errorf("task context ID cannot be empty")
	}
	if err := t.Status.Validate(); err != nil {
		return fmt.Errorf("task status is invalid: %w", err)
	}
	for i, msg := range t.History {
		if msg == nil {
			return fmt.Errorf("task history message at index %d cannot be nil", i)
		}
		if err := msg.Validate(); err != nil {
			return fmt.Errorf("task history message at index %d is invalid: %w", i, err)
		}
	}
	for i, artifact := range t.Artifacts {
		if artifact == nil {
			return fmt.Errorf("task artifact at index %d cannot be nil", i)
		}
		if err := artifact.Validate(); err != nil {
			return fmt.Errorf("task artifact at index %d is invalid: %w", i, err)
		}
	}
	return nil
}

// MessageSendParams represents parameters for sending a message to create a task.
type MessageSendParams struct {
	Message *Message `json:"message"`
}

// Validate ensures the MessageSendParams is valid.
func (msp MessageSendParams) Validate() error {
	if msp.Message == nil {
		return fmt.Errorf("message send params message cannot be nil")
	}
	return msp.Message.Validate()
}

// TaskArtifactUpdateEvent represents an event for updating a task's artifacts.
type TaskArtifactUpdateEvent struct {
	Artifact *Artifact `json:"artifact"`
	Append   bool      `json:"append"`
}

// Validate ensures the TaskArtifactUpdateEvent is valid.
func (taue TaskArtifactUpdateEvent) Validate() error {
	if taue.Artifact == nil {
		return fmt.Errorf("task artifact update event artifact cannot be nil")
	}
	return taue.Artifact.Validate()
}

// AuthenticationInfo represents authentication information for push notifications.
type AuthenticationInfo struct {
	Schemes     []string `json:"schemes,omitzero"`
	Credentials string   `json:"credentials,omitzero"`
}

// Validate ensures the AuthenticationInfo is valid.
func (ai AuthenticationInfo) Validate() error {
	return nil
}

// PushNotificationConfig represents push notification configuration.
type PushNotificationConfig struct {
	ID             string              `json:"id"`
	URL            string              `json:"url"`
	Token          string              `json:"token"`
	Authentication *AuthenticationInfo `json:"authentication,omitzero"`
}

// Validate ensures the PushNotificationConfig is valid.
func (pnc PushNotificationConfig) Validate() error {
	if pnc.ID == "" {
		return fmt.Errorf("push notification config ID cannot be empty")
	}
	if pnc.URL == "" {
		return fmt.Errorf("push notification config URL cannot be empty")
	}
	if pnc.Authentication != nil {
		if err := pnc.Authentication.Validate(); err != nil {
			return fmt.Errorf("push notification config authentication is invalid: %w", err)
		}
	}
	return nil
}

// TaskPushNotificationConfig represents task-specific push notification configuration.
type TaskPushNotificationConfig struct {
	Name                   string                  `json:"name"`
	PushNotificationConfig *PushNotificationConfig `json:"push_notification_config"`
}

// Validate ensures the TaskPushNotificationConfig is valid.
func (tpnc TaskPushNotificationConfig) Validate() error {
	if tpnc.Name == "" {
		return fmt.Errorf("task push notification config name cannot be empty")
	}
	if tpnc.PushNotificationConfig == nil {
		return fmt.Errorf("task push notification config push notification config cannot be nil")
	}
	return tpnc.PushNotificationConfig.Validate()
}

// TaskStatusUpdateEvent represents a task status update event.
type TaskStatusUpdateEvent struct {
	TaskID    string         `json:"task_id"`
	ContextID string         `json:"context_id,omitzero"`
	Status    TaskStatus     `json:"status"`
	Final     bool           `json:"final"`
	Metadata  map[string]any `json:"metadata,omitzero"`
}

// Validate ensures the TaskStatusUpdateEvent is valid.
func (tsue TaskStatusUpdateEvent) Validate() error {
	if tsue.TaskID == "" {
		return fmt.Errorf("task status update event task ID cannot be empty")
	}
	return tsue.Status.Validate()
}

// SendMessageConfiguration represents configuration for sending messages.
type SendMessageConfiguration struct {
	AcceptedOutputModes []string                `json:"accepted_output_modes,omitzero"`
	PushNotification    *PushNotificationConfig `json:"push_notification,omitzero"`
	HistoryLength       int32                   `json:"history_length,omitzero"`
	Blocking            bool                    `json:"blocking,omitzero"`
}

// Validate ensures the SendMessageConfiguration is valid.
func (smc SendMessageConfiguration) Validate() error {
	if smc.PushNotification != nil {
		if err := smc.PushNotification.Validate(); err != nil {
			return fmt.Errorf("send message configuration push notification is invalid: %w", err)
		}
	}
	return nil
}

// AgentExtension represents an extension supported by an agent.
type AgentExtension struct {
	URI         string `json:"uri"`
	Description string `json:"description,omitzero"`
	Required    bool   `json:"required,omitzero"`
}

// AgentCapabilities represents the capabilities of an agent.
type AgentCapabilities struct {
	Streaming         bool              `json:"streaming,omitzero"`
	PushNotifications bool              `json:"push_notifications,omitzero"`
	Extensions        []*AgentExtension `json:"extensions,omitzero"`
}

// Validate ensures the AgentCapabilities is valid.
func (ac AgentCapabilities) Validate() error {
	return nil
}

// AgentInterface represents an additional interface supported by an agent.
type AgentInterface struct {
	URL       string `json:"url"`
	Transport string `json:"transport"`
}

// Security represents security requirements for contacting the agent.
type Security struct {
	SecuritySchemes map[string][]string `json:"security_schemes,omitzero"`
}

// AgentCard represents an agent card with comprehensive information.
type AgentCard struct {
	ProtocolVersion                   string                     `json:"protocol_version,omitzero"`
	Name                              string                     `json:"name"`
	Description                       string                     `json:"description"`
	URL                               string                     `json:"url"`
	PreferredTransport                string                     `json:"preferred_transport,omitzero"`
	AdditionalInterfaces              []*AgentInterface          `json:"additional_interfaces,omitzero"`
	Provider                          *AgentProvider             `json:"provider"`
	Version                           string                     `json:"version"`
	DocumentationURL                  string                     `json:"documentation_url,omitzero"`
	Capabilities                      *AgentCapabilities         `json:"capabilities,omitzero"`
	SecuritySchemes                   map[string]*SecurityScheme `json:"security_schemes,omitzero"`
	Security                          []*Security                `json:"security,omitzero"`
	DefaultInputModes                 []string                   `json:"default_input_modes,omitzero"`
	DefaultOutputModes                []string                   `json:"default_output_modes,omitzero"`
	Skills                            []*AgentSkill              `json:"skills,omitzero"`
	SupportsAuthenticatedExtendedCard bool                       `json:"supports_authenticated_extended_card"`
}

// Validate ensures the AgentCard is valid.
func (ac AgentCard) Validate() error {
	if ac.Name == "" {
		return fmt.Errorf("agent card name cannot be empty")
	}
	if ac.Provider == nil {
		return fmt.Errorf("agent card provider cannot be nil")
	}
	if err := ac.Provider.Validate(); err != nil {
		return fmt.Errorf("agent card provider is invalid: %w", err)
	}
	if ac.Capabilities != nil {
		if err := ac.Capabilities.Validate(); err != nil {
			return fmt.Errorf("agent card capabilities is invalid: %w", err)
		}
	}
	for i, skill := range ac.Skills {
		if skill == nil {
			return fmt.Errorf("agent card skill at index %d cannot be nil", i)
		}
		if err := skill.Validate(); err != nil {
			return fmt.Errorf("agent card skill at index %d is invalid: %w", i, err)
		}
	}
	for name, scheme := range ac.SecuritySchemes {
		if scheme == nil {
			return fmt.Errorf("agent card security scheme %s cannot be nil", name)
		}
		if err := scheme.Validate(); err != nil {
			return fmt.Errorf("agent card security scheme %s is invalid: %w", name, err)
		}
	}
	return nil
}

// SecurityScheme represents a security scheme.
type SecurityScheme struct {
	APIKey        *APIKeySecurityScheme        `json:"api_key,omitzero"`
	HTTPAuth      *HTTPAuthSecurityScheme      `json:"http_auth,omitzero"`
	OAuth2        *OAuth2SecurityScheme        `json:"oauth2,omitzero"`
	OpenIDConnect *OpenIDConnectSecurityScheme `json:"openid_connect,omitzero"`
}

// Validate ensures the SecurityScheme is valid.
func (ss SecurityScheme) Validate() error {
	schemes := 0
	if ss.APIKey != nil {
		schemes++
		if err := ss.APIKey.Validate(); err != nil {
			return fmt.Errorf("security scheme API key is invalid: %w", err)
		}
	}
	if ss.HTTPAuth != nil {
		schemes++
		if err := ss.HTTPAuth.Validate(); err != nil {
			return fmt.Errorf("security scheme HTTP auth is invalid: %w", err)
		}
	}
	if ss.OAuth2 != nil {
		schemes++
		if err := ss.OAuth2.Validate(); err != nil {
			return fmt.Errorf("security scheme OAuth2 is invalid: %w", err)
		}
	}
	if ss.OpenIDConnect != nil {
		schemes++
		if err := ss.OpenIDConnect.Validate(); err != nil {
			return fmt.Errorf("security scheme OpenID Connect is invalid: %w", err)
		}
	}
	if schemes != 1 {
		return fmt.Errorf("security scheme must have exactly one scheme type, got %d", schemes)
	}
	return nil
}

// OAuth2SecurityScheme represents an OAuth2 security scheme.
type OAuth2SecurityScheme struct {
	Description string      `json:"description"`
	Flows       *OAuthFlows `json:"flows"`
}

// Validate ensures the OAuth2SecurityScheme is valid.
func (oss OAuth2SecurityScheme) Validate() error {
	if oss.Flows == nil {
		return fmt.Errorf("OAuth2 security scheme flows cannot be nil")
	}
	return oss.Flows.Validate()
}

// OpenIDConnectSecurityScheme represents an OpenID Connect security scheme.
type OpenIDConnectSecurityScheme struct {
	Description      string `json:"description"`
	OpenIDConnectURL string `json:"openid_connect_url"`
}

// Validate ensures the OpenIDConnectSecurityScheme is valid.
func (oidcss OpenIDConnectSecurityScheme) Validate() error {
	if oidcss.OpenIDConnectURL == "" {
		return fmt.Errorf("OpenID Connect security scheme OpenID Connect URL cannot be empty")
	}
	return nil
}

// OAuthFlows represents OAuth flows.
type OAuthFlows struct {
	AuthorizationCode *AuthorizationCodeFlow `json:"authorization_code,omitzero"`
	ClientCredentials *ClientCredentialsFlow `json:"client_credentials,omitzero"`
	Implicit          *ImplicitFlow          `json:"implicit,omitzero"`
	Password          *PasswordFlow          `json:"password,omitzero"`
}

// Validate ensures the OAuthFlows is valid.
func (of OAuthFlows) Validate() error {
	return nil
}

// AuthorizationCodeFlow represents an authorization code flow.
type AuthorizationCodeFlow struct {
	AuthorizationURL string            `json:"authorization_url"`
	TokenURL         string            `json:"token_url"`
	RefreshURL       string            `json:"refresh_url,omitzero"`
	Scopes           map[string]string `json:"scopes,omitzero"`
}

// ClientCredentialsFlow represents a client credentials flow.
type ClientCredentialsFlow struct {
	TokenURL   string            `json:"token_url"`
	RefreshURL string            `json:"refresh_url,omitzero"`
	Scopes     map[string]string `json:"scopes,omitzero"`
}

// ImplicitFlow represents an implicit flow.
type ImplicitFlow struct {
	AuthorizationURL string            `json:"authorization_url"`
	RefreshURL       string            `json:"refresh_url,omitzero"`
	Scopes           map[string]string `json:"scopes,omitzero"`
}

// PasswordFlow represents a password flow.
type PasswordFlow struct {
	TokenURL   string            `json:"token_url"`
	RefreshURL string            `json:"refresh_url,omitzero"`
	Scopes     map[string]string `json:"scopes,omitzero"`
}
