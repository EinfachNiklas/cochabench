package eval

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseJestJSON(t *testing.T) {
	h := JavascriptHandler{}

	tests := []struct {
		name         string
		input        string
		wantTotal    int
		wantPassed   int
		wantFailed   int
		wantSkipped  int
		wantErrCount int
		wantErr      bool
		errSubstr    string
	}{
		{
			name:      "AllPass",
			input:     `{"numTotalTests":3,"numPassedTests":3,"numFailedTests":0,"numPendingTests":0,"testResults":[]}`,
			wantTotal: 3, wantPassed: 3,
		},
		{
			name:      "WithFailures",
			input:     `{"numTotalTests":2,"numPassedTests":1,"numFailedTests":1,"numPendingTests":0,"testResults":[{"assertionResults":[{"title":"should fail","status":"failed","failureMessages":["Expected true to be false"]}]}]}`,
			wantTotal: 2, wantPassed: 1, wantFailed: 1, wantErrCount: 1,
		},
		{
			name:      "NoMarker",
			input:     `{"some":"other json"}`,
			wantErr:   true,
			errSubstr: "Jest did not produce valid JSON test output",
		},
		{
			name:      "WithPending",
			input:     `{"numTotalTests":3,"numPassedTests":2,"numFailedTests":0,"numPendingTests":1,"testResults":[]}`,
			wantTotal: 3, wantPassed: 2, wantSkipped: 1,
		},
		{
			name: "WithLeadingAndTrailingOutput",
			input: "npm notice something\n" +
				`{"numTotalTests":1,"numPassedTests":1,"numFailedTests":0,"numPendingTests":0,"testResults":[]}` +
				"\nJest did finish successfully\n",
			wantTotal: 1, wantPassed: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &TestResult{}
			err := h.parseJestJSON([]byte(tt.input), result)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q missing substring %q", err, tt.errSubstr)
				}
				if strings.Contains(err.Error(), "\n") {
					t.Errorf("error %q should not contain newlines", err)
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
			if tt.wantErrCount > 0 && len(result.Errors) != tt.wantErrCount {
				t.Errorf("len(Errors) = %d, want %d", len(result.Errors), tt.wantErrCount)
			}
		})
	}
}

func TestJavascriptHandler_PrepareEnvironment(t *testing.T) {
	runID := "test-run-js"

	challengeDir := t.TempDir()
	solutionDir := filepath.Join(challengeDir, "solutions", runID)
	os.MkdirAll(filepath.Join(solutionDir, "src"), 0755)
	os.WriteFile(filepath.Join(solutionDir, "src", "index.js"), []byte("module.exports = {};\n"), 0644)

	testDir := filepath.Join(challengeDir, "test")
	os.MkdirAll(testDir, 0755)
	os.WriteFile(filepath.Join(testDir, "index.test.js"), []byte("test('ok', () => {});\n"), 0644)

	os.WriteFile(filepath.Join(challengeDir, "package.json"), []byte(`{"name":"test"}`), 0644)

	h := JavascriptHandler{}
	tempDir, cleanup, err := h.PrepareEnvironment(challengeDir, runID)
	if err != nil {
		t.Fatalf("PrepareEnvironment failed: %v", err)
	}
	defer cleanup()

	// test files should be in tempDir/test
	if _, err := os.Stat(filepath.Join(tempDir, "test", "index.test.js")); os.IsNotExist(err) {
		t.Error("expected index.test.js in tempDir/test")
	}

	// solution files should be in tempDir
	if _, err := os.Stat(filepath.Join(tempDir, "src", "index.js")); os.IsNotExist(err) {
		t.Error("expected src/index.js in tempDir")
	}

	// package.json should be copied
	if _, err := os.Stat(filepath.Join(tempDir, "package.json")); os.IsNotExist(err) {
		t.Error("expected package.json in tempDir")
	}
}

func TestJavascriptHandler_PrepareEnvironment_Cleanup(t *testing.T) {
	runID := "test-cleanup-js"

	challengeDir := t.TempDir()
	solutionDir := filepath.Join(challengeDir, "solutions", runID)
	os.MkdirAll(solutionDir, 0755)
	os.WriteFile(filepath.Join(solutionDir, "index.js"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(challengeDir, "test"), 0755)

	h := JavascriptHandler{}
	tempDir, cleanup, err := h.PrepareEnvironment(challengeDir, runID)
	if err != nil {
		t.Fatalf("PrepareEnvironment failed: %v", err)
	}

	cleanup()

	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Errorf("expected tempDir %q to be removed after cleanup", tempDir)
	}
}
