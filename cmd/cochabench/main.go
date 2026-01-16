package main

import (
	"context"
	"log"
	"os"

	"github.com/EinfachNiklas/cochabench/internal/eval"
	timer "github.com/EinfachNiklas/cochabench/internal/run"

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
						Name:   "start",
						Usage:  "Starts Run for current challenge",
						Action: timer.Start,
					},
					{
						Name:   "stop",
						Usage:  "Stops Run for current challenge",
						Action: timer.Stop,
					},
					{
						Name:   "cancel",
						Usage:  "Cancels already started Run for current challenge",
						Action: timer.Cancel,
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
