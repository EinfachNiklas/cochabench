package challenge

import (
	"strings"
	"testing"
)

func TestManifest_toString(t *testing.T) {
	tests := []struct {
		name       string
		manifest   Manifest
		tag        string
		contains   []string
		exact      string
	}{
		{
			name:     "EmptyChallenges",
			manifest: Manifest{Challenges: map[string]Challenge{}},
			tag:      "v1.0",
			exact:    "No challenges available.",
		},
		{
			name:     "NilChallenges",
			manifest: Manifest{Challenges: nil},
			tag:      "v1.0",
			exact:    "No challenges available.",
		},
		{
			name: "SingleChallenge",
			manifest: Manifest{Challenges: map[string]Challenge{
				"001": {Title: "Foo", Language: "go", Difficulty: "easy"},
			}},
			tag:      "v1.0",
			contains: []string{"Release: v1.0", "001", "Foo", "go", "easy"},
		},
		{
			name: "SortedByID",
			manifest: Manifest{Challenges: map[string]Challenge{
				"003": {Title: "C"},
				"001": {Title: "A"},
				"002": {Title: "B"},
			}},
			tag: "v1.0",
		},
		{
			name: "ReleaseTag",
			manifest: Manifest{Challenges: map[string]Challenge{
				"001": {Title: "X"},
			}},
			tag:      "v2.3.1",
			contains: []string{"Release: v2.3.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.manifest.toString(tt.tag)

			if tt.exact != "" {
				if got != tt.exact {
					t.Errorf("toString() = %q, want %q", got, tt.exact)
				}
				return
			}

			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("toString() missing %q in output:\n%s", s, got)
				}
			}

			// Verify sort order for SortedByID case
			if tt.name == "SortedByID" {
				idx1 := strings.Index(got, "001")
				idx2 := strings.Index(got, "002")
				idx3 := strings.Index(got, "003")
				if idx1 >= idx2 || idx2 >= idx3 {
					t.Errorf("IDs not in sorted order: 001@%d, 002@%d, 003@%d\n%s", idx1, idx2, idx3, got)
				}
			}
		})
	}
}
