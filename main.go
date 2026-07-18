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
	"github.com/andrew-malikov/workspace/console"
	"github.com/andrew-malikov/workspace/view"

	"github.com/urfave/cli/v3"
)

func main() {
	os.Exit(run(context.Background(), os.Args, console.OS()))
}

func run(ctx context.Context, args []string, terminal console.Console) int {
	renderer, err := view.NewRenderer(terminal.Output, terminal.Color)
	if err != nil {
		presentFailure(terminal, err)
		return 1
	}

	command := newCommand(terminal, renderer)
	if err := command.Run(ctx, args); err != nil {
		presentFailure(terminal, err)
		return 1
	}
	return 0
}

func newCommand(terminal console.Console, renderer *view.Renderer) *cli.Command {
	return &cli.Command{
		Name:                   "ws",
		Usage:                  "workspace you way out",
		UseShortOptionHandling: true,
		Reader:                 terminal.Input,
		Writer:                 terminal.Output,
		ErrWriter:              terminal.Error,
		ExitErrHandler:         func(context.Context, *cli.Command, error) {},
		Commands: []*cli.Command{
			list.NewCommand(renderer),
			track.NewCommand(renderer),
			up.NewCommand(renderer),
			scaffold.NewCommand(renderer),
			test.NewCommand(terminal),
			containers.NewCommand(renderer),
			untrack.NewCommand(renderer),
			clear.NewCommand(terminal),
		},
	}
}

func presentFailure(terminal console.Console, err error) {
	_, _ = fmt.Fprintln(terminal.Error, view.SafeText(err.Error()))
}
