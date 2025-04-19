// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package a2a_test

import (
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
	gocmpopts "github.com/google/go-cmp/cmp/cmpopts"

	"github.com/go-a2a/a2a"
)

func TestID(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		v     any
		want  a2a.ID
		wantS string
	}{
		"string": {
			v:     "abc123",
			want:  a2a.NewID("abc123"),
			wantS: "abc123",
		},
		"int32": {
			v:     int32(3),
			want:  a2a.NewID(int32(3)),
			wantS: "3",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var id a2a.ID
			switch v := tt.v.(type) {
			case string:
				id = a2a.NewID(v)
			case int32:
				id = a2a.NewID(v)
			}

			if diff := gocmp.Diff(tt.want, id, gocmpopts.EquateComparable(a2a.ID{})); diff != "" {
				t.Errorf("ID: (-want +got):\n%s", diff)
			}

			if got, want := tt.wantS, id.String(); got != want {
				t.Fatalf("got %s but want %s", got, want)
			}
		})
	}
}

func TestIDMarshaling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       a2a.ID
		expected string
	}{
		{
			name:     "string",
			id:       a2a.NewID("abc123"),
			expected: `"abc123"`,
		},
		{
			name:     "int32",
			id:       a2a.NewID(int32(3)),
			expected: `3`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := tt.id.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}

			if diff := gocmp.Diff(tt.expected, string(data)); diff != "" {
				t.Errorf("MarshalJSON(): (-want +got):\n%s", diff)
			}

			var newID a2a.ID
			if err := newID.UnmarshalJSON(data); err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}

			if diff := gocmp.Diff(tt.id, newID, gocmpopts.EquateComparable(a2a.ID{})); diff != "" {
				t.Errorf("UnmarshalJSON(): (-want +got):\n%s", diff)
			}
		})
	}
}
