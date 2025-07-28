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

// // CardResolver defines the interface for resolving agent cards.
// type CardResolver interface {
// 	GetAgentCard(ctx context.Context, relativePath string, opts ...RequestOption) (*a2a.AgentCard, error)
// }
//
// // cardResolver implements CardResolver using HTTP requests.
// type cardResolver struct {
// 	httpClient    *http.Client
// 	baseURL       string
// 	agentCardPath string
// 	interceptors  []Interceptor
// }
//
// var _ CardResolver = (*cardResolver)(nil)
//
// // NewCardResolver creates a new HTTP-based card resolver.
// func NewCardResolver(baseURL string, opts ...ClientOption) *cardResolver {
// 	config := applyClientOptions(opts...)
//
// 	resolver := &cardResolver{
// 		httpClient:    config.httpClient,
// 		baseURL:       strings.TrimRight(baseURL, "/"),
// 		agentCardPath: strings.TrimLeft(a2a.AgentCardWellKnownPath, "/"),
// 		interceptors:  config.interceptors,
// 	}
//
// 	return resolver
// }
//
// // GetAgentCard fetches an agent card from a specified path relative to the baseURL.
// //
// // If relative_card_path is None, it defaults to the resolver's configured
// // agent_card_path (for the public agent card).
// func (r *cardResolver) GetAgentCard(ctx context.Context, relativeCardPath string, opts ...RequestOption) (*a2a.AgentCard, error) {
// 	if relativeCardPath == "" {
// 		relativeCardPath = r.agentCardPath
// 	}
// 	// Ensure the path starts with a slash
// 	if !strings.HasPrefix(relativeCardPath, "/") {
// 		relativeCardPath = "/" + relativeCardPath
// 	}
//
// 	// Remove leading slash from relative path if present
// 	relativeCardPath = strings.TrimLeft(relativeCardPath, "/")
//
// 	targetURL := fmt.Sprintf("%s/%s", r.baseURL, relativeCardPath)
//
// 	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
// 	if err != nil {
// 		return nil, NewConfigurationError("create request", err)
// 	}
//
// 	// Apply request options
// 	requestConfig := applyRequestOptions(opts...)
// 	for key, value := range requestConfig.headers {
// 		req.Header.Set(key, value)
// 	}
//
// 	// Set default headers
// 	req.Header.Set("Accept", "application/json")
// 	req.Header.Set("Content-Type", "application/json")
//
// 	// Apply interceptors
// 	invoker := func(ctx context.Context, req *http.Request) (*http.Response, error) {
// 		return r.httpClient.Do(req)
// 	}
// 	invoker = chainInterceptors(r.interceptors, invoker)
//
// 	resp, err := invoker(ctx, req)
// 	if err != nil {
// 		return nil, NewNetworkError("fetch agent card", err)
// 	}
// 	defer resp.Body.Close()
//
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, NewHTTPError(resp.StatusCode, fmt.Sprintf("fetch agent card from %s", targetURL), nil)
// 	}
//
// 	var agentCard a2a.AgentCard
// 	dec := jsontext.NewDecoder(resp.Body)
// 	if err := json.UnmarshalDecode(dec, &agentCard, json.DefaultOptionsV2()); err != nil {
// 		return nil, NewJSONError("decode agent card", err)
// 	}
//
// 	return &agentCard, nil
// }
//
// // GetPublicAgentCard fetches the public agent card using the default path.
// func (r *cardResolver) GetPublicAgentCard(ctx context.Context, opts ...RequestOption) (*a2a.AgentCard, error) {
// 	return r.GetAgentCard(ctx, "", opts...)
// }
//
// // CreateClientFromAgentCardURL creates a client from an agent card URL.
// func CreateClientFromAgentCardURL(ctx context.Context, baseURL string, opts ...ClientOption) (Client, error) {
// 	resolver := NewCardResolver(baseURL, opts...)
//
// 	agentCard, err := resolver.GetPublicAgentCard(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("fetch agent card: %w", err)
// 	}
//
// 	// Create client with the agent card
// 	clientOpts := append(opts, WithAgentCard(agentCard))
// 	return NewHTTPClient(clientOpts...), nil
// }
//
// // WithAgentCard sets the agent card for the client.
// func WithAgentCard(agentCard *a2a.AgentCard) ClientOption {
// 	return func(c *clientConfig) {
// 		// Store the agent card in the interceptors for now
// 		// In a real implementation, we'd extend clientConfig
// 		c.interceptors = append(c.interceptors, func(ctx context.Context, req *http.Request, invoker Invoker) (*http.Response, error) {
// 			// Add agent card to context
// 			callCtx := &ClientCallContext{
// 				Method:    req.Method,
// 				URL:       req.URL.String(),
// 				Headers:   make(map[string]string),
// 				Metadata:  make(map[string]any),
// 				AgentCard: agentCard,
// 			}
//
// 			for key, values := range req.Header {
// 				if len(values) > 0 {
// 					callCtx.Headers[key] = values[0]
// 				}
// 			}
//
// 			ctxWithCallCtx := WithClientCallContext(ctx, callCtx)
// 			return invoker(ctxWithCallCtx, req)
// 		})
// 	}
// }
