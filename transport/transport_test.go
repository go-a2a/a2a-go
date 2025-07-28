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
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/internal/jsonrpc2"
)

// Test helpers and mocks

// mockConnection implements Connection interface for testing
type mockConnection struct {
	mu        sync.Mutex
	sessionID string
	closed    bool
	messages  []jsonrpc2.Message
	readIdx   int
	writeErr  error
	readErr   error
	closeErr  error
}

func newMockConnection(sessionID string) *mockConnection {
	return &mockConnection{
		sessionID: sessionID,
		messages:  make([]jsonrpc2.Message, 0),
	}
}

func (m *mockConnection) SessionID() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessionID
}

func (m *mockConnection) Read(ctx context.Context) (jsonrpc2.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.readErr != nil {
		return nil, m.readErr
	}

	if m.closed {
		return nil, io.EOF
	}

	if m.readIdx >= len(m.messages) {
		return nil, io.EOF
	}

	msg := m.messages[m.readIdx]
	m.readIdx++
	return msg, nil
}

func (m *mockConnection) Write(ctx context.Context, msg jsonrpc2.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.writeErr != nil {
		return m.writeErr
	}

	if m.closed {
		return io.ErrClosedPipe
	}

	m.messages = append(m.messages, msg)
	return nil
}

func (m *mockConnection) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closeErr != nil {
		return m.closeErr
	}

	m.closed = true
	return nil
}

func (m *mockConnection) addMessage(msg jsonrpc2.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

func (m *mockConnection) setReadError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readErr = err
}

func (m *mockConnection) setWriteError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeErr = err
}

// mockTransport implements Transport interface for testing
type mockTransport struct {
	conn      Connection
	connErr   error
	connected bool
}

func newMockTransport(conn Connection) *mockTransport {
	return &mockTransport{
		conn: conn,
	}
}

func (m *mockTransport) Connect(ctx context.Context) (Connection, error) {
	if m.connErr != nil {
		return nil, m.connErr
	}
	m.connected = true
	return m.conn, nil
}

// mockMethodHandler implements MethodHandler for testing
type mockMethodHandler struct {
	handleFunc func(ctx context.Context, session any, method a2a.Method, params a2a.Params) (a2a.Result, error)
}

func (m *mockMethodHandler) Handle(ctx context.Context, session any, method a2a.Method, params a2a.Params) (a2a.Result, error) {
	if m.handleFunc != nil {
		return m.handleFunc(ctx, session, method, params)
	}
	return &a2a.EmptyResult{}, nil
}

// mockReadWriteCloser implements io.ReadWriteCloser for testing
type mockReadWriteCloser struct {
	readData  []byte
	writeData bytes.Buffer
	readErr   error
	writeErr  error
	closeErr  error
	closed    bool
	mu        sync.Mutex
}

func newMockReadWriteCloser(data []byte) *mockReadWriteCloser {
	return &mockReadWriteCloser{
		readData: data,
	}
}

func (m *mockReadWriteCloser) Read(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.readErr != nil {
		return 0, m.readErr
	}

	if m.closed {
		return 0, io.EOF
	}

	n = copy(p, m.readData)
	m.readData = m.readData[n:]

	if len(m.readData) == 0 {
		return n, io.EOF
	}

	return n, nil
}

func (m *mockReadWriteCloser) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.writeErr != nil {
		return 0, m.writeErr
	}

	if m.closed {
		return 0, io.ErrClosedPipe
	}

	return m.writeData.Write(p)
}

func (m *mockReadWriteCloser) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closeErr != nil {
		return m.closeErr
	}

	m.closed = true
	return nil
}

func (m *mockReadWriteCloser) getWrittenData() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.writeData.Bytes()
}

func (m *mockReadWriteCloser) setReadError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readErr = err
}

func (m *mockReadWriteCloser) setWriteError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeErr = err
}

func (m *mockReadWriteCloser) setCloseError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeErr = err
}

// Test data helpers

func createTestRequest(id, method string, params any) (*jsonrpc2.Request, error) {
	return jsonrpc2.NewCall(jsonrpc2.StringID(id), method, params)
}

func createTestResponse(id string, result any, err error) (*jsonrpc2.Response, error) {
	return jsonrpc2.NewResponse(jsonrpc2.StringID(id), result, err)
}

// Actual tests

