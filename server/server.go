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

package server

import (
	"context"
	"slices"
	"sync"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/internal/jsonrpc2"
	"github.com/go-a2a/a2a-go/transport"
)

// A Server is an instance of an A2A server.
//
// Servers expose server-side A2A features, which can serve one or more A2A
// sessions by using [Server.Start] or [Server.Run].
type Server struct {
	// fixed at creation
	opts ServerOptions

	mu                     sync.Mutex
	sessions               []*transport.ServerSession
	sendingMethodHandler   transport.MethodHandler[*transport.ServerSession]
	receivingMethodHandler transport.MethodHandler[*transport.ServerSession]
}

var (
	_ transport.Server                           = (*Server)(nil)
	_ transport.Binder[*transport.ServerSession] = (*Server)(nil)
)

// ServerOptions is used to configure behavior of the server.
type ServerOptions struct {
	// Optional instructions for connected clients.
	Instructions string

	SendMessageHandler                       func(ctx context.Context, ss *transport.ServerSession, params *a2a.MessageSendParams) (a2a.MessageOrTask, error)
	SendStreamMessageHandler                 func(ctx context.Context, ss *transport.ServerSession, params *a2a.MessageSendParams) (a2a.SendStreamingMessageResponse, error)
	GetTaskHandler                           func(ctx context.Context, ss *transport.ServerSession, params *a2a.TaskQueryParams) (*a2a.Task, error)
	CancelTaskHandler                        func(ctx context.Context, ss *transport.ServerSession, params *a2a.TaskIDParams) (*a2a.Task, error)
	SetTasksPushNotificationConfigHandler    func(ctx context.Context, ss *transport.ServerSession, params *a2a.TaskPushNotificationConfig) (*a2a.TaskPushNotificationConfig, error)
	GetTasksPushNotificationConfigHandler    func(ctx context.Context, ss *transport.ServerSession, params *a2a.GetTaskPushNotificationConfigParams) (*a2a.TaskPushNotificationConfig, error)
	ListTasksPushNotificationConfigHandler   func(ctx context.Context, ss *transport.ServerSession, params *a2a.ListTaskPushNotificationConfigParams) (a2a.TaskPushNotificationConfigs, error)
	DeleteTasksPushNotificationConfigHandler func(ctx context.Context, ss *transport.ServerSession, params *a2a.DeleteTaskPushNotificationConfigParams) (*a2a.EmptyResult, error)
	ResubscribeTasksHandler                  func(ctx context.Context, ss *transport.ServerSession, params *a2a.TaskIDParams) (a2a.SendStreamingMessageResponse, error)
}

// NewServer creates a new A2A server. The resulting server has no features:
// add features using the various Server.AddXXX methods, and the [AddTool] function.
//
// The server can be connected to one or more A2A clients using [Server.Start]
// or [Server.Run].
//
// The first argument must not be nil.
//
// If non-nil, the provided options are used to configure the server.
func NewServer(opts *ServerOptions) *Server {
	if opts == nil {
		opts = new(ServerOptions)
	}

	return &Server{
		opts:                   *opts,
		sendingMethodHandler:   transport.DefaultSendingMethodHandler[*transport.ServerSession],
		receivingMethodHandler: transport.DefaultReceivingMethodHandler[*transport.ServerSession],
	}
}

// Run runs the server over the given transport, which must be persistent.
//
// Run blocks until the client terminates the connection or the provided
// context is cancelled. If the context is cancelled, Run closes the connection.
//
// If tools have been added to the server before this call, then the server will
// advertise the capability for tools, including the ability to send list-changed notifications.
// If no tools have been added, the server will not have the tool capability.
// The same goes for other features like prompts and resources.
func (s *Server) Run(ctx context.Context, t transport.Transport) error {
	ss, err := s.Connect(ctx, t)
	if err != nil {
		return err
	}

	ssClosed := make(chan error)
	go func() {
		ssClosed <- ss.Wait()
	}()

	select {
	case <-ctx.Done():
		ss.Close()
		return ctx.Err()
	case err := <-ssClosed:
		return err
	}
}

// bind implements the binder[*ServerSession] interface, so that Servers can
// be connected using [connect].
func (s *Server) Bind(conn *jsonrpc2.Connection) *transport.ServerSession {
	ss := &transport.ServerSession{
		Server:     s,
		Connection: conn,
	}
	s.mu.Lock()
	s.sessions = append(s.sessions, ss)
	s.mu.Unlock()
	return ss
}

// disconnect implements the binder[*ServerSession] interface, so that
// Servers can be connected using [connect].
func (s *Server) Disconnect(cc *transport.ServerSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions = slices.DeleteFunc(s.sessions, func(cc2 *transport.ServerSession) bool {
		return cc2 == cc
	})
}

// Connect connects the A2A server over the given transport and starts handling
// messages.
//
// It returns a connection object that may be used to terminate the connection
// (with [Connection.Close]), or await client termination (with
// [Connection.Wait]).
func (s *Server) Connect(ctx context.Context, t transport.Transport) (*transport.ServerSession, error) {
	return transport.Connect(ctx, t, s)
}

