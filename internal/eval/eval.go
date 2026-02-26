package eval

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/EinfachNiklas/cochabench/internal/run"
	"github.com/EinfachNiklas/cochabench/internal/tools"
)

const TIMEOUT_DURATION = 5 * time.Minute

func Evaluate(ctx context.Context, cmd *cli.Command) error {
	runID := cmd.String("runID")
	if len(runID) == 0 {
		return fmt.Errorf("No runID provided. Nothing to evaluate")
	}
	dirPath := cmd.String("path")
	if len(dirPath) == 0 {
		dirPath = "./"
	}
	debugMode := cmd.Bool("debug")

	err := tools.ValidateDirPath(dirPath)
	if err != nil {
		return err
	}
	err = tools.ValidateDirStructure(cmd.String("path"))
	if err != nil {
		return err
	}
	challengeConfig, err := tools.LoadChallengeConfig(filepath.Join(cmd.String("path"), "challenge.config.json"))
	if err != nil {
		return err
	}

	runData, err := run.LoadEntry(dirPath, runID)
	if err != nil {
		return fmt.Errorf("Could not load Run Data: %w", err)
	}

	if runData.RunStatus != "F" {
		return fmt.Errorf("Run is not finished. Nothing to evaluate")
	}

	handler, err := createHandler(challengeConfig.ChallengeType)
	if err != nil {
		return err
	}

	tempDir, cleanup, err := handler.PrepareEnvironment(dirPath, runID)
	if err != nil {
		return fmt.Errorf("Failed to prepare temporary environment for evaluation: %w", err)
	}
	if !debugMode {
		defer cleanup()
	} else {
		log.Printf("Location of tmpDir for eval: %s\n", tempDir)
	}

	testResult, err, timedOut := executeTests(handler, tempDir)
	if err != nil {
		return fmt.Errorf("Failed to execute tests: %w", err)
	}

	runData.TestDuration = testResult.Duration
	runData.PassedTests = testResult.Passed
	runData.TimedOut = timedOut
	runData.NumTotalTests = testResult.TotalTests
	runData.NumPassedTests = testResult.PassedTests
	runData.NumFailedTests = testResult.FailedTests
	runData.NumSkippedTests = testResult.SkippedTests

	fmt.Println(runData)

	runData.Write(dirPath)

	return nil
}

func createHandler(challengeType string) (LanguageHandler, error) {
	switch challengeType {
	case "javascript", "js":
		return JavascriptHandler{}, nil
	case "python", "py":
		return PythonHandler{}, nil
	case "go", "golang":
		return GoHandler{}, nil
	default:
		return nil, fmt.Errorf("unsupported challenge type: %s", challengeType)
	}
}

func executeTests(handler LanguageHandler, tempDir string) (result *TestResult, err error, timedOut bool) {
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT_DURATION)
	defer cancel()

	resultChan := make(chan *TestResult)
	errorChan := make(chan error)

	go func() {
		result, err := handler.ExecuteTests(ctx, tempDir)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	select {
	case result = <-resultChan:
		return result, nil, false
	case err = <-errorChan:
		return nil, fmt.Errorf("Failed to execute tests: %w", err), false
	case <-ctx.Done():
		return nil, fmt.Errorf("Test execution timed out"), true
	}
}
