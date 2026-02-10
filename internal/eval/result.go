package eval

import "time"

type TestResult struct {
	RunID        string
	Passed       bool
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int
	Duration     time.Duration
	Output       string
	Errors       []TestError
}

type TestError struct {
	TestName string
	Message  string
	Stack    string
}
