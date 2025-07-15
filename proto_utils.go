// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package a2a provides Protocol Buffer conversion utilities for the Agent-to-Agent (A2A) protocol.
// This file contains bidirectional conversion functions between Go A2A types and Protocol Buffer messages,
// mirroring the functionality of the Python proto_utils.py implementation.
package a2a

import (
	"fmt"

	a2a_v1 "google.golang.org/a2a/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// ToProtoMessage converts a [Message] to a [a2a_v1.Message].
func ToProtoMessage(msg *Message) (*a2a_v1.Message, error) {
	if msg == nil {
		return nil, nil
	}

	protoMsg := &a2a_v1.Message{}

	if msg.ContextID != "" {
		protoMsg.ContextId = msg.ContextID
	}

	// Convert string content to a single text part
	if msg.Content != "" {
		textPart := &a2a_v1.Part{
			Part: &a2a_v1.Part_Text{
				Text: msg.Content,
			},
		}
		protoMsg.Content = []*a2a_v1.Part{textPart}
	}

	if msg.Metadata != nil {
		metadata, err := ToProtoMetadata(msg.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to convert metadata: %w", err)
		}
		protoMsg.Metadata = metadata
	}

	return protoMsg, nil
}

// FromProtoMessage converts a [a2a_v1.Message] to a [Message].
func FromProtoMessage(protoMsg *a2a_v1.Message) (*Message, error) {
	if protoMsg == nil {
		return nil, nil
	}

	msg := &Message{}

	if protoMsg.ContextId != "" {
		msg.ContextID = protoMsg.ContextId
	}

	// Convert parts to string content (simplified approach - take first text part)
	if len(protoMsg.Content) > 0 {
		for _, part := range protoMsg.Content {
			if textPart, ok := part.Part.(*a2a_v1.Part_Text); ok {
				msg.Content = textPart.Text
				break // Take the first text part
			}
		}
	}

	if protoMsg.Metadata != nil {
		metadata, err := FromProtoMetadata(protoMsg.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to convert metadata: %w", err)
		}
		msg.Metadata = metadata
	}

	return msg, nil
}

// ToProtoMetadata converts a Go map[string]any to a [structpb.Struct].
func ToProtoMetadata(metadata map[string]any) (*structpb.Struct, error) {
	if metadata == nil {
		return nil, nil
	}

	protoStruct, err := structpb.NewStruct(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create protobuf struct: %w", err)
	}

	return protoStruct, nil
}

// FromProtoMetadata converts a [structpb.Struct] to a Go map[string]any.
func FromProtoMetadata(protoStruct *structpb.Struct) (map[string]any, error) {
	if protoStruct == nil {
		return nil, nil
	}

	return protoStruct.AsMap(), nil
}

// ToProtoPart converts a [Part] to a [a2a_v1.Part].
func ToProtoPart(part Part) (*a2a_v1.Part, error) {
	if part == nil {
		return nil, nil
	}

	protoPart := &a2a_v1.Part{}

	switch p := part.(type) {
	case *TextPart:
		protoPart.Part = &a2a_v1.Part_Text{
			Text: p.Text,
		}
	case *DataPart:
		// Convert map[string]any to protobuf DataPart
		dataStruct, err := structpb.NewStruct(p.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to create struct for data part: %w", err)
		}
		protoPart.Part = &a2a_v1.Part_Data{
			Data: &a2a_v1.DataPart{
				Data: dataStruct,
			},
		}
	case *ArtifactFilePart:
		protoFilePart, err := ToProtoFile(p.File)
		if err != nil {
			return nil, fmt.Errorf("failed to convert file part: %w", err)
		}
		protoPart.Part = &a2a_v1.Part_File{
			File: protoFilePart,
		}
	default:
		return nil, fmt.Errorf("unknown part type: %T", part)
	}

	return protoPart, nil
}

// FromProtoPart converts a [a2a_v1.Part] to a [Part].
func FromProtoPart(protoPart *a2a_v1.Part) (Part, error) {
	if protoPart == nil {
		return nil, nil
	}

	switch p := protoPart.Part.(type) {
	case *a2a_v1.Part_Text:
		return &TextPart{
			Kind: "text",
			Text: p.Text,
		}, nil
	case *a2a_v1.Part_Data:
		// Convert protobuf DataPart to map[string]any
		data := p.Data.Data.AsMap()
		return &DataPart{
			Kind: "data",
			Data: data,
		}, nil
	case *a2a_v1.Part_File:
		file, err := FromProtoFile(p.File)
		if err != nil {
			return nil, fmt.Errorf("failed to convert file part: %w", err)
		}
		return &ArtifactFilePart{
			Kind: "file",
			File: file,
		}, nil
	default:
		return nil, fmt.Errorf("unknown proto part type: %T", protoPart.Part)
	}
}

// ToProtoData converts a Go map[string]any to a [a2a_v1.DataPart].
func ToProtoData(data map[string]any) (*a2a_v1.DataPart, error) {
	protoData, err := structpb.NewStruct(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert structpb.Struct: %w", err)
	}
	return &a2a_v1.DataPart{
		Data: protoData,
	}, nil
}

// FromProtoData converts a [a2a_v1.DataPart] to a Go map[string]any.
func FromProtoData(protoData *a2a_v1.DataPart) (map[string]any, error) {
	data := protoData.GetData()
	if data == nil {
		return nil, fmt.Errorf("data part is nil")
	}
	return data.AsMap(), nil
}

// ToProtoFile converts a [File] to a [a2a_v1.FilePart].
func ToProtoFile(file File) (*a2a_v1.FilePart, error) {
	if file == nil {
		return nil, nil
	}

	protoFilePart := &a2a_v1.FilePart{}

	switch f := file.(type) {
	case *FileWithURI:
		protoFilePart.MimeType = f.ContentType
		protoFilePart.File = &a2a_v1.FilePart_FileWithUri{
			FileWithUri: f.URI,
		}
	case *FileWithBytes:
		protoFilePart.MimeType = f.ContentType
		protoFilePart.File = &a2a_v1.FilePart_FileWithBytes{
			FileWithBytes: f.Bytes,
		}
	default:
		return nil, fmt.Errorf("unknown file type: %T", file)
	}

	return protoFilePart, nil
}

// FromProtoFile converts a [a2a_v1.FilePart] to a [File].
func FromProtoFile(protoFilePart *a2a_v1.FilePart) (File, error) {
	if protoFilePart == nil {
		return nil, nil
	}

	switch f := protoFilePart.File.(type) {
	case *a2a_v1.FilePart_FileWithUri:
		return &FileWithURI{
			FileBase: FileBase{
				ContentType: protoFilePart.MimeType,
			},
			URI: f.FileWithUri,
		}, nil
	case *a2a_v1.FilePart_FileWithBytes:
		return &FileWithBytes{
			FileBase: FileBase{
				ContentType: protoFilePart.MimeType,
			},
			Bytes: f.FileWithBytes,
		}, nil
	default:
		return nil, fmt.Errorf("unknown proto file type: %T", protoFilePart.File)
	}
}

// ToProtoTask converts a [Task] to a [a2a_v1.Task].
func ToProtoTask(task *Task) (*a2a_v1.Task, error) {
	if task == nil {
		return nil, nil
	}

	protoTask := &a2a_v1.Task{
		Id:        task.ID,
		ContextId: task.ContextID,
	}

	// Convert TaskStatus
	if status, err := ToProtoTaskStatus(&task.Status); err != nil {
		return nil, fmt.Errorf("failed to convert task status: %w", err)
	} else if status != nil {
		protoTask.Status = status
	}

	// Convert History
	if len(task.History) > 0 {
		history := make([]*a2a_v1.Message, len(task.History))
		for i, msg := range task.History {
			if msg == nil {
				return nil, fmt.Errorf("history message at index %d cannot be nil", i)
			}
			protoMsg, err := ToProtoMessage(msg)
			if err != nil {
				return nil, fmt.Errorf("failed to convert history message at index %d: %w", i, err)
			}
			history[i] = protoMsg
		}
		protoTask.History = history
	}

	// Convert Artifacts
	if len(task.Artifacts) > 0 {
		artifacts := make([]*a2a_v1.Artifact, len(task.Artifacts))
		for i, artifact := range task.Artifacts {
			if artifact == nil {
				return nil, fmt.Errorf("artifact at index %d cannot be nil", i)
			}
			protoArtifact, err := ToProtoArtifact(artifact)
			if err != nil {
				return nil, fmt.Errorf("failed to convert artifact at index %d: %w", i, err)
			}
			artifacts[i] = protoArtifact
		}
		protoTask.Artifacts = artifacts
	}

	return protoTask, nil
}

// FromProtoTask converts a [a2a_v1.Task] to a [Task].
func FromProtoTask(protoTask *a2a_v1.Task) (*Task, error) {
	if protoTask == nil {
		return nil, nil
	}

	task := &Task{
		ID:        protoTask.Id,
		ContextID: protoTask.ContextId,
	}

	// Convert TaskStatus
	if protoTask.Status != nil {
		status, err := FromProtoTaskStatus(protoTask.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to convert task status: %w", err)
		}
		task.Status = *status
	}

	// Convert History
	if len(protoTask.History) > 0 {
		history := make([]*Message, len(protoTask.History))
		for i, protoMsg := range protoTask.History {
			if protoMsg == nil {
				return nil, fmt.Errorf("history message at index %d cannot be nil", i)
			}
			msg, err := FromProtoMessage(protoMsg)
			if err != nil {
				return nil, fmt.Errorf("failed to convert history message at index %d: %w", i, err)
			}
			history[i] = msg
		}
		task.History = history
	}

	// Convert Artifacts
	if len(protoTask.Artifacts) > 0 {
		artifacts := make([]*Artifact, len(protoTask.Artifacts))
		for i, protoArtifact := range protoTask.Artifacts {
			if protoArtifact == nil {
				return nil, fmt.Errorf("artifact at index %d cannot be nil", i)
			}
			artifact, err := FromProtoArtifact(protoArtifact)
			if err != nil {
				return nil, fmt.Errorf("failed to convert artifact at index %d: %w", i, err)
			}
			artifacts[i] = artifact
		}
		task.Artifacts = artifacts
	}

	return task, nil
}

// ToProtoTaskStatus converts a [TaskStatus] to a [a2a_v1.TaskStatus].
func ToProtoTaskStatus(status *TaskStatus) (*a2a_v1.TaskStatus, error) {
	if status == nil {
		return nil, nil
	}

	protoState := ToProtoTaskState(status.State)
	return &a2a_v1.TaskStatus{
		State: protoState,
	}, nil
}

// FromProtoTaskStatus converts a [a2a_v1.TaskStatus] to a [TaskStatus].
func FromProtoTaskStatus(protoStatus *a2a_v1.TaskStatus) (*TaskStatus, error) {
	if protoStatus == nil {
		return nil, nil
	}

	state := FromProtoTaskState(protoStatus.State)
	return &TaskStatus{
		State: state,
	}, nil
}

// ToProtoTaskState converts a [TaskState] to a [a2a_v1.TaskState].
func ToProtoTaskState(state TaskState) a2a_v1.TaskState {
	switch state {
	case TaskStateSubmitted:
		return a2a_v1.TaskState_TASK_STATE_SUBMITTED
	case TaskStateRunning:
		return a2a_v1.TaskState_TASK_STATE_WORKING
	case TaskStateCompleted:
		return a2a_v1.TaskState_TASK_STATE_COMPLETED
	case TaskStateFailed:
		return a2a_v1.TaskState_TASK_STATE_FAILED
	case TaskStateCanceled:
		return a2a_v1.TaskState_TASK_STATE_CANCELLED
	default:
		return a2a_v1.TaskState_TASK_STATE_UNSPECIFIED
	}
}

// FromProtoTaskState converts a [a2a_v1.TaskState] to a [TaskState].
func FromProtoTaskState(protoState a2a_v1.TaskState) TaskState {
	switch protoState {
	case a2a_v1.TaskState_TASK_STATE_SUBMITTED:
		return TaskStateSubmitted
	case a2a_v1.TaskState_TASK_STATE_WORKING:
		return TaskStateRunning
	case a2a_v1.TaskState_TASK_STATE_COMPLETED:
		return TaskStateCompleted
	case a2a_v1.TaskState_TASK_STATE_FAILED:
		return TaskStateFailed
	case a2a_v1.TaskState_TASK_STATE_CANCELLED:
		return TaskStateCanceled
	default:
		return TaskStateSubmitted // Default to submitted for unknown states
	}
}

// ToProtoArtifact converts a [Artifact] to a [a2a_v1.Artifact].
func ToProtoArtifact(artifact *Artifact) (*a2a_v1.Artifact, error) {
	if artifact == nil {
		return nil, nil
	}

	protoArtifact := &a2a_v1.Artifact{
		ArtifactId:  artifact.ArtifactID,
		Name:        artifact.Name,
		Description: artifact.Description,
	}

	// Convert Parts
	if len(artifact.Parts) > 0 {
		parts := make([]*a2a_v1.Part, len(artifact.Parts))
		for i, partWrapper := range artifact.Parts {
			if partWrapper == nil {
				return nil, fmt.Errorf("part wrapper at index %d cannot be nil", i)
			}
			part := partWrapper.GetPart()
			if part == nil {
				return nil, fmt.Errorf("part at index %d cannot be nil", i)
			}
			protoPart, err := ToProtoPart(part)
			if err != nil {
				return nil, fmt.Errorf("failed to convert part at index %d: %w", i, err)
			}
			parts[i] = protoPart
		}
		protoArtifact.Parts = parts
	}

	// Convert Extensions
	if len(artifact.Extensions) > 0 {
		protoArtifact.Extensions = make([]string, len(artifact.Extensions))
		copy(protoArtifact.Extensions, artifact.Extensions)
	}

	// Convert Metadata
	if artifact.Metadata != nil {
		metadata, err := ToProtoMetadata(artifact.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to convert metadata: %w", err)
		}
		protoArtifact.Metadata = metadata
	}

	return protoArtifact, nil
}

// FromProtoArtifact converts a [a2a_v1.Artifact] to a [Artifact].
func FromProtoArtifact(protoArtifact *a2a_v1.Artifact) (*Artifact, error) {
	if protoArtifact == nil {
		return nil, nil
	}

	artifact := &Artifact{
		ArtifactID:  protoArtifact.ArtifactId,
		Name:        protoArtifact.Name,
		Description: protoArtifact.Description,
	}

	// Convert Parts
	if len(protoArtifact.Parts) > 0 {
		parts := make([]*PartWrapper, len(protoArtifact.Parts))
		for i, protoPart := range protoArtifact.Parts {
			if protoPart == nil {
				return nil, fmt.Errorf("proto part at index %d cannot be nil", i)
			}
			part, err := FromProtoPart(protoPart)
			if err != nil {
				return nil, fmt.Errorf("failed to convert part at index %d: %w", i, err)
			}
			parts[i] = NewPartWrapper(part)
		}
		artifact.Parts = parts
	}

	// Convert Extensions
	if len(protoArtifact.Extensions) > 0 {
		artifact.Extensions = make([]string, len(protoArtifact.Extensions))
		copy(artifact.Extensions, protoArtifact.Extensions)
	}

	// Convert Metadata
	if protoArtifact.Metadata != nil {
		metadata, err := FromProtoMetadata(protoArtifact.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to convert metadata: %w", err)
		}
		artifact.Metadata = metadata
	}

	return artifact, nil
}

// ToProtoAuthenticationInfo converts a [AuthenticationInfo] to a [a2a_v1.AuthenticationInfo].
func ToProtoAuthenticationInfo(auth *AuthenticationInfo) (*a2a_v1.AuthenticationInfo, error) {
	if auth == nil {
		return nil, nil
	}

	protoAuth := &a2a_v1.AuthenticationInfo{
		Credentials: auth.Credentials,
	}

	if len(auth.Schemes) > 0 {
		protoAuth.Schemes = make([]string, len(auth.Schemes))
		copy(protoAuth.Schemes, auth.Schemes)
	}

	return protoAuth, nil
}

// FromProtoAuthenticationInfo converts a [a2a_v1.AuthenticationInfo] to a [AuthenticationInfo].
func FromProtoAuthenticationInfo(protoAuth *a2a_v1.AuthenticationInfo) (*AuthenticationInfo, error) {
	if protoAuth == nil {
		return nil, nil
	}

	auth := &AuthenticationInfo{
		Credentials: protoAuth.Credentials,
	}

	if len(protoAuth.Schemes) > 0 {
		auth.Schemes = make([]string, len(protoAuth.Schemes))
		copy(auth.Schemes, protoAuth.Schemes)
	}

	return auth, nil
}

// ToProtoPushNotificationConfig converts a [PushNotificationConfig] to a [a2a_v1.PushNotificationConfig].
func ToProtoPushNotificationConfig(pnc *PushNotificationConfig) (*a2a_v1.PushNotificationConfig, error) {
	if pnc == nil {
		return nil, nil
	}

	protoPnc := &a2a_v1.PushNotificationConfig{
		Id:    pnc.ID,
		Url:   pnc.URL,
		Token: pnc.Token,
	}

	if pnc.Authentication != nil {
		protoAuth, err := ToProtoAuthenticationInfo(pnc.Authentication)
		if err != nil {
			return nil, fmt.Errorf("failed to convert authentication info: %w", err)
		}
		protoPnc.Authentication = protoAuth
	}

	return protoPnc, nil
}

// FromProtoPushNotificationConfig converts a [a2a_v1.PushNotificationConfig] to a [PushNotificationConfig].
func FromProtoPushNotificationConfig(protoPnc *a2a_v1.PushNotificationConfig) (*PushNotificationConfig, error) {
	if protoPnc == nil {
		return nil, nil
	}

	pnc := &PushNotificationConfig{
		ID:    protoPnc.Id,
		URL:   protoPnc.Url,
		Token: protoPnc.Token,
	}

	if protoPnc.Authentication != nil {
		auth, err := FromProtoAuthenticationInfo(protoPnc.Authentication)
		if err != nil {
			return nil, fmt.Errorf("failed to convert authentication info: %w", err)
		}
		pnc.Authentication = auth
	}

	return pnc, nil
}

// ToProtoTaskArtifactUpdateEvent converts a [TaskArtifactUpdateEvent] to a [a2a_v1.TaskArtifactUpdateEvent].
func ToProtoTaskArtifactUpdateEvent(event *TaskArtifactUpdateEvent) (*a2a_v1.TaskArtifactUpdateEvent, error) {
	if event == nil {
		return nil, nil
	}

	protoEvent := &a2a_v1.TaskArtifactUpdateEvent{
		Append: event.Append,
	}

	if event.Artifact != nil {
		protoArtifact, err := ToProtoArtifact(event.Artifact)
		if err != nil {
			return nil, fmt.Errorf("failed to convert artifact: %w", err)
		}
		protoEvent.Artifact = protoArtifact
	}

	return protoEvent, nil
}

// FromProtoTaskArtifactUpdateEvent converts a [a2a_v1.TaskArtifactUpdateEvent] to a [TaskArtifactUpdateEvent].
func FromProtoTaskArtifactUpdateEvent(protoEvent *a2a_v1.TaskArtifactUpdateEvent) (*TaskArtifactUpdateEvent, error) {
	if protoEvent == nil {
		return nil, nil
	}

	event := &TaskArtifactUpdateEvent{
		Append: protoEvent.Append,
	}

	if protoEvent.Artifact != nil {
		artifact, err := FromProtoArtifact(protoEvent.Artifact)
		if err != nil {
			return nil, fmt.Errorf("failed to convert artifact: %w", err)
		}
		event.Artifact = artifact
	}

	return event, nil
}

// ToProtoTaskStatusUpdateEvent converts a [TaskStatusUpdateEvent] to a [a2a_v1.TaskStatusUpdateEvent].
func ToProtoTaskStatusUpdateEvent(event *TaskStatusUpdateEvent) (*a2a_v1.TaskStatusUpdateEvent, error) {
	if event == nil {
		return nil, nil
	}

	protoEvent := &a2a_v1.TaskStatusUpdateEvent{
		TaskId:    event.TaskID,
		ContextId: event.ContextID,
		Final:     event.Final,
	}

	// Convert TaskStatus
	protoStatus, err := ToProtoTaskStatus(&event.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to convert task status: %w", err)
	}
	protoEvent.Status = protoStatus

	// Convert Metadata
	if event.Metadata != nil {
		metadata, err := ToProtoMetadata(event.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to convert metadata: %w", err)
		}
		protoEvent.Metadata = metadata
	}

	return protoEvent, nil
}

// FromProtoTaskStatusUpdateEvent converts a [a2a_v1.TaskStatusUpdateEvent] to a [TaskStatusUpdateEvent].
func FromProtoTaskStatusUpdateEvent(protoEvent *a2a_v1.TaskStatusUpdateEvent) (*TaskStatusUpdateEvent, error) {
	if protoEvent == nil {
		return nil, nil
	}

	event := &TaskStatusUpdateEvent{
		TaskID:    protoEvent.TaskId,
		ContextID: protoEvent.ContextId,
		Final:     protoEvent.Final,
	}

	// Convert TaskStatus
	if protoEvent.Status != nil {
		status, err := FromProtoTaskStatus(protoEvent.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to convert task status: %w", err)
		}
		event.Status = *status
	}

	// Convert Metadata
	if protoEvent.Metadata != nil {
		metadata, err := FromProtoMetadata(protoEvent.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to convert metadata: %w", err)
		}
		event.Metadata = metadata
	}

	return event, nil
}

// ToProtoSendMessageConfiguration converts a [SendMessageConfiguration] to a [a2a_v1.SendMessageConfiguration].
func ToProtoSendMessageConfiguration(smc *SendMessageConfiguration) (*a2a_v1.SendMessageConfiguration, error) {
	if smc == nil {
		return nil, nil
	}

	protoSmc := &a2a_v1.SendMessageConfiguration{
		HistoryLength: smc.HistoryLength,
		Blocking:      smc.Blocking,
	}

	if len(smc.AcceptedOutputModes) > 0 {
		protoSmc.AcceptedOutputModes = make([]string, len(smc.AcceptedOutputModes))
		copy(protoSmc.AcceptedOutputModes, smc.AcceptedOutputModes)
	}

	if smc.PushNotification != nil {
		protoPnc, err := ToProtoPushNotificationConfig(smc.PushNotification)
		if err != nil {
			return nil, fmt.Errorf("failed to convert push notification config: %w", err)
		}
		protoSmc.PushNotification = protoPnc
	}

	return protoSmc, nil
}

// FromProtoSendMessageConfiguration converts a [a2a_v1.SendMessageConfiguration] to a [SendMessageConfiguration].
func FromProtoSendMessageConfiguration(protoSmc *a2a_v1.SendMessageConfiguration) (*SendMessageConfiguration, error) {
	if protoSmc == nil {
		return nil, nil
	}

	smc := &SendMessageConfiguration{
		HistoryLength: protoSmc.HistoryLength,
		Blocking:      protoSmc.Blocking,
	}

	if len(protoSmc.AcceptedOutputModes) > 0 {
		smc.AcceptedOutputModes = make([]string, len(protoSmc.AcceptedOutputModes))
		copy(smc.AcceptedOutputModes, protoSmc.AcceptedOutputModes)
	}

	if protoSmc.PushNotification != nil {
		pnc, err := FromProtoPushNotificationConfig(protoSmc.PushNotification)
		if err != nil {
			return nil, fmt.Errorf("failed to convert push notification config: %w", err)
		}
		smc.PushNotification = pnc
	}

	return smc, nil
}

// ToProtoUpdateEvent converts various update event types to appropriate Protocol Buffer types.
//
// This function handles multiple types of update events that can occur in the A2A protocol.
func ToProtoUpdateEvent(event any) (any, error) {
	if event == nil {
		return nil, nil
	}

	switch e := event.(type) {
	case *TaskStatusUpdateEvent:
		return ToProtoTaskStatusUpdateEvent(e)
	case *TaskArtifactUpdateEvent:
		return ToProtoTaskArtifactUpdateEvent(e)
	default:
		return nil, fmt.Errorf("unknown update event type: %T", event)
	}
}

// ToProtoTaskOrMessage converts a Task or Message to the appropriate Protocol Buffer type.
//
// This function handles the polymorphic conversion between Task and Message types.
func ToProtoTaskOrMessage(payload any) (any, error) {
	if payload == nil {
		return nil, nil
	}

	switch p := payload.(type) {
	case *Task:
		return ToProtoTask(p)
	case *Message:
		return ToProtoMessage(p)
	default:
		return nil, fmt.Errorf("unknown task or message type: %T", payload)
	}
}

// ToProtoStreamResponse converts various types to a [a2a_v1.StreamResponse].
//
// This mirrors the Python implementation's polymorphic update_event method.
func ToProtoStreamResponse(payload any) (*a2a_v1.StreamResponse, error) {
	if payload == nil {
		return nil, nil
	}

	response := &a2a_v1.StreamResponse{}

	switch p := payload.(type) {
	case *Task:
		protoTask, err := ToProtoTask(p)
		if err != nil {
			return nil, fmt.Errorf("failed to convert task: %w", err)
		}
		response.Payload = &a2a_v1.StreamResponse_Task{
			Task: protoTask,
		}
	case *Message:
		protoMessage, err := ToProtoMessage(p)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message: %w", err)
		}
		response.Payload = &a2a_v1.StreamResponse_Msg{
			Msg: protoMessage,
		}
	case *TaskArtifactUpdateEvent:
		protoEvent, err := ToProtoTaskArtifactUpdateEvent(p)
		if err != nil {
			return nil, fmt.Errorf("failed to convert task artifact update event: %w", err)
		}
		response.Payload = &a2a_v1.StreamResponse_ArtifactUpdate{
			ArtifactUpdate: protoEvent,
		}
	default:
		return nil, fmt.Errorf("unknown payload type for stream response: %T", payload)
	}

	return response, nil
}

