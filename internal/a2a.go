// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package a2a provides an open protocol enabling communication and interoperability between opaque agentic applications for Go.
package a2a

// Version is the current version of the A2A protocol.
const Version = "0.1.0"

// APIKeySecurityScheme API Key security scheme.
type APIKeySecurityScheme struct {
	// Description of this security scheme.
	Description string `json:"description,omitzero"`

	// The location of the API key. Valid values are "query", "header", or "cookie".
	In APIKeySecuritySchemeIn `json:"in"`

	// The name of the header, query or cookie parameter to be used.
	Name string `json:"name"`

	// Type corresponds to the JSON schema field "type".
	Type string `json:"type"`
}

type APIKeySecuritySchemeIn string

const (
	APIKeySecuritySchemeInCookie APIKeySecuritySchemeIn = "cookie"
	APIKeySecuritySchemeInHeader APIKeySecuritySchemeIn = "header"
	APIKeySecuritySchemeInQuery  APIKeySecuritySchemeIn = "query"
)

// AgentCapabilities Defines optional capabilities supported by an agent.
type AgentCapabilities struct {
	// true if the agent can notify updates to client.
	PushNotifications bool `json:"pushNotifications,omitzero"`

	// true if the agent exposes status change history for tasks.
	StateTransitionHistory bool `json:"stateTransitionHistory,omitzero"`

	// true if the agent supports SSE.
	Streaming bool `json:"streaming,omitzero"`
}

// AgentCard An AgentCard conveys key information:
// - Overall details (version, name, description, uses)
// - Skills: A set of capabilities the agent can perform
// - Default modalities/content types supported by the agent.
// - Authentication requirements
type AgentCard struct {
	// Optional capabilities supported by the agent.
	Capabilities *AgentCapabilities `json:"capabilities"`

	// The set of interaction modes that the agent supports across all skills. This can be overridden per-skill.
	// Supported media types for input.
	DefaultInputModes []string `json:"defaultInputModes"`

	// Supported media types for output.
	DefaultOutputModes []string `json:"defaultOutputModes"`

	// A human-readable description of the agent. Used to assist users and
	// other agents in understanding what the agent can do.
	Description string `json:"description"`

	// A URL to documentation for the agent.
	DocumentationURL string `json:"documentationUrl,omitzero"`

	// A URL to an icon for the agent.
	IconURL string `json:"iconUrl,omitzero"`

	// Human readable name of the agent.
	Name string `json:"name"`

	// The service provider of the agent
	Provider *AgentProvider `json:"provider,omitzero"`

	// Security requirements for contacting the agent.
	Security []map[string][]string `json:"security,omitzero"`

	// Security scheme details used for authenticating with this agent.
	SecuritySchemes map[string]any `json:"securitySchemes,omitzero"`

	// Skills are a unit of capability that an agent can perform.
	Skills []AgentSkill `json:"skills"`

	// true if the agent supports providing an extended agent card when the user is authenticated.
	// Defaults to false if not specified.
	SupportsAuthenticatedExtendedCard bool `json:"supportsAuthenticatedExtendedCard,omitzero"`

	// A URL to the address the agent is hosted at.
	URL string `json:"url"`

	// The version of the agent - format is up to the provider.
	Version string `json:"version"`
}

// AgentProvider Represents the service provider of an agent.
type AgentProvider struct {
	// Agent provider's organization name.
	Organization string `json:"organization"`

	// Agent provider's URL.
	URL string `json:"url"`
}

// AgentSkill Represents a unit of capability that an agent can perform.
type AgentSkill struct {
	// Description of the skill - will be used by the client or a human
	// as a hint to understand what the skill does.
	Description string `json:"description"`

	// The set of example scenarios that the skill can perform.
	// Will be used by the client as a hint to understand how the skill can be used.
	Examples []string `json:"examples,omitzero"`

	// Unique identifier for the agent's skill.
	ID string `json:"id"`

	// The set of interaction modes that the skill supports
	// (if different than the default).
	// Supported media types for input.
	InputModes []string `json:"inputModes,omitzero"`

	// Human readable name of the skill.
	Name string `json:"name"`

	// Supported media types for output.
	OutputModes []string `json:"outputModes,omitzero"`

	// Set of tagwords describing classes of capabilities for this specific skill.
	Tags []string `json:"tags"`
}

