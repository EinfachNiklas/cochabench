package tools

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

func TestTableBuilder_Build(t *testing.T) {
	tests := []struct {
		name     string
		headers  []string
		rows     [][]string
		contains []string
		exact    string
	}{
		{
			name:    "EmptyHeaders",
			headers: []string{},
			exact:   "",
		},
		{
			name:     "HeadersOnly",
			headers:  []string{"ID", "Name"},
			contains: []string{"ID", "Name", "-+-"},
		},
		{
			name:     "SingleRow",
			headers:  []string{"A", "B"},
			rows:     [][]string{{"x", "y"}},
			contains: []string{"A", "B", "x", "y", " | ", "-+-"},
		},
		{
			name:     "ColumnWidthExpands",
			headers:  []string{"ID"},
			rows:     [][]string{{"long-value"}},
			contains: []string{"ID        ", "long-value"},
		},
		{
			name:     "MultipleRows",
			headers:  []string{"K", "V"},
			rows:     [][]string{{"a", "1"}, {"b", "2"}, {"c", "3"}},
			contains: []string{"a", "b", "c", "1", "2", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tb := NewTableBuilder(tt.headers)
			for _, row := range tt.rows {
				tb.AddRow(row)
			}
			got := tb.Build()

			if tt.exact != "" || (len(tt.headers) == 0) {
				if got != tt.exact {
					t.Errorf("Build() = %q, want %q", got, tt.exact)
				}
				return
			}

			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("Build() missing %q in output:\n%s", s, got)
				}
			}
		})
	}
}

func TestTableBuilder_Build_HeadersOnly_NoDataRows(t *testing.T) {
	tb := NewTableBuilder([]string{"ID", "Name"})
	got := tb.Build()

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines (header + separator), got %d:\n%s", len(lines), got)
	}
}

