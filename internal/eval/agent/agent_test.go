package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

// mockTool implements tools.Tool for testing csvFromTools
type mockTool struct {
	name string
	desc string
}

func (m mockTool) Name() string                                           { return m.name }
func (m mockTool) Description() string                                    { return m.desc }
func (m mockTool) Call(ctx context.Context, input string) (string, error) { return "", nil }

func TestCsvFromTools(t *testing.T) {
	toolList := []tools.Tool{
		mockTool{name: "tool_a", desc: "does A"},
		mockTool{name: "tool_b", desc: "does B"},
	}

	t.Run("Mode1_NamesOnly", func(t *testing.T) {
		got := csvFromTools(toolList, 1)
		if got != "tool_a; tool_b" {
			t.Errorf("csvFromTools(mode=1) = %q, want %q", got, "tool_a; tool_b")
		}
	})

	t.Run("Mode2_NamesAndDescriptions", func(t *testing.T) {
		got := csvFromTools(toolList, 2)
		if !strings.Contains(got, "tool_a: does A") || !strings.Contains(got, "tool_b: does B") {
			t.Errorf("csvFromTools(mode=2) = %q, expected name: desc format", got)
		}
	})

	t.Run("Mode0_Empty", func(t *testing.T) {
		got := csvFromTools(toolList, 0)
		if got != "" {
			t.Errorf("csvFromTools(mode=0) = %q, want empty", got)
		}
	})
}

func TestFileReader_Name(t *testing.T) {
	fr := FileReader{}
	if fr.Name() != "file_reader" {
		t.Errorf("Name() = %q, want %q", fr.Name(), "file_reader")
	}
}

func TestFileReader_Call(t *testing.T) {
	ctx := context.Background()

	t.Run("ExistingFile", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		os.WriteFile(path, []byte("hello world"), 0644)

		fr := FileReader{}
		got, err := fr.Call(ctx, path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hello world" {
			t.Errorf("Call() = %q, want %q", got, "hello world")
		}
	})

	t.Run("EmptyPath", func(t *testing.T) {
		fr := FileReader{}
		got, err := fr.Call(ctx, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "cannot be empty") {
			t.Errorf("Call('') = %q, expected 'cannot be empty'", got)
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		fr := FileReader{}
		got, err := fr.Call(ctx, "/nonexistent/path/file.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "Error: failed to read file") {
			t.Errorf("Call(nonexistent) = %q, expected error message", got)
		}
	})
}

func TestDirectoryLister_Name(t *testing.T) {
	dl := DirectoryLister{}
	if dl.Name() != "directory_lister" {
		t.Errorf("Name() = %q, want %q", dl.Name(), "directory_lister")
	}
}

func TestDirectoryLister_Call(t *testing.T) {
	ctx := context.Background()

	t.Run("DirWithFiles", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
		os.Mkdir(filepath.Join(dir, "subdir"), 0755)

		dl := DirectoryLister{}
		got, err := dl.Call(ctx, dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "[FILE]") {
			t.Errorf("expected [FILE] in output, got:\n%s", got)
		}
		if !strings.Contains(got, "a.txt") {
			t.Errorf("expected a.txt in output, got:\n%s", got)
		}
		if !strings.Contains(got, "[DIR]") {
			t.Errorf("expected [DIR] in output, got:\n%s", got)
		}
		if !strings.Contains(got, "2 items") {
			t.Errorf("expected '2 items' in output, got:\n%s", got)
		}
	})

	t.Run("EmptyDir", func(t *testing.T) {
		dir := t.TempDir()
		dl := DirectoryLister{}
		got, err := dl.Call(ctx, dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "is empty") {
			t.Errorf("expected 'is empty' in output, got: %q", got)
		}
	})

	t.Run("NonExistentDir", func(t *testing.T) {
		dl := DirectoryLister{}
		got, err := dl.Call(ctx, "/nonexistent/dir")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "does not exist") {
			t.Errorf("expected 'does not exist' in output, got: %q", got)
		}
	})

	t.Run("FileNotDir", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "file.txt")
		os.WriteFile(filePath, []byte("x"), 0644)

		dl := DirectoryLister{}
		got, err := dl.Call(ctx, filePath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "is not a directory") {
			t.Errorf("expected 'is not a directory' in output, got: %q", got)
		}
	})
}

func TestNormalizeAIEvaluationError(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		wantText string
		wantSame bool
	}{
		{
			name:     "StandardizedRateLimit",
			input:    llms.NewError(llms.ErrCodeRateLimit, "openai", "Rate limit exceeded"),
			wantText: aiRateLimitErrorMessage,
		},
		{
			name:     "StandardizedQuotaExceeded",
			input:    llms.NewError(llms.ErrCodeQuotaExceeded, "anthropic", "Quota exceeded"),
			wantText: aiQuotaExceededErrorMessage,
		},
		{
			name:     "StringFallbackRateLimit",
			input:    errors.New("429 too many requests"),
			wantText: aiRateLimitErrorMessage,
		},
		{
			name:     "StringFallbackQuotaExceeded",
			input:    errors.New("credit balance exhausted"),
			wantText: aiQuotaExceededErrorMessage,
		},
		{
			name:     "UnchangedOtherError",
			input:    errors.New("network down"),
			wantText: "network down",
			wantSame: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeAIEvaluationError(tt.input)
			if got == nil {
				t.Fatal("expected error, got nil")
			}
			if got.Error() != tt.wantText {
				t.Fatalf("normalizeAIEvaluationError() = %q, want %q", got.Error(), tt.wantText)
			}
			if tt.wantSame && got != tt.input {
				t.Fatal("expected error to be returned unchanged")
			}
			if strings.Contains(got.Error(), "\n") {
				t.Fatalf("error %q should not contain newlines", got.Error())
			}
		})
	}
}
