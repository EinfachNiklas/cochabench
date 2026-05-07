package main

import (
	"context"
	"log"
	"os"
	"time"

	challenges "github.com/EinfachNiklas/cochabench/internal/challenge"
	cochabenchdata "github.com/EinfachNiklas/cochabench/internal/cochabenchData"
	"github.com/EinfachNiklas/cochabench/internal/config"
	"github.com/EinfachNiklas/cochabench/internal/eval"
	"github.com/EinfachNiklas/cochabench/internal/run"
	"github.com/EinfachNiklas/cochabench/internal/tools"
	"github.com/google/uuid"

	"github.com/urfave/cli/v3"
)

var version = "dev"

func main() {
	cmd := &cli.Command{
		Name:    "cochabench",
		Usage:   "Handle the coding challenges",
		Version: tools.GetBuildVersion(version),
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
							&cli.BoolFlag{
								Name:    "print-id-only",
								Aliases: []string{"id-only"},
								Usage:   "Only prints the runID of the created run; used for automation",
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
							&cli.IntFlag{
								Name:    "number-of-agents",
								Aliases: []string{"n"},
								Value:   3,
								Usage:   "Number of AI evaluation agents to run",
							},
							&cli.IntFlag{
								Name:    "ai-eval-iterations",
								Aliases: []string{"iterations"},
								Value:   20,
								Usage:   "Maximum iterations the AI evaluation agents can use",
							},
							&cli.DurationFlag{
								Name:    "timeout",
								Aliases: []string{"t"},
								Value:   5 * time.Minute,
								Usage:   "Time until an evaluation times out",
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
			{
				Name:    "data",
				Usage:   "Handle combined run data",
				Aliases: []string{"d"},
				Commands: []*cli.Command{
					{
						Name:    "merge",
						Usage:   "Combine the data of multiple challenges into on database. Expects the challenges to be directories inside the provided path.",
						Aliases: []string{"m"},
						Action:  cochabenchdata.MergeDB,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "path",
								Usage:   "Provide the path the merge is executed from",
								Aliases: []string{"p"},
								Value:   "./",
							},
						},
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
