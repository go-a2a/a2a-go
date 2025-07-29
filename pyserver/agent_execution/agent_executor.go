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

// Package agent_execution provides the core execution framework for A2A agents.
package agent_execution

import (
	"context"

	"github.com/go-a2a/a2a-go"
)

// AgentExecutor defines the interface that all A2A agents must implement.
// It provides the core methods for handling task execution and cancellation.
type AgentExecutor interface {
	// Execute handles incoming requests and produces responses via the event queue.
	// The method should process the request context and emit appropriate events
	// such as status updates, messages, and artifacts.
	Execute(ctx context.Context, reqCtx *RequestContext, eventQueue EventQueue) error

	// Cancel handles cancellation requests for running tasks.
	// It should attempt to gracefully stop the task execution and update
	// the task status accordingly.
	Cancel(ctx context.Context, reqCtx *RequestContext, eventQueue EventQueue) error
}

// BaseAgentExecutor provides a base implementation of AgentExecutor with common functionality.
// Agents can embed this struct to inherit default implementations.
// Note: Implementers must override the Execute method.
type BaseAgentExecutor struct{}

// Execute must be implemented by the embedding struct.
// This method panics to ensure implementers provide their own implementation.
func (b *BaseAgentExecutor) Execute(ctx context.Context, reqCtx *RequestContext, eventQueue EventQueue) error {
	panic("Execute method must be implemented by the embedding struct")
}

// Cancel provides a default implementation that marks the task as canceled.
func (b *BaseAgentExecutor) Cancel(ctx context.Context, reqCtx *RequestContext, eventQueue EventQueue) error {
	return eventQueue.Put(ctx, &TaskStatusUpdateEvent{
		TaskID: reqCtx.TaskID,
		Status: &TaskStatus{
			State:   a2a.TaskStateCanceled,
			Message: "Task canceled by user request",
		},
	})
}