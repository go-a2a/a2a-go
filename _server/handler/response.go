// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"fmt"
	"log"

	"github.com/go-json-experiment/json"

	a2a "github.com/go-a2a/a2a-go"
)

// JSONRPCRequest represents a JSON-RPC request.
type JSONRPCRequest struct {
	ID      any    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	JSONRPC string `json:"jsonrpc"`
}

// Validate ensures the JSONRPCRequest is valid.
func (r *JSONRPCRequest) Validate() error {
	if r.Method == "" {
		return NewValidationError("method", "method cannot be empty")
	}
	if r.JSONRPC != "2.0" {
		return NewValidationError("jsonrpc", "jsonrpc must be '2.0'")
	}
	return nil
}

// JSONRPCResponse represents a JSON-RPC response.
type JSONRPCResponse struct {
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
	JSONRPC string `json:"jsonrpc"`
}

// Validate ensures the JSONRPCResponse is valid.
func (r *JSONRPCResponse) Validate() error {
	if r.JSONRPC != "2.0" {
		return NewValidationError("jsonrpc", "jsonrpc must be '2.0'")
	}
	if r.Result == nil && r.Error == nil {
		return NewValidationError("result", "either result or error must be present")
	}
	if r.Result != nil && r.Error != nil {
		return NewValidationError("result", "result and error cannot both be present")
	}
	return nil
}

// JSONRPCErrorResponse represents a JSON-RPC error response.
type JSONRPCErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Validate ensures the JSONRPCErrorResponse is valid.
func (r *JSONRPCErrorResponse) Validate() error {
	if r.Message == "" {
		return NewValidationError("message", "message cannot be empty")
	}
	return nil
}

// BuildErrorResponse creates a JSON-RPC error response from an error.
// This function converts various error types into properly formatted JSON-RPC error responses.
func BuildErrorResponse(requestID any, err error) *JSONRPCResponse {
	if err == nil {
		return nil
	}

	var errorResponse *JSONRPCErrorResponse

	// Handle different error types
	switch e := err.(type) {
	case *TaskNotFoundError:
		errorResponse = &JSONRPCErrorResponse{
			Code:    e.GetCode(),
			Message: e.GetMessage(),
			Data:    map[string]string{"task_id": e.TaskID},
		}
	case *TaskNotCancelableError:
		errorResponse = &JSONRPCErrorResponse{
			Code:    e.GetCode(),
			Message: e.GetMessage(),
			Data:    map[string]string{"task_id": e.TaskID, "reason": e.Reason},
		}
	case *ValidationError:
		errorResponse = &JSONRPCErrorResponse{
			Code:    a2a.ErrorCodeInvalidParams,
			Message: e.Error(),
			Data:    map[string]string{"field": e.Field},
		}
	case *InvalidRequestError:
		errorResponse = &JSONRPCErrorResponse{
			Code:    e.GetCode(),
			Message: e.GetMessage(),
			Data:    map[string]string{"details": e.Details},
		}
	case *ExecutionError:
		errorResponse = &JSONRPCErrorResponse{
			Code:    e.GetCode(),
			Message: e.GetMessage(),
			Data:    map[string]string{"task_id": e.TaskID, "details": e.Details},
		}
	case *StorageError:
		errorResponse = &JSONRPCErrorResponse{
			Code:    e.GetCode(),
			Message: e.GetMessage(),
			Data:    map[string]string{"operation": e.Operation, "task_id": e.TaskID, "details": e.Details},
		}
	case *InternalError:
		errorResponse = &JSONRPCErrorResponse{
			Code:    e.GetCode(),
			Message: e.GetMessage(),
			Data:    map[string]string{"details": e.Details},
		}
	default:
		// Generic error handling
		errorResponse = &JSONRPCErrorResponse{
			Code:    a2a.ErrorCodeInternalError,
			Message: "Internal server error",
			Data:    map[string]string{"details": err.Error()},
		}
	}

	return &JSONRPCResponse{
		ID:      requestID,
		Error:   errorResponse,
		JSONRPC: "2.0",
	}
}

