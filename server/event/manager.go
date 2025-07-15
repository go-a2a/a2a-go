// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"fmt"
	"sync"
)

// QueueManager defines the interface for managing event queues.
// It provides methods to create, retrieve, and manage EventQueues for different task IDs.
type QueueManager interface {
	// Add creates a new event queue for the given task ID.
	// Returns TaskQueueExistsError if the queue already exists.
	Add(taskID string, queue *EventQueue) error

	// Get retrieves an EventQueue for the given task ID.
	// Returns nil if the queue does not exist.
	Get(taskID string) *EventQueue

	// Tap creates a child queue for the given task ID.
	// Returns nil if the parent queue does not exist.
	Tap(taskID string) *EventQueue

	// Close closes and removes the event queue for the given task ID.
	// Returns NoTaskQueueError if the queue does not exist.
	Close(taskID string) error

	// CreateOrTap creates a new EventQueue for the task ID if one doesn't exist,
	// or taps an existing one.
	CreateOrTap(taskID string) *EventQueue

	// Exists checks if a queue exists for the given task ID.
	Exists(taskID string) bool

	// List returns all task IDs that have queues.
	List() []string

	// Count returns the number of queues managed by this manager.
	Count() int

	// CloseAll closes all queues managed by this manager.
	CloseAll() error

	// String returns a string representation of the manager.
	String() string
}

// QueueManagerConfig holds configuration for queue managers.
type QueueManagerConfig struct {
	// DefaultMaxQueueSize is the default maximum size for new queues.
	DefaultMaxQueueSize int

	// AutoCreateQueues determines whether to automatically create queues
	// when they don't exist during Get operations.
	AutoCreateQueues bool
}

// DefaultQueueManagerConfig returns a default configuration for queue managers.
func DefaultQueueManagerConfig() *QueueManagerConfig {
	return &QueueManagerConfig{
		DefaultMaxQueueSize: DefaultMaxQueueSize,
		AutoCreateQueues:    false,
	}
}

// QueueInfo holds information about a queue.
type QueueInfo struct {
	TaskID    string
	Queue     *EventQueue
	CreatedAt int64
	Closed    bool
}

// QueueStats holds statistics about queues.
type QueueStats struct {
	TotalQueues   int
	ActiveQueues  int
	ClosedQueues  int
	TotalEvents   int
	TotalChildren int
}

// BaseQueueManager provides common functionality for queue managers.
type BaseQueueManager struct {
	config *QueueManagerConfig
	mu     sync.RWMutex
}

// NewBaseQueueManager creates a new BaseQueueManager with the given configuration.
func NewBaseQueueManager(config *QueueManagerConfig) *BaseQueueManager {
	if config == nil {
		config = DefaultQueueManagerConfig()
	}

	return &BaseQueueManager{
		config: config,
	}
}

// Config returns the configuration for this queue manager.
func (bqm *BaseQueueManager) Config() *QueueManagerConfig {
	bqm.mu.RLock()
	defer bqm.mu.RUnlock()
	return bqm.config
}

// createQueue creates a new EventQueue with the default configuration.
func (bqm *BaseQueueManager) createQueue(taskID string) *EventQueue {
	return NewEventQueueWithName(
		fmt.Sprintf("TaskQueue-%s", taskID),
		bqm.config.DefaultMaxQueueSize,
	)
}

// QueueManagerMetrics provides metrics for queue managers.
type QueueManagerMetrics struct {
	mu       sync.RWMutex
	stats    QueueStats
	lastSeen map[string]int64
}

// NewQueueManagerMetrics creates a new QueueManagerMetrics.
func NewQueueManagerMetrics() *QueueManagerMetrics {
	return &QueueManagerMetrics{
		stats:    QueueStats{},
		lastSeen: make(map[string]int64),
	}
}

// GetStats returns the current statistics.
func (qmm *QueueManagerMetrics) GetStats() QueueStats {
	qmm.mu.RLock()
	defer qmm.mu.RUnlock()
	return qmm.stats
}

