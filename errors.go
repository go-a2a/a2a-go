// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"github.com/go-a2a/a2a/internal/jsonrpc2"
)

// A2A specific errors.
var (
	// ErrorTaskNotFound is the error code for task not found.
	ErrorTaskNotFound = jsonrpc2.NewError(-32001, "task not found")

	// ErrorTaskNotCancelable is the error code for task not cancelable.
	ErrorTaskNotCancelable = jsonrpc2.NewError(-32002, "Task cannot be canceled")

	// ErrorPushNotificationNotSupported is the error code for push notification not supported.
	ErrorPushNotificationNotSupported = jsonrpc2.NewError(-32003, "Push Notification is not supported")

	// ErrorUnsupportedOperation is the error code for unsupported operation.
	ErrorUnsupportedOperation = jsonrpc2.NewError(-32004, "This operation is not supported")

	// ErrorContentTypeNotSupported is the error code for content type not supported
	ErrorContentTypeNotSupported = jsonrpc2.NewError(-32005, "Content type not supported")
)
