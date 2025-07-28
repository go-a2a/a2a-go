// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"context"
	"fmt"
	"log"
	"strings"

	a2a_v1 "github.com/go-a2a/a2a-grpc/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/server"
)

// GRPCHandler maps incoming gRPC requests to the appropriate request handler method.
//
// It implements the [a2a_v1.A2AServiceServer] interface and delegates to the underlying [RequestHandler].
type GRPCHandler struct {
	a2a_v1.UnimplementedA2AServiceServer

	agentCard      *a2a.AgentCard
	requestHandler RequestHandler
	contextBuilder server.CallContextBuilder
}

// NewGRPCHandler creates a new GrpcHandler with the provided configuration.
func NewGRPCHandler(agentCard *a2a.AgentCard, requestHandler RequestHandler, contextBuilder server.CallContextBuilder) *GRPCHandler {
	if agentCard == nil {
		panic("agent card cannot be nil")
	}
	if requestHandler == nil {
		panic("request handler cannot be nil")
	}
	if contextBuilder == nil {
		contextBuilder = server.NewDefaultCallContextBuilder()
	}

	return &GRPCHandler{
		agentCard:      agentCard,
		requestHandler: requestHandler,
		contextBuilder: contextBuilder,
	}
}

// SendMessage handles the 'SendMessage' gRPC method.
func (h *GRPCHandler) SendMessage(ctx context.Context, request *a2a_v1.SendMessageRequest) (*a2a_v1.SendMessageResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Build server context
	serverContext, err := h.contextBuilder.Build(ctx, request)
	if err != nil {
		log.Printf("Failed to build server context: %v", err)
		return nil, status.Error(codes.Internal, "failed to build server context")
	}

	// Transform the proto object to internal objects
	a2aRequest, err := a2a.FromProtoMessageSendParams(request)
	if err != nil {
		return nil, h.abortContext(err)
	}

	// Call the handler
	response, err := h.requestHandler.OnMessageSend(ctx, a2aRequest, serverContext)
	if err != nil {
		return nil, h.abortContext(err)
	}

	result, err := a2a.ToProtoTaskOrMessage(response)
	if err != nil {
		return nil, h.abortContext(err)
	}

	// Transform response back to proto
	return result, nil
}

