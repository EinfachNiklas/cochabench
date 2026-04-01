package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type PythonHandler struct{}

func (h PythonHandler) ExecuteTests(ctx context.Context, tempDir string) (*TestResult, error) {
	startTime := time.Now()

	// Determine python command (try python3 first, fallback to python)
	pythonCmd := "python3"
	if _, err := exec.LookPath(pythonCmd); err != nil {
		pythonCmd = "python"
	}

	// Create virtual environment
	fmt.Println("Creating virtual environment...")
	venvPath := filepath.Join(tempDir, "venv")
	cmd := exec.CommandContext(ctx, pythonCmd, "-m", "venv", venvPath)
	cmd.Dir = tempDir
	_, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("Python environment setup timed out")
		}
		return nil, fmt.Errorf("Could not create Python virtual environment: %w", err)
	}

	var pipPath, pythonPath string
	if _, err := os.Stat(filepath.Join(venvPath, "Scripts")); err == nil {
		// Windows
		pipPath = filepath.Join(venvPath, "Scripts", "pip.exe")
		pythonPath = filepath.Join(venvPath, "Scripts", "python.exe")
	} else {
		// Unix-like systems
		pipPath = filepath.Join(venvPath, "bin", "pip")
		pythonPath = filepath.Join(venvPath, "bin", "python")
	}

	fmt.Println("Installing pytest...")
	cmd = exec.CommandContext(ctx, pipPath, "install", "pytest", "pytest-json-report", "--quiet")
	cmd.Dir = tempDir
	_, err = cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("Pytest installation timed out")
		}
		return nil, fmt.Errorf("Could not install pytest: %w", err)
	}

	requirementsPath := filepath.Join(tempDir, "requirements.txt")
	if _, err := os.Stat(requirementsPath); err == nil {
		fmt.Println("Installing Python packages...")
		cmd = exec.CommandContext(ctx, pipPath, "install", "-r", "requirements.txt", "--quiet")
		cmd.Dir = tempDir
		_, err = cmd.CombinedOutput()
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("Python dependency installation timed out")
			}
			return nil, fmt.Errorf("Could not install Python dependencies: %w", err)
		}
	}

	fmt.Println("Executing Tests...")
	cmd = exec.CommandContext(ctx, pythonPath, "-m", "pytest", "test", "--json-report", "--json-report-file=report.json", "-v")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "PYTHONPATH="+tempDir)

	_, err = cmd.CombinedOutput()
	duration := time.Since(startTime)

	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("Test execution timed out")
	}

	result := &TestResult{
		Duration: duration,
	}

	reportPath := filepath.Join(tempDir, "report.json")
	reportData, readErr := os.ReadFile(reportPath)
	if readErr != nil {
		return nil, fmt.Errorf("Could not read pytest report: %w", readErr)
	}

	if parseErr := h.parsePytestJSON(reportData, result); parseErr != nil {
		return nil, fmt.Errorf("Could not parse pytest results: %w", parseErr)
	}

	if err != nil {
		result.Passed = false
	} else {
		result.Passed = result.FailedTests == 0
	}

	return result, nil
}

func (h PythonHandler) parsePytestJSON(data []byte, result *TestResult) error {
	var pytestReport struct {
		Summary struct {
			Total   int `json:"total"`
			Passed  int `json:"passed"`
			Failed  int `json:"failed"`
			Skipped int `json:"skipped"`
		} `json:"summary"`
		Tests []struct {
			NodeID  string `json:"nodeid"`
			Outcome string `json:"outcome"`
			Call    struct {
				Longrepr string `json:"longrepr"`
			} `json:"call"`
		} `json:"tests"`
	}

	if err := json.Unmarshal(data, &pytestReport); err != nil {
		return err
	}

	result.TotalTests = pytestReport.Summary.Total
	result.PassedTests = pytestReport.Summary.Passed
	result.FailedTests = pytestReport.Summary.Failed
	result.SkippedTests = pytestReport.Summary.Skipped

	for _, test := range pytestReport.Tests {
		if test.Outcome == "failed" {
			result.Errors = append(result.Errors, TestError{
				TestName: test.NodeID,
				Message:  test.Call.Longrepr,
				Stack:    "",
			})
		}
	}

	return nil
}

func (h PythonHandler) PrepareEnvironment(challengePath string, runID string) (tempDir string, cleanup func(), err error) {
	tempDir, err = os.MkdirTemp("", fmt.Sprintf("cochabench-%s-*", runID))
	if err != nil {
		return "", nil, fmt.Errorf("Could not create temporary evaluation directory: %w", err)
	}

	cleanup = func() {
		os.RemoveAll(tempDir)
	}

	testSrc := filepath.Join(challengePath, "test")
	testDst := filepath.Join(tempDir, "test")
	if err := os.CopyFS(testDst, os.DirFS(testSrc)); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("Could not copy test files: %w", err)
	}

	solutionSrc := filepath.Join(challengePath, "solutions", runID)
	solutionDst := filepath.Join(tempDir, "src")
	if err := os.CopyFS(solutionDst, os.DirFS(solutionSrc)); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("Could not copy solution files: %w", err)
	}

	requirementsSrc := filepath.Join(challengePath, "requirements.txt")
	requirementsDst := filepath.Join(tempDir, "requirements.txt")
	if _, statErr := os.Stat(requirementsSrc); statErr == nil {
		data, readErr := os.ReadFile(requirementsSrc)
		if readErr == nil {
			err = os.WriteFile(requirementsDst, data, 0644)
			if err != nil {
				return "", cleanup, fmt.Errorf("Could not copy requirements.txt: %w", err)
			}
		}
	}

	return tempDir, cleanup, nil
}