// Artifact Represents an artifact generated for a task.
type Artifact struct {
	// Unique identifier for the artifact.
	ArtifactID string `json:"artifactId"`

	// Optional description for the artifact.
	Description string `json:"description,omitzero"`

	// Extension metadata.
	Metadata map[string]any `json:"metadata,omitzero"`

	// Optional name for the artifact.
	Name string `json:"name,omitzero"`

	// Artifact parts.
	Parts []any `json:"parts"`
}

// AuthorizationCodeOAuthFlow Configuration details for a supported OAuth Flow
type AuthorizationCodeOAuthFlow struct {
	// The authorization URL to be used for this flow. This MUST be in the form of a URL. The OAuth2
	// standard requires the use of TLS
	AuthorizationURL string `json:"authorizationUrl"`

	// The URL to be used for obtaining refresh tokens. This MUST be in the form of a URL. The OAuth2
	// standard requires the use of TLS.
	RefreshURL string `json:"refreshUrl,omitzero"`

	// The available scopes for the OAuth2 security scheme. A map between the scope name and a short
	// description for it. The map MAY be empty.
	Scopes map[string]string `json:"scopes"`

	// The token URL to be used for this flow. This MUST be in the form of a URL. The OAuth2 standard
	// requires the use of TLS.
	TokenURL string `json:"tokenUrl"`
}

// CancelTaskRequest JSON-RPC request model for the 'tasks/cancel' method.
type CancelTaskRequest struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// A Structured value that holds the parameter values to be used during the invocation of the method.
	Params *TaskIDParams `json:"params"`
}

// CancelTaskSuccessResponse JSON-RPC success response model for the 'tasks/cancel' method.
type CancelTaskSuccessResponse struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// The result object on success.
	Result *Task `json:"result"`
}

// ClientCredentialsOAuthFlow Configuration details for a supported OAuth Flow
type ClientCredentialsOAuthFlow struct {
	// The URL to be used for obtaining refresh tokens. This MUST be in the form of a URL. The OAuth2
	// standard requires the use of TLS.
	RefreshURL string `json:"refreshUrl,omitzero"`

	// The available scopes for the OAuth2 security scheme. A map between the scope name and a short
	// description for it. The map MAY be empty.
	Scopes map[string]string `json:"scopes"`

	// The token URL to be used for this flow. This MUST be in the form of a URL. The OAuth2 standard
	// requires the use of TLS.
	TokenURL string `json:"tokenUrl"`
}

// ContentTypeNotSupportedError A2A specific error indicating incompatible content types between request and agent capabilities.
type ContentTypeNotSupportedError struct {
	// A Number that indicates the error type that occurred.
	Code int64 `json:"code"`

	// A Primitive or Structured value that contains additional information about the error.
	// This may be omitted.
	Data any `json:"data,omitzero"`

	// A String providing a short description of the error.
	Message string `json:"message"`
}

// DataPart Represents a structured data segment within a message part.
type DataPart struct {
	// Structured data content
	Data map[string]any `json:"data"`

	// Part type - data for DataParts
	Kind string `json:"kind"`

	// Optional metadata associated with the part.
	Metadata map[string]any `json:"metadata,omitzero"`
}

// FileBase Represents the base entity for FileParts
type FileBase struct {
	// Optional mimeType for the file
	MimeType string `json:"mimeType,omitzero"`

	// Optional name for the file
	Name string `json:"name,omitzero"`
}

