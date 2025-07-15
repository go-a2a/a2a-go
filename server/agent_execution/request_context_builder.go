// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package agent_execution

import (
	"context"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/server"
)

// RequestContextBuilder defines the interface for building RequestContext instances.
// This is equivalent to Python's RequestContextBuilder abstract base class.
//
// Implementations are responsible for building the RequestContext that is supplied
// to the AgentExecutor.
type RequestContextBuilder interface {
	// Build creates a RequestContext from the provided parameters.
	// This is equivalent to Python's build method.
	//
	// Parameters:
	//   ctx: Go context for cancellation and timeouts
	//   params: The incoming message send parameters
	//   taskID: The ID of the task (may be empty, will be generated if needed)
	//   contextID: The ID of the conversation context (may be empty, will be generated if needed)
	//   currentTask: The existing Task object being processed (may be nil)
	//   callContext: The server call context associated with this request
	//
	// Returns:
	//   *RequestContext: The built request context
	//   error: Any error that occurred during building
	Build(ctx context.Context, params *a2a.MessageSendParams, taskID, contextID string, currentTask *a2a.Task, callContext *server.ServerCallContext) (*RequestContext, error)
}
