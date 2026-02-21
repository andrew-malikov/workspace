package main

import (
	"context"
	"fmt"
	"os"

	"github.com/andrew-malikov/workspace/commands/git"
	"github.com/andrew-malikov/workspace/commands/track"
	"github.com/andrew-malikov/workspace/commands/untrack"

	"github.com/urfave/cli/v3"
)

func main() {
	commands := &cli.Command{
		Name:  "ws",
		Usage: "workspace you way out",
		Commands: []*cli.Command{
			track.NewCommand(),
			untrack.NewCommand(),
			git.NewCommand(),
		},
	}

	if err := commands.Run(context.Background(), os.Args); err != nil {
		fmt.Print(err)
	}
}
