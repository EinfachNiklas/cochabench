package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type JavascriptHandler struct{}

func (h JavascriptHandler) ExecuteTests(ctx context.Context, tempDir string) (*TestResult, error) {

	fmt.Println("Installing Packages...")
	cmd := exec.CommandContext(ctx, "npm", "install")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("Dependency installation timed out")
		}
		return nil, fmt.Errorf("Could not install Node.js dependencies: %w", err)
	}

	fmt.Println("Executing Tests...")
	cmd = exec.CommandContext(ctx, "npm", "test", "--", "--json", "--silent", "--noStackTrace")
	cmd.Dir = tempDir

	output, err = cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("Test execution timed out")
	}

	result := &TestResult{}

	// Try parsing as Jest JSON
	if parseErr := h.parseJestJSON(output, result); parseErr != nil {
		return nil, fmt.Errorf("Could not parse Jest test results: %w", parseErr)
	}

	// npm test returns non-zero exit code if tests fail
	if err != nil {
		result.Passed = false
	} else {
		result.Passed = result.FailedTests == 0
	}

	return result, nil
}

// parseJestJSON parses Jest JSON output
func (h JavascriptHandler) parseJestJSON(data []byte, result *TestResult) error {
	marker := []byte(`"numTotalTests"`)
	markerIndex := bytes.Index(data, marker)

	if markerIndex == -1 {
		return fmt.Errorf("Jest did not produce valid JSON test output")
	}

	// Search backwards from marker to find the opening '{'
	jsonStart := -1
	for i := markerIndex; i >= 0; i-- {
		if data[i] == '{' {
			jsonStart = i
			break
		}
	}

	if jsonStart == -1 {
		return fmt.Errorf("Jest output did not contain a valid JSON report")
	}

	jsonEnd := findJSONEnd(data, jsonStart)
	if jsonEnd == -1 {
		return fmt.Errorf("Jest output contained incomplete JSON")
	}

	jsonData := data[jsonStart : jsonEnd+1]

	var jestReport struct {
		NumTotalTests   int `json:"numTotalTests"`
		NumPassedTests  int `json:"numPassedTests"`
		NumFailedTests  int `json:"numFailedTests"`
		NumPendingTests int `json:"numPendingTests"`
		TestResults     []struct {
			AssertionResults []struct {
				Title           string   `json:"title"`
				Status          string   `json:"status"`
				FailureMessages []string `json:"failureMessages"`
			} `json:"assertionResults"`
		} `json:"testResults"`
	}

	if err := json.Unmarshal(jsonData, &jestReport); err != nil {
		return err
	}

	result.TotalTests = jestReport.NumTotalTests
	result.PassedTests = jestReport.NumPassedTests
	result.FailedTests = jestReport.NumFailedTests
	result.SkippedTests = jestReport.NumPendingTests

	for _, testFile := range jestReport.TestResults {
		for _, assertion := range testFile.AssertionResults {
			if assertion.Status == "failed" {
				message := ""
				if len(assertion.FailureMessages) > 0 {
					message = assertion.FailureMessages[0]
				}
				result.Errors = append(result.Errors, TestError{
					TestName: assertion.Title,
					Message:  message,
					Stack:    "",
				})
			}
		}
	}

	return nil
}

func findJSONEnd(data []byte, start int) int {
	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(data); i++ {
		ch := data[i]

		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

func (h JavascriptHandler) PrepareEnvironment(challengePath string, runID string) (tempDir string, cleanup func(), err error) {
	tempDir, err = os.MkdirTemp("", fmt.Sprintf("cochabench-%s-*", runID))
	if err != nil {
		return "", nil, fmt.Errorf("Could not create temporary evaluation directory: %w", err)
	}

	cleanup = func() {
		os.RemoveAll(tempDir)
	}

	//Copy challenge files
	err = os.CopyFS(filepath.Join(tempDir, "test"), os.DirFS(filepath.Join(challengePath, "test")))
	if err != nil {
		return "", cleanup, fmt.Errorf("Could not copy test files: %w", err)
	}

	err = os.CopyFS(tempDir, os.DirFS(filepath.Join(challengePath, "solutions", runID)))
	if err != nil {
		return "", cleanup, fmt.Errorf("Could not copy solution files: %w", err)
	}

	packageJSONSrc := filepath.Join(challengePath, "package.json")
	_, statErr := os.Stat(packageJSONSrc)
	if statErr == nil {
		data, readErr := os.ReadFile(packageJSONSrc)
		if readErr == nil {
			err = os.WriteFile(filepath.Join(tempDir, "package.json"), data, 0644)
			if err != nil {
				return "", cleanup, fmt.Errorf("Could not copy package.json: %w", err)
			}
		}
	}

	return tempDir, cleanup, nil

}
