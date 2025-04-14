// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"time"

	"github.com/go-a2a/a2a"
)

func Example() {
	// Create a simple agent card
	agentCard := a2a.AgentCard{
		Name:    "Example Agent",
		URL:     "http://localhost:8080",
		Version: "1.0.0",
		Capabilities: a2a.AgentCapabilities{
			Streaming:         true,
			PushNotifications: true,
		},
		DefaultInputModes:  []string{"text"},
		DefaultOutputModes: []string{"text"},
		Skills: []a2a.AgentSkill{
			{
				ID:          "echo",
				Name:        "Echo",
				Description: stringPtr("Echoes the input back to the user"),
			},
		},
	}

	// Create a task manager
	taskManager := a2a.NewInMemoryTaskManager()

	// Create a server
	server := a2a.NewServer(agentCard, taskManager)

	// Create a test HTTP server
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()

	// Create a client
	client, err := a2a.NewClient(httpServer.URL)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}

	// Set a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the agent card
	card, err := client.GetAgentCard(ctx)
	if err != nil {
		fmt.Printf("Failed to get agent card: %v\n", err)
		return
	}

	fmt.Printf("Agent: %s (v%s)\n", card.Name, card.Version)
	fmt.Printf("Capabilities: Streaming=%v, PushNotifications=%v\n",
		card.Capabilities.Streaming, card.Capabilities.PushNotifications)

	// Create a task
	task, err := client.SendTask(ctx, a2a.TaskSendParams{
		ID: "task-1",
		Message: a2a.Message{
			Role: "user",
			Parts: []a2a.Part{
				a2a.TextPart{
					Type: "text",
					Text: "Hello, world!",
				},
			},
		},
	})
	if err != nil {
		fmt.Printf("Failed to send task: %v\n", err)
		return
	}

	fmt.Printf("Task created: %s (state: %s)\n", task.ID, task.Status.State)

	// Get the task
	task, err = client.GetTask(ctx, "task-1", nil)
	if err != nil {
		fmt.Printf("Failed to get task: %v\n", err)
		return
	}

	fmt.Printf("Retrieved task: %s (state: %s)\n", task.ID, task.Status.State)

	// Cancel the task
	task, err = client.CancelTask(ctx, "task-1")
	if err != nil {
		fmt.Printf("Failed to cancel task: %v\n", err)
		return
	}

	fmt.Printf("Task canceled: %s (state: %s)\n", task.ID, task.Status.State)

	// Output:
	// Agent: Example Agent (v1.0.0)
	// Capabilities: Streaming=true, PushNotifications=true
	// Task created: task-1 (state: working)
	// Retrieved task: task-1 (state: working)
	// Task canceled: task-1 (state: canceled)
}

// Helper function to create a string pointer.
func stringPtr(s string) *string {
	return &s
}
