package eval

type LanguageHandler interface {
	PrepareEnvironment(challengePath, runID string) (tempDir string, cleanup func(), err error)

	ExecuteTests(tempDir string) (*TestResult, error)
}
