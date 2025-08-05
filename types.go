// Copyright 2025 The Go A2A Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"fmt"
	"net/http"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// Method represents the A2A Protocol RPC Method.
type Method string

const (
	// MethodMessageSend sends a message to an agent to initiate a new interaction or to continue an existing one.
	// This method is suitable for synchronous request/response interactions or when client-side polling (using `tasks/get`) is acceptable for monitoring longer-running tasks.
	MethodMessageSend Method = "message/send"

	// MethodMessageStream sends a message to an agent to initiate/continue a task AND subscribes the client to real-time updates for that task via Server-Sent Events (SSE).
	// This method requires the server to have `AgentCard.capabilities.streaming` is true.
	MethodMessageStream Method = "message/stream"

	// MethodTasksGet retrieves the current state (including status, artifacts, and optionally history) of a previously initiated task.
	// This is typically used for polling the status of a task initiated with `message/send`, or for fetching the final state of a task after being notified via a push notification or after an SSE stream has ended.
	MethodTasksGet Method = "tasks/get"

	// MethodTasksList retrieves a list of tasks associated with the client.
	MethodTasksList Method = "tasks/list"

	// MethodTasksCancel requests the cancellation of an ongoing task.
	// The server will attempt to cancel the task, but success is not guaranteed (e.g., the task might have already completed or failed, or cancellation might not be supported at its current stage).
	MethodTasksCancel Method = "tasks/cancel"

	// MethodTasksPushNotificationConfigSet sets or updates the push notification configuration for a specified task.
	// This allows the client to tell the server where and how to send asynchronous updates for the task. Requires the server to have `AgentCard.capabilities.pushNotifications` is true.
	MethodTasksPushNotificationConfigSet Method = "tasks/pushNotificationConfig/set"

	// MethodTasksPushNotificationConfigGet retrieves the current push notification configuration for a specified task.
	// Requires the server to have `AgentCard.capabilities.pushNotifications` is true.
	MethodTasksPushNotificationConfigGet Method = "tasks/pushNotificationConfig/get"

	// MethodTasksPushNotificationConfigList retrieves the associated push notification configurations for a specified task.
	// Requires the server to have `AgentCard.capabilities.pushNotifications` is true.
	MethodTasksPushNotificationConfigList Method = "tasks/pushNotificationConfig/list"

	// MethodTasksPushNotificationConfigDelete deletes an associated push notification configuration for a task.
	// Requires the server to have `AgentCard.capabilities.pushNotifications` is true.
	MethodTasksPushNotificationConfigDelete Method = "tasks/pushNotificationConfig/delete"

	// MethodTasksResubscribe allows a client to reconnect to an SSE stream for an ongoing task after a previous connection (from `message/stream` or an earlier `tasks/resubscribe`) was interrupted.
	// Requires the server to have `AgentCard.capabilities.streaming` is true.
	MethodTasksResubscribe Method = "tasks/resubscribe"

	// MethodAgentAuthenticatedExtendedCard retrieves a potentially more detailed version of the Agent Card after the client has authenticated.
	// This endpoint is available only if `AgentCard.supportsAuthenticatedExtendedCard` is true.
	MethodAgentAuthenticatedExtendedCard Method = "agent/authenticatedExtendedCard"
)

// In represents an API Key Location enumeration.
type In string

const (
	// InCookie, InHeader, and InQuery represent the possible locations for API keys.
	InCookie In = "cookie"

	// InHeader represents the location of the API key in the HTTP header.
	InHeader In = "header"

	// InQuery represents the location of the API key in the query string.
	InQuery In = "query"
)

// Role represents the message role enumeration.
type Role string

const (
	// RoleAgent and RoleUser represent the possible roles in a message.
	RoleAgent Role = "agent"

	// RoleUser represents the user role in a message.
	RoleUser Role = "user"
)

// PartKind represents the type of part in a message.
type PartKind string

const (
	// TextPartKind represents a text part in a message.
	TextPartKind PartKind = "text"

	// FilePartKind represents a file part in a message.
	FilePartKind PartKind = "file"

	// DataPartKind represents a data part in a message.
	DataPartKind PartKind = "data"
)

// EventKind represents the type of event in the A2A protocol.
type EventKind string

const (
	// MessageEventKind represents a message event.
	MessageEventKind EventKind = "message"

	// ArtifactUpdateEventKind represents an artifact update event.
	ArtifactUpdateEventKind EventKind = "artifact-update"

	// StatusUpdateEventKind represents a status update event.
	StatusUpdateEventKind EventKind = "status-update"

	// TaskEventKind represents a task event.
	TaskEventKind EventKind = "task"
)

// SecuritySchemeType represents the type of security scheme.
type SecuritySchemeType string

const (
	// APIKeySecuritySchemeType represents API Key security scheme.
	APIKeySecuritySchemeType SecuritySchemeType = "apiKey"

	// HTTPSecuritySchemeType represents HTTP Authentication security scheme.
	HTTPSecuritySchemeType SecuritySchemeType = "http"

	// OpenIDConnectSecuritySchemeType represents OpenID Connect security scheme.
	OpenIDConnectSecuritySchemeType SecuritySchemeType = "openIdConnect"

	// OAuth2SecuritySchemeType represents OAuth2 security scheme.
	OAuth2SecuritySchemeType SecuritySchemeType = "oauth2"
)

// TaskState represents the possible states of a Task.
type TaskState string

