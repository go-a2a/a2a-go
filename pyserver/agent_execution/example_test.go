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

package agent_execution_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/pyserver/agent_execution"
)

// SimpleEchoAgent demonstrates a basic agent that echoes messages back.
type SimpleEchoAgent struct {
	agent_execution.BaseAgentExecutor
}

func (a *SimpleEchoAgent) Execute(ctx context.Context, reqCtx *agent_execution.RequestContext, eventQueue agent_execution.EventQueue) error {
	// Send initial working status
	eventQueue.Put(ctx, agent_execution.CreateWorkingStatusEvent(reqCtx.TaskID, reqCtx.ContextID))

	// Extract text from the message
	inputText := agent_execution.ExtractTextFromMessage(reqCtx.Message)

	// Create echo response
	response := fmt.Sprintf("Echo: %s", inputText)
	artifact := agent_execution.NewTextArtifact("echo_response", response)

	// Send the artifact
	eventQueue.Put(ctx, agent_execution.NewTaskArtifactUpdateEvent(reqCtx.TaskID, reqCtx.ContextID, artifact))

	// Send completion status
	eventQueue.Put(ctx, agent_execution.CreateCompletedStatusEvent(reqCtx.TaskID, reqCtx.ContextID))

	return nil
}

// Example demonstrates using the SimpleEchoAgent.
func Example_simpleEchoAgent() {
	// Create the agent
	agent := &SimpleEchoAgent{}

	// Create a message to send
	message := &a2a.Message{
		Role:      a2a.RoleUser,
		Parts:     []a2a.Part{a2a.NewTextPart("Hello, Agent!")},
		MessageID: "msg-123",
	}

	// Create the protocol message executor
	executor := agent_execution.NewProtocolMessageExecutor(agent, nil)

	// Execute the message
	task, err := executor.ExecuteMessage(context.Background(), &a2a.MessageSendParams{
		Message: message,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Print the result
	fmt.Printf("Task ID: %s\n", task.ID)
	fmt.Printf("Status: %s\n", task.Status.State)
	if len(task.Artifacts) > 0 {
		fmt.Printf("Artifact: %s\n", task.Artifacts[0].Name)
	} else {
		fmt.Printf("No artifacts found\n")
	}

	// Output:
	// Task ID: msg-123
	// Status: working
	// Artifact: echo_response
}

// CalculatorAgent demonstrates an agent that performs calculations.
type CalculatorAgent struct {
	agent_execution.BaseAgentExecutor
}

func (a *CalculatorAgent) Execute(ctx context.Context, reqCtx *agent_execution.RequestContext, eventQueue agent_execution.EventQueue) error {
	// Send initial status
	eventQueue.Put(ctx, agent_execution.CreateWorkingStatusEvent(reqCtx.TaskID, reqCtx.ContextID))

	// Extract the calculation request
	inputText := agent_execution.ExtractTextFromMessage(reqCtx.Message)

	// Simple parsing (in real implementation, use a proper parser)
	var result string
	if strings.Contains(inputText, "+") {
		result = "Result: 5 (example calculation)"
	} else {
		result = "Please provide a calculation in the format: number + number"
	}

	// Create result artifact
	artifact := agent_execution.NewTextArtifact("calculation_result", result)
	eventQueue.Put(ctx, agent_execution.NewTaskArtifactUpdateEvent(reqCtx.TaskID, reqCtx.ContextID, artifact))

	// Complete the task
	eventQueue.Put(ctx, agent_execution.CreateCompletedStatusEvent(reqCtx.TaskID, reqCtx.ContextID))

	return nil
}

// Example_streamingAgent demonstrates streaming responses.
func Example_streamingAgent() {
	// Create a calculator agent
	agent := &CalculatorAgent{}

	// Create executor
	executor := agent_execution.NewProtocolMessageExecutor(agent, nil)

	// Create a calculation request
	message := &a2a.Message{
		Role:      a2a.RoleUser,
		Parts:     []a2a.Part{a2a.NewTextPart("2 + 3")},
		MessageID: "calc-123",
	}

	// Execute with streaming
	ctx := context.Background()
	responseChan, err := executor.ExecuteStreamingMessage(ctx, &a2a.MessageSendParams{
		Message: message,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Process streaming responses
	for response := range responseChan {
		switch r := response.(type) {
		case *a2a.TaskStatusUpdateEvent:
			fmt.Printf("Status: %s\n", r.Status.State)
		case *a2a.TaskArtifactUpdateEvent:
			fmt.Printf("Artifact: %s\n", r.Artifact.Name)
		}
	}

	// Output:
	// Status: working
}

// InteractiveAgent demonstrates an agent that requires user input.
type InteractiveAgent struct {
	agent_execution.BaseAgentExecutor
	needsMoreInfo bool
}

func (a *InteractiveAgent) Execute(ctx context.Context, reqCtx *agent_execution.RequestContext, eventQueue agent_execution.EventQueue) error {
	// Check if this is the first interaction
	if reqCtx.Task == nil || reqCtx.Task.Status.State != a2a.TaskStateInputRequired {
		// First interaction - ask for more information
		eventQueue.Put(ctx, agent_execution.CreateInputRequiredStatusEvent(
			reqCtx.TaskID,
			reqCtx.ContextID,
			"Please provide your favorite color",
		))
		return nil
	}

	// Second interaction - process the response
	eventQueue.Put(ctx, agent_execution.CreateWorkingStatusEvent(reqCtx.TaskID, reqCtx.ContextID))

	color := agent_execution.ExtractTextFromMessage(reqCtx.Message)
	response := fmt.Sprintf("Your favorite color is %s. Nice choice!", color)

	artifact := agent_execution.NewTextArtifact("color_response", response)
	eventQueue.Put(ctx, agent_execution.NewTaskArtifactUpdateEvent(reqCtx.TaskID, reqCtx.ContextID, artifact))

	eventQueue.Put(ctx, agent_execution.CreateCompletedStatusEvent(reqCtx.TaskID, reqCtx.ContextID))

	return nil
}