func TestValidateDirPath(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "ValidDir",
			setup:   func(t *testing.T) string { return t.TempDir() },
			wantErr: false,
		},
		{
			name: "NonExistent",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nope")
			},
			wantErr:   true,
			errSubstr: "Directory does not exist",
		},
		{
			name: "IsFile",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				p := filepath.Join(dir, "file.txt")
				if err := os.WriteFile(p, []byte("hi"), 0644); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr:   true,
			errSubstr: "not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			err := ValidateDirPath(path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q missing substring %q", err, tt.errSubstr)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func setupChallengeDir(t *testing.T, createSrc, createTest, createConfig bool) string {
	t.Helper()
	dir := t.TempDir()
	if createSrc {
		if err := os.Mkdir(filepath.Join(dir, "src"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if createTest {
		if err := os.Mkdir(filepath.Join(dir, "test"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if createConfig {
		if err := os.WriteFile(filepath.Join(dir, "challenge.config.json"), []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestValidateDirStructure(t *testing.T) {
	tests := []struct {
		name              string
		src, test, config bool
		wantErr           bool
		errSubstr         string
	}{
		{"Valid", true, true, true, false, ""},
		{"MissingSrc", false, true, true, true, "missing required folder: src"},
		{"MissingTest", true, false, true, true, "missing required folder: test"},
		{"MissingConfig", true, true, false, true, "missing required file: challenge.config.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupChallengeDir(t, tt.src, tt.test, tt.config)
			err := ValidateDirStructure(dir)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q missing substring %q", err, tt.errSubstr)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoadChallengeConfig(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		writeFile bool
		wantErr   bool
		errSubstr string
		wantName  string
		wantID    string
		wantType  string
	}{
		{
			name:      "ValidJSON",
			content:   `{"Name":"foo","ChallengeID":"c1","ChallengeType":"go"}`,
			writeFile: true,
			wantName:  "foo",
			wantID:    "c1",
			wantType:  "go",
		},
		{
			name:      "MalformedJSON",
			content:   `{bad json`,
			writeFile: true,
			wantErr:   true,
			errSubstr: "challenge.config.json is invalid",
		},
		{
			name:      "FileNotFound",
			writeFile: false,
			wantErr:   true,
		},
		{
			name:      "EmptyJSON",
			content:   `{}`,
			writeFile: true,
			wantName:  "",
			wantID:    "",
			wantType:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.writeFile {
				dir := t.TempDir()
				path = filepath.Join(dir, "config.json")
				if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
					t.Fatal(err)
				}
			} else {
				path = filepath.Join(t.TempDir(), "nonexistent.json")
			}

			cfg, err := LoadChallengeConfig(path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q missing substring %q", err, tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", cfg.Name, tt.wantName)
			}
			if cfg.ChallengeID != tt.wantID {
				t.Errorf("ChallengeID = %q, want %q", cfg.ChallengeID, tt.wantID)
			}
			if cfg.ChallengeType != tt.wantType {
				t.Errorf("ChallengeType = %q, want %q", cfg.ChallengeType, tt.wantType)
			}
		})
	}
}

func TestLoadEnv(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		t.Setenv("LLM_API_KEY", "test-key-123")
		env, err := LoadEnv()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if env.LLM_API_KEY != "test-key-123" {
			t.Errorf("LLM_API_KEY = %q, want %q", env.LLM_API_KEY, "test-key-123")
		}
	})

	t.Run("Empty", func(t *testing.T) {
		t.Setenv("LLM_API_KEY", "")
		_, err := LoadEnv()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "LLM_API_KEY") {
			t.Errorf("error %q should mention LLM_API_KEY", err)
		}
	})
}

func TestFmtTime(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{"ZeroTime", time.Time{}, "-"},
		{"SpecificTime", time.Date(2024, 6, 15, 14, 30, 45, 0, time.Local), "2024-06-15 14:30:45"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			got := FmtTime(tt.t)
			if got != tt.want {
				t2.Errorf("FmtTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetBuildVersion(t *testing.T) {
	t.Run("UsesExternalVersionWhenProvided", func(t *testing.T) {
		got := GetBuildVersion("v1.2.3")
		if got != "v1.2.3" {
			t.Fatalf("GetBuildVersion() = %q, want %q", got, "v1.2.3")
		}
	})

	tests := []struct {
		name            string
		externalVersion string
		info            *debug.BuildInfo
		ok              bool
		want            string
	}{
		{
			name:            "ReturnsExternalVersionWithoutReadingBuildInfo",
			externalVersion: "v1.2.3",
			info: &debug.BuildInfo{
				Main: debug.Module{
					Version: "v9.9.9",
				},
			},
			ok:   true,
			want: "v1.2.3",
		},
		{
			name:            "UsesMainVersionFromBuildInfo",
			externalVersion: "dev",
			info: &debug.BuildInfo{
				Main: debug.Module{
					Version: "v0.1.3+dirty",
				},
			},
			ok:   true,
			want: "v0.1.3+dirty",
		},
		{
			name:            "FallsBackToShortVCSRevision",
			externalVersion: "dev",
			info: &debug.BuildInfo{
				Main: debug.Module{
					Version: "(devel)",
				},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abcdef1234567890"},
				},
			},
			ok:   true,
			want: "abcdef1",
		},
		{
			name:            "ReturnsShortRevisionWithoutTruncationWhenAlreadyShort",
			externalVersion: "dev",
			info: &debug.BuildInfo{
				Main: debug.Module{
					Version: "(devel)",
				},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc123"},
				},
			},
			ok:   true,
			want: "abc123",
		},
		{
			name:            "FallsBackToExternalVersionWithoutBuildInfo",
			externalVersion: "dev",
			info:            nil,
			ok:              false,
			want:            "dev",
		},
		{
			name:            "FallsBackToExternalVersionWhenBuildInfoHasNoUsefulVersion",
			externalVersion: "dev",
			info: &debug.BuildInfo{
				Main: debug.Module{
					Version: "(devel)",
				},
			},
			ok:   true,
			want: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveBuildVersion(tt.externalVersion, tt.info, tt.ok)
			if got != tt.want {
				t.Fatalf("resolveBuildVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}