const (
	TaskStateSubmitted     TaskState = "submitted"
	TaskStateWorking       TaskState = "working"
	TaskStateInputRequired TaskState = "input-required"
	TaskStateCompleted     TaskState = "completed"
	TaskStateCanceled      TaskState = "canceled"
	TaskStateFailed        TaskState = "failed"
	TaskStateRejected      TaskState = "rejected"
	TaskStateAuthRequired  TaskState = "auth-required"
	TaskStateUnknown       TaskState = "unknown"
)

// TransportProtocol supported A2A transport protocols.
type TransportProtocol string

const (
	TransportProtocolJSONRPC  TransportProtocol = "JSONRPC"
	TransportProtocolGRPC     TransportProtocol = "GRPC"
	TransportProtocolHTTPJSON TransportProtocol = "HTTP+JSON"
)

// Params represents all possible A2A request params.
type Params interface {
	GetMetadata() map[string]any
	SetMetadata(map[string]any)
}

// Result represents all possible A2A response result.
type Result interface {
	GetTaskID() string
}

// MessageOrTask represents the result of SendMessage which can be either Message or Task.
type MessageOrTask interface {
	Result
	GetKind() EventKind
}

// StreamingMessageResult represents the result of SendStreamingMessage.
type SendStreamingMessageResponse interface {
	Result
	GetEventKind() EventKind
}

// SecurityScheme represents various security schemes.
type SecurityScheme interface {
	GetType() SecuritySchemeType
	GetDescription() string
}

// Part represents a part of a message, which can be text, a file, or structured data.
type Part interface {
	GetKind() PartKind
}

// File represents file content with either bytes or URI.
type File interface {
	GetMIMEType() string
	GetName() string
}

// Metadata is additional metadata for requests, responses and other types.
type Metadata map[string]any

// GetMeta returns metadata from a value.
func (md Metadata) GetMetadata() map[string]any { return md }

// SetMeta sets the metadata on a value.
func (md *Metadata) SetMetadata(x map[string]any) { *md = x }

// EmptyParams represents an empty set of parameters.
type EmptyParams struct {
	Metadata `json:"metadata,omitzero"`
}

var _ Params = (*EmptyParams)(nil)

// EmptyResult represents an empty result, used when no result is expected.
type EmptyResult struct{}

var _ Result = (*EmptyResult)(nil)

func (r EmptyResult) GetTaskID() string { return "" }

// APIKeySecurityScheme represents API Key security scheme.
type APIKeySecurityScheme struct {
	// Description of this security scheme.
	Description string `json:"description,omitzero"`
	// In specifies the location of the API key.
	// Valid values are "query", "header", or "cookie".
	In In `json:"in"`
	// Name is the name of the header, query or cookie parameter to be used.
	Name string `json:"name"`
	// Type is the security scheme type, always "apiKey".
	Type SecuritySchemeType `json:"type"`
}

var _ SecurityScheme = (*APIKeySecurityScheme)(nil)

func (s APIKeySecurityScheme) GetType() SecuritySchemeType { return s.Type }
func (s APIKeySecurityScheme) GetDescription() string      { return s.Description }

// AgentCardSignature represents a JWS signature of an [AgentCard].
// This follows the JSON format of an RFC 7515 JSON Web Signature (JWS).
type AgentCardSignature struct {
	// header is the unprotected JWS header values.
	Header http.Header `json:"header,omitzero"`
	// Protected is the protected JWS header for the signature. This is a Base64url-encoded.
	// JSON object, as per RFC 7515.
	Protected string `json:"protected"`
	// The computed signature, Base64url-encoded.
	Signature string `json:"signature"`
}

// AgentExtension represents a declaration of an extension supported by an Agent.
type AgentExtension struct {
	// Description of how this agent uses this extension.
	Description string `json:"description,omitzero"`
	// Params contains optional configuration for the extension.
	Params map[string]any `json:"params,omitzero"`
	// Required indicates whether the client must follow specific requirements of the extension.
	Required bool `json:"required,omitzero"`
	// URI is the URI of the extension.
	URI string `json:"uri"`
}

// AgentInterface provides a declaration of a combination of the target url and the supported transport.
type AgentInterface struct {
	// Transport is the transport supported by this URL.
	// This is an open form string, to be easily extended for many transport protocols.
	// The core ones officially supported are JSONRPC, GRPC and HTTP+JSON.
	Transport TransportProtocol `json:"transport"`
	// URL is the target URL for this interface.
	URL string `json:"url"`
}

// AgentProvider represents the service provider of an agent.
type AgentProvider struct {
	// Organization is the agent provider's organization name.
	Organization string `json:"organization"`
	// URL is the agent provider's URL.
	URL string `json:"url"`
}

// AgentSkill represents a unit of capability that an agent can perform.
type AgentSkill struct {
	// Description of the skill - will be used by the client or a human
	// as a hint to understand what the skill does.
	Description string `json:"description"`
	// Examples contains example scenarios that the skill can perform.
	// Will be used by the client as a hint to understand how the skill can be used.
	// Example: ["I need a recipe for bread"]
	Examples []string `json:"examples,omitzero"`
	// ID is the unique identifier for the agent's skill.
	ID string `json:"id"`
	// InputModes contains the set of interaction modes that the skill supports
	// (if different than the default). Supported media types for input.
	InputModes []string `json:"inputModes,omitzero"`
	// Name is the human readable name of the skill.
	Name string `json:"name"`
	// OutputModes contains supported media types for output.
	OutputModes []string `json:"outputModes,omitzero"`
	// Tags is a set of tagwords describing classes of capabilities for this specific skill.
	// Examples: ["cooking", "customer support", "billing"]
	Tags []string `json:"tags"`
}