// PrepareResponseObject creates a JSON-RPC response object from a result or error.
// This function handles the conversion of handler results into properly formatted JSON-RPC responses.
func PrepareResponseObject(requestID, result any, err error) *JSONRPCResponse {
	if err != nil {
		return BuildErrorResponse(requestID, err)
	}

	if result == nil {
		log.Printf("Warning: nil result provided to PrepareResponseObject")
		result = map[string]any{}
	}

	response := &JSONRPCResponse{
		ID:      requestID,
		Result:  result,
		JSONRPC: "2.0",
	}

	if validationErr := response.Validate(); validationErr != nil {
		log.Printf("Response validation failed: %v", validationErr)
		return BuildErrorResponse(requestID, NewInternalError("failed to create valid response"))
	}

	return response
}

// BuildSuccessResponse creates a successful JSON-RPC response.
func BuildSuccessResponse(requestID, result any) *JSONRPCResponse {
	return PrepareResponseObject(requestID, result, nil)
}

// SerializeResponse serializes a JSON-RPC response to JSON bytes.
func SerializeResponse(response *JSONRPCResponse) ([]byte, error) {
	if response == nil {
		return nil, fmt.Errorf("response cannot be nil")
	}

	if err := response.Validate(); err != nil {
		return nil, fmt.Errorf("response validation failed: %w", err)
	}

	data, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return data, nil
}

// DeserializeRequest deserializes JSON bytes into a JSON-RPC request.
func DeserializeRequest(data []byte) (*JSONRPCRequest, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("request data cannot be empty")
	}

	var request JSONRPCRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	return &request, nil
}

// ExtractRequestParams extracts and validates request parameters from a JSON-RPC request.
func ExtractRequestParams(request *JSONRPCRequest, target any) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.Params == nil {
		return fmt.Errorf("request params cannot be nil")
	}

	// Convert params to JSON and then unmarshal into target
	paramsJSON, err := json.Marshal(request.Params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	if err := json.Unmarshal(paramsJSON, target); err != nil {
		return fmt.Errorf("failed to unmarshal params: %w", err)
	}

	return nil
}

// ResponseHelper provides utility methods for building responses.
type ResponseHelper struct {
	requestID any
}

// NewResponseHelper creates a new ResponseHelper.
func NewResponseHelper(requestID any) *ResponseHelper {
	return &ResponseHelper{requestID: requestID}
}

// Success creates a successful response.
func (h *ResponseHelper) Success(result any) *JSONRPCResponse {
	return BuildSuccessResponse(h.requestID, result)
}

// Error creates an error response.
func (h *ResponseHelper) Error(err error) *JSONRPCResponse {
	return BuildErrorResponse(h.requestID, err)
}

// Serialize serializes a response to JSON bytes.
func (h *ResponseHelper) Serialize(response *JSONRPCResponse) ([]byte, error) {
	return SerializeResponse(response)
}

// MethodRouter provides method routing for JSON-RPC requests.
type MethodRouter struct {
	methods map[string]func(*JSONRPCRequest) *JSONRPCResponse
}

// NewMethodRouter creates a new MethodRouter.
func NewMethodRouter() *MethodRouter {
	return &MethodRouter{
		methods: make(map[string]func(*JSONRPCRequest) *JSONRPCResponse),
	}
}

// RegisterMethod registers a method handler.
func (r *MethodRouter) RegisterMethod(method string, handler func(*JSONRPCRequest) *JSONRPCResponse) {
	r.methods[method] = handler
}

// Route routes a JSON-RPC request to the appropriate handler.
func (r *MethodRouter) Route(request *JSONRPCRequest) *JSONRPCResponse {
	if err := request.Validate(); err != nil {
		return BuildErrorResponse(request.ID, NewInvalidRequestError(err.Error()))
	}

	handler, exists := r.methods[request.Method]
	if !exists {
		return BuildErrorResponse(request.ID, NewInvalidRequestError(fmt.Sprintf("method not found: %s", request.Method)))
	}

	return handler(request)
}
