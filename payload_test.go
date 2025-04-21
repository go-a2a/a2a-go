// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a_test

import (
	"testing"

	"github.com/bytedance/sonic"
	gocmp "github.com/google/go-cmp/cmp"
	gocmpopts "github.com/google/go-cmp/cmp/cmpopts"

	"github.com/go-a2a/a2a"
)

func TestSendTaskRequest(t *testing.T) {
	t.Parallel()

	params := a2a.TaskSendParams{
		TaskIDParams: a2a.TaskIDParams{
			ID: "test-id",
		},
		Message: a2a.Message{
			Role: a2a.RoleUser,
			Parts: []a2a.Part{
				&a2a.TextPart{Text: "hello"},
			},
		},
	}
	req := a2a.NewSendTaskRequest(a2a.NewID("req-id"), params)

	// Test method name
	if methodName := req.Method; methodName != a2a.MethodTasksSend {
		t.Errorf("req.Method = %v, want %v", methodName, a2a.MethodTasksSend)
	}

	// Test request fields
	if req.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %v, want %v", req.JSONRPC, "2.0")
	}
	if req.Method != a2a.MethodTasksSend {
		t.Errorf("Method = %v, want %v", req.Method, a2a.MethodTasksSend)
	}
	if req.ID.String() != "req-id" {
		t.Errorf("ID = %v, want %v", req.ID.String(), "req-id")
	}

	// Test marshaling/unmarshaling
	data, err := sonic.ConfigFastest.Marshal(&req)
	if err != nil {
		t.Fatalf("sonic.ConfigFastest.Marshal() error = %v", err)
	}

	got := new(a2a.SendTaskRequest)
	if err := sonic.ConfigFastest.Unmarshal(data, got); err != nil {
		t.Fatalf("sonic.ConfigFastest.Unmarshal() error = %v", err)
	}

	// Using EquateEmpty to handle empty maps/slices
	opts := []gocmp.Option{
		gocmpopts.IgnoreFields(a2a.SendTaskRequest{}, "ID"),
		gocmpopts.EquateEmpty(),
	}
	if diff := gocmp.Diff(req, got, opts...); diff != "" {
		t.Errorf("SendTaskRequest mismatch (-want +got):\n%s", diff)
	}
}

func TestSendTaskStreamingRequest(t *testing.T) {
	t.Parallel()

	params := a2a.TaskSendParams{
		TaskIDParams: a2a.TaskIDParams{
			ID: "test-id",
		},
		Message: a2a.Message{
			Role: "user",
			Parts: []a2a.Part{
				&a2a.TextPart{Text: "hello"},
			},
		},
		AcceptedOutputModes: []string{"streaming"},
	}

	req := a2a.NewSendTaskStreamingRequest(a2a.NewID("req-id"), params)

	// Test method name
	if methodName := req.Method; methodName != a2a.MethodTasksSendSubscribe {
		t.Errorf("req.Method = %v, want %v", methodName, a2a.MethodTasksSendSubscribe)
	}

	// Test request fields
	if req.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %v, want %v", req.JSONRPC, "2.0")
	}
	if req.Method != a2a.MethodTasksSendSubscribe {
		t.Errorf("Method = %v, want %v", req.Method, a2a.MethodTasksSendSubscribe)
	}
	if req.ID.String() != "req-id" {
		t.Errorf("ID = %v, want %v", req.ID.String(), "req-id")
	}

	// Test marshaling/unmarshaling
	data, err := sonic.ConfigFastest.Marshal(&req)
	if err != nil {
		t.Fatalf("sonic.ConfigFastest.Marshal() error = %v", err)
	}

	got := new(a2a.SendTaskStreamingRequest)
	if err := sonic.ConfigFastest.Unmarshal(data, got); err != nil {
		t.Fatalf("sonic.ConfigFastest.Unmarshal() error = %v", err)
	}

	// Using EquateEmpty to handle empty maps/slices
	opts := []gocmp.Option{
		gocmpopts.IgnoreFields(a2a.SendTaskStreamingRequest{}, "ID"),
		gocmpopts.EquateEmpty(),
	}
	if diff := gocmp.Diff(req, got, opts...); diff != "" {
		t.Errorf("SendTaskStreamingRequest mismatch (-want +got):\n%s", diff)
	}
}

