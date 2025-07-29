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

package pyclient_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/pyclient"
)

func TestClient_New(t *testing.T) {
	tests := map[string]struct {
		opts    []pyclient.Option
		wantErr bool
		errMsg  string
	}{
		"success: with base URL": {
			opts: []pyclient.Option{
				pyclient.WithBaseURL("https://example.com"),
			},
			wantErr: false,
		},
		"error: missing base URL": {
			opts:    []pyclient.Option{},
			wantErr: true,
			errMsg:  "base URL is required",
		},
		"error: empty base URL": {
			opts: []pyclient.Option{
				pyclient.WithBaseURL(""),
			},
			wantErr: true,
			errMsg:  "base URL cannot be empty",
		},
		"success: with multiple options": {
			opts: []pyclient.Option{
				pyclient.WithBaseURL("https://example.com"),
				pyclient.WithTimeout(30 * time.Second),
				pyclient.WithAuthToken("test-token"),
				pyclient.WithDebugLogging(true),
			},
			wantErr: false,
		},
		"error: invalid timeout": {
			opts: []pyclient.Option{
				pyclient.WithBaseURL("https://example.com"),
				pyclient.WithTimeout(0),
			},
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
		"error: nil HTTP client": {
			opts: []pyclient.Option{
				pyclient.WithBaseURL("https://example.com"),
				pyclient.WithHTTPClient(nil),
			},
			wantErr: true,
			errMsg:  "HTTP client cannot be nil",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client, err := pyclient.New(tc.opts...)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("expected error containing %q, got %q", tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if client == nil {
					t.Error("expected client, got nil")
				} else {
					// Clean up
					client.Close()
				}
			}
		})
	}
}

func TestClient_Connect(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		serverFunc func(*testing.T) *httptest.Server
		wantErr    bool
		errMsg     string
	}{
		"success: valid agent card": {
			serverFunc: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == a2a.AgentCardWellKnownPath {
						agentCard := &a2a.AgentCard{
							Name:        "Test Agent",
							Version:     "1.0.0",
							Description: "Test agent for unit tests",
							Capabilities: &a2a.AgentCapabilities{
								Streaming: true,
							},
						}
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(agentCard)
					} else {
						http.NotFound(w, r)
					}
				}))
			},
			wantErr: false,
		},
		"error: agent card not found": {
			serverFunc: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.NotFound(w, r)
				}))
			},
			wantErr: true,
			errMsg:  "404",
		},
		"error: invalid agent card JSON": {
			serverFunc: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == a2a.AgentCardWellKnownPath {
						w.Header().Set("Content-Type", "application/json")
						w.Write([]byte("invalid json"))
					}
				}))
			},
			wantErr: true,
			errMsg:  "decode agent card",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			server := tc.serverFunc(t)
			defer server.Close()

			client, err := pyclient.New(
				pyclient.WithBaseURL(server.URL),
			)
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}
			defer client.Close()

			err = client.Connect(ctx)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("expected error containing %q, got %q", tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestClient_ConnectionState(t *testing.T) {
	ctx := t.Context()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == a2a.AgentCardWellKnownPath {
			agentCard := &a2a.AgentCard{
				Name:    "Test Agent",
				Version: "1.0.0",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(agentCard)
		}
	}))
	defer server.Close()

	stateChanges := make([]string, 0)
	client, err := pyclient.New(
		pyclient.WithBaseURL(server.URL),
		pyclient.WithConnectionStateCallback(func(oldState, newState pyclient.ConnectionState) {
			stateChanges = append(stateChanges, fmt.Sprintf("%s->%s", oldState, newState))
		}),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Check initial state
	if state := client.State(); state != pyclient.ConnectionStateDisconnected {
		t.Errorf("expected initial state to be disconnected, got %s", state)
	}

	// Connect
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	// Check connected state
	if state := client.State(); state != pyclient.ConnectionStateConnected {
		t.Errorf("expected state to be connected, got %s", state)
	}

	// Verify state changes
	expectedChanges := []string{
		"disconnected->connecting",
		"connecting->connected",
	}
	if diff := cmp.Diff(expectedChanges, stateChanges); diff != "" {
		t.Errorf("state changes mismatch (-want +got):\n%s", diff)
	}
}

