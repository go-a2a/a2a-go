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

// JSONRPCHandler provides JSON-RPC protocol adaptation for request handlers.
// It translates JSON-RPC requests into handler calls and formats responses
// according to the JSON-RPC specification.
type JSONRPCHandler struct {
	handler        RequestHandler
	contextBuilder server.CallContextBuilder
	agentCard      *a2a.AgentCard
	router         *MethodRouter
}

// JSONRPCHandlerOption defines a function type for configuring JSONRPCHandler.
type JSONRPCHandlerOption func(*JSONRPCHandler)

// WithAgentCard sets the agent card for capability validation.
func WithAgentCard(card *a2a.AgentCard) JSONRPCHandlerOption {
	return func(h *JSONRPCHandler) {
		h.agentCard = card
	}
}

// WithJSONRPCContextBuilder sets the context builder for JSON-RPC requests.
func WithJSONRPCContextBuilder(builder server.CallContextBuilder) JSONRPCHandlerOption {
	return func(h *JSONRPCHandler) {
		h.contextBuilder = builder
	}
}

// NewJSONRPCHandler creates a new JSONRPCHandler with the provided request handler.
func NewJSONRPCHandler(handler RequestHandler, opts ...JSONRPCHandlerOption) *JSONRPCHandler {
	if handler == nil {
		panic("request handler cannot be nil")
	}

	jsonrpcHandler := &JSONRPCHandler{
		handler:        handler,
		contextBuilder: server.NewDefaultCallContextBuilder(),
		router:         NewMethodRouter(),
	}

	for _, opt := range opts {
		opt(jsonrpcHandler)
	}

	// Register method handlers
	jsonrpcHandler.registerMethods()

	return jsonrpcHandler
}

// registerMethods registers all JSON-RPC method handlers.
func (h *JSONRPCHandler) registerMethods() {
	h.router.RegisterMethod("get_task", h.handleGetTask)
	h.router.RegisterMethod("cancel_task", h.handleCancelTask)
	h.router.RegisterMethod("message_send", h.handleMessageSend)
	h.router.RegisterMethod("message_send_stream", h.handleMessageSendStream)
}

// HandleRequest processes a JSON-RPC request and returns a response.
func (h *JSONRPCHandler) HandleRequest(ctx context.Context, request *JSONRPCRequest) *JSONRPCResponse {
	if request == nil {
		return BuildErrorResponse(nil, NewInvalidRequestError("request cannot be nil"))
	}

	if err := request.Validate(); err != nil {
		return BuildErrorResponse(request.ID, NewInvalidRequestError(err.Error()))
	}

	return h.router.Route(request)
}

// handleGetTask handles the get_task JSON-RPC method.
func (h *JSONRPCHandler) handleGetTask(request *JSONRPCRequest) *JSONRPCResponse {
	var params GetTaskRequest
	if err := ExtractRequestParams(request, &params); err != nil {
		return BuildErrorResponse(request.ID, NewInvalidRequestError(err.Error()))
	}

	// Build call context
	callCtx, err := h.contextBuilder.Build(context.Background(), request)
	if err != nil {
		log.Printf("Failed to build call context: %v", err)
		return BuildErrorResponse(request.ID, NewInternalError("failed to build call context"))
	}

	// Call the handler
	response, err := h.handler.OnGetTask(context.Background(), callCtx, &params)
	if err != nil {
		log.Printf("Handler OnGetTask failed: %v", err)
		return BuildErrorResponse(request.ID, err)
	}

	return BuildSuccessResponse(request.ID, response)
}

// handleCancelTask handles the cancel_task JSON-RPC method.
func (h *JSONRPCHandler) handleCancelTask(request *JSONRPCRequest) *JSONRPCResponse {
	var params CancelTaskRequest
	if err := ExtractRequestParams(request, &params); err != nil {
		return BuildErrorResponse(request.ID, NewInvalidRequestError(err.Error()))
	}

	// Build call context
	callCtx, err := h.contextBuilder.Build(context.Background(), request)
	if err != nil {
		log.Printf("Failed to build call context: %v", err)
		return BuildErrorResponse(request.ID, NewInternalError("failed to build call context"))
	}

	// Call the handler
	response, err := h.handler.OnCancelTask(context.Background(), callCtx, &params)
	if err != nil {
		log.Printf("Handler OnCancelTask failed: %v", err)
		return BuildErrorResponse(request.ID, err)
	}

	return BuildSuccessResponse(request.ID, response)
}

