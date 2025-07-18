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
	"github.com/go-a2a/a2a-go/internal/jsonrpc2"
)

// Error type alias of [jsonrpc2.WireError].
type Error = jsonrpc2.WireError

// NewError returns an error that will encode on the wire correctly.
// The standard codes are made available from this package, this function should
// only be used to build errors for application specific codes as allowed by the
// specification.
//
// NewError creates a new [Error] with the specified code and message.
func NewError(code int64, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// List of Standard JSON-RPC Errors.
var (
	// ErrParse is used when invalid JSON was received by the server.
	ErrParse = NewError(-32700, "JSON RPC parse error")
	// ErrInvalidRequest is used when the JSON sent is not a valid Request object.
	ErrInvalidRequest = NewError(-32600, "JSON RPC invalid request")
	// ErrMethodNotFound should be returned by the handler when the method does
	// not exist / is not available.
	ErrMethodNotFound = NewError(-32601, "JSON RPC method not found")
	// ErrInvalidParams should be returned by the handler when method
	// parameter(s) were invalid.
	ErrInvalidParams = NewError(-32602, "JSON RPC invalid params")
	// ErrInternal indicates a failure to process a call correctly
	ErrInternal = NewError(-32603, "JSON RPC internal error")

	// The following errors are not part of the json specification, but
	// compliant extensions specific to this implementation.
	//
	// TODO(zchee): Are these errors also necessary for A2A?

	// ErrServerOverloaded is returned when a message was refused due to a
	// server being temporarily unable to accept any new messages.
	ErrServerOverloaded = NewError(-32000, "JSON RPC overloaded")
	// ErrUnknown should be used for all non coded errors.
	ErrUnknown = NewError(-32001, "JSON RPC unknown error")
	// ErrServerClosing is returned for calls that arrive while the server is closing.
	ErrServerClosing = NewError(-32004, "JSON RPC server is closing")
	// ErrClientClosing is a dummy error returned for calls initiated while the client is closing.
	ErrClientClosing = NewError(-32003, "JSON RPC client is closing")
)

// List of A2A-Specific Errors.
var (
	// ErrTaskNotFound is returned when a specified task ID does not correspond to an existing or active task.
	// It might be invalid, expired, or already completed and purged.
	ErrTaskNotFound = NewError(-32001, "JSON RPC task not found")

	// ErrTaskNotCancelable is returned when an attempt is made to cancel a task that is not in a cancelable state (e.g., it has already reached a terminal state like completed, failed, or canceled).
	ErrTaskNotCancelable = NewError(-32002, "JSON RPC task not cancelable")

	// ErrPushNotificationNotSupported is returned when a client attempted to use push notification features (e.g., tasks/pushNotificationConfig/set)
	// but the server agent does not support them (i.e., AgentCard.capabilities.pushNotifications is false).
	ErrPushNotificationNotSupported = NewError(-32003, "JSON RPC push notification not supported")

	// ErrUnsupportedOperation is returned when a requested operation or
	// a specific aspect of it (perhaps implied by parameters) is not supported by
	// this server agent implementation. Broader than just method not found.
	ErrUnsupportedOperation = NewError(-32004, "JSON RPC unsupported operation")

	// ErrContentTypeNotSupported is returned when a Media Type is provided in the
	// request's message.parts (or implied for an artifact) that is not supported by the agent or
	// the specific skill being invoked.
	ErrContentTypeNotSupported = NewError(-32005, "JSON RPC content type not supported")

	// ErrInvalidAgentResponse is returned when an agent generated an invalid response for the requested method.
	ErrInvalidAgentResponse = NewError(-32006, "JSON RPC invalid agent response")

	// ErrMethodNotImplemented is returned when an agent generates a method that is
	// not implemented for the requested method.
	ErrMethodNotImplemented = NewError(-32099, "JSON RPC method not implemented")
)