// AuthorizationCodeOAuthFlow represents configuration details for a supported OAuth Flow.
type AuthorizationCodeOAuthFlow struct {
	// AuthorizationURL is the authorization URL to be used for this flow.
	// This MUST be in the form of a URL. The OAuth2 standard requires the use of TLS.
	AuthorizationURL string `json:"authorizationUrl"`
	// RefreshURL is the URL to be used for obtaining refresh tokens.
	// This MUST be in the form of a URL. The OAuth2 standard requires the use of TLS.
	RefreshURL string `json:"refreshUrl,omitzero"`
	// Scopes contains the available scopes for the OAuth2 security scheme.
	// A map between the scope name and a short description for it. The map MAY be empty.
	Scopes map[string]string `json:"scopes"`
	// TokenURL is the token URL to be used for this flow.
	// This MUST be in the form of a URL. The OAuth2 standard requires the use of TLS.
	TokenURL string `json:"tokenUrl"`
}

// ClientCredentialsOAuthFlow represents configuration details for a supported OAuth Flow.
type ClientCredentialsOAuthFlow struct {
	// RefreshURL is the URL to be used for obtaining refresh tokens.
	// This MUST be in the form of a URL. The OAuth2 standard requires the use of TLS.
	RefreshURL string `json:"refreshUrl,omitzero"`
	// Scopes contains the available scopes for the OAuth2 security scheme.
	// A map between the scope name and a short description for it. The map MAY be empty.
	Scopes map[string]string `json:"scopes"`
	// TokenURL is the token URL to be used for this flow.
	// This MUST be in the form of a URL. The OAuth2 standard requires the use of TLS.
	TokenURL string `json:"tokenUrl"`
}

// DataPart represents a structured data segment within a message part.
type DataPart struct {
	// Data contains the structured data content.
	Data map[string]any `json:"data"`
	// Kind is the part type, always "data" for DataParts.
	Kind PartKind `json:"kind"`
	// Metadata contains optional metadata associated with the part.
	Metadata map[string]any `json:"metadata,omitzero"`
}

var _ Part = (*DataPart)(nil)

// NewDataPart creates a new [DataPart] with the provided data.
func NewDataPart(data map[string]any) *DataPart {
	return &DataPart{
		Data: data,
		Kind: DataPartKind,
	}
}

func (p DataPart) GetKind() PartKind { return p.Kind }

// DeleteTaskPushNotificationConfigParams represents parameters for removing pushNotificationConfiguration.
type DeleteTaskPushNotificationConfigParams struct {
	Metadata `json:"metadata,omitzero"`

	// ID is the task id.
	ID string `json:"id"`
	// PushNotificationConfigID is the push notification configuration ID to delete.
	PushNotificationConfigID string `json:"pushNotificationConfigId"`
}

var _ Params = (*DeleteTaskPushNotificationConfigParams)(nil)

// FileWithBytes represents a file with bytes content.
type FileWithBytes struct {
	// Bytes contains the base64 encoded content of the file.
	Bytes string `json:"bytes"`
	// MIMEType is the optional mimeType for the file.
	MIMEType string `json:"mimeType,omitzero"`
	// Name is the optional name for the file.
	Name string `json:"name,omitzero"`
}

var _ File = (*FileWithBytes)(nil)

func (f FileWithBytes) GetMIMEType() string { return f.MIMEType }
func (f FileWithBytes) GetName() string     { return f.Name }

// FileWithURI represents a file with URI content.
type FileWithURI struct {
	// MIMEType is the optional mimeType for the file.
	MIMEType string `json:"mimeType,omitzero"`
	// Name is the optional name for the file.
	Name string `json:"name,omitzero"`
	// URI is the URL for the file content.
	URI string `json:"uri"`
}

var _ File = (*FileWithURI)(nil)

func (f FileWithURI) GetMIMEType() string { return f.MIMEType }
func (f FileWithURI) GetName() string     { return f.Name }

// GetTaskPushNotificationConfigParams represents parameters for fetching pushNotificationConfiguration.
type GetTaskPushNotificationConfigParams struct {
	Metadata `json:"metadata,omitzero"`

	// ID is the task id.
	ID string `json:"id"`
	// Metadata contains extension metadata.
	// PushNotificationConfigID is the push notification configuration ID to fetch (optional).
	PushNotificationConfigID string `json:"pushNotificationConfigId,omitzero"`
}

var _ Params = (*GetTaskPushNotificationConfigParams)(nil)

// HTTPAuthSecurityScheme represents HTTP Authentication security scheme.
type HTTPAuthSecurityScheme struct {
	// BearerFormat is a hint to the client to identify how the bearer token is formatted.
	// Bearer tokens are usually generated by an authorization server,
	// so this information is primarily for documentation purposes.
	BearerFormat string `json:"bearerFormat,omitzero"`
	// Description of this security scheme.
	Description string `json:"description,omitzero"`
	// Scheme is the name of the HTTP Authentication scheme to be used in the Authorization header
	// as defined in RFC7235. The values used SHOULD be registered in the IANA Authentication Scheme registry.
	// The value is case-insensitive, as defined in RFC7235.
	Scheme string `json:"scheme"`
	// Type is the security scheme type, always "http".
	Type SecuritySchemeType `json:"type"`
}

var _ SecurityScheme = (*HTTPAuthSecurityScheme)(nil)

func (s HTTPAuthSecurityScheme) GetType() SecuritySchemeType { return s.Type }
func (s HTTPAuthSecurityScheme) GetDescription() string      { return s.Description }

