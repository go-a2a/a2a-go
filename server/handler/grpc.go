// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"context"
	"fmt"
	"log"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/server"
)

// GRPCHandler provides gRPC protocol adaptation for request handlers.
// It translates gRPC requests into handler calls and formats responses
// according to the gRPC specification.
type GRPCHandler struct {
	handler        RequestHandler
	contextBuilder server.CallContextBuilder
	agentCard      *a2a.AgentCard
}

// GRPCHandlerOption defines a function type for configuring GRPCHandler.
type GRPCHandlerOption func(*GRPCHandler)

// WithGRPCAgentCard sets the agent card for capability validation.
func WithGRPCAgentCard(card *a2a.AgentCard) GRPCHandlerOption {
	return func(h *GRPCHandler) {
		h.agentCard = card
	}
}

// WithGRPCContextBuilder sets the context builder for gRPC requests.
func WithGRPCContextBuilder(builder server.CallContextBuilder) GRPCHandlerOption {
	return func(h *GRPCHandler) {
		h.contextBuilder = builder
	}
}

// NewGRPCHandler creates a new GRPCHandler with the provided request handler.
func NewGRPCHandler(handler RequestHandler, opts ...GRPCHandlerOption) *GRPCHandler {
	if handler == nil {
		panic("request handler cannot be nil")
	}

	grpcHandler := &GRPCHandler{
		handler:        handler,
		contextBuilder: server.NewDefaultCallContextBuilder(),
	}

	for _, opt := range opts {
		opt(grpcHandler)
	}

	return grpcHandler
}

// GetTask handles gRPC GetTask requests.
func (h *GRPCHandler) GetTask(ctx context.Context, req *GetTaskRequest) (*GetTaskResponse, error) {
	if req == nil {
		return nil, NewInvalidRequestError("request cannot be nil")
	}

	if err := req.Validate(); err != nil {
		return nil, NewInvalidRequestError(err.Error())
	}

	// Build call context
	callCtx, err := h.contextBuilder.Build(ctx, req)
	if err != nil {
		log.Printf("Failed to build call context: %v", err)
		return nil, NewInternalError("failed to build call context")
	}

	// Call the handler
	response, err := h.handler.OnGetTask(ctx, callCtx, req)
	if err != nil {
		log.Printf("Handler OnGetTask failed: %v", err)
		return nil, err
	}

	return response, nil
}

// CancelTask handles gRPC CancelTask requests.
func (h *GRPCHandler) CancelTask(ctx context.Context, req *CancelTaskRequest) (*CancelTaskResponse, error) {
	if req == nil {
		return nil, NewInvalidRequestError("request cannot be nil")
	}

	if err := req.Validate(); err != nil {
		return nil, NewInvalidRequestError(err.Error())
	}

	// Build call context
	callCtx, err := h.contextBuilder.Build(ctx, req)
	if err != nil {
		log.Printf("Failed to build call context: %v", err)
		return nil, NewInternalError("failed to build call context")
	}

	// Call the handler
	response, err := h.handler.OnCancelTask(ctx, callCtx, req)
	if err != nil {
		log.Printf("Handler OnCancelTask failed: %v", err)
		return nil, err
	}

	return response, nil
}

// MessageSend handles gRPC MessageSend requests.
func (h *GRPCHandler) MessageSend(ctx context.Context, req *MessageSendRequest) (*MessageSendResponse, error) {
	if req == nil {
		return nil, NewInvalidRequestError("request cannot be nil")
	}

	if err := req.Validate(); err != nil {
		return nil, NewInvalidRequestError(err.Error())
	}

	// Build call context
	callCtx, err := h.contextBuilder.Build(ctx, req)
	if err != nil {
		log.Printf("Failed to build call context: %v", err)
		return nil, NewInternalError("failed to build call context")
	}

	// Call the handler
	response, err := h.handler.OnMessageSend(ctx, callCtx, req)
	if err != nil {
		log.Printf("Handler OnMessageSend failed: %v", err)
		return nil, err
	}

	return response, nil
}

// MessageSendStream handles gRPC MessageSendStream requests with streaming response.
func (h *GRPCHandler) MessageSendStream(ctx context.Context, req *MessageSendStreamRequest) (*MessageSendStreamResponse, error) {
	if req == nil {
		return nil, NewInvalidRequestError("request cannot be nil")
	}

	if err := req.Validate(); err != nil {
		return nil, NewInvalidRequestError(err.Error())
	}

	// Validate streaming capability
	if err := h.validateStreamingCapability(); err != nil {
		return nil, err
	}

	// Build call context
	callCtx, err := h.contextBuilder.Build(ctx, req)
	if err != nil {
		log.Printf("Failed to build call context: %v", err)
		return nil, NewInternalError("failed to build call context")
	}

	// Call the handler
	response, err := h.handler.OnMessageSendStream(ctx, callCtx, req)
	if err != nil {
		log.Printf("Handler OnMessageSendStream failed: %v", err)
		return nil, err
	}

	return response, nil
}