// FromProtoStreamResponse converts a [a2a_v1.StreamResponse] to the appropriate Go type.
func FromProtoStreamResponse(protoResponse *a2a_v1.StreamResponse) (any, error) {
	if protoResponse == nil {
		return nil, nil
	}

	switch payload := protoResponse.Payload.(type) {
	case *a2a_v1.StreamResponse_Task:
		return FromProtoTask(payload.Task)
	case *a2a_v1.StreamResponse_Msg:
		return FromProtoMessage(payload.Msg)
	case *a2a_v1.StreamResponse_StatusUpdate:
		// Note: TaskStatusUpdateEvent is not defined in types.go, so we return the protobuf type
		return payload.StatusUpdate, nil
	case *a2a_v1.StreamResponse_ArtifactUpdate:
		return FromProtoTaskArtifactUpdateEvent(payload.ArtifactUpdate)
	default:
		return nil, fmt.Errorf("unknown payload type in stream response: %T", protoResponse.Payload)
	}
}

// ToProtoTaskPushNotificationConfig converts a [TaskPushNotificationConfig] to a [a2a_v1.TaskPushNotificationConfig].
func ToProtoTaskPushNotificationConfig(tpnc *TaskPushNotificationConfig) (*a2a_v1.TaskPushNotificationConfig, error) {
	if tpnc == nil {
		return nil, nil
	}

	protoTpnc := &a2a_v1.TaskPushNotificationConfig{
		Name: tpnc.Name,
	}

	if tpnc.PushNotificationConfig != nil {
		protoPnc, err := ToProtoPushNotificationConfig(tpnc.PushNotificationConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to convert push notification config: %w", err)
		}
		protoTpnc.PushNotificationConfig = protoPnc
	}

	return protoTpnc, nil
}

