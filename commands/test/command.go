package testcmd

import (
	"context"
	"fmt"
	"os"

	cfg "github.com/andrew-malikov/workspace/config"
	"github.com/andrew-malikov/workspace/dotnet"
	"github.com/andrew-malikov/workspace/projects"
	"github.com/andrew-malikov/workspace/view"

	"github.com/urfave/cli/v3"
)

func NewCommand() *cli.Command {
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
			config, err := cfg.LoadConfig()
			if err != nil {
				return err
			}

			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			project := config.ResolveProjectByDir(dir)
			if project == nil {
				return view.RenderDirectoryIsNotTrackedYet(dir)
			}

			requestedKinds := resolveRequestedKinds(cmd.Bool(string(projects.UnitTestKind)), cmd.Bool(string(projects.IntegrationTestKind)), cmd.Bool(string(projects.ComponentTestKind)))
			if len(requestedKinds) == 0 {
				requestedKinds = project.Test.ConfiguredKinds()
			}

			if len(requestedKinds) == 0 {
				return fmt.Errorf("no tests are configured for project %s", project.Alias)
			}

			for _, kind := range requestedKinds {
				target := project.Test.Target(kind)
				if !target.IsConfigured() {
					return fmt.Errorf("%s tests are not configured for project %s", kind, project.Alias)
				}

				if err := dotnet.RunTest(ctx, project.Dir, kind, target); err != nil {
					return err
				}
			}

			return nil
		},
	}
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
