// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// InMemoryQueueManager implements QueueManager using in-memory storage.
// It is suitable for single-instance deployments.
type InMemoryQueueManager struct {
	*BaseQueueManager
	queues  map[string]*EventQueue
	metrics *QueueManagerMetrics
	mu      sync.RWMutex
}

var _ QueueManager = (*InMemoryQueueManager)(nil)

// NewInMemoryQueueManager creates a new InMemoryQueueManager.
func NewInMemoryQueueManager() *InMemoryQueueManager {
	return NewInMemoryQueueManagerWithConfig(DefaultQueueManagerConfig())
}

// NewInMemoryQueueManagerWithConfig creates a new InMemoryQueueManager with the given configuration.
func NewInMemoryQueueManagerWithConfig(config *QueueManagerConfig) *InMemoryQueueManager {
	return &InMemoryQueueManager{
		BaseQueueManager: NewBaseQueueManager(config),
		queues:           make(map[string]*EventQueue),
		metrics:          NewQueueManagerMetrics(),
	}
}

// NewInMemoryQueueManagerWithOptions creates a new InMemoryQueueManager with the given options.
func NewInMemoryQueueManagerWithOptions(options ...QueueManagerOption) *InMemoryQueueManager {
	config := DefaultQueueManagerConfig()
	ApplyOptions(config, options...)
	return NewInMemoryQueueManagerWithConfig(config)
}

// Add creates a new event queue for the given task ID.
// Returns TaskQueueExistsError if the queue already exists.
func (imqm *InMemoryQueueManager) Add(taskID string, queue *EventQueue) error {
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if queue == nil {
		return fmt.Errorf("queue cannot be nil")
	}

	imqm.mu.Lock()
	defer imqm.mu.Unlock()

	if _, exists := imqm.queues[taskID]; exists {
		return &TaskQueueExistsError{TaskID: taskID}
	}

	imqm.queues[taskID] = queue
	imqm.metrics.RecordQueueAccess(taskID, time.Now().Unix())
	imqm.updateMetricsLocked()

	return nil
}

// Get retrieves an EventQueue for the given task ID.
// Returns nil if the queue does not exist.
func (imqm *InMemoryQueueManager) Get(taskID string) *EventQueue {
	if taskID == "" {
		return nil
	}

	imqm.mu.RLock()
	queue, exists := imqm.queues[taskID]
	config := imqm.config
	imqm.mu.RUnlock()

	if exists {
		imqm.metrics.RecordQueueAccess(taskID, time.Now().Unix())
		return queue
	}

	// Auto-create queue if configured to do so
	if config.AutoCreateQueues {
		newQueue := imqm.createQueue(taskID)
		if err := imqm.Add(taskID, newQueue); err != nil {
			// If add fails (e.g., queue was created concurrently), return existing queue
			imqm.mu.RLock()
			queue, exists := imqm.queues[taskID]
			imqm.mu.RUnlock()
			if exists {
				return queue
			}
			return nil
		}
		return newQueue
	}

	return nil
}

// Tap creates a child queue for the given task ID.
// Returns nil if the parent queue does not exist.
func (imqm *InMemoryQueueManager) Tap(taskID string) *EventQueue {
	if taskID == "" {
		return nil
	}

	imqm.mu.RLock()
	queue, exists := imqm.queues[taskID]
	imqm.mu.RUnlock()

	if !exists {
		return nil
	}

	imqm.metrics.RecordQueueAccess(taskID, time.Now().Unix())
	return queue.Tap()
}

// Close closes and removes the event queue for the given task ID.
// Returns NoTaskQueueError if the queue does not exist.
func (imqm *InMemoryQueueManager) Close(taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	imqm.mu.Lock()
	defer imqm.mu.Unlock()

	queue, exists := imqm.queues[taskID]
	if !exists {
		return &NoTaskQueueError{TaskID: taskID}
	}

	if err := queue.Close(); err != nil {
		return NewQueueOperationError("close", taskID, "failed to close queue", err)
	}

	delete(imqm.queues, taskID)
	imqm.metrics.ClearLastSeen(taskID)
	imqm.updateMetricsLocked()

	return nil
}

// CreateOrTap creates a new EventQueue for the task ID if one doesn't exist, or taps an existing one.
func (imqm *InMemoryQueueManager) CreateOrTap(taskID string) *EventQueue {
	if taskID == "" {
		return nil
	}

	// First try to get existing queue
	if queue := imqm.Get(taskID); queue != nil {
		return queue.Tap()
	}

	// Create new queue
	newQueue := imqm.createQueue(taskID)
	if err := imqm.Add(taskID, newQueue); err != nil {
		// If add fails, try to get existing queue and tap it
		if queue := imqm.Get(taskID); queue != nil {
			return queue.Tap()
		}
		return nil
	}

	return newQueue
}

