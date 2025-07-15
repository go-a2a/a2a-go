package task

import (
	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/server/event"
	"testing"
)

func TestEventTypes(t *testing.T) {
	// Test TaskStatusUpdateEvent
	statusEvent := &event.TaskStatusUpdateEvent{
		TaskID: "test",
		Status: a2a.TaskStatus{State: a2a.TaskStateRunning},
	}

	if statusEvent.EventType() != "task_status_update" {
		t.Errorf("Expected event type 'task_status_update', got %s", statusEvent.EventType())
	}

	// Test TaskArtifactUpdateEvent
	artifactEvent := &event.TaskArtifactUpdateEvent{
		TaskID: "test",
		Artifact: &a2a.Artifact{
			ArtifactID: "test-artifact",
			Name:       "test",
			Parts:      []*a2a.PartWrapper{},
		},
	}

	if artifactEvent.EventType() != "task_artifact_update" {
		t.Errorf("Expected event type 'task_artifact_update', got %s", artifactEvent.EventType())
	}

	// Test TaskEvent
	task := &a2a.Task{
		ID:        "test",
		ContextID: "test",
		Status:    a2a.TaskStatus{State: a2a.TaskStateRunning},
		History:   []*a2a.Message{},
		Artifacts: []*a2a.Artifact{},
	}

	taskEvent := &event.TaskEvent{
		Task: task,
	}

	if taskEvent.EventType() != "task" {
		t.Errorf("Expected event type 'task', got %s", taskEvent.EventType())
	}
}
