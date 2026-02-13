package git

import (
	"github.com/andrew-malikov/workspace/commands/git/clear"

	"github.com/urfave/cli/v3"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "git",
		Aliases:     []string{"g"},
		Description: "manage git",
		Commands: []*cli.Command{
			clear.NewCommand(),
		},
	}
}