// FromProtoTaskPushNotificationConfig converts a [a2a_v1.TaskPushNotificationConfig] to a [TaskPushNotificationConfig].
func FromProtoTaskPushNotificationConfig(protoTpnc *a2a_v1.TaskPushNotificationConfig) (*TaskPushNotificationConfig, error) {
	if protoTpnc == nil {
		return nil, nil
	}

	tpnc := &TaskPushNotificationConfig{
		Name: protoTpnc.Name,
	}

	if protoTpnc.PushNotificationConfig != nil {
		pnc, err := FromProtoPushNotificationConfig(protoTpnc.PushNotificationConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to convert push notification config: %w", err)
		}
		tpnc.PushNotificationConfig = pnc
	}

	return tpnc, nil
}

// ToProtoAgentCard converts a [AgentCard] to a [a2a_v1.AgentCard].
func ToProtoAgentCard(ac *AgentCard) (*a2a_v1.AgentCard, error) {
	if ac == nil {
		return nil, nil
	}

	protoCard := &a2a_v1.AgentCard{
		ProtocolVersion:                   ac.ProtocolVersion,
		Name:                              ac.Name,
		Description:                       ac.Description,
		Url:                               ac.URL,
		PreferredTransport:                ac.PreferredTransport,
		Version:                           ac.Version,
		DocumentationUrl:                  ac.DocumentationURL,
		SupportsAuthenticatedExtendedCard: ac.SupportsAuthenticatedExtendedCard,
	}

	// Convert Provider
	if ac.Provider != nil {
		protoProvider, err := ToProtoProvider(ac.Provider)
		if err != nil {
			return nil, fmt.Errorf("failed to convert provider: %w", err)
		}
		protoCard.Provider = protoProvider
	}

	// Convert Capabilities
	if ac.Capabilities != nil {
		protoCapabilities, err := ToProtoAgentCapabilities(ac.Capabilities)
		if err != nil {
			return nil, fmt.Errorf("failed to convert capabilities: %w", err)
		}
		protoCard.Capabilities = protoCapabilities
	}

	// Convert Skills
	if len(ac.Skills) > 0 {
		protoSkills := make([]*a2a_v1.AgentSkill, len(ac.Skills))
		for i, skill := range ac.Skills {
			protoSkill, err := ToProtoSkill(skill)
			if err != nil {
				return nil, fmt.Errorf("failed to convert skill at index %d: %w", i, err)
			}
			protoSkills[i] = protoSkill
		}
		protoCard.Skills = protoSkills
	}

	// Convert SecuritySchemes
	if len(ac.SecuritySchemes) > 0 {
		protoSecuritySchemes := make(map[string]*a2a_v1.SecurityScheme)
		for name, scheme := range ac.SecuritySchemes {
			protoScheme, err := ToProtoSecurityScheme(scheme)
			if err != nil {
				return nil, fmt.Errorf("failed to convert security scheme %s: %w", name, err)
			}
			protoSecuritySchemes[name] = protoScheme
		}
		protoCard.SecuritySchemes = protoSecuritySchemes
	}

	// Convert DefaultInputModes
	if len(ac.DefaultInputModes) > 0 {
		protoCard.DefaultInputModes = make([]string, len(ac.DefaultInputModes))
		copy(protoCard.DefaultInputModes, ac.DefaultInputModes)
	}

	// Convert DefaultOutputModes
	if len(ac.DefaultOutputModes) > 0 {
		protoCard.DefaultOutputModes = make([]string, len(ac.DefaultOutputModes))
		copy(protoCard.DefaultOutputModes, ac.DefaultOutputModes)
	}

	return protoCard, nil
}

