package config

import (
	"path/filepath"
	"testing"
)

func TestConfigFields(t *testing.T) {
	cfg := Config{
		LLM_PROVIDER:     "anthropic",
		LLM_BASE_PATH:    "https://api.example.com",
		LLM_MODEL:        "claude-sonnet-4-6",
		CHALLENGE_SERVER: "https://github.com/example",
	}

	fields := configFields(&cfg)

	// Verify all 4 keys exist
	expectedKeys := []string{"LLM_PROVIDER", "LLM_BASE_PATH", "LLM_MODEL", "CHALLENGE_SERVER"}
	if len(fields) != len(expectedKeys) {
		t.Fatalf("configFields() returned %d entries, want %d", len(fields), len(expectedKeys))
	}
	for _, key := range expectedKeys {
		if _, ok := fields[key]; !ok {
			t.Errorf("missing key %q in configFields()", key)
		}
	}

	// Verify pointer values match struct fields
	if *fields["LLM_PROVIDER"] != "anthropic" {
		t.Errorf("LLM_PROVIDER = %q, want %q", *fields["LLM_PROVIDER"], "anthropic")
	}
	if *fields["LLM_BASE_PATH"] != "https://api.example.com" {
		t.Errorf("LLM_BASE_PATH = %q, want %q", *fields["LLM_BASE_PATH"], "https://api.example.com")
	}
	if *fields["LLM_MODEL"] != "claude-sonnet-4-6" {
		t.Errorf("LLM_MODEL = %q, want %q", *fields["LLM_MODEL"], "claude-sonnet-4-6")
	}
	if *fields["CHALLENGE_SERVER"] != "https://github.com/example" {
		t.Errorf("CHALLENGE_SERVER = %q, want %q", *fields["CHALLENGE_SERVER"], "https://github.com/example")
	}

	// Mutation test: writing through pointer changes the struct
	*fields["LLM_MODEL"] = "new-model"
	if cfg.LLM_MODEL != "new-model" {
		t.Errorf("mutation through pointer failed: LLM_MODEL = %q, want %q", cfg.LLM_MODEL, "new-model")
	}
}

func TestGetConfigPath(t *testing.T) {
	path, err := getConfigPath()
	if err != nil {
		t.Fatalf("getConfigPath() error: %v", err)
	}

	if filepath.Base(path) != "config.json" {
		t.Errorf("expected filename config.json, got %q", filepath.Base(path))
	}
	if filepath.Base(filepath.Dir(path)) != "cochabench" {
		t.Errorf("expected parent dir cochabench, got %q", filepath.Base(filepath.Dir(path)))
	}
}