// SendStreamingMessage handles the 'SendStreamingMessage' gRPC method.
func (h *GRPCHandler) SendStreamingMessage(request *a2a_v1.SendMessageRequest, stream a2a_v1.A2AService_SendStreamingMessageServer) error {
	if request == nil {
		return status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Validate streaming capability
	if err := h.validateStreamingCapability(); err != nil {
		return err
	}

	// Build server context
	serverContext, err := h.contextBuilder.Build(stream.Context(), request)
	if err != nil {
		log.Printf("Failed to build server context: %v", err)
		return status.Error(codes.Internal, "failed to build server context")
	}

	// Transform the proto object to internal objects
	a2aRequest, err := h.transformSendStreamingMessageRequest(request)
	if err != nil {
		return h.abortContext(err)
	}

	// Call the handler
	response, err := h.requestHandler.OnMessageSendStream(stream.Context(), serverContext, a2aRequest)
	if err != nil {
		return h.abortContext(err)
	}

	// Send initial messages
	for _, msg := range response.Messages {
		streamResp, err := h.transformToStreamResponse(msg)
		if err != nil {
			return h.abortContext(err)
		}
		if err := stream.Send(streamResp); err != nil {
			log.Printf("Failed to send stream response: %v", err)
			return status.Error(codes.Internal, "failed to send stream response")
		}
	}

	// Handle streaming messages if available
	if response.Stream != nil {
		for msg := range response.Stream {
			if msg == nil {
				continue
			}
			streamResp, err := h.transformToStreamResponse(msg)
			if err != nil {
				return h.abortContext(err)
			}
			if err := stream.Send(streamResp); err != nil {
				log.Printf("Failed to send streaming message: %v", err)
				return status.Error(codes.Internal, "failed to send streaming message")
			}
		}
	}

	return nil
}

// CancelTask handles the 'CancelTask' gRPC method.
func (h *GRPCHandler) CancelTask(ctx context.Context, request *a2a_v1.CancelTaskRequest) (*a2a_v1.Task, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Build server context
	serverContext, err := h.contextBuilder.Build(ctx, request)
	if err != nil {
		log.Printf("Failed to build server context: %v", err)
		return nil, status.Error(codes.Internal, "failed to build server context")
	}

	// Transform the proto object to internal objects
	a2aRequest, err := h.transformCancelTaskRequest(request)
	if err != nil {
		return nil, h.abortContext(err)
	}

	// Call the handler
	response, err := h.requestHandler.OnCancelTask(ctx, serverContext, a2aRequest)
	if err != nil {
		return nil, h.abortContext(err)
	}

	// Transform response back to proto
	return h.transformTaskResponse(response)
}

// TaskSubscription handles the 'TaskSubscription' gRPC method.
func (h *GRPCHandler) TaskSubscription(request *a2a_v1.TaskSubscriptionRequest, stream a2a_v1.A2AService_TaskSubscriptionServer) error {
	if request == nil {
		return status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Validate streaming capability
	if err := h.validateStreamingCapability(); err != nil {
		return err
	}

	// Build server context
	serverContext, err := h.contextBuilder.Build(stream.Context(), request)
	if err != nil {
		log.Printf("Failed to build server context: %v", err)
		return status.Error(codes.Internal, "failed to build server context")
	}

	// Transform the proto object to internal objects
	a2aRequest, err := h.transformTaskSubscriptionRequest(request)
	if err != nil {
		return h.abortContext(err)
	}

	// Call the handler
	response, err := h.requestHandler.OnResubscribeToTask(stream.Context(), serverContext, a2aRequest)
	if err != nil {
		return h.abortContext(err)
	}

	// Send initial messages
	for _, msg := range response.Messages {
		streamResp, err := h.transformToStreamResponse(msg)
		if err != nil {
			return h.abortContext(err)
		}
		if err := stream.Send(streamResp); err != nil {
			log.Printf("Failed to send stream response: %v", err)
			return status.Error(codes.Internal, "failed to send stream response")
		}
	}

	// Handle streaming messages if available
	if response.Stream != nil {
		for msg := range response.Stream {
			if msg == nil {
				continue
			}
			streamResp, err := h.transformToStreamResponse(msg)
			if err != nil {
				return h.abortContext(err)
			}
			if err := stream.Send(streamResp); err != nil {
				log.Printf("Failed to send streaming message: %v", err)
				return status.Error(codes.Internal, "failed to send streaming message")
			}
		}
	}

	return nil
}

// GetTaskPushNotificationConfig handles the 'GetTaskPushNotificationConfig' gRPC method.
func (h *GRPCHandler) GetTaskPushNotificationConfig(ctx context.Context, request *a2a_v1.GetTaskPushNotificationConfigRequest) (*a2a_v1.TaskPushNotificationConfig, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Build server context
	serverContext, err := h.contextBuilder.Build(ctx, request)
	if err != nil {
		log.Printf("Failed to build server context: %v", err)
		return nil, status.Error(codes.Internal, "failed to build server context")
	}

	// Transform the proto object to internal objects
	a2aRequest, err := h.transformGetTaskPushNotificationConfigRequest(request)
	if err != nil {
		return nil, h.abortContext(err)
	}

	// Call the handler
	response, err := h.requestHandler.OnGetTaskPushNotificationConfig(ctx, serverContext, a2aRequest)
	if err != nil {
		return nil, h.abortContext(err)
	}

	// Transform response back to proto
	return h.transformTaskPushNotificationConfigResponse(response)
}

// CreateTaskPushNotificationConfig handles the 'CreateTaskPushNotificationConfig' gRPC method.
func (h *GRPCHandler) CreateTaskPushNotificationConfig(ctx context.Context, request *a2a_v1.CreateTaskPushNotificationConfigRequest) (*a2a_v1.TaskPushNotificationConfig, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Validate push notification capability
	if err := h.validatePushNotificationCapability(); err != nil {
		return nil, err
	}

	// Build server context
	serverContext, err := h.contextBuilder.Build(ctx, request)
	if err != nil {
		log.Printf("Failed to build server context: %v", err)
		return nil, status.Error(codes.Internal, "failed to build server context")
	}

	// Transform the proto object to internal objects
	a2aRequest, err := h.transformCreateTaskPushNotificationConfigRequest(request)
	if err != nil {
		return nil, h.abortContext(err)
	}

	// Call the handler
	response, err := h.requestHandler.OnSetTaskPushNotificationConfig(ctx, serverContext, a2aRequest)
	if err != nil {
		return nil, h.abortContext(err)
	}

	// Transform response back to proto
	return h.transformSetTaskPushNotificationConfigResponse(response)
}

// ListTaskPushNotificationConfig handles the 'ListTaskPushNotificationConfig' gRPC method.
func (h *GRPCHandler) ListTaskPushNotificationConfig(ctx context.Context, request *a2a_v1.ListTaskPushNotificationConfigRequest) (*a2a_v1.ListTaskPushNotificationConfigResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Build server context
	serverContext, err := h.contextBuilder.Build(ctx, request)
	if err != nil {
		log.Printf("Failed to build server context: %v", err)
		return nil, status.Error(codes.Internal, "failed to build server context")
	}

	// Transform the proto object to internal objects
	a2aRequest, err := h.transformListTaskPushNotificationConfigRequest(request)
	if err != nil {
		return nil, h.abortContext(err)
	}

	// Call the handler
	response, err := h.requestHandler.OnListTaskPushNotificationConfig(ctx, serverContext, a2aRequest)
	if err != nil {
		return nil, h.abortContext(err)
	}

	// Transform response back to proto
	return h.transformListTaskPushNotificationConfigResponse(response)
}

// DeleteTaskPushNotificationConfig handles the 'DeleteTaskPushNotificationConfig' gRPC method.
func (h *GRPCHandler) DeleteTaskPushNotificationConfig(ctx context.Context, request *a2a_v1.DeleteTaskPushNotificationConfigRequest) (*emptypb.Empty, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Build server context
	serverContext, err := h.contextBuilder.Build(ctx, request)
	if err != nil {
		log.Printf("Failed to build server context: %v", err)
		return nil, status.Error(codes.Internal, "failed to build server context")
	}

	// Transform the proto object to internal objects
	a2aRequest, err := h.transformDeleteTaskPushNotificationConfigRequest(request)
	if err != nil {
		return nil, h.abortContext(err)
	}

	// Call the handler
	_, err = h.requestHandler.OnDeleteTaskPushNotificationConfig(ctx, serverContext, a2aRequest)
	if err != nil {
		return nil, h.abortContext(err)
	}

	// Return empty response
	return &emptypb.Empty{}, nil
}

// GetTask handles the 'GetTask' gRPC method.
func (h *GRPCHandler) GetTask(ctx context.Context, request *a2a_v1.GetTaskRequest) (*a2a_v1.Task, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Build server context
	serverContext, err := h.contextBuilder.Build(ctx, request)
	if err != nil {
		log.Printf("Failed to build server context: %v", err)
		return nil, status.Error(codes.Internal, "failed to build server context")
	}

	// Transform the proto object to internal objects
	a2aRequest, err := h.transformGetTaskRequest(request)
	if err != nil {
		return nil, h.abortContext(err)
	}

	// Call the handler
	response, err := h.requestHandler.OnGetTask(ctx, serverContext, a2aRequest)
	if err != nil {
		return nil, h.abortContext(err)
	}

	// Transform response back to proto
	return h.transformTaskFromGetTaskResponse(response)
}

// GetAgentCard handles the 'GetAgentCard' gRPC method.
func (h *GRPCHandler) GetAgentCard(ctx context.Context, request *a2a_v1.GetAgentCardRequest) (*a2a_v1.AgentCard, error) {
	// Transform agent card to proto
	return h.transformAgentCardToProto(h.agentCard)
}

// validateStreamingCapability validates that the agent supports streaming.
func (h *GRPCHandler) validateStreamingCapability() error {
	if h.agentCard == nil {
		return status.Error(codes.InvalidArgument, "agent card not configured")
	}

	if h.agentCard.Capabilities == nil || !h.agentCard.Capabilities.Streaming {
		return status.Error(codes.Unimplemented, "streaming is not supported by the agent")
	}

	return nil
}

// validatePushNotificationCapability validates that the agent supports push notifications.
func (h *GRPCHandler) validatePushNotificationCapability() error {
	if h.agentCard == nil {
		return status.Error(codes.InvalidArgument, "agent card not configured")
	}

	if h.agentCard.Capabilities == nil || !h.agentCard.Capabilities.PushNotifications {
		return status.Error(codes.Unimplemented, "push notifications are not supported by the agent")
	}

	return nil
}

// abortContext sets the gRPC errors appropriately based on the error type.
func (h *GRPCHandler) abortContext(err error) error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *ValidationError:
		return status.Error(codes.InvalidArgument, fmt.Sprintf("ValidationError: %s", e.Error()))
	case *InvalidRequestError:
		return status.Error(codes.InvalidArgument, fmt.Sprintf("InvalidRequestError: %s", e.Error()))
	case *TaskNotFoundError:
		return status.Error(codes.NotFound, fmt.Sprintf("TaskNotFoundError: %s", e.Error()))
	case *TaskNotCancelableError:
		return status.Error(codes.Unimplemented, fmt.Sprintf("TaskNotCancelableError: %s", e.Error()))
	case *UnsupportedOperationError:
		return status.Error(codes.Unimplemented, fmt.Sprintf("UnsupportedOperationError: %s", e.Error()))
	case *InternalError:
		return status.Error(codes.Internal, fmt.Sprintf("InternalError: %s", e.Error()))
	case *ExecutionError:
		return status.Error(codes.Internal, fmt.Sprintf("ExecutionError: %s", e.Error()))
	case *StorageError:
		return status.Error(codes.Internal, fmt.Sprintf("StorageError: %s", e.Error()))
	default:
		return status.Error(codes.Unknown, fmt.Sprintf("Unknown error type: %s", err.Error()))
	}
}

