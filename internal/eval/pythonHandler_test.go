package eval

import (
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
