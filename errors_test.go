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

package a2a

import (
	"errors"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

func TestNewError(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		code    int64
		message string
		want    *Error
	}{
		"positive code": {
			code:    100,
			message: "test error",
			want: &Error{
				Code:    100,
				Message: "test error",
			},
		},
		"negative code": {
			code:    -32600,
			message: "JSON RPC invalid request",
			want: &Error{
				Code:    -32600,
				Message: "JSON RPC invalid request",
			},
		},
		"zero code": {
			code:    0,
			message: "zero error",
			want: &Error{
				Code:    0,
				Message: "zero error",
			},
		},
		"empty message": {
			code:    1,
			message: "",
			want: &Error{
				Code:    1,
				Message: "",
			},
		},
		"long message": {
			code:    999,
			message: "this is a very long error message that exceeds typical message lengths to test handling of extended error descriptions",
			want: &Error{
				Code:    999,
				Message: "this is a very long error message that exceeds typical message lengths to test handling of extended error descriptions",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := NewError(tt.code, tt.message)
			if got.Code != tt.want.Code {
				t.Errorf("NewError().Code = %v, want %v", got.Code, tt.want.Code)
			}
			if got.Message != tt.want.Message {
				t.Errorf("NewError().Message = %v, want %v", got.Message, tt.want.Message)
			}
		})
	}
}

func TestPredefinedStandardJSONRPCErrors(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		name string
		err  *Error
		code int64
		msg  string
	}{
		"ErrParse": {
			err:  ErrParse,
			code: -32700,
			msg:  "JSON RPC parse error",
		},
		"ErrInvalidRequest": {
			err:  ErrInvalidRequest,
			code: -32600,
			msg:  "JSON RPC invalid request",
		},
		"ErrMethodNotFound": {
			err:  ErrMethodNotFound,
			code: -32601,
			msg:  "JSON RPC method not found",
		},
		"ErrInvalidParams": {
			err:  ErrInvalidParams,
			code: -32602,
			msg:  "JSON RPC invalid params",
		},
		"ErrInternal": {
			err:  ErrInternal,
			code: -32603,
			msg:  "JSON RPC internal error",
		},
		"ErrServerOverloaded": {
			err:  ErrServerOverloaded,
			code: -32000,
			msg:  "JSON RPC overloaded",
		},
		"ErrUnknown": {
			err:  ErrUnknown,
			code: -32001,
			msg:  "JSON RPC unknown error",
		},
		"ErrServerClosing": {
			err:  ErrServerClosing,
			code: -32004,
			msg:  "JSON RPC server is closing",
		},
		"ErrClientClosing": {
			err:  ErrClientClosing,
			code: -32003,
			msg:  "JSON RPC client is closing",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tt.err.Code != tt.code {
				t.Errorf("%s.Code = %v, want %v", tt.name, tt.err.Code, tt.code)
			}
			if tt.err.Message != tt.msg {
				t.Errorf("%s.Message = %v, want %v", tt.name, tt.err.Message, tt.msg)
			}
		})
	}
}

func TestPredefinedA2ASpecificErrors(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err  *Error
		code int64
		msg  string
	}{
		"ErrTaskNotFound": {
			err:  ErrTaskNotFound,
			code: -32001,
			msg:  "JSON RPC task not found",
		},
		"ErrTaskNotCancelable": {
			err:  ErrTaskNotCancelable,
			code: -32002,
			msg:  "JSON RPC task not cancelable",
		},
		"ErrPushNotificationNotSupported": {
			err:  ErrPushNotificationNotSupported,
			code: -32003,
			msg:  "JSON RPC push notification not supported",
		},
		"ErrUnsupportedOperation": {
			err:  ErrUnsupportedOperation,
			code: -32004,
			msg:  "JSON RPC unsupported operation",
		},
		"ErrContentTypeNotSupported": {
			err:  ErrContentTypeNotSupported,
			code: -32005,
			msg:  "JSON RPC content type not supported",
		},
		"ErrInvalidAgentResponse": {
			err:  ErrInvalidAgentResponse,
			code: -32006,
			msg:  "JSON RPC invalid agent response",
		},
		"ErrMethodNotImplemented": {
			err:  ErrMethodNotImplemented,
			code: -32099,
			msg:  "JSON RPC method not implemented",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tt.err.Code != tt.code {
				t.Errorf("%s.Code = %v, want %v", name, tt.err.Code, tt.code)
			}
			if tt.err.Message != tt.msg {
				t.Errorf("%s.Message = %v, want %v", name, tt.err.Message, tt.msg)
			}
		})
	}
}

