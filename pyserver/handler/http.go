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

package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/internal/jsonrpc2"
)

// HTTPHandler handles A2A requests over HTTP transport.
type HTTPHandler struct {
	router Router
}

// NewHTTPHandler creates a new HTTP handler with the given router.
func NewHTTPHandler(router Router) *HTTPHandler {
	return &HTTPHandler{
		router: router,
	}
}

// ServeHTTP implements http.Handler interface.
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract method from URL path
	method := extractMethod(r.URL.Path)
	if method == "" {
		writeHTTPError(w, http.StatusNotFound, "invalid endpoint")
		return
	}

	// Only accept POST requests for RPC methods
	if r.Method != http.MethodPost {
		writeHTTPError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeHTTPError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	// Create request context
	reqCtx := &RequestContext{
		Method:    a2a.Method(method),
		Headers:   r.Header,
		RequestID: r.Header.Get("X-Request-ID"),
	}

	// Add session if available
	session := extractSession(r)
	if session != nil {
		reqCtx.Session = session
	}

	// Handle based on content type
	contentType := r.Header.Get("Content-Type")
	switch {
	case strings.Contains(contentType, "application/json"):
		h.handleJSONRequest(ctx, w, reqCtx, body)
	case strings.Contains(contentType, "application/jsonrpc"):
		h.handleJSONRPCRequest(ctx, w, reqCtx, body)
	default:
		writeHTTPError(w, http.StatusUnsupportedMediaType, "unsupported content type")
	}
}

// handleJSONRequest handles plain JSON requests.
func (h *HTTPHandler) handleJSONRequest(ctx context.Context, w http.ResponseWriter, reqCtx *RequestContext, body []byte) {
	// Parse params based on method
	params, err := a2a.UnmarshalParams(body)
	if err != nil {
		writeHTTPError(w, http.StatusBadRequest, fmt.Sprintf("invalid request parameters: %v", err))
		return
	}
	reqCtx.Params = params

	// Route to handler
	handler, err := h.router.Route(reqCtx.Method)
	if err != nil {
		writeHTTPError(w, http.StatusNotFound, "method not found")
		return
	}

	// Execute handler
	result, err := handler.Handle(ctx, reqCtx.Method, params)
	if err != nil {
		writeHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.MarshalWrite(w, result); err != nil {
		// Log error but response is already partially written
		fmt.Printf("failed to write response: %v\n", err)
	}
}

// handleJSONRPCRequest handles JSON-RPC 2.0 requests.
func (h *HTTPHandler) handleJSONRPCRequest(ctx context.Context, w http.ResponseWriter, reqCtx *RequestContext, body []byte) {
	// Parse JSON-RPC request
	var req jsonrpc2.Request
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONRPCError(w, jsonrpc2.ID{}, a2a.ErrParse.Code, "parse error")
		return
	}

	reqCtx.Method = a2a.Method(req.Method)

	// Parse params
	if req.Params != nil {
		params, err := parseJSONRPCParams(reqCtx.Method, req.Params)
		if err != nil {
			writeJSONRPCError(w, req.ID, a2a.ErrInvalidParams.Code, "invalid params")
			return
		}
		reqCtx.Params = params
	}

	// Route to handler
	handler, err := h.router.Route(reqCtx.Method)
	if err != nil {
		writeJSONRPCError(w, req.ID, a2a.ErrMethodNotFound.Code, "method not found")
		return
	}

	// Execute handler
	result, err := handler.Handle(ctx, reqCtx.Method, reqCtx.Params)
	if err != nil {
		// Convert error to JSON-RPC error
		code := a2a.ErrInternal.Code
		var jsonrpcErr *a2a.Error
		if e, ok := err.(*a2a.Error); ok {
			jsonrpcErr = e
		} else {
			jsonrpcErr = &a2a.Error{
				Code:    code,
				Message: err.Error(),
			}
		}
		writeJSONRPCError(w, req.ID, jsonrpcErr.Code, jsonrpcErr.Message)
		return
	}

	data, err := json.Marshal(result)
	if err != nil {
		writeJSONRPCError(w, req.ID, a2a.ErrInvalidParams.Code, err.Error())
	}

	// Write JSON-RPC response
	resp := &jsonrpc2.Response{
		ID:     req.ID,
		Result: data,
	}

	w.Header().Set("Content-Type", "application/jsonrpc+json")
	w.WriteHeader(http.StatusOK)
	if err := json.MarshalWrite(w, resp); err != nil {
		fmt.Printf("failed to write response: %v\n", err)
	}
}

// extractMethod extracts the A2A method from the URL path.
func extractMethod(path string) string {
	// Remove leading slash and any query parameters
	path = strings.TrimPrefix(path, "/")
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}
	return path
}

// extractSession extracts session information from the request.
func extractSession(r *http.Request) *Session {
	// Look for session in cookie or header
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		if cookie, err := r.Cookie("session"); err == nil {
			sessionID = cookie.Value
		}
	}

	if sessionID == "" {
		return nil
	}

	// TODO: Load session from store
	return &Session{
		ID:       sessionID,
		Metadata: make(map[string]any),
	}
}

// parseJSONRPCParams parses JSON-RPC params into A2A params.
func parseJSONRPCParams(method a2a.Method, params jsontext.Value) (a2a.Params, error) {
	// Create a temporary structure to include method for unmarshaling
	type methodParams struct {
		Method a2a.Method `json:"method"`
	}

	// Marshal the params with method
	mp := methodParams{Method: method}
	methodJSON, err := json.Marshal(mp)
	if err != nil {
		return nil, err
	}

	// Merge method JSON with params
	var merged map[string]any
	if err := json.Unmarshal(methodJSON, &merged); err != nil {
		return nil, err
	}

	var paramsMap map[string]any
	if err := json.Unmarshal(params, &paramsMap); err != nil {
		return nil, err
	}

	for k, v := range paramsMap {
		merged[k] = v
	}

	// Marshal back to JSON
	mergedJSON, err := json.Marshal(merged)
	if err != nil {
		return nil, err
	}

	// Unmarshal into proper params type
	return a2a.UnmarshalParams(mergedJSON)
}

// writeHTTPError writes an HTTP error response.
func writeHTTPError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := map[string]string{
		"error": message,
	}
	_ = json.MarshalWrite(w, resp)
}

// writeJSONRPCError writes a JSON-RPC error response.
func writeJSONRPCError(w http.ResponseWriter, id jsonrpc2.ID, code int64, message string) {
	resp := &jsonrpc2.Response{
		ID: id,
		Error: &a2a.Error{
			Code:    code,
			Message: message,
		},
	}

	w.Header().Set("Content-Type", "application/jsonrpc+json")
	w.WriteHeader(http.StatusOK) // JSON-RPC errors still use 200 OK
	_ = json.MarshalWrite(w, resp)
}

