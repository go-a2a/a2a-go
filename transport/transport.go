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
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"sync"
	"syscall"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/internal/jsonrpc2"
)

// ErrConnectionClosed is returned when sending a message to a connection that
// is closed or in the process of closing.
var ErrConnectionClosed = errors.New("connection closed")

// IsConnectionClosedError reports whether err indicates a connection has been closed.
// It checks for various error types and patterns that indicate connection closure.
func IsConnectionClosedError(err error) bool {
	if err == nil {
		return false
	}

	// Standard I/O errors indicating closure
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
		return true
	}

	// Check for network operation errors
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		if netErr.Op == "read" || netErr.Op == "write" {
			// Common closed connection error patterns
			errStr := netErr.Err.Error()
			return errStr == "use of closed network connection" || errStr == "broken pipe"
		}
	}

	// Check for syscall errors indicating closed connections
	var syscallErr *os.SyscallError
	if errors.As(err, &syscallErr) {
		return syscallErr.Err == syscall.EPIPE || syscallErr.Err == syscall.ECONNRESET
	}

	return false
}

// Transport is used to create a bidirectional connection between A2A client
// and server.
//
// Transports should be used for at most one call to [Server.Connect] or
// [Client.Connect].
type Transport interface {
	// Connect returns the logical JSON-RPC connection..
	//
	// It is called exactly once by [Server.Connect] or [Client.Connect].
	Connect(ctx context.Context) (Connection, error)
}

// Connection is a logical bidirectional JSON-RPC connection.
type Connection interface {
	SessionID() string
	Read(context.Context) (jsonrpc2.Message, error)
	Write(context.Context, jsonrpc2.Message) error
	Close() error // may be called concurrently by both peers
}

// ioTransport is a [Transport] that communicates using newline-delimited
// JSON over an io.ReadWriteCloser.
type ioTransport struct {
	rwc io.ReadWriteCloser
}

var _ Transport = (*ioTransport)(nil)

// Connect implements [Transport].
func (t *ioTransport) Connect(context.Context) (Connection, error) {
	return newIOConn(t.rwc), nil
}

// InMemoryTransport is a [Transport] that communicates over an in-memory
// network connection, using newline-delimited JSON.
type InMemoryTransport struct {
	ioTransport
}

var _ Transport = (*InMemoryTransport)(nil)

// NewInMemoryTransports returns two InMemoryTransports that connect to each
// other.
func NewInMemoryTransports() (Transport, Transport) {
	// Create buffered in-memory connections to avoid synchronization deadlocks
	conn1 := &bufferedConn{ch: make(chan []byte, 100)}
	conn2 := &bufferedConn{ch: make(chan []byte, 100)}
	conn1.peer = conn2
	conn2.peer = conn1
	return &InMemoryTransport{ioTransport{conn1}}, &InMemoryTransport{ioTransport{conn2}}
}

type Handler interface {
	Handle(ctx context.Context, req *jsonrpc2.Request) (any, error)
	SetConn(Connection)
}

type Binder[T Handler] interface {
	Bind(*jsonrpc2.Connection) T
	Disconnect(T)
}

func Connect[H Handler](ctx context.Context, t Transport, b Binder[H]) (h H, err error) {
	conn, err := t.Connect(ctx)
	if err != nil {
		return h, fmt.Errorf("connect with transport: %w", err)
	}

	// If logging is configured, write message logs.
	reader, writer := jsonrpc2.Reader(conn), jsonrpc2.Writer(conn)

	var preempter canceller
	bind := func(conn *jsonrpc2.Connection) jsonrpc2.Handler {
		h = b.Bind(conn)
		preempter.conn = conn
		return jsonrpc2.HandlerFunc(h.Handle)
	}

	_ = jsonrpc2.NewConnection(ctx, jsonrpc2.ConnectionConfig{
		Reader:    reader,
		Writer:    writer,
		Closer:    conn,
		Preempter: &preempter,
		Bind:      bind,
		OnDone: func() {
			b.Disconnect(h)
		},
	})

	if preempter.conn == nil {
		panic("unbound preempter")
	}

	h.SetConn(conn)
	return h, nil
}

// canceller is a [jsonrpc2.Preempter] that cancels in-flight requests on A2A
// cancelled notifications.
type canceller struct {
	conn *jsonrpc2.Connection
}

var _ jsonrpc2.Preempter = (*canceller)(nil)