func TestGetTaskRequest(t *testing.T) {
	t.Parallel()

	params := a2a.TaskQueryParams{
		TaskIDParams: a2a.TaskIDParams{
			ID: "test-id",
		},
		HistoryLength: 10,
	}

	req := a2a.NewGetTaskRequest(a2a.NewID("req-id"), params)

	// Test method name
	if methodName := req.Method; methodName != a2a.MethodTasksGet {
		t.Errorf("req.Method = %v, want %v", methodName, a2a.MethodTasksGet)
	}

	// Test request fields
	if req.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %v, want %v", req.JSONRPC, "2.0")
	}
	if req.Method != a2a.MethodTasksGet {
		t.Errorf("Method = %v, want %v", req.Method, a2a.MethodTasksGet)
	}
	if req.ID.String() != "req-id" {
		t.Errorf("ID = %v, want %v", req.ID.String(), "req-id")
	}

	// Test marshaling/unmarshaling
	data, err := sonic.ConfigFastest.Marshal(req)
	if err != nil {
		t.Fatalf("sonic.ConfigFastest.Marshal() error = %v", err)
	}

	got := new(a2a.GetTaskRequest)
	if err := sonic.ConfigFastest.Unmarshal(data, got); err != nil {
		t.Fatalf("sonic.ConfigFastest.Unmarshal() error = %v", err)
	}

	// Using EquateEmpty to handle empty maps/slices
	opts := []gocmp.Option{
		gocmpopts.IgnoreFields(a2a.GetTaskRequest{}, "ID"),
		gocmpopts.EquateEmpty(),
	}
	if diff := gocmp.Diff(req, got, opts...); diff != "" {
		t.Errorf("GetTaskRequest mismatch (-want +got):\n%s", diff)
	}
}

func TestCancelTaskRequest(t *testing.T) {
	t.Parallel()

	params := a2a.TaskIDParams{
		ID: "test-id",
	}

	req := a2a.NewCancelTaskRequest(a2a.NewID("req-id"), params)

	// Test method name
	if methodName := req.Method; methodName != a2a.MethodTasksCancel {
		t.Errorf("req.Method = %v, want %v", methodName, a2a.MethodTasksCancel)
	}

	// Test request fields
	if req.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %v, want %v", req.JSONRPC, "2.0")
	}
	if req.Method != a2a.MethodTasksCancel {
		t.Errorf("Method = %v, want %v", req.Method, a2a.MethodTasksCancel)
	}
	if req.ID.String() != "req-id" {
		t.Errorf("ID = %v, want %v", req.ID.String(), "req-id")
	}

	// Test marshaling/unmarshaling
	data, err := sonic.ConfigFastest.Marshal(req)
	if err != nil {
		t.Fatalf("sonic.ConfigFastest.Marshal() error = %v", err)
	}

	got := new(a2a.CancelTaskRequest)
	if err := sonic.ConfigFastest.Unmarshal(data, got); err != nil {
		t.Fatalf("sonic.ConfigFastest.Unmarshal() error = %v", err)
	}

	// Using EquateEmpty to handle empty maps/slices
	opts := []gocmp.Option{
		gocmpopts.IgnoreFields(a2a.CancelTaskRequest{}, "ID"),
		gocmpopts.EquateEmpty(),
	}
	if diff := gocmp.Diff(req, got, opts...); diff != "" {
		t.Errorf("CancelTaskRequest mismatch (-want +got):\n%s", diff)
	}
}

func TestSetTaskPushNotificationRequest(t *testing.T) {
	t.Parallel()

	params := a2a.TaskPushNotificationConfig{
		ID: "test-id",
		PushNotificationConfig: a2a.PushNotificationConfig{
			URL:   "https://example.com/push",
			Token: "token123",
		},
	}

	req := a2a.NewSetTaskPushNotificationRequest(a2a.NewID("req-id"), params)

	// Test method name
	if methodName := req.Method; methodName != a2a.MethodTasksPushNotificationSet {
		t.Errorf("req.Method = %v, want %v", methodName, a2a.MethodTasksPushNotificationSet)
	}

	// Test request fields
	if req.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %v, want %v", req.JSONRPC, "2.0")
	}
	if req.Method != a2a.MethodTasksPushNotificationSet {
		t.Errorf("Method = %v, want %v", req.Method, a2a.MethodTasksPushNotificationSet)
	}
	if req.ID.String() != "req-id" {
		t.Errorf("ID = %v, want %v", req.ID.String(), "req-id")
	}

	// Test marshaling/unmarshaling
	data, err := sonic.ConfigFastest.Marshal(req)
	if err != nil {
		t.Fatalf("sonic.ConfigFastest.Marshal() error = %v", err)
	}

	got := new(a2a.SetTaskPushNotificationRequest)
	if err := sonic.ConfigFastest.Unmarshal(data, got); err != nil {
		t.Fatalf("sonic.ConfigFastest.Unmarshal() error = %v", err)
	}

	// Using EquateEmpty to handle empty maps/slices
	opts := []gocmp.Option{
		gocmpopts.IgnoreFields(a2a.SetTaskPushNotificationRequest{}, "ID"),
		gocmpopts.EquateEmpty(),
	}
	if diff := gocmp.Diff(req, got, opts...); diff != "" {
		t.Errorf("SetTaskPushNotificationRequest mismatch (-want +got):\n%s", diff)
	}
}

