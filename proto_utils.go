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

package a2a

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	a2a_v1 "github.com/go-a2a/a2a-grpc/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProtoMessage converts a [Message] to a [a2a_v1.Message].
func ToProtoMessage(msg *Message) (*a2a_v1.Message, error) {
	if msg == nil {
		return nil, nil
	}

	protoMsg := &a2a_v1.Message{
		MessageId: msg.MessageID,
		ContextId: msg.ContextID,
		TaskId:    msg.TaskID,
		Role:      ToProtoRole(msg.Role),
	}

	// Convert string content to a single text part
	protoMsg.Content = make([]*a2a_v1.Part, len(msg.Parts))
	for i, part := range msg.Parts {
		p, err := ToProtoPart(part)
		if err != nil {
			return nil, err
		}
		protoMsg.Content[i] = p
	}

	if msg.Metadata != nil {
		metadata, err := ToProtoMetadata(msg.Metadata)
		if err != nil {
			return nil, fmt.Errorf("convert metadata: %w", err)
		}
		protoMsg.Metadata = metadata
	}

	return protoMsg, nil
}

// ToProtoMetadata converts a Go map[string]any to a [structpb.Struct].
func ToProtoMetadata(metadata map[string]any) (*structpb.Struct, error) {
	if metadata == nil {
		return nil, nil
	}

	protoStruct, err := structpb.NewStruct(metadata)
	if err != nil {
		return nil, fmt.Errorf("create protobuf struct: %w", err)
	}

	return protoStruct, nil
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

	case *FilePart:
		protoFilePart, err := ToProtoFile(p.File)
		if err != nil {
			return nil, err
		}
		protoPart.Part = &a2a_v1.Part_File{
			File: protoFilePart,
		}

	case *DataPart:
		protoData, err := ToProtoData(p.Data)
		if err != nil {
			return nil, err
		}
		protoPart.Part = &a2a_v1.Part_Data{
			Data: protoData,
		}

	default:
		return nil, fmt.Errorf("unsupported part type: %T", part)
	}

	return protoPart, nil
}

// ToProtoData converts a Go map[string]any to a [a2a_v1.DataPart].
func ToProtoData(data map[string]any) (*a2a_v1.DataPart, error) {
	protoData, err := structpb.NewStruct(data)
	if err != nil {
		return nil, fmt.Errorf("convert structpb.Struct: %w", err)
	}
	return &a2a_v1.DataPart{
		Data: protoData,
	}, nil
}

// ToProtoFile converts a [File] to a [a2a_v1.FilePart].
func ToProtoFile(file File) (*a2a_v1.FilePart, error) {
	if file == nil {
		return nil, nil
	}

	protoFilePart := &a2a_v1.FilePart{}

	switch f := file.(type) {
	case *FileWithURI:
		protoFilePart.MimeType = f.GetMIMEType()
		protoFilePart.File = &a2a_v1.FilePart_FileWithUri{
			FileWithUri: f.URI,
		}

	case *FileWithBytes:
		protoFilePart.MimeType = f.GetMIMEType()
		protoFilePart.File = &a2a_v1.FilePart_FileWithBytes{
			FileWithBytes: []byte(f.Bytes),
		}

	default:
		return nil, fmt.Errorf("unknown file type: %T", file)
	}

	return protoFilePart, nil
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
	status, err := ToProtoTaskStatus(task.Status)
	if err != nil {
		return nil, fmt.Errorf("convert task status: %w", err)
	}
	if status != nil {
		protoTask.Status = status
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
				return nil, err
			}
			artifacts[i] = protoArtifact
		}
		protoTask.Artifacts = artifacts
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
				return nil, err
			}
			history[i] = protoMsg
		}
		protoTask.History = history
	}

	return protoTask, nil
}

// ToProtoTaskStatus converts a [TaskStatus] to a [a2a_v1.TaskStatus].
func ToProtoTaskStatus(status TaskStatus) (*a2a_v1.TaskStatus, error) {
	protoState := &a2a_v1.TaskStatus{
		State: ToProtoTaskState(status.State),
	}

	if status.Timestamp != "" {
		t, err := time.Parse(time.RFC3339, status.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("parse %s timestamp", status.Timestamp)
		}
		protoState.Timestamp = timestamppb.New(t)
	}

	return protoState, nil
}

// ToProtoTaskState converts a [TaskState] to a [a2a_v1.TaskState].
func ToProtoTaskState(state TaskState) a2a_v1.TaskState {
	switch state {
	case TaskStateSubmitted:
		return a2a_v1.TaskState_TASK_STATE_SUBMITTED
	case TaskStateWorking:
		return a2a_v1.TaskState_TASK_STATE_WORKING
	case TaskStateInputRequired:
		return a2a_v1.TaskState_TASK_STATE_INPUT_REQUIRED
	case TaskStateCompleted:
		return a2a_v1.TaskState_TASK_STATE_COMPLETED
	case TaskStateCanceled:
		return a2a_v1.TaskState_TASK_STATE_CANCELLED
	case TaskStateFailed:
		return a2a_v1.TaskState_TASK_STATE_FAILED
	case TaskStateRejected:
		return a2a_v1.TaskState_TASK_STATE_REJECTED
	case TaskStateAuthRequired:
		return a2a_v1.TaskState_TASK_STATE_AUTH_REQUIRED
	default:
		return a2a_v1.TaskState_TASK_STATE_UNSPECIFIED
	}
}

// ToProtoArtifact converts a [Artifact] to a [a2a_v1.Artifact].
func ToProtoArtifact(artifact *Artifact) (*a2a_v1.Artifact, error) {
	if artifact == nil {
		return nil, nil
	}

	protoArtifact := &a2a_v1.Artifact{
		ArtifactId:  artifact.ArtifactID,
		Description: artifact.Description,
		Name:        artifact.Name,
		Extensions:  artifact.Extensions,
	}

	// Convert Metadata
	if artifact.Metadata != nil {
		metadata, err := ToProtoMetadata(artifact.Metadata)
		if err != nil {
			return nil, err
		}
		protoArtifact.Metadata = metadata
	}

	// Convert Parts
	if len(artifact.Parts) > 0 {
		parts := make([]*a2a_v1.Part, len(artifact.Parts))
		for i, part := range artifact.Parts {
			if part == nil {
				return nil, fmt.Errorf("part at index %d cannot be nil", i)
			}
			protoPart, err := ToProtoPart(part)
			if err != nil {
				return nil, err
			}
			parts[i] = protoPart
		}
		protoArtifact.Parts = parts
	}

	return protoArtifact, nil
}