func TestIsConnectionClosedError(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err  error
		want bool
	}{
		"nil error": {
			err:  nil,
			want: false,
		},
		"EOF error": {
			err:  io.EOF,
			want: true,
		},
		"ErrClosedPipe": {
			err:  io.ErrClosedPipe,
			want: true,
		},
		"wrapped EOF": {
			err:  fmt.Errorf("wrapped: %w", io.EOF),
			want: true, // Properly wrapped with errors.Is
		},
		"network operation error - broken pipe": {
			err: &net.OpError{
				Op:  "read",
				Err: errors.New("broken pipe"),
			},
			want: true,
		},
		"network operation error - use of closed network connection": {
			err: &net.OpError{
				Op:  "write",
				Err: errors.New("use of closed network connection"),
			},
			want: true,
		},
		"network operation error - other error": {
			err: &net.OpError{
				Op:  "read",
				Err: errors.New("some other error"),
			},
			want: false,
		},
		"syscall error EPIPE": {
			err: &os.SyscallError{
				Syscall: "write",
				Err:     syscall.EPIPE,
			},
			want: true,
		},
		"syscall error ECONNRESET": {
			err: &os.SyscallError{
				Syscall: "read",
				Err:     syscall.ECONNRESET,
			},
			want: true,
		},
		"syscall error other": {
			err: &os.SyscallError{
				Syscall: "read",
				Err:     syscall.EINVAL,
			},
			want: false,
		},
		"unrelated error": {
			err:  errors.New("unrelated error"),
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := IsConnectionClosedError(tt.err)
			if got != tt.want {
				t.Errorf("IsConnectionClosedError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestIOTransportConnect(t *testing.T) {
	t.Parallel()

	t.Run("successful connection", func(t *testing.T) {
		t.Parallel()

		rwc := newMockReadWriteCloser([]byte(`{"jsonrpc":"2.0","method":"test","id":"1"}`))
		transport := &ioTransport{rwc: rwc}

		ctx := t.Context()
		conn, err := transport.Connect(ctx)

		if err != nil {
			t.Fatalf("Connect() returned error: %v", err)
		}

		if conn == nil {
			t.Fatal("Connect() returned nil connection")
		}

		// Verify it's an ioConn
		ioConn, ok := conn.(*ioConn)
		if !ok {
			t.Fatalf("Connect() returned wrong type: %T, want *ioConn", conn)
		}

		if ioConn.rwc != rwc {
			t.Error("ioConn.rwc doesn't match original ReadWriteCloser")
		}
	})
}

func TestNewInMemoryTransports(t *testing.T) {
	t.Parallel()

	t.Run("creates working pair", func(t *testing.T) {
		t.Parallel()

		t1, t2 := NewInMemoryTransports()

		if t1 == nil || t2 == nil {
			t.Fatal("NewInMemoryTransports() returned nil transport(s)")
		}

		// Both should be InMemoryTransport
		_, ok1 := t1.(*InMemoryTransport)
		_, ok2 := t2.(*InMemoryTransport)

		if !ok1 || !ok2 {
			t.Fatalf("Wrong transport types: %T, %T", t1, t2)
		}

		ctx := t.Context()

		// Connect both transports
		conn1, err1 := t1.Connect(ctx)
		conn2, err2 := t2.Connect(ctx)

		if err1 != nil || err2 != nil {
			t.Fatalf("Connect errors: %v, %v", err1, err2)
		}

		if conn1 == nil || conn2 == nil {
			t.Fatal("Connect() returned nil connection(s)")
		}
	})

	t.Run("bidirectional communication", func(t *testing.T) {
		t.Parallel()

		t1, t2 := NewInMemoryTransports()
		ctx := t.Context()

		conn1, err := t1.Connect(ctx)
		if err != nil {
			t.Fatalf("t1.Connect() error: %v", err)
		}

		conn2, err := t2.Connect(ctx)
		if err != nil {
			t.Fatalf("t2.Connect() error: %v", err)
		}

		// Create test message
		testMsg, err := createTestRequest("1", "test/method", map[string]string{"param": "value"})
		if err != nil {
			t.Fatalf("createTestRequest() error: %v", err)
		}

		// Use channels to coordinate the write and read operations to avoid deadlock
		writeErr := make(chan error, 1)
		readResult := make(chan jsonrpc2.Message, 1)
		readErr := make(chan error, 1)

		// Send message from conn1 to conn2 in a goroutine
		go func() {
			writeErr <- conn1.Write(ctx, testMsg)
		}()

		// Read message on conn2 in a goroutine
		go func() {
			msg, err := conn2.Read(ctx)
			if err != nil {
				readErr <- err
				return
			}
			readResult <- msg
		}()

		// Wait for write to complete
		select {
		case err := <-writeErr:
			if err != nil {
				t.Fatalf("conn1.Write() error: %v", err)
			}
		case <-ctx.Done():
			t.Fatal("Write operation timed out")
		}

		// Wait for read to complete
		var receivedMsg jsonrpc2.Message
		select {
		case receivedMsg = <-readResult:
		case err := <-readErr:
			t.Fatalf("conn2.Read() error: %v", err)
		case <-ctx.Done():
			t.Fatal("Read operation timed out")
		}

		receivedReq, ok := receivedMsg.(*jsonrpc2.Request)
		if !ok {
			t.Fatalf("Received wrong message type: %T", receivedMsg)
		}

		if receivedReq.Method != "test/method" {
			t.Errorf("Method mismatch: got %q, want %q", receivedReq.Method, "test/method")
		}

		if receivedReq.ID.Raw() != "1" {
			t.Errorf("ID mismatch: got %v, want %q", receivedReq.ID.Raw(), "1")
		}
	})
}

func TestLoggingTransport(t *testing.T) {
	t.Parallel()

	t.Run("logs read and write operations", func(t *testing.T) {
		t.Parallel()

		// Create a mock connection
		mockConn := newMockConnection("test-session")
		testMsg, _ := createTestRequest("1", "test/method", nil)
		mockConn.addMessage(testMsg)

		// Create delegate transport
		delegate := newMockTransport(mockConn)

		// Create logging transport with string buffer
		var logBuffer strings.Builder
		loggingTransport := NewLoggingTransport(delegate, &logBuffer)

		ctx := t.Context()
		conn, err := loggingTransport.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect() error: %v", err)
		}

		// Test read operation
		msg, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("Read() error: %v", err)
		}

		if msg == nil {
			t.Fatal("Read() returned nil message")
		}

		// Check that read was logged
		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "read:") {
			t.Error("Read operation was not logged")
		}

		// Test write operation
		logBuffer.Reset()
		testResponse, _ := createTestResponse("1", "result", nil)
		err = conn.Write(ctx, testResponse)
		if err != nil {
			t.Fatalf("Write() error: %v", err)
		}

		// Check that write was logged
		logOutput = logBuffer.String()
		if !strings.Contains(logOutput, "write:") {
			t.Error("Write operation was not logged")
		}
	})

	t.Run("logs read errors", func(t *testing.T) {
		t.Parallel()

		mockConn := newMockConnection("test-session")
		mockConn.setReadError(errors.New("read error"))

		delegate := newMockTransport(mockConn)
		var logBuffer strings.Builder
		loggingTransport := NewLoggingTransport(delegate, &logBuffer)

		ctx := t.Context()
		conn, err := loggingTransport.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect() error: %v", err)
		}

		_, err = conn.Read(ctx)
		if err == nil {
			t.Fatal("Expected read error")
		}

		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "read error") {
			t.Error("Read error was not logged")
		}
	})

	t.Run("logs write errors", func(t *testing.T) {
		t.Parallel()

		mockConn := newMockConnection("test-session")
		mockConn.setWriteError(errors.New("write error"))

		delegate := newMockTransport(mockConn)
		var logBuffer strings.Builder
		loggingTransport := NewLoggingTransport(delegate, &logBuffer)

		ctx := t.Context()
		conn, err := loggingTransport.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect() error: %v", err)
		}

		testMsg, _ := createTestRequest("1", "test", nil)
		err = conn.Write(ctx, testMsg)
		if err == nil {
			t.Fatal("Expected write error")
		}

		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "write error") {
			t.Error("Write error was not logged")
		}
	})

	t.Run("doesn't log connection closed errors on read", func(t *testing.T) {
		t.Parallel()

		mockConn := newMockConnection("test-session")
		mockConn.setReadError(io.EOF)

		delegate := newMockTransport(mockConn)
		var logBuffer strings.Builder
		loggingTransport := NewLoggingTransport(delegate, &logBuffer)

		ctx := t.Context()
		conn, err := loggingTransport.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect() error: %v", err)
		}

		_, err = conn.Read(ctx)
		if err != io.EOF {
			t.Fatalf("Expected EOF, got: %v", err)
		}

		logOutput := logBuffer.String()
		if strings.Contains(logOutput, "read error") {
			t.Error("Connection closed error was logged when it shouldn't be")
		}
	})
}

