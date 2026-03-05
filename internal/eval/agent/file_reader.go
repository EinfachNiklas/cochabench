package agent

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/tools"
)

// FileReader is a tool that reads the contents of a file
type FileReader struct {
	RunID string // If set, restricts access to solutions/{RunID} directory
}

var _ tools.Tool = FileReader{}

func (f FileReader) Name() string {
	return "file_reader"
}

func (f FileReader) Description() string {
	return `Reads the contents of a file and returns it as a string.
Input: The absolute or relative file path to read.
Output: The file contents as text.
Example input: "./example.txt" or "/path/to/file.go"`
}

func (f FileReader) Call(ctx context.Context, input string) (string, error) {
	if input == "" {
		return "Error: file path cannot be empty", nil
	}

	content, err := os.ReadFile(input)
	if err != nil {
		return fmt.Sprintf("Error: failed to read file '%s': %v", input, err), nil
	}

	return string(content), nil
}