// FromProtoAgentCard converts a [a2a_v1.AgentCard] to a [AgentCard].
func FromProtoAgentCard(protoCard *a2a_v1.AgentCard) (*AgentCard, error) {
	if protoCard == nil {
		return nil, nil
	}

	ac := &AgentCard{
		ProtocolVersion:                   protoCard.ProtocolVersion,
		Name:                              protoCard.Name,
		Description:                       protoCard.Description,
		URL:                               protoCard.Url,
		PreferredTransport:                protoCard.PreferredTransport,
		Version:                           protoCard.Version,
		DocumentationURL:                  protoCard.DocumentationUrl,
		SupportsAuthenticatedExtendedCard: protoCard.SupportsAuthenticatedExtendedCard,
	}

	// Convert Provider
	if protoCard.Provider != nil {
		provider, err := FromProtoProvider(protoCard.Provider)
		if err != nil {
			return nil, fmt.Errorf("failed to convert provider: %w", err)
		}
		ac.Provider = provider
	}

	// Convert Capabilities
	if protoCard.Capabilities != nil {
		capabilities, err := FromProtoAgentCapabilities(protoCard.Capabilities)
		if err != nil {
			return nil, fmt.Errorf("failed to convert capabilities: %w", err)
		}
		ac.Capabilities = capabilities
	}

	// Convert Skills
	if len(protoCard.Skills) > 0 {
		skills := make([]*AgentSkill, len(protoCard.Skills))
		for i, protoSkill := range protoCard.Skills {
			skill, err := FromProtoSkill(protoSkill)
			if err != nil {
				return nil, fmt.Errorf("failed to convert skill at index %d: %w", i, err)
			}
			skills[i] = skill
		}
		ac.Skills = skills
	}

	// Convert SecuritySchemes
	if len(protoCard.SecuritySchemes) > 0 {
		securitySchemes := make(map[string]*SecurityScheme)
		for name, protoScheme := range protoCard.SecuritySchemes {
			scheme, err := FromProtoSecurityScheme(protoScheme)
			if err != nil {
				return nil, fmt.Errorf("failed to convert security scheme %s: %w", name, err)
			}
			securitySchemes[name] = scheme
		}
		ac.SecuritySchemes = securitySchemes
	}

	// Convert DefaultInputModes
	if len(protoCard.DefaultInputModes) > 0 {
		ac.DefaultInputModes = make([]string, len(protoCard.DefaultInputModes))
		copy(ac.DefaultInputModes, protoCard.DefaultInputModes)
	}

	// Convert DefaultOutputModes
	if len(protoCard.DefaultOutputModes) > 0 {
		ac.DefaultOutputModes = make([]string, len(protoCard.DefaultOutputModes))
		copy(ac.DefaultOutputModes, protoCard.DefaultOutputModes)
	}

	return ac, nil
}