// Preempt implements [jsonrpc2.Preempter].
//
// TODO(zchee): corrent implementations to TasksCancel method?
func (c *canceller) Preempt(ctx context.Context, req *jsonrpc2.Request) (result any, err error) {
	if req.Method == string(a2a.MethodTasksCancel) {
		var params a2a.TaskIDParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, err
		}
		id, err := jsonrpc2.MakeID(params.ID)
		if err != nil {
			return nil, err
		}
		go c.conn.Cancel(id)
	}
	return nil, jsonrpc2.ErrNotHandled
}

// LoggingTransport is a [Transport] that delegates to another transport,
// writing RPC logs to an io.Writer.
type LoggingTransport struct {
	delegate Transport
	w        io.Writer
}

var _ Transport = (*LoggingTransport)(nil)

// NewLoggingTransport creates a new LoggingTransport that delegates to the
// provided transport, writing RPC logs to the provided io.Writer.
func NewLoggingTransport(delegate Transport, w io.Writer) Transport {
	return &LoggingTransport{delegate, w}
}

// Connect connects the underlying transport, returning a [Connection] that writes
// logs to the configured destination.
//
// Connect implements [Transport].
func (t *LoggingTransport) Connect(ctx context.Context) (Connection, error) {
	delegate, err := t.delegate.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &loggingConn{delegate, t.w}, nil
}

type loggingConn struct {
	delegate Connection
	w        io.Writer
}

var _ Connection = (*loggingConn)(nil)

// SessionID implements [Connection].
func (c *loggingConn) SessionID() string { return c.delegate.SessionID() }

// Read is a stream middleware that logs incoming messages.
//
// Read implements [Connection].
func (s *loggingConn) Read(ctx context.Context) (jsonrpc2.Message, error) {
	msg, err := s.delegate.Read(ctx)
	if err != nil {
		// Don't log connection closed errors since they're expected during shutdown
		if !IsConnectionClosedError(err) {
			fmt.Fprintf(s.w, "read error: %v", err)
		}
	} else {
		data, err := jsonrpc2.EncodeMessage(msg)
		if err != nil {
			fmt.Fprintf(s.w, "LoggingTransport: failed to marshal: %v", err)
		}
		fmt.Fprintf(s.w, "read: %s\n", string(data))
	}
	return msg, err
}

// Write is a stream middleware that logs outgoing messages.
//
// Write implements [Connection].
func (s *loggingConn) Write(ctx context.Context, msg jsonrpc2.Message) error {
	err := s.delegate.Write(ctx, msg)
	if err != nil {
		fmt.Fprintf(s.w, "write error: %v", err)
	} else {
		data, err := jsonrpc2.EncodeMessage(msg)
		if err != nil {
			fmt.Fprintf(s.w, "LoggingTransport: failed to marshal: %v", err)
		}
		fmt.Fprintf(s.w, "write: %s\n", string(data))
	}
	return err
}

// Close implements [Connection].
func (s *loggingConn) Close() error {
	return s.delegate.Close()
}

// rwc binds an io.ReadCloser and io.WriteCloser together to create an
// io.ReadWriteCloser.
type rwc struct {
	rc io.ReadCloser
	wc io.WriteCloser
}

var _ io.ReadWriteCloser = (*rwc)(nil)

func (r rwc) Read(p []byte) (n int, err error) {
	return r.rc.Read(p)
}

func (r rwc) Write(p []byte) (n int, err error) {
	return r.wc.Write(p)
}

func (r rwc) Close() error {
	return errors.Join(r.rc.Close(), r.wc.Close())
}

// bufferedConn implements [io.ReadWriteCloser] using buffered channels for in-memory transport.
type bufferedConn struct {
	ch     chan []byte
	peer   *bufferedConn
	closed bool
	mu     sync.Mutex
	buf    []byte // current read buffer
}

var _ io.ReadWriteCloser = (*bufferedConn)(nil)

func (b *bufferedConn) Read(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return 0, io.EOF
	}

	// If we have data in our current buffer, read from it first
	if len(b.buf) > 0 {
		n = copy(p, b.buf)
		b.buf = b.buf[n:]
		return n, nil
	}

	// Get new data from channel (blocking)
	b.mu.Unlock() // Unlock before blocking read
	data, ok := <-b.ch
	b.mu.Lock() // Re-lock after read

	if !ok {
		return 0, io.EOF
	}
	n = copy(p, data)
	if n < len(data) {
		// Store remaining data for next read
		b.buf = data[n:]
	}
	return n, nil
}

