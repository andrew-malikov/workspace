package scaffold

import (
	"github.com/andrew-malikov/workspace/view"
	"github.com/andrew-malikov/workspace/workspaces"
	"github.com/urfave/cli/v3"
)

func NewCommand(renderer *view.Renderer, session workspaces.Session) *cli.Command {
	return &cli.Command{
		Name:    "scaffold",
		Aliases: []string{"sc"},
		Usage:   "scaffold tracked project config",
		Commands: []*cli.Command{
			newTestCommand(renderer, session),
		},
	}
}
