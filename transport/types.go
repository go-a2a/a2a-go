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

package transport

import (
	"fmt"

	a2a "github.com/go-a2a/a2a-go"
)

// messageOrTaskResult is a wrapper that can unmarshal JSON into either Message or Task
// based on the "kind" discriminator field, implementing the MessageOrTask interface.
type messageOrTaskResult struct {
	result a2a.MessageOrTask
}

var (
	_ a2a.Result        = (*messageOrTaskResult)(nil)
	_ a2a.MessageOrTask = (*messageOrTaskResult)(nil)
)

func (m *messageOrTaskResult) GetTaskID() string {
	if m.result != nil {
		return m.result.GetTaskID()
	}
	return ""
}

// GetKind implements the MessageOrTask interface.
func (m *messageOrTaskResult) GetKind() a2a.EventKind {
	if m.result != nil {
		return m.result.GetKind()
	}
	return a2a.EventKind("")
}

// UnmarshalJSON implements json.Unmarshaler for [messageOrTaskResult].
func (m *messageOrTaskResult) UnmarshalJSON(data []byte) error {
	result, err := a2a.UnmarshalMessageOrTask(data)
	if err != nil {
		return fmt.Errorf("messageOrTaskResult.UnmarshalJSON: %w", err)
	}
	if result == nil {
		return fmt.Errorf("messageOrTaskResult.UnmarshalJSON: unmarshaled nil result")
	}
	m.result = result
	return nil
}

// sendStreamingMessageResponse is a wrapper that can unmarshal JSON into either Message or Task
// based on the "kind" discriminator field, implementing the [a2a.SendStreamingMessageResponse] interface.
type sendStreamingMessageResponse struct {
	result a2a.SendStreamingMessageResponse
}

var (
	_ a2a.Result                       = (*sendStreamingMessageResponse)(nil)
	_ a2a.SendStreamingMessageResponse = (*sendStreamingMessageResponse)(nil)
)

func (m *sendStreamingMessageResponse) GetTaskID() string {
	if m.result != nil {
		return m.result.GetTaskID()
	}
	return ""
}

// GetKind implements the MessageOrTask interface.
func (m *sendStreamingMessageResponse) GetEventKind() a2a.EventKind {
	if m.result != nil {
		return m.result.GetEventKind()
	}
	return a2a.EventKind("")
}

// UnmarshalJSON implements json.Unmarshaler for [sendStreamingMessageResponse].
func (m *sendStreamingMessageResponse) UnmarshalJSON(data []byte) error {
	result, err := a2a.UnmarshalSendStreamingMessageResponse(data)
	if err != nil {
		return fmt.Errorf("sendStreamingMessageResponse.UnmarshalJSON: %w", err)
	}
	if result == nil {
		return fmt.Errorf("sendStreamingMessageResponse.UnmarshalJSON: unmarshaled nil result")
	}
	m.result = result
	return nil
}
