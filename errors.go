// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"github.com/go-a2a/a2a/internal/jsonrpc2"
)

// NewJSONParseError creates a new JSONParseError.
func NewJSONParseError() error {
	return jsonrpc2.NewError(JSONParseErrorCode, "Invalid JSON payload")
}

// NewInvalidRequestError creates a new InvalidRequestError.
func NewInvalidRequestError() error {
	return jsonrpc2.NewError(InvalidRequestErrorCode, "Request payload validation error")
}

// NewMethodNotFoundError creates a new MethodNotFoundError.
func NewMethodNotFoundError() error {
	return jsonrpc2.NewError(MethodNotFoundErrorCode, "Method not found")
}

// NewInvalidParamsError creates a new InvalidParamsError.
func NewInvalidParamsError() error {
	return jsonrpc2.NewError(InvalidParamsErrorCode, "Invalid parameters")
}

// NewInternalError creates a new InternalError.
func NewInternalError() error {
	return jsonrpc2.NewError(InternalErrorCode, "Internal error")
}

// NewTaskNotFoundError creates a new TaskNotFoundError.
func NewTaskNotFoundError() error {
	return jsonrpc2.NewError(TaskNotFoundErrorCode, "Task not found")
}

// NewTaskNotCancelableError creates a new TaskNotCancelableError.
func NewTaskNotCancelableError() error {
	return jsonrpc2.NewError(TaskNotCancelableErrorCode, "Task cannot be canceled")
}

// NewPushNotificationNotSupportedError creates a new PushNotificationNotSupportedError.
func NewPushNotificationNotSupportedError() error {
	return jsonrpc2.NewError(PushNotificationNotSupportedErrorCode, "Push Notification is not supported")
}

// NewUnsupportedOperationError creates a new UnsupportedOperationError.
func NewUnsupportedOperationError() error {
	return jsonrpc2.NewError(UnsupportedOperationErrorCode, "This operation is not supported")
}

// NewContentTypeNotSupportedError creates a new ContentTypeNotSupportedError.
func NewContentTypeNotSupportedError() error {
	return jsonrpc2.NewError(ContentTypeNotSupportedErrorCode, "Content type not supported")
}
