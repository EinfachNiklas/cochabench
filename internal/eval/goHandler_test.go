package eval

import (
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
