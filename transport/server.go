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
	"sync"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/internal/jsonrpc2"
)

// Server provides methods for handling incoming requests from A2A clients.
type Server interface {
	Connect(ctx context.Context, t Transport) (*ServerSession, error)
	ServerSendingMethodHandler() any
	ServerReceivingMethodHandler() any

	SendMessage(ctx context.Context, ss *ServerSession, params *a2a.MessageSendParams) (a2a.MessageOrTask, error)
	SendStreamMessage(ctx context.Context, ss *ServerSession, params *a2a.MessageSendParams) (a2a.SendStreamingMessageResponse, error)
	GetTask(ctx context.Context, ss *ServerSession, params *a2a.TaskQueryParams) (*a2a.Task, error)
	CancelTask(ctx context.Context, ss *ServerSession, params *a2a.TaskIDParams) (*a2a.Task, error)
	SetTasksPushNotificationConfig(ctx context.Context, ss *ServerSession, params *a2a.TaskPushNotificationConfig) (*a2a.TaskPushNotificationConfig, error)
	GetTasksPushNotificationConfig(ctx context.Context, ss *ServerSession, params *a2a.GetTaskPushNotificationConfigParams) (*a2a.TaskPushNotificationConfig, error)
	ListTasksPushNotificationConfig(ctx context.Context, ss *ServerSession, params *a2a.ListTaskPushNotificationConfigParams) (a2a.TaskPushNotificationConfigs, error)
	DeleteTasksPushNotificationConfig(ctx context.Context, ss *ServerSession, params *a2a.DeleteTaskPushNotificationConfigParams) (*a2a.EmptyResult, error)
	ResubscribeTasks(ctx context.Context, ss *ServerSession, params *a2a.TaskIDParams) (a2a.SendStreamingMessageResponse, error)
}

// serverMethodInfo maps from the RPC method name to server [methodInfo].
var serverMethodInfo = map[a2a.Method]methodInfo{
	a2a.MethodMessageSend:                       newMethodInfo(serverMethod((Server).SendMessage)),
	a2a.MethodMessageStream:                     newMethodInfo(serverMethod((Server).SendStreamMessage)),
	a2a.MethodTasksGet:                          newMethodInfo(serverMethod((Server).GetTask)),
	a2a.MethodTasksCancel:                       newMethodInfo(serverMethod((Server).CancelTask)),
	a2a.MethodTasksPushNotificationConfigSet:    newMethodInfo(serverMethod((Server).SetTasksPushNotificationConfig)),
	a2a.MethodTasksPushNotificationConfigGet:    newMethodInfo(serverMethod((Server).GetTasksPushNotificationConfig)),
	a2a.MethodTasksPushNotificationConfigList:   newMethodInfo(serverMethod((Server).GetTasksPushNotificationConfig)),
	a2a.MethodTasksPushNotificationConfigDelete: newMethodInfo(serverMethod((Server).DeleteTasksPushNotificationConfig)),
	a2a.MethodTasksResubscribe:                  newMethodInfo(serverMethod((Server).ResubscribeTasks)),
}

// A ServerSession is a logical connection from a single A2A client. Its
// methods can be used to send requests or notifications to the client. Create
// a session by calling [Server.Connect].
//
// Call [ServerSession.Close] to close the connection, or await client
// termination with [ServerSession.Wait].
type ServerSession struct {
	Server     Server
	Connection *jsonrpc2.Connection
	A2aConn    Connection
	Mu         sync.Mutex
}

var _ Handler = (*ServerSession)(nil)

// handle invokes the method described by the given JSON RPC request.
//
// Handle implements [Handler].
func (ss *ServerSession) Handle(ctx context.Context, req *jsonrpc2.Request) (any, error) {
	// For the streamable transport, we need the request ID to correlate
	// server->client calls and notifications to the incoming request from which
	// they originated. See [idContextKey] for details.
	ctx = context.WithValue(ctx, idContextKey{}, req.ID)
	return handleReceive(ctx, ss, req)
}

// SetConn implements [Handler].
func (ss *ServerSession) SetConn(c Connection) {
	ss.A2aConn = c
}

// ID implements [Session].
func (ss *ServerSession) ID() string {
	if ss.A2aConn == nil {
		return ""
	}
	return ss.A2aConn.SessionID()
}

// sendingMethodInfo implements [Session].
func (ss *ServerSession) sendingMethodInfo() map[a2a.Method]methodInfo {
	return clientMethodInfo
}

// receivingMethodInfo implements [Session].
func (ss *ServerSession) receivingMethodInfo() map[a2a.Method]methodInfo {
	return serverMethodInfo
}

// sendingMethodHandler implements [Session].
func (ss *ServerSession) sendingMethodHandler() methodHandler {
	ss.Mu.Lock()
	h := ss.Server.ServerSendingMethodHandler()
	ss.Mu.Unlock()
	return h
}

// receivingMethodHandler implements [Session].
func (ss *ServerSession) receivingMethodHandler() methodHandler {
	ss.Mu.Lock()
	h := ss.Server.ServerReceivingMethodHandler()
	ss.Mu.Unlock()
	return h
}

// Conn implements [Session].
func (ss *ServerSession) Conn() *jsonrpc2.Connection { return ss.Connection }

// Close performs a graceful shutdown of the connection, preventing new
// requests from being handled, and waiting for ongoing requests to return.
// Close then terminates the connection.
//
// Close implements [Session].
func (ss *ServerSession) Close() error {
	return ss.Connection.Close()
}

// Wait waits for the connection to be closed by the client.
//
// Wait implements [Session].
func (ss *ServerSession) Wait() error {
	return ss.Connection.Wait()
}
