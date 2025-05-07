// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"
	"time"

	"github.com/bytedance/sonic"

	"github.com/go-a2a/a2a"
	"github.com/go-a2a/a2a/internal/jsonrpc2"
)

// handleTasksSend handles the tasks/send method
func (s *Server) handleTasksSend(w http.ResponseWriter, req jsonrpc2.Request) {
	var params a2a.TaskSendParams
	if err := sonic.ConfigDefault.Unmarshal(req.Params, &params); err != nil {
		if err != nil {
			s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
			return
		}
	}

	// Validate required fields
	if params.ID == "" || len(params.Message.Parts) == 0 {
		s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
		return
	}

	// Create a new task or update an existing one
	task := &a2a.Task{
		ID:        params.ID,
		SessionID: params.SessionID,
		Status: a2a.TaskStatus{
			State:     a2a.TaskStateSubmitted,
			Message:   &params.Message,
			Timestamp: time.Now(),
		},
		Metadata: params.Metadata,
	}

	// Process the task
	processedTask, err := s.taskManager.ProcessTask(task)
	if err != nil {
		s.sendErrorResponse(w, err)
		return
	}

	// Set up push notification if requested
	if params.PushNotification != nil && s.capabilities.PushNotifications {
		if err := s.notifier.SetPushNotification(params.ID, params.PushNotification); err != nil {
			// Log error but don't fail the request
		}
	}

	// Include history if requested
	if params.HistoryLength > 0 && s.capabilities.StateTransitionHistory {
		history, err := s.taskManager.GetTaskHistory(params.ID, params.HistoryLength)
		if err == nil {
			processedTask.History = history
		}
	} else {
		// Don't include history if not requested
		processedTask.History = nil
	}

	resp, err := jsonrpc2.NewResponse(req.ID, processedTask, nil)
	if err != nil {
		s.sendErrorResponse(w, err)
		return
	}
	s.sendResponse(w, resp)
}

// handleTasksGet handles the tasks/get method
func (s *Server) handleTasksGet(w http.ResponseWriter, req jsonrpc2.Request) {
	var params a2a.TaskQueryParams
	if err := sonic.ConfigDefault.Unmarshal(req.Params, &params); err != nil {
		s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
		return
	}

	// Validate required fields
	if params.ID == "" {
		s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
		return
	}

	// Get the task
	task, err := s.taskManager.GetTask(params.ID)
	if err != nil {
		s.sendErrorResponse(w, err)
		return
	}

	// Include history if requested
	if params.HistoryLength > 0 && s.capabilities.StateTransitionHistory {
		history, err := s.taskManager.GetTaskHistory(params.ID, params.HistoryLength)
		if err == nil {
			task.History = history
		}
	} else {
		// Don't include history if not requested
		task.History = nil
	}

	resp, err := jsonrpc2.NewResponse(req.ID, task, nil)
	if err != nil {
		s.sendErrorResponse(w, err)
		return
	}
	s.sendResponse(w, resp)
}

// handleTasksCancel handles the tasks/cancel method
func (s *Server) handleTasksCancel(w http.ResponseWriter, req jsonrpc2.Request) {
	var params a2a.TaskIdParams
	if err := sonic.ConfigDefault.Unmarshal(req.Params, &params); err != nil {
		s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
		return
	}

	// Validate required fields
	if params.ID == "" {
		s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
		return
	}

	// Cancel the task
	task, err := s.taskManager.CancelTask(params.ID)
	if err != nil {
		s.sendErrorResponse(w, err)
		return
	}

	// Notify via push notification if configured
	if s.capabilities.PushNotifications {
		s.notifier.SendStatusUpdate(params.ID, task.Status, true)
	}

	resp, err := jsonrpc2.NewResponse(req.ID, task, nil)
	if err != nil {
		s.sendErrorResponse(w, err)
		return
	}
	s.sendResponse(w, resp)
}

