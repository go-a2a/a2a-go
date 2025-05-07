# A2A Go Client

This package provides a Go client for the Agent-to-Agent (A2A) protocol, allowing applications to communicate with A2A-compatible agents.

## Features

- Task management (send, get, cancel)
- Streaming task updates
- Push notifications
- Agent card discovery and validation
- Helper functions for creating messages with different content types

## Usage

### Basic Task Interaction

```go
// Create a client
client, err := client.New("https://example.com/a2a")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Create a message
message := client.CreateTextMessage("Hello, agent!")

// Send a task
params := a2a.TaskSendParams{
    ID:      "task123",
    Message: message,
}

task, err := client.SendTask(context.Background(), params)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Task ID: %s\n", task.ID)
fmt.Printf("Task State: %s\n", task.Status.State)
```

### Streaming Updates

```go
stream, err := client.SendTaskStream(context.Background(), params)
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

// Process updates as they arrive
for {
    update, final, err := stream.NextUpdate(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    switch u := update.(type) {
    case a2a.TaskStatusUpdateEvent:
        fmt.Printf("Status update: %s\n", u.Status.State)
    case a2a.TaskArtifactUpdateEvent:
        fmt.Printf("Artifact update: %s\n", u.Artifact.Name)
    }

    if final {
        break
    }
}
```

### Agent Card Discovery

```go
card, err := client.GetAgentCard(context.Background())
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Agent name: %s\n", card.Name)
fmt.Printf("Agent version: %s\n", card.Version)
fmt.Printf("Skills: %d\n", len(card.Skills))
```

### Push Notifications

```go
// Create a push server
ps := client.NewPushServer(":8080")

// Register a handler for a specific task
ps.RegisterHandler("task123", func(ctx context.Context, taskID string, event any) error {
    fmt.Printf("Received notification for task %s\n", taskID)
    return nil
})

// Start the server in a goroutine
go func() {
    if err := ps.Start(); err != nil {
        log.Printf("Push server error: %v", err)
    }
}()

// Create a push notification config
config := client.CreatePushNotificationConfig(
    "http://localhost:8080/task123",
    "token123",
    nil,
)

// Set push notification for a task
taskConfig := client.CreateTaskPushNotificationConfig("task123", config)
result, err := client.SetTaskPushNotification(context.Background(), taskConfig)
if err != nil {
    log.Fatal(err)
}
```

## Error Handling

The client provides helper functions for checking specific error types:

```go
if client.IsTaskNotFoundError(err) {
    // Handle task not found
}

if client.IsTaskNotCancelableError(err) {
    // Handle task not cancelable
}

if client.IsPushNotificationNotSupportedError(err) {
    // Handle push notifications not supported
}
```

## Configuration

The client can be configured with various options:

```go
client, err := client.New(
    "https://example.com/a2a",
    client.WithTimeout(10*time.Second),
    client.WithBearerToken("my-token"),
    client.WithMaxRetries(5),
)
```
