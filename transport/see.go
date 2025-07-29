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
	"bytes"
	"context"
	crand "crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"

	"github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/internal/jsonrpc2"
)

// GetServer represents a function that returns a [Server] for a given [*http.Request].
type GetServer func(request *http.Request) Server

// ServerHandler represents an HTTP handler that serves A2A sessions.
type ServerHandler struct {
	getServer  GetServer
	sseHandler *SSEHandler
	sessions   sync.Map // map[string]*SSEServerTransport
}

var _ http.Handler = (*ServerHandler)(nil)

// NewServerHandler returns a new [ServerHandler] that serves A2A sessions.
func NewServerHandler(getServer GetServer, opts *SSEHandlerOptions) *ServerHandler {
	h := &ServerHandler{
		getServer:  getServer,
		sseHandler: NewSSEHandler(getServer, opts),
	}

	return h
}

// ServeHTTP implements [http.Handler].
func (h *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("r.Method: %#v\n", r.Method)
	switch r.Method {
	case http.MethodGet:
		switch r.URL.Path {
		case a2a.AgentCardWellKnownPath:
			h.handleAgentCard(w, r)
		default:
			h.sseHandler.ServeHTTP(w, r)
		}

	case http.MethodPost:
		accept := r.Header.Get("Accept")
		if accept == "" {
			http.Error(w, "Accept header is empty", http.StatusBadRequest)
			return
		}
		fmt.Printf("accept: %#v\n", accept)

		switch accept {
		case "application/json":
			h.handleRequest(w, r)

		case "text/event-stream":
			// For POST requests and accept "text/event-stream", the message body is a message to send to a session.
			h.sseHandler.handleSSE(w, r)
		}
	}
}

func (h *ServerHandler) handleAgentCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enc := jsontext.NewEncoder(w)
	if err := json.MarshalEncode(enc, h.getServer(r).AgentCard()); err != nil {
		http.Error(w, "marshal agent card", http.StatusInternalServerError)
	}
}

func (h *ServerHandler) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Read and parse the message.
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	// Optionally, we could just push the data onto a channel, and let the
	// message fail to parse when it is read. This failure seems a bit more
	// useful
	msg, err := jsonrpc2.DecodeMessage(data)
	if err != nil {
		http.Error(w, "parse body failed", http.StatusBadRequest)
		return
	}

	req := msg.(*jsonrpc2.Request)
	switch a2a.Method(req.Method) {
	case a2a.MethodMessageSend:
		fmt.Printf("req: %#v\n", req)
		// server := h.getServer(r)
		// ss, err := server.Connect(ctx, newIOConn(w))
		// if err != nil {
		// 	http.Error(w, "parse body failed", http.StatusBadRequest)
		// 	return
		// }

		var params a2a.MessageSendParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			http.Error(w, "parse body failed", http.StatusBadRequest)
			return
		}
		// resp, err := server.SendMessage(r.Context(), ss, &params)
		response, _ := jsonrpc2.NewResponse(req.ID, &params, err)
		data, _ := jsonrpc2.EncodeMessage(response)
		w.Write(data)
	}
}

// This file implements support for SSE (HTTP with server-sent events)
// transport server and client.
// https://modelcontextprotocol.io/specification/2024-11-05/basic/transports
//
// The transport is simple, at least relative to the new streamable transport
// introduced in the 2025-03-26 version of the spec. In short:
//
//  1. Sessions are initiated via a hanging GET request, which streams
//     server->client messages as SSE 'message' events.
//  2. The first event in the SSE stream must be an 'endpoint' event that
//     informs the client of the session endpoint.
//  3. The client POSTs client->server messages to the session endpoint.
//
// Therefore, the each new GET request hands off its responsewriter to an
// [SSEServerTransport] type that abstracts the transport as follows:
//  - Write writes a new event to the responseWriter, or fails if the GET has
//  exited.
//  - Read reads off a message queue that is pushed to via POST requests.
//  - Close causes the hanging GET to exit.