// ToProtoAuthenticationInfo converts a [AuthenticationInfo] to a [a2a_v1.AuthenticationInfo].
func ToProtoAuthenticationInfo(auth *PushNotificationAuthenticationInfo) (*a2a_v1.AuthenticationInfo, error) {
	if auth == nil {
		return nil, nil
	}

	protoAuth := &a2a_v1.AuthenticationInfo{
		Schemes:     auth.Schemes,
		Credentials: auth.Credentials,
	}

	return protoAuth, nil
}

// ToProtoPushNotificationConfig converts a [PushNotificationConfig] to a [a2a_v1.PushNotificationConfig].
func ToProtoPushNotificationConfig(config *PushNotificationConfig) (*a2a_v1.PushNotificationConfig, error) {
	if config == nil {
		return nil, nil
	}

	protoNotifyConfig := &a2a_v1.PushNotificationConfig{
		Id:    config.ID,
		Url:   config.URL,
		Token: config.Token,
	}

	if config.Authentication != nil {
		protoAuth, err := ToProtoAuthenticationInfo(config.Authentication)
		if err != nil {
			return nil, err
		}
		protoNotifyConfig.Authentication = protoAuth
	}

	return protoNotifyConfig, nil
}

// ToProtoTaskArtifactUpdateEvent converts a [TaskArtifactUpdateEvent] to a [a2a_v1.TaskArtifactUpdateEvent].
func ToProtoTaskArtifactUpdateEvent(event *TaskArtifactUpdateEvent) (*a2a_v1.TaskArtifactUpdateEvent, error) {
	if event == nil {
		return nil, nil
	}

	protoEvent := &a2a_v1.TaskArtifactUpdateEvent{
		TaskId:    event.TaskID,
		ContextId: event.ContextID,
		Append:    event.Append,
		LastChunk: event.LastChunk,
	}

	if event.Artifact != nil {
		protoArtifact, err := ToProtoArtifact(event.Artifact)
		if err != nil {
			return nil, err
		}
		protoEvent.Artifact = protoArtifact
	}

	return protoEvent, nil
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
	protoStatus, err := ToProtoTaskStatus(event.Status)
	if err != nil {
		return nil, err
	}
	protoEvent.Status = protoStatus

	// Convert Metadata
	if event.Metadata != nil {
		metadata, err := ToProtoMetadata(event.Metadata)
		if err != nil {
			return nil, err
		}
		protoEvent.Metadata = metadata
	}

	return protoEvent, nil
}

// ToProtoMessageSendConfiguration converts a [SendMessageConfiguration] to a [a2a_v1.SendMessageConfiguration].
func ToProtoMessageSendConfiguration(config *MessageSendConfiguration) (*a2a_v1.SendMessageConfiguration, error) {
	if config == nil {
		return &a2a_v1.SendMessageConfiguration{}, nil
	}

	protoSmc := &a2a_v1.SendMessageConfiguration{
		AcceptedOutputModes: config.AcceptedOutputModes,
		HistoryLength:       int32(config.HistoryLength),
		Blocking:            config.Blocking,
	}

	if config.PushNotificationConfig != nil {
		protoNotifyConfig, err := ToProtoPushNotificationConfig(config.PushNotificationConfig)
		if err != nil {
			return nil, fmt.Errorf("convert push notification config: %w", err)
		}
		protoSmc.PushNotification = protoNotifyConfig
	}

	return protoSmc, nil
}

// ToProtoUpdateEvent converts various update event types to appropriate Protocol Buffer types.
//
// This function handles multiple types of update events that can occur in the A2A protocol.
func ToProtoUpdateEvent(event SendStreamingMessageResponse) (*a2a_v1.StreamResponse, error) {
	if event == nil {
		return nil, nil
	}

	response := &a2a_v1.StreamResponse{}

	switch e := event.(type) {
	case *TaskStatusUpdateEvent:
		payload, err := ToProtoTaskStatusUpdateEvent(e)
		if err != nil {
			return nil, fmt.Errorf("convert TaskStatusUpdateEvent: %w", err)
		}
		response.Payload = &a2a_v1.StreamResponse_StatusUpdate{
			StatusUpdate: payload,
		}
	case *TaskArtifactUpdateEvent:
		payload, err := ToProtoTaskArtifactUpdateEvent(e)
		if err != nil {
			return nil, fmt.Errorf("convert TaskArtifactUpdateEvent: %w", err)
		}
		response.Payload = &a2a_v1.StreamResponse_ArtifactUpdate{
			ArtifactUpdate: payload,
		}
	case *Message:
		payload, err := ToProtoMessage(e)
		if err != nil {
			return nil, fmt.Errorf("convert TaskArtifactUpdateEvent: %w", err)
		}
		response.Payload = &a2a_v1.StreamResponse_Msg{
			Msg: payload,
		}
	case *Task:
		payload, err := ToProtoTask(e)
		if err != nil {
			return nil, fmt.Errorf("convert TaskArtifactUpdateEvent: %w", err)
		}
		response.Payload = &a2a_v1.StreamResponse_Task{
			Task: payload,
		}

	default:
		return nil, fmt.Errorf("unsupported event type: %T", event)
	}

	return response, nil
}