// ImplicitOAuthFlow represents configuration details for a supported OAuth Flow.
type ImplicitOAuthFlow struct {
	// AuthorizationURL is the authorization URL to be used for this flow.
	// This MUST be in the form of a URL. The OAuth2 standard requires the use of TLS.
	AuthorizationURL string `json:"authorizationUrl"`
	// RefreshURL is the URL to be used for obtaining refresh tokens.
	// This MUST be in the form of a URL. The OAuth2 standard requires the use of TLS.
	RefreshURL string `json:"refreshUrl,omitzero"`
	// Scopes contains the available scopes for the OAuth2 security scheme.
	// A map between the scope name and a short description for it. The map MAY be empty.
	Scopes map[string]string `json:"scopes"`
}

// ListTaskPushNotificationConfigParams represents parameters for getting list of pushNotificationConfigurations.
type ListTaskPushNotificationConfigParams struct {
	Metadata `json:"metadata,omitzero"`

	// ID is the task id.
	ID string `json:"id"`
}

var _ Params = (*ListTaskPushNotificationConfigParams)(nil)

// OpenIDConnectSecurityScheme represents OpenID Connect security scheme configuration.
type OpenIDConnectSecurityScheme struct {
	// Description of this security scheme.
	Description string `json:"description,omitzero"`
	// OpenIDConnectURL is the well-known URL to discover the OpenID Connect provider metadata.
	OpenIDConnectURL string `json:"openIdConnectUrl"`
	// Type is the security scheme type, always "openIdConnect".
	Type SecuritySchemeType `json:"type"`
}

var _ SecurityScheme = (*OpenIDConnectSecurityScheme)(nil)

func (s OpenIDConnectSecurityScheme) GetType() SecuritySchemeType { return s.Type }
func (s OpenIDConnectSecurityScheme) GetDescription() string      { return s.Description }

// PasswordOAuthFlow represents configuration details for a supported OAuth Flow.
type PasswordOAuthFlow struct {
	// RefreshURL is the URL to be used for obtaining refresh tokens.
	// This MUST be in the form of a URL. The OAuth2 standard requires the use of TLS.
	RefreshURL string `json:"refreshUrl,omitzero"`
	// Scopes contains the available scopes for the OAuth2 security scheme.
	// A map between the scope name and a short description for it. The map MAY be empty.
	Scopes map[string]string `json:"scopes"`
	// TokenURL is the token URL to be used for this flow.
	// This MUST be in the form of a URL. The OAuth2 standard requires the use of TLS.
	TokenURL string `json:"tokenUrl"`
}

// PushNotificationAuthenticationInfo defines authentication details for push notifications.
type PushNotificationAuthenticationInfo struct {
	// Credentials contains optional credentials.
	Credentials string `json:"credentials,omitzero"`
	// Schemes contains supported authentication schemes (e.g., Basic, Bearer).
	Schemes []string `json:"schemes"`
}

// PushNotificationConfig represents configuration for setting up push notifications for task updates.
type PushNotificationConfig struct {
	// Authentication contains authentication details for push notifications.
	Authentication *PushNotificationAuthenticationInfo `json:"authentication,omitzero"`
	// ID is the Push Notification ID created by server to support multiple callbacks.
	ID string `json:"id,omitzero"`
	// Token is unique to this task/session.
	Token string `json:"token,omitzero"`
	// URL is the URL for sending the push notifications.
	URL string `json:"url"`
}

// SecuritySchemeBase represents base properties shared by all security schemes.
type SecuritySchemeBase struct {
	// Description of this security scheme.
	Description string `json:"description,omitzero"`
}

// TaskIDParams represents parameters containing only a task ID, used for simple task operations.
type TaskIDParams struct {
	Metadata `json:"metadata,omitzero"`

	// ID is the task id.
	ID string `json:"id"`
}

var _ Params = (*TaskIDParams)(nil)

// TaskPushNotificationConfig represents parameters for setting or getting push notification configuration.
type TaskPushNotificationConfig struct {
	// PushNotificationConfig contains the push notification configuration.
	PushNotificationConfig *PushNotificationConfig `json:"pushNotificationConfig"`
	// TaskID is the task id.
	TaskID string `json:"taskId"`
}

var (
	_ Params = (*TaskPushNotificationConfig)(nil)
	_ Result = (*TaskPushNotificationConfig)(nil)
)

func (e TaskPushNotificationConfig) GetMetadata() map[string]any { return nil }
func (e TaskPushNotificationConfig) SetMetadata(map[string]any)  {}
func (e TaskPushNotificationConfig) GetTaskID() string           { return e.TaskID }

type TaskPushNotificationConfigs []*TaskPushNotificationConfig

var _ Result = (TaskPushNotificationConfigs)(nil)

func (e TaskPushNotificationConfigs) GetTaskID() string { return e[0].TaskID }

// TaskQueryParams represents parameters for querying a task, including optional history length.
type TaskQueryParams struct {
	Metadata `json:"metadata,omitzero"`

	// HistoryLength is the number of recent messages to be retrieved.
	HistoryLength int `json:"historyLength,omitzero"`
	// ID is the task id.
	ID string `json:"id"`
}

var _ Params = (*TaskQueryParams)(nil)

// TextPart represents a text segment within parts.
type TextPart struct {
	// Kind is the part type, always "text" for TextParts.
	Kind PartKind `json:"kind"`
	// Metadata contains optional metadata associated with the part.
	Metadata map[string]any `json:"metadata,omitzero"`
	// Text contains the text content.
	Text string `json:"text"`
}

