package main

import (
	"context"
	"log"
	"os"

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
				Name:    "eval",
				Aliases: []string{"e"},
				Usage:   "Evaluate Coding Challenge",
				Action:  eval.Evaluate,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "path",
						Aliases: []string{"p"},
						Usage:   "Path to directory of challenge",
					},
				},
			},
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
				},
			},
		},
	}
	cmd.Suggest = true
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
