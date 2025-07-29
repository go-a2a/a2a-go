# A2A Go Implementation TODO

## Overview

This document outlines the implementation status of the A2A (Agent-to-Agent) protocol in Go, comparing the current implementation against the protocol specification documented in `docs/a2a.llms.txt`.

## Implementation Status

### ✅ Completed

#### Core Protocol Implementation
- **JSON-RPC 2.0 Foundation**: Fully implemented via `internal/jsonrpc2` package with proper message structures, ID handling, and wire format
- **AgentCard**: Complete implementation with all fields including capabilities, security schemes, skills, provider info
- **Task & TaskStatus**: Complete implementation with all task states and proper status handling
- **Message**: Implemented with role, parts, metadata, and proper JSON marshaling
- **Part Types**: All three types implemented (TextPart, FilePart, DataPart) with proper interfaces
- **Artifact**: Complete implementation with helper functions for creating text/data artifacts
- **Security Schemes**: All four types implemented (APIKey, HTTP, OAuth2, OpenIDConnect)

#### RPC Methods
All documented methods are defined as constants:
- `message/send`, `message/stream`
- `tasks/get`, `tasks/list`, `tasks/cancel`, `tasks/resubscribe`
- `tasks/pushNotificationConfig/*` (set/get/list/delete)
- `agent/authenticatedExtendedCard`

#### Error Handling
- Standard JSON-RPC errors (-32700 to -32603)
- A2A-specific errors (-32001 to -32006)
- Additional implementation error (-32099)

### 🔄 Partially Implemented

#### SSE Streaming Support
- Basic SSE transport implemented in `transport/see.go`
- Server-side event streaming with proper event types
- Client-side SSE connection handling
- **Missing**: Complete integration with message/stream method

#### Transport Layer
- Multiple transport protocols defined (JSONRPC, gRPC, HTTP+JSON)
- SSE transport partially implemented
- Connection management and session handling in place

### ❌ Not Implemented

#### Push Notifications
- [ ] Webhook sender implementation
- [ ] JWT signing/verification for push notification authentication
- [ ] Asynchronous task update mechanism
- [ ] Push notification configuration storage

#### Authentication/Security
- [ ] Authentication middleware implementation
- [ ] JWT token validation
- [ ] JWKS endpoint for signature verification
- [ ] TLS configuration and enforcement
- [ ] API key validation middleware
- [ ] OAuth2 flow implementation

#### Server Implementation
- [ ] Complete request handler for all RPC methods
- [ ] TaskManager implementation
- [ ] Request routing for:
  - [ ] `tasks/get`
  - [ ] `tasks/list`
  - [ ] `tasks/cancel`
  - [ ] `tasks/pushNotificationConfig/set`
  - [ ] `tasks/pushNotificationConfig/get`
  - [ ] `tasks/pushNotificationConfig/list`
  - [ ] `tasks/pushNotificationConfig/delete`
  - [ ] `agent/authenticatedExtendedCard`

#### Agent Discovery
- [ ] Card resolver implementation for `/.well-known/agent-card.json`
- [ ] Automatic discovery mechanism
- [ ] Agent card caching

#### Task Management
- [ ] TaskManager interface implementation
- [ ] Task storage (in-memory and persistent options)
- [ ] Task status transition logic
- [ ] Task history tracking
- [ ] Task artifact management
- [ ] Context management for grouped tasks

#### Streaming
- [ ] Complete `message/stream` implementation
- [ ] SSE event types for streaming responses
- [ ] Reconnection support for `tasks/resubscribe`
- [ ] Stream state management

## Priority Implementation Tasks

### High Priority
1. **Server Request Router**: Implement complete request handling for all RPC methods
2. **TaskManager**: Create task lifecycle management system
3. **Authentication Middleware**: Add security layer for protected endpoints
4. **SSE Streaming Completion**: Finish streaming implementation for real-time updates

### Medium Priority
5. **Push Notifications**: Implement webhook-based task updates
6. **Agent Card Resolver**: Add discovery mechanism
7. **Examples**: Create working examples demonstrating usage patterns
8. **Integration Tests**: Add comprehensive test coverage

### Low Priority
9. **Alternative Transports**: Add gRPC and WebSocket support
10. **Performance Optimization**: Add connection pooling and caching
11. **Monitoring**: Add metrics and observability

## Architecture Recommendations

### TaskManager Design
```go
type TaskManager interface {
    CreateTask(ctx context.Context, message *Message) (*Task, error)
    GetTask(ctx context.Context, taskID string) (*Task, error)
    UpdateTaskStatus(ctx context.Context, taskID string, status TaskStatus) error
    ListTasks(ctx context.Context, contextID string) ([]*Task, error)
    CancelTask(ctx context.Context, taskID string) error
}
```

### Authentication Middleware
```go
type AuthMiddleware interface {
    ValidateRequest(r *http.Request) (*AuthContext, error)
    RequiredScopes(method string) []string
}
```

### Push Notification System
```go
type PushNotificationSender interface {
    SendTaskUpdate(ctx context.Context, config *PushNotificationConfig, event TaskUpdateEvent) error
    ValidateWebhookURL(url string) error
}
```

## Testing Requirements

- Unit tests for all new components
- Integration tests for end-to-end flows
- Security testing for authentication
- Load testing for streaming connections
- Example implementations for common use cases

## Documentation Needs

- API usage guide
- Security configuration guide
- Deployment best practices
- Migration guide from Python implementation
- Performance tuning guide