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

package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"

	a2a "github.com/go-a2a/a2a-go"
)

// CardResolver represents an Agent Card resolver.
type CardResolver interface {
	// GetAgentCard fetches an agent card from a specified path relative to the baseURL.
	//
	// If relativeCardPath is empty, it defaults to the resolver's configured
	// agentCardPath (for the public agent card).
	GetAgentCard(ctx context.Context, relativeCardPath string) (*a2a.AgentCard, error)
}

type cardResolver struct {
	hc            *http.Client
	baseURL       string
	agentCardPath string
}

var _ CardResolver = (*cardResolver)(nil)

func NewCardResolver(baseURL, agentCardPath string) *cardResolver {
	return &cardResolver{
		hc:            &http.Client{},
		baseURL:       strings.TrimRight(baseURL, "/"),
		agentCardPath: strings.TrimLeft(a2a.AgentCardWellKnownPath, "/"),
	}
}

func (r *cardResolver) GetAgentCard(ctx context.Context, relativeCardPath string) (*a2a.AgentCard, error) {
	if relativeCardPath == "" {
		relativeCardPath = r.agentCardPath
	} else {
		relativeCardPath = strings.TrimLeft(relativeCardPath, "/")
	}

	targetURL := fmt.Sprintf("%s/%s", r.baseURL, relativeCardPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	invoker := func(_ context.Context, req *http.Request) (*http.Response, error) {
		return r.hc.Do(req)
	}
	resp, err := invoker(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fetch agent card: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch agent card from %s", targetURL)
	}

	var agentCard a2a.AgentCard
	dec := jsontext.NewDecoder(resp.Body)
	if err := json.UnmarshalDecode(dec, &agentCard, json.DefaultOptionsV2()); err != nil {
		return nil, fmt.Errorf("decode agent card: %w", err)
	}

	return &agentCard, nil
}