// SendingMethodHandler returns the current sending method handler.
//
// SendingMethodHandler implements [transport.Client].
func (s *Server) ServerSendingMethodHandler() any {
	return s.sendingMethodHandler
}

// ReceivingMethodHandler returns the current receiving method handler.
//
// ReceivingMethodHandler implements [transport.Client].
func (s *Server) ServerReceivingMethodHandler() any {
	return s.receivingMethodHandler
}

// AddSendingMiddleware wraps the current sending method handler using the provided
// middleware. Middleware is applied from right to left, so that the first one is
// executed first.
//
// For example, AddSendingMiddleware(m1, m2, m3) augments the method handler as
// m1(m2(m3(handler))).
//
// Sending middleware is called when a request is sent. It is useful for tasks
// such as tracing, metrics, and adding progress tokens.
func (s *Server) AddSendingMiddleware(middleware ...transport.Middleware[*transport.ServerSession]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	transport.AddMiddleware(&s.sendingMethodHandler, middleware)
}

// AddReceivingMiddleware wraps the current receiving method handler using
// the provided middleware. Middleware is applied from right to left, so that the
// first one is executed first.
//
// For example, AddReceivingMiddleware(m1, m2, m3) augments the method handler as
// m1(m2(m3(handler))).
//
// Receiving middleware is called when a request is received. It is useful for tasks
// such as authentication, request logging and metrics.
func (s *Server) AddReceivingMiddleware(middleware ...transport.Middleware[*transport.ServerSession]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	transport.AddMiddleware(&s.receivingMethodHandler, middleware)
}

func (s *Server) SendMessage(ctx context.Context, ss *transport.ServerSession, params *a2a.MessageSendParams) (a2a.MessageOrTask, error) {
	if s.opts.SendMessageHandler == nil {
		return nil, jsonrpc2.ErrMethodNotFound
	}
	res, err := s.opts.SendMessageHandler(ctx, ss, params)

	return res, err
}

func (s *Server) SendStreamMessage(ctx context.Context, ss *transport.ServerSession, params *a2a.MessageSendParams) (a2a.SendStreamingMessageResponse, error) {
	if s.opts.SendStreamMessageHandler == nil {
		return nil, jsonrpc2.ErrMethodNotFound
	}
	return s.opts.SendStreamMessageHandler(ctx, ss, params)
}

func (s *Server) GetTask(ctx context.Context, ss *transport.ServerSession, params *a2a.TaskQueryParams) (*a2a.Task, error) {
	if s.opts.GetTaskHandler == nil {
		return nil, jsonrpc2.ErrMethodNotFound
	}
	return s.opts.GetTaskHandler(ctx, ss, params)
}

func (s *Server) CancelTask(ctx context.Context, ss *transport.ServerSession, params *a2a.TaskIDParams) (*a2a.Task, error) {
	if s.opts.CancelTaskHandler == nil {
		return nil, jsonrpc2.ErrMethodNotFound
	}
	return s.opts.CancelTaskHandler(ctx, ss, params)
}

func (s *Server) SetTasksPushNotificationConfig(ctx context.Context, ss *transport.ServerSession, params *a2a.TaskPushNotificationConfig) (*a2a.TaskPushNotificationConfig, error) {
	if s.opts.SetTasksPushNotificationConfigHandler == nil {
		return nil, jsonrpc2.ErrMethodNotFound
	}
	return s.opts.SetTasksPushNotificationConfigHandler(ctx, ss, params)
}

func (s *Server) GetTasksPushNotificationConfig(ctx context.Context, ss *transport.ServerSession, params *a2a.GetTaskPushNotificationConfigParams) (*a2a.TaskPushNotificationConfig, error) {
	if s.opts.GetTasksPushNotificationConfigHandler == nil {
		return nil, jsonrpc2.ErrMethodNotFound
	}
	return s.opts.GetTasksPushNotificationConfigHandler(ctx, ss, params)
}

func (s *Server) ListTasksPushNotificationConfig(ctx context.Context, ss *transport.ServerSession, params *a2a.ListTaskPushNotificationConfigParams) (a2a.TaskPushNotificationConfigs, error) {
	if s.opts.ListTasksPushNotificationConfigHandler == nil {
		return nil, jsonrpc2.ErrMethodNotFound
	}
	return s.opts.ListTasksPushNotificationConfigHandler(ctx, ss, params)
}

func (s *Server) DeleteTasksPushNotificationConfig(ctx context.Context, ss *transport.ServerSession, params *a2a.DeleteTaskPushNotificationConfigParams) (*a2a.EmptyResult, error) {
	if s.opts.DeleteTasksPushNotificationConfigHandler == nil {
		return nil, jsonrpc2.ErrMethodNotFound
	}
	return s.opts.DeleteTasksPushNotificationConfigHandler(ctx, ss, params)
}

func (s *Server) ResubscribeTasks(ctx context.Context, ss *transport.ServerSession, params *a2a.TaskIDParams) (a2a.SendStreamingMessageResponse, error) {
	if s.opts.ResubscribeTasksHandler == nil {
		return nil, jsonrpc2.ErrMethodNotFound
	}
	return s.opts.ResubscribeTasksHandler(ctx, ss, params)
}