var _ Part = (*TextPart)(nil)

// NewTextPart creates a new [TextPart] with the provided text.
func NewTextPart(text string) *TextPart {
	return &TextPart{
		Kind: TextPartKind,
		Text: text,
	}
}

func (p TextPart) GetKind() PartKind { return p.Kind }

// AgentCapabilities defines optional capabilities supported by an agent.
type AgentCapabilities struct {
	// Extensions contains extensions supported by this agent.
	Extensions []*AgentExtension `json:"extensions,omitzero"`
	// PushNotifications indicates if the agent can notify updates to client.
	PushNotifications bool `json:"pushNotifications,omitzero"`
	// StateTransitionHistory indicates if the agent exposes status change history for tasks.
	StateTransitionHistory bool `json:"stateTransitionHistory,omitzero"`
	// Streaming indicates if the agent supports Server-Sent Events (SSE).
	Streaming bool `json:"streaming,omitzero"`
}

// FilePart represents a File segment within parts.
type FilePart struct {
	// File contains the file content either as URL or bytes.
	File File `json:"file"`
	// Kind is the part type, always "file" for FileParts.
	Kind PartKind `json:"kind"`
	// Metadata contains optional metadata associated with the part.
	Metadata map[string]any `json:"metadata,omitzero"`
}

var _ Part = (*FilePart)(nil)

func (p FilePart) GetKind() PartKind { return p.Kind }

// NewFilePart creates a new [FilePart] with the provided file.
func NewFilePart(file File) *FilePart {
	return &FilePart{
		File: file,
		Kind: FilePartKind,
	}
}

// MarshalJSON implements json.Marshaler for FilePart
func (f *FilePart) MarshalJSON() ([]byte, error) {
	type Alias FilePart
	return json.Marshal((*Alias)(f))
}

// UnmarshalJSON implements json.Unmarshaler for FilePart
func (f *FilePart) UnmarshalJSON(data []byte) error {
	type Alias FilePart
	alias := &struct {
		File jsontext.Value `json:"file"`
		*Alias
	}{
		Alias: (*Alias)(f),
	}

	if err := json.Unmarshal(data, alias); err != nil {
		return err
	}

	file, err := UnmarshalFileContent(alias.File)
	if err != nil {
		return err
	}
	f.File = file

	return nil
}

// MessageSendConfiguration represents configuration for the send message request.
type MessageSendConfiguration struct {
	// AcceptedOutputModes contains accepted output modalities by the client.
	AcceptedOutputModes []string `json:"acceptedOutputModes"`
	// Blocking indicates if the server should treat the client as a blocking request.
	Blocking bool `json:"blocking,omitzero"`
	// HistoryLength is the number of recent messages to be retrieved.
	HistoryLength int `json:"historyLength,omitzero"`
	// PushNotificationConfig specifies where the server should send notifications when disconnected.
	PushNotificationConfig *PushNotificationConfig `json:"pushNotificationConfig,omitzero"`
}

// OAuthFlows represents OAuth flow configurations.
type OAuthFlows struct {
	// AuthorizationCode contains configuration for the OAuth Authorization Code flow.
	// Previously called accessCode in OpenAPI 2.0.
	AuthorizationCode *AuthorizationCodeOAuthFlow `json:"authorizationCode,omitzero"`
	// ClientCredentials contains configuration for the OAuth Client Credentials flow.
	// Previously called application in OpenAPI 2.0.
	ClientCredentials *ClientCredentialsOAuthFlow `json:"clientCredentials,omitzero"`
	// Implicit contains configuration for the OAuth Implicit flow.
	Implicit *ImplicitOAuthFlow `json:"implicit,omitzero"`
	// Password contains configuration for the OAuth Resource Owner Password flow.
	Password *PasswordOAuthFlow `json:"password,omitzero"`
}

// OAuth2SecurityScheme represents OAuth2.0 security scheme configuration.
type OAuth2SecurityScheme struct {
	// Description of this security scheme.
	Description string `json:"description,omitzero"`
	// Flows contains configuration information for the flow types supported.
	Flows *OAuthFlows `json:"flows"`
	// Type is the security scheme type, always "oauth2".
	Type SecuritySchemeType `json:"type"`
}

var _ SecurityScheme = (*OAuth2SecurityScheme)(nil)

func (s OAuth2SecurityScheme) GetType() SecuritySchemeType { return s.Type }
func (s OAuth2SecurityScheme) GetDescription() string      { return s.Description }

// TaskArtifactUpdateEvent represents event sent by server during sendStream or subscribe requests.
type TaskArtifactUpdateEvent struct {
	Metadata `json:"metadata,omitzero"`

	// Append indicates if this artifact appends to a previous one.
	Append bool `json:"append,omitzero"`
	// Artifact is the generated artifact.
	Artifact *Artifact `json:"artifact"`
	// ContextID is the context the task is associated with.
	ContextID string `json:"contextId"`
	// Kind is the event type, always "artifact-update".
	Kind EventKind `json:"kind"`
	// LastChunk indicates if this is the last chunk of the artifact.
	LastChunk bool `json:"lastChunk,omitzero"`
	// TaskID is the task id.
	TaskID string `json:"taskId"`
}

var (
	_ SendStreamingMessageResponse = (*TaskArtifactUpdateEvent)(nil)
	_ Result                       = (*TaskArtifactUpdateEvent)(nil)
)

func (e TaskArtifactUpdateEvent) GetEventKind() EventKind { return e.Kind }
func (e TaskArtifactUpdateEvent) GetTaskID() string       { return e.TaskID }

