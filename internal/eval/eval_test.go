package eval

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantType  string
		wantErr   bool
		errSubstr string
	}{
		{"Go", "go", "GoHandler", false, ""},
		{"Golang", "golang", "GoHandler", false, ""},
		{"Javascript", "javascript", "JavascriptHandler", false, ""},
		{"JS", "js", "JavascriptHandler", false, ""},
		{"Python", "python", "PythonHandler", false, ""},
		{"Py", "py", "PythonHandler", false, ""},
		{"Unknown", "unknown", "", true, "Unsupported challenge type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := createHandler(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q missing substring %q", err, tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			typeName := typeString(handler)
			if typeName != tt.wantType {
				t.Errorf("createHandler(%q) type = %q, want %q", tt.input, typeName, tt.wantType)
			}
		})
	}
}

type stubLanguageHandler struct {
	executeTests func(context.Context, string) (*TestResult, error)
}

func (h stubLanguageHandler) PrepareEnvironment(challengePath string, runID string) (string, func(), error) {
	return "", func() {}, nil
}

func (h stubLanguageHandler) ExecuteTests(ctx context.Context, tempDir string) (*TestResult, error) {
	return h.executeTests(ctx, tempDir)
}

func TestExecuteTests_WrapsHandlerErrorOnce(t *testing.T) {
	handler := stubLanguageHandler{
		executeTests: func(ctx context.Context, tempDir string) (*TestResult, error) {
			return nil, errors.New("boom")
		},
	}

	_, err, timedOut := executeTests(handler, t.TempDir(), 5*time.Minute)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if timedOut {
		t.Fatal("timedOut = true, want false")
	}
	if !strings.Contains(err.Error(), "Test execution failed: boom") {
		t.Fatalf("error = %q, want wrapped handler error", err)
	}
	if strings.Count(err.Error(), "Test execution failed") != 1 {
		t.Fatalf("error = %q, expected single wrap", err)
	}
	if strings.Contains(err.Error(), "\n") {
		t.Fatalf("error = %q, should not contain newlines", err)
	}
}

func typeString(h LanguageHandler) string {
	switch h.(type) {
	case GoHandler:
		return "GoHandler"
	case JavascriptHandler:
		return "JavascriptHandler"
	case PythonHandler:
		return "PythonHandler"
	default:
		return "unknown"
	}
}

func TestIsBinary(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"NormalText", "hello world\nfoo bar", false},
		{"WithNullByte", "hello\x00world", true},
		{"EmptyString", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBinary(tt.input)
			if got != tt.want {
				t.Errorf("isBinary(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSortedKeys(t *testing.T) {
	t.Run("Unsorted", func(t *testing.T) {
		m := map[string]bool{"c": true, "a": true, "b": true}
		got := sortedKeys(m)
		if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
			t.Errorf("sortedKeys() = %v, want [a b c]", got)
		}
	})

	t.Run("Empty", func(t *testing.T) {
		got := sortedKeys(map[string]bool{})
		if len(got) != 0 {
			t.Errorf("sortedKeys(empty) = %v, want []", got)
		}
	})
}

func TestCollectFiles(t *testing.T) {
	t.Run("NestedFiles", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "sub"), 0755)
		os.WriteFile(filepath.Join(dir, "a.go"), []byte("a"), 0644)
		os.WriteFile(filepath.Join(dir, "sub", "b.go"), []byte("b"), 0644)

		files, err := collectFiles(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 2 {
			t.Fatalf("expected 2 files, got %d: %v", len(files), files)
		}
		if !files["a.go"] {
			t.Error("missing a.go")
		}
		if !files["sub/b.go"] {
			t.Error("missing sub/b.go")
		}
	})

	t.Run("EmptyDir", func(t *testing.T) {
		dir := t.TempDir()
		files, err := collectFiles(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("expected empty map, got %v", files)
		}
	})
}

func TestReadFileOrEmpty(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "exists.txt"), []byte("content"), 0644)

	t.Run("Exists", func(t *testing.T) {
		got := readFileOrEmpty(dir, "exists.txt")
		if got != "content" {
			t.Errorf("readFileOrEmpty() = %q, want %q", got, "content")
		}
	})

	t.Run("NotExists", func(t *testing.T) {
		got := readFileOrEmpty(dir, "nope.txt")
		if got != "" {
			t.Errorf("readFileOrEmpty() = %q, want empty", got)
		}
	})
}

func TestGenerateDiff(t *testing.T) {
	t.Run("IdenticalFiles", func(t *testing.T) {
		src := t.TempDir()
		sol := t.TempDir()
		os.WriteFile(filepath.Join(src, "main.go"), []byte("package main\n"), 0644)
		os.WriteFile(filepath.Join(sol, "main.go"), []byte("package main\n"), 0644)

		diff, err := GenerateDiff(src, sol)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if diff != "" {
			t.Errorf("expected empty diff, got %q", diff)
		}
	})

	t.Run("ModifiedFile", func(t *testing.T) {
		src := t.TempDir()
		sol := t.TempDir()
		os.WriteFile(filepath.Join(src, "main.go"), []byte("package main\n"), 0644)
		os.WriteFile(filepath.Join(sol, "main.go"), []byte("package main\n\nfunc hello() {}\n"), 0644)

		diff, err := GenerateDiff(src, sol)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(diff, "---") || !strings.Contains(diff, "+++") {
			t.Errorf("expected unified diff markers, got:\n%s", diff)
		}
	})

	t.Run("NewFile", func(t *testing.T) {
		src := t.TempDir()
		sol := t.TempDir()
		os.WriteFile(filepath.Join(sol, "new.go"), []byte("package new\n"), 0644)

		diff, err := GenerateDiff(src, sol)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(diff, "=== NEW FILE") {
			t.Errorf("expected NEW FILE marker, got:\n%s", diff)
		}
	})

	t.Run("DeletedFile", func(t *testing.T) {
		src := t.TempDir()
		sol := t.TempDir()
		os.WriteFile(filepath.Join(src, "old.go"), []byte("package old\n"), 0644)

		diff, err := GenerateDiff(src, sol)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(diff, "=== DELETED FILE") {
			t.Errorf("expected DELETED FILE marker, got:\n%s", diff)
		}
	})
}