// Exists checks if a queue exists for the given task ID.
func (imqm *InMemoryQueueManager) Exists(taskID string) bool {
	if taskID == "" {
		return false
	}

	imqm.mu.RLock()
	defer imqm.mu.RUnlock()

	_, exists := imqm.queues[taskID]
	return exists
}

// List returns all task IDs that have queues.
func (imqm *InMemoryQueueManager) List() []string {
	imqm.mu.RLock()
	defer imqm.mu.RUnlock()

	taskIDs := make([]string, 0, len(imqm.queues))
	for taskID := range imqm.queues {
		taskIDs = append(taskIDs, taskID)
	}

	sort.Strings(taskIDs)
	return taskIDs
}

// Count returns the number of queues managed by this manager.
func (imqm *InMemoryQueueManager) Count() int {
	imqm.mu.RLock()
	defer imqm.mu.RUnlock()
	return len(imqm.queues)
}

// CloseAll closes all queues managed by this manager.
func (imqm *InMemoryQueueManager) CloseAll() error {
	imqm.mu.Lock()
	defer imqm.mu.Unlock()

	var errors []error
	for taskID, queue := range imqm.queues {
		if err := queue.Close(); err != nil {
			errors = append(errors, NewQueueOperationError("close_all", taskID, "failed to close queue", err))
		}
	}

	// Clear all queues
	imqm.queues = make(map[string]*EventQueue)
	imqm.updateMetricsLocked()

	if len(errors) > 0 {
		return fmt.Errorf("failed to close %d queues: %v", len(errors), errors)
	}

	return nil
}

// String returns a string representation of the manager.
func (imqm *InMemoryQueueManager) String() string {
	imqm.mu.RLock()
	defer imqm.mu.RUnlock()

	stats := imqm.metrics.GetStats()
	return fmt.Sprintf("InMemoryQueueManager{queues: %d, active: %d, closed: %d, totalEvents: %d}",
		stats.TotalQueues, stats.ActiveQueues, stats.ClosedQueues, stats.TotalEvents)
}

// GetStats returns statistics about the queues.
func (imqm *InMemoryQueueManager) GetStats() QueueStats {
	return imqm.metrics.GetStats()
}

// GetQueueInfo returns information about a specific queue.
func (imqm *InMemoryQueueManager) GetQueueInfo(taskID string) (*QueueInfo, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	imqm.mu.RLock()
	queue, exists := imqm.queues[taskID]
	imqm.mu.RUnlock()

	if !exists {
		return nil, &NoTaskQueueError{TaskID: taskID}
	}

	lastSeen, _ := imqm.metrics.GetLastSeen(taskID)

	return &QueueInfo{
		TaskID:    taskID,
		Queue:     queue,
		CreatedAt: lastSeen,
		Closed:    queue.IsClosed(),
	}, nil
}

// GetAllQueueInfo returns information about all queues.
func (imqm *InMemoryQueueManager) GetAllQueueInfo() []*QueueInfo {
	imqm.mu.RLock()
	defer imqm.mu.RUnlock()

	infos := make([]*QueueInfo, 0, len(imqm.queues))
	for taskID, queue := range imqm.queues {
		lastSeen, _ := imqm.metrics.GetLastSeen(taskID)
		infos = append(infos, &QueueInfo{
			TaskID:    taskID,
			Queue:     queue,
			CreatedAt: lastSeen,
			Closed:    queue.IsClosed(),
		})
	}

	return infos
}

// Filter returns queues that match the given filter.
func (imqm *InMemoryQueueManager) Filter(filter QueueFilter) map[string]*EventQueue {
	if filter == nil {
		return make(map[string]*EventQueue)
	}

	imqm.mu.RLock()
	defer imqm.mu.RUnlock()

	result := make(map[string]*EventQueue)
	for taskID, queue := range imqm.queues {
		if filter(taskID, queue) {
			result[taskID] = queue
		}
	}

	return result
}

// FindQueues finds queues based on multiple criteria.
func (imqm *InMemoryQueueManager) FindQueues(filters ...QueueFilter) map[string]*EventQueue {
	if len(filters) == 0 {
		return make(map[string]*EventQueue)
	}

	imqm.mu.RLock()
	defer imqm.mu.RUnlock()

	result := make(map[string]*EventQueue)
	for taskID, queue := range imqm.queues {
		match := true
		for _, filter := range filters {
			if !filter(taskID, queue) {
				match = false
				break
			}
		}
		if match {
			result[taskID] = queue
		}
	}

	return result
}