func TestRWCWrapper(t *testing.T) {
	t.Parallel()

	t.Run("successful read/write/close", func(t *testing.T) {
		t.Parallel()

		data := []byte("test data")
		rc := newMockReadWriteCloser(data)
		wc := newMockReadWriteCloser(nil)

		wrapper := rwc{rc: rc, wc: wc}

		// Test read
		readBuf := make([]byte, len(data))
		n, err := wrapper.Read(readBuf)
		if err != nil && err != io.EOF {
			t.Fatalf("Read() error: %v", err)
		}
		if n != len(data) {
			t.Errorf("Read() got %d bytes, want %d", n, len(data))
		}
		if !bytes.Equal(readBuf[:n], data) {
			t.Errorf("Read() got %q, want %q", readBuf[:n], data)
		}

		// Test write
		writeData := []byte("write test")
		n, err = wrapper.Write(writeData)
		if err != nil {
			t.Fatalf("Write() error: %v", err)
		}
		if n != len(writeData) {
			t.Errorf("Write() got %d bytes, want %d", n, len(writeData))
		}

		// Verify write data
		writtenData := wc.getWrittenData()
		if !bytes.Equal(writtenData, writeData) {
			t.Errorf("Written data %q, want %q", writtenData, writeData)
		}

		// Test close
		err = wrapper.Close()
		if err != nil {
			t.Fatalf("Close() error: %v", err)
		}
	})

	t.Run("close with errors", func(t *testing.T) {
		t.Parallel()

		rc := newMockReadWriteCloser(nil)
		wc := newMockReadWriteCloser(nil)

		rcErr := errors.New("read closer error")
		wcErr := errors.New("write closer error")

		rc.setCloseError(rcErr)
		wc.setCloseError(wcErr)

		wrapper := rwc{rc: rc, wc: wc}

		err := wrapper.Close()
		if err == nil {
			t.Fatal("Expected close error")
		}

		// Should contain both errors
		errStr := err.Error()
		if !strings.Contains(errStr, "read closer error") {
			t.Error("Close error should contain read closer error")
		}
		if !strings.Contains(errStr, "write closer error") {
			t.Error("Close error should contain write closer error")
		}
	})
}

