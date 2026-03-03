package eval

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type GoHandler struct{}

func (h GoHandler) ExecuteTests(ctx context.Context, tempDir string) (*TestResult, error) {
	startTime := time.Now()

	// Download dependencies
	fmt.Println("Downloading Go dependencies...")
	downloadCmd := exec.CommandContext(ctx, "go", "mod", "download")
	downloadCmd.Dir = tempDir
	if output, err := downloadCmd.CombinedOutput(); err != nil {
		fmt.Printf("Warning: go mod download failed: %s\n", output)
	}

	fmt.Println("Running go mod tidy...")
	tidyCmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	tidyCmd.Dir = tempDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("go mod tidy timed out")
		}
		return nil, fmt.Errorf("failed to run go mod tidy: %s", output)
	}

	fmt.Println("Executing Go Tests...")

	cmd := exec.CommandContext(ctx, "go", "test", "-json", "./...")
	cmd.Dir = tempDir

	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("Test execution timed out")
	}

	result := &TestResult{
		Duration: duration,
	}

	if parseErr := h.parseGoTestJSON(output, result); parseErr != nil {
		return nil, fmt.Errorf("failed to parse test output: %w\n\n%s", parseErr, output)
	}

	if err != nil {
		result.Passed = false
	} else {
		result.Passed = result.FailedTests == 0
	}

	return result, nil
}

func (h GoHandler) parseGoTestJSON(data []byte, result *TestResult) error {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	testResults := make(map[string]string)
	testOutputs := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		var event struct {
			Action  string  `json:"Action"`
			Test    string  `json:"Test"`
			Output  string  `json:"Output"`
			Elapsed float64 `json:"Elapsed"`
		}

		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		switch event.Action {
		case "pass":
			testResults[event.Test] = "pass"
		case "fail":
			testResults[event.Test] = "fail"
			if event.Output != "" {
				testOutputs[event.Test] += event.Output
			}
		case "skip":
			testResults[event.Test] = "skip"
		case "output":

			if event.Test != "" {
				testOutputs[event.Test] += event.Output
			}
		}
	}

	for testName, outcome := range testResults {
		switch outcome {
		case "pass":
			result.PassedTests++
		case "fail":
			result.FailedTests++

			message := strings.TrimSpace(testOutputs[testName])
			result.Errors = append(result.Errors, TestError{
				TestName: testName,
				Message:  message,
				Stack:    "",
			})
		case "skip":
			result.SkippedTests++
		}
	}

	result.TotalTests = result.PassedTests + result.FailedTests + result.SkippedTests

	return nil
}

func (h GoHandler) PrepareEnvironment(challengePath string, runID string) (tempDir string, cleanup func(), err error) {
	tempDir, err = os.MkdirTemp("", fmt.Sprintf("cochabench-%s-*", runID))
	if err != nil {
		return "", nil, fmt.Errorf("Failed to create temp directory: %w", err)
	}

	cleanup = func() {
		os.RemoveAll(tempDir)
	}

	// Copy Files
	solutionSrc := filepath.Join(challengePath, "solutions", runID)
	srcDst := filepath.Join(tempDir, "src")
	if err := os.CopyFS(srcDst, os.DirFS(solutionSrc)); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("Failed to copy solution files: %w", err)
	}
	goModData, err := os.ReadFile(filepath.Join(srcDst, "go.mod"))
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("Failed to copy go.mod from src/: %v", err)
	}
	err = os.WriteFile(filepath.Join(srcDst, "..", "go.mod"), goModData, 0644)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("Failed to copy go.mod: %v", err)
	}

	if removeErr := os.Remove(filepath.Join(srcDst, "go.mod")); removeErr != nil {
		cleanup()
		return "", nil, fmt.Errorf("Failed to remove src/go.mod: %w", removeErr)
	}

	testSrc := filepath.Join(challengePath, "test")
	testFiles, err := os.ReadDir(testSrc)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("Failed to read test directory: %w", err)
	}

	for _, file := range testFiles {
		if !file.IsDir() {
			srcPath := filepath.Join(testSrc, file.Name())
			var dstPath string
			dstPath = filepath.Join(srcDst, file.Name())
			data, readErr := os.ReadFile(srcPath)
			if readErr != nil {
				cleanup()
				return "", nil, fmt.Errorf("Failed to read test file %s: %w", file.Name(), readErr)
			}
			if writeErr := os.WriteFile(dstPath, data, 0644); writeErr != nil {
				cleanup()
				return "", nil, fmt.Errorf("Failed to write test file %s: %w", file.Name(), writeErr)
			}
		}
	}

	goSumSrc := filepath.Join(srcDst, "go.sum")
	goSumDst := filepath.Join(tempDir, "go.sum")
	if _, statErr := os.Stat(goSumSrc); statErr == nil {
		data, readErr := os.ReadFile(goSumSrc)
		if readErr != nil {
			cleanup()
			return "", nil, fmt.Errorf("Failed to read src/go.sum: %w", readErr)
		}
		if writeErr := os.WriteFile(goSumDst, data, 0644); writeErr != nil {
			cleanup()
			return "", nil, fmt.Errorf("Failed to write go.sum: %w", writeErr)
		}
		// Delete src/go.sum after copying
		if removeErr := os.Remove(goSumSrc); removeErr != nil {
			cleanup()
			return "", nil, fmt.Errorf("Failed to remove src/go.sum: %w", removeErr)
		}
	}

	return tempDir, cleanup, nil
}
