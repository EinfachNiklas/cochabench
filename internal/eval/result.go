package eval

type TestResult struct {
	Passed       bool
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int
	Errors       []TestError
}

type TestError struct {
	TestName string
	Message  string
	Stack    string
}