// handleMessageSend handles the message_send JSON-RPC method.
func (h *JSONRPCHandler) handleMessageSend(request *JSONRPCRequest) *JSONRPCResponse {
	var params MessageSendRequest
	if err := ExtractRequestParams(request, &params); err != nil {
		return BuildErrorResponse(request.ID, NewInvalidRequestError(err.Error()))
	}

	// Build call context
	callCtx, err := h.contextBuilder.Build(context.Background(), request)
	if err != nil {
		log.Printf("Failed to build call context: %v", err)
		return BuildErrorResponse(request.ID, NewInternalError("failed to build call context"))
	}

	// Call the handler
	response, err := h.handler.OnMessageSend(context.Background(), callCtx, &params)
	if err != nil {
		log.Printf("Handler OnMessageSend failed: %v", err)
		return BuildErrorResponse(request.ID, err)
	}

	return BuildSuccessResponse(request.ID, response)
}

// handleMessageSendStream handles the message_send_stream JSON-RPC method.
func (h *JSONRPCHandler) handleMessageSendStream(request *JSONRPCRequest) *JSONRPCResponse {
	// Validate streaming capability
	if err := h.validateStreamingCapability(); err != nil {
		return BuildErrorResponse(request.ID, err)
	}

	var params MessageSendStreamRequest
	if err := ExtractRequestParams(request, &params); err != nil {
		return BuildErrorResponse(request.ID, NewInvalidRequestError(err.Error()))
	}

	// Build call context
	callCtx, err := h.contextBuilder.Build(context.Background(), request)
	if err != nil {
		log.Printf("Failed to build call context: %v", err)
		return BuildErrorResponse(request.ID, NewInternalError("failed to build call context"))
	}

	// Call the handler
	response, err := h.handler.OnMessageSendStream(context.Background(), callCtx, &params)
	if err != nil {
		log.Printf("Handler OnMessageSendStream failed: %v", err)
		return BuildErrorResponse(request.ID, err)
	}

	return BuildSuccessResponse(request.ID, response)
}

// validateStreamingCapability validates that the agent supports streaming.
func (h *JSONRPCHandler) validateStreamingCapability() error {
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
func (h *JSONRPCHandler) validatePushNotificationCapability() error {
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

// HandleBatchRequest processes a batch of JSON-RPC requests.
func (h *JSONRPCHandler) HandleBatchRequest(ctx context.Context, requests []*JSONRPCRequest) []*JSONRPCResponse {
	if len(requests) == 0 {
		return []*JSONRPCResponse{
			BuildErrorResponse(nil, NewInvalidRequestError("batch request cannot be empty")),
		}
	}

	responses := make([]*JSONRPCResponse, len(requests))
	for i, request := range requests {
		responses[i] = h.HandleRequest(ctx, request)
	}

	return responses
}

// ProcessJSONRPCRequest processes a raw JSON-RPC request from bytes.
func (h *JSONRPCHandler) ProcessJSONRPCRequest(ctx context.Context, data []byte) ([]byte, error) {
	request, err := DeserializeRequest(data)
	if err != nil {
		response := BuildErrorResponse(nil, NewInvalidRequestError(err.Error()))
		return SerializeResponse(response)
	}

	response := h.HandleRequest(ctx, request)
	return SerializeResponse(response)
}

// SetAgentCard sets the agent card for capability validation.
func (h *JSONRPCHandler) SetAgentCard(card *a2a.AgentCard) {
	h.agentCard = card
}

// GetAgentCard returns the current agent card.
func (h *JSONRPCHandler) GetAgentCard() *a2a.AgentCard {
	return h.agentCard
}

// Validate ensures the JSONRPCHandler is in a valid state.
func (h *JSONRPCHandler) Validate() error {
	if h.handler == nil {
		return fmt.Errorf("request handler cannot be nil")
	}
	if h.contextBuilder == nil {
		return fmt.Errorf("context builder cannot be nil")
	}
	if h.router == nil {
		return fmt.Errorf("method router cannot be nil")
	}
	return nil
}

// String returns a string representation of the JSONRPCHandler for debugging.
func (h *JSONRPCHandler) String() string {
	hasAgentCard := h.agentCard != nil
	return fmt.Sprintf("JSONRPCHandler{has_agent_card: %t}", hasAgentCard)
}
