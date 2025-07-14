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

	for range 100 {
		go func() {
			_ = user.IsAuthenticated()
			_ = user.UserName()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for range 100 {
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

	for b.Loop() {
		_ = user.IsAuthenticated()
	}
}

// BenchmarkUnauthenticatedUser_UserName benchmarks the UserName method.
func BenchmarkUnauthenticatedUser_UserName(b *testing.B) {
	user := UnauthenticatedUser{}

	for b.Loop() {
		_ = user.UserName()
	}
}

// BenchmarkUnauthenticatedUser_InterfaceCall benchmarks method calls through the interface.
func BenchmarkUnauthenticatedUser_InterfaceCall(b *testing.B) {
	var user User = UnauthenticatedUser{}

	for b.Loop() {
		_ = user.IsAuthenticated()
		_ = user.UserName()
	}
}

// BenchmarkUnauthenticatedUser_Creation benchmarks struct creation.
func BenchmarkUnauthenticatedUser_Creation(b *testing.B) {
	for b.Loop() {
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