// FilePart Represents a File segment within parts.
type FilePart struct {
	// File content either as url or bytes
	File any `json:"file"`

	// Part type - file for FileParts
	Kind string `json:"kind"`

	// Optional metadata associated with the part.
	Metadata map[string]any `json:"metadata,omitzero"`
}

// FileWithBytes Define the variant where 'bytes' is present and 'uri' is absent
type FileWithBytes struct {
	// base64 encoded content of the file
	Bytes string `json:"bytes"`

	// Optional mimeType for the file
	MimeType string `json:"mimeType,omitzero"`

	// Optional name for the file
	Name string `json:"name,omitzero"`
}

// FileWithUri Define the variant where 'uri' is present and 'bytes' is absent
type FileWithUri struct {
	// Optional mimeType for the file
	MimeType string `json:"mimeType,omitzero"`

	// Optional name for the file
	Name string `json:"name,omitzero"`

	// URL for the File content
	URI string `json:"uri"`
}

// GetTaskPushNotificationConfigRequest JSON-RPC request model for the 'tasks/pushNotificationConfig/get' method.
type GetTaskPushNotificationConfigRequest struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// A Structured value that holds the parameter values to be used during the invocation of the method.
	Params *TaskIDParams `json:"params"`
}

// GetTaskPushNotificationConfigSuccessResponse JSON-RPC success response model for the 'tasks/pushNotificationConfig/get' method.
type GetTaskPushNotificationConfigSuccessResponse struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// The result object on success.
	Result *TaskPushNotificationConfig `json:"result"`
}

// GetTaskRequest JSON-RPC request model for the 'tasks/get' method.
type GetTaskRequest struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// A Structured value that holds the parameter values to be used during the invocation of the method.
	Params *TaskQueryParams `json:"params"`
}

// GetTaskSuccessResponse JSON-RPC success response for the 'tasks/get' method.
type GetTaskSuccessResponse struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// The result object on success.
	Result *Task `json:"result"`
}

// HTTPAuthSecurityScheme HTTP Authentication security scheme.
type HTTPAuthSecurityScheme struct {
	// A hint to the client to identify how the bearer token is formatted. Bearer tokens are usually
	// generated by an authorization server, so this information is primarily for documentation
	// purposes.
	BearerFormat string `json:"bearerFormat,omitzero"`

	// Description of this security scheme.
	Description string `json:"description,omitzero"`

	// The name of the HTTP Authentication scheme to be used in the Authorization header as defined
	// in RFC7235. The values used SHOULD be registered in the IANA Authentication Scheme registry.
	// The value is case-insensitive, as defined in RFC7235.
	Scheme string `json:"scheme"`

	// Type corresponds to the JSON schema field "type".
	Type string `json:"type"`
}

// ImplicitOAuthFlow Configuration details for a supported OAuth Flow
type ImplicitOAuthFlow struct {
	// The authorization URL to be used for this flow. This MUST be in the form of a URL. The OAuth2
	// standard requires the use of TLS
	AuthorizationURL string `json:"authorizationUrl"`

	// The URL to be used for obtaining refresh tokens. This MUST be in the form of a URL. The OAuth2
	// standard requires the use of TLS.
	RefreshURL string `json:"refreshUrl,omitzero"`

	// The available scopes for the OAuth2 security scheme. A map between the scope name and a short
	// description for it. The map MAY be empty.
	Scopes map[string]string `json:"scopes"`
}

// InternalError JSON-RPC error indicating an internal JSON-RPC error on the server.
type InternalError struct {
	// A Number that indicates the error type that occurred.
	Code int64 `json:"code"`

	// A Primitive or Structured value that contains additional information about the error.
	// This may be omitted.
	Data any `json:"data,omitzero"`

	// A String providing a short description of the error.
	Message string `json:"message"`
}

