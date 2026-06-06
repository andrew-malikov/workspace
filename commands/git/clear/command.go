package clear

import (
	"context"
	"os"

	"github.com/andrew-malikov/workspace/vcs"
	"github.com/andrew-malikov/workspace/view"
	"github.com/andrew-malikov/workspace/workspaces"

	"github.com/urfave/cli/v3"

	tea "charm.land/bubbletea/v2"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "clear",
		Aliases: []string{"clr", "c"},
		Usage:   "clear stale branches matched by configured ownership filters from branch history",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			workspace, err := workspaces.LoadWorkspace()
			if err != nil {
				return err
			}

			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			project := workspace.ResolveProjectByDir(dir)

			// todo: use the wd as a draft project instead of exiting
			if project == nil {
				return view.RenderDirectoryIsNotTrackedYet(dir)
			}

			git, err := vcs.NewProjectGit(project.Dir)
			if err != nil {
				return err
			}

			ownership, err := vcs.NewBranchOwnershipOptions(
				workspace.Git.Clear.Ownership.LookbackCommits,
				vcs.BranchOwnershipFilterInput{
					AuthorEmails:    workspace.Git.Clear.Ownership.Include.AuthorEmails,
					AuthorNames:     workspace.Git.Clear.Ownership.Include.AuthorNames,
					MessagePatterns: workspace.Git.Clear.Ownership.Include.MessagePatterns,
				},
				vcs.BranchOwnershipFilterInput{
					AuthorEmails:    workspace.Git.Clear.Ownership.Exclude.AuthorEmails,
					AuthorNames:     workspace.Git.Clear.Ownership.Exclude.AuthorNames,
					MessagePatterns: workspace.Git.Clear.Ownership.Exclude.MessagePatterns,
				},
			)
			if err != nil {
				return err
			}

			branches, err := project.ListStaleBranches(false, ownership)
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
