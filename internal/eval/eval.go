package eval

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/EinfachNiklas/cochabench/internal/run"
	"github.com/EinfachNiklas/cochabench/internal/tools"
)

func Evaluate(ctx context.Context, cmd *cli.Command) error {
	runID := cmd.String("runID")
	if len(runID) == 0 {
		return fmt.Errorf("No runID provided. Nothing to evaluate")
	}
	dirPath := cmd.String("path")
	if len(dirPath) == 0 {
		dirPath = "./"
	}
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

	var h LanguageHandler
	var tempDir string
	var cleanup func()
	if challengeConfig.ChallengeType == "javascript" {
		h = JavascriptHandler{}
		tempDir, cleanup, err = h.PrepareEnvironment(dirPath, runID)
	}
	defer cleanup()
	if err != nil {
		return fmt.Errorf("Failed to prepare temporary environment for evaluation: %w", err)
	}
	result, err := h.ExecuteTests(tempDir)
	if err != nil {
		return fmt.Errorf("Failed to execute tests: %w", err)
	}

	fmt.Println(result)

	return nil
}
