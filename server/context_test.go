// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

// mockUser is a simple mock implementation of auth.User for testing
type mockUser struct {
	authenticated bool
	username      string
}

func (m mockUser) IsAuthenticated() bool {
	return m.authenticated
}

func (m mockUser) UserName() string {
	return m.username
}

func TestNewServerCallContext(t *testing.T) {
	user := mockUser{authenticated: true, username: "testuser"}
	scc := NewServerCallContext(user)

	if scc.User() != user {
		t.Errorf("Expected user %v, got %v", user, scc.User())
	}

	state := scc.State()
	if len(state) != 0 {
		t.Errorf("Expected empty state, got %v", state)
	}
}

func TestNewServerCallContextWithNilUser(t *testing.T) {
	scc := NewServerCallContext(nil)

	if scc.User().IsAuthenticated() {
		t.Error("Expected unauthenticated user when nil user provided")
	}

	if scc.User().UserName() != "" {
		t.Error("Expected empty username for unauthenticated user")
	}
}

func TestNewServerCallContextWithState(t *testing.T) {
	user := mockUser{authenticated: true, username: "testuser"}
	initialState := map[string]any{
		"key1": "value1",
		"key2": 42,
	}

	scc := NewServerCallContextWithState(user, initialState)

	if scc.User() != user {
		t.Errorf("Expected user %v, got %v", user, scc.User())
	}

	state := scc.State()
	if len(state) != 2 {
		t.Errorf("Expected state with 2 keys, got %d", len(state))
	}

	if state["key1"] != "value1" {
		t.Errorf("Expected key1 to be 'value1', got %v", state["key1"])
	}

	if state["key2"] != 42 {
		t.Errorf("Expected key2 to be 42, got %v", state["key2"])
	}

	// Verify that modifying the returned state doesn't affect the internal state
	state["key3"] = "new_value"
	newState := scc.State()
	if len(newState) != 2 {
		t.Error("Internal state was modified by external changes")
	}
}

func TestSetAndGetState(t *testing.T) {
	user := mockUser{authenticated: true, username: "testuser"}
	scc := NewServerCallContext(user)

	// Test setting and getting state
	scc.SetState("test_key", "test_value")
	value, ok := scc.GetState("test_key")
	if !ok {
		t.Error("Expected to find test_key in state")
	}
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %v", value)
	}

	// Test getting non-existent key
	_, ok = scc.GetState("non_existent_key")
	if ok {
		t.Error("Expected not to find non_existent_key in state")
	}
}

func TestDeleteState(t *testing.T) {
	user := mockUser{authenticated: true, username: "testuser"}
	scc := NewServerCallContext(user)

	// Set a value
	scc.SetState("to_delete", "value")
	_, ok := scc.GetState("to_delete")
	if !ok {
		t.Error("Expected to find key before deletion")
	}

	// Delete the value
	scc.DeleteState("to_delete")
	_, ok = scc.GetState("to_delete")
	if ok {
		t.Error("Expected not to find key after deletion")
	}
}

func TestThreadSafety(t *testing.T) {
	user := mockUser{authenticated: true, username: "testuser"}
	scc := NewServerCallContext(user)

	// Test concurrent access
	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)

				scc.SetState(key, value)
				retrievedValue, ok := scc.GetState(key)
				if !ok {
					t.Errorf("Expected to find key %s", key)
				}
				if retrievedValue != value {
					t.Errorf("Expected %s, got %v", value, retrievedValue)
				}

				// Also test State() method
				state := scc.State()
				if len(state) == 0 {
					t.Error("Expected non-empty state")
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestValidate(t *testing.T) {
	user := mockUser{authenticated: true, username: "testuser"}
	scc := NewServerCallContext(user)

	// Valid context should pass validation
	if err := scc.Validate(); err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}

	// Test with nil user (should not happen with constructor, but test anyway)
	scc.user = nil
	if err := scc.Validate(); err == nil {
		t.Error("Expected validation to fail with nil user")
	}
}

func TestString(t *testing.T) {
	user := mockUser{authenticated: true, username: "testuser"}
	scc := NewServerCallContext(user)
	scc.SetState("key1", "value1")
	scc.SetState("key2", "value2")

	str := scc.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}

	// Check that it contains expected information
	if !contains(str, "testuser") {
		t.Error("Expected string to contain username")
	}
	if !contains(str, "true") {
		t.Error("Expected string to contain authentication status")
	}
	if !contains(str, "2") {
		t.Error("Expected string to contain state keys count")
	}
}

func TestDefaultCallContextBuilder(t *testing.T) {
	builder := NewDefaultCallContextBuilder()
	ctx := context.Background()

	// Test successful build
	scc, err := builder.Build(ctx, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if scc.User().IsAuthenticated() {
		t.Error("Expected unauthenticated user from default builder")
	}

	// Check that Go context is stored in state
	goCtx, ok := scc.GetState("go_context")
	if !ok {
		t.Error("Expected go_context in state")
	}
	if goCtx != ctx {
		t.Error("Expected stored context to match input context")
	}

	// Test with nil context
	_, err = builder.Build(nil, nil)
	if err == nil {
		t.Error("Expected error when building with nil context")
	}
}

func TestCallContextBuilderInterface(t *testing.T) {
	// Verify that DefaultCallContextBuilder implements CallContextBuilder
	var builder CallContextBuilder = NewDefaultCallContextBuilder()

	ctx := context.Background()
	scc, err := builder.Build(ctx, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if scc == nil {
		t.Error("Expected non-nil ServerCallContext")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
