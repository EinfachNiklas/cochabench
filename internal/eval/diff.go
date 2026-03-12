package eval

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

const (
	MaxDiffChars     = 30000
	DiffContextLines = 3
)

// GenerateDiff produces a unified diff between the original template (srcDir)
// and the user's solution (solutionDir).
// Returns an empty string if there are no changes.
func GenerateDiff(srcDir, solutionDir string) (string, error) {
	srcFiles, err := collectFiles(srcDir)
	if err != nil {
		return "", fmt.Errorf("failed to scan source directory: %w", err)
	}
	solFiles, err := collectFiles(solutionDir)
	if err != nil {
		return "", fmt.Errorf("failed to scan solution directory: %w", err)
	}

	allPaths := make(map[string]bool)
	for p := range srcFiles {
		allPaths[p] = true
	}
	for p := range solFiles {
		allPaths[p] = true
	}

	sortedPaths := sortedKeys(allPaths)

	var result strings.Builder
	for _, relPath := range sortedPaths {
		srcContent := readFileOrEmpty(srcDir, relPath)
		solContent := readFileOrEmpty(solutionDir, relPath)

		if srcContent == solContent {
			continue
		}

		if isBinary(srcContent) || isBinary(solContent) {
			fmt.Fprintf(&result, "Binary file changed: %s\n", relPath)
			continue
		}

		_, inSrc := srcFiles[relPath]
		_, inSol := solFiles[relPath]

		if !inSrc {
			fmt.Fprintf(&result, "=== NEW FILE: %s ===\n", relPath)
		} else if !inSol {
			fmt.Fprintf(&result, "=== DELETED FILE: %s ===\n", relPath)
		}

		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(srcContent),
			B:        difflib.SplitLines(solContent),
			FromFile: "original/" + relPath,
			ToFile:   "solution/" + relPath,
			Context:  DiffContextLines,
		}
		text, err := difflib.GetUnifiedDiffString(diff)
		if err != nil {
			return "", fmt.Errorf("failed to generate diff for %s: %w", relPath, err)
		}

		result.WriteString(text)
	}

	diffStr := result.String()

	if len(diffStr) > MaxDiffChars {
		totalLen := len(diffStr)
		diffStr = diffStr[:MaxDiffChars] +
			fmt.Sprintf("\n\n[DIFF TRUNCATED - showing first %d of %d characters. Use file_reader tool to inspect full files.]\n",
				MaxDiffChars, totalLen)
	}

	fmt.Println(diffStr)

	return diffStr, nil
}

func collectFiles(root string) (map[string]bool, error) {
	files := make(map[string]bool)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = true
		return nil
	})
	return files, err
}

func readFileOrEmpty(root, relPath string) string {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relPath)))
	if err != nil {
		return ""
	}
	return string(data)
}

func isBinary(content string) bool {
	checkLen := len(content)
	if checkLen > 8192 {
		checkLen = 8192
	}
	return strings.ContainsRune(content[:checkLen], '\x00')
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