// ToProtoTaskOrMessage converts a Task or Message to the appropriate Protocol Buffer type.
//
// This function handles the polymorphic conversion between Task and Message types.
func ToProtoTaskOrMessage(event MessageOrTask) (*a2a_v1.SendMessageResponse, error) {
	if event == nil {
		return nil, nil
	}

	switch e := event.(type) {
	case *Message:
		payload, err := ToProtoMessage(e)
		if err != nil {
			return nil, fmt.Errorf("convert Message: %w", err)
		}
		return &a2a_v1.SendMessageResponse{
			Payload: &a2a_v1.SendMessageResponse_Msg{
				Msg: payload,
			},
		}, nil
	case *Task:
		payload, err := ToProtoTask(e)
		if err != nil {
			return nil, fmt.Errorf("convert Message: %w", err)
		}
		return &a2a_v1.SendMessageResponse{
			Payload: &a2a_v1.SendMessageResponse_Task{
				Task: payload,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown task or message type: %T", event)
	}
}

// ToProtoStreamResponse converts various types to a [a2a_v1.StreamResponse].
//
// This mirrors the Python implementation's polymorphic update_event method.
func ToProtoStreamResponse(event SendStreamingMessageResponse) (*a2a_v1.StreamResponse, error) {
	if event == nil {
		return nil, nil
	}

	response := &a2a_v1.StreamResponse{}

	switch p := event.(type) {
	case *Message:
		protoMessage, err := ToProtoMessage(p)
		if err != nil {
			return nil, fmt.Errorf("convert message: %w", err)
		}
		response.Payload = &a2a_v1.StreamResponse_Msg{
			Msg: protoMessage,
		}
	case *Task:
		protoTask, err := ToProtoTask(p)
		if err != nil {
			return nil, fmt.Errorf("convert task: %w", err)
		}
		response.Payload = &a2a_v1.StreamResponse_Task{
			Task: protoTask,
		}
	case *TaskStatusUpdateEvent:
		protoEvent, err := ToProtoTaskStatusUpdateEvent(p)
		if err != nil {
			return nil, fmt.Errorf("convert task status update event: %w", err)
		}
		response.Payload = &a2a_v1.StreamResponse_StatusUpdate{
			StatusUpdate: protoEvent,
		}
	case *TaskArtifactUpdateEvent:
		protoEvent, err := ToProtoTaskArtifactUpdateEvent(p)
		if err != nil {
			return nil, fmt.Errorf("convert task artifact update event: %w", err)
		}
		response.Payload = &a2a_v1.StreamResponse_ArtifactUpdate{
			ArtifactUpdate: protoEvent,
		}
	default:
		return nil, fmt.Errorf("unknown payload type for stream response: %T", event)
	}

	return response, nil
}

// ToProtoTaskPushNotificationConfig converts a [TaskPushNotificationConfig] to a [a2a_v1.TaskPushNotificationConfig].
func ToProtoTaskPushNotificationConfig(config *TaskPushNotificationConfig) (*a2a_v1.TaskPushNotificationConfig, error) {
	if config == nil {
		return nil, nil
	}

	protoTpnc := &a2a_v1.TaskPushNotificationConfig{
		Name: fmt.Sprintf("tasks/%[1]s/pushNotificationConfigs/%[1]s", config.TaskID),
	}

	if config.PushNotificationConfig != nil {
		protoConfig, err := ToProtoPushNotificationConfig(config.PushNotificationConfig)
		if err != nil {
			return nil, fmt.Errorf("convert push notification config: %w", err)
		}
		protoTpnc.PushNotificationConfig = protoConfig
	}

	return protoTpnc, nil
}

// ToProtoAgentCard converts a [AgentCard] to a [a2a_v1.AgentCard].
func ToProtoAgentCard(card *AgentCard) (*a2a_v1.AgentCard, error) {
	if card == nil {
		return nil, nil
	}

	protoCard := &a2a_v1.AgentCard{
		DefaultInputModes:                 card.DefaultInputModes,
		DefaultOutputModes:                card.DefaultOutputModes,
		Description:                       card.Description,
		DocumentationUrl:                  card.DocumentationURL,
		Name:                              card.Name,
		Url:                               card.URL,
		Version:                           card.Version,
		SupportsAuthenticatedExtendedCard: card.SupportsAuthenticatedExtendedCard,
		ProtocolVersion:                   card.ProtocolVersion,
		PreferredTransport:                string(card.PreferredTransport),
	}

	// Convert Capabilities
	if card.Capabilities != nil {
		protoCapabilities, err := ToProtoAgentCapabilities(card.Capabilities)
		if err != nil {
			return nil, fmt.Errorf("convert capabilities: %w", err)
		}
		protoCard.Capabilities = protoCapabilities
	}

	// Convert Provider
	if card.Provider != nil {
		protoProvider, err := ToProtoProvider(card.Provider)
		if err != nil {
			return nil, fmt.Errorf("convert provider: %w", err)
		}
		protoCard.Provider = protoProvider
	}

	if len(card.Security) > 0 {
		protoSecurity, err := ToProtoSecurity(card.Security)
		if err != nil {
			return nil, fmt.Errorf("convert security: %w", err)
		}
		protoCard.Security = protoSecurity
	}

	// Convert SecuritySchemes
	if len(card.SecuritySchemes) > 0 {
		protoSecuritySchemes := make(map[string]*a2a_v1.SecurityScheme)
		for name, scheme := range card.SecuritySchemes {
			protoScheme, err := ToProtoSecurityScheme(scheme)
			if err != nil {
				return nil, fmt.Errorf("convert security scheme %s: %w", name, err)
			}
			protoSecuritySchemes[name] = protoScheme
		}
		protoCard.SecuritySchemes = protoSecuritySchemes
	}

	// Convert Skills
	if len(card.Skills) > 0 {
		protoSkills := make([]*a2a_v1.AgentSkill, len(card.Skills))
		for i, skill := range card.Skills {
			protoSkill, err := ToProtoSkill(skill)
			if err != nil {
				return nil, fmt.Errorf("convert skill at index %d: %w", i, err)
			}
			protoSkills[i] = protoSkill
		}
		protoCard.Skills = protoSkills
	}

	return protoCard, nil
}

// ToProtoAgentCapabilities converts a [AgentCapabilities] to a [a2a_v1.AgentCapabilities].
func ToProtoAgentCapabilities(capabilities *AgentCapabilities) (*a2a_v1.AgentCapabilities, error) {
	if capabilities == nil {
		return nil, nil
	}

	protoCapabilities := &a2a_v1.AgentCapabilities{
		Streaming:         capabilities.Streaming,
		PushNotifications: capabilities.PushNotifications,
	}

	if len(capabilities.Extensions) > 0 {
		protoExtensions := make([]*a2a_v1.AgentExtension, len(capabilities.Extensions))
		for i, ext := range capabilities.Extensions {
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

// ToProtoProvider converts a Go []map[string][]string to a slice of [a2a_v1.Security].
func ToProtoSecurity(security []map[string]SecurityScopes) ([]*a2a_v1.Security, error) {
	if security == nil {
		return nil, nil
	}

	protoSecurities := []*a2a_v1.Security{}
	for _, m := range security {
		schemes := make(map[string]*a2a_v1.StringList)
		for k, v := range m {
			schemes[k] = &a2a_v1.StringList{
				List: v.AsSlices(),
			}
		}
		protoSecurities = append(protoSecurities, &a2a_v1.Security{
			Schemes: schemes,
		})
	}

	return protoSecurities, nil
}

// ToProtoSecuritySchemes converts a Go map of [SecuritySchemes] to a map of [a2a_v1.SecurityScheme].
func ToProtoSecuritySchemes(schemes map[string]SecurityScheme) (map[string]*a2a_v1.SecurityScheme, error) {
	if schemes == nil {
		return nil, nil
	}

	protoSchemes := make(map[string]*a2a_v1.SecurityScheme)
	for name, scheme := range schemes {
		protoScheme, err := ToProtoSecurityScheme(scheme)
		if err != nil {
			return nil, fmt.Errorf("convert security scheme %s: %w", name, err)
		}
		protoSchemes[name] = protoScheme
	}

	return protoSchemes, nil
}

// ToProtoSecurityScheme converts a [SecurityScheme] to a [a2a_v1.SecurityScheme].
func ToProtoSecurityScheme(scheme SecurityScheme) (*a2a_v1.SecurityScheme, error) {
	if scheme == nil {
		return nil, nil
	}

	protoScheme := &a2a_v1.SecurityScheme{}

	switch scheme := scheme.(type) {
	case *APIKeySecurityScheme:
		protoAPIKey := &a2a_v1.APIKeySecurityScheme{
			Description: scheme.Description,
			Location:    string(scheme.In),
			Name:        scheme.Name,
		}
		protoScheme.Scheme = &a2a_v1.SecurityScheme_ApiKeySecurityScheme{
			ApiKeySecurityScheme: protoAPIKey,
		}

	case *HTTPAuthSecurityScheme:
		protoHTTPAuth := &a2a_v1.HTTPAuthSecurityScheme{
			Description:  scheme.Description,
			Scheme:       scheme.Scheme,
			BearerFormat: scheme.BearerFormat,
		}
		protoScheme.Scheme = &a2a_v1.SecurityScheme_HttpAuthSecurityScheme{
			HttpAuthSecurityScheme: protoHTTPAuth,
		}

	case *OAuth2SecurityScheme:
		protoOAuth2 := &a2a_v1.OAuth2SecurityScheme{
			Description: scheme.Description,
		}
		if scheme.Flows != nil {
			protoFlows, err := ToProtoOAuthFlows(scheme.Flows)
			if err != nil {
				return nil, fmt.Errorf("convert OAuth flows: %w", err)
			}
			protoOAuth2.Flows = protoFlows
		}
		protoScheme.Scheme = &a2a_v1.SecurityScheme_Oauth2SecurityScheme{
			Oauth2SecurityScheme: protoOAuth2,
		}

	case *OpenIDConnectSecurityScheme:
		protoOpenIDConnect := &a2a_v1.OpenIdConnectSecurityScheme{
			Description:      scheme.Description,
			OpenIdConnectUrl: scheme.OpenIDConnectURL,
		}
		protoScheme.Scheme = &a2a_v1.SecurityScheme_OpenIdConnectSecurityScheme{
			OpenIdConnectSecurityScheme: protoOpenIDConnect,
		}

	default:
		return nil, fmt.Errorf("security scheme has no valid scheme type")
	}

	return protoScheme, nil
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
			RefreshUrl:       flows.AuthorizationCode.RefreshURL,
			Scopes:           flows.AuthorizationCode.Scopes,
			TokenUrl:         flows.AuthorizationCode.TokenURL,
		}
		protoFlows.Flow = &a2a_v1.OAuthFlows_AuthorizationCode{
			AuthorizationCode: protoAuthCode,
		}
	} else if flows.ClientCredentials != nil {
		protoClientCreds := &a2a_v1.ClientCredentialsOAuthFlow{
			RefreshUrl: flows.ClientCredentials.RefreshURL,
			Scopes:     flows.ClientCredentials.Scopes,
			TokenUrl:   flows.ClientCredentials.TokenURL,
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
			RefreshUrl: flows.Password.RefreshURL,
			Scopes:     flows.Password.Scopes,
			TokenUrl:   flows.Password.TokenURL,
		}
		protoFlows.Flow = &a2a_v1.OAuthFlows_Password{
			Password: protoPassword,
		}
	}

	return protoFlows, nil
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
		Tags:        skill.Tags,
		Examples:    skill.Examples,
		InputModes:  skill.InputModes,
		OutputModes: skill.InputModes,
	}

	return protoSkill, nil
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

// FromProtoMessage converts a [a2a_v1.Message] to a [Message].
func FromProtoMessage(protoMsg *a2a_v1.Message) (*Message, error) {
	if protoMsg == nil {
		return nil, nil
	}

	msg := &Message{
		MessageID: protoMsg.GetMessageId(),
		ContextID: protoMsg.GetContextId(),
		TaskID:    protoMsg.GetTaskId(),
		Kind:      MessageEventKind,
		Role:      FromProtoRole(protoMsg.GetRole()),
	}

	// Convert parts to string content (simplified approach - take first text part)
	if protoContent := protoMsg.GetContent(); len(protoContent) > 0 {
		parts := make([]Part, len(protoContent))
		for i, protoPart := range protoContent {
			part, err := FromProtoPart(protoPart)
			if err != nil {
				return nil, fmt.Errorf("convert Protobuf part: %w", err)
			}
			parts[i] = part
		}
		msg.Parts = parts
	}

	if protoMetadata := protoMsg.GetMetadata(); protoMetadata != nil {
		metadata, err := FromProtoMetadata(protoMetadata)
		if err != nil {
			return nil, fmt.Errorf("convert metadata: %w", err)
		}
		msg.Metadata = metadata
	}

	return msg, nil
}

// FromProtoMetadata converts a [structpb.Struct] to a Go map[string]any.
func FromProtoMetadata(protoStruct *structpb.Struct) (map[string]any, error) {
	if protoStruct == nil {
		return nil, nil
	}

	return protoStruct.AsMap(), nil
}

// FromProtoPart converts a [a2a_v1.Part] to a [Part].
func FromProtoPart(protoPart *a2a_v1.Part) (Part, error) {
	if protoPart == nil {
		return nil, nil
	}

	switch p := protoPart.GetPart().(type) {
	case *a2a_v1.Part_Text:
		return NewTextPart(p.Text), nil
	case *a2a_v1.Part_Data:
		// Convert protobuf DataPart to map[string]any
		data, err := FromProtoData(p.Data)
		if err != nil {
			return nil, fmt.Errorf("convert data part: %w", err)
		}
		return NewDataPart(data), nil
	case *a2a_v1.Part_File:
		file, err := FromProtoFile(p.File)
		if err != nil {
			return nil, fmt.Errorf("convert file part: %w", err)
		}
		return NewFilePart(file), nil
	default:
		return nil, fmt.Errorf("unsupported part type: %T", protoPart.Part)
	}
}

// FromProtoData converts a [a2a_v1.DataPart] to a Go map[string]any.
func FromProtoData(protoData *a2a_v1.DataPart) (map[string]any, error) {
	data := protoData.GetData()
	if data == nil {
		return nil, fmt.Errorf("data part is nil")
	}
	return data.AsMap(), nil
}

// FromProtoFile converts a [a2a_v1.FilePart] to a [File].
func FromProtoFile(protoFilePart *a2a_v1.FilePart) (File, error) {
	if protoFilePart == nil {
		return nil, nil
	}

	switch f := protoFilePart.GetFile().(type) {
	case *a2a_v1.FilePart_FileWithUri:
		return &FileWithURI{
			MIMEType: protoFilePart.GetMimeType(),
			URI:      f.FileWithUri,
		}, nil
	case *a2a_v1.FilePart_FileWithBytes:
		return &FileWithBytes{
			MIMEType: protoFilePart.GetMimeType(),
			Bytes:    string(f.FileWithBytes),
		}, nil
	default:
		return nil, fmt.Errorf("unknown proto file type: %T", protoFilePart.File)
	}
}

// FromProtoTask converts a [a2a_v1.Task] to a [Task].
func FromProtoTask(protoTask *a2a_v1.Task) (*Task, error) {
	if protoTask == nil {
		return nil, nil
	}

	task := &Task{
		ID:        protoTask.GetId(),
		ContextID: protoTask.GetContextId(),
	}

	// Convert TaskStatus
	if protoStatus := protoTask.GetStatus(); protoStatus != nil {
		status, err := FromProtoTaskStatus(protoStatus)
		if err != nil {
			return nil, fmt.Errorf("convert task status: %w", err)
		}
		task.Status = status
	}

	// Convert History
	if protoHistory := protoTask.GetHistory(); len(protoHistory) > 0 {
		history := make([]*Message, len(protoHistory))
		for i, protoMsg := range protoHistory {
			if protoMsg == nil {
				return nil, fmt.Errorf("history message at index %d cannot be nil", i)
			}
			msg, err := FromProtoMessage(protoMsg)
			if err != nil {
				return nil, fmt.Errorf("convert history message at index %d: %w", i, err)
			}
			history[i] = msg
		}
		task.History = history
	}

	// Convert Artifacts
	if protoArtifacts := protoTask.GetArtifacts(); len(protoArtifacts) > 0 {
		artifacts := make([]*Artifact, len(protoArtifacts))
		for i, protoArtifact := range protoArtifacts {
			if protoArtifact == nil {
				return nil, fmt.Errorf("artifact at index %d cannot be nil", i)
			}
			artifact, err := FromProtoArtifact(protoArtifact)
			if err != nil {
				return nil, fmt.Errorf("convert artifact at index %d: %w", i, err)
			}
			artifacts[i] = artifact
		}
		task.Artifacts = artifacts
	}

	return task, nil
}

// FromProtoTaskStatus converts a [a2a_v1.TaskStatus] to a [TaskStatus].
func FromProtoTaskStatus(protoStatus *a2a_v1.TaskStatus) (TaskStatus, error) {
	if protoStatus == nil {
		return TaskStatus{}, nil
	}

	state := FromProtoTaskState(protoStatus.GetState())
	return TaskStatus{
		State: state,
	}, nil
}

// FromProtoTaskState converts a [a2a_v1.TaskState] to a [TaskStatus].
func FromProtoTaskState(protoState a2a_v1.TaskState) TaskState {
	switch protoState {
	case a2a_v1.TaskState_TASK_STATE_SUBMITTED:
		return TaskStateSubmitted
	case a2a_v1.TaskState_TASK_STATE_WORKING:
		return TaskStateWorking
	case a2a_v1.TaskState_TASK_STATE_COMPLETED:
		return TaskStateCompleted
	case a2a_v1.TaskState_TASK_STATE_FAILED:
		return TaskStateFailed
	case a2a_v1.TaskState_TASK_STATE_CANCELLED:
		return TaskStateCanceled
	case a2a_v1.TaskState_TASK_STATE_REJECTED:
		return TaskStateRejected
	case a2a_v1.TaskState_TASK_STATE_INPUT_REQUIRED:
		return TaskStateInputRequired
	case a2a_v1.TaskState_TASK_STATE_AUTH_REQUIRED:
		return TaskStateAuthRequired
	default:
		return TaskStateUnknown // Default to submitted for unknown states
	}
}

// FromProtoArtifact converts a [a2a_v1.Artifact] to a [Artifact].
func FromProtoArtifact(protoArtifact *a2a_v1.Artifact) (*Artifact, error) {
	if protoArtifact == nil {
		return nil, nil
	}

	artifact := &Artifact{
		ArtifactID:  protoArtifact.GetArtifactId(),
		Description: protoArtifact.GetDescription(),
		Name:        protoArtifact.GetName(),
		Extensions:  protoArtifact.GetExtensions(),
	}

	// Convert Metadata
	if protoMetadata := protoArtifact.GetMetadata(); protoMetadata != nil {
		metadata, err := FromProtoMetadata(protoMetadata)
		if err != nil {
			return nil, fmt.Errorf("convert metadata: %w", err)
		}
		artifact.Metadata = metadata
	}

	// Convert Parts
	if protoParts := protoArtifact.GetParts(); len(protoParts) > 0 {
		parts := make([]Part, len(protoParts))
		for i, protoPart := range protoParts {
			if protoPart == nil {
				return nil, fmt.Errorf("proto part at index %d cannot be nil", i)
			}
			part, err := FromProtoPart(protoPart)
			if err != nil {
				return nil, fmt.Errorf("convert part at index %d: %w", i, err)
			}
			parts[i] = part
		}
		artifact.Parts = parts
	}

	return artifact, nil
}

// FromProtoTaskArtifactUpdateEvent converts a [a2a_v1.TaskArtifactUpdateEvent] to a [TaskArtifactUpdateEvent].
func FromProtoTaskArtifactUpdateEvent(protoEvent *a2a_v1.TaskArtifactUpdateEvent) (*TaskArtifactUpdateEvent, error) {
	if protoEvent == nil {
		return nil, nil
	}

	event := &TaskArtifactUpdateEvent{
		Append:    protoEvent.GetAppend(),
		ContextID: protoEvent.GetContextId(),
		Kind:      ArtifactUpdateEventKind,
		LastChunk: protoEvent.GetLastChunk(),
		Metadata:  protoEvent.GetMetadata().AsMap(),
		TaskID:    protoEvent.GetTaskId(),
	}

	if protoArtifact := protoEvent.GetArtifact(); protoArtifact != nil {
		artifact, err := FromProtoArtifact(protoArtifact)
		if err != nil {
			return nil, fmt.Errorf("convert artifact: %w", err)
		}
		event.Artifact = artifact
	}

	return event, nil
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
		Kind:      StatusUpdateEventKind,
	}

	// Convert TaskStatus
	if protoEvent.Status != nil {
		status, err := FromProtoTaskStatus(protoEvent.Status)
		if err != nil {
			return nil, fmt.Errorf("convert task status: %w", err)
		}
		event.Status = status
	}

	// Convert Metadata
	if protoEvent.Metadata != nil {
		metadata, err := FromProtoMetadata(protoEvent.Metadata)
		if err != nil {
			return nil, fmt.Errorf("convert metadata: %w", err)
		}
		event.Metadata = metadata
	}

	return event, nil
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
			return nil, fmt.Errorf("convert authentication info: %w", err)
		}
		pnc.Authentication = auth
	}

	return pnc, nil
}