func TestIOConnRead(t *testing.T) {
	t.Parallel()

	t.Run("read single message", func(t *testing.T) {
		t.Parallel()

		jsonMsg := `{"jsonrpc":"2.0","method":"test","id":"1"}` + "\n"
		rwc := newMockReadWriteCloser([]byte(jsonMsg))
		conn := newIOConn(rwc)

		ctx := t.Context()
		msg, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("Read() error: %v", err)
		}

		req, ok := msg.(*jsonrpc2.Request)
		if !ok {
			t.Fatalf("Message type %T, want *jsonrpc2.Request", msg)
		}

		if req.Method != "test" {
			t.Errorf("Method = %q, want %q", req.Method, "test")
		}

		if req.ID.Raw() != "1" {
			t.Errorf("ID = %v, want %q", req.ID.Raw(), "1")
		}
	})

	t.Run("read batched messages", func(t *testing.T) {
		t.Parallel()

		// JSON-RPC batch format
		jsonBatch := `[
			{"jsonrpc":"2.0","method":"test1","id":"1"},
			{"jsonrpc":"2.0","method":"test2","id":"2"}
		]` + "\n"

		rwc := newMockReadWriteCloser([]byte(jsonBatch))
		conn := newIOConn(rwc)

		ctx := t.Context()

		// First read should return first message from batch
		msg1, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("Read() first message error: %v", err)
		}

		req1, ok := msg1.(*jsonrpc2.Request)
		if !ok {
			t.Fatalf("First message type %T, want *jsonrpc2.Request", msg1)
		}

		if req1.Method != "test1" {
			t.Errorf("First message method = %q, want %q", req1.Method, "test1")
		}

		// Second read should return second message from batch
		msg2, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("Read() second message error: %v", err)
		}

		req2, ok := msg2.(*jsonrpc2.Request)
		if !ok {
			t.Fatalf("Second message type %T, want *jsonrpc2.Request", msg2)
		}

		if req2.Method != "test2" {
			t.Errorf("Second message method = %q, want %q", req2.Method, "test2")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		t.Parallel()

		rwc := newMockReadWriteCloser([]byte(""))
		conn := newIOConn(rwc)

		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel immediately

		_, err := conn.Read(ctx)
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got: %v", err)
		}
	})

	t.Run("connection closed error handling", func(t *testing.T) {
		t.Parallel()

		rwc := newMockReadWriteCloser(nil)
		rwc.setReadError(io.EOF)
		conn := newIOConn(rwc)

		ctx := t.Context()
		_, err := conn.Read(ctx)
		if err != io.EOF {
			t.Errorf("Expected io.EOF, got: %v", err)
		}
	})

	t.Run("invalid JSON handling", func(t *testing.T) {
		t.Parallel()

		rwc := newMockReadWriteCloser([]byte(`{invalid json}` + "\n"))
		conn := newIOConn(rwc)

		ctx := t.Context()
		_, err := conn.Read(ctx)
		if err == nil {
			t.Fatal("Expected error for invalid JSON")
		}
	})
}

