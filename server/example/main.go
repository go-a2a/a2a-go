// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/server"
)

func main() {
	// Create a task manager with default handler
	taskManager := server.NewInMemoryTaskManager(nil)

	// Create agent card
	agentCard := &a2a.AgentCard{
		Name:        "Example A2A Agent",
		Description: "A simple example agent implementing the A2A protocol",
		URL:         "http://localhost:8080/a2a",
		Version:     "1.0.0",
		Provider: &a2a.AgentProvider{
			Organization: "Example Org",
			URL:          "https://example.org",
		},
		Capabilities: a2a.AgentCapabilities{
			Streaming:              true,
			PushNotifications:      true,
			StateTransitionHistory: true,
		},
		DefaultInputModes:  []string{"text"},
		DefaultOutputModes: []string{"text"},
		Skills: []a2a.AgentSkill{
			{
				ID:          "example-skill",
				Name:        "Example Skill",
				Description: "A simple example skill",
				Tags:        []string{"example", "demo"},
			},
		},
	}

	// Create server configuration
	config := server.Config{
		AgentCard:               agentCard,
		TaskManager:             taskManager,
		EnableStreaming:         true,
		EnablePushNotifications: true,
		EnableStateHistory:      true,
	}

	// Create server
	a2aServer, err := server.NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: a2aServer,
	}

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		log.Println("Shutting down server...")
		if err := httpServer.Close(); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
	}()

	// Start server
	fmt.Println("A2A server running on http://localhost:8080")
	fmt.Println("Agent card available at http://localhost:8080/.well-known/agent.json")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
