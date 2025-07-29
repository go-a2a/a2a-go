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

// Package pyclient provides a Go implementation of the A2A client that follows
// Python client patterns and conventions, adapted to Go idioms.
//
// This package provides a high-level API for communicating with A2A agents,
// handling connection management, message serialization, streaming responses,
// and error handling.
//
// # Basic Usage
//
//	ctx := context.Background()
//	client, err := pyclient.New(
//		pyclient.WithBaseURL("https://example.com"),
//		pyclient.WithTimeout(30 * time.Second),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Send a message
//	response, err := client.SendMessage(ctx, &a2a.MessageSendParams{
//		Message: &a2a.Message{
//			Role: a2a.RoleUser,
//			Parts: []a2a.Part{
//				&a2a.TextPart{Text: "Hello, agent!"},
//			},
//		},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Streaming Support
//
// The client supports Server-Sent Events (SSE) for real-time updates:
//
//	stream, err := client.SendStreamMessage(ctx, params)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stream.Close()
//
//	for event := range stream.Events() {
//		switch ev := event.(type) {
//		case *a2a.Message:
//			fmt.Printf("Message: %v\n", ev)
//		case *a2a.StatusUpdate:
//			fmt.Printf("Status: %v\n", ev)
//		case error:
//			log.Printf("Stream error: %v", ev)
//		}
//	}
//
// # Error Handling
//
// The package provides typed errors that correspond to A2A protocol errors:
//
//	if err := client.CancelTask(ctx, taskID); err != nil {
//		var a2aErr *pyclient.Error
//		if errors.As(err, &a2aErr) {
//			switch a2aErr.Code {
//			case pyclient.ErrCodeTaskNotFound:
//				// Handle task not found
//			case pyclient.ErrCodeTaskNotCancelable:
//				// Handle non-cancelable task
//			}
//		}
//	}
//
// # Connection Management
//
// The client handles connection pooling, automatic reconnection, and
// graceful shutdown internally. Use context for request-level cancellation
// and timeouts.
package pyclient
