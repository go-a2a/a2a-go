// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-a2a/a2a-go/client"
	"github.com/go-a2a/a2a-go/client/auth"
	"github.com/go-a2a/a2a-go/client/internal/helpers"
)

// ExampleHTTPClient demonstrates basic usage of the HTTP client.
func ExampleHTTPClient() {
	// Create a client with default options
	httpClient := client.NewHTTPClientWithURL("https://api.example.com/a2a")
	defer httpClient.Close()

	// Create a message
	message := helpers.CreateTextMessageObject("Hello, A2A agent!")

	// Create a send message request
	request := &client.SendMessageRequest{
		Message: message,
		Configuration: helpers.CreateSendMessageConfiguration(
			helpers.WithAcceptedOutputModes([]string{"text", "json"}),
			helpers.WithHistoryLength(10),
			helpers.WithBlocking(true),
		),
	}

	// Send the message
	ctx := context.Background()
	response, _ := httpClient.SendMessage(ctx, request)
	_ = response
}

// ExampleHTTPClient_withAuthentication demonstrates using the client with authentication.
func ExampleHTTPClient_withAuthentication() {
	// Create API key credentials
	credentials := auth.CreateAPIKeyCredentials("your-api-key")

	// Create a credential service with the credentials
	credentialService := auth.NewStaticCredentialService(credentials)

	// Create an authentication interceptor
	authInterceptor := auth.NewAuthInterceptor(credentialService)
	_ = authInterceptor

	// Create a client with authentication - simplified for example
	httpClient := client.NewHTTPClientWithURL("https://api.example.com/a2a")
	defer httpClient.Close()

	// Use the client as normal
	message := helpers.CreateTextMessageObject("Authenticated message")
	request := &client.SendMessageRequest{
		Message: message,
	}

	ctx := context.Background()
	response, _ := httpClient.SendMessage(ctx, request)
	_ = response
}

// ExampleHTTPClient_streaming demonstrates streaming message responses.
func ExampleHTTPClient_streaming() {
	httpClient := client.NewHTTPClientWithURL("https://api.example.com/a2a")
	defer httpClient.Close()

	// Create a streaming message request
	message := helpers.CreateTextMessageObject("Stream me some data")
	request := &client.SendStreamingMessageRequest{
		Message: message,
		Configuration: helpers.CreateSendMessageConfiguration(
			helpers.WithBlocking(false),
		),
	}

	// Send the streaming message
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	responseChannel, err := httpClient.SendMessageStreaming(ctx, request)
	if err != nil {
		log.Printf("Error starting stream: %v", err)
		return
	}

	// Process streaming responses
	for response := range responseChannel {
		if response.Error != nil {
			log.Printf("Stream error: %v", response.Error)
			break
		}
		fmt.Printf("Streaming response: %s\n", response.ID)
	}
}

// ExampleCreateClientFromAgentCardURL demonstrates creating a client from an agent card URL.
func ExampleCreateClientFromAgentCardURL() {
	ctx := context.Background()

	// Create a client from an agent card URL
	httpClient, err := client.CreateClientFromAgentCardURL(
		ctx,
		"https://api.example.com",
		client.WithTimeout(10*time.Second),
	)
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}
	defer httpClient.Close()

	// Use the client
	message := helpers.CreateTextMessageObject("Hello from auto-configured client!")
	request := &client.SendMessageRequest{
		Message: message,
	}

	response, err := httpClient.SendMessage(ctx, request)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", response.ID)
}

