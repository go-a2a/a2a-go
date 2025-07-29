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

package pyclient

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-a2a/a2a-go/internal/jsonrpc2"
	"github.com/go-a2a/a2a-go/transport"
)

// connectionManager manages the client's connection lifecycle.
type connectionManager struct {
	client    *Client
	session   *transport.ClientSession
	transport transport.Transport

	state   atomic.Int32 // ConnectionState
	stateMu sync.RWMutex
	stateCV *sync.Cond

	reconnectCh   chan struct{}
	reconnectStop chan struct{}

	lastError     error
	lastErrorTime time.Time
}

// newConnectionManager creates a new connection manager.
func newConnectionManager(client *Client, t transport.Transport) *connectionManager {
	cm := &connectionManager{
		client:        client,
		transport:     t,
		reconnectCh:   make(chan struct{}, 1),
		reconnectStop: make(chan struct{}),
	}
	cm.stateCV = sync.NewCond(&cm.stateMu)
	cm.setState(ConnectionStateDisconnected)
	return cm
}

// connect establishes a connection to the A2A agent.
func (cm *connectionManager) connect(ctx context.Context) error {
	cm.setState(ConnectionStateConnecting)

	// Create client binder
	binder := &clientBinder{
		client: cm.client,
		cm:     cm,
	}

	// Connect using transport
	session, err := transport.Connect(ctx, cm.transport, binder)
	if err != nil {
		cm.setError(err)
		cm.setState(ConnectionStateDisconnected)
		return &ConnectionError{
			Operation: "connect",
			URL:       cm.client.opts.baseURL,
			Err:       err,
		}
	}

	cm.stateMu.Lock()
	cm.session = session
	cm.stateMu.Unlock()

	cm.setState(ConnectionStateConnected)

	// Start auto-reconnect if enabled
	if cm.client.opts.enableAutoReconnect {
		go cm.autoReconnectLoop()
	}

	return nil
}

// disconnect closes the connection.
func (cm *connectionManager) disconnect() error {
	cm.setState(ConnectionStateClosed)

	// Stop auto-reconnect
	close(cm.reconnectStop)

	cm.stateMu.Lock()
	session := cm.session
	cm.session = nil
	cm.stateMu.Unlock()

	if session != nil {
		return session.Close()
	}

	return nil
}

// getSession returns the current session if connected.
func (cm *connectionManager) getSession() (*transport.ClientSession, error) {
	cm.stateMu.RLock()
	defer cm.stateMu.RUnlock()

	if cm.getState() != ConnectionStateConnected {
		return nil, ErrNotConnected
	}

	if cm.session == nil {
		return nil, ErrNotConnected
	}

	return cm.session, nil
}

// setState updates the connection state.
func (cm *connectionManager) setState(state ConnectionState) {
	oldState := ConnectionState(cm.state.Load())
	cm.state.Store(int32(state))

	// Notify state change callback
	if cm.client.opts.onConnectionStateChange != nil && oldState != state {
		cm.client.opts.onConnectionStateChange(oldState, state)
	}

	// Signal any waiters
	cm.stateCV.Broadcast()
}

// getState returns the current connection state.
func (cm *connectionManager) getState() ConnectionState {
	return ConnectionState(cm.state.Load())
}

// setError records the last error.
func (cm *connectionManager) setError(err error) {
	cm.stateMu.Lock()
	cm.lastError = err
	cm.lastErrorTime = time.Now()
	cm.stateMu.Unlock()
}

// waitForState waits for a specific connection state.
func (cm *connectionManager) waitForState(ctx context.Context, state ConnectionState) error {
	cm.stateMu.Lock()
	defer cm.stateMu.Unlock()

	for cm.getState() != state {
		// Check context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Wait for state change
		done := make(chan struct{})
		go func() {
			cm.stateMu.Lock()
			cm.stateCV.Wait()
			cm.stateMu.Unlock()
			close(done)
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			// Continue loop
		}
	}

	return nil
}

// triggerReconnect triggers a reconnection attempt.
func (cm *connectionManager) triggerReconnect() {
	select {
	case cm.reconnectCh <- struct{}{}:
	default:
		// Already triggered
	}
}

// autoReconnectLoop handles automatic reconnection.
func (cm *connectionManager) autoReconnectLoop() {
	backoff := cm.client.opts.retryConfig.InitialDelay

	for {
		select {
		case <-cm.reconnectStop:
			return

		case <-cm.reconnectCh:
			// Wait for backoff
			select {
			case <-time.After(backoff):
			case <-cm.reconnectStop:
				return
			}

			// Attempt reconnection
			cm.setState(ConnectionStateReconnecting)

			ctx, cancel := context.WithTimeout(context.Background(), cm.client.opts.connectTimeout)
			err := cm.connect(ctx)
			cancel()

			if err != nil {
				// Exponential backoff
				backoff = min(time.Duration(float64(backoff)*cm.client.opts.retryConfig.Multiplier), cm.client.opts.retryConfig.MaxDelay)

				// Trigger another attempt
				cm.triggerReconnect()
			} else {
				// Reset backoff on success
				backoff = cm.client.opts.retryConfig.InitialDelay
			}
		}
	}
}

// clientBinder implements transport.Binder for the client.
type clientBinder struct {
	client        *Client
	cm            *connectionManager
	clientSession *transport.ClientSession
}

// Bind implements transport.Binder.
func (b *clientBinder) Bind(conn *jsonrpc2.Connection) *transport.ClientSession {
	// For now, create a mock ClientSession
	// In a real implementation, this would properly integrate with the transport layer
	b.clientSession = &transport.ClientSession{
		Connection:   conn,
		Client:       nil, // Will be set later
		A2AConn:      nil, // Will be set later
		HTTPClient:   b.client.opts.httpClient,
		Interceptors: b.client.opts.interceptors,
	}
	return b.clientSession
}

// Disconnect implements transport.Binder.
func (b *clientBinder) Disconnect(h *transport.ClientSession) {
	// Handle disconnection
	if b.cm.getState() == ConnectionStateConnected {
		b.cm.setState(ConnectionStateDisconnected)

		// Trigger reconnect if auto-reconnect is enabled
		if b.client.opts.enableAutoReconnect {
			b.cm.triggerReconnect()
		}
	}
}

// clientHandler implements transport.Handler for the client.
type clientHandler struct {
	client *Client
	cm     *connectionManager
	conn   *jsonrpc2.Connection
}

// Handle implements transport.Handler.
func (h *clientHandler) Handle(ctx context.Context, req *jsonrpc2.Request) (any, error) {
	// Log incoming message if callback is set
	if h.client.opts.onMessage != nil {
		h.client.opts.onMessage("receive", req)
	}

	// For now, return not handled - the actual handling will be implemented
	// when we have the full client implementation
	return nil, jsonrpc2.ErrNotHandled
}

// SetConn implements transport.Handler.
func (h *clientHandler) SetConn(conn transport.Connection) {
	// Store the connection if needed
}