// FromProtoAuthenticationInfo converts a [a2a_v1.AuthenticationInfo] to a [AuthenticationInfo].
func FromProtoAuthenticationInfo(protoAuth *a2a_v1.AuthenticationInfo) (*PushNotificationAuthenticationInfo, error) {
	if protoAuth == nil {
		return nil, nil
	}

	auth := &PushNotificationAuthenticationInfo{
		Credentials: protoAuth.Credentials,
		Schemes:     protoAuth.GetSchemes(),
	}

	return auth, nil
}

// FromProtoSendMessageConfiguration converts a [a2a_v1.SendMessageConfiguration] to a [SendMessageConfiguration].
func FromProtoSendMessageConfiguration(config *a2a_v1.SendMessageConfiguration) (*MessageSendConfiguration, error) {
	if config == nil {
		return nil, nil
	}

	smc := &MessageSendConfiguration{
		AcceptedOutputModes: config.GetAcceptedOutputModes(),
		HistoryLength:       int(config.GetHistoryLength()),
		Blocking:            config.GetBlocking(),
	}

	if protoNotify := config.GetPushNotification(); protoNotify != nil {
		notifyConfig, err := FromProtoPushNotificationConfig(protoNotify)
		if err != nil {
			return nil, fmt.Errorf("convert push notification config: %w", err)
		}
		smc.PushNotificationConfig = notifyConfig
	}

	return smc, nil
}

