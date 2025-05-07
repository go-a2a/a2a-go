// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"net/http"

	"github.com/bytedance/sonic"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/internal/jsonrpc2"
)

// Server implements the A2A protocol server
type Server struct {
	taskManager    TaskManager
	mux            *http.ServeMux
	agentCard      *a2a.AgentCard
	streamRegistry *StreamRegistry
	notifier       *Notifier
	capabilities   a2a.AgentCapabilities
}

// Config holds configuration for the A2A server
type Config struct {
	// AgentCard represents metadata about the agent
	AgentCard *a2a.AgentCard
	// TaskManager is the task management implementation
	TaskManager TaskManager
	// EnableStreaming enables tasks/sendSubscribe and tasks/resubscribe
	EnableStreaming bool
	// EnablePushNotifications enables tasks/pushNotification endpoints
	EnablePushNotifications bool
	// EnableStateHistory enables state transition history
	EnableStateHistory bool
}

// NewServer creates a new A2A server instance with the provided configuration
func NewServer(cfg Config) (*Server, error) {
	if cfg.AgentCard == nil {
		return nil, fmt.Errorf("agent card is required")
	}

	if cfg.TaskManager == nil {
		return nil, fmt.Errorf("task manager is required")
	}

	// Set up capabilities based on config
	capabilities := a2a.AgentCapabilities{
		Streaming:              cfg.EnableStreaming,
		PushNotifications:      cfg.EnablePushNotifications,
		StateTransitionHistory: cfg.EnableStateHistory,
	}

	// Update agent card with actual capabilities
	cfg.AgentCard.Capabilities = capabilities

	s := &Server{
		taskManager:    cfg.TaskManager,
		mux:            http.NewServeMux(),
		agentCard:      cfg.AgentCard,
		streamRegistry: NewStreamRegistry(),
		notifier:       NewNotifier(),
		capabilities:   capabilities,
	}

	// Register handlers
	s.registerHandlers()

	return s, nil
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// registerHandlers sets up all the HTTP routes for the A2A server
func (s *Server) registerHandlers() {
	// Agent card endpoint
	s.mux.HandleFunc("GET /.well-known/agent.json", s.handleAgentCard)

	// A2A endpoints
	s.mux.HandleFunc("POST /", s.handleA2ARequest)
}

// handleAgentCard serves the agent card
func (s *Server) handleAgentCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := sonic.ConfigDefault.NewEncoder(w).Encode(s.agentCard); err != nil {
		http.Error(w, "Failed to encode agent card", http.StatusInternalServerError)
	}
}

// handleA2ARequest handles all JSON-RPC requests
func (s *Server) handleA2ARequest(w http.ResponseWriter, r *http.Request) {
	var req jsonrpc2.Request
	if err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, jsonrpc2.ErrParse)
		return
	}
	defer r.Body.Close()

	// Route to appropriate handler based on method
	switch req.Method {
	case a2a.MethodTasksSend:
		s.handleTasksSend(w, req)
	case a2a.MethodTasksGet:
		s.handleTasksGet(w, req)
	case a2a.MethodTasksCancel:
		s.handleTasksCancel(w, req)
	case a2a.MethodTasksPushNotificationSet:
		if !s.capabilities.PushNotifications {
			_, err := jsonrpc2.NewResponse(req.ID, nil, a2a.ErrorPushNotificationNotSupported)
			s.sendErrorResponse(w, err)
		} else {
			s.handleTasksPushNotificationSet(w, req)
		}
	case a2a.MethodTasksPushNotificationGet:
		if !s.capabilities.PushNotifications {
			_, err := jsonrpc2.NewResponse(req.ID, nil, a2a.ErrorPushNotificationNotSupported)
			s.sendErrorResponse(w, err)
		} else {
			s.handleTasksPushNotificationGet(w, req)
		}
	case a2a.MethodTasksSendSubscribe:
		s.handleTasksSendSubscribe(w, r, req)
	case a2a.MethodTasksResubscribe:
		s.handleTasksResubscribe(w, r, req)
	default:
		_, err := jsonrpc2.NewResponse(req.ID, nil, jsonrpc2.ErrMethodNotFound)
		s.sendErrorResponse(w, err)
	}
}

// sendResponse sends a JSON-RPC response
func (s *Server) sendResponse(w http.ResponseWriter, resp *jsonrpc2.Response) {
	w.Header().Set("Content-Type", "application/json")
	if err := sonic.ConfigDefault.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// sendErrorResponse sends a JSON-RPC error response
func (s *Server) sendErrorResponse(w http.ResponseWriter, err error) {
	resp := &jsonrpc2.Response{
		Error: err,
	}
	s.sendResponse(w, resp)
}

// validateTaskExistence checks if a task exists and returns appropriate error if not
func (s *Server) validateTaskExistence(taskID string) (*a2a.Task, error) {
	task, err := s.taskManager.GetTask(taskID)
	if err != nil {
		return nil, a2a.ErrorTaskNotFound
	}
	return task, nil
}
