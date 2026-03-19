package eval

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseGoTestJSON(t *testing.T) {
	h := GoHandler{}

	tests := []struct {
		name          string
		input         string
		wantTotal     int
		wantPassed    int
		wantFailed    int
		wantSkipped   int
		wantErrCount  int
	}{
		{
			name: "AllPass",
			input: `{"Action":"pass","Test":"TestA"}
{"Action":"pass","Test":"TestB"}`,
			wantTotal: 2, wantPassed: 2,
		},
		{
			name: "WithFailure",
			input: `{"Action":"output","Test":"TestA","Output":"some output\n"}
{"Action":"pass","Test":"TestA"}
{"Action":"output","Test":"TestB","Output":"failure details\n"}
{"Action":"fail","Test":"TestB"}`,
			wantTotal: 2, wantPassed: 1, wantFailed: 1, wantErrCount: 1,
		},
		{
			name: "WithSkip",
			input: `{"Action":"skip","Test":"TestA"}
{"Action":"pass","Test":"TestB"}`,
			wantTotal: 2, wantPassed: 1, wantSkipped: 1,
		},
		{
			name:      "EmptyInput",
			input:     "",
			wantTotal: 0,
		},
		{
			name:      "InvalidJSON",
			input:     "not json at all\nstill not json",
			wantTotal: 0,
		},
		{
			name: "PackageLevelEventsIgnored",
			input: `{"Action":"pass","Test":"TestA"}
{"Action":"pass","Test":""}`,
			wantTotal: 1, wantPassed: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &TestResult{}
			err := h.parseGoTestJSON([]byte(tt.input), result)
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
			if tt.wantErrCount > 0 && len(result.Errors) != tt.wantErrCount {
				t.Errorf("len(Errors) = %d, want %d", len(result.Errors), tt.wantErrCount)
			}
			if tt.wantErrCount > 0 && len(result.Errors) > 0 {
				if !strings.Contains(result.Errors[0].TestName, "Test") {
					t.Errorf("Error TestName = %q, expected to contain 'Test'", result.Errors[0].TestName)
				}
			}
		})
	}
}

func TestGoHandler_PrepareEnvironment(t *testing.T) {
	runID := "test-run-001"

	// Build a fake challenge directory
	challengeDir := t.TempDir()
	solutionDir := filepath.Join(challengeDir, "solutions", runID)
	os.MkdirAll(solutionDir, 0755)
	os.WriteFile(filepath.Join(solutionDir, "go.mod"), []byte("module example\n\ngo 1.21\n"), 0644)
	os.WriteFile(filepath.Join(solutionDir, "main.go"), []byte("package main\n"), 0644)
	os.WriteFile(filepath.Join(solutionDir, "go.sum"), []byte("h1:abc\n"), 0644)

	testDir := filepath.Join(challengeDir, "test")
	os.MkdirAll(testDir, 0755)
	os.WriteFile(filepath.Join(testDir, "main_test.go"), []byte("package main\n"), 0644)

	h := GoHandler{}
	tempDir, cleanup, err := h.PrepareEnvironment(challengeDir, runID)
	if err != nil {
		t.Fatalf("PrepareEnvironment failed: %v", err)
	}
	defer cleanup()

	// go.mod should be moved to tempDir root
	if _, err := os.Stat(filepath.Join(tempDir, "go.mod")); os.IsNotExist(err) {
		t.Error("expected go.mod at tempDir root")
	}

	// main.go should be in tempDir/src
	if _, err := os.Stat(filepath.Join(tempDir, "src", "main.go")); os.IsNotExist(err) {
		t.Error("expected main.go in tempDir/src")
	}

	// go.mod should NOT remain in src/
	if _, err := os.Stat(filepath.Join(tempDir, "src", "go.mod")); !os.IsNotExist(err) {
		t.Error("expected go.mod to be removed from tempDir/src")
	}

	// test files should be merged into src/
	if _, err := os.Stat(filepath.Join(tempDir, "src", "main_test.go")); os.IsNotExist(err) {
		t.Error("expected main_test.go in tempDir/src (merged from test/)")
	}

	// go.sum should be moved to tempDir root
	if _, err := os.Stat(filepath.Join(tempDir, "go.sum")); os.IsNotExist(err) {
		t.Error("expected go.sum at tempDir root")
	}

	// go.sum should NOT remain in src/
	if _, err := os.Stat(filepath.Join(tempDir, "src", "go.sum")); !os.IsNotExist(err) {
		t.Error("expected go.sum to be removed from tempDir/src")
	}
}

func TestGoHandler_PrepareEnvironment_Cleanup(t *testing.T) {
	runID := "test-cleanup"

	challengeDir := t.TempDir()
	solutionDir := filepath.Join(challengeDir, "solutions", runID)
	os.MkdirAll(solutionDir, 0755)
	os.WriteFile(filepath.Join(solutionDir, "go.mod"), []byte("module example\n\ngo 1.21\n"), 0644)
	os.WriteFile(filepath.Join(solutionDir, "main.go"), []byte("package main\n"), 0644)
	os.MkdirAll(filepath.Join(challengeDir, "test"), 0755)

	h := GoHandler{}
	tempDir, cleanup, err := h.PrepareEnvironment(challengeDir, runID)
	if err != nil {
		t.Fatalf("PrepareEnvironment failed: %v", err)
	}

	cleanup()

	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Errorf("expected tempDir %q to be removed after cleanup", tempDir)
	}
}