// Transform methods - these need to be implemented based on the proto definitions
// For now, returning placeholder implementations

func (h *GRPCHandler) transformSendMessageRequest(req *a2a_v1.SendMessageRequest) (*MessageSendRequest, error) {
	// TODO: Implement proto to internal type conversion
	// The task ID is in the Message field, not directly in the request
	var taskID string
	if req.GetRequest() != nil {
		taskID = req.GetRequest().GetTaskId()
	}

	return &MessageSendRequest{
		Messages: []*a2a.Message{}, // Placeholder - need to convert req.GetRequest()
		TaskID:   taskID,
	}, nil
}

func (h *GRPCHandler) transformSendMessageResponse(resp *MessageSendResponse) (*a2a_v1.SendMessageResponse, error) {
	// TODO: Implement internal to proto type conversion
	return &a2a_v1.SendMessageResponse{
		// TODO: Need to check the actual field name in the proto
		// TaskId: resp.TaskID,
	}, nil
}

func (h *GRPCHandler) transformSendStreamingMessageRequest(req *a2a_v1.SendMessageRequest) (*MessageSendStreamRequest, error) {
	// TODO: Implement proto to internal type conversion
	// The task ID is in the Message field, not directly in the request
	var taskID string
	if req.GetRequest() != nil {
		taskID = req.GetRequest().GetTaskId()
	}

	return &MessageSendStreamRequest{
		Messages: []*a2a.Message{}, // Placeholder - need to convert req.GetRequest()
		TaskID:   taskID,
	}, nil
}

