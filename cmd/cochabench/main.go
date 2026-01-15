package main

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/EinfachNiklas/CodingChallengesBenchmark/internal/eval"
)

func validatePath(path string) error {
	stat, err := os.Stat(path)
	if len(path) == 0 {
		return errors.New("No path provided")
	}
	if err != nil {
		return errors.New("This directory does not exist")
	}
	if !stat.IsDir() {
		return errors.New("The provided path is not a directory")
	}
	return nil
}

func evaluate(ctx context.Context, cmd *cli.Command) error {
	err := validatePath(cmd.String("path"))
	if err != nil {
		return err
	}
	err = eval.ValidateDirStructure(cmd.String("path"))
	if err != nil {
		return err
	}
	challengeConfig, err := eval.LoadChallengeConfig(filepath.Join(cmd.String("path"), "config.json"))
	if err != nil {
		return err
	}
	log.Println(challengeConfig.Name)

	return nil
}

func main() {
	cmd := &cli.Command{
		Name:  "cochabench",
		Usage: "Handle the coding challenges",
		Commands: []*cli.Command{
			{
				Name:    "eval",
				Aliases: []string{"e"},
				Usage:   "Evaluate Coding Challenge",
				Action:  evaluate,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "path",
						Aliases: []string{"p"},
						Usage:   "Path to directory of challenge",
					},
				},
			},
		},
	}
	cmd.Suggest = true
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
