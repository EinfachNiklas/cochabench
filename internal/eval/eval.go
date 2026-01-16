package eval

import (
	"context"
	"log"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/EinfachNiklas/cochabench/internal/tools"
)

func Evaluate(ctx context.Context, cmd *cli.Command) error {
	err := tools.ValidateDirPath(cmd.String("path"))
	if err != nil {
		return err
	}
	err = tools.ValidateDirStructure(cmd.String("path"))
	if err != nil {
		return err
	}
	challengeConfig, err := tools.LoadChallengeConfig(filepath.Join(cmd.String("path"), "config.json"))
	if err != nil {
		return err
	}
	log.Println(challengeConfig.Name)

	return nil
}
