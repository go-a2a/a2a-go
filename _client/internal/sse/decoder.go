// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package sse

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-json-experiment/json"
)

// Event represents a Server-Sent Event.
type Event struct {
	Type  string `json:"type,omitempty"`
	Data  string `json:"data,omitempty"`
	ID    string `json:"id,omitempty"`
	Retry int    `json:"retry,omitempty"`
}

// Decoder decodes Server-Sent Events from an io.Reader.
type Decoder struct {
	reader  io.Reader
	scanner *bufio.Scanner
}

// NewDecoder creates a new SSE decoder.
func NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{
		reader:  reader,
		scanner: bufio.NewScanner(reader),
	}
}

// Decode decodes the next SSE event from the stream.
func (d *Decoder) Decode() (*Event, error) {
	event := &Event{}

	for d.scanner.Scan() {
		line := d.scanner.Text()

		// Empty line indicates end of event
		if line == "" {
			if event.Data != "" || event.Type != "" {
				return event, nil
			}
			continue
		}

		// Comments (lines starting with :) are ignored
		if strings.HasPrefix(line, ":") {
			continue
		}

		// Parse field and value
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		field := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch field {
		case "event":
			event.Type = value
		case "data":
			if event.Data != "" {
				event.Data += "\n"
			}
			event.Data += value
		case "id":
			event.ID = value
		case "retry":
			// Parse retry as integer
			var retry int
			if _, err := fmt.Sscanf(value, "%d", &retry); err == nil {
				event.Retry = retry
			}
		}
	}

	if err := d.scanner.Err(); err != nil {
		return nil, fmt.Errorf("SSE scanner error: %w", err)
	}

	// EOF reached
	if event.Data != "" || event.Type != "" {
		return event, nil
	}

	return nil, io.EOF
}

// DecodeJSON decodes the next SSE event and unmarshals the data as JSON.
func (d *Decoder) DecodeJSON(v any) error {
	event, err := d.Decode()
	if err != nil {
		return err
	}

	if event.Data == "" {
		return fmt.Errorf("SSE event has no data")
	}

	if err := json.Unmarshal([]byte(event.Data), v); err != nil {
		return fmt.Errorf("failed to unmarshal SSE event data: %w", err)
	}

	return nil
}

// StreamDecoder provides a channel-based interface for decoding SSE events.
type StreamDecoder struct {
	decoder *Decoder
	events  chan *Event
	errors  chan error
	done    chan struct{}
}

// NewStreamDecoder creates a new streaming SSE decoder.
func NewStreamDecoder(reader io.Reader) *StreamDecoder {
	return &StreamDecoder{
		decoder: NewDecoder(reader),
		events:  make(chan *Event, 10),
		errors:  make(chan error, 1),
		done:    make(chan struct{}),
	}
}

// Start starts the streaming decoder.
func (sd *StreamDecoder) Start(ctx context.Context) {
	go func() {
		defer close(sd.events)
		defer close(sd.errors)
		defer close(sd.done)

		for {
			select {
			case <-ctx.Done():
				sd.errors <- ctx.Err()
				return
			default:
				event, err := sd.decoder.Decode()
				if err != nil {
					if err == io.EOF {
						return
					}
					sd.errors <- err
					return
				}

				select {
				case sd.events <- event:
				case <-ctx.Done():
					sd.errors <- ctx.Err()
					return
				}
			}
		}
	}()
}

// Events returns a channel of SSE events.
func (sd *StreamDecoder) Events() <-chan *Event {
	return sd.events
}

// Errors returns a channel of errors.
func (sd *StreamDecoder) Errors() <-chan error {
	return sd.errors
}

// Done returns a channel that closes when the decoder is done.
func (sd *StreamDecoder) Done() <-chan struct{} {
	return sd.done
}

// JSONStreamDecoder provides a channel-based interface for decoding SSE events as JSON.
type JSONStreamDecoder struct {
	decoder *Decoder
	results chan any
	errors  chan error
	done    chan struct{}
}

// NewJSONStreamDecoder creates a new JSON streaming SSE decoder.
func NewJSONStreamDecoder(reader io.Reader) *JSONStreamDecoder {
	return &JSONStreamDecoder{
		decoder: NewDecoder(reader),
		results: make(chan any, 10),
		errors:  make(chan error, 1),
		done:    make(chan struct{}),
	}
}

// Start starts the JSON streaming decoder.
func (jsd *JSONStreamDecoder) Start(ctx context.Context, resultType any) {
	go func() {
		defer close(jsd.results)
		defer close(jsd.errors)
		defer close(jsd.done)

		for {
			select {
			case <-ctx.Done():
				jsd.errors <- ctx.Err()
				return
			default:
				event, err := jsd.decoder.Decode()
				if err != nil {
					if err == io.EOF {
						return
					}
					jsd.errors <- err
					return
				}

				if event.Data == "" {
					continue
				}

				// Create a new instance of the result type
				result := reflect.New(reflect.TypeOf(resultType)).Interface()

				if err := json.Unmarshal([]byte(event.Data), result); err != nil {
					jsd.errors <- fmt.Errorf("failed to unmarshal SSE event data: %w", err)
					return
				}

				select {
				case jsd.results <- result:
				case <-ctx.Done():
					jsd.errors <- ctx.Err()
					return
				}
			}
		}
	}()
}

// Results returns a channel of decoded JSON results.
func (jsd *JSONStreamDecoder) Results() <-chan any {
	return jsd.results
}

// Errors returns a channel of errors.
func (jsd *JSONStreamDecoder) Errors() <-chan error {
	return jsd.errors
}

// Done returns a channel that closes when the decoder is done.
func (jsd *JSONStreamDecoder) Done() <-chan struct{} {
	return jsd.done
}

// Client provides a simple SSE client.
type Client struct {
	httpClient *http.Client
	headers    map[string]string
}

// NewClient creates a new SSE client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 0}, // No timeout for streaming
		headers:    make(map[string]string),
	}
}

// WithHTTPClient sets the HTTP client.
func (c *Client) WithHTTPClient(client *http.Client) *Client {
	c.httpClient = client
	return c
}

// WithHeader adds a header to the client.
func (c *Client) WithHeader(key, value string) *Client {
	c.headers[key] = value
	return c
}

// Connect connects to an SSE endpoint and returns a decoder.
func (c *Client) Connect(ctx context.Context, url string) (*Decoder, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set SSE headers
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	// Add custom headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSE endpoint: %w", err)
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("SSE connection failed with status %d", resp.StatusCode)
	}

	return NewDecoder(resp.Body), nil
}

// ConnectStream connects to an SSE endpoint and returns a streaming decoder.
func (c *Client) ConnectStream(ctx context.Context, url string) (*StreamDecoder, error) {
	decoder, err := c.Connect(ctx, url)
	if err != nil {
		return nil, err
	}

	streamDecoder := &StreamDecoder{
		decoder: decoder,
		events:  make(chan *Event, 10),
		errors:  make(chan error, 1),
		done:    make(chan struct{}),
	}

	streamDecoder.Start(ctx)
	return streamDecoder, nil
}