func (b *bufferedConn) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed || b.peer == nil {
		return 0, io.ErrClosedPipe
	}

	// Make a copy of the data to send
	data := make([]byte, len(p))
	copy(data, p)

	// Send to peer's channel (non-blocking)
	select {
	case b.peer.ch <- data:
		return len(p), nil
	default:
		// Channel full, return error
		return 0, errors.New("write would block")
	}
}

func (b *bufferedConn) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.closed {
		b.closed = true
		close(b.ch)
	}
	return nil
}

// ioConn is a transport that delimits messages with newlines across
// a bidirectional stream, and supports jsonrpc.2 message batching.
//
// See https://github.com/ndjson/ndjson-spec for discussion of newline
// delimited JSON.
//
// See [msgBatch] for more discussion of message batching.
type ioConn struct {
	rwc io.ReadWriteCloser // the underlying stream
	in  *jsontext.Decoder  // a decoder bound to rwc

	// If outgoiBatch has a positive capacity, it will be used to batch requests
	// and notifications before sending.
	outgoingBatch []jsonrpc2.Message

	// Unread messages in the last batch. Since reads are serialized, there is no
	// need to guard here.
	queue []jsonrpc2.Message

	// batches correlate incoming requests to the batch in which they arrived.
	// Since writes may be concurrent to reads, we need to guard this with a mutex.
	batchMu sync.Mutex
	batches map[jsonrpc2.ID]*msgBatch // lazily allocated
}

var _ Connection = (*ioConn)(nil)

func newIOConn(rwc io.ReadWriteCloser) *ioConn {
	return &ioConn{
		rwc: rwc,
		in:  jsontext.NewDecoder(rwc),
	}
}

// SessionID implements [Connection].
func (c *ioConn) SessionID() string { return "" }

// Read implements [Connection].
func (t *ioConn) Read(ctx context.Context) (jsonrpc2.Message, error) {
	return t.read(ctx, t.in)
}

