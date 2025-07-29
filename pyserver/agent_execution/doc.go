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
// It includes interfaces and implementations for handling agent tasks, managing
// event queues, and processing A2A protocol messages.
//
// The package is designed to be compatible with the Python A2A agent execution
// module, providing equivalent functionality in idiomatic Go.
//
// # Core Components
//
// AgentExecutor: The main interface that all agents must implement. It defines
// the Execute and Cancel methods for handling task execution.
//
// RequestContext: Contains all the information about an incoming request,
// including the task ID, message, authentication info, and metadata.
//
// EventQueue: Manages the flow of events during agent execution. Events include
// status updates, messages, and artifacts that are sent back to the client.
//
// ProtocolMessageExecutor: Handles the execution of A2A protocol messages,
// managing the lifecycle of tasks and converting between internal events and
// protocol responses.
//
// # Usage Example
//
//	type MyAgent struct {
//		agent_execution.BaseAgentExecutor
//	}
//
//	func (a *MyAgent) Execute(ctx context.Context, reqCtx *RequestContext, eventQueue EventQueue) error {
//		// Send initial status
//		eventQueue.Put(ctx, CreateWorkingStatusEvent(reqCtx.TaskID, reqCtx.ContextID))
//
//		// Process the request
//		result := processRequest(reqCtx.Message)
//
//		// Send result as artifact
//		artifact := &a2a.Artifact{
//			ArtifactID: uuid.New().String(),
//			Name:       "result",
//			Parts:      []a2a.Part{a2a.NewTextPart(result)},
//		}
//		eventQueue.Put(ctx, NewTaskArtifactUpdateEvent(reqCtx.TaskID, reqCtx.ContextID, artifact))
//
//		// Send completion status
//		eventQueue.Put(ctx, CreateCompletedStatusEvent(reqCtx.TaskID, reqCtx.ContextID))
//
//		return nil
//	}
//
// # Event Types
//
// The package defines several event types that agents can emit:
//
// - TaskStatusUpdateEvent: Updates the task status (working, completed, failed, etc.)
// - TaskArtifactUpdateEvent: Sends artifacts (results) back to the client
// - MessageEvent: Sends messages during task execution
// - FinalEvent: Indicates the end of a streaming response
//
// # Context Building
//
// The package provides context builders for creating RequestContext instances:
//
// - SimpleContextBuilder: Basic context building with validation
// - EnhancedContextBuilder: Advanced context building with validators and enrichers
//
// # Thread Safety
//
// All interfaces and types in this package are designed to be thread-safe.
// The ChannelEventQueue implementation uses channels and mutexes to ensure
// safe concurrent access.
package agent_execution