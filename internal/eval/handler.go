package eval

import "context"

type LanguageHandler interface {
	PrepareEnvironment(challengePath string, runID string) (tempDir string, cleanup func(), err error)

	ExecuteTests(ctx context.Context, tempDir string) (*TestResult, error)
}