func (t *ioConn) read(ctx context.Context, in *jsontext.Decoder) (jsonrpc2.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// If we have a queued message, return it.
	if len(t.queue) > 0 {
		next := t.queue[0]
		t.queue = t.queue[1:]
		return next, nil
	}

	var raw jsontext.Value
	if err := json.UnmarshalDecode(in, &raw); err != nil {
		// Handle connection closed errors gracefully during shutdown
		if IsConnectionClosedError(err) {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("ioConn.read: unmarshal: %w", err)
	}
	msgs, batch, err := readBatch(raw)
	if err != nil {
		return nil, fmt.Errorf("ioConn.read: readBatch: %w", err)
	}
	t.queue = msgs[1:]

	if batch {
		var respBatch *msgBatch // track incoming requests in the batch
		for _, msg := range msgs {
			if req, ok := msg.(*jsonrpc2.Request); ok {
				if respBatch == nil {
					respBatch = &msgBatch{
						unresolved: make(map[jsonrpc2.ID]int),
					}
				}
				if _, ok := respBatch.unresolved[req.ID]; ok {
					return nil, fmt.Errorf("duplicate message ID %q", req.ID.Raw())
				}
				respBatch.unresolved[req.ID] = len(respBatch.responses)
				respBatch.responses = append(respBatch.responses, nil)
			}
		}
		if respBatch != nil {
			// The batch contains one or more incoming requests to track.
			if err := t.addBatch(respBatch); err != nil {
				return nil, fmt.Errorf("ioConn.read: addBatch: %w", err)
			}
		}
	}

	return msgs[0], err
}

// readBatch reads batch data, which may be either a single JSON-RPC message,
// or an array of JSON-RPC messages.
func readBatch(data []byte) (msgs []jsonrpc2.Message, isBatch bool, err error) {
	// Try to read an array of messages first.
	var rawBatch []jsontext.Value
	if err := json.Unmarshal(data, &rawBatch); err == nil {
		if len(rawBatch) == 0 {
			return nil, true, fmt.Errorf("empty batch")
		}

		msgs = slices.Grow(msgs, len(rawBatch))
		for _, raw := range rawBatch {
			msg, err := jsonrpc2.DecodeMessage(raw)
			if err != nil {
				return nil, true, err
			}
			msgs = append(msgs, msg)
		}
		return msgs, true, nil
	}

	// Try again with a single message.
	msg, err := jsonrpc2.DecodeMessage(data)
	return []jsonrpc2.Message{msg}, false, err
}

// msgBatch records information about an incoming batch of jsonrpc.2 calls.
//
// The jsonrpc.2 spec (https://www.jsonrpc.org/specification#batch) says:
//
// "The Server should respond with an Array containing the corresponding
// Response objects, after all of the batch Request objects have been
// processed. A Response object SHOULD exist for each Request object, except
// that there SHOULD NOT be any Response objects for notifications. The Server
// MAY process a batch rpc call as a set of concurrent tasks, processing them
// in any order and with any width of parallelism."
//
// Therefore, a msgBatch keeps track of outstanding calls and their responses.
// When there are no unresolved calls, the response payload is sent.
type msgBatch struct {
	unresolved map[jsonrpc2.ID]int
	responses  []*jsonrpc2.Response
}

// addBatch records a msgBatch for an incoming batch payload.
// It returns an error if batch is malformed, containing previously seen IDs.
//
// See [msgBatch] for more.
func (t *ioConn) addBatch(batch *msgBatch) error {
	t.batchMu.Lock()
	defer t.batchMu.Unlock()

	for id := range batch.unresolved {
		if _, ok := t.batches[id]; ok {
			return fmt.Errorf("%w: batch contains previously seen request %v", a2a.ErrInvalidRequest, id.Raw())
		}
	}
	for id := range batch.unresolved {
		if t.batches == nil {
			t.batches = make(map[jsonrpc2.ID]*msgBatch)
		}
		t.batches[id] = batch
	}

	return nil
}

// updateBatch records a response in the message batch tracking the
// corresponding incoming call, if any.
//
// The second result reports whether resp was part of a batch. If this is true,
// the first result is nil if the batch is still incomplete, or the full set of
// batch responses if resp completed the batch.
func (t *ioConn) updateBatch(resp *jsonrpc2.Response) ([]*jsonrpc2.Response, bool) {
	t.batchMu.Lock()
	defer t.batchMu.Unlock()

	if batch, ok := t.batches[resp.ID]; ok {
		idx, ok := batch.unresolved[resp.ID]
		if !ok {
			// Internal consistency error - batch exists but ID not found in unresolved map
			// This should never happen in normal operation, but instead of panicking,
			// return false to allow normal write processing
			return nil, false
		}

		batch.responses[idx] = resp
		delete(batch.unresolved, resp.ID)
		delete(t.batches, resp.ID)
		if len(batch.unresolved) == 0 {
			return batch.responses, true
		}
		return nil, true
	}

	return nil, false
}

// Write implements [Connection].
func (t *ioConn) Write(ctx context.Context, msg jsonrpc2.Message) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Batching support: if msg is a Response, it may have completed a batch, so
	// check that first. Otherwise, it is a request or notification, and we may
	// want to collect it into a batch before sending, if we're configured to use
	// outgoing batches.
	if resp, ok := msg.(*jsonrpc2.Response); ok {
		if batch, ok := t.updateBatch(resp); ok {
			if len(batch) > 0 {
				data, err := marshalMessages(batch)
				if err != nil {
					return err
				}
				data = append(data, '\n')
				_, err = t.rwc.Write(data)
				return err
			}
			return nil
		}
	} else if len(t.outgoingBatch) < cap(t.outgoingBatch) {
		t.outgoingBatch = append(t.outgoingBatch, msg)
		if len(t.outgoingBatch) == cap(t.outgoingBatch) {
			data, err := marshalMessages(t.outgoingBatch)
			t.outgoingBatch = t.outgoingBatch[:0]
			if err != nil {
				return err
			}
			data = append(data, '\n')
			_, err = t.rwc.Write(data)
			return err
		}
		return nil
	}

	data, err := jsonrpc2.EncodeMessage(msg)
	if err != nil {
		return fmt.Errorf("marshaling message: %v", err)
	}
	data = append(data, '\n') // newline delimited
	_, err = t.rwc.Write(data)

	return err
}

// Close implements [Connection].
func (t *ioConn) Close() error {
	return t.rwc.Close()
}

func marshalMessages[T jsonrpc2.Message](msgs []T) ([]byte, error) {
	var rawMsgs []jsontext.Value
	for _, msg := range msgs {
		raw, err := jsonrpc2.EncodeMessage(msg)
		if err != nil {
			return nil, fmt.Errorf("encoding batch message: %w", err)
		}
		rawMsgs = append(rawMsgs, raw)
	}
	return json.Marshal(rawMsgs)
}