func TestGetTaskPushNotificationRequest(t *testing.T) {
	t.Parallel()

	params := a2a.TaskIDParams{
		ID: "test-id",
	}

	req := a2a.NewGetTaskPushNotificationRequest(a2a.NewID("req-id"), params)

	// Test method name
	if methodName := req.Method; methodName != a2a.MethodTasksPushNotificationGet {
		t.Errorf("req.Method = %v, want %v", methodName, a2a.MethodTasksPushNotificationGet)
	}

	// Test request fields
	if req.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %v, want %v", req.JSONRPC, "2.0")
	}
	if req.Method != a2a.MethodTasksPushNotificationGet {
		t.Errorf("Method = %v, want %v", req.Method, a2a.MethodTasksPushNotificationGet)
	}
	if req.ID.String() != "req-id" {
		t.Errorf("ID = %v, want %v", req.ID.String(), "req-id")
	}

	// Test marshaling/unmarshaling
	data, err := sonic.ConfigFastest.Marshal(req)
	if err != nil {
		t.Fatalf("sonic.ConfigFastest.Marshal() error = %v", err)
	}

	got := new(a2a.GetTaskPushNotificationRequest)
	if err := sonic.ConfigFastest.Unmarshal(data, got); err != nil {
		t.Fatalf("sonic.ConfigFastest.Unmarshal() error = %v", err)
	}

	// Using EquateEmpty to handle empty maps/slices
	opts := []gocmp.Option{
		gocmpopts.IgnoreFields(a2a.GetTaskPushNotificationRequest{}, "ID"),
		gocmpopts.EquateEmpty(),
	}
	if diff := gocmp.Diff(req, got, opts...); diff != "" {
		t.Errorf("GetTaskPushNotificationRequest mismatch (-want +got):\n%s", diff)
	}
}

func TestTaskResubscriptionRequest(t *testing.T) {
	t.Parallel()

	params := a2a.TaskIDParams{
		ID: "test-id",
	}

	req := a2a.NewTaskResubscriptionRequest(a2a.NewID("req-id"), params)

	// Test method name
	if methodName := req.Method; methodName != a2a.MethodTasksResubscribe {
		t.Errorf("req.Method = %v, want %v", methodName, a2a.MethodTasksResubscribe)
	}

	// Test request fields
	if req.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %v, want %v", req.JSONRPC, "2.0")
	}
	if req.Method != a2a.MethodTasksResubscribe {
		t.Errorf("Method = %v, want %v", req.Method, a2a.MethodTasksResubscribe)
	}
	if req.ID.String() != "req-id" {
		t.Errorf("ID = %v, want %v", req.ID.String(), "req-id")
	}

	// Test marshaling/unmarshaling
	data, err := sonic.ConfigFastest.Marshal(req)
	if err != nil {
		t.Fatalf("sonic.ConfigFastest.Marshal() error = %v", err)
	}

	got := new(a2a.TaskResubscriptionRequest)
	if err := sonic.ConfigFastest.Unmarshal(data, got); err != nil {
		t.Fatalf("sonic.ConfigFastest.Unmarshal() error = %v", err)
	}

	// Using EquateEmpty to handle empty maps/slices
	opts := []gocmp.Option{
		gocmpopts.IgnoreFields(a2a.TaskResubscriptionRequest{}, "ID"),
		gocmpopts.EquateEmpty(),
	}
	if diff := gocmp.Diff(req, got, opts...); diff != "" {
		t.Errorf("TaskResubscriptionRequest mismatch (-want +got):\n%s", diff)
	}
}
