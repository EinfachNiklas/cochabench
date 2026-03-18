package eval

import (
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
			name: "AllPass",
			input: `{"numTotalTests":3,"numPassedTests":3,"numFailedTests":0,"numPendingTests":0,"testResults":[]}`,
			wantTotal: 3, wantPassed: 3,
		},
		{
			name: "WithFailures",
			input: `{"numTotalTests":2,"numPassedTests":1,"numFailedTests":1,"numPendingTests":0,"testResults":[{"assertionResults":[{"title":"should fail","status":"failed","failureMessages":["Expected true to be false"]}]}]}`,
			wantTotal: 2, wantPassed: 1, wantFailed: 1, wantErrCount: 1,
		},
		{
			name:      "NoMarker",
			input:     `{"some":"other json"}`,
			wantErr:   true,
			errSubstr: "no Jest JSON found",
		},
		{
			name: "WithPending",
			input: `{"numTotalTests":3,"numPassedTests":2,"numFailedTests":0,"numPendingTests":1,"testResults":[]}`,
			wantTotal: 3, wantPassed: 2, wantSkipped: 1,
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
