// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"testing"

	"github.com/go-a2a/a2a"
	"github.com/google/go-cmp/cmp"
)

func TestValidateAgentCard(t *testing.T) {
	tests := []struct {
		name    string
		card    *a2a.AgentCard
		wantErr bool
	}{
		{
			name:    "nil card",
			card:    nil,
			wantErr: true,
		},
		{
			name: "missing name",
			card: &a2a.AgentCard{
				URL:     "https://example.com",
				Version: "1.0",
				Skills:  []a2a.AgentSkill{{ID: "skill1", Name: "Skill 1"}},
			},
			wantErr: true,
		},
		{
			name: "missing url",
			card: &a2a.AgentCard{
				Name:    "Test Agent",
				Version: "1.0",
				Skills:  []a2a.AgentSkill{{ID: "skill1", Name: "Skill 1"}},
			},
			wantErr: true,
		},
		{
			name: "missing version",
			card: &a2a.AgentCard{
				Name:   "Test Agent",
				URL:    "https://example.com",
				Skills: []a2a.AgentSkill{{ID: "skill1", Name: "Skill 1"}},
			},
			wantErr: true,
		},
		{
			name: "no skills",
			card: &a2a.AgentCard{
				Name:    "Test Agent",
				URL:     "https://example.com",
				Version: "1.0",
				Skills:  []a2a.AgentSkill{},
			},
			wantErr: true,
		},
		{
			name: "skill missing id",
			card: &a2a.AgentCard{
				Name:    "Test Agent",
				URL:     "https://example.com",
				Version: "1.0",
				Skills:  []a2a.AgentSkill{{Name: "Skill 1"}},
			},
			wantErr: true,
		},
		{
			name: "skill missing name",
			card: &a2a.AgentCard{
				Name:    "Test Agent",
				URL:     "https://example.com",
				Version: "1.0",
				Skills:  []a2a.AgentSkill{{ID: "skill1"}},
			},
			wantErr: true,
		},
		{
			name: "valid card",
			card: &a2a.AgentCard{
				Name:    "Test Agent",
				URL:     "https://example.com",
				Version: "1.0",
				Skills:  []a2a.AgentSkill{{ID: "skill1", Name: "Skill 1"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAgentCard(tt.card)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAgentCard() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFindSkill(t *testing.T) {
	card := &a2a.AgentCard{
		Skills: []a2a.AgentSkill{
			{ID: "skill1", Name: "Skill 1"},
			{ID: "skill2", Name: "Skill 2"},
		},
	}

	tests := []struct {
		name     string
		skillID  string
		want     *a2a.AgentSkill
		wantBool bool
	}{
		{
			name:     "existing skill",
			skillID:  "skill1",
			want:     &a2a.AgentSkill{ID: "skill1", Name: "Skill 1"},
			wantBool: true,
		},
		{
			name:     "non-existent skill",
			skillID:  "skill3",
			want:     nil,
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotBool := FindSkill(card, tt.skillID)
			if gotBool != tt.wantBool {
				t.Errorf("FindSkill() gotBool = %v, want %v", gotBool, tt.wantBool)
				return
			}
			if gotBool && !cmp.Equal(got, tt.want) {
				t.Errorf("FindSkill() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSupportedInputModes(t *testing.T) {
	card := &a2a.AgentCard{
		DefaultInputModes: []string{"text"},
		Skills: []a2a.AgentSkill{
			{ID: "skill1", Name: "Skill 1", InputModes: []string{"text", "file"}},
			{ID: "skill2", Name: "Skill 2"},
		},
	}

	tests := []struct {
		name    string
		skillID string
		want    []string
	}{
		{
			name:    "skill with custom input modes",
			skillID: "skill1",
			want:    []string{"text", "file"},
		},
		{
			name:    "skill with default input modes",
			skillID: "skill2",
			want:    []string{"text"},
		},
		{
			name:    "non-existent skill",
			skillID: "skill3",
			want:    []string{"text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSupportedInputModes(card, tt.skillID)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("GetSupportedInputModes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSupportedOutputModes(t *testing.T) {
	card := &a2a.AgentCard{
		DefaultOutputModes: []string{"text"},
		Skills: []a2a.AgentSkill{
			{ID: "skill1", Name: "Skill 1", OutputModes: []string{"text", "file"}},
			{ID: "skill2", Name: "Skill 2"},
		},
	}

	tests := []struct {
		name    string
		skillID string
		want    []string
	}{
		{
			name:    "skill with custom output modes",
			skillID: "skill1",
			want:    []string{"text", "file"},
		},
		{
			name:    "skill with default output modes",
			skillID: "skill2",
			want:    []string{"text"},
		},
		{
			name:    "non-existent skill",
			skillID: "skill3",
			want:    []string{"text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSupportedOutputModes(card, tt.skillID)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("GetSupportedOutputModes() = %v, want %v", got, tt.want)
			}
		})
	}
}
