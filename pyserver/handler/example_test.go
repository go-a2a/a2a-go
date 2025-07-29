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

package handler_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/pyserver/agent_execution"
	"github.com/go-a2a/a2a-go/pyserver/handler"
)

// ExampleHandler demonstrates basic handler usage.
func ExampleHandler() {
	// Create a simple handler
	h := &echoHandler{}

	// Create a router
	router := handler.NewRouter()

	// Register the handler
	router.Register(h, a2a.MethodMessageSend)

	// Create HTTP handler
	httpHandler := handler.NewHTTPHandler(router)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", httpHandler))
}

// ExampleRouter demonstrates router with middleware.
func ExampleRouter() {
	// Create router with middleware
	router := handler.NewRouter()
	router.Use(
		handler.RecoveryMiddleware(),
		handler.LoggingMiddleware(),
		handler.TimeoutMiddleware(30 * time.Second),
		handler.ValidationMiddleware(),
		handler.AuthenticationMiddleware(true), // optional auth
	)

	// Register handlers
	router.Register(&echoHandler{}, a2a.MethodMessageSend)
	router.RegisterStreaming(&streamingEchoHandler{}, a2a.MethodMessageStream)

	// Create HTTP handler
	httpHandler := handler.NewHTTPHandler(router)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", httpHandler))
}

// ExampleSSEHandler demonstrates SSE streaming.
func ExampleSSEHandler() {
	// Create router
	router := handler.NewRouter()

	// Register streaming handler
	router.RegisterStreaming(&streamingEchoHandler{}, a2a.MethodMessageStream)

	// Create SSE handler
	sseHandler := handler.NewSSEHandler(router)

	// Register routes
	http.Handle("/message/stream", sseHandler)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// ExampleWebSocketHandler demonstrates WebSocket usage.
func ExampleWebSocketHandler() {
	// Create router
	router := handler.NewRouter()

	// Register handlers
	router.Register(&echoHandler{}, a2a.MethodMessageSend)
	router.RegisterStreaming(&streamingEchoHandler{}, a2a.MethodMessageStream)

	// Create WebSocket handler
	wsHandler := handler.NewWebSocketHandler(router)

	// Register route
	http.Handle("/ws", wsHandler)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// ExampleAgentHandler demonstrates wrapping an AgentExecutor.
func ExampleAgentHandler() {
	// Create an agent executor
	executor := &myAgentExecutor{}

	// Wrap it as a handler
	h := handler.NewAgentHandler(executor)

	// Create router and register
	router := handler.NewRouter()
	router.Register(h, h.SupportedMethods()...)

	// Create HTTP handler
	httpHandler := handler.NewHTTPHandler(router)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", httpHandler))
}

// echoHandler is a simple handler that echoes messages.
type echoHandler struct{}

func (h *echoHandler) Handle(ctx context.Context, method a2a.Method, params a2a.Params) (a2a.Result, error) {
	switch method {
	case a2a.MethodMessageSend:
		msg := params.(*a2a.MessageSendParams)
		// Echo the message back
		return &a2a.Message{
			Kind:      a2a.MessageEventKind,
			MessageID: "echo-" + msg.Message.MessageID,
			Role:      a2a.RoleAgent,
			Parts:     msg.Message.Parts,
			TaskID:    "task-" + time.Now().Format("20060102150405"),
		}, nil
	default:
		return nil, handler.ErrMethodNotFound
	}
}

func (h *echoHandler) SupportedMethods() []a2a.Method {
	return []a2a.Method{a2a.MethodMessageSend}
}

// streamingEchoHandler demonstrates streaming responses.
type streamingEchoHandler struct{}

func (h *streamingEchoHandler) Handle(ctx context.Context, method a2a.Method, params a2a.Params) (a2a.Result, error) {
	// For non-streaming calls, return a task
	return &a2a.Task{
		Kind:      a2a.TaskEventKind,
		ID:        "task-" + time.Now().Format("20060102150405"),
		ContextID: "ctx-123",
		Status: a2a.TaskStatus{
			State: a2a.TaskStateWorking,
		},
	}, nil
}

func (h *streamingEchoHandler) SupportedMethods() []a2a.Method {
	return []a2a.Method{a2a.MethodMessageStream}
}

func (h *streamingEchoHandler) HandleStream(ctx context.Context, method a2a.Method, params a2a.Params) (<-chan a2a.SendStreamingMessageResponse, error) {
	ch := make(chan a2a.SendStreamingMessageResponse)

	go func() {
		defer close(ch)

		taskID := "task-" + time.Now().Format("20060102150405")

		// Send initial status
		select {
		case ch <- &a2a.TaskStatusUpdateEvent{
			Kind:      a2a.StatusUpdateEventKind,
			TaskID:    taskID,
			ContextID: "ctx-123",
			Status: a2a.TaskStatus{
				State: a2a.TaskStateWorking,
			},
		}:
		case <-ctx.Done():
			return
		}

		// Simulate work with periodic updates
		for i := 0; i < 5; i++ {
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
				// Send message update
				ch <- &a2a.Message{
					Kind:      a2a.MessageEventKind,
					MessageID: fmt.Sprintf("msg-%d", i),
					Role:      a2a.RoleAgent,
					Parts: []a2a.Part{
						a2a.NewTextPart(fmt.Sprintf("Processing step %d", i+1)),
					},
					TaskID: taskID,
				}
			}
		}

		// Send completion
		ch <- &a2a.TaskStatusUpdateEvent{
			Kind:      a2a.StatusUpdateEventKind,
			TaskID:    taskID,
			ContextID: "ctx-123",
			Final:     true,
			Status: a2a.TaskStatus{
				State: a2a.TaskStateCompleted,
			},
		}
	}()

	return ch, nil
}

// myAgentExecutor is a sample AgentExecutor implementation.
type myAgentExecutor struct{}

func (e *myAgentExecutor) Execute(ctx context.Context, reqCtx *agent_execution.RequestContext, eventQueue agent_execution.EventQueue) error {
	// Send initial status
	eventQueue.Send(&a2a.TaskStatusUpdateEvent{
		Kind:      a2a.StatusUpdateEventKind,
		TaskID:    reqCtx.TaskID,
		ContextID: reqCtx.ContextID,
		Status: a2a.TaskStatus{
			State: a2a.TaskStateWorking,
		},
	})

	// Process the request
	time.Sleep(1 * time.Second)

	// Send result
	eventQueue.Send(&a2a.Message{
		Kind:      a2a.MessageEventKind,
		MessageID: "result-1",
		Role:      a2a.RoleAgent,
		Parts: []a2a.Part{
			a2a.NewTextPart("Task completed successfully"),
		},
		TaskID: reqCtx.TaskID,
	})

	// Send completion
	eventQueue.Send(&a2a.TaskStatusUpdateEvent{
		Kind:      a2a.StatusUpdateEventKind,
		TaskID:    reqCtx.TaskID,
		ContextID: reqCtx.ContextID,
		Final:     true,
		Status: a2a.TaskStatus{
			State: a2a.TaskStateCompleted,
		},
	})

	return nil
}

func (e *myAgentExecutor) Cancel(ctx context.Context, reqCtx *agent_execution.RequestContext, eventQueue agent_execution.EventQueue) error {
	// Send cancellation status
	eventQueue.Send(&a2a.TaskStatusUpdateEvent{
		Kind:      a2a.StatusUpdateEventKind,
		TaskID:    reqCtx.TaskID,
		ContextID: reqCtx.ContextID,
		Final:     true,
		Status: a2a.TaskStatus{
			State: a2a.TaskStateCanceled,
		},
	})

	return nil
}