func (h *GRPCHandler) transformToStreamResponse(msg *a2a.Message) (*a2a_v1.StreamResponse, error) {
	// TODO: Implement message to stream response conversion
	return &a2a_v1.StreamResponse{}, nil
}

func (h *GRPCHandler) transformCancelTaskRequest(req *a2a_v1.CancelTaskRequest) (*CancelTaskRequest, error) {
	// The task ID is in the "name" field in format "tasks/{id}"
	taskID := extractTaskIDFromName(req.GetName())
	return &CancelTaskRequest{
		TaskID: taskID,
	}, nil
}

func (h *GRPCHandler) transformTaskResponse(resp *CancelTaskResponse) (*a2a_v1.Task, error) {
	// TODO: Implement proper task conversion
	return &a2a_v1.Task{}, nil
}

func (h *GRPCHandler) transformTaskSubscriptionRequest(req *a2a_v1.TaskSubscriptionRequest) (*ResubscribeToTaskRequest, error) {
	// The task ID is in the "name" field in format "tasks/{id}"
	taskID := extractTaskIDFromName(req.GetName())
	return &ResubscribeToTaskRequest{
		TaskID: taskID,
	}, nil
}

func (h *GRPCHandler) transformGetTaskPushNotificationConfigRequest(req *a2a_v1.GetTaskPushNotificationConfigRequest) (*GetTaskPushNotificationConfigRequest, error) {
	// The task ID is embedded in the "name" field in format "tasks/{id}/pushNotificationConfigs/{config_id}"
	taskID, configID := extractTaskAndConfigIDFromName(req.GetName())
	return &GetTaskPushNotificationConfigRequest{
		TaskID:                   taskID,
		PushNotificationConfigID: configID,
	}, nil
}

func (h *GRPCHandler) transformCreateTaskPushNotificationConfigRequest(req *a2a_v1.CreateTaskPushNotificationConfigRequest) (*SetTaskPushNotificationConfigRequest, error) {
	// The task ID is embedded in the "parent" field in format "tasks/{id}/pushNotificationConfigs"
	taskID := extractTaskIDFromParent(req.GetParent())
	return &SetTaskPushNotificationConfigRequest{
		TaskID: taskID,
		// TODO: Convert req.GetConfig() to *a2a.PushNotificationConfig
	}, nil
}

func (h *GRPCHandler) transformTaskPushNotificationConfigResponse(resp *GetTaskPushNotificationConfigResponse) (*a2a_v1.TaskPushNotificationConfig, error) {
	// TODO: Implement internal to proto type conversion
	return &a2a_v1.TaskPushNotificationConfig{}, nil
}

