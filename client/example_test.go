// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/client"
)

func ExampleClient_SendTask() {
	// Create a new client
	c, err := client.New("https://example.com/a2a")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Create a message
	message := client.CreateTextMessage("Hello, agent!")

	// Send a task
	ctx := context.Background()
	params := a2a.TaskSendParams{
		ID:      "task123",
		Message: message,
	}

	task, err := c.SendTask(ctx, params)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Task ID: %s\n", task.ID)
	fmt.Printf("Task State: %s\n", task.Status.State)
}

func ExampleClient_SendTaskStream() {
	// Create a new client
	c, err := client.New("https://example.com/a2a")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Create a message
	message := client.CreateTextMessage("Hello, streaming agent!")

	// Send a task with streaming
	ctx := context.Background()
	params := a2a.TaskSendParams{
		ID:      "task456",
		Message: message,
	}

	stream, err := c.SendTaskStream(ctx, params)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	// Wait for the task to complete or timeout after 30 seconds
	state, err := stream.WaitForCompletionWithTimeout(30 * time.Second)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Final task state: %s\n", state)
}

func ExampleClient_GetAgentCard() {
	// Create a new client
	c, err := client.New("https://example.com/a2a")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Get the agent card
	ctx := context.Background()
	card, err := c.GetAgentCard(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Agent name: %s\n", card.Name)
	fmt.Printf("Agent version: %s\n", card.Version)
	fmt.Printf("Number of skills: %d\n", len(card.Skills))
}

func ExamplePushServer() {
	// Create a push server
	ps := client.NewPushServer(":8080")

	// Register a handler for a specific task
	ps.RegisterHandler("task789", func(ctx context.Context, taskID string, event any) error {
		fmt.Printf("Received notification for task %s: %v\n", taskID, event)
		return nil
	})

	// Start the server (this blocks, so typically you'd run it in a goroutine)
	go func() {
		if err := ps.Start(); err != nil {
			log.Printf("Push server error: %v", err)
		}
	}()

	// When done, shut down the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = ps.Shutdown(ctx)
}
