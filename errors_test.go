// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"errors"
	"testing"
)

func TestServerError(t *testing.T) {
	taskErr := &TaskNotFoundError{TaskID: "test-task"}

	tests := map[string]struct {
		setupError    func() *ServerError
		expectedError string
		expectedCode  int
		expectedMsg   string
		expectedCause A2AError
	}{
		"new server error": {
			setupError: func() *ServerError {
				return NewServerError(taskErr)
			},
			expectedError: "server error: task not found: test-task",
			expectedCode:  ErrorCodeTaskNotFound,
			expectedMsg:   "A2A specific error indicating the requested task ID was not found",
			expectedCause: taskErr,
		},
		"new server error with message": {
			setupError: func() *ServerError {
				return NewServerErrorWithMessage(taskErr, "additional context")
			},
			expectedError: "server error: additional context: task not found: test-task",
			expectedCode:  ErrorCodeTaskNotFound,
			expectedMsg:   "A2A specific error indicating the requested task ID was not found",
			expectedCause: taskErr,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.setupError()

			if got := err.Error(); got != tt.expectedError {
				t.Errorf("ServerError.Error() = %q, want %q", got, tt.expectedError)
			}

			if got := err.Code(); got != tt.expectedCode {
				t.Errorf("ServerError.Code() = %d, want %d", got, tt.expectedCode)
			}

			if got := err.Message(); got != tt.expectedMsg {
				t.Errorf("ServerError.Message() = %q, want %q", got, tt.expectedMsg)
			}

			if got := err.Cause(); got != tt.expectedCause {
				t.Errorf("ServerError.Cause() = %v, want %v", got, tt.expectedCause)
			}
		})
	}
}

func TestServerErrorUnwrap(t *testing.T) {
	taskErr := &TaskNotFoundError{TaskID: "test-task"}
	serverErr := NewServerError(taskErr)

	// Test Unwrap method
	unwrapped := serverErr.Unwrap()
	if unwrapped != taskErr {
		t.Errorf("ServerError.Unwrap() = %v, want %v", unwrapped, taskErr)
	}

	// Test with errors.Is
	if !errors.Is(serverErr, taskErr) {
		t.Error("errors.Is(serverErr, taskErr) should return true")
	}

	// Test with errors.As
	var retrievedTaskErr *TaskNotFoundError
	if !errors.As(serverErr, &retrievedTaskErr) {
		t.Error("errors.As(serverErr, &retrievedTaskErr) should return true")
	}

	if retrievedTaskErr.TaskID != "test-task" {
		t.Errorf("Retrieved task ID = %q, want %q", retrievedTaskErr.TaskID, "test-task")
	}
}

func TestServerErrorIs(t *testing.T) {
	taskErr := &TaskNotFoundError{TaskID: "test-task"}
	serverErr := NewServerError(taskErr)

	tests := map[string]struct {
		target   error
		expected bool
	}{
		"same task error": {
			target:   taskErr,
			expected: true,
		},
		"different task error": {
			target:   &TaskNotFoundError{TaskID: "different-task"},
			expected: false,
		},
		"server error with same cause": {
			target:   NewServerError(taskErr),
			expected: true,
		},
		"different error type": {
			target:   &JSONParseError{Msg: "parse error"},
			expected: false,
		},
		"nil error": {
			target:   nil,
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := serverErr.Is(tt.target); got != tt.expected {
				t.Errorf("ServerError.Is(%v) = %v, want %v", tt.target, got, tt.expected)
			}
		})
	}
}

func TestServerErrorAs(t *testing.T) {
	taskErr := &TaskNotFoundError{TaskID: "test-task"}
	serverErr := NewServerError(taskErr)

	// Test As with ServerError
	var retrievedServerErr *ServerError
	if !serverErr.As(&retrievedServerErr) {
		t.Error("ServerError.As(&retrievedServerErr) should return true")
	}

	if retrievedServerErr != serverErr {
		t.Error("Retrieved server error should be the same instance")
	}

	// Test As with underlying error
	var retrievedTaskErr *TaskNotFoundError
	if !serverErr.As(&retrievedTaskErr) {
		t.Error("ServerError.As(&retrievedTaskErr) should return true")
	}

	if retrievedTaskErr.TaskID != "test-task" {
		t.Errorf("Retrieved task ID = %q, want %q", retrievedTaskErr.TaskID, "test-task")
	}

	// Test As with different error type
	var jsonErr *JSONParseError
	if serverErr.As(&jsonErr) {
		t.Error("ServerError.As(&jsonErr) should return false")
	}
}