// ToProtoAgentCapabilities converts a [AgentCapabilities] to a [a2a_v1.AgentCapabilities].
func ToProtoAgentCapabilities(ac *AgentCapabilities) (*a2a_v1.AgentCapabilities, error) {
	if ac == nil {
		return nil, nil
	}

	protoCapabilities := &a2a_v1.AgentCapabilities{
		Streaming:         ac.Streaming,
		PushNotifications: ac.PushNotifications,
	}

	if len(ac.Extensions) > 0 {
		protoExtensions := make([]*a2a_v1.AgentExtension, len(ac.Extensions))
		for i, ext := range ac.Extensions {
			protoExtensions[i] = &a2a_v1.AgentExtension{
				Uri:         ext.URI,
				Description: ext.Description,
				Required:    ext.Required,
			}
		}
		protoCapabilities.Extensions = protoExtensions
	}

	return protoCapabilities, nil
}

// FromProtoAgentCapabilities converts a [a2a_v1.AgentCapabilities] to a [AgentCapabilities].
func FromProtoAgentCapabilities(protoCapabilities *a2a_v1.AgentCapabilities) (*AgentCapabilities, error) {
	if protoCapabilities == nil {
		return nil, nil
	}

	ac := &AgentCapabilities{
		Streaming:         protoCapabilities.Streaming,
		PushNotifications: protoCapabilities.PushNotifications,
	}

	if len(protoCapabilities.Extensions) > 0 {
		extensions := make([]*AgentExtension, len(protoCapabilities.Extensions))
		for i, protoExt := range protoCapabilities.Extensions {
			extensions[i] = &AgentExtension{
				URI:         protoExt.Uri,
				Description: protoExt.Description,
				Required:    protoExt.Required,
			}
		}
		ac.Extensions = extensions
	}

	return ac, nil
}

