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

package event

import "errors"

var (
	// ErrQueueClosed is returned when attempting to use a closed queue.
	// This corresponds to Python's QueueClosed (asyncio.QueueEmpty/QueueShutDown).
	ErrQueueClosed = errors.New("event queue is closed")

	// ErrQueueEmpty is returned when attempting to dequeue from an empty queue
	// in non-blocking mode.
	ErrQueueEmpty = errors.New("event queue is empty")

	// ErrInvalidQueueSize is returned when attempting to create a queue with
	// invalid size.
	ErrInvalidQueueSize = errors.New("max queue size must be greater than 0")
)