// FromProtoPushNotificationConfig converts a [a2a_v1.SendMessageRequest] to a [MessageSendParams].
func FromProtoMessageSendParams(request *a2a_v1.SendMessageRequest) (*MessageSendParams, error) {
	if request == nil {
		return nil, nil
	}

	configuration, err := FromProtoSendMessageConfiguration(request.GetConfiguration())
	if err != nil {
		return nil, err
	}
	message, err := FromProtoMessage(request.GetRequest())
	if err != nil {
		return nil, err
	}
	metadata, err := FromProtoMetadata(request.GetMetadata())
	if err != nil {
		return nil, err
	}

	return &MessageSendParams{
		Configuration: configuration,
		Message:       message,
		Metadata:      metadata,
	}, nil
}

var (
	reTaskPushConfigName = regexp.MustCompile(`tasks/(\w+)/pushNotificationConfigs/(\w+)`)
	reTaskName           = regexp.MustCompile(`tasks/(\w+)`)
)

// FromProtoTaskIDParams converts a [a2a_v1.CancelTaskRequest], [a2a_v1.TaskSubscriptionRequest], or [a2a_v1.GetTaskPushNotificationConfigRequest] to a [TaskIDParams].
func FromProtoTaskIDParams[R *a2a_v1.CancelTaskRequest | *a2a_v1.TaskSubscriptionRequest | *a2a_v1.GetTaskPushNotificationConfigRequest](request R) (*TaskIDParams, error) {
	if request == nil {
		return nil, nil
	}

	// This is currently incomplete until the core sdk supports multiple
	// configs for a single task.
	switch request := any(request).(type) {
	case *a2a_v1.GetTaskPushNotificationConfigRequest:
		m := reTaskPushConfigName.FindStringSubmatch(request.GetName())
		if len(m) == 0 {
			return nil, fmt.Errorf("%w: no task for %s", ErrInvalidParams, request.GetName())
		}
		return &TaskIDParams{
			ID: m[1],
		}, nil

	case *a2a_v1.CancelTaskRequest:
		m := reTaskName.FindStringSubmatch(request.GetName())
		if len(m) == 0 {
			return nil, fmt.Errorf("%w: no task for %s", ErrInvalidParams, request.GetName())
		}
		return &TaskIDParams{
			ID: m[1],
		}, nil

	case *a2a_v1.TaskSubscriptionRequest:
		m := reTaskName.FindStringSubmatch(request.GetName())
		if len(m) == 0 {
			return nil, fmt.Errorf("%w: no task for %s", ErrInvalidParams, request.GetName())
		}
		return &TaskIDParams{
			ID: m[1],
		}, nil
	}

	return nil, errors.New("unreachable")
}

