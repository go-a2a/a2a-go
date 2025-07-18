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

// A2A protocol path and URL constants.
// These constants define standard paths and URLs used in the A2A protocol
// for agent card resolution, authentication, and JSON-RPC endpoints.
const (
	// AgentCardWellKnownPath is the standard path for retrieving an agent's public AgentCard.
	// This path follows the well-known URI pattern and is used by A2A card resolvers
	// to fetch agent cards from remote agents.
	//
	// Example usage: https://agent.example.com/.well-known/agent.json
	AgentCardWellKnownPath = "/.well-known/agent.json"

	// ExtendedAgentCardPath is the path for an authenticated extended agent card.
	// This path is used to retrieve additional agent information that requires
	// authentication and is conditionally available based on agent capabilities.
	//
	// Example usage: https://agent.example.com/agent/authenticatedExtendedCard
	ExtendedAgentCardPath = "/agent/authenticatedExtendedCard"

	// DefaultRPCURL is the default URL path for the A2A JSON-RPC endpoint.
	// This path handles POST requests for JSON-RPC communication between agents
	// and is used as the default endpoint for A2A protocol operations.
	//
	// Example usage: https://agent.example.com/
	DefaultRPCURL = "/"
)
