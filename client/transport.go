// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bytedance/sonic"
)

// Transport handles JSON-RPC communication with the A2A server.
type Transport struct {
	baseURL    string
	httpClient *http.Client
	headers    http.Header
}

// NewTransport creates a new Transport.
func NewTransport(baseURL string, httpClient *http.Client) *Transport {
	// Ensure baseURL ends with /
	if !strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL + "/"
	}

	return &Transport{
		baseURL:    baseURL,
		httpClient: httpClient,
		headers:    make(http.Header),
	}
}

// SetHeaders sets the headers to include in requests.
func (t *Transport) SetHeaders(headers http.Header) {
	t.headers = headers.Clone()
}

// Call sends a JSON-RPC request and parses the response.
func (t *Transport) Call(ctx context.Context, req any, resp any) error {
	// Encode the request
	data, err := sonic.ConfigFastest.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", t.baseURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	for k, vs := range t.headers {
		for _, v := range vs {
			httpReq.Header.Add(k, v)
		}
	}

	// Send request
	httpResp, err := t.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("sending HTTP request: %w", err)
	}
	defer httpResp.Body.Close()

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("server returned non-OK status: %s, body: %s", httpResp.Status, string(bodyBytes))
	}

	// Decode response
	if err := sonic.ConfigFastest.NewDecoder(httpResp.Body).Decode(resp); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	// Check for JSON-RPC error
	if respWithError, ok := resp.(interface{ Error() error }); ok {
		if err := respWithError.Error(); err != nil {
			return err
		}
	}

	return nil
}

// Stream creates a streaming connection for JSON-RPC notifications.
func (t *Transport) Stream(ctx context.Context, req any) (*StreamConn, error) {
	// Encode the request
	data, err := sonic.ConfigFastest.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", t.baseURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	for k, vs := range t.headers {
		for _, v := range vs {
			httpReq.Header.Add(k, v)
		}
	}

	// Send request
	httpResp, err := t.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending HTTP request: %w", err)
	}

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		httpResp.Body.Close()
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("server returned non-OK status: %s, body: %s", httpResp.Status, string(bodyBytes))
	}

	// Create and return stream connection
	return NewStreamConn(httpResp.Body), nil
}