func TestMethodNotImplementedError(t *testing.T) {
	err := &MethodNotImplementedError{Method: "TestMethod"}

	expectedError := "method not implemented: TestMethod"
	if got := err.Error(); got != expectedError {
		t.Errorf("MethodNotImplementedError.Error() = %q, want %q", got, expectedError)
	}

	expectedCode := ErrorCodeMethodNotImplemented
	if got := err.Code(); got != expectedCode {
		t.Errorf("MethodNotImplementedError.Code() = %d, want %d", got, expectedCode)
	}

	expectedMsg := "Method not implemented"
	if got := err.Message(); got != expectedMsg {
		t.Errorf("MethodNotImplementedError.Message() = %q, want %q", got, expectedMsg)
	}
}

func TestUnsupportedOperationError(t *testing.T) {
	err := &UnsupportedOperationError{Operation: "TestOperation"}

	expectedError := "unsupported operation: TestOperation"
	if got := err.Error(); got != expectedError {
		t.Errorf("UnsupportedOperationError.Error() = %q, want %q", got, expectedError)
	}

	expectedCode := ErrorCodeUnsupportedOperation
	if got := err.Code(); got != expectedCode {
		t.Errorf("UnsupportedOperationError.Code() = %d, want %d", got, expectedCode)
	}

	expectedMsg := "Unsupported operation"
	if got := err.Message(); got != expectedMsg {
		t.Errorf("UnsupportedOperationError.Message() = %q, want %q", got, expectedMsg)
	}
}