// SSEHandler is an http.Handler that serves SSE-based A2A sessions.
type SSEHandler struct {
	getServer    func(request *http.Request) Server
	onConnection func(*ServerSession) // for testing; must not block
	opts         SSEHandlerOptions

	sessions sync.Map // map[string]*SSEServerTransport
}

var _ http.Handler = (*SSEHandler)(nil)

// SSEHandlerOptions provides options for the [NewSSEHandler] constructor.
type SSEHandlerOptions struct{}

// NewSSEHandler returns a new [SSEHandler] that creates and manages A2A
// sessions created via incoming HTTP requests.
//
// Sessions are created when the client issues a GET request to the server,
// which must accept text/event-stream responses (server-sent events).
// For each such request, a new [SSEServerTransport] is created with a distinct
// messages endpoint, and connected to the server returned by getServer.
// The SSEHandler also handles requests to the message endpoints, by
// delegating them to the relevant server transport.
//
// The getServer function may return a distinct [Server] for each new
// request, or reuse an existing server. If it returns nil, the handler
// will return a 400 Bad Request.
func NewSSEHandler(getServer func(request *http.Request) Server, opts *SSEHandlerOptions) *SSEHandler {
	h := &SSEHandler{
		getServer: getServer,
	}
	if opts != nil {
		h.opts = *opts
	}

	return h
}

func (h *SSEHandler) handleSSE(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionid")
	// Look up the session.
	if sessionID == "" {
		http.Error(w, "sessionid must be provided", http.StatusBadRequest)
		return
	}
	session, ok := h.sessions.Load(sessionID)
	if !ok {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	session.(*SSEServerTransport).ServeHTTP(w, r)
}

// ServeHTTP implements [http.Handler].
func (h *SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	// GET requests create a new session, and serve messages over SSE.

	// TODO: it's not entirely documented whether we should check Accept here.
	// Let's again be lax and assume the client will accept SSE.

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	sessionID := crand.Text()
	endpoint, err := r.URL.Parse("?sessionid=" + sessionID)
	if err != nil {
		http.Error(w, "internal error: failed to create endpoint", http.StatusInternalServerError)
		return
	}

	transport := NewSSEServerTransport(endpoint.RequestURI(), w, sessionID)

	// The session is terminated when the request exits.
	h.sessions.Store(sessionID, transport)
	defer func() {
		h.sessions.Delete(sessionID)
	}()

	server := h.getServer(r)
	if server == nil {
		// The getServer argument to NewSSEHandler returned nil.
		http.Error(w, "no server available", http.StatusBadRequest)
		return
	}
	ss, err := server.Connect(r.Context(), transport)
	if err != nil {
		http.Error(w, "connection failed", http.StatusInternalServerError)
		return
	}
	if h.onConnection != nil {
		h.onConnection(ss)
	}
	defer ss.Close() // close the transport when the GET exits

	select {
	case <-r.Context().Done():
	case <-transport.done:
	}
}

// SSEServerTransport is a logical SSE session created through a hanging GET
// request.
//
// When connected, it returns the following [Connection] implementation:
//   - Writes are SSE 'message' events to the GET response.
//   - Reads are received from POSTs to the session endpoint, via
//     [SSEServerTransport.ServeHTTP].
//   - Close terminates the hanging GET.
type SSEServerTransport struct {
	endpoint string
	incoming chan jsonrpc2.Message // queue of incoming messages; never closed

	sessionID string

	// We must guard both pushes to the incoming queue and writes to the response
	// writer, because incoming POST requests are arbitrarily concurrent and we
	// need to ensure we don't write push to the queue, or write to the
	// ResponseWriter, after the session GET request exits.
	mu     sync.Mutex
	w      http.ResponseWriter // the hanging response body
	closed bool                // set when the stream is closed
	done   chan struct{}       // closed when the connection is closed
}

var _ Transport = (*SSEServerTransport)(nil)

// NewSSEServerTransport creates a new SSE transport for the given messages
// endpoint, and hanging GET response.
//
// Use [SSEServerTransport.Connect] to initiate the flow of messages.
//
// The transport is itself an [http.Handler]. It is the caller's responsibility
// to ensure that the resulting transport serves HTTP requests on the given
// session endpoint.
//
// Most callers should instead use an [SSEHandler], which transparently handles
// the delegation to SSEServerTransports.
func NewSSEServerTransport(endpoint string, w http.ResponseWriter, sessionID string) *SSEServerTransport {
	return &SSEServerTransport{
		endpoint:  endpoint,
		sessionID: sessionID,
		w:         w,
		incoming:  make(chan jsonrpc2.Message, 100),
		done:      make(chan struct{}),
	}
}

// ServeHTTP handles POST requests to the transport endpoint.
//
// ServeHTTP implements [http.Handler].
func (t *SSEServerTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Read and parse the message.
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	// Optionally, we could just push the data onto a channel, and let the
	// message fail to parse when it is read. This failure seems a bit more
	// useful
	msg, err := jsonrpc2.DecodeMessage(data)
	if err != nil {
		http.Error(w, "parse body failed", http.StatusBadRequest)
		return
	}

	select {
	case t.incoming <- msg:
		w.WriteHeader(http.StatusAccepted)
	case <-t.done:
		http.Error(w, "session closed", http.StatusBadRequest)
	}
}

// Connect sends the 'endpoint' event to the client.
// See [SSEServerTransport] for more details on the [Connection] implementation.
func (t *SSEServerTransport) Connect(context.Context) (Connection, error) {
	t.mu.Lock()
	_, err := writeEvent(t.w, Event{
		Name: "endpoint",
		Data: []byte(t.endpoint),
	})
	t.mu.Unlock()
	if err != nil {
		return nil, err
	}

	return sseServerConn{t}, nil
}

// sseServerConn implements the [Connection] interface for a single [SSEServerTransport].
// It hides the Connection interface from the [SSEServerTransport] API.
type sseServerConn struct {
	t *SSEServerTransport
}

var _ Connection = (*sseServerConn)(nil)

func (s sseServerConn) SessionID() string { return s.t.sessionID }

// Read implements [jsonrpc2.Reader].
func (s sseServerConn) Read(ctx context.Context) (jsonrpc2.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case msg := <-s.t.incoming:
		return msg, nil

	case <-s.t.done:
		return nil, io.EOF
	}
}