func TestClient_SendStreamMessage(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		serverFunc func(*testing.T) *httptest.Server
		params     *a2a.MessageSendParams
		wantEvents []string
		wantErr    bool
		errMsg     string
	}{
		"success: stream events": {
			serverFunc: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case a2a.AgentCardWellKnownPath:
						agentCard := &a2a.AgentCard{
							Name:    "Test Agent",
							Version: "1.0.0",
							Capabilities: &a2a.AgentCapabilities{
								Streaming: true,
							},
						}
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(agentCard)

					case "/message/stream":
						w.Header().Set("Content-Type", "text/event-stream")
						w.Header().Set("X-Task-ID", "test-task-123")
						w.WriteHeader(http.StatusOK)

						// Send SSE events
						fmt.Fprintf(w, "event: message\n")
						fmt.Fprintf(w, "data: {\"role\":\"agent\",\"parts\":[{\"kind\":\"text\",\"text\":\"Hello from agent\"}]}\n\n")

						fmt.Fprintf(w, "event: status-update\n")
						fmt.Fprintf(w, "data: {\"state\":\"completed\"}\n\n")

						w.(http.Flusher).Flush()

					default:
						http.NotFound(w, r)
					}
				}))
			},
			params: &a2a.MessageSendParams{
				Message: &a2a.Message{
					Role: a2a.RoleUser,
					Parts: []a2a.Part{
						&a2a.TextPart{Text: "Hello, agent!"},
					},
				},
			},
			wantEvents: []string{"MessageEvent", "StatusUpdateEvent"},
			wantErr:    false,
		},
		"error: streaming not supported": {
			serverFunc: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == a2a.AgentCardWellKnownPath {
						agentCard := &a2a.AgentCard{
							Name:    "Test Agent",
							Version: "1.0.0",
							Capabilities: &a2a.AgentCapabilities{
								Streaming: false, // Streaming not supported
							},
						}
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(agentCard)
					}
				}))
			},
			params: &a2a.MessageSendParams{
				Message: &a2a.Message{
					Role: a2a.RoleUser,
					Parts: []a2a.Part{
						&a2a.TextPart{Text: "Hello"},
					},
				},
			},
			wantErr: true,
			errMsg:  "agent does not support streaming",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			server := tc.serverFunc(t)
			defer server.Close()

			client, err := pyclient.New(
				pyclient.WithBaseURL(server.URL),
			)
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}
			defer client.Close()

			if err := client.Connect(ctx); err != nil {
				t.Fatalf("failed to connect: %v", err)
			}

			stream, err := client.SendStreamMessage(ctx, tc.params)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("expected error containing %q, got %q", tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if stream == nil {
					t.Error("expected stream, got nil")
				} else {
					defer stream.Close()

					// Collect events
					var eventTypes []string
					timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
					defer cancel()

					for {
						select {
						case event, ok := <-stream.Events():
							if !ok {
								goto done
							}
							switch event.(type) {
							case *pyclient.MessageEvent:
								eventTypes = append(eventTypes, "MessageEvent")
							case *pyclient.StatusUpdateEvent:
								eventTypes = append(eventTypes, "StatusUpdateEvent")
							case *pyclient.ErrorEvent:
								eventTypes = append(eventTypes, "ErrorEvent")
							}
						case <-timeoutCtx.Done():
							goto done
						}
					}
				done:

					if diff := cmp.Diff(tc.wantEvents, eventTypes); diff != "" {
						t.Errorf("event types mismatch (-want +got):\n%s", diff)
					}
				}
			}
		})
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	tests := map[string]struct {
		err      error
		wantCode int64
		wantMsg  string
	}{
		"pyclient error": {
			err:      pyclient.NewError(pyclient.ErrCodeTimeout, "operation timed out"),
			wantCode: pyclient.ErrCodeTimeout,
			wantMsg:  "operation timed out",
		},
		"pyclient error with cause": {
			err:      pyclient.NewErrorWithCause(pyclient.ErrCodeConnectionFailed, "connection failed", errors.New("network error")),
			wantCode: pyclient.ErrCodeConnectionFailed,
			wantMsg:  "connection failed",
		},
		"validation error": {
			err:      &pyclient.ValidationError{Field: "timeout", Message: "must be positive"},
			wantCode: 0, // Not a pyclient.Error
			wantMsg:  "validation error for field timeout: must be positive",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var clientErr *pyclient.Error
			if errors.As(tc.err, &clientErr) {
				if clientErr.Code != tc.wantCode {
					t.Errorf("expected code %d, got %d", tc.wantCode, clientErr.Code)
				}
				if !strings.Contains(clientErr.Message, tc.wantMsg) {
					t.Errorf("expected message containing %q, got %q", tc.wantMsg, clientErr.Message)
				}
			} else if tc.wantCode != 0 {
				t.Errorf("expected pyclient.Error, got %T", tc.err)
			}

			// Check error message
			if !strings.Contains(tc.err.Error(), tc.wantMsg) {
				t.Errorf("expected error message containing %q, got %q", tc.wantMsg, tc.err.Error())
			}
		})
	}
}

func TestRetryConfig(t *testing.T) {
	config := pyclient.DefaultRetryConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts to be 3, got %d", config.MaxAttempts)
	}

	if config.InitialDelay != 1*time.Second {
		t.Errorf("expected InitialDelay to be 1s, got %v", config.InitialDelay)
	}

	if config.MaxDelay != 30*time.Second {
		t.Errorf("expected MaxDelay to be 30s, got %v", config.MaxDelay)
	}

	if config.Multiplier != 2.0 {
		t.Errorf("expected Multiplier to be 2.0, got %f", config.Multiplier)
	}

	// Test retryable error detection
	retryableErrors := []error{
		&pyclient.ConnectionError{Operation: "connect", URL: "https://example.com", Err: errors.New("network error")},
		&pyclient.TimeoutError{Operation: "read", Duration: "30s"},
		a2a.NewError(-32000, "server overloaded"),
	}

	for _, err := range retryableErrors {
		if !config.RetryableErrors(err) {
			t.Errorf("expected %v to be retryable", err)
		}
	}

	// Test non-retryable errors
	nonRetryableErrors := []error{
		pyclient.NewError(pyclient.ErrCodeInvalidConfiguration, "invalid config"),
		&pyclient.ValidationError{Field: "url", Message: "invalid URL"},
	}

	for _, err := range nonRetryableErrors {
		if config.RetryableErrors(err) {
			t.Errorf("expected %v to NOT be retryable", err)
		}
	}
}