// ToProtoProvider converts a Go []map[string][]string to a slice of [a2a_v1.Security].
func ToProtoSecurity(security []map[string][]string) ([]*a2a_v1.Security, error) {
	if security == nil {
		return nil, nil
	}

	protoSecurities := []*a2a_v1.Security{}
	for _, m := range security {
		schemes := make(map[string]*a2a_v1.StringList)
		for k, v := range m {
			schemes[k] = &a2a_v1.StringList{
				List: v,
			}
		}
		protoSecurities = append(protoSecurities, &a2a_v1.Security{
			Schemes: schemes,
		})
	}

	return protoSecurities, nil
}

// FromProtoSecurity converts a slice of [a2a_v1.Security] to a Go []map[string][]string.
func FromProtoSecurity(protoSecurity []*a2a_v1.Security) ([]map[string][]string, error) {
	if protoSecurity == nil {
		return nil, nil
	}

	security := []map[string][]string{}
	for _, s := range protoSecurity {
		m := map[string][]string{}
		for k, v := range s.GetSchemes() {
			m[k] = v.List
		}
		security = append(security, m)
	}

	return security, nil
}

// ToProtoProvider converts a [AgentProvider] to a [a2a_v1.AgentProvider].
func ToProtoProvider(provider *AgentProvider) (*a2a_v1.AgentProvider, error) {
	if provider == nil {
		return nil, nil
	}

	protoProvider := &a2a_v1.AgentProvider{
		Organization: provider.Organization,
		Url:          provider.URL,
	}

	return protoProvider, nil
}

// FromProtoProvider converts a [a2a_v1.AgentProvider] to a [AgentProvider].
func FromProtoProvider(protoProvider *a2a_v1.AgentProvider) (*AgentProvider, error) {
	if protoProvider == nil {
		return nil, nil
	}

	provider := &AgentProvider{
		Organization: protoProvider.Organization,
		URL:          protoProvider.Url,
	}

	return provider, nil
}