// FromProtoTaskPushNotificationConfig converts a [a2a_v1.CreateTaskPushNotificationConfigRequest] to a [TaskPushNotificationConfig].
func FromProtoTaskPushNotificationConfig(request *a2a_v1.CreateTaskPushNotificationConfigRequest) (*TaskPushNotificationConfig, error) {
	if request == nil {
		return nil, nil
	}

	m := reTaskName.FindStringSubmatch(request.GetParent())
	if len(m) == 0 {
		return nil, fmt.Errorf("%w: no task for %s", ErrInvalidParams, request.GetParent())
	}

	response := &TaskPushNotificationConfig{
		TaskID: m[1],
	}

	if notifyConfig := request.GetConfig().GetPushNotificationConfig(); notifyConfig != nil {
		config, err := FromProtoPushNotificationConfig(notifyConfig)
		if err != nil {
			return nil, fmt.Errorf("convert push notification config: %w", err)
		}
		response.PushNotificationConfig = config
	}

	return response, nil
}

// FromProtoAgentCard converts a [a2a_v1.AgentCard] to a [AgentCard].
func FromProtoAgentCard(protoCard *a2a_v1.AgentCard) (*AgentCard, error) {
	if protoCard == nil {
		return nil, nil
	}

	agentCard := &AgentCard{
		DefaultInputModes:                 protoCard.GetDefaultInputModes(),
		DefaultOutputModes:                protoCard.GetDefaultOutputModes(),
		Description:                       protoCard.GetDescription(),
		DocumentationURL:                  protoCard.GetDocumentationUrl(),
		Name:                              protoCard.GetName(),
		URL:                               protoCard.GetUrl(),
		Version:                           protoCard.GetVersion(),
		SupportsAuthenticatedExtendedCard: protoCard.GetSupportsAuthenticatedExtendedCard(),
		ProtocolVersion:                   protoCard.GetProtocolVersion(),
		PreferredTransport:                TransportProtocol(protoCard.GetPreferredTransport()),
	}

	// Convert Capabilities
	if protoCapabilities := protoCard.GetCapabilities(); protoCapabilities != nil {
		capabilities, err := FromProtoAgentCapabilities(protoCapabilities)
		if err != nil {
			return nil, fmt.Errorf("convert capabilities: %w", err)
		}
		agentCard.Capabilities = capabilities
	}

	// Convert Provider
	if protoProvider := protoCard.GetProvider(); protoProvider != nil {
		provider, err := FromProtoProvider(protoProvider)
		if err != nil {
			return nil, fmt.Errorf("convert provider: %w", err)
		}
		agentCard.Provider = provider
	}

	// Convert Security
	if protoSecurity := protoCard.GetSecurity(); protoSecurity != nil {
		security, err := FromProtoSecurity(protoSecurity)
		if err != nil {
			return nil, fmt.Errorf("convert provider: %w", err)
		}
		agentCard.Security = security
	}

	// Convert SecuritySchemes
	if protoSecuritySchemes := protoCard.GetSecuritySchemes(); len(protoSecuritySchemes) > 0 {
		securitySchemes := make(map[string]SecurityScheme)
		for name, protoScheme := range protoSecuritySchemes {
			scheme, err := FromProtoSecurityScheme(protoScheme)
			if err != nil {
				return nil, fmt.Errorf("convert security scheme %s: %w", name, err)
			}
			securitySchemes[name] = scheme
		}
		agentCard.SecuritySchemes = securitySchemes
	}

	// Convert Skills
	if protoSkills := protoCard.GetSkills(); len(protoSkills) > 0 {
		skills := make([]*AgentSkill, len(protoSkills))
		for i, protoSkill := range protoSkills {
			skill, err := FromProtoSkill(protoSkill)
			if err != nil {
				return nil, fmt.Errorf("convert skill at index %d: %w", i, err)
			}
			skills[i] = skill
		}
		agentCard.Skills = skills
	}

	return agentCard, nil
}

