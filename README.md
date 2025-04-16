# A2A for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/go-a2a/a2a.svg)](https://pkg.go.dev/github.com/go-a2a/a2a)
[![Go](https://github.com/go-a2a/a2a/actions/workflows/go.yml/badge.svg)](https://github.com/go-a2a/a2a/actions/workflows/test.yml)

> [!IMPORTANT]
> This project is in the alpha stage.
>
> Flags, configuration, behavior, and design may change significantly.

A Go implementation of the [Agent-to-Agent (A2A)](https://github.com/google/A2A) protocol. 

A2A is an open protocol enabling communication and interoperability between opaque agentic applications.

## Installation

```bash
go get github.com/go-a2a/a2a
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-a2a/a2a"
)

func main() {
	// Create a new task
	task := a2a.NewTaskBuilder().
		WithAgentCard(a2a.AgentCard{
			Name:    "Example Agent",
			URL:     "https://example.com/agent",
			Version: "1.0.0",
		}).
		WithClient("example-client").
		AddMessage(
			a2a.NewMessageBuilder("user").
				AddTextPart("Hello, agent!").
				Build(),
		).
		Build()

	fmt.Printf("Task ID: %s\n", task.ID)
	fmt.Printf("Task State: %s\n", task.State)
}
```

## Features

- Full implementation of the A2A protocol in Go
- Builder pattern for creating tasks, messages, and artifacts
- OpenTelemetry integration for tracing and metrics
- High-performance JSON serialization using Bytedance's Sonic

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](./LICENSE) file for details.