func TestIOConnWrite(t *testing.T) {
	t.Parallel()

	t.Run("write single message", func(t *testing.T) {
		t.Parallel()

		rwc := newMockReadWriteCloser(nil)
		conn := newIOConn(rwc)

		ctx := t.Context()
		testMsg, _ := createTestRequest("1", "test/method", nil)

		err := conn.Write(ctx, testMsg)
		if err != nil {
			t.Fatalf("Write() error: %v", err)
		}

		written := rwc.getWrittenData()
		if len(written) == 0 {
			t.Fatal("Nothing was written")
		}

		// Should be JSON + newline
		if !bytes.HasSuffix(written, []byte("\n")) {
			t.Error("Written data should end with newline")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		t.Parallel()

		rwc := newMockReadWriteCloser(nil)
		conn := newIOConn(rwc)

		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel immediately

		testMsg, _ := createTestRequest("1", "test", nil)
		err := conn.Write(ctx, testMsg)
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got: %v", err)
		}
	})

	t.Run("write error handling", func(t *testing.T) {
		t.Parallel()

		rwc := newMockReadWriteCloser(nil)
		rwc.setWriteError(errors.New("write failed"))
		conn := newIOConn(rwc)

		ctx := t.Context()
		testMsg, _ := createTestRequest("1", "test", nil)

		err := conn.Write(ctx, testMsg)
		if err == nil {
			t.Fatal("Expected write error")
		}

		if !strings.Contains(err.Error(), "write failed") {
			t.Errorf("Error should contain 'write failed', got: %v", err)
		}
	})
}

func TestIOConnMessageBatching(t *testing.T) {
	t.Parallel()

	t.Run("response batching", func(t *testing.T) {
		t.Parallel()

		// Create a batch of requests
		jsonBatch := `[
			{"jsonrpc":"2.0","method":"test1","id":"1"},
			{"jsonrpc":"2.0","method":"test2","id":"2"}
		]` + "\n"

		rwc := newMockReadWriteCloser([]byte(jsonBatch))
		conn := newIOConn(rwc)

		ctx := t.Context()

		// Read the batch (which sets up internal batch tracking)
		msg1, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("Read() first message error: %v", err)
		}
		msg2, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("Read() second message error: %v", err)
		}

		// Cast to requests to get IDs (for clarity)
		_ = msg1.(*jsonrpc2.Request)
		_ = msg2.(*jsonrpc2.Request)

		// Write responses for the batch
		resp1, _ := createTestResponse("1", "result1", nil)
		resp2, _ := createTestResponse("2", "result2", nil)

		// Write first response
		err = conn.Write(ctx, resp1)
		if err != nil {
			t.Fatalf("Write() first response error: %v", err)
		}

		// At this point, nothing should be written to rwc yet (batch incomplete)
		written := rwc.getWrittenData()
		if len(written) > 0 {
			t.Error("Response should not be written until batch is complete")
		}

		// Write second response
		err = conn.Write(ctx, resp2)
		if err != nil {
			t.Fatalf("Write() second response error: %v", err)
		}

		// Now the complete batch should be written
		written = rwc.getWrittenData()
		if len(written) == 0 {
			t.Fatal("Batch response should be written")
		}

		// Should be a JSON array + newline
		if !bytes.HasPrefix(written, []byte("[")) {
			t.Error("Batch response should start with '['")
		}
		if !bytes.HasSuffix(written, []byte("]\n")) {
			t.Error("Batch response should end with ']\\n'")
		}
	})

	t.Run("duplicate batch ID error", func(t *testing.T) {
		t.Parallel()

		// Create a batch with duplicate IDs (invalid)
		jsonBatch := `[
			{"jsonrpc":"2.0","method":"test1","id":"1"},
			{"jsonrpc":"2.0","method":"test2","id":"1"}
		]` + "\n"

		rwc := newMockReadWriteCloser([]byte(jsonBatch))
		conn := newIOConn(rwc)

		ctx := t.Context()

		// First read should fail immediately due to duplicate ID in batch
		_, err := conn.Read(ctx)
		if err == nil {
			t.Fatal("Expected error for duplicate batch ID")
		}

		if !strings.Contains(err.Error(), "duplicate message ID") {
			t.Errorf("Expected duplicate message ID error, got: %v", err)
		}
	})
}

