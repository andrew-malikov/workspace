package clear

import (
	"context"
	"os"

	cfg "github.com/andrew-malikov/workspace/config"
	"github.com/andrew-malikov/workspace/view"

	"github.com/urfave/cli/v3"

	tea "charm.land/bubbletea/v2"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "clear",
		Aliases: []string{"clr", "c"},
		Usage:   "clear hanging branches",
		// todo: probably outdated by tui
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"d"},
				Value:   false,
				Usage:   "only prints out the plan",
			},
			&cli.BoolFlag{
				Name:    "team",
				Aliases: []string{"t"},
				Value:   false,
				Usage:   "include team members",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := cfg.LoadConfig()
			if err != nil {
				return err
			}

			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			project := config.ResolveProjectByDir(dir)

			// todo: use the wd as a draft project instead of exiting
			if project == nil {
				return view.RenderDirectoryIsNotTrackedYet(dir)
			}

			branches, err := project.ListStaleBranches(false)
			if err != nil {
				return err
			}

			if _, err := tea.NewProgram(newUi(branches)).Run(); err != nil {
				return err
			}

			return nil
		},
	}
}
