package clear

import (
	"context"
	"os"

	cfg "github.com/andrew-malikov/workspace/config"
	"github.com/andrew-malikov/workspace/vcs"
	"github.com/andrew-malikov/workspace/view"

	"github.com/urfave/cli/v3"

	tea "charm.land/bubbletea/v2"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "clear",
		Aliases: []string{"clr", "c"},
		Usage:   "clear branches owned by the current git user based on branch history",
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

			git, err := vcs.NewProjectGit(project.Dir)
			if err != nil {
				return err
			}

			branches, err := project.ListStaleBranches(false)
			if err != nil {
				return err
			}

			if _, err := tea.NewProgram(newUi(branches, git)).Run(); err != nil {
				return err
			}

			return nil
		},
	}
}
