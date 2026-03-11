package eval

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v3"

	cochabenchdata "github.com/EinfachNiklas/cochabench/internal/cochabenchData"
	"github.com/EinfachNiklas/cochabench/internal/eval/agent"
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
	noAIEval := cmd.Bool("no-ai-eval")

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

	store, err := cochabenchdata.LoadCochabenchStore(dirPath)
	if err != nil {
		return fmt.Errorf("Could not load CochabenchStore: %v\n", err)
	}
	defer store.Close()

	runData, found, err := store.GetEntry(runID)
	if err != nil {
		return fmt.Errorf("Could not load Run Data: %w", err)
	}
	if !found {
		return fmt.Errorf("This run does not exist")
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

	if noAIEval {
		runData.QualityScore = -1
		runData.MaintainabilityScore = -1
		runData.SecurityScore = -1
	} else {
		evaluator := agent.NewEvaluator()

		aiEvaluation, err := evaluator.Evaluate(tempDir)
		if err != nil {
			return err
		}
		runData.QualityScore = aiEvaluation.Quality
		runData.MaintainabilityScore = aiEvaluation.Maintainability
		runData.SecurityScore = aiEvaluation.Security
	}

	runData.TestDuration = testResult.Duration
	runData.TimedOut = timedOut
	runData.NumTotalTests = testResult.TotalTests
	runData.NumPassedTests = testResult.PassedTests
	runData.NumFailedTests = testResult.FailedTests

	printEvaluationAsTable(runData)

	err = store.SaveEntry(runData)
	if err != nil {
		return err
	}

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

func printEvaluationAsTable(entry *cochabenchdata.CochabenchEntry) {
	tb := tools.NewTableBuilder([]string{"RunID", "RunName", "Status", "StartTime", "EndTime", "Duration", "TimedOut", "Total", "Passed", "Failed", "Quality", "Maintainability", "Security"})
	tb.AddRow([]string{
		entry.RunID,
		entry.RunName,
		entry.RunStatus,
		tools.FmtTime(entry.StartTime),
		tools.FmtTime(entry.EndTime),
		entry.TestDuration.String(),
		fmt.Sprintf("%v", entry.TimedOut),
		fmt.Sprintf("%d", entry.NumTotalTests),
		fmt.Sprintf("%d", entry.NumPassedTests),
		fmt.Sprintf("%d", entry.NumFailedTests),
		fmt.Sprintf("%.2f", entry.QualityScore),
		fmt.Sprintf("%.2f", entry.MaintainabilityScore),
		fmt.Sprintf("%.2f", entry.SecurityScore),
	})
	fmt.Println(tb.Build())
}
