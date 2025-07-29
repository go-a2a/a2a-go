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

package transport

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/internal/jsonrpc2"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

type Client interface {
	ClientSendingMethodHandler() any
	ClientReceivingMethodHandler() any
}

// clientMethodInfo maps from the RPC method name to client [methodInfo].
var clientMethodInfo = map[a2a.Method]methodInfo{}

// Interceptor defines a middleware function that can intercept and modify requests/responses.
type Interceptor func(ctx context.Context, req *http.Request, invoker Invoker) (*http.Response, error)

// Invoker represents the next handler in the interceptor chain.
type Invoker func(ctx context.Context, req *http.Request) (*http.Response, error)

// chainInterceptors chains multiple interceptors together.
func chainInterceptors(interceptors []Interceptor, invoker Invoker) Invoker {
	if len(interceptors) == 0 {
		return invoker
	}

	// Build the chain from right to left
	for i := len(interceptors) - 1; i >= 0; i-- {
		interceptor := interceptors[i]
		next := invoker
		invoker = func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return interceptor(ctx, req, next)
		}
	}

	return invoker
}

// ClientSession is a logical connection with an A2A server. Its
// methods can be used to send requests or notifications to the server. Create
// a session by calling [Client.Connect].
//
// Call [ClientSession.Close] to close the connection, or await client
// termination with [ServerSession.Wait].
type ClientSession struct {
	Connection   *jsonrpc2.Connection
	Client       Client
	A2AConn      Connection
	HTTPClient   *http.Client
	Interceptors []Interceptor
	mu           sync.Mutex
}

var _ Handler = (*ClientSession)(nil)

// Handle implements [Handler].
func (cs *ClientSession) Handle(ctx context.Context, req *jsonrpc2.Request) (any, error) {
	return handleReceive(ctx, cs, req)
}

// SetConn implements [Handler].
func (cs *ClientSession) SetConn(c Connection) {
	cs.A2AConn = c
}

// ID implements [Session].
func (cs *ClientSession) ID() string {
	if cs.A2AConn == nil {
		return ""
	}
	return cs.A2AConn.SessionID()
}

// Close performs a graceful close of the connection, preventing new requests
// from being handled, and waiting for ongoing requests to return. Close then
// terminates the connection.
//
// Close implements [Session].
func (cs *ClientSession) Close() error {
	return cs.Connection.Close()
}

// Wait waits for the connection to be closed by the server.
// Generally, clients should be responsible for closing the connection.
//
// Wait implements [Session].
func (cs *ClientSession) Wait() error {
	return cs.Connection.Wait()
}

// sendingMethodInfo implements [Session].
func (cs *ClientSession) sendingMethodInfo() map[a2a.Method]methodInfo {
	return serverMethodInfo
}

// receivingMethodInfo implements [Session].
func (cs *ClientSession) receivingMethodInfo() map[a2a.Method]methodInfo {
	return clientMethodInfo
}

// sendingMethodHandler implements [Session].
func (cs *ClientSession) sendingMethodHandler() methodHandler {
	cs.mu.Lock()
	h := cs.Client.ClientSendingMethodHandler()
	cs.mu.Unlock()
	return h
}

// receivingMethodHandler implements [Session].
func (cs *ClientSession) receivingMethodHandler() methodHandler {
	cs.mu.Lock()
	h := cs.Client.ClientReceivingMethodHandler()
	cs.mu.Unlock()
	return h
}

// Conn implements [Session].
func (cs *ClientSession) Conn() *jsonrpc2.Connection { return cs.Connection }