// InvalidAgentResponseError A2A specific error indicating agent returned invalid response for the current method
type InvalidAgentResponseError struct {
	// A Number that indicates the error type that occurred.
	Code int64 `json:"code"`

	// A Primitive or Structured value that contains additional information about the error.
	// This may be omitted.
	Data any `json:"data,omitzero"`

	// A String providing a short description of the error.
	Message string `json:"message"`
}

// InvalidParamsError JSON-RPC error indicating invalid method parameter(s).
type InvalidParamsError struct {
	// A Number that indicates the error type that occurred.
	Code int64 `json:"code"`

	// A Primitive or Structured value that contains additional information about the error.
	// This may be omitted.
	Data any `json:"data,omitzero"`

	// A String providing a short description of the error.
	Message string `json:"message"`
}

// InvalidRequestError JSON-RPC error indicating the JSON sent is not a valid Request object.
type InvalidRequestError struct {
	// A Number that indicates the error type that occurred.
	Code int64 `json:"code"`

	// A Primitive or Structured value that contains additional information about the error.
	// This may be omitted.
	Data any `json:"data,omitzero"`

	// A String providing a short description of the error.
	Message string `json:"message"`
}

// JSONParseError JSON-RPC error indicating invalid JSON was received by the server.
type JSONParseError struct {
	// A Number that indicates the error type that occurred.
	Code int64 `json:"code"`

	// A Primitive or Structured value that contains additional information about the error.
	// This may be omitted.
	Data any `json:"data,omitzero"`

	// A String providing a short description of the error.
	Message string `json:"message"`
}

// JSONRPCError Represents a JSON-RPC 2.0 Error object.
// This is typically included in a JSONRPCErrorResponse when an error occurs.
type JSONRPCError struct {
	// A Number that indicates the error type that occurred.
	Code int64 `json:"code"`

	// A Primitive or Structured value that contains additional information about the error.
	// This may be omitted.
	Data any `json:"data,omitzero"`

	// A String providing a short description of the error.
	Message string `json:"message"`
}

// JSONRPCErrorResponse Represents a JSON-RPC 2.0 Error Response object.
type JSONRPCErrorResponse struct {
	Error any `json:"error"`

	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`
}

// JSONRPCMessage Base interface for any JSON-RPC 2.0 request or response.
type JSONRPCMessage struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id,omitzero"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`
}

// JSONRPCRequest Represents a JSON-RPC 2.0 Request object.
type JSONRPCRequest struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id,omitzero"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// A Structured value that holds the parameter values to be used during the invocation of the method.
	Params map[string]any `json:"params,omitzero"`
}

// JSONRPCSuccessResponse Represents a JSON-RPC 2.0 Success Response object.
type JSONRPCSuccessResponse struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// The result object on success
	Result any `json:"result"`
}

// Message Represents a single message exchanged between user and agent.
type Message struct {
	// The context the message is associated with
	ContextID string `json:"contextId,omitzero"`

	// Event type
	Kind string `json:"kind"`

	// Identifier created by the message creator
	MessageID string `json:"messageId"`

	// Extension metadata.
	Metadata map[string]any `json:"metadata,omitzero"`

	// Message content
	Parts []any `json:"parts"`

	// List of tasks referenced as context by this message.
	ReferenceTaskIDs []string `json:"referenceTaskIds,omitzero"`

	// Message sender's role
	Role MessageRole `json:"role"`

	// Identifier of task the message is related to
	TaskID string `json:"taskId,omitzero"`
}

type MessageRole string

const (
	MessageRoleAgent MessageRole = "agent"
	MessageRoleUser  MessageRole = "user"
)

// MessageSendConfiguration Configuration for the send message request.
type MessageSendConfiguration struct {
	// Accepted output modalities by the client.
	AcceptedOutputModes []string `json:"acceptedOutputModes"`

	// If the server should treat the client as a blocking request.
	Blocking bool `json:"blocking,omitzero"`

	// Number of recent messages to be retrieved.
	HistoryLength int64 `json:"historyLength,omitzero"`

	// Where the server should send notifications when disconnected.
	PushNotificationConfig *PushNotificationConfig `json:"pushNotificationConfig,omitzero"`
}