func TestErrorString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err     *Error
		wantStr string
	}{
		"standard error": {
			err:     NewError(-32600, "Invalid request"),
			wantStr: "Invalid request",
		},
		"predefined error": {
			err:     ErrParse,
			wantStr: "JSON RPC parse error",
		},
		"empty message": {
			err:     NewError(1, ""),
			wantStr: "",
		},
		"A2A specific error": {
			err:     ErrTaskNotFound,
			wantStr: "JSON RPC task not found",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tt.err.Error(); got != tt.wantStr {
				t.Errorf("Error.Error() = %v, want %v", got, tt.wantStr)
			}
		})
	}
}

func TestErrorIs(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err    *Error
		other  error
		wantIs bool
	}{
		"same error instance": {
			err:    ErrParse,
			other:  ErrParse,
			wantIs: true,
		},
		"same code different instance": {
			err:    NewError(-32700, "Different message"),
			other:  ErrParse,
			wantIs: true,
		},
		"different code same message": {
			err:    NewError(-32600, "JSON RPC parse error"),
			other:  ErrParse,
			wantIs: false,
		},
		"different error types": {
			err:    ErrParse,
			other:  errors.New("standard error"),
			wantIs: false,
		},
		"nil error": {
			err:    ErrParse,
			other:  nil,
			wantIs: false,
		},
		"A2A specific errors": {
			err:    ErrTaskNotFound,
			other:  NewError(-32001, "Different message but same code"),
			wantIs: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := errors.Is(tt.err, tt.other); got != tt.wantIs {
				t.Errorf("errors.Is(%v, %v) = %v, want %v", tt.err, tt.other, got, tt.wantIs)
			}
		})
	}
}

func TestErrorJSONMarshaling(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err      *Error
		wantJSON string
	}{
		"basic error": {
			err:      NewError(-32600, "Invalid request"),
			wantJSON: `{"code":-32600,"message":"Invalid request"}`,
		},
		"error with zero code": {
			err:      NewError(0, "Zero code error"),
			wantJSON: `{"code":0,"message":"Zero code error"}`,
		},
		"error with empty message": {
			err:      NewError(100, ""),
			wantJSON: `{"code":100,"message":""}`,
		},
		"predefined error": {
			err:      ErrParse,
			wantJSON: `{"code":-32700,"message":"JSON RPC parse error"}`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tt.err)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if string(got) != tt.wantJSON {
				t.Errorf("json.Marshal() = %s, want %s", got, tt.wantJSON)
			}
		})
	}
}

func TestErrorJSONUnmarshaling(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		jsonStr  string
		wantErr  *Error
		wantFail bool
	}{
		"basic error": {
			jsonStr: `{"code":-32600,"message":"Invalid request"}`,
			wantErr: &Error{Code: -32600, Message: "Invalid request"},
		},
		"error with data field": {
			jsonStr: `{"code":-32602,"message":"Invalid params","data":"additional info"}`,
			wantErr: &Error{
				Code:    -32602,
				Message: "Invalid params",
				Data:    jsontext.Value(`"additional info"`),
			},
		},
		"error with zero code": {
			jsonStr: `{"code":0,"message":"Zero error"}`,
			wantErr: &Error{Code: 0, Message: "Zero error"},
		},
		"invalid json": {
			jsonStr:  `{"code":"invalid","message":"test"}`,
			wantFail: true,
		},
		"missing code field": {
			jsonStr: `{"message":"test"}`,
			wantErr: &Error{Code: 0, Message: "test"},
		},
		"missing message field": {
			jsonStr: `{"code":-32000}`,
			wantErr: &Error{Code: -32000, Message: ""},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got Error
			err := json.Unmarshal([]byte(tt.jsonStr), &got)

			if tt.wantFail {
				if err == nil {
					t.Errorf("json.Unmarshal() should have failed but didn't")
				}
				return
			}

			if err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if got.Code != tt.wantErr.Code {
				t.Errorf("Error.Code = %v, want %v", got.Code, tt.wantErr.Code)
			}
			if got.Message != tt.wantErr.Message {
				t.Errorf("Error.Message = %v, want %v", got.Message, tt.wantErr.Message)
			}
			if string(got.Data) != string(tt.wantErr.Data) {
				t.Errorf("Error.Data = %s, want %s", got.Data, tt.wantErr.Data)
			}
		})
	}
}

