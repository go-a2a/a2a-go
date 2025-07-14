// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestUserInterface verifies that UnauthenticatedUser implements the User interface.
func TestUserInterface(t *testing.T) {
	var _ User = UnauthenticatedUser{}
}

func TestUnauthenticatedUser_IsAuthenticated(t *testing.T) {
	tests := map[string]struct {
		user User
		want bool
	}{
		"zero value": {
			user: UnauthenticatedUser{},
			want: false,
		},
		"via interface": {
			user: User(UnauthenticatedUser{}),
			want: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.user.IsAuthenticated()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("IsAuthenticated() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUnauthenticatedUser_UserName(t *testing.T) {
	tests := map[string]struct {
		user User
		want string
	}{
		"zero value": {
			user: UnauthenticatedUser{},
			want: "",
		},
		"via interface": {
			user: User(UnauthenticatedUser{}),
			want: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.user.UserName()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("UserName() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUnauthenticatedUser_ZeroValue(t *testing.T) {
	// Test that zero value is usable
	var user UnauthenticatedUser

	if got := user.IsAuthenticated(); got != false {
		t.Errorf("Zero value IsAuthenticated() = %v, want false", got)
	}

	if got := user.UserName(); got != "" {
		t.Errorf("Zero value UserName() = %q, want empty string", got)
	}
}

func TestUnauthenticatedUser_Immutability(t *testing.T) {
	user1 := UnauthenticatedUser{}
	user2 := UnauthenticatedUser{}

	// Test that multiple instances behave identically
	if user1.IsAuthenticated() != user2.IsAuthenticated() {
		t.Error("Multiple instances should have identical IsAuthenticated() behavior")
	}

	if user1.UserName() != user2.UserName() {
		t.Error("Multiple instances should have identical UserName() behavior")
	}
}

func TestUnauthenticatedUser_ThreadSafety(t *testing.T) {
	// Test concurrent access to the same instance
	user := UnauthenticatedUser{}

	// Run multiple goroutines accessing the same instance
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func() {
			_ = user.IsAuthenticated()
			_ = user.UserName()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify the instance still works correctly
	if got := user.IsAuthenticated(); got != false {
		t.Errorf("After concurrent access, IsAuthenticated() = %v, want false", got)
	}

	if got := user.UserName(); got != "" {
		t.Errorf("After concurrent access, UserName() = %q, want empty string", got)
	}
}

// BenchmarkUnauthenticatedUser_IsAuthenticated benchmarks the IsAuthenticated method.
func BenchmarkUnauthenticatedUser_IsAuthenticated(b *testing.B) {
	user := UnauthenticatedUser{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.IsAuthenticated()
	}
}

// BenchmarkUnauthenticatedUser_UserName benchmarks the UserName method.
func BenchmarkUnauthenticatedUser_UserName(b *testing.B) {
	user := UnauthenticatedUser{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.UserName()
	}
}

// BenchmarkUnauthenticatedUser_InterfaceCall benchmarks method calls through the interface.
func BenchmarkUnauthenticatedUser_InterfaceCall(b *testing.B) {
	var user User = UnauthenticatedUser{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.IsAuthenticated()
		_ = user.UserName()
	}
}

// BenchmarkUnauthenticatedUser_Creation benchmarks struct creation.
func BenchmarkUnauthenticatedUser_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UnauthenticatedUser{}
	}
}

// ExampleUnauthenticatedUser demonstrates basic usage of UnauthenticatedUser.
func ExampleUnauthenticatedUser() {
	user := UnauthenticatedUser{}

	// Check authentication status
	if !user.IsAuthenticated() {
		// Handle unauthenticated user
		_ = user.UserName() // Returns empty string
	}

	// Output:
}

// ExampleUser demonstrates using UnauthenticatedUser through the User interface.
func ExampleUser() {
	var user User = UnauthenticatedUser{}

	// Interface usage
	authenticated := user.IsAuthenticated()
	username := user.UserName()

	_, _ = authenticated, username

	// Output:
}