// MessageSendParams Sent by the client to the agent as a request. May create, continue or restart a task.
type MessageSendParams struct {
	// Send message configuration.
	Configuration *MessageSendConfiguration `json:"configuration,omitzero"`

	// The message being sent to the server.
	Message *Message `json:"message"`

	// Extension metadata.
	Metadata map[string]any `json:"metadata,omitzero"`
}

// MethodNotFoundError JSON-RPC error indicating the method does not exist or is not available.
type MethodNotFoundError struct {
	// A Number that indicates the error type that occurred.
	Code int64 `json:"code"`

	// A Primitive or Structured value that contains additional information about the error.
	// This may be omitted.
	Data any `json:"data,omitzero"`

	// A String providing a short description of the error.
	Message string `json:"message"`
}

// OAuth2SecurityScheme OAuth2.0 security scheme configuration.
type OAuth2SecurityScheme struct {
	// Description of this security scheme.
	Description string `json:"description,omitzero"`

	// An object containing configuration information for the flow types supported.
	Flows *OAuthFlows `json:"flows"`

	// Type corresponds to the JSON schema field "type".
	Type string `json:"type"`
}

// OAuthFlows Allows configuration of the supported OAuth Flows
type OAuthFlows struct {
	// Configuration for the OAuth Authorization Code flow. Previously called accessCode in OpenAPI 2.0.
	AuthorizationCode *AuthorizationCodeOAuthFlow `json:"authorizationCode,omitzero"`

	// Configuration for the OAuth Client Credentials flow. Previously called application in OpenAPI 2.0
	ClientCredentials *ClientCredentialsOAuthFlow `json:"clientCredentials,omitzero"`

	// Configuration for the OAuth Implicit flow
	Implicit *ImplicitOAuthFlow `json:"implicit,omitzero"`

	// Configuration for the OAuth Resource Owner Password flow
	Password *PasswordOAuthFlow `json:"password,omitzero"`
}

// OpenIDConnectSecurityScheme OpenID Connect security scheme configuration.
type OpenIDConnectSecurityScheme struct {
	// Description of this security scheme.
	Description string `json:"description,omitzero"`

	// Well-known URL to discover the [[OpenID-Connect-Discovery]] provider metadata.
	OpenIDConnectURL string `json:"openIdConnectUrl"`

	// Type corresponds to the JSON schema field "type".
	Type string `json:"type"`
}

// PartBase Base properties common to all message parts.
type PartBase struct {
	// Optional metadata associated with the part.
	Metadata map[string]any `json:"metadata,omitzero"`
}

// PasswordOAuthFlow Configuration details for a supported OAuth Flow
type PasswordOAuthFlow struct {
	// The URL to be used for obtaining refresh tokens. This MUST be in the form of a URL. The OAuth2
	// standard requires the use of TLS.
	RefreshURL string `json:"refreshUrl,omitzero"`

	// The available scopes for the OAuth2 security scheme. A map between the scope name and a short
	// description for it. The map MAY be empty.
	Scopes map[string]string `json:"scopes"`

	// The token URL to be used for this flow. This MUST be in the form of a URL. The OAuth2 standard
	// requires the use of TLS.
	TokenURL string `json:"tokenUrl"`
}

// PushNotificationAuthenticationInfo Defines authentication details for push notifications.
type PushNotificationAuthenticationInfo struct {
	// Optional credentials
	Credentials string `json:"credentials,omitzero"`

	// Supported authentication schemes - e.g. Basic, Bearer
	Schemes []string `json:"schemes"`
}

// PushNotificationConfig Configuration for setting up push notifications for task updates.
type PushNotificationConfig struct {
	Authentication *PushNotificationAuthenticationInfo `json:"authentication,omitzero"`

	// Push Notification ID - created by server to support multiple callbacks
	ID string `json:"id,omitzero"`

	// Token unique to this task/session.
	Token string `json:"token,omitzero"`

	// URL for sending the push notifications.
	URL string `json:"url"`
}

