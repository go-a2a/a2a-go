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

import (
	"context"
	"sync"
)

// QueueManager manages event queues for tasks.
type QueueManager interface {
	// Get returns the event queue for a task, creating it if necessary.
	Get(taskID string) (*EventQueue, error)
	// Tap creates a child queue for the specified task that receives
	// copies of all events enqueued to the parent queue.
	Tap(taskID string) (*EventQueue, error)
	// Close closes and removes the queue for a task.
	Close(taskID string) error
	// CloseAll closes all managed queues.
	CloseAll() error
}

// InMemoryQueueManager provides in-memory event queue management.
// This corresponds to Python's InMemoryQueueManager.
type InMemoryQueueManager struct {
	mu         sync.RWMutex
	queues     map[string]*EventQueue
	maxSize    int
	defaultCtx context.Context
}

// NewInMemoryQueueManager creates a new in-memory queue manager.
func NewInMemoryQueueManager(maxQueueSize int) *InMemoryQueueManager {
	if maxQueueSize <= 0 {
		maxQueueSize = DefaultMaxQueueSize
	}
	return &InMemoryQueueManager{
		queues:     make(map[string]*EventQueue),
		maxSize:    maxQueueSize,
		defaultCtx: context.Background(),
	}
}

// Get returns the event queue for a task, creating it if necessary.
func (m *InMemoryQueueManager) Get(taskID string) (*EventQueue, error) {
	m.mu.RLock()
	queue, exists := m.queues[taskID]
	m.mu.RUnlock()

	if exists {
		return queue, nil
	}

	// Need to create a new queue
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check in case another goroutine created it
	if queue, exists = m.queues[taskID]; exists {
		return queue, nil
	}

	// Create new queue
	queue, err := NewEventQueue(m.maxSize)
	if err != nil {
		return nil, err
	}

	m.queues[taskID] = queue
	return queue, nil
}

// Tap creates a child queue for the specified task.
func (m *InMemoryQueueManager) Tap(taskID string) (*EventQueue, error) {
	queue, err := m.Get(taskID)
	if err != nil {
		return nil, err
	}
	return queue.Tap()
}

// Close closes and removes the queue for a task.
func (m *InMemoryQueueManager) Close(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue, exists := m.queues[taskID]
	if !exists {
		return nil
	}

	delete(m.queues, taskID)
	return queue.Close()
}

// CloseAll closes all managed queues.
func (m *InMemoryQueueManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	for taskID, queue := range m.queues {
		if err := queue.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(m.queues, taskID)
	}

	return firstErr
}

// Size returns the number of managed queues.
func (m *InMemoryQueueManager) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.queues)
}

// TaskIDs returns a slice of all task IDs with active queues.
func (m *InMemoryQueueManager) TaskIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.queues))
	for id := range m.queues {
		ids = append(ids, id)
	}
	return ids
}