// TaskStatusUpdateEvent represents event sent by server during sendStream or subscribe requests.
type TaskStatusUpdateEvent struct {
	// ContextID is the context the task is associated with.
	ContextID string `json:"contextId"`
	// Final indicates the end of the event stream.
	Final bool `json:"final"`
	// Kind is the event type, always "status-update".
	Kind EventKind `json:"kind"`
	// Metadata contains extension metadata.
	Metadata map[string]any `json:"metadata,omitzero"`
	// Status is the current status of the task.
	Status TaskStatus `json:"status"`
	// TaskID is the task id.
	TaskID string `json:"taskId"`
}

var (
	_ SendStreamingMessageResponse = (*TaskStatusUpdateEvent)(nil)
	_ Result                       = (*TaskStatusUpdateEvent)(nil)
)

func (e TaskStatusUpdateEvent) GetEventKind() EventKind { return e.Kind }
func (e TaskStatusUpdateEvent) GetTaskID() string       { return e.TaskID }

// Artifact represents an artifact generated for a task.
type Artifact struct {
	// ArtifactID is the unique identifier for the artifact.
	ArtifactID string `json:"artifactId"`
	// Description is an optional description for the artifact.
	Description string `json:"description,omitzero"`
	// Extensions contains the URIs of extensions that are present or contributed to this Artifact.
	Extensions []string `json:"extensions,omitzero"`
	// Metadata contains extension metadata.
	Metadata map[string]any `json:"metadata,omitzero"`
	// Name is an optional name for the artifact.
	Name string `json:"name,omitzero"`
	// Parts contains the artifact parts.
	Parts []Part `json:"parts"`
}

// MarshalJSON implements json.Marshaler for Artifact
func (a *Artifact) MarshalJSON() ([]byte, error) {
	type Alias Artifact
	return json.Marshal((*Alias)(a))
}

// UnmarshalJSON implements json.Unmarshaler for Artifact
func (a *Artifact) UnmarshalJSON(data []byte) error {
	type Alias Artifact
	alias := &struct {
		Parts []jsontext.Value `json:"parts"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, alias); err != nil {
		return err
	}

	a.Parts = make([]Part, len(alias.Parts))
	for i, rawPart := range alias.Parts {
		part, err := UnmarshalPart(rawPart)
		if err != nil {
			return err
		}
		a.Parts[i] = part
	}

	return nil
}

// MessageSendParams represents parameters sent by the client to the agent as a request.
type MessageSendParams struct {
	Metadata `json:"metadata,omitzero"`

	// Configuration contains send message configuration.
	Configuration *MessageSendConfiguration `json:"configuration,omitzero"`
	// Message is the message being sent to the server.
	Message *Message `json:"message"`
}

var _ Params = (*MessageSendParams)(nil)

// Message represents a single message exchanged between user and agent.
type Message struct {
	// ContextID is the context the message is associated with.
	ContextID string `json:"contextId,omitzero"`
	// Extensions contains the URIs of extensions that are present or contributed to this Message.
	Extensions []string `json:"extensions,omitzero"`
	// Kind is the event type, always "message".
	Kind EventKind `json:"kind"`
	// MessageID is the identifier created by the message creator.
	MessageID string `json:"messageId"`
	// Metadata contains extension metadata.
	Metadata map[string]any `json:"metadata,omitzero"`
	// Parts contains the message content.
	Parts []Part `json:"parts"`
	// ReferenceTaskIDs contains list of tasks referenced as context by this message.
	ReferenceTaskIDs []string `json:"referenceTaskIds,omitzero"`
	// Role is the message sender's role ("user" or "agent").
	Role Role `json:"role"`
	// TaskID is the identifier of task the message is related to.
	TaskID string `json:"taskId,omitzero"`
}

var (
	_ MessageOrTask                = (*Message)(nil)
	_ SendStreamingMessageResponse = (*Message)(nil)
	_ Result                       = (*Message)(nil)
)

func (m Message) GetKind() EventKind      { return m.Kind }
func (e Message) GetEventKind() EventKind { return e.Kind }
func (m Message) GetTaskID() string       { return m.TaskID }

// MarshalJSON implements json.Marshaler for Message
func (m *Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	return json.Marshal((*Alias)(m))
}

// UnmarshalJSON implements json.Unmarshaler for Message
func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	alias := &struct {
		Parts []jsontext.Value `json:"parts"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, alias); err != nil {
		return err
	}

	m.Parts = make([]Part, len(alias.Parts))
	for i, rawPart := range alias.Parts {
		part, err := UnmarshalPart(rawPart)
		if err != nil {
			return err
		}
		m.Parts[i] = part
	}

	return nil
}

// TaskStatus represents TaskState and accompanying message.
type TaskStatus struct {
	// Message contains additional status updates for client.
	Message *Message `json:"message,omitzero"`
	// State is the current state of the task.
	State TaskState `json:"state"`
	// Timestamp is the ISO 8601 datetime string when the status was recorded.
	// Example: "2023-10-27T10:00:00Z"
	Timestamp string `json:"timestamp,omitzero"`
}

type SecurityScope string

// SecurityScopes slice of [SecurityScope].
type SecurityScopes []SecurityScope

func NewSecurityScopes(scopes ...string) SecurityScopes {
	ss := make(SecurityScopes, len(scopes))
	for i, s := range scopes {
		ss[i] = SecurityScope(s)
	}
	return ss
}

func (ss SecurityScopes) AsSlices() []string {
	scopes := make([]string, len(ss))
	for i, s := range ss {
		scopes[i] = string(s)
	}
	return scopes
}

