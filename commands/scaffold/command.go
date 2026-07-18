package scaffold

import (
	"github.com/andrew-malikov/workspace/view"
	"github.com/urfave/cli/v3"
)

func NewCommand(renderer *view.Renderer) *cli.Command {
	return &cli.Command{
		Name:    "scaffold",
		Aliases: []string{"sc"},
		Usage:   "scaffold tracked project config",
		Commands: []*cli.Command{
			newTestCommand(renderer),
		},
	}
}
