package eval

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParsePytestJSON(t *testing.T) {
	h := PythonHandler{}

	tests := []struct {
		name         string
		input        string
		wantTotal    int
		wantPassed   int
		wantFailed   int
		wantSkipped  int
		wantErrCount int
		wantErr      bool
	}{
		{
			name:       "AllPass",
			input:      `{"summary":{"total":3,"passed":3,"failed":0,"skipped":0},"tests":[]}`,
			wantTotal:  3,
			wantPassed: 3,
		},
		{
			name: "WithFailures",
			input: `{"summary":{"total":2,"passed":1,"failed":1,"skipped":0},"tests":[
				{"nodeid":"test_math.py::test_add","outcome":"passed","call":{"longrepr":""}},
				{"nodeid":"test_math.py::test_sub","outcome":"failed","call":{"longrepr":"AssertionError: expected 5"}}
			]}`,
			wantTotal: 2, wantPassed: 1, wantFailed: 1, wantErrCount: 1,
		},
		{
			name:    "InvalidJSON",
			input:   `{not valid json`,
			wantErr: true,
		},
		{
			name:        "WithSkipped",
			input:       `{"summary":{"total":3,"passed":1,"failed":0,"skipped":2},"tests":[]}`,
			wantTotal:   3,
			wantPassed:  1,
			wantSkipped: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &TestResult{}
			err := h.parsePytestJSON([]byte(tt.input), result)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.TotalTests != tt.wantTotal {
				t.Errorf("TotalTests = %d, want %d", result.TotalTests, tt.wantTotal)
			}
			if result.PassedTests != tt.wantPassed {
				t.Errorf("PassedTests = %d, want %d", result.PassedTests, tt.wantPassed)
			}
			if result.FailedTests != tt.wantFailed {
				t.Errorf("FailedTests = %d, want %d", result.FailedTests, tt.wantFailed)
			}
			if result.SkippedTests != tt.wantSkipped {
				t.Errorf("SkippedTests = %d, want %d", result.SkippedTests, tt.wantSkipped)
			}
			if tt.wantErrCount > 0 {
				if len(result.Errors) != tt.wantErrCount {
					t.Errorf("len(Errors) = %d, want %d", len(result.Errors), tt.wantErrCount)
				}
				if len(result.Errors) > 0 && !strings.Contains(result.Errors[0].TestName, "test_") {
					t.Errorf("Error TestName = %q, expected to contain 'test_'", result.Errors[0].TestName)
				}
			}
		})
	}
}

func TestPythonHandler_PrepareEnvironment(t *testing.T) {
	runID := "test-run-py"

	challengeDir := t.TempDir()
	solutionDir := filepath.Join(challengeDir, "solutions", runID)
	os.MkdirAll(solutionDir, 0755)
	os.WriteFile(filepath.Join(solutionDir, "main.py"), []byte("print('hello')\n"), 0644)

	testDir := filepath.Join(challengeDir, "test")
	os.MkdirAll(testDir, 0755)
	os.WriteFile(filepath.Join(testDir, "test_main.py"), []byte("def test_ok(): pass\n"), 0644)

	os.WriteFile(filepath.Join(challengeDir, "requirements.txt"), []byte("requests==2.31.0\n"), 0644)

	h := PythonHandler{}
	tempDir, cleanup, err := h.PrepareEnvironment(challengeDir, runID)
	if err != nil {
		t.Fatalf("PrepareEnvironment failed: %v", err)
	}
	defer cleanup()

	// test files should be in tempDir/test
	if _, err := os.Stat(filepath.Join(tempDir, "test", "test_main.py")); os.IsNotExist(err) {
		t.Error("expected test_main.py in tempDir/test")
	}

	// solution files should be in tempDir/src
	if _, err := os.Stat(filepath.Join(tempDir, "src", "main.py")); os.IsNotExist(err) {
		t.Error("expected main.py in tempDir/src")
	}

	// requirements.txt should be copied
	if _, err := os.Stat(filepath.Join(tempDir, "requirements.txt")); os.IsNotExist(err) {
		t.Error("expected requirements.txt in tempDir")
	}
}

func TestPythonHandler_PrepareEnvironment_NoRequirements(t *testing.T) {
	runID := "test-run-py-noreq"

	challengeDir := t.TempDir()
	solutionDir := filepath.Join(challengeDir, "solutions", runID)
	os.MkdirAll(solutionDir, 0755)
	os.WriteFile(filepath.Join(solutionDir, "main.py"), []byte("print('hello')\n"), 0644)
	os.MkdirAll(filepath.Join(challengeDir, "test"), 0755)

	h := PythonHandler{}
	tempDir, cleanup, err := h.PrepareEnvironment(challengeDir, runID)
	if err != nil {
		t.Fatalf("PrepareEnvironment failed: %v", err)
	}
	defer cleanup()

	// requirements.txt should NOT exist
	if _, err := os.Stat(filepath.Join(tempDir, "requirements.txt")); !os.IsNotExist(err) {
		t.Error("expected no requirements.txt in tempDir when source has none")
	}
}

func TestPythonHandler_PrepareEnvironment_Cleanup(t *testing.T) {
	runID := "test-cleanup-py"

	challengeDir := t.TempDir()
	solutionDir := filepath.Join(challengeDir, "solutions", runID)
	os.MkdirAll(solutionDir, 0755)
	os.WriteFile(filepath.Join(solutionDir, "main.py"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(challengeDir, "test"), 0755)

	h := PythonHandler{}
	tempDir, cleanup, err := h.PrepareEnvironment(challengeDir, runID)
	if err != nil {
		t.Fatalf("PrepareEnvironment failed: %v", err)
	}

	cleanup()

	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Errorf("expected tempDir %q to be removed after cleanup", tempDir)
	}
}
