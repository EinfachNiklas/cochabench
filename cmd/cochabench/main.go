package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func evaluate(ctx context.Context, cmd *cli.Command) error {
	fmt.Println("Hello World")
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
			},
		},
	}
	cmd.Suggest = true
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
