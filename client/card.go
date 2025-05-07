// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/go-a2a/a2a"
)

const (
	// AgentCardPath is the well-known path where agent cards are published.
	AgentCardPath = "/.well-known/agent.json"
)

// FetchAgentCard fetches an agent card from the specified URL.
func FetchAgentCard(ctx context.Context, baseURL string, httpClient *http.Client) (*a2a.AgentCard, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	// Build the agent card URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL: %w", err)
	}
	u.Path = path.Join(u.Path, AgentCardPath)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	// Send request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned non-OK status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	// Decode agent card
	var card a2a.AgentCard
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, fmt.Errorf("decoding agent card: %w", err)
	}

	return &card, nil
}

// ValidateAgentCard validates an agent card.
func ValidateAgentCard(card *a2a.AgentCard) error {
	if card == nil {
		return fmt.Errorf("agent card is nil")
	}

	if card.Name == "" {
		return fmt.Errorf("agent card missing required field: name")
	}

	if card.URL == "" {
		return fmt.Errorf("agent card missing required field: url")
	}

	if card.Version == "" {
		return fmt.Errorf("agent card missing required field: version")
	}

	if len(card.Skills) == 0 {
		return fmt.Errorf("agent card has no skills")
	}

	for i, skill := range card.Skills {
		if skill.ID == "" {
			return fmt.Errorf("skill #%d missing required field: id", i+1)
		}
		if skill.Name == "" {
			return fmt.Errorf("skill #%d missing required field: name", i+1)
		}
	}

	return nil
}

// FindSkill finds a skill by ID in an agent card.
func FindSkill(card *a2a.AgentCard, skillID string) (*a2a.AgentSkill, bool) {
	for _, skill := range card.Skills {
		if skill.ID == skillID {
			return &skill, true
		}
	}
	return nil, false
}

// GetSupportedInputModes returns the input modes supported by a skill.
// If the skill has no specific input modes, the agent's default input modes are returned.
func GetSupportedInputModes(card *a2a.AgentCard, skillID string) []string {
	skill, found := FindSkill(card, skillID)
	if !found {
		return card.DefaultInputModes
	}

	if len(skill.InputModes) > 0 {
		return skill.InputModes
	}

	return card.DefaultInputModes
}

// GetSupportedOutputModes returns the output modes supported by a skill.
// If the skill has no specific output modes, the agent's default output modes are returned.
func GetSupportedOutputModes(card *a2a.AgentCard, skillID string) []string {
	skill, found := FindSkill(card, skillID)
	if !found {
		return card.DefaultOutputModes
	}

	if len(skill.OutputModes) > 0 {
		return skill.OutputModes
	}

	return card.DefaultOutputModes
}