// FromProtoTaskQueryParams converts a [a2a_v1.GetTaskRequest] to a [TaskQueryParams].
func FromProtoTaskQueryParams(request *a2a_v1.GetTaskRequest) (*TaskQueryParams, error) {
	if request == nil {
		return nil, nil
	}

	m := reTaskName.FindStringSubmatch(request.GetName())
	if len(m) == 0 {
		return nil, fmt.Errorf("no task for %s", request.GetName())
	}

	return &TaskQueryParams{
		HistoryLength: int(request.GetHistoryLength()),
		ID:            m[1],
	}, nil
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

	if protoExtensions := protoCapabilities.GetExtensions(); len(protoExtensions) > 0 {
		extensions := make([]*AgentExtension, len(protoExtensions))
		for i, protoExt := range protoExtensions {
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

// FromProtoSecurity converts a slice of [a2a_v1.Security] to a Go []map[string]SecurityScopes.
func FromProtoSecurity(protoSecurity []*a2a_v1.Security) ([]map[string]SecurityScopes, error) {
	if protoSecurity == nil {
		return nil, nil
	}

	security := []map[string]SecurityScopes{}
	for _, s := range protoSecurity {
		m := map[string]SecurityScopes{}
		for k, v := range s.GetSchemes() {
			m[k] = NewSecurityScopes(v.List...)
		}
		security = append(security, m)
	}

	return security, nil
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

// FromProtoSecuritySchemes converts a map of [a2a_v1.SecurityScheme] to a Go map of [SecuritySchemes].
func FromProtoSecuritySchemes(protoSchemes map[string]*a2a_v1.SecurityScheme) (map[string]SecurityScheme, error) {
	if protoSchemes == nil {
		return nil, nil
	}

	schemes := make(map[string]SecurityScheme)
	for name, protoScheme := range protoSchemes {
		scheme, err := FromProtoSecurityScheme(protoScheme)
		if err != nil {
			return nil, fmt.Errorf("convert security scheme %s: %w", name, err)
		}
		schemes[name] = scheme
	}

	return schemes, nil
}

// FromProtoSecurityScheme converts a [a2a_v1.SecurityScheme] to a [SecurityScheme].
func FromProtoSecurityScheme(protoScheme *a2a_v1.SecurityScheme) (SecurityScheme, error) {
	if protoScheme == nil {
		return nil, nil
	}

	switch protoScheme := protoScheme.Scheme.(type) {
	case *a2a_v1.SecurityScheme_ApiKeySecurityScheme:
		return &APIKeySecurityScheme{
			Description: protoScheme.ApiKeySecurityScheme.GetDescription(),
			Name:        protoScheme.ApiKeySecurityScheme.GetName(),
			In:          In(protoScheme.ApiKeySecurityScheme.GetLocation()),
			Type:        APIKeySecuritySchemeType,
		}, nil
	case *a2a_v1.SecurityScheme_HttpAuthSecurityScheme:
		return &HTTPAuthSecurityScheme{
			Description:  protoScheme.HttpAuthSecurityScheme.GetDescription(),
			Scheme:       protoScheme.HttpAuthSecurityScheme.GetScheme(),
			BearerFormat: protoScheme.HttpAuthSecurityScheme.GetBearerFormat(),
			Type:         HTTPSecuritySchemeType,
		}, nil
	case *a2a_v1.SecurityScheme_Oauth2SecurityScheme:
		scheme := &OAuth2SecurityScheme{
			Description: protoScheme.Oauth2SecurityScheme.Description,
			Type:        OAuth2SecuritySchemeType,
		}
		if protoFlow := protoScheme.Oauth2SecurityScheme.GetFlows(); protoFlow != nil {
			flows, err := FromProtoOAuthFlows(protoFlow)
			if err != nil {
				return nil, fmt.Errorf("convert OAuth flows: %w", err)
			}
			scheme.Flows = flows
		}
		return scheme, nil
	case *a2a_v1.SecurityScheme_OpenIdConnectSecurityScheme:
		return &OpenIDConnectSecurityScheme{
			Description:      protoScheme.OpenIdConnectSecurityScheme.GetDescription(),
			OpenIDConnectURL: protoScheme.OpenIdConnectSecurityScheme.GetOpenIdConnectUrl(),
			Type:             OpenIDConnectSecuritySchemeType,
		}, nil
	default:
		return nil, fmt.Errorf("unknown security scheme type: %T", protoScheme)
	}
}

// FromProtoOAuthFlows converts a [a2a_v1.OAuthFlows] to a [OAuthFlows].
func FromProtoOAuthFlows(protoFlows *a2a_v1.OAuthFlows) (*OAuthFlows, error) {
	if protoFlows == nil {
		return nil, nil
	}

	flows := &OAuthFlows{}

	switch protoFlow := protoFlows.Flow.(type) {
	case *a2a_v1.OAuthFlows_AuthorizationCode:
		flows.AuthorizationCode = &AuthorizationCodeOAuthFlow{
			AuthorizationURL: protoFlow.AuthorizationCode.GetAuthorizationUrl(),
			TokenURL:         protoFlow.AuthorizationCode.GetTokenUrl(),
			RefreshURL:       protoFlow.AuthorizationCode.GetRefreshUrl(),
			Scopes:           protoFlow.AuthorizationCode.GetScopes(),
		}
	case *a2a_v1.OAuthFlows_ClientCredentials:
		flows.ClientCredentials = &ClientCredentialsOAuthFlow{
			RefreshURL: protoFlow.ClientCredentials.GetRefreshUrl(),
			Scopes:     protoFlow.ClientCredentials.GetScopes(),
			TokenURL:   protoFlow.ClientCredentials.GetTokenUrl(),
		}
	case *a2a_v1.OAuthFlows_Implicit:
		flows.Implicit = &ImplicitOAuthFlow{
			AuthorizationURL: protoFlow.Implicit.GetAuthorizationUrl(),
			RefreshURL:       protoFlow.Implicit.GetRefreshUrl(),
			Scopes:           protoFlow.Implicit.GetScopes(),
		}
	case *a2a_v1.OAuthFlows_Password:
		flows.Password = &PasswordOAuthFlow{
			RefreshURL: protoFlow.Password.GetRefreshUrl(),
			Scopes:     protoFlow.Password.GetScopes(),
			TokenURL:   protoFlow.Password.GetTokenUrl(),
		}
	}

	return flows, nil
}

// FromProtoSkill converts a [a2a_v1.AgentSkill] to a [AgentSkill].
func FromProtoSkill(protoSkill *a2a_v1.AgentSkill) (*AgentSkill, error) {
	if protoSkill == nil {
		return nil, nil
	}

	skill := &AgentSkill{
		ID:          protoSkill.GetId(),
		Name:        protoSkill.GetName(),
		Description: protoSkill.GetDescription(),
		Tags:        protoSkill.GetTags(),
		Examples:    protoSkill.GetExamples(),
		InputModes:  protoSkill.GetInputModes(),
		OutputModes: protoSkill.GetOutputModes(),
	}

	return skill, nil
}

// FromProtoRole converts a [a2a_v1.Role] to a [Role].
func FromProtoRole(protoRole a2a_v1.Role) Role {
	switch protoRole {
	case a2a_v1.Role_ROLE_USER:
		return RoleUser
	case a2a_v1.Role_ROLE_AGENT:
		return RoleAgent
	default:
		return RoleAgent // Default to agent for unknown roles
	}
}