func TestIsServerError(t *testing.T) {
	taskErr := &TaskNotFoundError{TaskID: "test-task"}
	serverErr := NewServerError(taskErr)

	tests := map[string]struct {
		err      error
		expected bool
	}{
		"server error": {
			err:      serverErr,
			expected: true,
		},
		"task error": {
			err:      taskErr,
			expected: false,
		},
		"nil error": {
			err:      nil,
			expected: false,
		},
		"wrapped server error": {
			err:      NewServerError(serverErr),
			expected: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := IsServerError(tt.err); got != tt.expected {
				t.Errorf("IsServerError(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}

func TestAsServerError(t *testing.T) {
	taskErr := &TaskNotFoundError{TaskID: "test-task"}
	serverErr := NewServerError(taskErr)

	tests := map[string]struct {
		err           error
		expectedErr   *ServerError
		expectedFound bool
	}{
		"server error": {
			err:           serverErr,
			expectedErr:   serverErr,
			expectedFound: true,
		},
		"task error": {
			err:           taskErr,
			expectedErr:   nil,
			expectedFound: false,
		},
		"nil error": {
			err:           nil,
			expectedErr:   nil,
			expectedFound: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotErr, gotFound := AsServerError(tt.err)

			if gotFound != tt.expectedFound {
				t.Errorf("AsServerError(%v) found = %v, want %v", tt.err, gotFound, tt.expectedFound)
			}

			if gotErr != tt.expectedErr {
				t.Errorf("AsServerError(%v) error = %v, want %v", tt.err, gotErr, tt.expectedErr)
			}
		})
	}
}

func TestWrapA2AError(t *testing.T) {
	taskErr := &TaskNotFoundError{TaskID: "test-task"}
	serverErr := NewServerError(taskErr)
	regularErr := errors.New("regular error")

	tests := map[string]struct {
		err           error
		msg           string
		expectedType  string
		expectedError string
	}{
		"wrap task error": {
			err:           taskErr,
			msg:           "",
			expectedType:  "*a2a.ServerError",
			expectedError: "server error: task not found: test-task",
		},
		"wrap task error with message": {
			err:           taskErr,
			msg:           "context message",
			expectedType:  "*a2a.ServerError",
			expectedError: "server error: context message: task not found: test-task",
		},
		"wrap server error": {
			err:           serverErr,
			msg:           "",
			expectedType:  "*a2a.ServerError",
			expectedError: "server error: task not found: test-task",
		},
		"wrap regular error": {
			err:           regularErr,
			msg:           "",
			expectedType:  "*a2a.ServerError",
			expectedError: "server error: internal error: regular error",
		},
		"wrap regular error with message": {
			err:           regularErr,
			msg:           "context message",
			expectedType:  "*a2a.ServerError",
			expectedError: "server error: context message: internal error: regular error",
		},
		"wrap nil error": {
			err:           nil,
			msg:           "",
			expectedType:  "<nil>",
			expectedError: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := WrapA2AError(tt.err, tt.msg)

			if got == nil {
				if tt.expectedType != "<nil>" {
					t.Errorf("WrapA2AError(%v, %q) = nil, want %s", tt.err, tt.msg, tt.expectedType)
				}
				return
			}

			if got.Error() != tt.expectedError {
				t.Errorf("WrapA2AError(%v, %q).Error() = %q, want %q", tt.err, tt.msg, got.Error(), tt.expectedError)
			}
		})
	}
}

func TestIsMethodNotImplementedError(t *testing.T) {
	methodErr := &MethodNotImplementedError{Method: "TestMethod"}
	taskErr := &TaskNotFoundError{TaskID: "test-task"}

	tests := map[string]struct {
		err      error
		expected bool
	}{
		"method not implemented error": {
			err:      methodErr,
			expected: true,
		},
		"task error": {
			err:      taskErr,
			expected: false,
		},
		"nil error": {
			err:      nil,
			expected: false,
		},
		"wrapped method error": {
			err:      NewServerError(methodErr),
			expected: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := IsMethodNotImplementedError(tt.err); got != tt.expected {
				t.Errorf("IsMethodNotImplementedError(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}

func TestAsMethodNotImplementedError(t *testing.T) {
	methodErr := &MethodNotImplementedError{Method: "TestMethod"}
	taskErr := &TaskNotFoundError{TaskID: "test-task"}

	tests := map[string]struct {
		err           error
		expectedErr   *MethodNotImplementedError
		expectedFound bool
	}{
		"method not implemented error": {
			err:           methodErr,
			expectedErr:   methodErr,
			expectedFound: true,
		},
		"task error": {
			err:           taskErr,
			expectedErr:   nil,
			expectedFound: false,
		},
		"nil error": {
			err:           nil,
			expectedErr:   nil,
			expectedFound: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotErr, gotFound := AsMethodNotImplementedError(tt.err)

			if gotFound != tt.expectedFound {
				t.Errorf("AsMethodNotImplementedError(%v) found = %v, want %v", tt.err, gotFound, tt.expectedFound)
			}

			if gotErr != tt.expectedErr {
				t.Errorf("AsMethodNotImplementedError(%v) error = %v, want %v", tt.err, gotErr, tt.expectedErr)
			}
		})
	}
}

func TestIsUnsupportedOperationError(t *testing.T) {
	opErr := &UnsupportedOperationError{Operation: "TestOperation"}
	taskErr := &TaskNotFoundError{TaskID: "test-task"}

	tests := map[string]struct {
		err      error
		expected bool
	}{
		"unsupported operation error": {
			err:      opErr,
			expected: true,
		},
		"task error": {
			err:      taskErr,
			expected: false,
		},
		"nil error": {
			err:      nil,
			expected: false,
		},
		"wrapped operation error": {
			err:      NewServerError(opErr),
			expected: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := IsUnsupportedOperationError(tt.err); got != tt.expected {
				t.Errorf("IsUnsupportedOperationError(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}

func TestAsUnsupportedOperationError(t *testing.T) {
	opErr := &UnsupportedOperationError{Operation: "TestOperation"}
	taskErr := &TaskNotFoundError{TaskID: "test-task"}

	tests := map[string]struct {
		err           error
		expectedErr   *UnsupportedOperationError
		expectedFound bool
	}{
		"unsupported operation error": {
			err:           opErr,
			expectedErr:   opErr,
			expectedFound: true,
		},
		"task error": {
			err:           taskErr,
			expectedErr:   nil,
			expectedFound: false,
		},
		"nil error": {
			err:           nil,
			expectedErr:   nil,
			expectedFound: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotErr, gotFound := AsUnsupportedOperationError(tt.err)

			if gotFound != tt.expectedFound {
				t.Errorf("AsUnsupportedOperationError(%v) found = %v, want %v", tt.err, gotFound, tt.expectedFound)
			}

			if gotErr != tt.expectedErr {
				t.Errorf("AsUnsupportedOperationError(%v) error = %v, want %v", tt.err, gotErr, tt.expectedErr)
			}
		})
	}
}

func TestGetA2AErrorCode(t *testing.T) {
	taskErr := &TaskNotFoundError{TaskID: "test-task"}
	serverErr := NewServerError(taskErr)
	regularErr := errors.New("regular error")

	tests := map[string]struct {
		err          error
		expectedCode int
		expectedOK   bool
	}{
		"task error": {
			err:          taskErr,
			expectedCode: ErrorCodeTaskNotFound,
			expectedOK:   true,
		},
		"server error": {
			err:          serverErr,
			expectedCode: ErrorCodeTaskNotFound,
			expectedOK:   true,
		},
		"regular error": {
			err:          regularErr,
			expectedCode: 0,
			expectedOK:   false,
		},
		"nil error": {
			err:          nil,
			expectedCode: 0,
			expectedOK:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotCode, gotOK := GetA2AErrorCode(tt.err)

			if gotOK != tt.expectedOK {
				t.Errorf("GetA2AErrorCode(%v) ok = %v, want %v", tt.err, gotOK, tt.expectedOK)
			}

			if gotCode != tt.expectedCode {
				t.Errorf("GetA2AErrorCode(%v) code = %d, want %d", tt.err, gotCode, tt.expectedCode)
			}
		})
	}
}

func TestGetA2AErrorMessage(t *testing.T) {
	taskErr := &TaskNotFoundError{TaskID: "test-task"}
	serverErr := NewServerError(taskErr)
	regularErr := errors.New("regular error")

	tests := map[string]struct {
		err             error
		expectedMessage string
		expectedOK      bool
	}{
		"task error": {
			err:             taskErr,
			expectedMessage: "A2A specific error indicating the requested task ID was not found",
			expectedOK:      true,
		},
		"server error": {
			err:             serverErr,
			expectedMessage: "A2A specific error indicating the requested task ID was not found",
			expectedOK:      true,
		},
		"regular error": {
			err:             regularErr,
			expectedMessage: "",
			expectedOK:      false,
		},
		"nil error": {
			err:             nil,
			expectedMessage: "",
			expectedOK:      false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotMessage, gotOK := GetA2AErrorMessage(tt.err)

			if gotOK != tt.expectedOK {
				t.Errorf("GetA2AErrorMessage(%v) ok = %v, want %v", tt.err, gotOK, tt.expectedOK)
			}

			if gotMessage != tt.expectedMessage {
				t.Errorf("GetA2AErrorMessage(%v) message = %q, want %q", tt.err, gotMessage, tt.expectedMessage)
			}
		})
	}
}

func TestErrorCodes(t *testing.T) {
	tests := map[string]struct {
		name         string
		code         int
		expectedCode int
	}{
		"method not implemented": {
			code:         ErrorCodeMethodNotImplemented,
			expectedCode: -32004,
		},
		"unsupported operation": {
			code:         ErrorCodeUnsupportedOperation,
			expectedCode: -32005,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.code != tt.expectedCode {
				t.Errorf("Error code %s = %d, want %d", tt.name, tt.code, tt.expectedCode)
			}
		})
	}
}

func TestErrorChaining(t *testing.T) {
	// Test complex error chaining scenario
	taskErr := &TaskNotFoundError{TaskID: "test-task"}
	serverErr := NewServerError(taskErr)
	wrappedErr := NewServerError(serverErr)

	// Should be able to unwrap through multiple layers
	if !errors.Is(wrappedErr, taskErr) {
		t.Error("Complex error chain should support errors.Is")
	}

	var retrievedTaskErr *TaskNotFoundError
	if !errors.As(wrappedErr, &retrievedTaskErr) {
		t.Error("Complex error chain should support errors.As")
	}

	if retrievedTaskErr.TaskID != "test-task" {
		t.Errorf("Retrieved task ID = %q, want %q", retrievedTaskErr.TaskID, "test-task")
	}
}

// Benchmarks

func BenchmarkServerErrorCreation(b *testing.B) {
	taskErr := &TaskNotFoundError{TaskID: "test-task"}

	for b.Loop() {
		_ = NewServerError(taskErr)
	}
}

func BenchmarkServerErrorWithMessage(b *testing.B) {
	taskErr := &TaskNotFoundError{TaskID: "test-task"}

	for b.Loop() {
		_ = NewServerErrorWithMessage(taskErr, "context message")
	}
}

func BenchmarkWrapA2AError(b *testing.B) {
	taskErr := &TaskNotFoundError{TaskID: "test-task"}

	for b.Loop() {
		_ = WrapA2AError(taskErr, "context message")
	}
}

func BenchmarkErrorUnwrapping(b *testing.B) {
	taskErr := &TaskNotFoundError{TaskID: "test-task"}
	serverErr := NewServerError(taskErr)

	for b.Loop() {
		var retrievedErr *TaskNotFoundError
		_ = errors.As(serverErr, &retrievedErr)
	}
}

func BenchmarkGetA2AErrorCode(b *testing.B) {
	taskErr := &TaskNotFoundError{TaskID: "test-task"}
	serverErr := NewServerError(taskErr)

	for b.Loop() {
		_, _ = GetA2AErrorCode(serverErr)
	}
}