// ConfigurePushNotification handles gRPC push notification configuration.
func (h *GRPCHandler) ConfigurePushNotification(ctx context.Context, config *PushNotificationConfig) error {
	if config == nil {
		return NewInvalidRequestError("config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return NewInvalidRequestError(err.Error())
	}

	// Validate push notification capability
	if err := h.validatePushNotificationCapability(); err != nil {
		return err
	}

	// This would typically be handled by a push notification service
	// For now, we'll just log the configuration
	log.Printf("Push notification configured for task %s: enabled=%t", config.TaskID, config.Enabled)

	return nil
}

// validateStreamingCapability validates that the agent supports streaming.
func (h *GRPCHandler) validateStreamingCapability() error {
	if h.agentCard == nil {
		return NewInvalidRequestError("agent card not configured")
	}

	// Check if the agent supports streaming
	// TODO: Implement actual streaming capability check based on agent capabilities
	if h.agentCard.Capabilities == nil {
		return NewInvalidRequestError("streaming not supported by this agent")
	}

	return nil
}

// validatePushNotificationCapability validates that the agent supports push notifications.
func (h *GRPCHandler) validatePushNotificationCapability() error {
	if h.agentCard == nil {
		return NewInvalidRequestError("agent card not configured")
	}

	// Check if the agent supports push notifications
	// TODO: Implement actual push notification capability check based on agent capabilities
	if h.agentCard.Capabilities == nil {
		return NewInvalidRequestError("push notifications not supported by this agent")
	}

	return nil
}

// SetAgentCard sets the agent card for capability validation.
func (h *GRPCHandler) SetAgentCard(card *a2a.AgentCard) {
	h.agentCard = card
}

// GetAgentCard returns the current agent card.
func (h *GRPCHandler) GetAgentCard() *a2a.AgentCard {
	return h.agentCard
}

// Validate ensures the GRPCHandler is in a valid state.
func (h *GRPCHandler) Validate() error {
	if h.handler == nil {
		return fmt.Errorf("request handler cannot be nil")
	}
	if h.contextBuilder == nil {
		return fmt.Errorf("context builder cannot be nil")
	}
	return nil
}

// String returns a string representation of the GRPCHandler for debugging.
func (h *GRPCHandler) String() string {
	hasAgentCard := h.agentCard != nil
	return fmt.Sprintf("GRPCHandler{has_agent_card: %t}", hasAgentCard)
}

// StreamingServer defines the interface for gRPC streaming servers.
type StreamingServer interface {
	Send(*a2a.Message) error
	Context() context.Context
}

// HandleMessageStream processes a streaming message request.
func (h *GRPCHandler) HandleMessageStream(req *MessageSendStreamRequest, stream StreamingServer) error {
	if req == nil {
		return NewInvalidRequestError("request cannot be nil")
	}

	if err := req.Validate(); err != nil {
		return NewInvalidRequestError(err.Error())
	}

	// Validate streaming capability
	if err := h.validateStreamingCapability(); err != nil {
		return err
	}

	// Build call context
	callCtx, err := h.contextBuilder.Build(stream.Context(), req)
	if err != nil {
		log.Printf("Failed to build call context: %v", err)
		return NewInternalError("failed to build call context")
	}

	// Call the handler
	response, err := h.handler.OnMessageSendStream(stream.Context(), callCtx, req)
	if err != nil {
		log.Printf("Handler OnMessageSendStream failed: %v", err)
		return err
	}

	// Send initial messages
	for _, msg := range response.Messages {
		if err := stream.Send(msg); err != nil {
			log.Printf("Failed to send message: %v", err)
			return NewInternalError("failed to send message")
		}
	}

	// Stream additional messages
	if response.Stream != nil {
		for msg := range response.Stream {
			if msg == nil {
				continue
			}
			if err := stream.Send(msg); err != nil {
				log.Printf("Failed to send streaming message: %v", err)
				return NewInternalError("failed to send streaming message")
			}
		}
	}

	return nil
}

// ErrorToGRPCStatus converts handler errors to gRPC status codes.
func ErrorToGRPCStatus(err error) error {
	if err == nil {
		return nil
	}

	// This would typically use gRPC's status package to create proper gRPC errors
	// For now, we'll return the error as-is
	return err
}