// handleTasksPushNotificationSet handles the tasks/pushNotification/set method
func (s *Server) handleTasksPushNotificationSet(w http.ResponseWriter, req jsonrpc2.Request) {
	var params a2a.TaskPushNotificationConfig
	if err := sonic.ConfigDefault.Unmarshal(req.Params, &params); err != nil {
		s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
		return
	}

	// Validate required fields
	if params.ID == "" || params.PushNotificationConfig.URL == "" {
		s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
		return
	}

	// Check if task exists
	_, err := s.validateTaskExistence(params.ID)
	if err != nil {
		s.sendErrorResponse(w, err)
		return
	}

	// Set push notification config
	if err := s.notifier.SetPushNotification(params.ID, &params.PushNotificationConfig); err != nil {
		s.sendErrorResponse(w, jsonrpc2.ErrInternal)
		return
	}

	resp, err := jsonrpc2.NewResponse(req.ID, &params, nil)
	if err != nil {
		s.sendErrorResponse(w, err)
		return
	}
	s.sendResponse(w, resp)
}

// handleTasksPushNotificationGet handles the tasks/pushNotification/get method
func (s *Server) handleTasksPushNotificationGet(w http.ResponseWriter, req jsonrpc2.Request) {
	var params a2a.TaskIdParams
	if err := sonic.ConfigDefault.Unmarshal(req.Params, &params); err != nil {
		s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
		return
	}

	// Validate required fields
	if params.ID == "" {
		s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
		return
	}

	// Check if task exists
	if _, err := s.validateTaskExistence(params.ID); err != nil {
		s.sendErrorResponse(w, err)
		return
	}

	// Get push notification config
	config := s.notifier.GetPushNotification(params.ID)
	if config == nil {
		s.sendErrorResponse(w, a2a.ErrorPushNotificationNotSupported)
		return
	}

	result := &a2a.TaskPushNotificationConfig{
		ID:                     params.ID,
		PushNotificationConfig: *config,
	}

	resp, err := jsonrpc2.NewResponse(req.ID, result, nil)
	if err != nil {
		s.sendErrorResponse(w, err)
		return
	}
	s.sendResponse(w, resp)
}

// handleTasksSendSubscribe handles the tasks/sendSubscribe streaming method
func (s *Server) handleTasksSendSubscribe(w http.ResponseWriter, r *http.Request, req jsonrpc2.Request) {
	var params a2a.TaskSendParams
	if err := sonic.ConfigDefault.Unmarshal(req.Params, &params); err != nil {
		s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
		return
	}

	// Validate required fields
	if params.ID == "" || len(params.Message.Parts) == 0 {
		s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
		return
	}

	// Create a stream for the task
	stream := s.streamRegistry.CreateStream(params.ID, w, r)
	if stream == nil {
		s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
		return
	}

	// Create a task for processing
	task := &a2a.Task{
		ID:        params.ID,
		SessionID: params.SessionID,
		Status: a2a.TaskStatus{
			State:     a2a.TaskStateSubmitted,
			Message:   &params.Message,
			Timestamp: time.Now(),
		},
		Metadata: params.Metadata,
	}

	// Set up push notification if requested
	if params.PushNotification != nil && s.capabilities.PushNotifications {
		if err := s.notifier.SetPushNotification(params.ID, params.PushNotification); err != nil {
			// Log error but don't fail the request
		}
	}

	// Start processing the task with streaming updates
	if err := s.taskManager.StartTask(task, stream); err != nil {
		s.sendErrorResponse(w, err)
		return
	}
}

// handleTasksResubscribe handles the tasks/resubscribe streaming method
func (s *Server) handleTasksResubscribe(w http.ResponseWriter, r *http.Request, req jsonrpc2.Request) {
	var params a2a.TaskQueryParams
	if err := sonic.ConfigDefault.Unmarshal(req.Params, &params); err != nil {
		s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
		return
	}

	// Validate required fields
	if params.ID == "" {
		s.sendErrorResponse(w, jsonrpc2.ErrInvalidParams)
		return
	}

	// Get the task
	task, err := s.taskManager.GetTask(params.ID)
	if err != nil {
		s.sendErrorResponse(w, err)
		return
	}

	// Create a stream for the task
	stream := s.streamRegistry.CreateStream(params.ID, w, r)
	if stream == nil {
		s.sendErrorResponse(w, jsonrpc2.ErrInternal)
		return
	}

	// Resume sending updates
	if err := s.taskManager.ResumeTaskUpdates(task, stream); err != nil {
		s.sendErrorResponse(w, err)
		return
	}
}