// AgentCard represents an AgentCard conveys key information about an agent.
type AgentCard struct {
	// AdditionalInterfaces contains announcement of additional supported transports.
	// Client can use any of the supported transports.
	AdditionalInterfaces []*AgentInterface `json:"additionalInterfaces,omitzero"`
	// Capabilities contains optional capabilities supported by the agent.
	Capabilities *AgentCapabilities `json:"capabilities"`
	// DefaultInputModes contains the set of interaction modes that the agent supports across all skills.
	// This can be overridden per-skill. Supported media types for input.
	DefaultInputModes []string `json:"defaultInputModes"`
	// DefaultOutputModes contains supported media types for output.
	DefaultOutputModes []string `json:"defaultOutputModes"`
	// Description is a human-readable description of the agent.
	// Used to assist users and other agents in understanding what the agent can do.
	// Example: "Agent that helps users with recipes and cooking."
	Description string `json:"description"`
	// DocumentationURL is a URL to documentation for the agent.
	DocumentationURL string `json:"documentationUrl,omitzero"`
	// IconURL is a URL to an icon for the agent.
	IconURL string `json:"iconUrl,omitzero"`
	// Name is the human readable name of the agent.
	// Example: "Recipe Agent"
	Name string `json:"name"`
	// PreferredTransport is the transport protocol for the preferred endpoint (the main 'url' field).
	// If not specified, defaults to 'JSONRPC'.
	//
	// IMPORTANT: The transport specified here MUST be available at the main 'url'.
	// This creates a binding between the main URL and its supported transport protocol.
	// Clients should prefer this transport and URL combination when both are supported.
	PreferredTransport TransportProtocol `json:"preferredTransport,omitzero"`
	// ProtocolVersion is the version of the A2A protocol this agent supports.
	// Default: "0.2.5"
	ProtocolVersion string `json:"protocolVersion,omitzero"`
	// Provider is the service provider of the agent.
	Provider *AgentProvider `json:"provider,omitzero"`
	// Security contains security requirements for contacting the agent.
	Security []map[string]SecurityScopes `json:"security,omitzero"`
	// SecuritySchemes contains security scheme details used for authenticating with this agent.
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitzero"`
	// Signatures is the JSON Web Signatures computed for this AgentCard.
	Signatures *AgentCardSignature `json:"signatures,omitzero"`
	// Skills contains the skills that are units of capability that an agent can perform.
	Skills []*AgentSkill `json:"skills"`
	// SupportsAuthenticatedExtendedCard indicates if the agent supports providing an extended
	// agent card when the user is authenticated. Defaults to false if not specified.
	SupportsAuthenticatedExtendedCard bool `json:"supportsAuthenticatedExtendedCard,omitzero"`
	// URL is the URL to the address the agent is hosted at.
	// This represents the preferred endpoint as declared by the agent.
	URL string `json:"url"`
	// Version is the version of the agent - format is up to the provider.
	// Example: "1.0.0"
	Version string `json:"version"`
}

var _ Result = (*AgentCard)(nil)

func (e AgentCard) GetTaskID() string { return "" }

// MarshalJSON implements json.Marshaler for AgentCard
func (a *AgentCard) MarshalJSON() ([]byte, error) {
	type Alias AgentCard
	return json.Marshal((*Alias)(a))
}

// UnmarshalJSON implements json.Unmarshaler for AgentCard
func (a *AgentCard) UnmarshalJSON(data []byte) error {
	type Alias AgentCard
	alias := &struct {
		SecuritySchemes map[string]jsontext.Value `json:"securitySchemes"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, alias); err != nil {
		return err
	}

	if alias.SecuritySchemes != nil {
		a.SecuritySchemes = make(map[string]SecurityScheme)
		for key, rawScheme := range alias.SecuritySchemes {
			scheme, err := UnmarshalSecurityScheme(rawScheme)
			if err != nil {
				return err
			}
			a.SecuritySchemes[key] = scheme
		}
	}

	return nil
}

// Task represents a task in the A2A system.
type Task struct {
	// Artifacts contains the collection of artifacts created by the agent.
	Artifacts []*Artifact `json:"artifacts,omitzero"`
	// ContextID is the server-generated id for contextual alignment across interactions.
	ContextID string `json:"contextId"`
	// History contains the message history for this task.
	History []*Message `json:"history,omitzero"`
	// ID is the unique identifier for the task.
	ID string `json:"id"`
	// Kind is the event type, always "task".
	Kind EventKind `json:"kind"`
	// Metadata contains extension metadata.
	Metadata map[string]any `json:"metadata,omitzero"`
	// Status is the current status of the task.
	Status TaskStatus `json:"status"`
}

var (
	_ MessageOrTask                = (*Task)(nil)
	_ SendStreamingMessageResponse = (*Task)(nil)
	_ Result                       = (*Task)(nil)
)

func (t Task) GetKind() EventKind      { return t.Kind }
func (e Task) GetEventKind() EventKind { return e.Kind }
func (e Task) GetTaskID() string       { return e.ID }

// MarshalJSON implements json.Marshaler for Task
func (t *Task) MarshalJSON() ([]byte, error) {
	type Alias Task
	return json.Marshal((*Alias)(t))
}

// UnmarshalJSON implements json.Unmarshaler for Task
func (t *Task) UnmarshalJSON(data []byte) error {
	type Alias Task
	alias := &struct {
		Messages []jsontext.Value `json:"messages"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, alias); err != nil {
		return err
	}

	if alias.Messages != nil {
		t.History = make([]*Message, len(alias.History))
		for i, rawMessage := range alias.Messages {
			var msg Message
			if err := msg.UnmarshalJSON(rawMessage); err != nil {
				return err
			}
			t.History[i] = &msg
		}
	}

	return nil
}