// GetAgentCard
func (cs *ClientSession) GetAgentCard(ctx context.Context, baseURL string) (*a2a.AgentCard, error) {
	targetURL, err := url.JoinPath(baseURL, a2a.AgentCardWellKnownPath)
	if err != nil {
		return nil, fmt.Errorf("join %q and %q path: %w", baseURL, a2a.AgentCardWellKnownPath, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Apply interceptors
	invoker := func(ctx context.Context, req *http.Request) (*http.Response, error) {
		req = req.WithContext(ctx)
		return cs.HTTPClient.Do(req)
	}
	invoker = chainInterceptors(cs.Interceptors, invoker)

	resp, err := invoker(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fetch agent card: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch agent card from %s", targetURL)
	}

	var agentCard a2a.AgentCard
	dec := jsontext.NewDecoder(resp.Body)
	if err := json.UnmarshalDecode(dec, &agentCard, json.DefaultOptionsV2()); err != nil {
		return nil, fmt.Errorf("decode agent card: %w", err)
	}

	return &agentCard, nil
}

// SendMessage sends a message to an agent to initiate a new interaction or to continue an existing one.
func (cs *ClientSession) SendMessage(ctx context.Context, req *a2a.MessageSendParams) (a2a.MessageOrTask, error) {
	return handleSend[a2a.MessageOrTask](ctx, cs, a2a.MethodMessageSend, orZero[a2a.Params](req))
}

// SendStreamMessage sends a message to an agent to initiate/continue a task AND subscribes the client to real-time updates for that task via Server-Sent Events (SSE).
func (cs *ClientSession) SendStreamMessage(ctx context.Context, req *a2a.MessageSendParams) (a2a.SendStreamingMessageResponse, error) {
	return handleSend[a2a.SendStreamingMessageResponse](ctx, cs, a2a.MethodMessageStream, orZero[a2a.Params](req))
}

// GetTask retrieves the current state (including status, artifacts, and optionally history) of a previously initiated task.
func (cs *ClientSession) GetTask(ctx context.Context, req *a2a.TaskQueryParams) (*a2a.Task, error) {
	return handleSend[*a2a.Task](ctx, cs, a2a.MethodTasksGet, orZero[a2a.Params](req))
}

// CancelTask requests the cancellation of an ongoing task.
func (cs *ClientSession) CancelTask(ctx context.Context, req *a2a.TaskIDParams) (*a2a.Task, error) {
	return handleSend[*a2a.Task](ctx, cs, a2a.MethodTasksCancel, orZero[a2a.Params](req))
}

// SetTasksPushNotificationConfig sets or updates the push notification configuration for a specified task.
func (cs *ClientSession) SetTasksPushNotificationConfig(ctx context.Context, req *a2a.TaskPushNotificationConfig) (*a2a.TaskPushNotificationConfig, error) {
	return handleSend[*a2a.TaskPushNotificationConfig](ctx, cs, a2a.MethodTasksPushNotificationConfigSet, orZero[a2a.Params](req))
}

// GetTasksPushNotificationConfig retrieves the current push notification configuration for a specified task.
func (cs *ClientSession) GetTasksPushNotificationConfig(ctx context.Context, req *a2a.GetTaskPushNotificationConfigParams) (*a2a.TaskPushNotificationConfig, error) {
	return handleSend[*a2a.TaskPushNotificationConfig](ctx, cs, a2a.MethodTasksPushNotificationConfigGet, orZero[a2a.Params](req))
}

// ListTasksPushNotificationConfig retrieves the associated push notification configurations for a specified task.
func (cs *ClientSession) ListTasksPushNotificationConfig(ctx context.Context, req *a2a.ListTaskPushNotificationConfigParams) (a2a.TaskPushNotificationConfigs, error) {
	return handleSend[a2a.TaskPushNotificationConfigs](ctx, cs, a2a.MethodTasksPushNotificationConfigList, orZero[a2a.Params](req))
}

// DeleteTasksPushNotificationConfig deletes an associated push notification configuration for a task.
func (cs *ClientSession) DeleteTasksPushNotificationConfig(ctx context.Context, req *a2a.DeleteTaskPushNotificationConfigParams) (*a2a.EmptyResult, error) {
	return handleSend[*a2a.EmptyResult](ctx, cs, a2a.MethodTasksPushNotificationConfigDelete, orZero[a2a.Params](req))
}

// ResubscribeTasks allows a client to reconnect to an SSE stream for an ongoing task after a previous connection (from message/stream or an earlier tasks/resubscribe) was interrupted.
func (cs *ClientSession) ResubscribeTasks(ctx context.Context, req *a2a.TaskIDParams) (a2a.SendStreamingMessageResponse, error) {
	return handleSend[a2a.SendStreamingMessageResponse](ctx, cs, a2a.MethodTasksResubscribe, orZero[a2a.Params](req))
}

// AuthenticatedExtendedCard retrieves a potentially more detailed version of the [a2a.AgentCard] after the client has authenticated.
func (cs *ClientSession) AuthenticatedExtendedCard(ctx context.Context, req *a2a.EmptyParams) (*a2a.AgentCard, error) {
	return handleSend[*a2a.AgentCard](ctx, cs, a2a.MethodAgentAuthenticatedExtendedCard, orZero[a2a.Params](req))
}