func (h *GRPCHandler) transformSetTaskPushNotificationConfigResponse(resp *SetTaskPushNotificationConfigResponse) (*a2a_v1.TaskPushNotificationConfig, error) {
	// TODO: Implement internal to proto type conversion
	return &a2a_v1.TaskPushNotificationConfig{}, nil
}

func (h *GRPCHandler) transformListTaskPushNotificationConfigRequest(req *a2a_v1.ListTaskPushNotificationConfigRequest) (*ListTaskPushNotificationConfigRequest, error) {
	// The task ID is embedded in the "parent" field in format "tasks/{id}"
	taskID := extractTaskIDFromParent(req.GetParent())
	return &ListTaskPushNotificationConfigRequest{
		ID: taskID,
	}, nil
}

func (h *GRPCHandler) transformListTaskPushNotificationConfigResponse(resp *ListTaskPushNotificationConfigResponse) (*a2a_v1.ListTaskPushNotificationConfigResponse, error) {
	// TODO: Implement internal to proto type conversion
	return &a2a_v1.ListTaskPushNotificationConfigResponse{}, nil
}

func (h *GRPCHandler) transformDeleteTaskPushNotificationConfigRequest(req *a2a_v1.DeleteTaskPushNotificationConfigRequest) (*DeleteTaskPushNotificationConfigRequest, error) {
	// The task ID is embedded in the "name" field in format "tasks/{id}/pushNotificationConfigs/{config_id}"
	taskID, configID := extractTaskAndConfigIDFromName(req.GetName())
	return &DeleteTaskPushNotificationConfigRequest{
		TaskID:                   taskID,
		PushNotificationConfigID: configID,
	}, nil
}

// transformDeleteTaskPushNotificationConfigResponse is not needed since the method returns emptypb.Empty

func (h *GRPCHandler) transformGetTaskRequest(req *a2a_v1.GetTaskRequest) (*GetTaskRequest, error) {
	// The task ID is in the "name" field in format "tasks/{id}"
	taskID := extractTaskIDFromName(req.GetName())
	historyLength := int(req.GetHistoryLength())
	return &GetTaskRequest{
		TaskID:        taskID,
		HistoryLength: &historyLength,
	}, nil
}

func (h *GRPCHandler) transformTaskFromGetTaskResponse(resp *GetTaskResponse) (*a2a_v1.Task, error) {
	// TODO: Implement internal to proto type conversion
	return &a2a_v1.Task{}, nil
}

func (h *GRPCHandler) transformAgentCardToProto(card *a2a.AgentCard) (*a2a_v1.AgentCard, error) {
	// TODO: Implement internal to proto type conversion
	return &a2a_v1.AgentCard{}, nil
}

// Helper functions to extract task IDs from resource names

// extractTaskIDFromName extracts task ID from "tasks/{id}" format
func extractTaskIDFromName(name string) string {
	if name == "" {
		return ""
	}
	// Expected format: "tasks/{id}"
	if len(name) > 6 && name[:6] == "tasks/" {
		return name[6:]
	}
	return name
}

// extractTaskIDFromParent extracts task ID from "tasks/{id}" or "tasks/{id}/pushNotificationConfigs" format
func extractTaskIDFromParent(parent string) string {
	if parent == "" {
		return ""
	}
	// Expected format: "tasks/{id}" or "tasks/{id}/pushNotificationConfigs"
	if len(parent) > 6 && parent[:6] == "tasks/" {
		// Find the next slash or end of string
		remaining := parent[6:]
		slashIndex := strings.Index(remaining, "/")
		if slashIndex == -1 {
			return remaining
		}
		return remaining[:slashIndex]
	}
	return parent
}

// extractTaskAndConfigIDFromName extracts task ID and config ID from "tasks/{id}/pushNotificationConfigs/{config_id}" format
func extractTaskAndConfigIDFromName(name string) (string, string) {
	if name == "" {
		return "", ""
	}
	// Expected format: "tasks/{id}/pushNotificationConfigs/{config_id}"
	if len(name) > 6 && name[:6] == "tasks/" {
		remaining := name[6:]
		slashIndex := strings.Index(remaining, "/")
		if slashIndex == -1 {
			return remaining, ""
		}

		taskID := remaining[:slashIndex]
		configPath := remaining[slashIndex+1:]

		// Expected: "pushNotificationConfigs/{config_id}"
		if len(configPath) > 26 && configPath[:26] == "pushNotificationConfigs/" {
			configID := configPath[26:]
			return taskID, configID
		}

		return taskID, ""
	}
	return name, ""
}