// PushNotificationNotSupportedError A2A specific error indicating the agent does not support push notifications.
type PushNotificationNotSupportedError struct {
	// A Number that indicates the error type that occurred.
	Code int64 `json:"code"`

	// A Primitive or Structured value that contains additional information about the error.
	// This may be omitted.
	Data any `json:"data,omitzero"`

	// A String providing a short description of the error.
	Message string `json:"message"`
}

// SecuritySchemeBase Base properties shared by all security schemes.
type SecuritySchemeBase struct {
	// Description of this security scheme.
	Description string `json:"description,omitzero"`
}

// SendMessageRequest JSON-RPC request model for the 'message/send' method.
type SendMessageRequest struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// A Structured value that holds the parameter values to be used during the invocation of the method.
	Params *MessageSendParams `json:"params"`
}

// SendMessageSuccessResponse JSON-RPC success response model for the 'message/send' method.
type SendMessageSuccessResponse struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// The result object on success
	Result any `json:"result"`
}

// SendStreamingMessageRequest JSON-RPC request model for the 'message/stream' method.
type SendStreamingMessageRequest struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// A Structured value that holds the parameter values to be used during the invocation of the method.
	Params *MessageSendParams `json:"params"`
}

// SendStreamingMessageSuccessResponse JSON-RPC success response model for the 'message/stream' method.
type SendStreamingMessageSuccessResponse struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// The result object on success
	Result any `json:"result"`
}

// SetTaskPushNotificationConfigRequest JSON-RPC request model for the 'tasks/pushNotificationConfig/set' method.
type SetTaskPushNotificationConfigRequest struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// A Structured value that holds the parameter values to be used during the invocation of the method.
	Params *TaskPushNotificationConfig `json:"params"`
}

// SetTaskPushNotificationConfigSuccessResponse JSON-RPC success response model for the 'tasks/pushNotificationConfig/set' method.
type SetTaskPushNotificationConfigSuccessResponse struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// The result object on success.
	Result *TaskPushNotificationConfig `json:"result"`
}

// Task
type Task struct {
	// Collection of artifacts created by the agent.
	Artifacts []Artifact `json:"artifacts,omitzero"`

	// Server-generated id for contextual alignment across interactions
	ContextID string `json:"contextId"`

	// History corresponds to the JSON schema field "history".
	History []Message `json:"history,omitzero"`

	// Unique identifier for the task
	ID string `json:"id"`

	// Event type
	Kind string `json:"kind"`

	// Extension metadata.
	Metadata map[string]any `json:"metadata,omitzero"`

	// Current status of the task
	Status *TaskStatus `json:"status"`
}

// TaskArtifactUpdateEvent Sent by server during sendStream or subscribe requests
type TaskArtifactUpdateEvent struct {
	// Indicates if this artifact appends to a previous one
	Append bool `json:"append,omitzero"`

	// Generated artifact
	Artifact *Artifact `json:"artifact"`

	// The context the task is associated with
	ContextID string `json:"contextId"`

	// Event type
	Kind string `json:"kind"`

	// Indicates if this is the last chunk of the artifact
	LastChunk bool `json:"lastChunk,omitzero"`

	// Extension metadata.
	Metadata map[string]any `json:"metadata,omitzero"`

	// Task id
	TaskID string `json:"taskId"`
}

// TaskIDParams Parameters containing only a task ID, used for simple task operations.
type TaskIDParams struct {
	// Task id.
	ID string `json:"id"`

	// Metadata corresponds to the JSON schema field "metadata".
	Metadata map[string]any `json:"metadata,omitzero"`
}

