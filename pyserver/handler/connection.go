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

package handler

import (
	"context"
	"sync"
	"time"
)

// ConnectionManager manages active connections and sessions.
type ConnectionManager interface {
	// Register registers a new connection.
	Register(conn Connection) error

	// Unregister removes a connection.
	Unregister(connID string) error

	// Get retrieves a connection by ID.
	Get(connID string) (Connection, bool)

	// GetBySession retrieves connections for a session.
	GetBySession(sessionID string) []Connection

	// Broadcast sends a message to all connections.
	Broadcast(ctx context.Context, message any) error

	// BroadcastToSession sends a message to all connections in a session.
	BroadcastToSession(ctx context.Context, sessionID string, message any) error

	// Count returns the number of active connections.
	Count() int

	// Close closes all connections and cleans up.
	Close() error
}

// Connection represents an active connection.
type Connection interface {
	// ID returns the unique connection identifier.
	ID() string

	// SessionID returns the associated session ID.
	SessionID() string

	// Send sends a message through the connection.
	Send(ctx context.Context, message any) error

	// Close closes the connection.
	Close() error

	// IsAlive checks if the connection is still active.
	IsAlive() bool
}

// connectionManager is the default implementation of ConnectionManager.
type connectionManager struct {
	mu          sync.RWMutex
	connections map[string]Connection
	sessions    map[string][]string // sessionID -> []connectionID
	closed      bool
}

// NewConnectionManager creates a new connection manager.
func NewConnectionManager() ConnectionManager {
	return &connectionManager{
		connections: make(map[string]Connection),
		sessions:    make(map[string][]string),
	}
}

// Register registers a new connection.
func (m *connectionManager) Register(conn Connection) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrConnectionManagerClosed
	}

	connID := conn.ID()
	sessionID := conn.SessionID()

	// Store connection
	m.connections[connID] = conn

	// Update session mapping
	if sessionID != "" {
		m.sessions[sessionID] = append(m.sessions[sessionID], connID)
	}

	return nil
}

// Unregister removes a connection.
func (m *connectionManager) Unregister(connID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, exists := m.connections[connID]
	if !exists {
		return nil
	}

	// Remove from connections
	delete(m.connections, connID)

	// Remove from session mapping
	sessionID := conn.SessionID()
	if sessionID != "" {
		conns := m.sessions[sessionID]
		for i, id := range conns {
			if id == connID {
				m.sessions[sessionID] = append(conns[:i], conns[i+1:]...)
				break
			}
		}
		// Clean up empty session
		if len(m.sessions[sessionID]) == 0 {
			delete(m.sessions, sessionID)
		}
	}

	// Close the connection
	return conn.Close()
}

// Get retrieves a connection by ID.
func (m *connectionManager) Get(connID string) (Connection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, exists := m.connections[connID]
	return conn, exists
}

// GetBySession retrieves connections for a session.
func (m *connectionManager) GetBySession(sessionID string) []Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	connIDs, exists := m.sessions[sessionID]
	if !exists {
		return nil
	}

	var connections []Connection
	for _, connID := range connIDs {
		if conn, exists := m.connections[connID]; exists && conn.IsAlive() {
			connections = append(connections, conn)
		}
	}

	return connections
}

// Broadcast sends a message to all connections.
func (m *connectionManager) Broadcast(ctx context.Context, message any) error {
	m.mu.RLock()
	connections := make([]Connection, 0, len(m.connections))
	for _, conn := range m.connections {
		if conn.IsAlive() {
			connections = append(connections, conn)
		}
	}
	m.mu.RUnlock()

	// Send to all connections concurrently
	var wg sync.WaitGroup
	for _, conn := range connections {
		wg.Add(1)
		go func(c Connection) {
			defer wg.Done()
			// Ignore individual send errors during broadcast
			_ = c.Send(ctx, message)
		}(conn)
	}
	wg.Wait()

	return nil
}

// BroadcastToSession sends a message to all connections in a session.
func (m *connectionManager) BroadcastToSession(ctx context.Context, sessionID string, message any) error {
	connections := m.GetBySession(sessionID)
	if len(connections) == 0 {
		return ErrSessionNotFound
	}

	// Send to all session connections concurrently
	var wg sync.WaitGroup
	for _, conn := range connections {
		wg.Add(1)
		go func(c Connection) {
			defer wg.Done()
			// Ignore individual send errors during broadcast
			_ = c.Send(ctx, message)
		}(conn)
	}
	wg.Wait()

	return nil
}

// Count returns the number of active connections.
func (m *connectionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}

// Close closes all connections and cleans up.
func (m *connectionManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}

	m.closed = true

	// Close all connections
	for _, conn := range m.connections {
		conn.Close()
	}

	// Clear maps
	m.connections = nil
	m.sessions = nil

	return nil
}

// ErrConnectionManagerClosed is returned when operations are attempted on a closed manager.
var ErrConnectionManagerClosed = NewHandlerError(500, "connection manager closed")

// SessionStore manages session persistence.
type SessionStore interface {
	// Create creates a new session.
	Create(ctx context.Context) (*Session, error)

	// Get retrieves a session by ID.
	Get(ctx context.Context, sessionID string) (*Session, error)

	// Update updates a session.
	Update(ctx context.Context, session *Session) error

	// Delete removes a session.
	Delete(ctx context.Context, sessionID string) error

	// Cleanup removes expired sessions.
	Cleanup(ctx context.Context) error
}

// memorySessionStore is an in-memory session store.
type memorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*sessionData
	ttl      time.Duration
}

type sessionData struct {
	session   *Session
	createdAt time.Time
	updatedAt time.Time
}

// NewMemorySessionStore creates a new in-memory session store.
func NewMemorySessionStore(ttl time.Duration) SessionStore {
	store := &memorySessionStore{
		sessions: make(map[string]*sessionData),
		ttl:      ttl,
	}

	// Start cleanup routine
	go store.cleanupRoutine()

	return store
}

// Create creates a new session.
func (s *memorySessionStore) Create(ctx context.Context) (*Session, error) {
	session := &Session{
		ID:       generateSessionID(),
		Metadata: make(map[string]any),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = &sessionData{
		session:   session,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}

	return session, nil
}

// Get retrieves a session by ID.
func (s *memorySessionStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	// Check if expired
	if time.Since(data.updatedAt) > s.ttl {
		return nil, ErrSessionNotFound
	}

	return data.session, nil
}

// Update updates a session.
func (s *memorySessionStore) Update(ctx context.Context, session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.sessions[session.ID]
	if !exists {
		return ErrSessionNotFound
	}

	data.session = session
	data.updatedAt = time.Now()

	return nil
}

// Delete removes a session.
func (s *memorySessionStore) Delete(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	return nil
}

// Cleanup removes expired sessions.
func (s *memorySessionStore) Cleanup(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, data := range s.sessions {
		if now.Sub(data.updatedAt) > s.ttl {
			delete(s.sessions, id)
		}
	}

	return nil
}

// cleanupRoutine periodically cleans up expired sessions.
func (s *memorySessionStore) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.Cleanup(context.Background())
	}
}

// generateSessionID generates a unique session ID.
func generateSessionID() string {
	// Simple implementation; production should use crypto/rand
	return "session-" + time.Now().Format("20060102150405.000000")
}
