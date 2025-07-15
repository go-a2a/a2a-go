// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package a2a provides constants for the Agent-to-Agent (A2A) protocol.
// This file contains well-known path and URL constants used throughout the A2A SDK,
// mirroring the Python A2A constants for protocol compatibility.
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
