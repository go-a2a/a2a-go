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

package client

import (
	"fmt"
	"net/http"
)

// ClientError represents a client-side error.
type ClientError interface {
	error
	Code() int
	Message() string
	IsRetryable() bool
}

// HTTPError represents an HTTP-related error.
type HTTPError struct {
	StatusCode int
	Msg        string
	Err        error
}

var _ ClientError = (*HTTPError)(nil)

// NewHTTPError creates a new HTTP error.
func NewHTTPError(statusCode int, message string, err error) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Msg:        message,
		Err:        err,
	}
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("HTTP error %d: %s: %v", e.StatusCode, e.Msg, e.Err)
	}
	return fmt.Sprintf("HTTP error %d: %s", e.StatusCode, e.Msg)
}

// Code returns the HTTP status code.
func (e *HTTPError) Code() int {
	return e.StatusCode
}

// Message returns the error message.
func (e *HTTPError) Message() string {
	return e.Msg
}

// IsRetryable returns true if the error is retryable.
func (e *HTTPError) IsRetryable() bool {
	return e.StatusCode >= 500 || e.StatusCode == http.StatusRequestTimeout || e.StatusCode == http.StatusTooManyRequests
}

// Unwrap returns the underlying error.
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// JSONError represents a JSON parsing or validation error.
type JSONError struct {
	Msg string
	Err error
}

var _ ClientError = (*JSONError)(nil)

// NewJSONError creates a new JSON error.
func NewJSONError(message string, err error) *JSONError {
	return &JSONError{
		Msg: message,
		Err: err,
	}
}

// Error implements the error interface.
func (e *JSONError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("JSON error: %s: %v", e.Msg, e.Err)
	}
	return fmt.Sprintf("JSON error: %s", e.Msg)
}

// Code returns a fixed error code for JSON errors.
func (e *JSONError) Code() int {
	return -32700 // JSON parse error code from JSON-RPC
}

// Message returns the error message.
func (e *JSONError) Message() string {
	return e.Msg
}

// IsRetryable returns false for JSON errors.
func (e *JSONError) IsRetryable() bool {
	return false
}

// Unwrap returns the underlying error.
func (e *JSONError) Unwrap() error {
	return e.Err
}

// TimeoutError represents a timeout errors during a request.
type TimeoutError struct {
	Msg string
	Err error
}

var _ ClientError = (*TimeoutError)(nil)

// NewNetworkError creates a new network error.
func NewTimeoutError(message string, err error) *TimeoutError {
	return &TimeoutError{
		Msg: message,
		Err: err,
	}
}

// Error implements the error interface.
func (e *TimeoutError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Network error: %s: %v", e.Msg, e.Err)
	}
	return fmt.Sprintf("Network error: %s", e.Msg)
}

// Code returns a fixed error code for network errors.
func (e *TimeoutError) Code() int {
	return 503 // Service unavailable
}

// Message returns the error message.
func (e *TimeoutError) Message() string {
	return e.Msg
}

// IsRetryable returns true for network errors.
func (e *TimeoutError) IsRetryable() bool {
	return true
}

// Unwrap returns the underlying error.
func (e *TimeoutError) Unwrap() error {
	return e.Err
}

// NetworkError represents a network-related error.
type NetworkError struct {
	Msg string
	Err error
}

var _ ClientError = (*NetworkError)(nil)

// NewNetworkError creates a new network error.
func NewNetworkError(message string, err error) *NetworkError {
	return &NetworkError{
		Msg: message,
		Err: err,
	}
}

// Error implements the error interface.
func (e *NetworkError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Network error: %s: %v", e.Msg, e.Err)
	}
	return fmt.Sprintf("Network error: %s", e.Msg)
}

// Code returns a fixed error code for network errors.
func (e *NetworkError) Code() int {
	return 503 // Service unavailable
}

// Message returns the error message.
func (e *NetworkError) Message() string {
	return e.Msg
}

// IsRetryable returns true for network errors.
func (e *NetworkError) IsRetryable() bool {
	return true
}

// Unwrap returns the underlying error.
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// ConfigurationError represents a configuration error.
type ConfigurationError struct {
	Msg string
	Err error
}

var _ ClientError = (*ConfigurationError)(nil)

// NewConfigurationError creates a new configuration error.
func NewConfigurationError(message string, err error) *ConfigurationError {
	return &ConfigurationError{
		Msg: message,
		Err: err,
	}
}

// Error implements the error interface.
func (e *ConfigurationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Configuration error: %s: %v", e.Msg, e.Err)
	}
	return fmt.Sprintf("Configuration error: %s", e.Msg)
}

// Code returns a fixed error code for configuration errors.
func (e *ConfigurationError) Code() int {
	return -32600 // Invalid request error code from JSON-RPC
}

// Message returns the error message.
func (e *ConfigurationError) Message() string {
	return e.Msg
}

// IsRetryable returns false for configuration errors.
func (e *ConfigurationError) IsRetryable() bool {
	return false
}

// Unwrap returns the underlying error.
func (e *ConfigurationError) Unwrap() error {
	return e.Err
}