func ExampleGetTaskRequest() {
	httpClient := client.NewHTTPClientWithURL("https://api.example.com/a2a")
	defer httpClient.Close()

	ctx := context.Background()
	taskID := "example-task-id"

	// Get task
	getTaskRequest := &client.GetTaskRequest{
		TaskID:        taskID,
		HistoryLength: 5,
	}

	taskResponse, err := httpClient.GetTask(ctx, getTaskRequest)
	if err != nil {
		log.Printf("Error getting task: %v", err)
		return
	}

	fmt.Printf("Task status: %s\n", taskResponse.Result.Status.State)

	// Cancel task
	cancelRequest := &client.CancelTaskRequest{
		TaskID: taskID,
	}

	cancelResponse, err := httpClient.CancelTask(ctx, cancelRequest)
	if err != nil {
		log.Printf("Error canceling task: %v", err)
		return
	}

	fmt.Printf("Task canceled: %s\n", cancelResponse.Result.Status.State)
}

func ExampleNewInMemoryCredentialStore() {
	// Create an in-memory credential store
	store := auth.NewInMemoryCredentialStore(24 * time.Hour)

	// Create some credentials
	credentials := auth.CreateBearerCredentials("access-token", nil)

	// Store credentials
	ctx := context.Background()
	err := store.StoreCredentials(ctx, "user-123", credentials)
	if err != nil {
		log.Printf("Error storing credentials: %v", err)
		return
	}

	// Retrieve credentials
	retrievedCreds, err := store.GetCredentials(ctx, "user-123")
	if err != nil {
		log.Printf("Error retrieving credentials: %v", err)
		return
	}

	fmt.Printf("Credential type: %s\n", retrievedCreds.Type)
	// Output: Credential type: bearer
}

// ExampleRetryPolicy demonstrates using retry policies.
func ExampleRetryPolicy() {
	// Create a custom retry policy
	retryPolicy := &client.RetryPolicy{
		MaxAttempts:  5,
		InitialDelay: 2 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}

	// Create a client with retry policy
	httpClient := client.NewHTTPClientWithURL(
		"https://api.example.com/a2a",
		client.WithRetryPolicy(retryPolicy),
	)
	defer httpClient.Close()

	// Use the client - requests will be retried automatically
	message := helpers.CreateTextMessageObject("This will be retried if it fails")
	request := &client.SendMessageRequest{
		Message: message,
	}

	ctx := context.Background()
	response, err := httpClient.SendMessage(ctx, request)
	if err != nil {
		log.Printf("Error (after retries): %v", err)
		return
	}

	fmt.Printf("Response: %s\n", response.ID)
}

func ExampleWithInterceptors() {
	// Create a client with logging interceptor
	httpClient := client.NewHTTPClientWithURL(
		"https://api.example.com/a2a",
		client.WithInterceptors(client.LoggingInterceptor(client.NoopLogger{})),
	)
	defer httpClient.Close()

	// Use the client - all requests will be logged
	message := helpers.CreateTextMessageObject("This request will be logged")
	request := &client.SendMessageRequest{
		Message: message,
	}

	ctx := context.Background()
	_, _ = httpClient.SendMessage(ctx, request)
}

func ExampleCreateTextMessageObject() {
	httpClient := client.NewHTTPClientWithURL("https://api.example.com/a2a")
	defer httpClient.Close()

	message := helpers.CreateTextMessageObject("Test message")
	request := &client.SendMessageRequest{
		Message: message,
	}

	ctx := context.Background()
	response, err := httpClient.SendMessage(ctx, request)
	if err != nil {
		// Check if it's a specific client error
		if clientErr, ok := err.(client.ClientError); ok {
			fmt.Printf("Client error code: %d\n", clientErr.Code())
			fmt.Printf("Client error message: %s\n", clientErr.Message())
			fmt.Printf("Is retryable: %v\n", clientErr.IsRetryable())
		}

		// Check if it's an HTTP error
		if httpErr, ok := err.(*client.HTTPError); ok {
			fmt.Printf("HTTP status code: %d\n", httpErr.StatusCode)
		}

		// Check if it's a JSON error
		if jsonErr, ok := err.(*client.JSONError); ok {
			fmt.Printf("JSON error: %s\n", jsonErr.Message())
		}

		return
	}

	fmt.Printf("Success: %s\n", response.ID)
}