// Write implements [jsonrpc2.Writer].
func (s sseServerConn) Write(ctx context.Context, msg jsonrpc2.Message) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	data, err := jsonrpc2.EncodeMessage(msg)
	if err != nil {
		return err
	}

	s.t.mu.Lock()
	defer s.t.mu.Unlock()

	// Note that it is invalid to write to a ResponseWriter after ServeHTTP has
	// exited, and so we must lock around this write and check isDone, which is
	// set before the hanging GET exits.
	if s.t.closed {
		return io.EOF
	}

	_, err = writeEvent(s.t.w, Event{
		Name: "message",
		Data: data,
	})
	return err
}

// Close implements io.Closer, and closes the session.
//
// It must be safe to call Close more than once, as the close may
// asynchronously be initiated by either the server closing its connection, or
// by the hanging GET exiting.
func (s sseServerConn) Close() error {
	s.t.mu.Lock()
	defer s.t.mu.Unlock()
	if !s.t.closed {
		s.t.closed = true
		close(s.t.done)
	}
	return nil
}

// SSEClientTransport is a [Transport] that can communicate with an MCP
// endpoint serving the SSE transport defined by the 2024-11-05 version of the
// spec.
//
// https://modelcontextprotocol.io/specification/2024-11-05/basic/transports
type SSEClientTransport struct {
	sseEndpoint *url.URL
	opts        SSEClientTransportOptions
}

var _ Transport = (*SSEClientTransport)(nil)

