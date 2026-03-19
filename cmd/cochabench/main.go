package main

import (
	"context"
	"log"
	"os"

	challenges "github.com/EinfachNiklas/cochabench/internal/challenge"
	"github.com/EinfachNiklas/cochabench/internal/config"
	"github.com/EinfachNiklas/cochabench/internal/eval"
	"github.com/EinfachNiklas/cochabench/internal/run"
	"github.com/google/uuid"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "cochabench",
		Usage: "Handle the coding challenges",
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Handle Run Events",
				Commands: []*cli.Command{
					{
						Name:   "init",
						Usage:  "Initializes Run for current challenge",
						Action: run.Init,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "name",
								Value:   uuid.NewString(),
								Aliases: []string{"n"},
								Usage:   "Name of run to initialize",
							},
						},
					}, {
						Name:   "start",
						Usage:  "Starts Run for current challenge",
						Action: run.Start,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "id",
								Aliases: []string{"i"},
								Usage:   "ID of run to start",
							},
						},
					},
					{
						Name:   "stop",
						Usage:  "Stops Run for current challenge",
						Action: run.Stop,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "id",
								Aliases: []string{"i"},
								Usage:   "ID of run to stop",
							},
						},
					},
					{
						Name:    "eval",
						Aliases: []string{"e"},
						Usage:   "Evaluate Coding Challenge",
						Action:  eval.Evaluate,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "runID",
								Aliases: []string{"i"},
								Usage:   "RunID of Run to evaluate",
							},
							&cli.StringFlag{
								Name:    "path",
								Aliases: []string{"p"},
								Usage:   "Path to directory of challenge",
							},
							&cli.BoolFlag{
								Name:    "debug",
								Aliases: []string{"d"},
								Usage:   "Debug Mode: Keep tmp dir and print tmp location",
							},
							&cli.BoolFlag{
								Name:    "no-ai-eval",
								Aliases: []string{"no-ai"},
								Usage:   "Disables the AI Evaluation: Values Quality, Maintainability and Security will be set to -1",
							},
						},
					},
					{
						Name:   "cancel",
						Usage:  "Cancels already started Run for current challenge",
						Action: run.Cancel,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "id",
								Aliases: []string{"i"},
								Usage:   "ID of run to cancel",
							},
						},
					},
					{
						Name:    "list",
						Usage:   "Lists all runs for current challenge",
						Aliases: []string{"l"},
						Action:  run.List,
					},
				},
			},
			{
				Name:    "challenge",
				Usage:   "Handle challenges",
				Aliases: []string{"c"},
				Commands: []*cli.Command{
					{
						Name:    "list",
						Usage:   "Lists all available challenges",
						Aliases: []string{"l"},
						Action:  challenges.List,
					},
					{
						Name:   "get",
						Usage:  "Downloads a set challenge",
						Action: challenges.Get,
						Commands: []*cli.Command{
							{
								Name:    "all",
								Usage:   "Downloads all challenges",
								Aliases: []string{"a"},
								Action:  challenges.GetAll,
							},
						},
					},
				},
			},
			{
				Name:  "config",
				Usage: "Manage configuration",
				Commands: []*cli.Command{
					{
						Name:    "init",
						Usage:   "Initialize the config file",
						Aliases: []string{"initialize", "i"},
						Action:  config.Initialize,
					},
					{
						Name:    "show",
						Usage:   "Show all config values",
						Aliases: []string{"s"},
						Action:  config.Show,
					},
					{
						Name:   "get",
						Usage:  "Get a config value. Usage: config get <key>",
						Action: config.Get,
					},
					{
						Name:   "set",
						Usage:  "Set a config value. Usage: config set <key> <value>",
						Action: config.Set,
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