// UpdateStats updates the statistics.
func (qmm *QueueManagerMetrics) UpdateStats(stats QueueStats) {
	qmm.mu.Lock()
	defer qmm.mu.Unlock()
	qmm.stats = stats
}

// RecordQueueAccess records that a queue was accessed.
func (qmm *QueueManagerMetrics) RecordQueueAccess(taskID string, timestamp int64) {
	qmm.mu.Lock()
	defer qmm.mu.Unlock()
	qmm.lastSeen[taskID] = timestamp
}

// GetLastSeen returns the last seen timestamp for a task ID.
func (qmm *QueueManagerMetrics) GetLastSeen(taskID string) (int64, bool) {
	qmm.mu.RLock()
	defer qmm.mu.RUnlock()
	timestamp, exists := qmm.lastSeen[taskID]
	return timestamp, exists
}

// ClearLastSeen removes the last seen timestamp for a task ID.
func (qmm *QueueManagerMetrics) ClearLastSeen(taskID string) {
	qmm.mu.Lock()
	defer qmm.mu.Unlock()
	delete(qmm.lastSeen, taskID)
}

// QueueManagerOption is a function that configures a queue manager.
type QueueManagerOption func(*QueueManagerConfig)

// WithDefaultMaxQueueSize sets the default maximum queue size.
func WithDefaultMaxQueueSize(size int) QueueManagerOption {
	return func(config *QueueManagerConfig) {
		config.DefaultMaxQueueSize = size
	}
}

// WithAutoCreateQueues sets whether to automatically create queues.
func WithAutoCreateQueues(autoCreate bool) QueueManagerOption {
	return func(config *QueueManagerConfig) {
		config.AutoCreateQueues = autoCreate
	}
}

// ApplyOptions applies the given options to the configuration.
func ApplyOptions(config *QueueManagerConfig, options ...QueueManagerOption) {
	for _, option := range options {
		option(config)
	}
}

// QueueOperation represents a queue operation for batch processing.
type QueueOperation struct {
	Type   string // "add", "get", "tap", "close"
	TaskID string
	Queue  *EventQueue
}

// QueueOperationResult represents the result of a queue operation.
type QueueOperationResult struct {
	TaskID string
	Queue  *EventQueue
	Error  error
}

// BatchQueueManager provides batch operations for queue management.
type BatchQueueManager interface {
	QueueManager

	// BatchAdd adds multiple queues in a single operation.
	BatchAdd(operations []QueueOperation) []QueueOperationResult

	// BatchGet retrieves multiple queues in a single operation.
	BatchGet(taskIDs []string) []QueueOperationResult

	// BatchClose closes multiple queues in a single operation.
	BatchClose(taskIDs []string) []QueueOperationResult
}

var _ QueueManager = (BatchQueueManager)(nil)

// AsyncQueueManager provides asynchronous operations for queue management.
type AsyncQueueManager interface {
	QueueManager

	// AddAsync adds a queue asynchronously.
	AddAsync(ctx context.Context, taskID string, queue *EventQueue) <-chan error

	// GetAsync retrieves a queue asynchronously.
	GetAsync(ctx context.Context, taskID string) <-chan *EventQueue

	// CloseAsync closes a queue asynchronously.
	CloseAsync(ctx context.Context, taskID string) <-chan error
}

var _ QueueManager = (AsyncQueueManager)(nil)

// QueueFilter is a function that filters queues based on some criteria.
type QueueFilter func(taskID string, queue *EventQueue) bool

// FilterableQueueManager provides filtering capabilities for queue management.
type FilterableQueueManager interface {
	QueueManager

	// Filter returns queues that match the given filter.
	Filter(filter QueueFilter) map[string]*EventQueue

	// FindQueues finds queues based on multiple criteria.
	FindQueues(filters ...QueueFilter) map[string]*EventQueue

	// CountQueues counts queues that match the given filter.
	CountQueues(filter QueueFilter) int
}

var _ QueueManager = (FilterableQueueManager)(nil)
