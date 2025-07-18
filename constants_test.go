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
	"strings"
	"testing"
)

// TestConstantsValues verifies that all constants have the expected values
// that match the Python A2A implementation for protocol compatibility.
func TestConstantsValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		constant string
		expected string
	}{
		"AgentCardWellKnownPath": {
			constant: AgentCardWellKnownPath,
			expected: "/.well-known/agent.json",
		},
		"ExtendedAgentCardPath": {
			constant: ExtendedAgentCardPath,
			expected: "/agent/authenticatedExtendedCard",
		},
		"DefaultRPCURL": {
			constant: DefaultRPCURL,
			expected: "/",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tt.constant != tt.expected {
				t.Errorf("constant %s = %q, want %q", name, tt.constant, tt.expected)
			}
		})
	}
}

// TestConstantsFormat verifies that the constants follow expected formatting rules.
func TestConstantsFormat(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		constant string
		checks   []struct {
			name string
			fn   func(string) bool
		}
	}{
		"AgentCardWellKnownPath": {
			constant: AgentCardWellKnownPath,
			checks: []struct {
				name string
				fn   func(string) bool
			}{
				{"starts with /", func(s string) bool { return strings.HasPrefix(s, "/") }},
				{"ends with .json", func(s string) bool { return strings.HasSuffix(s, ".json") }},
				{"contains .well-known", func(s string) bool { return strings.Contains(s, ".well-known") }},
				{"not empty", func(s string) bool { return s != "" }},
			},
		},
		"ExtendedAgentCardPath": {
			constant: ExtendedAgentCardPath,
			checks: []struct {
				name string
				fn   func(string) bool
			}{
				{"starts with /", func(s string) bool { return strings.HasPrefix(s, "/") }},
				{"contains agent", func(s string) bool { return strings.Contains(s, "agent") }},
				{"not empty", func(s string) bool { return s != "" }},
			},
		},
		"DefaultRPCURL": {
			constant: DefaultRPCURL,
			checks: []struct {
				name string
				fn   func(string) bool
			}{
				{"starts with /", func(s string) bool { return strings.HasPrefix(s, "/") }},
				{"not empty", func(s string) bool { return s != "" }},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, check := range tt.checks {
				t.Run(check.name, func(t *testing.T) {
					if !check.fn(tt.constant) {
						t.Errorf("constant %s failed check %s: %q", name, check.name, tt.constant)
					}
				})
			}
		})
	}
}

// TestConstantsAreStrings verifies that all constants are string types.
func TestConstantsAreStrings(t *testing.T) {
	t.Parallel()

	// This test ensures the constants are properly typed as strings
	var _ string = AgentCardWellKnownPath
	var _ string = ExtendedAgentCardPath
	var _ string = DefaultRPCURL
}

// TestConstantsImmutability verifies that constants maintain their values
// throughout the test execution (they should be immutable).
func TestConstantsImmutability(t *testing.T) {
	t.Parallel()

	// Store initial values
	initialAgentCard := AgentCardWellKnownPath
	initialExtendedCard := ExtendedAgentCardPath
	initialRPCURL := DefaultRPCURL

	// Verify values haven't changed
	if AgentCardWellKnownPath != initialAgentCard {
		t.Errorf("AgentCardWellKnownPath changed from %q to %q", initialAgentCard, AgentCardWellKnownPath)
	}
	if ExtendedAgentCardPath != initialExtendedCard {
		t.Errorf("ExtendedAgentCardPath changed from %q to %q", initialExtendedCard, ExtendedAgentCardPath)
	}
	if DefaultRPCURL != initialRPCURL {
		t.Errorf("DefaultRPCURL changed from %q to %q", initialRPCURL, DefaultRPCURL)
	}
}

// TestConstantsCompatibility verifies that the constants maintain compatibility
// with the Python A2A implementation.
func TestConstantsCompatibility(t *testing.T) {
	t.Parallel()

	// These are the exact values from the Python implementation
	pythonConstants := map[string]string{
		"AGENT_CARD_WELL_KNOWN_PATH": "/.well-known/agent.json",
		"EXTENDED_AGENT_CARD_PATH":   "/agent/authenticatedExtendedCard",
		"DEFAULT_RPC_URL":            "/",
	}

	goConstants := map[string]string{
		"AGENT_CARD_WELL_KNOWN_PATH": AgentCardWellKnownPath,
		"EXTENDED_AGENT_CARD_PATH":   ExtendedAgentCardPath,
		"DEFAULT_RPC_URL":            DefaultRPCURL,
	}

	for name, pythonValue := range pythonConstants {
		goValue, exists := goConstants[name]
		if !exists {
			t.Errorf("Go constant for %s not found", name)
			continue
		}

		if goValue != pythonValue {
			t.Errorf("Go constant for %s = %q, Python value = %q", name, goValue, pythonValue)
		}
	}
}

// TestConstantsUsageScenarios tests the constants in typical usage scenarios.
func TestConstantsUsageScenarios(t *testing.T) {
	t.Parallel()

	// Test URL construction scenarios
	baseURL := "https://agent.example.com"

	agentCardURL := baseURL + AgentCardWellKnownPath
	expectedAgentCardURL := "https://agent.example.com/.well-known/agent.json"
	if agentCardURL != expectedAgentCardURL {
		t.Errorf("Agent card URL = %q, want %q", agentCardURL, expectedAgentCardURL)
	}

	extendedCardURL := baseURL + ExtendedAgentCardPath
	expectedExtendedCardURL := "https://agent.example.com/agent/authenticatedExtendedCard"
	if extendedCardURL != expectedExtendedCardURL {
		t.Errorf("Extended card URL = %q, want %q", extendedCardURL, expectedExtendedCardURL)
	}

	rpcURL := baseURL + DefaultRPCURL
	expectedRPCURL := "https://agent.example.com/"
	if rpcURL != expectedRPCURL {
		t.Errorf("RPC URL = %q, want %q", rpcURL, expectedRPCURL)
	}
}

// BenchmarkConstantsAccess benchmarks the performance of accessing constants.
func BenchmarkConstantsAccess(b *testing.B) {
	benchmarks := map[string]struct {
		constant string
	}{
		"AgentCardWellKnownPath": {AgentCardWellKnownPath},
		"ExtendedAgentCardPath":  {ExtendedAgentCardPath},
		"DefaultRPCURL":          {DefaultRPCURL},
	}

	for name, bm := range benchmarks {
		b.Run(name, func(b *testing.B) {
			for b.Loop() {
				_ = bm.constant
			}
		})
	}
}

// BenchmarkConstantsStringOps benchmarks string operations with constants.
func BenchmarkConstantsStringOps(b *testing.B) {
	baseURL := "https://agent.example.com"

	b.Run("AgentCardURLConstruction", func(b *testing.B) {
		for b.Loop() {
			_ = baseURL + AgentCardWellKnownPath
		}
	})

	b.Run("ExtendedCardURLConstruction", func(b *testing.B) {
		for b.Loop() {
			_ = baseURL + ExtendedAgentCardPath
		}
	})

	b.Run("RPCURLConstruction", func(b *testing.B) {
		for b.Loop() {
			_ = baseURL + DefaultRPCURL
		}
	})
}
