// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package a2a provides an open protocol enabling communication and interoperability between opaque agentic applications for Go.
package a2a

import "time"

// Version is the current version of the A2A protocol.
const Version = "v0.2.5"

// TaskState represents the lifecycle state of a task.
type TaskState string

// Task state constants.
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

// MessageRole represents the possible roles in a message.
type MessageRole string

const (
	// RoleAgent indicates the message is from the agent.
	MessageRoleAgent MessageRole = "agent"

	// RoleUser indicates the message is from the client.
	MessageRoleUser MessageRole = "user"
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

// Represents the service provider of an agent.
type AgentProvider struct {
	// Agent provider's organization name.
	Organization string `json:"organization"`

	// Agent provider's URL.
	URL string `json:"url"`
}

// Defines optional capabilities supported by an agent.
type AgentCapabilities struct {
	// true if the agent supports SSE.
	Streaming bool `json:"streaming,omitzero"`

	// true if the agent can notify updates to client.
	PushNotifications bool `json:"pushNotifications,omitzero"`

	// true if the agent exposes status change history for tasks.
	StateTransitionHistory bool `json:"stateTransitionHistory,omitzero"`

	// extensions supported by this agent.
	Extensions []AgentExtension `json:"extensions,omitempty"`
}

// A declaration of an extension supported by an Agent.
type AgentExtension struct {
	// The URI of the extension.
	URI string `json:"uri"`

	// A description of how this agent uses this extension.
	Description string `json:"description,omitzero"`

	// Whether the client must follow specific requirements of the extension.
	Required bool `json:"required,omitzero"`

	// Optional configuration for the extension.
	Params map[string]any `json:"params,omitempty"`
}

// Represents a unit of capability that an agent can perform.
type AgentSkill struct {
	// Unique identifier for the agent's skill.
	ID string `json:"id"`

	// Human readable name of the skill.
	Name string `json:"name"`

	// Description of the skill - will be used by the client or a human
	// as a hint to understand what the skill does.
	Description string `json:"description"`

	// Set of tagwords describing classes of capabilities for this specific skill.
	Tags []string `json:"tags"`

	// The set of example scenarios that the skill can perform.
	// Will be used by the client as a hint to understand how the skill can be used.
	Examples []string `json:"examples,omitempty"`

	// The set of interaction modes that the skill supports
	// (if different than the default).
	// Supported media types for input.
	InputModes []string `json:"inputModes,omitempty"`

	// Supported media types for output.
	OutputModes []string `json:"outputModes,omitempty"`
}

// AgentInterface provides a declaration of a combination of the
// target url and the supported transport to interact with the agent.
type AgentInterface struct {
	// URL corresponds to the JSON schema field "url".
	URL string `json:"url"`

	// The transport supported this url. This is an open form string, to be
	// easily extended for many transport protocols. The core ones officially
	// supported are JSONRPC, GRPC and HTTP+JSON.
	Transport string `json:"transport"`
}

// An AgentCard conveys key information:
// - Overall details (version, name, description, uses)
// - Skills: A set of capabilities the agent can perform
// - Default modalities/content types supported by the agent.
// - Authentication requirements
type AgentCard struct {
	// The version of the A2A protocol this agent supports.
	ProtocolVersion string `json:"protocolVersion"`

	// Human readable name of the agent.
	Name string `json:"name"`

	// A human-readable description of the agent. Used to assist users and
	// other agents in understanding what the agent can do.
	Description string `json:"description"`

	// A URL to the address the agent is hosted at. This represents the
	// preferred endpoint as declared by the agent.
	URL string `json:"url"`

	// The transport of the preferred endpoint. If empty, defaults to JSONRPC.
	PreferredTransport string `json:"preferredTransport,omitzero"`

	// Announcement of additional supported transports. Client can use any of
	// the supported transports.
	AdditionalInterfaces []AgentInterface `json:"additionalInterfaces,omitempty"`

	// A URL to an icon for the agent.
	IconURL string `json:"iconUrl,omitzero"`

	// The service provider of the agent
	Provider *AgentProvider `json:"provider,omitempty"`

	// The version of the agent - format is up to the provider.
	Version string `json:"version"`

	// A URL to documentation for the agent.
	DocumentationURL string `json:"documentationUrl,omitzero"`

	// Optional capabilities supported by the agent.
	Capabilities AgentCapabilities `json:"capabilities"`

	// Security scheme details used for authenticating with this agent.
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`

	// Security requirements for contacting the agent.
	Security []map[string][]string `json:"security,omitempty"`

	// The set of interaction modes that the agent supports across all skills. This
	// can be overridden per-skill.
	// Supported media types for input.
	DefaultInputModes []string `json:"defaultInputModes"`

	// Supported media types for output.
	DefaultOutputModes []string `json:"defaultOutputModes"`

	// Skills are a unit of capability that an agent can perform.
	Skills []AgentSkill `json:"skills"`

	// true if the agent supports providing an extended agent card when the user is
	// authenticated.
	// Defaults to false if not specified.
	SupportsAuthenticatedExtendedCard bool `json:"supportsAuthenticatedExtendedCard,omitzero"`
}

type Task struct {
	// Unique identifier for the task
	ID string `json:"id"`

	// Server-generated id for contextual alignment across interactions
	ContextID string `json:"contextId"`

	// Current status of the task
	Status TaskStatus `json:"status"`

	// History corresponds to the JSON schema field "history".
	History []Message `json:"history,omitempty"`

	// Collection of artifacts created by the agent.
	Artifacts []Artifact `json:"artifacts,omitempty"`

	// Extension metadata.
	Metadata map[string]any `json:"metadata,omitempty"`

	// Event type. Always "task".
	Kind string `json:"kind"`
}

// TaskState and accompanying message.
type TaskStatus struct {
	// State corresponds to the JSON schema field "state".
	State TaskState `json:"state"`

	// Additional status updates for client
	Message *Message `json:"message,omitempty"`

	// ISO 8601 datetime string when the status was recorded.
	Timestamp time.Time `json:"timestamp,omitzero"`
}

// Sent by server during sendStream or subscribe requests
type TaskStatusUpdateEvent struct {
	// Task id
	TaskID string `json:"taskId"`

	// The context the task is associated with
	ContextID string `json:"contextId"`

	// Event type. Always "status-update".
	Kind string `json:"kind"`

	// Current status of the task
	Status TaskStatus `json:"status"`

	// Indicates the end of the event stream
	Final bool `json:"final"`

	// Extension metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Sent by server during sendStream or subscribe requests
type TaskArtifactUpdateEvent struct {
	// Task id
	TaskID string `json:"taskId"`

	// The context the task is associated with
	ContextID string `json:"contextId"`

	// Event type. Always "artifact-update".
	Kind string `json:"kind"`

	// Generated artifact
	Artifact Artifact `json:"artifact"`

	// Indicates if this artifact appends to a previous one
	Append bool `json:"append,omitzero"`

	// Indicates if this is the last chunk of the artifact
	LastChunk bool `json:"lastChunk,omitzero"`

	// Extension metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Parameters containing only a task ID, used for simple task operations.
type TaskIDParams struct {
	// Task id.
	ID string `json:"id"`

	// Metadata corresponds to the JSON schema field "metadata".
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Parameters for querying a task, including optional history length.
type TaskQueryParams struct {
	// Task id.
	ID string `json:"id"`

	// Metadata corresponds to the JSON schema field "metadata".
	Metadata map[string]any `json:"metadata,omitempty"`

	// Number of recent messages to be retrieved.
	HistoryLength int `json:"historyLength,omitzero"`
}

// Parameters for fetching a pushNotificationConfiguration associated with a Task
type GetTaskPushNotificationConfigParams struct {
	// Task id.
	ID string `json:"id"`

	// Metadata corresponds to the JSON schema field "metadata".
	Metadata map[string]any `json:"metadata,omitempty"`

	// PushNotificationConfigID corresponds to the JSON schema field
	// "pushNotificationConfigId".
	PushNotificationConfigID string `json:"pushNotificationConfigId,omitzero"`
}

// Parameters for getting list of pushNotificationConfigurations associated with a
// Task
type ListTaskPushNotificationConfigParams struct {
	// Task id.
	ID string `json:"id"`

	// Metadata corresponds to the JSON schema field "metadata".
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Parameters for removing pushNotificationConfiguration associated with a Task
type DeleteTaskPushNotificationConfigParams struct {
	// Task id.
	ID string `json:"id"`

	// Metadata corresponds to the JSON schema field "metadata".
	Metadata map[string]any `json:"metadata,omitempty"`

	// PushNotificationConfigID corresponds to the JSON schema field
	// "pushNotificationConfigId".
	PushNotificationConfigID string `json:"pushNotificationConfigId"`
}

// Configuration for the send message request.
type MessageSendConfiguration struct {
	// Accepted output modalities by the client.
	AcceptedOutputModes []string `json:"acceptedOutputModes"`

	// Number of recent messages to be retrieved.
	HistoryLength int `json:"historyLength,omitzero"`

	// Where the server should send notifications when disconnected.
	PushNotificationConfig *PushNotificationConfig `json:"pushNotificationConfig,omitempty"`

	// If the server should treat the client as a blocking request.
	Blocking bool `json:"blocking,omitzero"`
}

// Sent by the client to the agent as a request. May create, continue or restart a
// task.
type MessageSendParams struct {
	// The message being sent to the server.
	Message Message `json:"message"`

	// Send message configuration.
	Configuration *MessageSendConfiguration `json:"configuration,omitempty"`

	// Extension metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Represents an artifact generated for a task.
type Artifact struct {
	// Unique identifier for the artifact.
	ArtifactID string `json:"artifactId"`

	// Optional name for the artifact.
	Name string `json:"name,omitzero"`

	// Optional description for the artifact.
	Description string `json:"description,omitzero"`

	// Artifact parts.
	Parts []Part `json:"parts"`

	// Extension metadata.
	Metadata map[string]any `json:"metadata,omitempty"`

	// The URIs of extensions that are present or contributed to this Artifact.
	Extensions []string `json:"extensions,omitempty"`
}

// Represents a single message exchanged between user and agent.
type Message struct {
	// Message sender's role
	Role MessageRole `json:"role"`

	// Message content
	Parts []Part `json:"parts"`

	// Extension metadata.
	Metadata map[string]any `json:"metadata,omitempty"`

	// The URIs of extensions that are present or contributed to this Message.
	Extensions []string `json:"extensions,omitempty"`

	// List of tasks referenced as context by this message.
	ReferenceTaskIds []string `json:"referenceTaskIds,omitempty"`

	// Identifier created by the message creator
	MessageID string `json:"messageId"`

	// Identifier of task the message is related to
	TaskID string `json:"taskId,omitzero"`

	// The context the message is associated with
	ContextID string `json:"contextId,omitzero"`

	// Event type. Always "message"
	Kind string `json:"kind"`
}

type Part interface {
	PartType() PartType
}

// Represents a text segment within parts.
type TextPart struct {
	// Optional metadata associated with the part.
	Metadata map[string]any `json:"metadata,omitempty"`

	// Part type - text for TextParts. Always "text".
	Kind string `json:"kind"`

	// Text content
	Text string `json:"text"`
}

var _ Part = (*TextPart)(nil)

func (*TextPart) PartType() PartType {
	return PartTypeText
}

type File interface {
	isFile()
}

// Define the variant where 'bytes' is present and 'uri' is absent
type FileWithBytes struct {
	// Optional name for the file
	Name string `json:"name,omitzero"`

	// Optional mimeType for the file
	MIMEType string `json:"mimeType,omitzero"`

	// base64 encoded content of the file
	Bytes string `json:"bytes"`
}

var _ File = (*FileWithBytes)(nil)

func (*FileWithBytes) isFile()

// Define the variant where 'uri' is present and 'bytes' is absent
type FileWithURI struct {
	// Optional name for the file
	Name string `json:"name,omitzero"`

	// Optional mimeType for the file
	MIMEType string `json:"mimeType,omitzero"`

	// URL for the File content
	URI string `json:"uri"`
}

var _ File = (*FileWithURI)(nil)

func (*FileWithURI) isFile()

// Represents a File segment within parts.
type FilePart struct {
	// Optional metadata associated with the part.
	Metadata map[string]any `json:"metadata,omitempty"`

	// Part type - file for FileParts. Always "file".
	Kind string `json:"kind"`

	// File content either as url or bytes
	File File `json:"file"`
}

var _ Part = (*FilePart)(nil)

func (*FilePart) PartType() PartType {
	return PartTypeFile
}

// Represents a structured data segment within a message part.
type DataPart struct {
	// Optional metadata associated with the part.
	Metadata map[string]any `json:"metadata,omitempty"`

	// Part type - data for DataParts. Always "data".
	Kind string `json:"kind"`

	// Structured data content
	Data map[string]any `json:"data"`
}

var _ Part = (*DataPart)(nil)

func (*DataPart) PartType() PartType {
	return PartTypeData
}

// Defines authentication details for push notifications.
type PushNotificationAuthenticationInfo struct {
	// Supported authentication schemes - e.g. Basic, Bearer
	Schemes []string `json:"schemes"`

	// Optional credentials
	Credentials string `json:"credentials,omitzero"`
}

// Configuration for setting up push notifications for task updates.
type PushNotificationConfig struct {
	// Push Notification ID - created by server to support multiple callbacks
	ID string `json:"id,omitzero"`

	// URL for sending the push notifications.
	URL string `json:"url"`

	// Token unique to this task/session.
	Token string `json:"token,omitzero"`

	// Authentication corresponds to the JSON schema field "authentication".
	Authentication *PushNotificationAuthenticationInfo `json:"authentication,omitempty"`
}

// Parameters for setting or getting push notification configuration for a task
type TaskPushNotificationConfig struct {
	// Task id.
	TaskID string `json:"taskId"`

	// Push notification configuration.
	PushNotificationConfig PushNotificationConfig `json:"pushNotificationConfig"`
}

type SecurityScheme interface {
	isSecurityScheme()
}

type APIKeySecuritySchemeIn string

const (
	APIKeySecuritySchemeInCookie APIKeySecuritySchemeIn = "cookie"
	APIKeySecuritySchemeInHeader APIKeySecuritySchemeIn = "header"
	APIKeySecuritySchemeInQuery  APIKeySecuritySchemeIn = "query"
)

// API Key security scheme.
type APIKeySecurityScheme struct {
	// Description of this security scheme.
	Description string `json:"description,omitzero"`

	// Type corresponds to the JSON schema field "type". Always "apiKey".
	Type string `json:"type"`

	// The location of the API key. Valid values are "query", "header", or "cookie".
	In APIKeySecuritySchemeIn `json:"in"`

	// The name of the header, query or cookie parameter to be used.
	Name string `json:"name"`
}

var _ SecurityScheme = (*APIKeySecurityScheme)(nil)

func (*APIKeySecurityScheme) isSecurityScheme()

// HTTP Authentication security scheme.
type HTTPAuthSecurityScheme struct {
	// Description of this security scheme.
	Description string `json:"description,omitzero"`

	// Type corresponds to the JSON schema field "type". Always "http".
	Type string `json:"type"`

	// The name of the HTTP Authentication scheme to be used in the Authorization
	// header as defined
	// in RFC7235. The values used SHOULD be registered in the IANA Authentication
	// Scheme registry.
	// The value is case-insensitive, as defined in RFC7235.
	Scheme string `json:"scheme"`

	// A hint to the client to identify how the bearer token is formatted. Bearer
	// tokens are usually
	// generated by an authorization server, so this information is primarily for
	// documentation
	// purposes.
	BearerFormat string `json:"bearerFormat,omitzero"`
}

var _ SecurityScheme = (*HTTPAuthSecurityScheme)(nil)

func (*HTTPAuthSecurityScheme) isSecurityScheme()

// OAuth2.0 security scheme configuration.
type OAuth2SecurityScheme struct {
	// Description of this security scheme.
	Description string `json:"description,omitzero"`

	// Type corresponds to the JSON schema field "type". Always "oauth2".
	Type string `json:"type"`

	// An object containing configuration information for the flow types supported.
	Flows OAuthFlows `json:"flows"`
}

var _ SecurityScheme = (*OAuth2SecurityScheme)(nil)

func (*OAuth2SecurityScheme) isSecurityScheme()

// OpenID Connect security scheme configuration.
type OpenIDConnectSecurityScheme struct {
	// Description of this security scheme.
	Description string `json:"description,omitzero"`

	// Type corresponds to the JSON schema field "type".
	Type string `json:"type"`

	// Well-known URL to discover the [[OpenID-Connect-Discovery]] provider metadata.
	OpenIdConnectURL string `json:"openIdConnectUrl"`
}

var _ SecurityScheme = (*OpenIDConnectSecurityScheme)(nil)

func (*OpenIDConnectSecurityScheme) isSecurityScheme()

// Allows configuration of the supported OAuth Flows
type OAuthFlows struct {
	// Configuration for the OAuth Authorization Code flow. Previously called
	// accessCode in OpenAPI 2.0.
	AuthorizationCode *AuthorizationCodeOAuthFlow `json:"authorizationCode,omitempty"`

	// Configuration for the OAuth Client Credentials flow. Previously called
	// application in OpenAPI 2.0
	ClientCredentials *ClientCredentialsOAuthFlow `json:"clientCredentials,omitempty"`

	// Configuration for the OAuth Implicit flow
	Implicit *ImplicitOAuthFlow `json:"implicit,omitempty"`

	// Configuration for the OAuth Resource Owner Password flow
	Password *PasswordOAuthFlow `json:"password,omitempty"`
}

// Configuration details for a supported OAuth Flow
type AuthorizationCodeOAuthFlow struct {
	// The authorization URL to be used for this flow. This MUST be in the form of a
	// URL. The OAuth2
	// standard requires the use of TLS
	AuthorizationURL string `json:"authorizationUrl"`

	// The token URL to be used for this flow. This MUST be in the form of a URL. The
	// OAuth2 standard
	// requires the use of TLS.
	TokenURL string `json:"tokenUrl"`

	// The URL to be used for obtaining refresh tokens. This MUST be in the form of a
	// URL. The OAuth2
	// standard requires the use of TLS.
	RefreshURL string `json:"refreshUrl,omitzero"`

	// The available scopes for the OAuth2 security scheme. A map between the scope
	// name and a short
	// description for it. The map MAY be empty.
	Scopes map[string]string `json:"scopes"`
}

// Configuration details for a supported OAuth Flow
type ClientCredentialsOAuthFlow struct {
	// The token URL to be used for this flow. This MUST be in the form of a URL. The
	// OAuth2 standard
	// requires the use of TLS.
	TokenURL string `json:"tokenUrl"`

	// The URL to be used for obtaining refresh tokens. This MUST be in the form of a
	// URL. The OAuth2
	// standard requires the use of TLS.
	RefreshURL string `json:"refreshUrl,omitzero"`

	// The available scopes for the OAuth2 security scheme. A map between the scope
	// name and a short
	// description for it. The map MAY be empty.
	Scopes map[string]string `json:"scopes"`
}

// Configuration details for a supported OAuth Flow
type ImplicitOAuthFlow struct {
	// The authorization URL to be used for this flow. This MUST be in the form of a
	// URL. The OAuth2
	// standard requires the use of TLS
	AuthorizationURL string `json:"authorizationUrl"`

	// The URL to be used for obtaining refresh tokens. This MUST be in the form of a
	// URL. The OAuth2
	// standard requires the use of TLS.
	RefreshURL string `json:"refreshUrl,omitzero"`

	// The available scopes for the OAuth2 security scheme. A map between the scope
	// name and a short
	// description for it. The map MAY be empty.
	Scopes map[string]string `json:"scopes"`
}

// Configuration details for a supported OAuth Flow
type PasswordOAuthFlow struct {
	// The token URL to be used for this flow. This MUST be in the form of a URL. The
	// OAuth2 standard
	// requires the use of TLS.
	TokenURL string `json:"tokenUrl"`

	// The URL to be used for obtaining refresh tokens. This MUST be in the form of a
	// URL. The OAuth2
	// standard requires the use of TLS.
	RefreshURL string `json:"refreshUrl,omitzero"`

	// The available scopes for the OAuth2 security scheme. A map between the scope
	// name and a short
	// description for it. The map MAY be empty.
	Scopes map[string]string `json:"scopes"`
}