// CountQueues counts queues that match the given filter.
func (imqm *InMemoryQueueManager) CountQueues(filter QueueFilter) int {
	if filter == nil {
		return 0
	}

	imqm.mu.RLock()
	defer imqm.mu.RUnlock()

	count := 0
	for taskID, queue := range imqm.queues {
		if filter(taskID, queue) {
			count++
		}
	}

	return count
}

// updateMetrics updates the internal metrics (must be called with lock held).
func (imqm *InMemoryQueueManager) updateMetrics() {
	imqm.mu.Lock()
	defer imqm.mu.Unlock()
	imqm.updateMetricsLocked()
}

// updateMetricsLocked updates the internal metrics (must be called with lock held).
func (imqm *InMemoryQueueManager) updateMetricsLocked() {
	stats := QueueStats{
		TotalQueues:   len(imqm.queues),
		ActiveQueues:  0,
		ClosedQueues:  0,
		TotalEvents:   0,
		TotalChildren: 0,
	}

	for _, queue := range imqm.queues {
		if queue.IsClosed() {
			stats.ClosedQueues++
		} else {
			stats.ActiveQueues++
		}
		stats.TotalEvents += queue.Len()
		stats.TotalChildren += queue.ChildCount()
	}

	imqm.metrics.UpdateStats(stats)
}

// Batch operations implementation

// BatchAdd adds multiple queues in a single operation.
func (imqm *InMemoryQueueManager) BatchAdd(operations []QueueOperation) []QueueOperationResult {
	results := make([]QueueOperationResult, len(operations))

	for i, op := range operations {
		err := imqm.Add(op.TaskID, op.Queue)
		results[i] = QueueOperationResult{
			TaskID: op.TaskID,
			Queue:  op.Queue,
			Error:  err,
		}
	}

	return results
}

// BatchGet retrieves multiple queues in a single operation.
func (imqm *InMemoryQueueManager) BatchGet(taskIDs []string) []QueueOperationResult {
	results := make([]QueueOperationResult, len(taskIDs))

	for i, taskID := range taskIDs {
		queue := imqm.Get(taskID)
		var err error
		if queue == nil {
			err = &NoTaskQueueError{TaskID: taskID}
		}
		results[i] = QueueOperationResult{
			TaskID: taskID,
			Queue:  queue,
			Error:  err,
		}
	}

	return results
}

// BatchClose closes multiple queues in a single operation.
func (imqm *InMemoryQueueManager) BatchClose(taskIDs []string) []QueueOperationResult {
	results := make([]QueueOperationResult, len(taskIDs))

	for i, taskID := range taskIDs {
		err := imqm.Close(taskID)
		results[i] = QueueOperationResult{
			TaskID: taskID,
			Queue:  nil,
			Error:  err,
		}
	}

	return results
}

// Async operations implementation

// AddAsync adds a queue asynchronously.
func (imqm *InMemoryQueueManager) AddAsync(ctx context.Context, taskID string, queue *EventQueue) <-chan error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		select {
		case ch <- imqm.Add(taskID, queue):
		case <-ctx.Done():
			ch <- ctx.Err()
		}
	}()
	return ch
}

// GetAsync retrieves a queue asynchronously.
func (imqm *InMemoryQueueManager) GetAsync(ctx context.Context, taskID string) <-chan *EventQueue {
	ch := make(chan *EventQueue, 1)
	go func() {
		defer close(ch)
		select {
		case ch <- imqm.Get(taskID):
		case <-ctx.Done():
			ch <- nil
		}
	}()
	return ch
}

// CloseAsync closes a queue asynchronously.
func (imqm *InMemoryQueueManager) CloseAsync(ctx context.Context, taskID string) <-chan error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		select {
		case ch <- imqm.Close(taskID):
		case <-ctx.Done():
			ch <- ctx.Err()
		}
	}()
	return ch
}

// Common queue filters

// ActiveQueueFilter returns queues that are not closed.
func ActiveQueueFilter(taskID string, queue *EventQueue) bool {
	return !queue.IsClosed()
}

// ClosedQueueFilter returns queues that are closed.
func ClosedQueueFilter(taskID string, queue *EventQueue) bool {
	return queue.IsClosed()
}

// NonEmptyQueueFilter returns queues that have events.
func NonEmptyQueueFilter(taskID string, queue *EventQueue) bool {
	return queue.Len() > 0
}

// TaskIDPrefixFilter returns a filter that matches task IDs with the given prefix.
func TaskIDPrefixFilter(prefix string) QueueFilter {
	return func(taskID string, queue *EventQueue) bool {
		return len(taskID) >= len(prefix) && taskID[:len(prefix)] == prefix
	}
}

// QueueCapacityFilter returns a filter that matches queues with capacity above the threshold.
func QueueCapacityFilter(minCapacity int) QueueFilter {
	return func(taskID string, queue *EventQueue) bool {
		return queue.Cap() >= minCapacity
	}
}