// ToProtoSecuritySchemes converts a Go map of [SecuritySchemes] to a map of [a2a_v1.SecurityScheme].
func ToProtoSecuritySchemes(schemes map[string]*SecurityScheme) (map[string]*a2a_v1.SecurityScheme, error) {
	if schemes == nil {
		return nil, nil
	}

	protoSchemes := make(map[string]*a2a_v1.SecurityScheme)
	for name, scheme := range schemes {
		protoScheme, err := ToProtoSecurityScheme(scheme)
		if err != nil {
			return nil, fmt.Errorf("failed to convert security scheme %s: %w", name, err)
		}
		protoSchemes[name] = protoScheme
	}

	return protoSchemes, nil
}

// FromProtoSecuritySchemes converts a map of [a2a_v1.SecurityScheme] to a Go map of [SecuritySchemes].
func FromProtoSecuritySchemes(protoSchemes map[string]*a2a_v1.SecurityScheme) (map[string]*SecurityScheme, error) {
	if protoSchemes == nil {
		return nil, nil
	}

	schemes := make(map[string]*SecurityScheme)
	for name, protoScheme := range protoSchemes {
		scheme, err := FromProtoSecurityScheme(protoScheme)
		if err != nil {
			return nil, fmt.Errorf("failed to convert security scheme %s: %w", name, err)
		}
		schemes[name] = scheme
	}

	return schemes, nil
}

// ToProtoSecurityScheme converts a [SecurityScheme] to a [a2a_v1.SecurityScheme].
func ToProtoSecurityScheme(scheme *SecurityScheme) (*a2a_v1.SecurityScheme, error) {
	if scheme == nil {
		return nil, nil
	}

	protoScheme := &a2a_v1.SecurityScheme{}

	switch {
	case scheme.APIKey != nil:
		protoAPIKey := &a2a_v1.APIKeySecurityScheme{
			Description: scheme.APIKey.Description,
			Name:        scheme.APIKey.Name,
			Location:    string(scheme.APIKey.In),
		}
		protoScheme.Scheme = &a2a_v1.SecurityScheme_ApiKeySecurityScheme{
			ApiKeySecurityScheme: protoAPIKey,
		}

	case scheme.HTTPAuth != nil:
		protoHTTPAuth := &a2a_v1.HTTPAuthSecurityScheme{
			Description:  scheme.HTTPAuth.Description,
			Scheme:       scheme.HTTPAuth.Scheme,
			BearerFormat: scheme.HTTPAuth.BearerFormat,
		}
		protoScheme.Scheme = &a2a_v1.SecurityScheme_HttpAuthSecurityScheme{
			HttpAuthSecurityScheme: protoHTTPAuth,
		}

	case scheme.OAuth2 != nil:
		protoOAuth2 := &a2a_v1.OAuth2SecurityScheme{
			Description: scheme.OAuth2.Description,
		}
		if scheme.OAuth2.Flows != nil {
			protoFlows, err := ToProtoOAuthFlows(scheme.OAuth2.Flows)
			if err != nil {
				return nil, fmt.Errorf("failed to convert OAuth flows: %w", err)
			}
			protoOAuth2.Flows = protoFlows
		}
		protoScheme.Scheme = &a2a_v1.SecurityScheme_Oauth2SecurityScheme{
			Oauth2SecurityScheme: protoOAuth2,
		}

	case scheme.OpenIDConnect != nil:
		protoOpenIDConnect := &a2a_v1.OpenIdConnectSecurityScheme{
			Description:      scheme.OpenIDConnect.Description,
			OpenIdConnectUrl: scheme.OpenIDConnect.OpenIDConnectURL,
		}
		protoScheme.Scheme = &a2a_v1.SecurityScheme_OpenIdConnectSecurityScheme{
			OpenIdConnectSecurityScheme: protoOpenIDConnect,
		}

	default:
		return nil, fmt.Errorf("security scheme has no valid scheme type")
	}

	return protoScheme, nil
}

// FromProtoSecurityScheme converts a [a2a_v1.SecurityScheme] to a [SecurityScheme].
func FromProtoSecurityScheme(protoScheme *a2a_v1.SecurityScheme) (*SecurityScheme, error) {
	if protoScheme == nil {
		return nil, nil
	}

	scheme := &SecurityScheme{}

	switch s := protoScheme.Scheme.(type) {
	case *a2a_v1.SecurityScheme_ApiKeySecurityScheme:
		scheme.APIKey = &APIKeySecurityScheme{
			Description: s.ApiKeySecurityScheme.Description,
			Name:        s.ApiKeySecurityScheme.Name,
			In:          Location(s.ApiKeySecurityScheme.Location),
		}
	case *a2a_v1.SecurityScheme_HttpAuthSecurityScheme:
		scheme.HTTPAuth = &HTTPAuthSecurityScheme{
			Description:  s.HttpAuthSecurityScheme.Description,
			Scheme:       s.HttpAuthSecurityScheme.Scheme,
			BearerFormat: s.HttpAuthSecurityScheme.BearerFormat,
		}
	case *a2a_v1.SecurityScheme_Oauth2SecurityScheme:
		scheme.OAuth2 = &OAuth2SecurityScheme{
			Description: s.Oauth2SecurityScheme.Description,
		}
		if s.Oauth2SecurityScheme.Flows != nil {
			flows, err := FromProtoOAuthFlows(s.Oauth2SecurityScheme.Flows)
			if err != nil {
				return nil, fmt.Errorf("failed to convert OAuth flows: %w", err)
			}
			scheme.OAuth2.Flows = flows
		}
	case *a2a_v1.SecurityScheme_OpenIdConnectSecurityScheme:
		scheme.OpenIDConnect = &OpenIDConnectSecurityScheme{
			Description:      s.OpenIdConnectSecurityScheme.Description,
			OpenIDConnectURL: s.OpenIdConnectSecurityScheme.OpenIdConnectUrl,
		}
	default:
		return nil, fmt.Errorf("unknown security scheme type: %T", protoScheme.Scheme)
	}

	return scheme, nil
}

