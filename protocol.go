// Copyright 2025 The Go MCP SDK Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package a2a

// Protocol types for version 2025-06-18.

type CancelledParams struct {
	// This property is reserved by the protocol to allow clients and servers to
	// attach additional metadata to their responses.
	Meta `json:"_meta,omitempty"`
	// An optional string describing the reason for the cancellation. This may be
	// logged or presented to the user.
	Reason string `json:"reason,omitempty"`
	// The ID of the request to cancel.
	//
	// This must correspond to the ID of a request previously issued in the same
	// direction.
	RequestID any `json:"requestId"`
}

func (x *CancelledParams) GetProgressToken() any  { return getProgressToken(x) }
func (x *CancelledParams) SetProgressToken(t any) { setProgressToken(x, t) }

// Capabilities a client may support. Known capabilities are defined here, in
// this schema, but this is not a closed set: any client can define its own,
// additional capabilities.
type ClientCapabilities struct {
	// Experimental, non-standard capabilities that the client supports.
	Experimental map[string]struct{} `json:"experimental,omitempty"`
	// Present if the client supports listing roots.
	Roots struct {
		// Whether the client supports notifications for changes to the roots list.
		ListChanged bool `json:"listChanged,omitempty"`
	} `json:"roots,omitempty"`
	// Present if the client supports sampling from an LLM.
	Sampling *SamplingCapabilities `json:"sampling,omitempty"`
	// Present if the client supports elicitation from the server.
	Elicitation *ElicitationCapabilities `json:"elicitation,omitempty"`
}

type InitializeParams struct {
	// This property is reserved by the protocol to allow clients and servers to
	// attach additional metadata to their responses.
	Meta         `json:"_meta,omitempty"`
	Capabilities *ClientCapabilities `json:"capabilities"`
	ClientInfo   *Implementation     `json:"clientInfo"`
	// The latest version of the Model Context Protocol that the client supports.
	// The client may decide to support older versions as well.
	ProtocolVersion string `json:"protocolVersion"`
}

func (x *InitializeParams) GetProgressToken() any  { return getProgressToken(x) }
func (x *InitializeParams) SetProgressToken(t any) { setProgressToken(x, t) }

// After receiving an initialize request from the client, the server sends this
// response.
type InitializeResult struct {
	// This property is reserved by the protocol to allow clients and servers to
	// attach additional metadata to their responses.
	Meta         `json:"_meta,omitempty"`
	Capabilities *serverCapabilities `json:"capabilities"`
	// Instructions describing how to use the server and its features.
	//
	// This can be used by clients to improve the LLM's understanding of available
	// tools, resources, etc. It can be thought of like a "hint" to the model. For
	// example, this information may be added to the system prompt.
	Instructions string `json:"instructions,omitempty"`
	// The version of the Model Context Protocol that the server wants to use. This
	// may not match the version that the client requested. If the client cannot
	// support this version, it must disconnect.
	ProtocolVersion string          `json:"protocolVersion"`
	ServerInfo      *Implementation `json:"serverInfo"`
}

type InitializedParams struct {
	// This property is reserved by the protocol to allow clients and servers to
	// attach additional metadata to their responses.
	Meta `json:"_meta,omitempty"`
}

func (x *InitializedParams) GetProgressToken() any  { return getProgressToken(x) }
func (x *InitializedParams) SetProgressToken(t any) { setProgressToken(x, t) }

// The severity of a log message.
//
// These map to syslog message severities, as specified in RFC-5424:
// https://datatracker.ietf.org/doc/html/rfc5424#section-6.2.1
type LoggingLevel string

type LoggingMessageParams struct {
	// This property is reserved by the protocol to allow clients and servers to
	// attach additional metadata to their responses.
	Meta `json:"_meta,omitempty"`
	// The data to be logged, such as a string message or an object. Any JSON
	// serializable type is allowed here.
	Data any `json:"data"`
	// The severity of this log message.
	Level LoggingLevel `json:"level"`
	// An optional name of the logger issuing this message.
	Logger string `json:"logger,omitempty"`
}

