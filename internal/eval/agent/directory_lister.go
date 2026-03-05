package agent

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/langchaingo/tools"
)

// DirectoryLister is a tool that lists the contents of a directory
type DirectoryLister struct {
	RunID string // If set, restricts access to solutions/{RunID} directory
}

var _ tools.Tool = DirectoryLister{}

func (f DirectoryLister) Name() string {
	return "directory_lister"
}

func (f DirectoryLister) Description() string {
	return `Lists the contents of a directory.
Input: The absolute or relative directory path to list. Leave empty or use "." for current directory.
Output: A formatted list of files and directories with their types and sizes.
Example input: "./src" or "/path/to/directory" or "." or ""
The listing includes file sizes and indicates whether entries are files or directories.
Only paths inside the working directory are allowed`
}

func (f DirectoryLister) Call(ctx context.Context, input string) (string, error) {
	dirPath := strings.TrimSpace(input)
	if dirPath == "" {
		dirPath = "."
	}

	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("Error: directory '%s' does not exist", dirPath), nil
		}
		return fmt.Sprintf("Error: failed to access directory '%s': %v", dirPath, err), nil
	}

	// Ensure it's a directory
	if !info.IsDir() {
		return fmt.Sprintf("Error: '%s' is not a directory", dirPath), nil
	}

	// Read directory contents
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Sprintf("Error: failed to read directory '%s': %v", dirPath, err), nil
	}

	if len(entries) == 0 {
		return fmt.Sprintf("Directory '%s' is empty", dirPath), nil
	}

	// Build the output
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Contents of '%s' (%d items):\n\n", dirPath, len(entries)))

	for _, entry := range entries {
		entryInfo, err := entry.Info()
		if err != nil {
			result.WriteString(fmt.Sprintf("  [ERROR] %s - could not read info\n", entry.Name()))
			continue
		}

		entryType := "[FILE]"
		size := fmt.Sprintf("%d bytes", entryInfo.Size())
		if entry.IsDir() {
			entryType = "[DIR] "
			size = "-"
		}

		result.WriteString(fmt.Sprintf("  %s %-40s %s\n", entryType, entry.Name(), size))
	}

	return result.String(), nil
}