// ToProtoOAuthFlows converts a [OAuthFlows] to a [a2a_v1.OAuthFlows].
func ToProtoOAuthFlows(flows *OAuthFlows) (*a2a_v1.OAuthFlows, error) {
	if flows == nil {
		return nil, nil
	}

	protoFlows := &a2a_v1.OAuthFlows{}

	if flows.AuthorizationCode != nil {
		protoAuthCode := &a2a_v1.AuthorizationCodeOAuthFlow{
			AuthorizationUrl: flows.AuthorizationCode.AuthorizationURL,
			TokenUrl:         flows.AuthorizationCode.TokenURL,
			RefreshUrl:       flows.AuthorizationCode.RefreshURL,
			Scopes:           flows.AuthorizationCode.Scopes,
		}
		protoFlows.Flow = &a2a_v1.OAuthFlows_AuthorizationCode{
			AuthorizationCode: protoAuthCode,
		}
	} else if flows.ClientCredentials != nil {
		protoClientCreds := &a2a_v1.ClientCredentialsOAuthFlow{
			TokenUrl:   flows.ClientCredentials.TokenURL,
			RefreshUrl: flows.ClientCredentials.RefreshURL,
			Scopes:     flows.ClientCredentials.Scopes,
		}
		protoFlows.Flow = &a2a_v1.OAuthFlows_ClientCredentials{
			ClientCredentials: protoClientCreds,
		}
	} else if flows.Implicit != nil {
		protoImplicit := &a2a_v1.ImplicitOAuthFlow{
			AuthorizationUrl: flows.Implicit.AuthorizationURL,
			RefreshUrl:       flows.Implicit.RefreshURL,
			Scopes:           flows.Implicit.Scopes,
		}
		protoFlows.Flow = &a2a_v1.OAuthFlows_Implicit{
			Implicit: protoImplicit,
		}
	} else if flows.Password != nil {
		protoPassword := &a2a_v1.PasswordOAuthFlow{
			TokenUrl:   flows.Password.TokenURL,
			RefreshUrl: flows.Password.RefreshURL,
			Scopes:     flows.Password.Scopes,
		}
		protoFlows.Flow = &a2a_v1.OAuthFlows_Password{
			Password: protoPassword,
		}
	}

	return protoFlows, nil
}

// FromProtoOAuthFlows converts a [a2a_v1.OAuthFlows] to a [OAuthFlows].
func FromProtoOAuthFlows(protoFlows *a2a_v1.OAuthFlows) (*OAuthFlows, error) {
	if protoFlows == nil {
		return nil, nil
	}

	flows := &OAuthFlows{}

	switch f := protoFlows.Flow.(type) {
	case *a2a_v1.OAuthFlows_AuthorizationCode:
		flows.AuthorizationCode = &AuthorizationCodeFlow{
			AuthorizationURL: f.AuthorizationCode.AuthorizationUrl,
			TokenURL:         f.AuthorizationCode.TokenUrl,
			RefreshURL:       f.AuthorizationCode.RefreshUrl,
			Scopes:           f.AuthorizationCode.Scopes,
		}
	case *a2a_v1.OAuthFlows_ClientCredentials:
		flows.ClientCredentials = &ClientCredentialsFlow{
			TokenURL:   f.ClientCredentials.TokenUrl,
			RefreshURL: f.ClientCredentials.RefreshUrl,
			Scopes:     f.ClientCredentials.Scopes,
		}
	case *a2a_v1.OAuthFlows_Implicit:
		flows.Implicit = &ImplicitFlow{
			AuthorizationURL: f.Implicit.AuthorizationUrl,
			RefreshURL:       f.Implicit.RefreshUrl,
			Scopes:           f.Implicit.Scopes,
		}
	case *a2a_v1.OAuthFlows_Password:
		flows.Password = &PasswordFlow{
			TokenURL:   f.Password.TokenUrl,
			RefreshURL: f.Password.RefreshUrl,
			Scopes:     f.Password.Scopes,
		}
	}

	return flows, nil
}

// ToProtoSkill converts a [AgentSkill] to a [a2a_v1.AgentSkill].
func ToProtoSkill(skill *AgentSkill) (*a2a_v1.AgentSkill, error) {
	if skill == nil {
		return nil, nil
	}

	protoSkill := &a2a_v1.AgentSkill{
		Id:          skill.ID,
		Name:        skill.Name,
		Description: skill.Description,
	}

	if len(skill.Examples) > 0 {
		protoSkill.Examples = make([]string, len(skill.Examples))
		copy(protoSkill.Examples, skill.Examples)
	}

	if len(skill.Tags) > 0 {
		protoSkill.Tags = make([]string, len(skill.Tags))
		copy(protoSkill.Tags, skill.Tags)
	}

	if len(skill.InputModes) > 0 {
		protoSkill.InputModes = make([]string, len(skill.InputModes))
		copy(protoSkill.InputModes, skill.InputModes)
	}

	if len(skill.OutputModes) > 0 {
		protoSkill.OutputModes = make([]string, len(skill.OutputModes))
		copy(protoSkill.OutputModes, skill.OutputModes)
	}

	return protoSkill, nil
}

// FromProtoSkill converts a [a2a_v1.AgentSkill] to a [AgentSkill].
func FromProtoSkill(protoSkill *a2a_v1.AgentSkill) (*AgentSkill, error) {
	if protoSkill == nil {
		return nil, nil
	}

	skill := &AgentSkill{
		ID:          protoSkill.Id,
		Name:        protoSkill.Name,
		Description: protoSkill.Description,
	}

	if len(protoSkill.Examples) > 0 {
		skill.Examples = make([]string, len(protoSkill.Examples))
		copy(skill.Examples, protoSkill.Examples)
	}

	if len(protoSkill.Tags) > 0 {
		skill.Tags = make([]string, len(protoSkill.Tags))
		copy(skill.Tags, protoSkill.Tags)
	}

	if len(protoSkill.InputModes) > 0 {
		skill.InputModes = make([]string, len(protoSkill.InputModes))
		copy(skill.InputModes, protoSkill.InputModes)
	}

	if len(protoSkill.OutputModes) > 0 {
		skill.OutputModes = make([]string, len(protoSkill.OutputModes))
		copy(skill.OutputModes, protoSkill.OutputModes)
	}

	return skill, nil
}

// ToProtoRole converts a [Role] to a [a2a_v1.Role].
func ToProtoRole(role Role) a2a_v1.Role {
	switch role {
	case RoleUser:
		return a2a_v1.Role_ROLE_USER
	case RoleAgent:
		return a2a_v1.Role_ROLE_AGENT
	default:
		return a2a_v1.Role_ROLE_UNSPECIFIED
	}
}

// FromProtoRole converts a [a2a_v1.Role] to a [Role].
func FromProtoRole(protoRole a2a_v1.Role) Role {
	switch protoRole {
	case a2a_v1.Role_ROLE_USER:
		return RoleUser
	case a2a_v1.Role_ROLE_AGENT:
		return RoleAgent
	default:
		return RoleUser // Default to user for unknown roles
	}
}
