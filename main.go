package main

import (
	"context"
	"fmt"
	"os"

	"github.com/andrew-malikov/workspace/commands/containers"
	"github.com/andrew-malikov/workspace/commands/git/clear"
	"github.com/andrew-malikov/workspace/commands/list"
	"github.com/andrew-malikov/workspace/commands/scaffold"
	test "github.com/andrew-malikov/workspace/commands/test"
	"github.com/andrew-malikov/workspace/commands/track"
	"github.com/andrew-malikov/workspace/commands/untrack"
	"github.com/andrew-malikov/workspace/commands/up"

	"github.com/urfave/cli/v3"
)

func main() {
	commands := &cli.Command{
		Name:                   "ws",
		Usage:                  "workspace you way out",
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			list.NewCommand(),
			track.NewCommand(),
			up.NewCommand(),
			scaffold.NewCommand(),
			test.NewCommand(),
			containers.NewCommand(),
			untrack.NewCommand(),
			clear.NewCommand(),
		},
	}

	if err := commands.Run(context.Background(), os.Args); err != nil {
		fmt.Print(err)
	}
}
