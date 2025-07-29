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

package client

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/go-a2a/a2a-go/internal/jsonrpc2"
	"github.com/go-a2a/a2a-go/transport"
)

// Client is an A2A client, which may be connected to an A2A server
// using the [Client.Connect] method.
type Client struct {
	opts                   ClientOption
	sessions               []*transport.ClientSession
	sendingMethodHandler   transport.MethodHandler[*transport.ClientSession]
	receivingMethodHandler transport.MethodHandler[*transport.ClientSession]
	mu                     sync.Mutex
}

var (
	_ transport.Client                           = (*Client)(nil)
	_ transport.Binder[*transport.ClientSession] = (*Client)(nil)
)

// requestHandler is the HTTP request handler for a2a client.
type requestHandler struct{}

var _ transport.RequestHandler = (*requestHandler)(nil)

// Handle is the HTTP request handler for a2a client.
func (requestHandler) Handle(ctx context.Context, invoker transport.Invoker, req *http.Request) (resp *http.Response, err error) {
	defer func() {
		if err != nil && resp != nil {
			_ = resp.Body.Close()
		}
	}()

	resp, err = invoker(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("httpRequestHandler: http request failed: %w", err)
	}

	return resp, nil
}

type ClientOption struct {
	HTTPClient     *http.Client
	RequestHandler transport.RequestHandler
}

// NewClient creates a new [Client].
//
// Use [Client.Connect] to connect it to an A2A server.
func NewClient(opts *ClientOption) *Client {
	c := &Client{
		sendingMethodHandler:   transport.DefaultSendingMethodHandler[*transport.ClientSession],
		receivingMethodHandler: transport.DefaultReceivingMethodHandler[*transport.ClientSession],
		opts: ClientOption{
			HTTPClient: &http.Client{
				Timeout: 30 * time.Second,
			},
			RequestHandler: &requestHandler{},
		},
	}
	if opts != nil {
		if hc := opts.HTTPClient; hc != nil {
			c.opts.HTTPClient = hc
		}
		if h := opts.RequestHandler; h != nil {
			c.opts.RequestHandler = h
		}
	}

	return c
}

// Bind implements the binder[*ClientSession] interface, so that Clients can
// be connected using [connect].
func (c *Client) Bind(conn *jsonrpc2.Connection) *transport.ClientSession {
	cs := &transport.ClientSession{
		Connection:     conn,
		Client:         c,
		HTTPClient:     c.opts.HTTPClient,
		RequestHandler: c.opts.RequestHandler,
	}

	c.mu.Lock()
	c.sessions = append(c.sessions, cs)
	c.mu.Unlock()

	return cs
}

// Disconnect implements the binder[*Client] interface, so that
// Clients can be connected using [connect].
func (c *Client) Disconnect(cs *transport.ClientSession) {
	c.mu.Lock()
	c.sessions = slices.DeleteFunc(c.sessions, func(cs2 *transport.ClientSession) bool {
		return cs2 == cs
	})
	c.mu.Unlock()
}

// Connect begins an A2A session by connecting to a server over the given
// transport, and initializing the session.
//
// Typically, it is the responsibility of the client to close the connection
// when it is no longer needed. However, if the connection is closed by the
// server, calls or notifications will return an error wrapping
// [ErrConnectionClosed].
func (c *Client) Connect(ctx context.Context, t transport.Transport) (*transport.ClientSession, error) {
	cs, err := transport.Connect(ctx, t, c)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

// ClientSendingMethodHandler returns the current sending method handler.
//
// ClientSendingMethodHandler implements [transport.Client].
func (c *Client) ClientSendingMethodHandler() any {
	return c.sendingMethodHandler
}

// ClientReceivingMethodHandler returns the current receiving method handler.
//
// ClientReceivingMethodHandler implements [transport.Client].
func (c *Client) ClientReceivingMethodHandler() any {
	return c.receivingMethodHandler
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
func (c *Client) AddSendingMiddleware(middleware ...transport.Middleware[*transport.ClientSession]) {
	c.mu.Lock()
	transport.AddMiddleware(&c.sendingMethodHandler, middleware)
	c.mu.Unlock()
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
func (c *Client) AddReceivingMiddleware(middleware ...transport.Middleware[*transport.ClientSession]) {
	c.mu.Lock()
	transport.AddMiddleware(&c.receivingMethodHandler, middleware)
	c.mu.Unlock()
}
