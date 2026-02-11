package eval

import "time"

type TestResult struct {
	Passed       bool
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int
	Duration     time.Duration
	Errors       []TestError
}

type TestError struct {
	TestName string
	Message  string
	Stack    string
}
