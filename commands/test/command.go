package test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/andrew-malikov/workspace/console"
	"github.com/andrew-malikov/workspace/dotnet"
	"github.com/andrew-malikov/workspace/projects"
	"github.com/andrew-malikov/workspace/workspaces"

	"github.com/urfave/cli/v3"
)

func NewCommand(terminal console.Console) *cli.Command {
	return &cli.Command{
		Name:                   "test",
		Aliases:                []string{"t"},
		Usage:                  "run project dotnet tests",
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    string(projects.UnitTestKind),
				Aliases: []string{"u"},
				Usage:   "run unit tests",
			},
			&cli.BoolFlag{
				Name:    string(projects.IntegrationTestKind),
				Aliases: []string{"i"},
				Usage:   "run integration tests",
			},
			&cli.BoolFlag{
				Name:    string(projects.ComponentTestKind),
				Aliases: []string{"c"},
				Usage:   "run component tests",
			},
		},
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
			if project == nil {
				return workspaces.DirectoryNotTrackedError{Dir: dir}
			}

			requestedKinds := resolveRequestedKinds(cmd.Bool(string(projects.UnitTestKind)), cmd.Bool(string(projects.IntegrationTestKind)), cmd.Bool(string(projects.ComponentTestKind)))
			if len(requestedKinds) == 0 {
				requestedKinds = project.Test.ConfiguredKinds()
			}

			if len(requestedKinds) == 0 {
				return fmt.Errorf("no tests are configured for project %s", project.Alias)
			}

			runCtx, stop := interruptContext(ctx)
			defer stop()

			// todo: keep the data mapping here and move all the logic into project module
			testRunner := dotnet.NewTestRunner(dotnet.StdCommandRunner{Input: terminal.Input}, terminal.Output, terminal.Error, terminal.SharedTerminal)
			return runCategories(runCtx, project, requestedKinds, testRunner, dotnet.NewTestRun)

		},
	}
}

func interruptContext(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, os.Interrupt)
}

type categoryRunner interface {
	Run(context.Context, string, dotnet.TestRun, projects.TestKind, projects.TestTarget) error
}

type testRunFactory func(string) (dotnet.TestRun, error)

func runCategories(
	ctx context.Context,
	project *projects.Project,
	requestedKinds []projects.TestKind,
	runner categoryRunner,
	createRun testRunFactory,
) error {
	var run dotnet.TestRun
	runCreated := false
	for _, kind := range requestedKinds {
		target := project.Test.Target(kind)
		if !target.IsConfigured() {
			return fmt.Errorf("%s tests are not configured for project %s", kind, project.Alias)
		}

		if !runCreated {
			var err error
			run, err = createRun(project.Dir)
			if err != nil {
				return err
			}
			runCreated = true
		}

		if err := runner.Run(ctx, project.Dir, run, kind, target); err != nil {
			return err
		}
	}
	return nil
}

func resolveRequestedKinds(unit bool, integration bool, component bool) []projects.TestKind {
	kinds := make([]projects.TestKind, 0, len(projects.AllTestKinds))
	if unit {
		kinds = append(kinds, projects.UnitTestKind)
	}
	if integration {
		kinds = append(kinds, projects.IntegrationTestKind)
	}
	if component {
		kinds = append(kinds, projects.ComponentTestKind)
	}
	return kinds
}
