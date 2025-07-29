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

// Package event provides event queue and consumer implementations for A2A server events.
// It mirrors the Python a2a.server.events module with idiomatic Go patterns.
package event

import (
	"github.com/go-a2a/a2a-go"
)

// Event represents any event that can be sent through the event system.
// This corresponds to the Python Event union type:
// Event = Union[Message, Task, TaskStatusUpdateEvent, TaskArtifactUpdateEvent]
type Event interface {
	// GetEventKind returns the event kind for type discrimination.
	GetEventKind() a2a.EventKind
	// GetTaskID returns the task ID associated with this event.
	GetTaskID() string
}

// Ensure types implement Event interface
var (
	_ Event = (*a2a.Message)(nil)
	_ Event = (*a2a.Task)(nil)
	_ Event = (*a2a.TaskStatusUpdateEvent)(nil)
	_ Event = (*a2a.TaskArtifactUpdateEvent)(nil)
)

// IsFinalEvent determines if an event represents a final event in the stream.
// A final event is:
// - TaskStatusUpdateEvent with Final=true
// - Any Message
// - Task with terminal state (completed, canceled, failed, rejected)
func IsFinalEvent(event Event) bool {
	switch e := event.(type) {
	case *a2a.TaskStatusUpdateEvent:
		return e.Final
	case *a2a.Message:
		return true
	case *a2a.Task:
		state := e.Status.State
		return state == a2a.TaskStateCompleted ||
			state == a2a.TaskStateCanceled ||
			state == a2a.TaskStateFailed ||
			state == a2a.TaskStateRejected
	default:
		return false
	}
}