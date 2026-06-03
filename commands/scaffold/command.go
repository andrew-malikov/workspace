package scaffold

import "github.com/urfave/cli/v3"

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "scaffold",
		Aliases: []string{"sc"},
		Usage:   "scaffold tracked project config",
		Commands: []*cli.Command{
			newTestCommand(),
		},
	}
}
