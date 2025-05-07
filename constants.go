// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a

// TaskState represents the current state of a task
type TaskState string

const (
	// TaskStateSubmitted indicates the task has been received but not yet started
	TaskStateSubmitted TaskState = "submitted"
	// TaskStateWorking indicates the task is actively being processed
	TaskStateWorking TaskState = "working"
	// TaskStateInputRequired indicates the agent requires further input from the user/client
	TaskStateInputRequired TaskState = "input-required"
	// TaskStateCompleted indicates the task has finished successfully
	TaskStateCompleted TaskState = "completed"
	// TaskStateCanceled indicates the task was canceled
	TaskStateCanceled TaskState = "canceled"
	// TaskStateFailed indicates the task failed due to an error
	TaskStateFailed TaskState = "failed"
	// TaskStateUnknown indicates the state cannot be determined
	TaskStateUnknown TaskState = "unknown"
)

// PartType represents the type of a part in a message
type PartType string

const (
	// PartTypeText represents text content
	PartTypeText PartType = "text"
	// PartTypeFile represents file content
	PartTypeFile PartType = "file"
	// PartTypeData represents structured data content
	PartTypeData PartType = "data"
)

// MessageRole represents the role of a message sender
type MessageRole string

const (
	// MessageRoleUser represents a message from the user
	MessageRoleUser MessageRole = "user"
	// MessageRoleAgent represents a message from the agent
	MessageRoleAgent MessageRole = "agent"
)

// JSONRPC methods
const (
	// MethodTasksSend is the method for sending a task
	MethodTasksSend = "tasks/send"
	// MethodTasksSendSubscribe is the method for sending a task and subscribing to updates
	MethodTasksSendSubscribe = "tasks/sendSubscribe"
	// MethodTasksGet is the method for getting a task
	MethodTasksGet = "tasks/get"
	// MethodTasksCancel is the method for canceling a task
	MethodTasksCancel = "tasks/cancel"
	// MethodTasksPushNotificationSet is the method for setting a push notification
	MethodTasksPushNotificationSet = "tasks/pushNotification/set"
	// MethodTasksPushNotificationGet is the method for getting a push notification
	MethodTasksPushNotificationGet = "tasks/pushNotification/get"
	// MethodTasksResubscribe is the method for resubscribing to task updates
	MethodTasksResubscribe = "tasks/resubscribe"
)
