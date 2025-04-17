# A2A Go Implementation

[![Go Reference](https://pkg.go.dev/badge/github.com/go-a2a/a2a.svg)](https://pkg.go.dev/github.com/go-a2a/a2a)
[![Go](https://github.com/go-a2a/a2a/actions/workflows/go.yml/badge.svg)](https://github.com/go-a2a/a2a/actions/workflows/test.yml)

> [!IMPORTANT]
> This project is in the alpha stage.
>
> Flags, configuration, behavior, and design may change significantly.

This repository contains a Go implementation of Google's [Agent-to-Agent (A2A)](https://github.com/google/A2A) protocol, which enables communication and interoperability between AI agents built on different frameworks.

## Overview

A2A (Agent-to-Agent) is an open protocol that provides a standardized way for AI agents to communicate with each other, regardless of the framework or vendor they are built on. This implementation follows the [A2A protocol specification](https://github.com/google/A2A/blob/main/specification/json/a2a.json) and provides both client and server components.

## Features

- Complete implementation of the A2A protocol in Go
- Support for both client and server roles
- JSON-RPC over HTTP communication
- Streaming responses via Server-Sent Events (SSE)
- Push notifications for task updates
- JWT-based authentication
- In-memory task storage
- Comprehensive type definitions matching the A2A protocol

## Requirements

- Go 1.24 or higher

## Installation

```bash
go get github.com/go-a2a/a2a
```

## Usage

### Basic Example

Here's a minimal example of setting up an A2A server and client:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/client"
	"github.com/go-a2a/a2a/server"
)

func main() {
	// Create a simple echo server
	go runServer()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	// Run client
	if err := runClient(); err != nil {
		log.Fatalf("Client error: %v", err)
	}
}

func runServer() {
	// Create server components
	taskStore := server.NewInMemoryTaskStore()
	subscriptionManager := server.NewSubscriptionManager()
	eventEmitter := server.NewDefaultTaskEventEmitter(subscriptionManager)

	// Create agent card
	agentCard := &a2a.AgentCard{
		Name:                "EchoAgent",
		Description:         "A simple echo agent",
		Version:             "1.0.0",
		SupportedOutputModes: []string{"text"},
		Capabilities: a2a.AgentCapabilities{
			Streaming: true,
		},
	}

	// Create handler with echo functionality
	handler := server.NewDefaultA2AHandler(agentCard, taskStore, eventEmitter, func(ctx context.Context, task *a2a.Task) error {
		// Extract input text
		inputMsg := task.History[0]
		var inputText string
		for _, part := range inputMsg.Parts {
			if textPart, ok := part.(types.TextPart); ok {
				inputText = textPart.Text
				break
			}
		}

		// Create response
		task.Artifacts = []types.Artifact{
			{
				Parts: []types.Part{
					types.TextPart{Text: "Echo: " + inputText},
				},
				Index: 0,
			},
		}
		task.Status.State = types.TaskStateCompleted
		task.Status.Timestamp = time.Now()

		// Update task in store
		return taskStore.UpdateTask(ctx, task)
	})

	// Create and start the server
	a2aServer := server.NewA2AServer(":8080", handler, taskStore, eventEmitter)
	if err := a2aServer.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func runClient() error {
	// Create client
	a2aClient := client.NewA2AClient("http://localhost:8080")

	// Create message
	message := types.Message{
		Role: types.RoleUser,
		Parts: []types.Part{
			types.TextPart{Text: "Hello, agent!"},
		},
	}

	// Send task
	params := &types.TaskSendParams{
		ID:                 "task-1",
		Message:            message,
		AcceptedOutputModes: []string{"text"},
	}

	// Send request
	task, err := a2aClient.TaskSend(context.Background(), params)
	if err != nil {
		return err
	}

	// Print response
	fmt.Printf("Task status: %s\n", task.Status.State)
	if len(task.Artifacts) > 0 {
		fmt.Println("Response:")
		for _, part := range task.Artifacts[0].Parts {
			if textPart, ok := part.(types.TextPart); ok {
				fmt.Println(textPart.Text)
			}
		}
	}

	return nil
}
```

### Running the Example

The included example demonstrates both server and client functionality:

```bash
# Run as server
go run example/main.go -server

# In another terminal, run as client
go run example/main.go -client -message "Hello, A2A!"
```

## Package Structure

- `a2a`: Core data structures for the A2A protocol
    - `client`: Client implementation for making requests to A2A servers
    - `server`: Server implementation for handling A2A requests
    - `utils`: Utilities for authentication, push notifications, etc.

## Implementation Details

### Types Package

Provides Go equivalents of the A2A protocol data structures:

- Task, Message, Part (TextPart, FilePart, DataPart), Artifact
- JSON-RPC request/response types
- Event types for streaming and push notifications

### Client Package

Implements the client side of the A2A protocol:

- `A2AClient`: Core client for making A2A requests
- `CardResolver`: For resolving agent cards by URL
- Support for streaming responses

### Server Package

Implements the server side of the A2A protocol:

- `A2AServer`: HTTP server handling A2A requests
- `TaskStore`: Interface for storing tasks (with in-memory implementation)
- `SubscriptionManager`: For managing streaming subscriptions

### Utils Package

Provides utility functions:

- JWT signing and verification
- JWKS endpoint support
- Push notification sending

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](./LICENSE) file for details.

## Credits

This implementation is based on the [A2A protocol](https://github.com/google/A2A) developed by Google.
