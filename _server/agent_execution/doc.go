// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package agent_execution provides the core components for executing agent logic
// within the A2A server. This package is a Go port of the Python a2a-python
// server/agent_execution module.
//
// The key components are:
//   - AgentExecutor: Interface for implementing agent business logic
//   - RequestContext: Context container for request processing
//   - RequestContextBuilder: Interface for building request contexts
//   - EventQueue: Interface for publishing events during execution
//
// This package enables the A2A server to execute agent-specific logic in response
// to incoming requests, with proper context management and event streaming.
package agent_execution
