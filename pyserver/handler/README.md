# A2A Go Handler Package

This package provides request handling infrastructure for A2A servers, ported from Python's A2A server request handlers.

## Package Structure

- **handler.go** - Core handler interfaces and abstractions
- **http.go** - HTTP request handler with JSON and JSON-RPC support
- **websocket.go** - WebSocket handler for bidirectional communication
- **sse.go** - Server-Sent Events handler for streaming responses
- **router.go** - Request routing and method dispatch
- **middleware.go** - Common middleware implementations
- **errors.go** - Error handling and conversion utilities
- **connection.go** - Connection and session management

## Key Features

### Transport Support
- HTTP with plain JSON
- JSON-RPC 2.0 over HTTP
- WebSocket with JSON-RPC
- Server-Sent Events (SSE) for streaming

### Middleware Chain
- Recovery (panic handling)
- Logging with timing
- Timeout enforcement
- Request validation
- Authentication/authorization
- Metrics collection

### Handler Patterns
- Interface-based design for extensibility
- Context propagation for cancellation
- Streaming support via channels
- Session and connection management

## Usage Examples

See `example_test.go` for comprehensive usage examples.

## Implementation Notes

### Completed Features
- Full A2A protocol method routing
- Request/response handling for all transport types
- Middleware chain with common implementations
- Session management with in-memory store
- Connection pooling for WebSocket/SSE
- Error handling with protocol-specific formatting

### TODO/Known Issues
1. `AgentHandler.Handle()` implementation needs completion
2. WebSocket origin checking needs proper implementation
3. SSE client `readEvents()` for testing
4. Production session store (Redis/database)
5. Configurable buffer sizes for performance tuning

## Architecture

```
Transport Layer (HTTP/WS/SSE)
    ↓
Router (Method Dispatch)
    ↓
Middleware Chain
    ↓
Handler Implementation
    ↓
Agent Executor
```

## Testing

Unit tests and integration tests are pending implementation (tracked in TODO items 23-25).

## Performance Considerations

- WebSocket creates 2 goroutines per connection
- Channel buffers are currently fixed size
- In-memory session store not suitable for production scale