func TestIOConnClose(t *testing.T) {
	t.Parallel()

	t.Run("successful close", func(t *testing.T) {
		t.Parallel()

		rwc := newMockReadWriteCloser(nil)
		conn := newIOConn(rwc)

		err := conn.Close()
		if err != nil {
			t.Fatalf("Close() error: %v", err)
		}
	})

	t.Run("close error", func(t *testing.T) {
		t.Parallel()

		rwc := newMockReadWriteCloser(nil)
		rwc.setCloseError(errors.New("close failed"))
		conn := newIOConn(rwc)

		err := conn.Close()
		if err == nil {
			t.Fatal("Expected close error")
		}

		if !strings.Contains(err.Error(), "close failed") {
			t.Errorf("Error should contain 'close failed', got: %v", err)
		}
	})
}

func TestMarshalMessages(t *testing.T) {
	t.Parallel()

	t.Run("marshal single message", func(t *testing.T) {
		t.Parallel()

		msg, _ := createTestRequest("1", "test", nil)
		messages := []*jsonrpc2.Request{msg}

		data, err := marshalMessages(messages)
		if err != nil {
			t.Fatalf("marshalMessages() error: %v", err)
		}

		if len(data) == 0 {
			t.Fatal("No data marshaled")
		}

		// Should be a JSON array
		if !bytes.HasPrefix(data, []byte("[")) {
			t.Error("Marshaled data should start with '['")
		}
		if !bytes.HasSuffix(data, []byte("]")) {
			t.Error("Marshaled data should end with ']'")
		}
	})

	t.Run("marshal multiple messages", func(t *testing.T) {
		t.Parallel()

		msg1, _ := createTestRequest("1", "test1", nil)
		msg2, _ := createTestRequest("2", "test2", nil)
		messages := []*jsonrpc2.Request{msg1, msg2}

		data, err := marshalMessages(messages)
		if err != nil {
			t.Fatalf("marshalMessages() error: %v", err)
		}

		if len(data) == 0 {
			t.Fatal("No data marshaled")
		}

		// Should contain both methods
		dataStr := string(data)
		if !strings.Contains(dataStr, "test1") {
			t.Error("Should contain 'test1'")
		}
		if !strings.Contains(dataStr, "test2") {
			t.Error("Should contain 'test2'")
		}
	})

	t.Run("marshal empty slice", func(t *testing.T) {
		t.Parallel()

		messages := []*jsonrpc2.Request{}

		data, err := marshalMessages(messages)
		if err != nil {
			t.Fatalf("marshalMessages() error: %v", err)
		}

		if string(data) != "[]" {
			t.Errorf("Empty slice should marshal to '[]', got: %q", string(data))
		}
	})
}

func TestIOConnSessionID(t *testing.T) {
	t.Parallel()

	t.Run("returns empty session ID", func(t *testing.T) {
		t.Parallel()

		rwc := newMockReadWriteCloser(nil)
		conn := newIOConn(rwc)

		sessionID := conn.SessionID()
		if sessionID != "" {
			t.Errorf("SessionID() = %q, want empty string", sessionID)
		}
	})
}