// SSEClientTransportOptions provides options for the [NewSSEClientTransport]
// constructor.
type SSEClientTransportOptions struct {
	// HTTPClient is the client to use for making HTTP requests. If nil,
	// http.DefaultClient is used.
	HTTPClient *http.Client
}

// NewSSEClientTransport returns a new client transport that connects to the
// SSE server at the provided URL.
//
// NewSSEClientTransport panics if the given URL is invalid.
func NewSSEClientTransport(baseURL string, opts *SSEClientTransportOptions) *SSEClientTransport {
	url, err := url.Parse(baseURL)
	if err != nil {
		panic(fmt.Sprintf("invalid base url: %v", err))
	}

	t := &SSEClientTransport{
		sseEndpoint: url,
	}
	if opts != nil {
		t.opts = *opts
	}

	return t
}

// Connect connects through the client endpoint.
func (c *SSEClientTransport) Connect(ctx context.Context) (Connection, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.sseEndpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	httpClient := c.opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	msgEndpoint, err := func() (*url.URL, error) {
		var evt Event
		for evt, err = range scanEvents(resp.Body) {
			break
		}
		if err != nil {
			return nil, err
		}
		if evt.Name != "endpoint" {
			return nil, fmt.Errorf("first event is %q, want %q", evt.Name, "endpoint")
		}
		raw := string(evt.Data)
		return c.sseEndpoint.Parse(raw)
	}()
	if err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("missing endpoint: %v", err)
	}

	// From here on, the stream takes ownership of resp.Body.
	s := &sseClientConn{
		client:      httpClient,
		sseEndpoint: c.sseEndpoint,
		msgEndpoint: msgEndpoint,
		incoming:    make(chan []byte, 100),
		body:        resp.Body,
		done:        make(chan struct{}),
	}

	go func() {
		defer s.Close() // close the transport when the GET exits

		for evt, err := range scanEvents(resp.Body) {
			if err != nil {
				return
			}
			select {
			case s.incoming <- evt.Data:
			case <-s.done:
				return
			}
		}
	}()

	return s, nil
}

// sseClientConn is a logical jsonrpc2 connection that implements the client
// half of the SSE protocol:
//   - Writes are POSTS to the session endpoint.
//   - Reads are SSE 'message' events, and pushes them onto a buffered channel.
//   - Close terminates the GET request.
type sseClientConn struct {
	client      *http.Client // HTTP client to use for requests
	sseEndpoint *url.URL     // SSE endpoint for the GET
	msgEndpoint *url.URL     // session endpoint for POSTs
	incoming    chan []byte  // queue of incoming messages

	sessionID string // TODO(zchee): fill with any logic

	mu     sync.Mutex
	body   io.ReadCloser // body of the hanging GET
	closed bool          // set when the stream is closed
	done   chan struct{} // closed when the stream is closed
}

var _ Connection = (*sseClientConn)(nil)

// TODO(jba): get the session ID. (Not urgent because SSE transports have been removed from the spec.)
func (c *sseClientConn) SessionID() string { return "" }

func (c *sseClientConn) isDone() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

func (c *sseClientConn) Read(ctx context.Context) (jsonrpc2.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case <-c.done:
		return nil, io.EOF

	case data := <-c.incoming:
		// TODO(rfindley): do we really need to check this? We receive from c.done above.
		if c.isDone() {
			return nil, io.EOF
		}
		msg, err := jsonrpc2.DecodeMessage(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	}
}

func (c *sseClientConn) Write(ctx context.Context, msg jsonrpc2.Message) error {
	data, err := jsonrpc2.EncodeMessage(msg)
	if err != nil {
		return err
	}
	if c.isDone() {
		return io.EOF
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.msgEndpoint.String(), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream") // request SSE
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to write: %s", resp.Status)
	}
	return nil
}

func (c *sseClientConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.closed = true
		_ = c.body.Close()
		close(c.done)
	}
	return nil
}
