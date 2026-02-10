package eval

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type JavascriptHandler struct{}

func (h JavascriptHandler) ExecuteTests(tempDir string) (*TestResult, error) {
	startTime := time.Now()
	fmt.Println("Installing Packages...")
	cmd := exec.Command("npm", "install")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Error when installing Node Modules: %s", output)
	}
	fmt.Println("Executing Tests...")
	cmd = exec.Command("npm", "test", "--", "--json")
	cmd.Dir = tempDir

	output, err = cmd.CombinedOutput()
	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf(string(output), err)
	}

	result := &TestResult{
		Duration: duration,
		Output:   string(output),
	}
	return result, nil
}

func (h JavascriptHandler) PrepareEnvironment(challengePath string, runID string) (tempDir string, cleanup func(), err error) {
	tempDir, err = os.MkdirTemp("./", fmt.Sprintf("cochabench-%s-*", runID))
	if err != nil {
		return "", nil, fmt.Errorf("Failed to create temp directory: %w", err)
	}

	cleanup = func() {
		os.RemoveAll(tempDir)
	}

	//Copy challenge files
	err = os.CopyFS(filepath.Join(tempDir, "test"), os.DirFS(filepath.Join(challengePath, "test")))
	if err != nil {
		return "", cleanup, fmt.Errorf("Failed to copy test files to temp dir: %w", err)
	}

	err = os.CopyFS(filepath.Join(tempDir, "src"), os.DirFS(filepath.Join(challengePath, "solutions", runID)))
	if err != nil {
		return "", cleanup, fmt.Errorf("Failed to copy solution files to temp dir: %w", err)
	}

	err = os.CopyFS(filepath.Join(tempDir, ""), os.DirFS(filepath.Join(challengePath, "solutions", runID)))
	packageJSONSrc := filepath.Join(challengePath, "package.json")
	_, statErr := os.Stat(packageJSONSrc)
	if statErr == nil {
		data, readErr := os.ReadFile(packageJSONSrc)
		if readErr == nil {
			err = os.WriteFile(filepath.Join(tempDir, "package.json"), data, 0644)
			if err != nil {
				return "", cleanup, fmt.Errorf("Failed to copy package.json to temp dir: %w", err)
			}
		}
	}

	return tempDir, cleanup, nil

}