func (x *LoggingMessageParams) GetProgressToken() any  { return getProgressToken(x) }
func (x *LoggingMessageParams) SetProgressToken(t any) { setProgressToken(x, t) }

type PingParams struct {
	// This property is reserved by the protocol to allow clients and servers to
	// attach additional metadata to their responses.
	Meta `json:"_meta,omitempty"`
}

func (x *PingParams) GetProgressToken() any  { return getProgressToken(x) }
func (x *PingParams) SetProgressToken(t any) { setProgressToken(x, t) }

// SamplingCapabilities describes the capabilities for sampling.
type SamplingCapabilities struct{}

// ElicitationCapabilities describes the capabilities for elicitation.
type ElicitationCapabilities struct{}

// An Implementation describes the name and version of an MCP implementation, with an optional
// title for UI representation.
type Implementation struct {
	// Intended for programmatic or logical use, but used as a display name in past
	// specs or fallback (if title isn't present).
	Name string `json:"name"`
	// Intended for UI and end-user contexts — optimized to be human-readable and
	// easily understood, even by those unfamiliar with domain-specific terminology.
	Title   string `json:"title,omitempty"`
	Version string `json:"version"`
}

// Present if the server supports argument autocompletion suggestions.
type completionCapabilities struct{}

// Present if the server supports sending log messages to the client.
type loggingCapabilities struct{}

// Present if the server offers any prompt templates.
type promptCapabilities struct {
	// Whether this server supports notifications for changes to the prompt list.
	ListChanged bool `json:"listChanged,omitempty"`
}

// Present if the server offers any resources to read.
type resourceCapabilities struct {
	// Whether this server supports notifications for changes to the resource list.
	ListChanged bool `json:"listChanged,omitempty"`
	// Whether this server supports subscribing to resource updates.
	Subscribe bool `json:"subscribe,omitempty"`
}

// Capabilities that a server may support. Known capabilities are defined here,
// in this schema, but this is not a closed set: any server can define its own,
// additional capabilities.
type serverCapabilities struct {
	// Present if the server supports argument autocompletion suggestions.
	Completions *completionCapabilities `json:"completions,omitempty"`
	// Experimental, non-standard capabilities that the server supports.
	Experimental map[string]struct{} `json:"experimental,omitempty"`
	// Present if the server supports sending log messages to the client.
	Logging *loggingCapabilities `json:"logging,omitempty"`
	// Present if the server offers any prompt templates.
	Prompts *promptCapabilities `json:"prompts,omitempty"`
	// Present if the server offers any resources to read.
	Resources *resourceCapabilities `json:"resources,omitempty"`
	// Present if the server offers any tools to call.
	Tools *toolCapabilities `json:"tools,omitempty"`
}

// Present if the server offers any tools to call.
type toolCapabilities struct {
	// Whether this server supports notifications for changes to the tool list.
	ListChanged bool `json:"listChanged,omitempty"`
}

const (
	methodCallTool                  = "tools/call"
	notificationCancelled           = "notifications/cancelled"
	methodComplete                  = "completion/complete"
	methodCreateMessage             = "sampling/createMessage"
	methodElicit                    = "elicitation/create"
	methodGetPrompt                 = "prompts/get"
	methodInitialize                = "initialize"
	notificationInitialized         = "notifications/initialized"
	methodListPrompts               = "prompts/list"
	methodListResourceTemplates     = "resources/templates/list"
	methodListResources             = "resources/list"
	methodListRoots                 = "roots/list"
	methodListTools                 = "tools/list"
	notificationLoggingMessage      = "notifications/message"
	methodPing                      = "ping"
	notificationProgress            = "notifications/progress"
	notificationPromptListChanged   = "notifications/prompts/list_changed"
	methodReadResource              = "resources/read"
	notificationResourceListChanged = "notifications/resources/list_changed"
	notificationResourceUpdated     = "notifications/resources/updated"
	notificationRootsListChanged    = "notifications/roots/list_changed"
	methodSetLevel                  = "logging/setLevel"
	methodSubscribe                 = "resources/subscribe"
	notificationToolListChanged     = "notifications/tools/list_changed"
	methodUnsubscribe               = "resources/unsubscribe"
)
