package clear

import (
	"context"
	"errors"

	"github.com/andrew-malikov/workspace/console"
	"github.com/andrew-malikov/workspace/vcs"
	"github.com/andrew-malikov/workspace/workspaces"

	"github.com/urfave/cli/v3"

	tea "charm.land/bubbletea/v2"
)

func NewCommand(terminal console.Console, session workspaces.Session) *cli.Command {
	return newCommand(terminal, func(ctx context.Context) (tea.Model, error) {
		return buildUI(ctx, session)
	}, runProgram)
}

type programRunner func(tea.Model, ...tea.ProgramOption) error
type modelBuilder func(context.Context) (tea.Model, error)

func runProgram(model tea.Model, options ...tea.ProgramOption) error {
	_, err := tea.NewProgram(model, options...).Run()
	return err
}

func newCommand(terminal console.Console, build modelBuilder, run programRunner) *cli.Command {
	return &cli.Command{
		Name:    "clear",
		Aliases: []string{"clr", "c"},
		Usage:   "clear stale branches matched by configured ownership filters from branch history",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if !terminal.InputTerminal || !terminal.OutputTerminal {
				return errors.New("ws clear requires interactive input and output terminals")
			}
			model, err := build(ctx)
			if err != nil {
				return err
			}
			if err := run(model, tea.WithInput(terminal.Input), tea.WithOutput(terminal.Output)); err != nil {
				return err
			}

			return nil
		},
	}
}

func buildUI(ctx context.Context, session workspaces.Session) (tea.Model, error) {
	workspace, err := session.Load()
	if err != nil {
		return nil, err
	}

	dir, err := session.Cwd()
	if err != nil {
		return nil, err
	}

	project := workspace.ResolveProjectByDir(dir)
	if project == nil {
		return nil, workspaces.DirectoryNotTrackedError{Dir: dir}
	}

	git, err := vcs.NewProjectGit(project.Dir)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	branches, err := project.ListStaleBranches(false, ownership)
	if err != nil {
		return nil, err
	}
	return newUi(branches, git), nil
}