// TaskNotCancelableError A2A specific error indicating the task is in a state where it cannot be canceled.
type TaskNotCancelableError struct {
	// A Number that indicates the error type that occurred.
	Code int64 `json:"code"`

	// A Primitive or Structured value that contains additional information about the error.
	// This may be omitted.
	Data any `json:"data,omitzero"`

	// A String providing a short description of the error.
	Message string `json:"message"`
}

// TaskNotFoundError A2A specific error indicating the requested task ID was not found.
type TaskNotFoundError struct {
	// A Number that indicates the error type that occurred.
	Code int64 `json:"code"`

	// A Primitive or Structured value that contains additional information about the error.
	// This may be omitted.
	Data any `json:"data,omitzero"`

	// A String providing a short description of the error.
	Message string `json:"message"`
}

// TaskPushNotificationConfig Parameters for setting or getting push notification configuration for a task
type TaskPushNotificationConfig struct {
	// Push notification configuration.
	PushNotificationConfig *PushNotificationConfig `json:"pushNotificationConfig"`

	// Task id.
	TaskID string `json:"taskId"`
}

// TaskQueryParams Parameters for querying a task, including optional history length.
type TaskQueryParams struct {
	// Number of recent messages to be retrieved.
	HistoryLength int64 `json:"historyLength,omitzero"`

	// Task id.
	ID string `json:"id"`

	// Metadata corresponds to the JSON schema field "metadata".
	Metadata map[string]any `json:"metadata,omitzero"`
}

// TaskResubscriptionRequest JSON-RPC request model for the 'tasks/resubscribe' method.
type TaskResubscriptionRequest struct {
	// An identifier established by the Client that MUST contain a String, Number.
	// Numbers SHOULD NOT contain fractional parts.
	ID any `json:"id"`

	// Specifies the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// A Structured value that holds the parameter values to be used during the invocation of the method.
	Params *TaskIDParams `json:"params"`
}

// TaskStatus TaskState and accompanying message.
type TaskStatus struct {
	// Additional status updates for client
	Message *Message `json:"message,omitzero"`

	// State corresponds to the JSON schema field "state".
	State TaskState `json:"state"`

	// ISO 8601 datetime string when the status was recorded.
	Timestamp string `json:"timestamp,omitzero"`
}

type TaskState string

const (
	TaskStateUnknown       TaskState = "unknown"
	TaskStateFailed        TaskState = "failed"
	TaskStateCanceled      TaskState = "canceled"
	TaskStateCompleted     TaskState = "completed"
	TaskStateInputRequired TaskState = "input-required"
	TaskStateWorking       TaskState = "working"
	TaskStateSubmitted     TaskState = "submitted"
	TaskStateRejected      TaskState = "rejected"
	TaskStateAuthRequired  TaskState = "auth-required"
)

// TaskStatusUpdateEvent Sent by server during sendStream or subscribe requests
type TaskStatusUpdateEvent struct {
	// The context the task is associated with
	ContextID string `json:"contextId"`

	// Indicates the end of the event stream
	Final bool `json:"final"`

	// Event type
	Kind string `json:"kind"`

	// Extension metadata.
	Metadata map[string]any `json:"metadata,omitzero"`

	// Current status of the task
	Status *TaskStatus `json:"status"`

	// Task id
	TaskID string `json:"taskId"`
}

// TextPart Represents a text segment within parts.
type TextPart struct {
	// Part type - text for TextParts
	Kind string `json:"kind"`

	// Optional metadata associated with the part.
	Metadata map[string]any `json:"metadata,omitzero"`

	// Text content
	Text string `json:"text"`
}

// UnsupportedOperationError A2A specific error indicating the requested operation is not supported by the agent.
type UnsupportedOperationError struct {
	// A Number that indicates the error type that occurred.
	Code int64 `json:"code"`

	// A Primitive or Structured value that contains additional information about the error.
	// This may be omitted.
	Data any `json:"data,omitzero"`

	// A String providing a short description of the error.
	Message string `json:"message"`
}
