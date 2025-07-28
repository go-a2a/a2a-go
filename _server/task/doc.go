// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package task provides task management functionality for the A2A server.
//
// This package implements task lifecycle management, status updates, artifact handling,
// and event processing for the Agent-to-Agent (A2A) protocol. It provides a complete
// task management system that mirrors the functionality of the Python a2a.server.tasks
// module but adapted for Go's concurrency model and idioms.
//
// The main components include:
//
// - TaskUpdater: Interface for agents to publish task-related events
// - TaskManager: Manages task lifecycle during request execution
// - TaskStore: Interface for task persistence with multiple implementations
// - ResultAggregator: Processes event streams and provides consumption patterns
//
// The package integrates with the existing event system and database models
// to provide a complete task management solution.
package task