func TestErrorRoundTripJSONMarshaling(t *testing.T) {
	t.Parallel()

	tests := []*Error{
		NewError(-32600, "Invalid request"),
		NewError(0, "Zero code"),
		NewError(100, ""),
		ErrParse,
		ErrTaskNotFound,
	}

	for i, original := range tests {
		t.Run(string(rune(i+'A')), func(t *testing.T) {
			t.Parallel()

			// Marshal to JSON
			jsonBytes, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			// Unmarshal back to struct
			var unmarshaled Error
			err = json.Unmarshal(jsonBytes, &unmarshaled)
			if err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			// Compare
			if unmarshaled.Code != original.Code {
				t.Errorf("Code: got %v, want %v", unmarshaled.Code, original.Code)
			}
			if unmarshaled.Message != original.Message {
				t.Errorf("Message: got %v, want %v", unmarshaled.Message, original.Message)
			}
			if string(unmarshaled.Data) != string(original.Data) {
				t.Errorf("Data: got %s, want %s", unmarshaled.Data, original.Data)
			}
		})
	}
}

func TestErrorEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("large code values", func(t *testing.T) {
		t.Parallel()

		largePositive := NewError(9223372036854775807, "max int64")
		largeNegative := NewError(-9223372036854775808, "min int64")

		if largePositive.Code != 9223372036854775807 {
			t.Errorf("Large positive code not preserved")
		}
		if largeNegative.Code != -9223372036854775808 {
			t.Errorf("Large negative code not preserved")
		}
	})

	t.Run("unicode in message", func(t *testing.T) {
		t.Parallel()

		unicodeErr := NewError(100, "错误信息 🚨 émission")
		if unicodeErr.Message != "错误信息 🚨 émission" {
			t.Errorf("Unicode message not preserved: got %s", unicodeErr.Message)
		}

		// Test JSON round-trip with unicode
		jsonBytes, err := json.Marshal(unicodeErr)
		if err != nil {
			t.Fatalf("Marshal unicode error: %v", err)
		}

		var unmarshaled Error
		err = json.Unmarshal(jsonBytes, &unmarshaled)
		if err != nil {
			t.Fatalf("Unmarshal unicode error: %v", err)
		}

		if unmarshaled.Message != unicodeErr.Message {
			t.Errorf("Unicode message corrupted in JSON round-trip")
		}
	})

	t.Run("very long message", func(t *testing.T) {
		t.Parallel()

		longMsg := string(make([]byte, 10000))
		for i := range longMsg {
			longMsg = longMsg[:i] + "a" + longMsg[i+1:]
		}

		longErr := NewError(200, longMsg)
		if len(longErr.Message) != 10000 {
			t.Errorf("Long message length not preserved: got %d, want 10000", len(longErr.Message))
		}
	})
}

func TestErrorInterfaceCompliance(t *testing.T) {
	t.Parallel()

	var err error = NewError(100, "test error")

	if err.Error() != "test error" {
		t.Errorf("Error interface method failed")
	}

	// Test with errors.Is
	testErr := NewError(100, "different message")
	if !errors.Is(err, testErr) {
		t.Errorf("errors.Is should work with same code")
	}
}
