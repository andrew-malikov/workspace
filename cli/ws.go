package main

import (
	"context"
	"fmt"
	"os"
	"ws/commands/track"
	"ws/commands/untrack"

	"github.com/urfave/cli/v3"
)

func main() {
	commands := &cli.Command{
		Usage: "workspace you way out",
		Commands: []*cli.Command{
			track.NewCommand(),
			untrack.NewCommand(),
		},
	}

	if err := commands.Run(context.Background(), os.Args); err != nil {
		fmt.Print(err)
	}
}