// UnmarshalParams unmarshals [Params] from JSON data.
func UnmarshalParams(data []byte) (Params, error) {
	var base struct {
		Method Method `json:"method"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	switch base.Method {
	case MethodMessageSend:
		var req MessageSendParams
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return &req, nil
	case MethodMessageStream:
		var req MessageSendParams
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return &req, nil
	case MethodTasksGet:
		var req TaskQueryParams
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return &req, nil
	case MethodTasksList:
		var req EmptyParams
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return &req, nil
	case MethodTasksCancel:
		var req TaskQueryParams
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return &req, nil
	case MethodTasksPushNotificationConfigSet:
		var req TaskPushNotificationConfig
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return &req, nil
	case MethodTasksPushNotificationConfigGet:
		var req GetTaskPushNotificationConfigParams
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return &req, nil
	case MethodTasksPushNotificationConfigList:
		var req ListTaskPushNotificationConfigParams
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return &req, nil
	case MethodTasksPushNotificationConfigDelete:
		var req DeleteTaskPushNotificationConfigParams
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return &req, nil
	case MethodTasksResubscribe:
		var req TaskIDParams
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return &req, nil
	case MethodAgentAuthenticatedExtendedCard:
		var req EmptyParams
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return &req, nil
	default:
		return nil, fmt.Errorf("unknown method: %s", base.Method)
	}
}

// UnmarshalSecurityScheme unmarshals SecurityScheme from JSON data.
func UnmarshalSecurityScheme(data []byte) (SecurityScheme, error) {
	var base struct {
		Type SecuritySchemeType `json:"type"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	switch base.Type {
	case APIKeySecuritySchemeType:
		var scheme APIKeySecurityScheme
		if err := json.Unmarshal(data, &scheme); err != nil {
			return nil, err
		}
		return &scheme, nil
	case HTTPSecuritySchemeType:
		var scheme HTTPAuthSecurityScheme
		if err := json.Unmarshal(data, &scheme); err != nil {
			return nil, err
		}
		return &scheme, nil
	case OpenIDConnectSecuritySchemeType:
		var scheme OpenIDConnectSecurityScheme
		if err := json.Unmarshal(data, &scheme); err != nil {
			return nil, err
		}
		return &scheme, nil
	case OAuth2SecuritySchemeType:
		var scheme OAuth2SecurityScheme
		if err := json.Unmarshal(data, &scheme); err != nil {
			return nil, err
		}
		return &scheme, nil
	default:
		return nil, fmt.Errorf("unknown security scheme type: %s", base.Type)
	}
}

// UnmarshalPart unmarshals Part from JSON data.
func UnmarshalPart(data []byte) (Part, error) {
	var base struct {
		Kind PartKind `json:"kind"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	switch base.Kind {
	case TextPartKind:
		var part TextPart
		if err := json.Unmarshal(data, &part); err != nil {
			return nil, err
		}
		return &part, nil
	case FilePartKind:
		var part FilePart
		if err := json.Unmarshal(data, &part); err != nil {
			return nil, err
		}
		return &part, nil
	case DataPartKind:
		var part DataPart
		if err := json.Unmarshal(data, &part); err != nil {
			return nil, err
		}
		return &part, nil
	default:
		return nil, fmt.Errorf("unknown part kind: %s", base.Kind)
	}
}

// UnmarshalMessageOrTask unmarshals [MessageOrTask] from JSON data.
func UnmarshalMessageOrTask(data []byte) (MessageOrTask, error) {
	var base struct {
		Kind EventKind `json:"kind"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	switch base.Kind {
	case MessageEventKind:
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return &msg, nil
	case TaskEventKind:
		var task Task
		if err := json.Unmarshal(data, &task); err != nil {
			return nil, err
		}
		return &task, nil
	default:
		return nil, fmt.Errorf("unknown message or task kind: %s", base.Kind)
	}
}

// UnmarshalSendStreamingMessageResponse unmarshals [SendStreamingMessageResponse] from JSON data.
func UnmarshalSendStreamingMessageResponse(data []byte) (SendStreamingMessageResponse, error) {
	var base struct {
		Kind EventKind `json:"kind"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	switch base.Kind {
	case ArtifactUpdateEventKind:
		var msg TaskArtifactUpdateEvent
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return &msg, nil
	case StatusUpdateEventKind:
		var msg TaskStatusUpdateEvent
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return &msg, nil
	case MessageEventKind:
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return &msg, nil
	case TaskEventKind:
		var task Task
		if err := json.Unmarshal(data, &task); err != nil {
			return nil, err
		}
		return &task, nil
	default:
		return nil, fmt.Errorf("unknown message or task kind: %s", base.Kind)
	}
}

// UnmarshalFileContent unmarshals FileContent from JSON data.
func UnmarshalFileContent(data []byte) (File, error) {
	var base struct {
		Bytes string `json:"bytes"`
		URI   string `json:"uri"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	switch {
	case base.Bytes != "":
		var file FileWithBytes
		if err := json.Unmarshal(data, &file); err != nil {
			return nil, err
		}
		return &file, nil
	case base.URI != "":
		var file FileWithURI
		if err := json.Unmarshal(data, &file); err != nil {
			return nil, err
		}
		return &file, nil
	default:
		return nil, fmt.Errorf("file content must have either bytes or uri")
	}
}
